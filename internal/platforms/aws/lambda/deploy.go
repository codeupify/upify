package lambda

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/codeupify/upify/internal/config"
)

var excludedDirs = []string{".git", "node_modules", "venv", ".", ".upify"}

func Deploy(cfg *config.Config) error {
	if cfg.AWSLambda == nil {
		return fmt.Errorf("AWS Lambda configuration is missing")
	}

	tempDir, err := os.MkdirTemp("", "lambda_deployment_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = copyFilesToTempDir(".", tempDir)
	if err != nil {
		return fmt.Errorf("failed to copy files to temp directory: %v", err)
	}

	err = copyFile(".upify/lambda_handler.py", filepath.Join(tempDir, "lambda_handler.py"))
	if err != nil {
		return fmt.Errorf("failed to copy lambda_handler.py: %v", err)
	}

	err = installRequirements(tempDir, cfg.Framework)
	if err != nil {
		return fmt.Errorf("failed to install requirements: %v", err)
	}

	zipBuffer, err := createZip(tempDir)
	if err != nil {
		return fmt.Errorf("failed to create zip: %v", err)
	}

	zipPath := filepath.Join(tempDir, "output.zip")
	os.WriteFile(zipPath, zipBuffer.Bytes(), 0644)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSLambda.Region),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	err = getOrCreateLambda(sess, cfg, zipPath)
	if err != nil {
		return fmt.Errorf("failed to get or create Lambda: %v", err)
	}

	return nil
}

func copyFilesToTempDir(srcDir, destDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, exclude := range excludedDirs {
			if strings.HasSuffix(path, "/"+exclude) || strings.HasSuffix(path, "/"+exclude+"/") {
				return filepath.SkipDir
			}
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func installRequirements(dir string, framework string) error {
	if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
		cmd := exec.Command("pip", "install", "-r", "requirements.txt", "-t", dir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install requirements: %v", err)
		}
	}

	if strings.ToLower(framework) == "flask" {
		cmd := exec.Command("pip", "install", "apig-wsgi", "-t", dir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install apig-wsgi: %v", err)
		}
		fmt.Println("Installed apig-wsgi for Flask framework")
	}

	return nil
}

func createZip(dir string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buffer, nil
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

	return *result.Role.Arn, nil
}

func getOrCreateLambda(sess *session.Session, cfg *config.Config, zipPath string) error {
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
	_, err = lambdaSvc.GetFunction(&lambda.GetFunctionInput{
		FunctionName: aws.String(cfg.Name),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == lambda.ErrCodeResourceNotFoundException {
			roleArn, err := getOrCreateRole(iamSvc, cfg.AWSLambda.RoleName)
			if err != nil {
				return fmt.Errorf("failed to get or create IAM role: %v", err)
			}

			_, err = lambdaSvc.CreateFunction(&lambda.CreateFunctionInput{
				Code: &lambda.FunctionCode{
					ZipFile: zipBytes,
				},
				FunctionName: aws.String(cfg.Name),
				Handler:      aws.String("lambda_handler.handler"),
				Role:         aws.String(roleArn),
				Runtime:      aws.String(cfg.AWSLambda.Runtime),
			})
			if err != nil {
				return fmt.Errorf("failed to create function: %v", err)
			}
			fmt.Println("Successfully created and deployed Python app to AWS Lambda")
		} else {
			return fmt.Errorf("failed to check if function exists: %v", err)
		}
	} else {
		_, err = lambdaSvc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(cfg.Name),
			ZipFile:      zipBytes,
		})
		if err != nil {
			return fmt.Errorf("failed to update function code: %v", err)
		}
		fmt.Println("Successfully updated Python app on AWS Lambda")
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
			fmt.Printf("Existing Function URL: %s\n", *getUrlConfig.FunctionUrl)
		} else {
			return fmt.Errorf("failed to create function URL: %v", err)
		}
	} else {
		fmt.Printf("New Function URL: %s\n", *createUrlConfig.FunctionUrl)
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
