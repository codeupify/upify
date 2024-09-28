package lambda

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/fs"
	"github.com/codeupify/upify/internal/handler"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/lang/node"
	"github.com/codeupify/upify/internal/lang/python"
)

func Deploy(cfg *config.Config) error {
	if err := validateAWSLambdaConfig(cfg); err != nil {
		return err
	}

	handlerPath := handler.GetHandlerPath(cfg.Language)
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		return fmt.Errorf("%s not found in current working directory", handler.GetHandlerFileName(cfg.Language))
	}

	envVars, err := fs.LoadEnvVariables()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %v", err)
	}

	envVars["UPIFY_DEPLOY_PLATFORM"] = "aws-lambda"

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
	err = fs.CreateZip(tempDir, zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip: %v", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSLambda.Region),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	err = getOrCreateLambda(sess, cfg, zipPath, envVars)
	if err != nil {
		return fmt.Errorf("failed to get or create Lambda: %v", err)
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

func getOrCreateRole(svc *iam.IAM, roleName string) (string, error) {
	getRoleInput := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	if result, err := svc.GetRole(getRoleInput); err == nil {
		return *result.Role.Arn, nil
	}

	assumeRolePolicyDocument := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`

	createRoleInput := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDocument),
		RoleName:                 aws.String(roleName),
	}

	result, err := svc.CreateRole(createRoleInput)
	if err != nil {
		return "", err
	}

	attachPolicyInput := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		RoleName:  aws.String(roleName),
	}
	_, err = svc.AttachRolePolicy(attachPolicyInput)
	if err != nil {
		return "", err
	}

	fmt.Println("Waiting for role to be fully ready...")
	time.Sleep(15 * time.Second)

	return *result.Role.Arn, nil
}

func getOrCreateLambda(sess *session.Session, cfg *config.Config, zipPath string, envVariables map[string]string) error {
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipFile.Close()

	zipBytes, err := io.ReadAll(zipFile)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %v", err)
	}

	lambdaSvc := lambda.New(sess)
	iamSvc := iam.New(sess)

	awsEnvVariables := make(map[string]*string)
	for k, v := range envVariables {
		awsEnvVariables[k] = aws.String(v)
	}

	_, err = lambdaSvc.GetFunction(&lambda.GetFunctionInput{
		FunctionName: aws.String(cfg.Name),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == lambda.ErrCodeResourceNotFoundException {
			roleArn, err := getOrCreateRole(iamSvc, cfg.AWSLambda.RoleName)
			if err != nil {
				return fmt.Errorf("failed to get or create IAM role: %v", err)
			}

			fmt.Println("Creating Lambda function...")
			_, err = lambdaSvc.CreateFunction(&lambda.CreateFunctionInput{
				Code: &lambda.FunctionCode{
					ZipFile: zipBytes,
				},
				FunctionName: aws.String(cfg.Name),
				Handler:      aws.String("upify_handler.handler"),
				Role:         aws.String(roleArn),
				Runtime:      aws.String(cfg.AWSLambda.Runtime),
				Environment: &lambda.Environment{
					Variables: awsEnvVariables,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create function: %v", err)
			}
		} else {
			return fmt.Errorf("failed to check if function exists: %v", err)
		}
	} else {
		fmt.Println("Updating Lambda function...")
		_, err = lambdaSvc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(cfg.Name),
			ZipFile:      zipBytes,
		})
		if err != nil {
			return fmt.Errorf("failed to update function code: %v", err)
		}

		// Need to wait for function to be updated
		waiterCfg := request.WithWaiterMaxAttempts(15)
		waiterDelay := request.WithWaiterDelay(request.ConstantWaiterDelay(2 * time.Second))

		err = lambdaSvc.WaitUntilFunctionUpdatedWithContext(
			aws.BackgroundContext(),
			&lambda.GetFunctionConfigurationInput{
				FunctionName: aws.String(cfg.Name),
			},
			waiterCfg,
			waiterDelay,
		)
		if err != nil {
			return fmt.Errorf("failed to wait for function update (make sure you have lambda:GetFunctionConfiguration permission): %v", err)
		}

		_, err = lambdaSvc.UpdateFunctionConfiguration(&lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(cfg.Name),
			Environment: &lambda.Environment{
				Variables: awsEnvVariables,
			},
			Runtime: aws.String(cfg.AWSLambda.Runtime),
		})
		if err != nil {
			return fmt.Errorf("failed to update function configuration: %v", err)
		}
	}

	err = addPublicAccessPermission(lambdaSvc, cfg.Name)
	if err != nil {
		log.Printf("Warning: Failed to add public access permission: %v", err)
	}

	createUrlConfig, err := lambdaSvc.CreateFunctionUrlConfig(&lambda.CreateFunctionUrlConfigInput{
		AuthType:     aws.String("NONE"),
		FunctionName: aws.String(cfg.Name),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == lambda.ErrCodeResourceConflictException {
			getUrlConfig, err := lambdaSvc.GetFunctionUrlConfig(&lambda.GetFunctionUrlConfigInput{
				FunctionName: aws.String(cfg.Name),
			})
			if err != nil {
				return fmt.Errorf("failed to get function URL: %v", err)
			}
			fmt.Printf("\nExisting Function URL: %s\n\n", *getUrlConfig.FunctionUrl)
		} else {
			return fmt.Errorf("failed to create function URL: %v", err)
		}
	} else {
		fmt.Printf("\nNew Function URL: %s\n\n", *createUrlConfig.FunctionUrl)
	}

	return nil
}

func addPublicAccessPermission(lambdaSvc *lambda.Lambda, functionName string) error {
	_, err := lambdaSvc.AddPermission(&lambda.AddPermissionInput{
		Action:              aws.String("lambda:InvokeFunctionUrl"),
		FunctionName:        aws.String(functionName),
		Principal:           aws.String("*"),
		StatementId:         aws.String("FunctionURLAllowPublicAccess"),
		FunctionUrlAuthType: aws.String("NONE"),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeResourceConflictException:
				return nil
			default:
				return fmt.Errorf("failed to add permission: %v", err)
			}
		}
		return fmt.Errorf("failed to add permission: %v", err)
	}

	return nil
}

func validateAWSLambdaConfig(cfg *config.Config) error {
	if cfg.AWSLambda == nil {
		return fmt.Errorf("AWS Lambda configuration is missing")
	}
	if cfg.AWSLambda.Region == "" {
		return fmt.Errorf("AWS Lambda region is not specified")
	}
	if cfg.AWSLambda.RoleName == "" {
		return fmt.Errorf("AWS Lambda role name is not specified")
	}
	if cfg.AWSLambda.Runtime == "" {
		return fmt.Errorf("AWS Lambda runtime is not specified")
	}
	return nil
}
