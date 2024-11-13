package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/fs"
	"github.com/codeupify/upify/internal/infra"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/lang/node"
	"github.com/codeupify/upify/internal/lang/python"
	"github.com/codeupify/upify/internal/platform"
)

func Deploy(cfg *config.Config) error {
	if err := infra.PreDeployValidate(cfg, platform.AWS); err != nil {
		return err
	}

	if err := infra.WriteEnvironmentVariables(platform.AWS); err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "lambda_deployment_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = fs.CopyFilesToTempDir(".", tempDir)
	if err != nil {
		return fmt.Errorf("failed to copy files to temp directory: %v", err)
	}

	err = installRequirements(tempDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to install requirements: %v", err)
	}

	zipPath := filepath.Join(tempDir, "source.zip")
	fmt.Printf("Creating %s...\n", zipPath)
	err = fs.CreateZip(tempDir, zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip: %v", err)
	}

	terraformManager, err := infra.NewTerraformManager(infra.GetPlatformTerraformDir(platform.AWS))
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

func installRequirements(dir string, cfg *config.Config) error {
	switch cfg.Language {
	case lang.Python:
		err := python.InstallRequirements(dir)
		if err != nil {
			return err
		}

		err = python.InstallLibrary(dir, "flask")
		if err != nil {
			return err
		}

		err = python.InstallLibrary(dir, "apig-wsgi")
		if err != nil {
			return err
		}

	case lang.JavaScript, lang.TypeScript:
		err := node.InstallPackagesJSON(dir, cfg.PackageManager)
		if err != nil {
			return err
		}

		err = node.InstallPackage(dir, "express", cfg.PackageManager)
		if err != nil {
			return err
		}

		err = node.InstallPackage(dir, "serverless-http", cfg.PackageManager)
		if err != nil {
			return err
		}

		pkgJson, err := node.ParsePackageJSON(filepath.Join(dir, "package.json"))
		if err != nil {
			return fmt.Errorf("failed to parse package.json: %v", err)
		}

		node.Build(dir, pkgJson, cfg.PackageManager)
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	return nil
}
