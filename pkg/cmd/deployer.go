package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/alphagov/pay-cli/pkg/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/jedib0t/go-pretty/table"
	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

var dryRun, verbose bool
var targetProfile, managementProfile, targetUser, yubikeyProfile, yubikeyMgmtProfile string

// As we cannot have constant arrays in Go...
func getRequiredCommands() []string {
	return []string{"aws-vault", "credstash", "ykman"}
}

// Deployer is the command to rotate the Jenkins Deployer API key.
func Deployer() *cli.Command {
	return &cli.Command{
		Name:  "deployer",
		Usage: "Rotates the AWS API keys associated with the Jenkins Deployer",
		Description: ` This tool is intended to rotate the API keys for the ci.deployer user in multiple AWS environments. 
		You can specify the environment for the IAM user you wish to rotate, and the management environment 
		to store the credentials for later use by the Deployer.`,
		Flags: append(
			[]cli.Flag{
				&cli.StringFlag{
					Name:        "environment, e",
					Usage:       "The AWS environment to rotate the user for. (dev, test, staging, prod)",
					Required:    true,
					Destination: &targetProfile,
				},
				&cli.StringFlag{
					Name:        "user, u",
					Value:       "ci.deployer",
					Usage:       "User to rotate the API key for",
					Destination: &targetUser,
				},
				&cli.StringFlag{
					Name:        "management-profile, m",
					Usage:       "AWS Account used to store and retrieve the API keys (e.g. ci, deploy)",
					Required:    true,
					Destination: &managementProfile,
				},
				&cli.BoolFlag{
					Name:        "verbose, v",
					Value:       false,
					Usage:       "Enable verbose logging. For troubleshooting purposes.",
					Destination: &verbose,
				},
				&cli.BoolFlag{
					Name:        "dry-run",
					Value:       false,
					Usage:       "Does a dry-run of the rotation process. Useful to see if your credentials will work without actually rotating the keys",
					Destination: &dryRun,
				},
				&cli.StringFlag{
					Name:        "yubikey-profile",
					Usage:       "The name of the Yubikey credential to use for AWS",
					Destination: &yubikeyProfile,
				},
				&cli.StringFlag{
					Name:        "yubikey-management-profile",
					Usage:       "The name of the Yubikey credential to use for AWS management",
					Destination: &yubikeyMgmtProfile,
				},
			},
		),
		Before: SetGlobalFlags,
		Action: runDeployerCmd,
	}
}

// runDeployerCmd - This is the main function that creates a new key, stores it in credstash
// and removes the old one.
func runDeployerCmd(c *cli.Context) error {
	if verbose {
		log.SetLevel(log.DebugLevel)
		log.Warn("Verbose mode is enabled.")
	}

	envErr := checkEnvironment(c)
	if envErr != nil {
		return envErr
	}

	// Set up session for target - discard the Environment Variables.
	awsTargetCreds, setupErr := setupAWSSession(targetProfile, yubikeyProfile)
	if setupErr != nil {
		log.Fatalf("Fatal Error: There was a problem with setting up the AWS Session")
	}

	sessionTargetAWS, err := session.NewSession(&aws.Config{
		Region:                        aws.String("eu-west-1"),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials:                   awsTargetCreds,
	})
	if err != nil {
		log.Errorf("Error while opening AWS Session: %s", err)
		return err
	}

	// Create a IAM service client...
	svcIAM := iam.New(sessionTargetAWS)

	groupErr := verifyIAMGroup(svcIAM, targetProfile, "Admins")
	if groupErr != nil {
		log.Errorf("Error: %s", groupErr)
		return groupErr
	}

	currentKeys, getKeysErr := getCurrentAccessKeys(svcIAM)
	if getKeysErr != nil {
		log.Errorf("Error: %s", getKeysErr)
		return getKeysErr
	}

	if len(currentKeys.AccessKeyMetadata) > 1 {
		log.Errorf("There is already more than one IAM Access Key present for %s.\nWe require there to be only one key present before continuing.", targetUser)
		return fmt.Errorf("User %s already has two Access Keys in use. One of them must be removed before running this tool. ", targetUser)
	}

	// Take a note of the Old Key ID. We'll verify this later.
	oldIAMKey := strings.TrimSpace(*currentKeys.AccessKeyMetadata[0].AccessKeyId)

	// Get Credentials for Management Environment. We want these to authenticate with Credstash...
	managementCredentials, mgmtSetupErr := setupAWSSession(managementProfile, yubikeyMgmtProfile)
	if mgmtSetupErr != nil {
		log.Fatalf("Fatal Error: There was a problem with setting up the Management AWS Session")
	}

	getCredstashOldKey, getCredstashOldKeyErr := getCredstashValue(managementCredentials, "access_key.deployer")
	if getCredstashOldKeyErr != nil {
		return getCredstashOldKeyErr
	}

	// Compare the old IAM and Credstash Key to make sure they match...
	if oldIAMKey == getCredstashOldKey {
		log.Infof("Existing IAM Key and Credstash Key Match. Continuing...")
	} else {
		log.Warn("Existing IAM Key and Credstash Value don't match. Something funny is happening here...")
		log.Warnf("Key Reported by IAM: %s", oldIAMKey)
		log.Warnf("Key Reported by Credstash: %s", getCredstashOldKey)
		if !prompter.YesNo("Do you still want to continue?", false) {
			log.Fatalf("IAM and Credstash didn't match - user aborted.")
			return fmt.Errorf("IAM and Credstash didn't match - user aborted")
		}
	}

	if dryRun {
		log.Warn("This is a dry-run. Skipping Rotation.")
		log.Warn("Run this tool again without --dry-run to continue with creating a new key and rotating the values in credstash.")
	} else {
		// Create New Access Key...
		newKey, newKeyErr := createNewAccessKey(svcIAM)
		if newKeyErr != nil {
			log.Errorf("Error: %s", newKeyErr)
			return newKeyErr
		}

		putAccessKeyErr := putCredstashValue(managementCredentials, "access_key.deployer", *newKey.AccessKeyId)
		if putAccessKeyErr != nil {
			log.Fatalf("Error whilst storing access_key.deployer in credstash")
			return putAccessKeyErr
		}

		putSecretErr := putCredstashValue(managementCredentials, "secret_access_key.deployer", *newKey.SecretAccessKey)
		if putSecretErr != nil {
			log.Fatalf("Error whilst storing secret_access_key.deployer in credstash")
			return putSecretErr
		}

		if prompter.YesNo("Would you like to disable the old IAM Access Key?", false) {
			deactivateErr := deactivateAccessKey(svcIAM, oldIAMKey, targetUser)
			if deactivateErr != nil {
				return fmt.Errorf("Failed to disable old IAM key: %s", deactivateErr)
			}
		}

		fmt.Printf("New State of IAM Keys for %s:\n", targetUser)
		_, getKeysErr := getCurrentAccessKeys(svcIAM)
		if getKeysErr != nil {
			log.Errorf("Error: %s", getKeysErr)
			return getKeysErr
		}

		log.Info("You should manually verify that the Deployer user is correctly working and then delete the old (inactive) IAM Key.")
	}

	return nil
}

