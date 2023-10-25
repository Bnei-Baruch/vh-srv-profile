package testutil

import (
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	rel := filepath.Join(filepath.Dir(filename), "..", "..", ".env")
	godotenv.Load(rel)
}
