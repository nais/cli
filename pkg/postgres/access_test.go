package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsAll(t *testing.T) {
	expected := "blabla whatever ALL stuff"
	actual := setGrant("blabla whatever CHANGEME stuff", true)
	assert.Equal(t, expected, actual)
}

func TestPermissionsOther(t *testing.T) {
	expected := "blabla whatever SELECT stuff"
	actual := setGrant("blabla whatever CHANGEME stuff", false)
	assert.Equal(t, expected, actual)
}
