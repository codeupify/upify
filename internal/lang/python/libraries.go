package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallRequirements(dir string) error {
	requirementsFile := filepath.Join(dir, "requirements.txt")

	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		fmt.Println("No requirements.txt found; skipping installation...")
		return nil
	}

	cmd := exec.Command("pip", "install", "-r", requirementsFile, "-t", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Python requirements: %v", err)
	}

	return nil
}

func InstallLibrary(dir string, library string) error {
	installed, err := isLibraryInstalled(dir, library)
	if err != nil {
		return fmt.Errorf("failed to check if library is installed: %v", err)
	}

	if installed {
		fmt.Printf("Library %s is already installed. Skipping installation.\n", library)
		return nil
	}

	cmd := exec.Command("pip", "install", library, "-t", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %v", library, err)
	}
	return nil
}

func isLibraryInstalled(dir string, library string) (bool, error) {
	packageDir := filepath.Join(dir, library)

	if _, err := os.Stat(packageDir); !os.IsNotExist(err) {
		return true, nil
	}

	eggInfoDir := filepath.Join(dir, library+".egg-info")
	if _, err := os.Stat(eggInfoDir); !os.IsNotExist(err) {
		return true, nil
	}

	return false, nil
}
