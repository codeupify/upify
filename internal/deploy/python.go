package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallPythonLibraries(dir string) error {
	requirementsFile := filepath.Join(dir, "requirements.txt")
	if _, err := os.Stat(requirementsFile); err == nil {
		cmd := exec.Command("pip", "install", "-r", requirementsFile, "-t", dir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install Python requirements: %v", err)
		}
	}

	return nil
}

func InstallPythonLibrary(dir string, library string) error {
	cmd := exec.Command("pip", "install", library, "-t", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %v", library, err)
	}
	fmt.Println("Successfully installed", library)
	return nil
}