func checkEnvironment(c *cli.Context) error {
	// Check that all of our required commands are installed...
	for _, thisCommand := range getRequiredCommands() {
		if common.IsCommandAvailable(thisCommand) {
			log.Infof("App `%s` is installed. Continuing...", thisCommand)
		} else {
			log.Errorf("The binary `%s` was not found on the system.", thisCommand)
			log.Errorf("Run `brew install %s` to fix this.", thisCommand)
			return fmt.Errorf("Error: %s is not installed. ", thisCommand)
		}
	}

	requiredProfiles := []string{}
	requiredProfiles = append(requiredProfiles, targetProfile, managementProfile)

	err := common.CheckVaultProfiles(requiredProfiles)
	if err != nil {
		log.Fatal("An error occurred whilst checking aws-vault profiles.")
	}

	ykErr := common.CheckYubikey()
	if ykErr != nil {
		log.Error("An error occurred whilst checking for your Yubikey.")
		return ykErr
	}

	vpnErr := common.CheckVPN()
	if vpnErr != nil {
		log.Error("An error occurred whilst checking the VPN.")
		return vpnErr
	}

	return nil
}

// setupAWSSession - Takes an AWS env/profile name and sets the AWS_
// environment variables to create an AWS Session.
func setupAWSSession(targetEnv string, ykmanCredName string) (*credentials.Credentials, error) {
	awsAccountName := fmt.Sprintf("govuk-pay-%s", targetEnv)

	// Override the Yubikey Credential Name if required...
	if ykmanCredName != "" {
		os.Setenv("YKMAN_OATH_CREDENTIAL_NAME", ykmanCredName)
	} else {
		os.Setenv("YKMAN_OATH_CREDENTIAL_NAME", awsAccountName)
	}

	fmt.Printf("Running `aws-vault` for %s - You may need to enter your Yubikey passphrase.\n", awsAccountName)

	cmd := exec.Command("aws-vault", "exec", targetEnv, "--prompt", "ykman", "--", "env")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		log.Errorf("Error: Ran `aws-vault exec` and encountered error: %s", err)

		log.Errorf("Is your Yubikey connected and ykman working correctly?")
		return nil, err
	}
	outStr := string(stdoutBuf.Bytes())

	allEnvLines := strings.Split(outStr, "\n")
	awsEnvLines := common.Filter(allEnvLines, func(v string) bool {
		return strings.HasPrefix(v, "AWS")
	})

	awsEnvs := common.ConvertKeyValuesToMap(awsEnvLines)

	awsCreds := credentials.NewStaticCredentials(awsEnvs["AWS_ACCESS_KEY_ID"], awsEnvs["AWS_SECRET_ACCESS_KEY"], awsEnvs["AWS_SESSION_TOKEN"])
	return awsCreds, nil
}

