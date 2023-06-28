package appstarter

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	JvmGradle    = "JVM_GRADLE"
	JvmMaven     = "JVM_MAVEN"
	NodeJS       = "NODEJS"
	GoMake       = "GO_MAKE"
	PythonPoetry = "PYTHON_POETRY"
	PythonPip    = "PYTHON_PIP"
)

type StartNaisIoRequest struct {
	AppName       string   `json:"appName"`
	Team          string   `json:"team"`
	Platform      string   `json:"platform"`
	AppListenPort uint     `json:"appListenPort,omitempty"`
	Extras        []string `json:"extras,omitempty"`
	KafkaTopics   []string `json:"kafkaTopics,omitempty"`
}

var projectTypes = func() map[string]string {
	return map[string]string{
		"build.gradle":     JvmGradle,
		"build.gradle.kts": JvmGradle,
		"pom.xml":          JvmMaven,
		"package.json":     NodeJS,
		"Makefile":         GoMake,
		"requirements.txt": PythonPip,
		"poetry.lock":      PythonPoetry,
	}
}

var currentDir, _ = os.Getwd()

func Naisify(appName string, team string, extras []string, kafkaTopics []string, appPort uint) error {
	appType, err := determinePlatform()
	if err != nil || len(appType) == 0 {
		return fmt.Errorf("unable to determine app type in %s: %v", currentDir, err)
	}
	request := StartNaisIoRequest{
		AppName:       appName,
		Team:          team,
		Platform:      appType,
		Extras:        extras,
		KafkaTopics:   kafkaTopics,
		AppListenPort: appPort,
	}
	startNaisIoResponse, err := makeHttpRequest(&request)
	if err != nil {
		return fmt.Errorf("error while requesting config: %v", err)
	}
	currentDir, _ := os.Getwd()
	err = writeTo(currentDir, startNaisIoResponse)
	if err != nil {
		return fmt.Errorf("error while writing to disk: %v", err)
	}
	fmt.Printf("%d files created, please review them for any todo items\n", len(startNaisIoResponse))
	return nil
}

func determinePlatform() (string, error) {
	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		return "", fmt.Errorf("error reading directory contents: %v", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileType := projectTypes()[file.Name()]
		if fileType != "" {
			return fileType, nil
		}
	}
	return "", fmt.Errorf("no known build system files (such as pom.xml et al.) found")
}

func writeTo(baseDir string, startNaisIoResponse map[string]string) error {
	for filename, contents := range startNaisIoResponse {
		if strings.Contains(filename, "..") {
			return fmt.Errorf("%s looks funky, may be path traversal", filename)
		}
		absoluteFilePath := baseDir + string(os.PathSeparator) + filename
		dir, _ := filepath.Split(absoluteFilePath)
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return fmt.Errorf("unable to create dir %s: %v", dir, err)
		}
		err = os.WriteFile(absoluteFilePath, []byte(contents), 0744)
		if err != nil {
			return fmt.Errorf("error while writing file: %v", err)
		}
	}
	return nil
}

func makeHttpRequest(req *StartNaisIoRequest) (map[string]string, error) {
	postBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}
	httpReq, _ := http.NewRequest("POST", "https://start.nais.io/app", bytes.NewBuffer(postBody))
	httpReq.Header = http.Header{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json"},
	}
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("error while making http request: %v %d %s", err, resp.StatusCode, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %v", err)
	}
	return parse(b)
}

func parse(responseBody []byte) (map[string]string, error) {
	var parsed map[string]string
	err := json.Unmarshal(responseBody, &parsed)
	if err != nil {
		return nil, fmt.Errorf("error while parsing http response: %v", err)
	}
	for key, value := range parsed {
		b64Decoded, err := b64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("error while decoding b64: %v", err)
		}
		parsed[key] = string(b64Decoded)
	}
	return parsed, nil
}
