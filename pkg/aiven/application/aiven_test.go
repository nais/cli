package application

import (
	"gopkg.in/yaml.v3"
	"gotest.tools/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func TestCreateAndMarshalOfAivenApplication(t *testing.T) {

	username := "my-user"
	team := "my-team"
	pool := "my-pool"
	expiryDate := time.Now().Format(time.RFC3339)

	wantedAiven := Aiven{
		Kind:       AivenKind,
		ApiVersion: AivenApiVersion,
		Metadata: Metadata{
			Name:      username,
			Namespace: team,
		},
		Spec: AivenSpec{
			SecretName: "",
			Protected:  DefaultProtected,
			Kafka: KafkaSpec{
				Pool: pool,
			},
			ExpiresAt: expiryDate,
		},
	}

	createdAiven := CreateAiven(username, team, pool, expiryDate)
	assert.Equal(t, createdAiven, wantedAiven)

	err := createdAiven.SetSecretName("")
	wantedSecretName, err := createdAiven.secretName()

	assert.NilError(t, err)
	assert.Equal(t, createdAiven.Spec.SecretName, wantedSecretName)

	file := temporaryFile()
	testMarshaledFile(t, file, createdAiven)
	deleteTemp(file)
}

func testMarshaledFile(t *testing.T, file *os.File, createdAiven Aiven) {
	// Write new aivenApplication to file
	err := createdAiven.MarshalAndWriteToFile(file.Name())
	assert.NilError(t, err)

	// Read file
	data, err := ioutil.ReadFile(file.Name())
	assert.NilError(t, err)

	// Unmarshall file
	unmarshalledAivenApplication := Aiven{}
	err = yaml.Unmarshal(data, &unmarshalledAivenApplication)
	assert.NilError(t, err)
	assert.Equal(t, unmarshalledAivenApplication, createdAiven)
}

func temporaryFile() *os.File {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "aiven-test-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	return tmpFile
}

func deleteTemp(fileName *os.File) {
	defer os.Remove(fileName.Name())
}
