package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Up struct {
	awsUp          awsUp
	gcpUp          gcpUp
	envGetter      envGetter
	envIDGenerator envIDGenerator
}

type awsUp interface {
	Execute(awsUpConfig AWSUpConfig, state storage.State) error
}

type gcpUp interface {
	Execute(gcpUpConfig GCPUpConfig, state storage.State) error
}

type envGetter interface {
	Get(name string) string
}

type envIDGenerator interface {
	Generate() (string, error)
}

type upConfig struct {
	awsAccessKeyID       string
	awsSecretAccessKey   string
	awsRegion            string
	awsBOSHAZ            string
	gcpServiceAccountKey string
	gcpProjectID         string
	gcpZone              string
	gcpRegion            string
	iaas                 string
	name                 string
	opsFile              string
}

func NewUp(awsUp awsUp, gcpUp gcpUp, envGetter envGetter,
	envIDGenerator envIDGenerator) Up {
	return Up{
		awsUp:          awsUp,
		gcpUp:          gcpUp,
		envGetter:      envGetter,
		envIDGenerator: envIDGenerator,
	}
}

func (u Up) Execute(args []string, state storage.State) error {
	var desiredIAAS string

	config, err := u.parseArgs(args)
	if err != nil {
		return err
	}

	switch {
	case state.IAAS == "" && config.iaas == "":
		return errors.New("--iaas [gcp, aws] must be provided")
	case state.IAAS == "" && config.iaas != "":
		desiredIAAS = config.iaas
	case state.IAAS != "" && config.iaas == "":
		desiredIAAS = state.IAAS
	case state.IAAS != "" && config.iaas != "":
		if state.IAAS != config.iaas {
			return fmt.Errorf("The iaas type cannot be changed for an existing environment. The current iaas type is %s.", state.IAAS)
		} else {
			desiredIAAS = state.IAAS
		}
	}

	if state.EnvID != "" && config.name != "" {
		return fmt.Errorf("The director name cannot be changed for an existing environment. Current name is %s.", state.EnvID)
	}

	if state.EnvID == "" {
		if config.name == "" {
			state.EnvID, err = u.envIDGenerator.Generate()
			if err != nil {
				return err
			}
		} else {
			state.EnvID = config.name
		}
	}

	switch desiredIAAS {
	case "aws":
		err = u.awsUp.Execute(AWSUpConfig{
			AccessKeyID:     config.awsAccessKeyID,
			SecretAccessKey: config.awsSecretAccessKey,
			Region:          config.awsRegion,
			BOSHAZ:          config.awsBOSHAZ,
			OpsFilePath:     config.opsFile,
		}, state)
	case "gcp":
		err = u.gcpUp.Execute(GCPUpConfig{
			ServiceAccountKeyPath: config.gcpServiceAccountKey,
			ProjectID:             config.gcpProjectID,
			Zone:                  config.gcpZone,
			Region:                config.gcpRegion,
			OpsFilePath:           config.opsFile,
		}, state)
	default:
		return fmt.Errorf("%q is an invalid iaas type, supported values are: [gcp, aws]", desiredIAAS)
	}

	if err != nil {
		return err
	}

	return nil
}

func (u Up) parseArgs(args []string) (upConfig, error) {
	var config upConfig

	upFlags := flags.New("up")

	upFlags.String(&config.iaas, "iaas", u.envGetter.Get("BBL_IAAS"))

	upFlags.String(&config.awsAccessKeyID, "aws-access-key-id", u.envGetter.Get("BBL_AWS_ACCESS_KEY_ID"))
	upFlags.String(&config.awsSecretAccessKey, "aws-secret-access-key", u.envGetter.Get("BBL_AWS_SECRET_ACCESS_KEY"))
	upFlags.String(&config.awsRegion, "aws-region", u.envGetter.Get("BBL_AWS_REGION"))
	upFlags.String(&config.awsBOSHAZ, "aws-bosh-az", u.envGetter.Get("BBL_AWS_BOSH_AZ"))

	upFlags.String(&config.gcpServiceAccountKey, "gcp-service-account-key", u.envGetter.Get("BBL_GCP_SERVICE_ACCOUNT_KEY"))
	upFlags.String(&config.gcpProjectID, "gcp-project-id", u.envGetter.Get("BBL_GCP_PROJECT_ID"))
	upFlags.String(&config.gcpZone, "gcp-zone", u.envGetter.Get("BBL_GCP_ZONE"))
	upFlags.String(&config.gcpRegion, "gcp-region", u.envGetter.Get("BBL_GCP_REGION"))

	upFlags.String(&config.name, "name", "")
	upFlags.String(&config.opsFile, "ops-file", "")

	err := upFlags.Parse(args)
	if err != nil {
		return upConfig{}, err
	}

	return config, nil
}
