# pay-cli

⚠️ This project was a firebreak proof of concept. It has some useful scripts but is not maintained by the Pay team.

## Deployer Command
This is a command help automate the rotation of the `ci.deployer` AWS API key and secret in each of our environments.

As the process requires us to rotate the keys in one environment and then store the new ones using credstash in either of the CI or Deploy AWS Environments, the best way to do this is with a script to safely rotate the keys whilst minimising any opportunity for expired or invalid keys to be used.

### Requirements
Before using this script, you will need the following things to be installed:

* aws-vault
* credstash
* ykman

You will also need access to the necessary AWS accounts for the credentials you are rotating, e.g. `govuk-pay-dev` and `govuk-pay-test` require access to `govuk-pay-ci`, and `govuk-pay-staging` and `govuk-pay-prod` require access to `govuk-pay-deploy`.

### Usage
At the current time of writing, this is the usage of the script:

```sh
./pay deployer --environment [dev/test/staging/production] --management-profile [ci/deploy]
```

The `--environment` is the target AWS environment you wish to rotate the IAM Access Keys for. The `--management-profile` is the AWS environment used to store the secrets using credstash.

There are some other options too:

* `--user [user.name]` - Defaults to "ci.deployer". Use this to rotate the keys for a different IAM user.
* `--dry-run` - Optional. Allows you to do read-only stuff to simulate most of the tool without making changes.
* `--verbose` - Optional. Sets logging to Debug output. Useful if you're having problems.

For a help printout, simply run:

```
./pay deployer --help
```

By default, the script will look for a Yubikey credential following the naming convention of `govuk-pay-[env]`, with `[env]` taking the form of dev, test, staging or production. You can override this bu using `--yubikey-profile` and `--yubikey-management-profile` if needed.


## Building
If you want to build this app yourself, run:

```sh
make build
```

This is controlled by the `Makefile` in the root of the project.

## Testing
Right now, there are only some basic tests for the common package, but if you'd like to run all available tests, you can run:

```sh
go test ./... -v
```

You should get output that looks like this:

```sh
?   	github.com/alphagov/pay-cli/cmd/pay	[no test files]
?   	github.com/alphagov/pay-cli/pkg/api	[no test files]
?   	github.com/alphagov/pay-cli/pkg/card	[no test files]
?   	github.com/alphagov/pay-cli/pkg/ci	[no test files]
?   	github.com/alphagov/pay-cli/pkg/cmd	[no test files]
=== RUN   TestHelpers
Running Suite: Helper Test Suite
================================
Random Seed: 1605878108
Will run 2 of 2 specs

••
Ran 2 of 2 Specs in 0.015 seconds
SUCCESS! -- 2 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestHelpers (0.03s)
PASS
ok  	github.com/alphagov/pay-cli/pkg/common	(cached)
?   	github.com/alphagov/pay-cli/pkg/config	[no test files]
?   	github.com/alphagov/pay-cli/pkg/link	[no test files]
?   	github.com/alphagov/pay-cli/pkg/toolbox	[no test files]
```
