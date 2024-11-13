package infra

import (
	"path/filepath"

	"github.com/codeupify/upify/internal/platform"
)

func GetPlatformTerraformDir(platform platform.Platform) string {
	return filepath.Join(".upify", "environments", "prod", string(platform))
}
