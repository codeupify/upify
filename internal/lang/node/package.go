package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/codeupify/upify/internal/lang"
)

type PackageJSON struct {
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	Main         string            `json:"main,omitempty"`
	Other        map[string]interface{}
}

func ParsePackageJSON(path string) (*PackageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	pkg := &PackageJSON{
		Scripts:      make(map[string]string),
		Dependencies: make(map[string]string),
		Other:        make(map[string]interface{}),
	}

	if scripts, ok := temp["scripts"].(map[string]interface{}); ok {
		for k, v := range scripts {
			pkg.Scripts[k] = fmt.Sprint(v)
		}
	}

	if deps, ok := temp["dependencies"].(map[string]interface{}); ok {
		for k, v := range deps {
			pkg.Dependencies[k] = fmt.Sprint(v)
		}
	}

	if main, ok := temp["main"].(string); ok {
		pkg.Main = main
	}

	for k, v := range temp {
		if k != "scripts" && k != "dependencies" {
			pkg.Other[k] = v
		}
	}

	return pkg, nil
}

func WritePackageJSON(path string, pkg *PackageJSON) error {
	temp := pkg.Other
	temp["scripts"] = pkg.Scripts
	temp["dependencies"] = pkg.Dependencies
	if pkg.Main != "" {
		temp["main"] = pkg.Main
	}

	data, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func InstallPackagesJSON(dir string, packageManager lang.PackageManager) error {
	fmt.Printf("Installing package.json dependencies...\n")
	var installCmd *exec.Cmd
	if packageManager == lang.Npm {
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

	return nil
}

func InstallPackage(dir string, packageName string, packageManager lang.PackageManager) error {
	installed := isPackageInstalled(dir, packageName, packageManager)
	if installed {
		return nil
	}

	fmt.Printf("Installing package: %s...\n", packageName)
	var installCmd *exec.Cmd
	if packageManager == lang.Npm {
		installCmd = exec.Command("npm", "install", packageName, "--save")
	} else {
		installCmd = exec.Command("yarn", "add", packageName, "--save")
	}

	installCmd.Dir = dir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install Node.js dependency: %v", err)
	}

	return nil
}

func Build(dir string, pkg *PackageJSON, packageManager lang.PackageManager) error {
	if _, hasBuild := pkg.Scripts["build"]; hasBuild {
		fmt.Println("Building Node.js project...")
		var buildCmd *exec.Cmd
		if packageManager == lang.Npm {
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

func AddPackageToPackageJSON(pkg *PackageJSON, packageName string, version string) {
	if pkg.Dependencies == nil {
		pkg.Dependencies = make(map[string]string)
	}
	pkg.Dependencies[packageName] = version
}

func AddScriptToPackageJSON(pkg *PackageJSON, scriptName string, scriptValue string) {
	if pkg.Scripts == nil {
		pkg.Scripts = make(map[string]string)
	}
	pkg.Scripts[scriptName] = scriptValue
}

func SetMainInPackageJSON(pkg *PackageJSON, mainFile string) {
	pkg.Main = mainFile
}

func isPackageInstalled(dir, packageName string, packageManager lang.PackageManager) bool {
	var listCmd *exec.Cmd

	if packageManager == lang.Npm {
		listCmd = exec.Command("npm", "list", packageName, "--depth=0")
	} else {
		listCmd = exec.Command("yarn", "list", "--pattern", packageName)
	}

	listCmd.Dir = dir
	listCmd.Stdout = nil
	listCmd.Stderr = nil

	if err := listCmd.Run(); err == nil {

		return true
	}

	return false
}
