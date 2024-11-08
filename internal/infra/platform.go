package infra

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeupify/upify/internal/platform"
)

func AddPlatform(platform platform.Platform, environmentsMainContent string, modulesMainContent string) error {
	environmentsDir := filepath.Join(".upify", "environments", "prod", string(platform))
	if _, err := os.Stat(environmentsDir); err == nil {
		return fmt.Errorf("%s already exists", environmentsDir)
	}

	if err := os.MkdirAll(environmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create environments directory: %w", err)
	}

	envMainPath := filepath.Join(environmentsDir, "main.tf")
	fmt.Printf("Creating %s...\n", envMainPath)
	if err := os.WriteFile(envMainPath, []byte(environmentsMainContent), 0644); err != nil {
		return fmt.Errorf("failed to write environment main.tf: %w", err)
	}

	modulesDir := filepath.Join(".upify", "modules", string(platform))
	modulesDirCreated := false
	_, err := os.Stat(modulesDir)
	if err == nil {
		fmt.Printf("Module directory %s exists, reusing existing module\n", modulesDir)
	} else {
		if err := os.MkdirAll(modulesDir, 0755); err != nil {
			return fmt.Errorf("failed to create modules directory: %w", err)
		}

		moduleMainPath := filepath.Join(modulesDir, "main.tf")
		fmt.Printf("Creating %s...\n", moduleMainPath)
		if err := os.WriteFile(moduleMainPath, []byte(modulesMainContent), 0644); err != nil {
			return fmt.Errorf("failed to write module main.tf: %w", err)
		}

		modulesDirCreated = true
	}

	terraformManager, err := NewTerraformManager(environmentsDir)
	if err != nil {
		cleanUp(environmentsDir, modulesDir, modulesDirCreated)
		return fmt.Errorf("failed to create terraform manager: %v", err)
	}

	ctx := context.Background()
	if err := terraformManager.Init(ctx); err != nil {
		cleanUp(environmentsDir, modulesDir, modulesDirCreated)
		return fmt.Errorf("failed to initialize terraform: %v", err)
	}

	return nil
}

func ListPlatforms() []string {
	environmentsDir := filepath.Join(".upify", "environments")
	result := []string{}

	for _, platform := range platform.AllPlatforms {
		platformDir := filepath.Join(environmentsDir, "prod", string(platform))
		if _, err := os.Stat(platformDir); err == nil {
			result = append(result, string(platform))
		}
	}

	return result
}

func cleanUp(environmentsDir string, modulesDir string, modulesDirCreated bool) {
	fmt.Println("Removing environments directory...")
	if err := os.RemoveAll(environmentsDir); err != nil {
		fmt.Printf("Failed to remove environments directory %s: %v\n", environmentsDir, err)
	}

	if modulesDirCreated {
		fmt.Println("Removing modules directory...")
		if err := os.RemoveAll(modulesDir); err != nil {
			fmt.Printf("Failed to remove modules directory %s: %v\n", modulesDir, err)
		}
	}
}
