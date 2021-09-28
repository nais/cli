package aiven

import (
	"github.com/nais/nais-cli/cmd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAivenTidy(t *testing.T) {
	_, err := cmd.DefaultDestination("")
	assert.NoError(t, err, "Creating folder")

	// created folders are found
	folders, err := findFoldersToTidy()
	assert.True(t, len(folders) > 0)
	assert.NoError(t, err, "Folders found")

	// created folders id tidy
	err = Tidy(folders)
	folders, err = findFoldersToTidy()
	assert.NoError(t, err, "Folders found")
	assert.True(t, len(folders) == 0)
	assert.NoError(t, err, "Tidy")
}
