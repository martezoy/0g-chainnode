package util

import (
	"path/filepath"
)

// ZgChainHomePath returns the OS-specific filepath for the 0g-chain home directory
// Assumes network is running with kvtool installed from the sub-repository in tests/e2e/kvtool
func ZgChainHomePath() string {
	return filepath.Join("kvtool", "full_configs", "generated", "0gchaind", "initstate", ".0gchain")
}
