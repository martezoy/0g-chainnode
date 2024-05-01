package types_test

import (
	"os"
	"testing"

	"github.com/0glabs/0g-chain/app"
)

func TestMain(m *testing.M) {
	app.SetSDKConfig()
	os.Exit(m.Run())
}
