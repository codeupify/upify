package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func installPythonRequirements(dir string) error {
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
