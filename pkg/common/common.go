package common

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CheckVaultProfiles - This takes a map of environment names and checks that they
// are present in aws-vault.
func CheckVaultProfiles(requiredProfiles []string) error {
	vaultList, vaultErr := exec.Command("aws-vault", "list", "--credentials").Output()
	if vaultErr != nil {
		log.Errorf("Error: Ran `aws-vault` and encountered error: %s", vaultErr)

		if len(vaultList) > 0 {
			fmt.Printf(string(vaultList))
		}

		return vaultErr
	}

	availableProfiles := strings.Split(string(vaultList), "\n")

	missingProfiles, err := CompareAvailableVsNeeded(availableProfiles, requiredProfiles)
	if err != nil {
		log.Errorf("Error: One or more required profiles are missing from `aws-vault`.")
		log.Errorf("Make sure you have permission for all required AWS profiles and use `aws-vault add` to configure them.")
		log.Errorf("Missing aws-vault Profiles: %s", missingProfiles)
		return err
	}

	log.Infof("Vault contains credentials for all required AWS profiles.")

	return nil
}

// CheckVPN is a helper function to check Cisco AnyConnect VPN is connected.
func CheckVPN() error {
	vpnOutput, vpnErr := exec.Command("/opt/cisco/anyconnect/bin/vpn", "state").Output()
	if vpnErr != nil {
		return fmt.Errorf("Error: Ran `/opt/cisco/anyconnect/bin/vpn state` and encountered error: %s", vpnErr)
	}

	vpnString := string(vpnOutput)

	if strings.Contains(vpnString, "state: Connected") {
		log.Infof("Cisco AnyConnect reports that the VPN is Connected.")
		return nil
	}

	if strings.Contains(vpnString, "state: Disconnected") {
		return fmt.Errorf("Cisco AnyConnect reports the VPN is disconnected")
	}

	return fmt.Errorf("The VPN is in an unknown state: %s", vpnString)
}

// CheckYubikey is a helper function to check the presence and state of a Yubikey
func CheckYubikey() error {
	// Killing ykman
	killErr := exec.Command("killall", "ykman").Run()
	if killErr != nil {
		exitCode := killErr.(*exec.ExitError).ExitCode()
		if exitCode > 1 {
			return fmt.Errorf("Error trying to kill ykman: %s", killErr)
		}
		log.Debugf("Process 'ykman' was not running.")
	}

	ykmanList, ykmanErr := exec.Command("ykman", "list").Output()
	if ykmanErr != nil {
		return fmt.Errorf("Error: Ran `ykman list` and encountered error: %s", ykmanErr)
	}

	if len(ykmanList) == 0 {
		return fmt.Errorf("A Yubikey was not found. Have you inserted your Yubikey?")
	}

	log.Infof("Found a Yubikey connected.")

	return nil
}

// CompareAvailableVsNeeded is a helper function to compare two arrays of strings. Returns an array of
// missing items.
func CompareAvailableVsNeeded(available []string, needed []string) (notFound []string, err error) {
	for _, thisNeeded := range needed {
		found := false
		for _, thisAvailable := range available {
			if thisAvailable == thisNeeded {
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, thisNeeded)
		}
	}

	if len(notFound) > 0 {
		return notFound, errors.New("One or more Needed Items Not Found")
	}

	return
}

// ConvertKeyValuesToMap is a helper to convert an array of key=value environment style variables
// and convert them to a map of keyed values.
func ConvertKeyValuesToMap(inputs []string) map[string]string {
	outputs := make(map[string]string)
	for _, thisInput := range inputs {
		keyValSlice := strings.SplitN(thisInput, "=", 2)
		outputs[keyValSlice[0]] = keyValSlice[1]
	}

	return outputs
}

// Filter - Filters a slice of strings according to a function.
func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// IsCommandAvailable is a helper to test whether a given command name is present on the system.
func IsCommandAvailable(name string) bool {
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