// createNewAccessKey - Creates a new Access Key for the given IAM user.
func createNewAccessKey(svcIAM *iam.IAM) (*iam.AccessKey, error) {
	log.Infof("Creating new access key for IAM User %s", targetUser)
	accessKeyOutput, err := svcIAM.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: &targetUser,
	})

	if err != nil {
		return nil, err
	}

	return accessKeyOutput.AccessKey, nil
}

// deactivateAccessKey - Disable an IAM Access Key.
func deactivateAccessKey(svcIAM *iam.IAM, accessKey string, iamUser string) error {
	log.Infof("Deactivating access key %s for IAM User %s", accessKey, targetUser)
	_, err := svcIAM.UpdateAccessKey(&iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(accessKey),
		Status:      aws.String(iam.StatusTypeInactive),
		UserName:    aws.String(iamUser),
	})

	if err != nil {
		return err
	}

	return nil

}

// getCurrentAccessKeys - Fetches the current access keys for a given IAM user
// and outputs them into a printed table.
func getCurrentAccessKeys(svcIAM *iam.IAM) (*iam.ListAccessKeysOutput, error) {
	log.Infof("Fetching access keys for IAM User %s", targetUser)
	accessKeys, err := svcIAM.ListAccessKeys(&iam.ListAccessKeysInput{
		MaxItems: aws.Int64(5),
		UserName: &targetUser,
	})

	if err != nil {
		return nil, err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"User Name", "Key ID", "Time Created", "Status"})
	for _, thisKey := range accessKeys.AccessKeyMetadata {
		t.AppendRow(table.Row{
			*thisKey.UserName, *thisKey.AccessKeyId, thisKey.CreateDate, *thisKey.Status,
		})
	}
	t.Render()

	return accessKeys, nil
}

// getCredstashValue - Gets a secret from Credstash. A good way to test management credentials.
func getCredstashValue(awsCredentials *credentials.Credentials, credstashKey string) (string, error) {
	mgmtCreds, credErr := awsCredentials.Get()
	if credErr != nil {
		return "", credErr
	}

	os.Setenv("AWS_ACCESS_KEY_ID", mgmtCreds.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", mgmtCreds.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", mgmtCreds.SessionToken)
	os.Setenv("AWS_SECURITY_TOKEN", mgmtCreds.SessionToken)

	fullKeyName := fmt.Sprintf("%s.%s", targetProfile, credstashKey)
	outputBytes, cmdErr := exec.Command("credstash", "get", fullKeyName).Output()
	outputString := strings.TrimSpace(string(outputBytes))
	if cmdErr != nil {
		log.Errorf("Error Running Credstash: %s", cmdErr)
		log.Warnf("Output: %s", outputString)
		return "", cmdErr
	}

	log.Debugf("Credstash - got %s: %s", fullKeyName, outputString)
	return outputString, nil
}

// getCredstashValue - Stores a secret with Credstash.
func putCredstashValue(awsCredentials *credentials.Credentials, credstashKey string, credstashValue string) error {
	mgmtCreds, credErr := awsCredentials.Get()
	if credErr != nil {
		return credErr
	}

	os.Setenv("AWS_ACCESS_KEY_ID", mgmtCreds.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", mgmtCreds.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", mgmtCreds.SessionToken)
	os.Setenv("AWS_SECURITY_TOKEN", mgmtCreds.SessionToken)

	fullKeyName := fmt.Sprintf("%s.%s", targetProfile, credstashKey)
	outputBytes, cmdErr := exec.Command("credstash", "put", "-a", fullKeyName, credstashValue).Output()
	outputString := string(outputBytes)
	if cmdErr != nil {
		log.Errorf("Error Running Credstash: %s", cmdErr)
		log.Warnf("Output: %s", outputString)
		return cmdErr
	}

	log.Infof("Credstash - updated key: %s", fullKeyName)
	log.Debugf("Credstash - updated %s: %s", fullKeyName, credstashValue)
	return nil
}

func verifyIAMGroup(svcIAM *iam.IAM, targetProfile string, requiredGroupName string) error {
	// Get Caller Identity to check permissions.
	log.Infof("Checking your permissions for the %s profile", targetProfile)
	currentUser, err := svcIAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		return err
	}

	groupsForUser, err := svcIAM.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: currentUser.User.UserName,
	})
	if err != nil {
		return err
	}

	foundGroup := false
	for _, thisGroup := range groupsForUser.Groups {
		if *thisGroup.GroupName == requiredGroupName {
			foundGroup = true
			break
		}
	}

	if !foundGroup {
		log.Errorf("IAM user is not a member of the required %s group", requiredGroupName)
		return fmt.Errorf("Your IAM user is not a member of the required %s group", requiredGroupName)
	}

	log.Infof("You have the necessary permissions for the AWS %s profile", targetProfile)
	return nil
}
