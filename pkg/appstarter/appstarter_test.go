package appstarter

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestProjectTypeIsRecognized(t *testing.T) {
	projectTypes := map[string]string{
		"build.gradle":     JvmGradle,
		"build.gradle.kts": JvmGradle,
		"pom.xml":          JvmMaven,
		"package.json":     NodeJS,
		"Makefile":         GoMake,
		"requirements.txt": PythonPip,
		"poetry.lock":      PythonPoetry,
	}
	for filename, projectType := range projectTypes {
		path := createTempFile(filename)
		expected := projectTypes[projectType]
		actual, _ := determinePlatform()
		assert.Equal(t, expected, actual, "Project type matches")
		_ = os.Remove(path)
	}
}

func TestResponseIsWrittenToDir(t *testing.T) {
	tmp := os.TempDir()
	response := map[string]string{
		"a": "a contents",
		"b": "b contents",
	}
	_ = writeTo(tmp, response)
	for filename := range response {
		absolutePath := tmp + string(os.PathSeparator) + filename
		if !fileExists(absolutePath) {
			t.Errorf("%s should exist, but doesn't", absolutePath)
		}
		_ = os.Remove(absolutePath)
	}
}

func createTempFile(name string) string {
	tmp := os.TempDir()
	path := tmp + string(os.PathSeparator) + name
	_, _ = os.Create(path)
	return path
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
