package postgres

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsAll(t *testing.T) {
	expected := "blabla whatever ALL stuff"
	actual := fmt.Sprintf("blabla whatever %s stuff", "ALL")
	assert.Equal(t, expected, actual)
}

func TestPermissionsOther(t *testing.T) {
	expected := "blabla whatever SELECT stuff"
	actual := fmt.Sprintf("blabla whatever %s stuff", "SELECT")
	assert.Equal(t, expected, actual)
}
