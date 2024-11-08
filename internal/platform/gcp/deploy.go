package gcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/fs"
	"github.com/codeupify/upify/internal/infra"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/lang/node"
	"github.com/codeupify/upify/internal/platform"
)

func Deploy(cfg *config.Config) error {
	if err := infra.PreDeployValidate(cfg, platform.GCP); err != nil {
		return err
	}

	if err := infra.WriteEnvironmentVariables(platform.GCP); err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "cloudrun_deployment_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = fs.CopyFilesToTempDir(".", tempDir)
	if err != nil {
		return fmt.Errorf("failed to copy files to temp directory: %v", err)
	}

	err = adjustEntryPointFile(cfg, tempDir)
	if err != nil {
		return fmt.Errorf("failed to adjust entrypoint file: %v", err)
	}

	if cfg.Language == lang.JavaScript || cfg.Language == lang.TypeScript {
		err = updatePackageJson(cfg, tempDir)
		if err != nil {
			return fmt.Errorf("failed to update package.json: %v", err)
		}
	}

	zipPath := filepath.Join(tempDir, "source.zip")
	fmt.Printf("Creating %s...\n", zipPath)
	err = fs.CreateZip(tempDir, zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip: %v", err)
	}

	terraformManager, err := infra.NewTerraformManager(infra.GetPlatformTerraformDir(platform.GCP))
	if err != nil {
		return fmt.Errorf("failed to create terraform manager: %v", err)
	}

	vars := map[string]string{
		"source_zip_path": zipPath,
	}

	ctx := context.Background()
	if err := terraformManager.Apply(ctx, vars); err != nil {
		return err
	}

	return nil
}

func updatePackageJson(cfg *config.Config, tempDirPath string) error {
	pkgJson, err := node.ParsePackageJSON(filepath.Join(tempDirPath, "package.json"))
	if err != nil {
		return fmt.Errorf("failed to parse package.json: %v", err)
	}

	node.SetMainInPackageJSON(pkgJson, "upify_handler.js")

	node.AddPackageToPackageJSON(pkgJson, "@google-cloud/functions-framework", "^3.0.0")
	if pkgJson.Scripts != nil && pkgJson.Scripts["build"] != "" {
		buildCommand := "npm run build"
		if cfg.PackageManager == lang.Yarn {
			buildCommand = "yarn build"
		}

		node.AddScriptToPackageJSON(pkgJson, "gcp-build", buildCommand)
	}

	return node.WritePackageJSON(filepath.Join(tempDirPath, "package.json"), pkgJson)
}

func adjustEntryPointFile(cfg *config.Config, tempDirPath string) error {
	switch cfg.Language {
	case lang.Python:
		return adjustPythonEntryPointFile(tempDirPath)
	default:
		return nil
	}
}

// We have to do this because Cloud Run expects the entrypoint to be main.py,
// but we need our upify_handler.py to be the entrypoint. So we rename any
// existing main.py to _main.py and then rename upify_handler.py to main.py
func adjustPythonEntryPointFile(tempDirPath string) error {
	mainPath := filepath.Join(tempDirPath, "main.py")
	_mainPath := filepath.Join(tempDirPath, "_main.py")

	if _, err := os.Stat(mainPath); err == nil {
		err := os.Rename(mainPath, _mainPath)
		if err != nil {
			return fmt.Errorf("failed to rename main.py to _main.py: %v", err)
		}
	}

	wrapperFiles := []string{"upify_handler.py", "upify_main.py"}

	for _, wrapperFile := range wrapperFiles {
		wrapperPath := filepath.Join(tempDirPath, wrapperFile)
		if _, err := os.Stat(wrapperPath); err == nil {
			content, err := os.ReadFile(wrapperPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %v", wrapperFile, err)
			}

			reImportMain := regexp.MustCompile(`(?m)^\s*import\s+main\s*$`)
			updatedContent := reImportMain.ReplaceAllString(string(content), "import _main")

			reFromMain := regexp.MustCompile(`(?m)^\s*from\s+main\s+import\s+`)
			updatedContent = reFromMain.ReplaceAllString(updatedContent, "from _main import ")

			err = os.WriteFile(wrapperPath, []byte(updatedContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to update %s: %v", wrapperFile, err)
			}
		}
	}

	upifyWrapperPath := filepath.Join(tempDirPath, "upify_handler.py")
	newMainPath := filepath.Join(tempDirPath, "main.py")
	err := os.Rename(upifyWrapperPath, newMainPath)
	if err != nil {
		return fmt.Errorf("failed to rename upify_handler.py to main.py: %v", err)
	}

	return nil
}
