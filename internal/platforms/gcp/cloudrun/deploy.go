package cloudrun

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/deploy"

	functions "cloud.google.com/go/functions/apiv2"
	functionspb "cloud.google.com/go/functions/apiv2/functionspb"
	"cloud.google.com/go/storage"
)

func Deploy(cfg *config.Config) error {
	// TODO
	// if err := validateAWSLambdaConfig(cfg); err != nil {
	// 	return err
	// }

	tempDir, err := os.MkdirTemp("", "cloudrun_deployment_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	// defer os.RemoveAll(tempDir)

	err = deploy.CopyFilesToTempDir(".", tempDir)
	if err != nil {
		return fmt.Errorf("failed to copy files to temp directory: %v", err)
	}

	err = adjustEntryPointFile(cfg, tempDir)
	if err != nil {
		return fmt.Errorf("failed to adjust entrypoint file: %v", err)
	}

	zipPath := filepath.Join(tempDir, "source.zip")
	err = deploy.CreateZip(tempDir, zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip: %v", err)
	}

	ctx := context.Background()

	bucketName, objectName, err := uploadToStorage(cfg, ctx, tempDir)
	if err != nil {
		return fmt.Errorf("failed to upload files to storage: %v", err)
	}

	fmt.Printf("\n%s/%s\n", bucketName, objectName)

	// err = createFunction(cfg, ctx, bucketName, objectName)
	// if err != nil {
	// 	return fmt.Errorf("failed to create function: %v", err)
	// }

	return nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Errorf("failed to load config: %v", err)
	}
	if err := Deploy(cfg); err != nil {
		fmt.Errorf("failed to deploy: %v", err)
	}
}

func adjustEntryPointFile(cfg *config.Config, tempDirPath string) error {

	if cfg.Language == config.Python {
		mainPath := filepath.Join(tempDirPath, "main.py")
		_mainPath := filepath.Join(tempDirPath, "_main.py")

		if _, err := os.Stat(mainPath); err == nil {
			err := os.Rename(mainPath, _mainPath)
			if err != nil {
				return fmt.Errorf("failed to rename main.py to _main.py: %v", err)
			}
		}

		wrapperPath := filepath.Join(tempDirPath, "upify_wrapper.py")
		if _, err := os.Stat(wrapperPath); err == nil {
			content, err := os.ReadFile(wrapperPath)
			if err != nil {
				return fmt.Errorf("failed to read upify_wrapper.py: %v", err)
			}

			reImportMain := regexp.MustCompile(`(?m)^\s*import\s+main\s*$`)
			updatedContent := reImportMain.ReplaceAllString(string(content), "import _main")

			reFromMain := regexp.MustCompile(`(?m)^\s*from\s+main\s+import\s+`)
			updatedContent = reFromMain.ReplaceAllString(updatedContent, "from _main import ")

			err = os.WriteFile(wrapperPath, []byte(updatedContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to update upify_wrapper.py: %v", err)
			}
		}

		newMainPath := filepath.Join(tempDirPath, "main.py")
		err := os.Rename(wrapperPath, newMainPath)
		if err != nil {
			return fmt.Errorf("failed to rename upify_wrapper.py to main.py: %v", err)
		}
	}

	return nil
}

func uploadToStorage(cfg *config.Config, ctx context.Context, zipPath string) (string, string, error) {

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	bucketName := fmt.Sprintf("%s-%s-source", cfg.GCPCloudRun.ProjectId, cfg.Name)
	bucket := storageClient.Bucket(bucketName)

	if err := bucket.Create(ctx, cfg.GCPCloudRun.ProjectId, nil); err != nil {
		return bucketName, "", fmt.Errorf("failed to create bucket: %w", err)
	}

	file, err := os.Open(zipPath)
	if err != nil {
		return bucketName, "", fmt.Errorf("failed to open local zip file: %w", err)
	}
	defer file.Close()

	objectName := filepath.Base(zipPath)
	writer := bucket.Object(objectName).NewWriter(ctx)

	if _, err = io.Copy(writer, file); err != nil {
		return bucketName, objectName, fmt.Errorf("failed to upload zip to GCS: %w", err)
	}
	if err := writer.Close(); err != nil {
		return bucketName, objectName, fmt.Errorf("failed to finalize GCS upload: %w", err)
	}

	return bucketName, objectName, nil
}

func createFunction(cfg *config.Config, ctx context.Context, bucketName string, objectName string) error {
	client, err := functions.NewFunctionClient(ctx)
	if err != nil {
		fmt.Errorf("Failed to create client: %v", err)
	}
	defer client.Close()

	serviceConfig := &functionspb.ServiceConfig{
		TimeoutSeconds: 60,
	}

	if cfg.GCPCloudRun.Memory != 0 {
		serviceConfig.AvailableMemory = fmt.Sprintf("%dM", cfg.GCPCloudRun.Memory)
	}

	function := &functionspb.Function{
		Name: fmt.Sprintf("projects/%s/locations/%s/functions/%s", cfg.GCPCloudRun.ProjectId, cfg.GCPCloudRun.Region, cfg.Name),
		BuildConfig: &functionspb.BuildConfig{
			Runtime:    cfg.GCPCloudRun.Runtime,
			EntryPoint: "handler",
			Source: &functionspb.Source{
				Source: &functionspb.Source_StorageSource{
					StorageSource: &functionspb.StorageSource{
						Bucket: bucketName,
						Object: objectName,
					},
				},
			},
		},
		ServiceConfig: &functionspb.ServiceConfig{
			TimeoutSeconds: 60,
		},
	}

	req := &functionspb.CreateFunctionRequest{
		Parent:     fmt.Sprintf("projects/%s/locations/%s", cfg.GCPCloudRun.ProjectId, cfg.GCPCloudRun.Region),
		Function:   function,
		FunctionId: cfg.Name,
	}

	op, err := client.CreateFunction(ctx, req)
	if err != nil {
		fmt.Errorf("Failed to create function: %v", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for operation: %w", err)
	}

	fmt.Printf("Function deployed successfully: %s\n", resp.Name)

	return nil
}
