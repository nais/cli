package kubectl

import (
	"fmt"
	"os/exec"
	"time"
)

const (
	Kubectl = "kubectl"
)

func Apply(aivenYamlPath string) ([]byte, error) {
	apply := "apply"
	command := "-f"

	cmd := exec.Command(Kubectl, apply, command, aivenYamlPath)
	stdout, err := hasError(cmd)
	if err != nil {
		return nil, fmt.Errorf("apply failed: %s", err)
	}
	fmt.Printf("applied --> %s", stdout)
	return stdout, nil
}

func GetSecret(secretName string) ([]byte, error) {
	// Could be removed, secret creation can have latency?
	time.Sleep(2 * time.Second)
	get := "get"
	secret := "secret"
	command := "-oyaml"

	cmd := exec.Command(Kubectl, get, secret, secretName, command)
	stdout, err := hasError(cmd)
	if err != nil {
		return nil, fmt.Errorf("get secret failed: %s", err)
	}
	fmt.Sprintln("fetched secret from namespace successfully.")
	return stdout, nil
}

func hasError(command *exec.Cmd) ([]byte, error) {
	stdout, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command: %s\n%s", command.String(), string(stdout))
	}
	return stdout, nil
}
