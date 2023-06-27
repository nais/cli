package aiven

import (
	"github.com/nais/cli/pkg/aiven/secret"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAivenTidy(t *testing.T) {
	_, err := secret.CreateDefaultDestination()
	assert.NoError(t, err, "Creating folder")

	// created folders are found
	folders, err := findFoldersToRemove()
	assert.True(t, len(folders) > 0)
	assert.NoError(t, err, "Folders found")

	// created folders id tidy
	err = tidy(folders)
	folders, err = findFoldersToRemove()
	assert.NoError(t, err, "Folders found")
	assert.True(t, len(folders) == 0)
	assert.NoError(t, err, "tidy")
}
