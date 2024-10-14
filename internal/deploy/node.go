package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/codeupify/upify/internal/config"
)

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func parsePackageJSON(path string) (*PackageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

func installPackages(dir string, pkg *PackageJSON, packageManager config.PackageManager) error {
	var installCmd *exec.Cmd
	if packageManager == config.Npm {
		installCmd = exec.Command("npm", "install", "--production")
	} else {
		installCmd = exec.Command("yarn", "install", "--production")
	}

	installCmd.Dir = dir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install Node.js dependencies: %v", err)
	}

	if _, hasBuild := pkg.Scripts["build"]; hasBuild {
		var buildCmd *exec.Cmd
		if packageManager == config.Npm {
			buildCmd = exec.Command("npm", "run", "build")
		} else {
			buildCmd = exec.Command("yarn", "build")
		}

		buildCmd.Dir = dir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build Node.js project: %v", err)
		}
		fmt.Println("Successfully built Node.js project")
	} else {
		fmt.Println("No build script found; skipping build step")
	}

	return nil
}

func build(dir string, pkg *PackageJSON, packageManager config.PackageManager) error {
	if _, hasBuild := pkg.Scripts["build"]; hasBuild {
		var buildCmd *exec.Cmd
		if packageManager == config.Npm {
			buildCmd = exec.Command("npm", "run", "build")
		} else {
			buildCmd = exec.Command("yarn", "build")
		}

		buildCmd.Dir = dir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build Node.js project: %v", err)
		}
		fmt.Println("Successfully built Node.js project")
	} else {
		fmt.Println("No build script found; skipping build step")
	}

	return nil
}
