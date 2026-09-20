package workspace

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOpenRequiresExactlyOneExplicitSelector(t *testing.T) {
	names := []string{"index", "name", "user"}
	original := map[string]bool{}
	for _, name := range names {
		original[name] = openCmd.Flags().Lookup(name).Changed
	}
	defer func() {
		for _, name := range names {
			openCmd.Flags().Lookup(name).Changed = original[name]
		}
	}()
	// Exercise Cobra's public flag-group validation without running a live command.
	for _, selected := range [][]string{{}, {"index"}, {"name"}, {"user"}, {"index", "name"}} {
		for _, name := range names {
			openCmd.Flags().Lookup(name).Changed = false
		}
		for _, name := range selected {
			openCmd.Flags().Lookup(name).Changed = true
		}
		err := openCmd.ValidateFlagGroups()
		if len(selected) == 1 {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
