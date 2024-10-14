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
	"google.golang.org/api/iterator"
	"google.golang.org/api/run/v1"

	functions "cloud.google.com/go/functions/apiv2"
	functionspb "cloud.google.com/go/functions/apiv2/functionspb"
	"cloud.google.com/go/storage"
)

func Deploy(cfg *config.Config) error {
	// TODO
	// if err := validateAWSLambdaConfig(cfg); err != nil {
	// 	return err
	// }

	envVars, err := deploy.LoadEnvVariables()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %v", err)
	}

	envVars["UPIFY_DEPLOY_PLATFORM"] = "gcp-cloudrun"

	tempDir, err := os.MkdirTemp("", "cloudrun_deployment_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	bucketName := getBucketName(cfg)

	bucketName, objectName, err := uploadToStorage(ctx, cfg.GCPCloudRun.ProjectId, bucketName, zipPath)
	if err != nil {
		return fmt.Errorf("failed to upload files to storage: %v", err)
	}

	fmt.Print("Created bucket " + bucketName + " and object " + objectName + " in project " + cfg.GCPCloudRun.ProjectId + "\n")

	err = createFunction(cfg, ctx, bucketName, objectName, envVars)
	if err != nil {
		return fmt.Errorf("failed to create function: %v", err)
	}

	deleteBucket(ctx, bucketName)

	return nil
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

func getBucketName(cfg *config.Config) string {
	return fmt.Sprintf("upify-%s-%s-source", cfg.GCPCloudRun.ProjectId, cfg.Name)
}

func deleteBucket(ctx context.Context, bucketName string) error {
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	bucket := storageClient.Bucket(bucketName)

	it := bucket.Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == storage.ErrBucketNotExist {
			return fmt.Errorf("bucket %q not found", bucketName)
		}
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating through objects: %v", err)
		}
		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("error deleting object %q: %v", attrs.Name, err)
		}
	}

	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("error deleting bucket %q: %v", bucketName, err)
	}

	fmt.Printf("Bucket %q successfully deleted\n", bucketName)
	return nil
}

func uploadToStorage(ctx context.Context, projectId string, bucketName string, zipPath string) (string, string, error) {

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	bucket := storageClient.Bucket(bucketName)

	// Delete any existing buckets
	_, err = bucket.Attrs(ctx)
	if err == nil {
		fmt.Printf("Bucket %q already exists, attempting to delete\n", bucketName)
		if err := deleteBucket(ctx, bucketName); err != nil {
			return "", "", fmt.Errorf("failed to delete existing bucket contents: %w", err)
		}
	} else if err != storage.ErrBucketNotExist {
		return "", "", fmt.Errorf("error checking bucket existence: %w", err)
	}

	if err := bucket.Create(ctx, projectId, nil); err != nil {
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

	fmt.Printf("Uploaded zip to bucket %q, object %q\n", bucketName, objectName)
	return bucketName, objectName, nil
}

func setIAMPolicyForUnauthenticated(runService *run.APIService, serviceName string) error {
	policy, err := runService.Projects.Locations.Services.GetIamPolicy(serviceName).Do()
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %v", err)
	}

	policy.Bindings = append(policy.Bindings, &run.Binding{
		Role:    "roles/run.invoker",
		Members: []string{"allUsers"},
	})

	_, err = runService.Projects.Locations.Services.SetIamPolicy(serviceName, &run.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %v", err)
	}

	fmt.Println("Function set to allow unauthenticated invocations")
	return nil
}

func createFunction(cfg *config.Config, ctx context.Context, bucketName string, objectName string, envVariables map[string]string) error {

	fmt.Printf("Creating function %q in project %q\n", cfg.Name, cfg.GCPCloudRun.ProjectId)

	client, err := functions.NewFunctionClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	serviceConfig := &functionspb.ServiceConfig{
		TimeoutSeconds:       60,
		EnvironmentVariables: envVariables,
	}

	if cfg.GCPCloudRun.Memory != 0 {
		serviceConfig.AvailableMemory = fmt.Sprintf("%dM", cfg.GCPCloudRun.Memory)
	}

	functionName := fmt.Sprintf("projects/%s/locations/%s/functions/%s", cfg.GCPCloudRun.ProjectId, cfg.GCPCloudRun.Region, cfg.Name)

	function := &functionspb.Function{
		Name: functionName,
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
		ServiceConfig: serviceConfig,
	}

	var resp *functionspb.Function
	var isNewFunction bool

	_, err = client.GetFunction(ctx, &functionspb.GetFunctionRequest{Name: functionName})
	if err != nil {

		req := &functionspb.CreateFunctionRequest{
			Parent:     fmt.Sprintf("projects/%s/locations/%s", cfg.GCPCloudRun.ProjectId, cfg.GCPCloudRun.Region),
			Function:   function,
			FunctionId: cfg.Name,
		}

		op, err := client.CreateFunction(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create function: %v", err)
		}

		resp, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for operation: %w", err)
		}
		isNewFunction = true
	} else {
		fmt.Print("Function already exists, updating...")
		req := &functionspb.UpdateFunctionRequest{
			Function: function,
		}
		op, err := client.UpdateFunction(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update function: %v", err)
		}

		resp, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for operation: %w", err)
		}
		isNewFunction = false
	}

	if err != nil {
		return fmt.Errorf("failed to create/update function: %v", err)
	}

	serviceName := fmt.Sprintf("projects/%s/locations/%s/services/%s", cfg.GCPCloudRun.ProjectId, cfg.GCPCloudRun.Region, cfg.Name)
	runService, err := run.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	err = setIAMPolicyForUnauthenticated(runService, serviceName)
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %v", err)
	}

	service, err := runService.Projects.Locations.Services.Get(serviceName).Do()
	if err != nil {
		return fmt.Errorf("failed to get Cloud Run service details: %v", err)
	}
	fmt.Printf("Function deployed successfully: %s\n", resp.Name)

	if isNewFunction {
		fmt.Printf("\nNew Function URL: %s\n\n", service.Status.Url)
	} else {
		fmt.Printf("\nExisting Function URL: %s\n\n", service.Status.Url)
	}

	return nil
}
