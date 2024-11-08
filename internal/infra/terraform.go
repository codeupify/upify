package infra

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type TerraformManager struct {
	tf      *tfexec.Terraform
	workDir string
}

func NewTerraformManager(workDir string) (*TerraformManager, error) {
	ctx := context.Background()

	// Try to get it from PATH

	tfPath, err := exec.LookPath("terraform")
	if err != nil && runtime.GOOS == "windows" {
		tfPath, err = exec.LookPath("terraform.exe")
	}

	if err == nil {
		tf, err := tfexec.NewTerraform(workDir, tfPath)
		if err != nil {
			return nil, fmt.Errorf("error creating terraform executor: %w", err)
		}

		v, _, err := tf.Version(ctx, false)
		if err != nil {
			return nil, err
		}

		constraint, _ := version.NewConstraint(">= 1.0.0")
		if !constraint.Check(v) {
			return nil, fmt.Errorf("terraform version %s is too old, please upgrade to 1.0.0 or newer", v)
		}

		tf.SetStdout(os.Stdout)
		tf.SetStderr(os.Stderr)
		return &TerraformManager{tf: tf, workDir: workDir}, nil
	}

	// Try to get it from ~/.upify
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to determine home directory: %w", err)
	}
	customDir := filepath.Join(homeDir, ".upify")
	customExecPath := filepath.Join(customDir, "terraform")
	if runtime.GOOS == "windows" {
		customExecPath += ".exe"
	}

	if _, err := os.Stat(customExecPath); err == nil {
		fmt.Printf("Using existing Terraform binary at: %s\n", customExecPath)
		tf, err := tfexec.NewTerraform(workDir, customExecPath)
		if err != nil {
			return nil, fmt.Errorf("error creating terraform executor: %w", err)
		}

		tf.SetStdout(os.Stdout)
		tf.SetStderr(os.Stderr)
		return &TerraformManager{tf: tf, workDir: workDir}, nil
	}

	fmt.Println("Terraform not found in PATH or ~/.upify, installing...")

	if err := os.MkdirAll(customDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", customDir, err)
	}

	installer := install.NewInstaller()
	defer installer.Remove(ctx)

	execPath, err := installer.Install(ctx, []src.Installable{
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion("1.9.8")),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error installing terraform: %w", err)
	}

	if err := os.Rename(execPath, customExecPath); err != nil {
		return nil, fmt.Errorf("failed to move terraform binary to %s: %w", customExecPath, err)
	}

	fmt.Printf("Terraform successfully installed at: %s\n", customExecPath)

	tf, err := tfexec.NewTerraform(workDir, customExecPath)
	if err != nil {
		return nil, fmt.Errorf("error creating terraform executor: %w", err)
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)
	return &TerraformManager{tf: tf, workDir: workDir}, nil
}

func (m *TerraformManager) Init(ctx context.Context) error {
	return m.tf.Init(ctx)
}

func (m *TerraformManager) Plan(ctx context.Context, vars map[string]string) (bool, error) {
	return m.tf.Plan(ctx)
}

func (m *TerraformManager) Apply(ctx context.Context, vars map[string]string) error {
	var applyOpts []tfexec.ApplyOption
	for key, value := range vars {
		applyOpts = append(applyOpts, tfexec.Var(fmt.Sprintf("%s=%s", key, value)))
	}

	return m.tf.Apply(ctx, applyOpts...)
}

func (m *TerraformManager) Output(ctx context.Context) (map[string]tfexec.OutputMeta, error) {
	return m.tf.Output(ctx)
}
