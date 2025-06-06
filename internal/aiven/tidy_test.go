package aiven

import (
	"testing"

	"github.com/nais/cli/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestAivenTidy(t *testing.T) {
	_, err := createDefaultDestination()
	assert.NoError(t, err, "Creating folder")

	// created folders are found
	folders, err := findFoldersToRemove()
	assert.True(t, len(folders) > 0)
	assert.NoError(t, err, "Folders found")

	// created folders id tidy
	err = tidy(folders, cli.Stdout())
	assert.NoError(t, err)
	folders, err = findFoldersToRemove()
	assert.NoError(t, err, "Folders found")
	assert.True(t, len(folders) == 0)
	assert.NoError(t, err, "tidy")
}
