package testutil

import (
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	rel := filepath.Join(filepath.Dir(filename), "..", "..", ".env")
	godotenv.Load(rel)
	common.LoadConfig()
}
