package upgrade

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadUpgradeParams tests the LoadUpgradeParams function
func TestLoadUpgradeParams(t *testing.T) {
	_, err := os.Getwd()
	require.NoError(t, err, "can't get current working directory")

	_, err = RetrieveUpgradesList(upgradesPath)
	require.Error(t, err)
}
