package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The client create/update/delete commands are deliberately stubbed because
// the Canvus REST API exposes no corresponding endpoints. These tests pin
// that contract: the commands are registered, hidden, and return a clear
// "not supported" error when invoked.

func TestClientCommands_AreHiddenStubs(t *testing.T) {
	cases := []struct {
		name string
	}{
		{"create"},
		{"update"},
		{"delete"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, _, err := ClientCmd.Find([]string{tc.name})
			require.NoError(t, err, "%q should be registered under client", tc.name)
			assert.True(t, cmd.Hidden, "%q should be hidden — it is unsupported", tc.name)
		})
	}
}

func TestClientCreate_ReturnsUnsupportedError(t *testing.T) {
	err := createCmd.RunE(createCmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestClientUpdate_ReturnsUnsupportedError(t *testing.T) {
	err := updateCmd.RunE(updateCmd, []string{"client-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestClientDelete_ReturnsUnsupportedError(t *testing.T) {
	err := deleteCmd.RunE(deleteCmd, []string{"client-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}
