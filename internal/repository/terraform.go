package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"

	"one-click-aks-server/internal/config"
	"one-click-aks-server/internal/entity"
	"one-click-aks-server/internal/helper"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
	"golang.org/x/exp/slog"
)

type terraformRepository struct {
	appConfig *config.Config
}

func NewTerraformRepository(appConfig *config.Config) entity.TerraformRepository {
	return &terraformRepository{
		appConfig: appConfig,
	}
}

func (t *terraformRepository) TerraformAction(tfvar entity.TfvarConfigType, action string, storageAccountName string) (*exec.Cmd, *os.File, *os.File, error) {

	setEnvironmentVariable("terraform_directory", "tf")
	setEnvironmentVariable("root_directory", os.ExpandEnv("$ROOT_DIR"))
	setEnvironmentVariable("subscription_id", t.appConfig.ActLabsHubSubscriptionID)
	setEnvironmentVariable("resource_group_name", t.appConfig.ActLabsHubResourceGroupName)
	setEnvironmentVariable("storage_account_name", t.appConfig.ActLabsHubStorageAccountName)
	setEnvironmentVariable("container_name", "repro-project-tf-state-files")
	setEnvironmentVariable("tf_state_file_name", t.appConfig.UserAlias+"-terraform.tfstate")
	if t.appConfig.UseServicePrincipal {
		setEnvironmentVariable("ARM_CLIENT_ID", t.appConfig.AzureClientID)
		setEnvironmentVariable("ARM_CLIENT_SECRET", t.appConfig.AzureClientSecret)
		setEnvironmentVariable("ARM_SUBSCRIPTION_ID", t.appConfig.SubscriptionID)
		setEnvironmentVariable("ARM_TENANT_ID", t.appConfig.AzureTenantID)
	}

	// Sets terraform environment variables from tfvar

	tr := reflect.TypeOf(tfvar)
	// Loop over the fields in the struct.
	for i := 0; i < tr.NumField(); i++ {
		// Get the field and its value at the current index.
		field := reflect.TypeOf(tfvar).Field(i)
		value := reflect.ValueOf(tfvar).Field(i)

		// Set the environment variable of resource.
		encoded, _ := json.Marshal(conjson.NewMarshaler(value.Interface(), transform.ConventionalKeys()))

		slog.Debug("Field :" + field.Name + " Encoded String : " + string(encoded))

		// If a variable doesn't exist, just skip it and let terraform default do the magic.
		if string(encoded) != "null" {
			setEnvironmentVariable("TF_VAR_"+helper.CamelToConventional(field.Name), string(encoded))
		}
	}

	// Execute terraform script with appropriate action.
	cmd := exec.Command(os.ExpandEnv("$ROOT_DIR")+"/scripts/terraform.sh", action)
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		return cmd, rPipe, wPipe, err
	}
	cmd.Stdout = wPipe
	cmd.Stderr = wPipe
	if err := cmd.Start(); err != nil {
		return cmd, rPipe, wPipe, err
	}

	// Return stuff to the service.
	return cmd, rPipe, wPipe, nil
}

func (t *terraformRepository) ExecuteScript(script string, mode string, storageAccountName string) (*exec.Cmd, *os.File, *os.File, error) {
	setEnvironmentVariable("terraform_directory", "tf")
	setEnvironmentVariable("root_directory", os.ExpandEnv("$ROOT_DIR"))
	setEnvironmentVariable("subscription_id", t.appConfig.ActLabsHubSubscriptionID)
	setEnvironmentVariable("resource_group_name", t.appConfig.ActLabsHubResourceGroupName)
	setEnvironmentVariable("storage_account_name", t.appConfig.ActLabsHubStorageAccountName)
	setEnvironmentVariable("container_name", "repro-project-tf-state-files")
	setEnvironmentVariable("tf_state_file_name", t.appConfig.UserAlias+"-terraform.tfstate")
	setEnvironmentVariable("SCRIPT_MODE", mode)
	if t.appConfig.UseServicePrincipal {
		setEnvironmentVariable("ARM_CLIENT_ID", t.appConfig.AzureClientID)
		setEnvironmentVariable("ARM_CLIENT_SECRET", t.appConfig.AzureClientSecret)
		setEnvironmentVariable("ARM_SUBSCRIPTION_ID", t.appConfig.SubscriptionID)
		setEnvironmentVariable("ARM_TENANT_ID", t.appConfig.AzureTenantID)
	}

	// Execute terraform script with appropriate action.
	cmd := exec.Command("bash", "-c", "echo '"+script+"' | base64 -d | dos2unix | bash")
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		return cmd, rPipe, wPipe, err
	}
	cmd.Stdout = wPipe
	cmd.Stderr = wPipe
	if err := cmd.Start(); err != nil {
		return cmd, rPipe, wPipe, err
	}

	// Return stuff to the service.
	return cmd, rPipe, wPipe, nil
}

func (t *terraformRepository) UpdateAssignment(userId string, labId string, status string) error {

	// http call to actlabs-hub
	req, err := http.NewRequest("PUT", t.appConfig.ActlabsHubURL+"assignment/"+userId+"/"+labId+"/"+status, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACTLABS_AUTH_TOKEN"))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ProtectedLabSecret", entity.ProtectedLabSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not able to update assignment status")
	}

	return nil
}

func (t *terraformRepository) UpdateChallenge(userId string, labId string, status string) error {

	// http call to actlabs-hub
	req, err := http.NewRequest("PUT", t.appConfig.ActlabsHubURL+"challenge/"+userId+"/"+labId+"/"+status, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACTLABS_AUTH_TOKEN"))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ProtectedLabSecret", entity.ProtectedLabSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not able to update challenge status")
	}

	return nil
}
