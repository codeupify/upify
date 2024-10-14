package lambda

// func loadEnvVariables() (map[string]string, error) {
// 	cwd, err := os.Getwd()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get current working directory: %v", err)
// 	}

// 	envPath := filepath.Join(cwd, ".upify", ".env")

// 	if _, err := os.Stat(envPath); os.IsNotExist(err) {
// 		fmt.Println("No .upify/.env file found, not adding environment variables")
// 		return make(map[string]string), nil
// 	}

// 	env, err := godotenv.Read(envPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read .env file: %v", err)
// 	}

// 	return env, nil
// }

// func installRequirements(dir string, cfg *config.Config) error {
// 	switch cfg.PackageManager {
// 	case config.Pip:
// 		return installPythonRequirements(dir, cfg.Framework)
// 	case config.Npm, config.Yarn:
// 		return installNodeRequirements(dir, cfg.Framework, cfg.PackageManager)
// 	default:
// 		return fmt.Errorf("unsupported package manager: %s", cfg.PackageManager)
// 	}
// }

// func installPythonRequirements(dir string, framework config.Framework) error {
// 	requirementsFile := filepath.Join(dir, "requirements.txt")
// 	if _, err := os.Stat(requirementsFile); err == nil {
// 		cmd := exec.Command("pip", "install", "-r", requirementsFile, "-t", dir)
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 		if err := cmd.Run(); err != nil {
// 			return fmt.Errorf("failed to install Python requirements: %v", err)
// 		}
// 	}

// 	if framework == config.Flask {
// 		cmd := exec.Command("pip", "install", "apig-wsgi", "-t", dir)
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 		if err := cmd.Run(); err != nil {
// 			return fmt.Errorf("failed to install apig-wsgi: %v", err)
// 		}
// 		fmt.Println("Installed apig-wsgi for Flask framework")
// 	}

// 	return nil
// }

// func parsePackageJSON(path string) (*PackageJSON, error) {
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var pkg PackageJSON
// 	if err := json.Unmarshal(data, &pkg); err != nil {
// 		return nil, err
// 	}

// 	return &pkg, nil
// }

// func installNodeRequirements(dir string, framework config.Framework, packageManager config.PackageManager) error {
// 	packageJsonFile := filepath.Join(dir, "package.json")
// 	if _, err := os.Stat(packageJsonFile); err == nil {
// 		pkg, err := parsePackageJSON(packageJsonFile)
// 		if err != nil {
// 			return fmt.Errorf("failed to parse package.json: %v", err)
// 		}

// 		var installCmd *exec.Cmd
// 		if packageManager == config.Npm {
// 			installCmd = exec.Command("npm", "install", "--production")
// 		} else {
// 			installCmd = exec.Command("yarn", "install", "--production")
// 		}

// 		installCmd.Dir = dir
// 		installCmd.Stdout = os.Stdout
// 		installCmd.Stderr = os.Stderr
// 		if err := installCmd.Run(); err != nil {
// 			return fmt.Errorf("failed to install Node.js dependencies: %v", err)
// 		}

// 		if framework == config.Express {
// 			var cmd *exec.Cmd
// 			if packageManager == config.Npm {
// 				cmd = exec.Command("npm", "install", "serverless-http")
// 			} else {
// 				cmd = exec.Command("yarn", "add", "serverless-http")
// 			}
// 			cmd.Dir = dir
// 			cmd.Stdout = os.Stdout
// 			cmd.Stderr = os.Stderr
// 			if err := cmd.Run(); err != nil {
// 				return fmt.Errorf("failed to install serverless-http: %v", err)
// 			}
// 			fmt.Println("Installed serverless-http for Express framework")
// 		}

// 		if _, hasBuild := pkg.Scripts["build"]; hasBuild {
// 			var buildCmd *exec.Cmd
// 			if packageManager == config.Npm {
// 				buildCmd = exec.Command("npm", "run", "build")
// 			} else {
// 				buildCmd = exec.Command("yarn", "build")
// 			}

// 			buildCmd.Dir = dir
// 			buildCmd.Stdout = os.Stdout
// 			buildCmd.Stderr = os.Stderr
// 			if err := buildCmd.Run(); err != nil {
// 				return fmt.Errorf("failed to build Node.js project: %v", err)
// 			}
// 			fmt.Println("Successfully built Node.js project")
// 		} else {
// 			fmt.Println("No build script found; skipping build step")
// 		}
// 	} else {
// 		return fmt.Errorf("package.json not found in directory: %s", dir)
// 	}

// 	return nil
// }

// func getOrCreateRole(svc *iam.IAM, roleName string) (string, error) {
// 	getRoleInput := &iam.GetRoleInput{
// 		RoleName: aws.String(roleName),
// 	}
// 	if result, err := svc.GetRole(getRoleInput); err == nil {
// 		return *result.Role.Arn, nil
// 	}

// 	assumeRolePolicyDocument := `{
// 		"Version": "2012-10-17",
// 		"Statement": [
// 			{
// 				"Effect": "Allow",
// 				"Principal": {
// 					"Service": "lambda.amazonaws.com"
// 				},
// 				"Action": "sts:AssumeRole"
// 			}
// 		]
// 	}`

// 	createRoleInput := &iam.CreateRoleInput{
// 		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDocument),
// 		RoleName:                 aws.String(roleName),
// 	}

// 	result, err := svc.CreateRole(createRoleInput)
// 	if err != nil {
// 		return "", err
// 	}

// 	attachPolicyInput := &iam.AttachRolePolicyInput{
// 		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
// 		RoleName:  aws.String(roleName),
// 	}
// 	_, err = svc.AttachRolePolicy(attachPolicyInput)
// 	if err != nil {
// 		return "", err
// 	}

// 	fmt.Println("Waiting for role to be fully ready...")
// 	time.Sleep(15 * time.Second)

// 	return *result.Role.Arn, nil
// }

// func getOrCreateLambda(sess *session.Session, cfg *config.Config, zipPath string, envVariables map[string]string) error {
// 	zipFile, err := os.Open(zipPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to open zip file: %v", err)
// 	}
// 	defer zipFile.Close()

// 	zipBytes, err := io.ReadAll(zipFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to read zip file: %v", err)
// 	}

// 	lambdaSvc := lambda.New(sess)
// 	iamSvc := iam.New(sess)

// 	awsEnvVariables := make(map[string]*string)
// 	for k, v := range envVariables {
// 		awsEnvVariables[k] = aws.String(v)
// 	}

// 	_, err = lambdaSvc.GetFunction(&lambda.GetFunctionInput{
// 		FunctionName: aws.String(cfg.Name),
// 	})

// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == lambda.ErrCodeResourceNotFoundException {
// 			roleArn, err := getOrCreateRole(iamSvc, cfg.AWSLambda.RoleName)
// 			if err != nil {
// 				return fmt.Errorf("failed to get or create IAM role: %v", err)
// 			}

// 			_, err = lambdaSvc.CreateFunction(&lambda.CreateFunctionInput{
// 				Code: &lambda.FunctionCode{
// 					ZipFile: zipBytes,
// 				},
// 				FunctionName: aws.String(cfg.Name),
// 				Handler:      aws.String("lambda_handler.handler"),
// 				Role:         aws.String(roleArn),
// 				Runtime:      aws.String(cfg.AWSLambda.Runtime),
// 				Environment: &lambda.Environment{
// 					Variables: awsEnvVariables,
// 				},
// 			})
// 			if err != nil {
// 				return fmt.Errorf("failed to create function: %v", err)
// 			}
// 			fmt.Println("Successfully created and deployed app to AWS Lambda")
// 		} else {
// 			return fmt.Errorf("failed to check if function exists: %v", err)
// 		}
// 	} else {
// 		_, err = lambdaSvc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
// 			FunctionName: aws.String(cfg.Name),
// 			ZipFile:      zipBytes,
// 		})
// 		if err != nil {
// 			return fmt.Errorf("failed to update function code: %v", err)
// 		}

// 		// Need to wait for function to be updated
// 		waiterCfg := request.WithWaiterMaxAttempts(15)
// 		waiterDelay := request.WithWaiterDelay(request.ConstantWaiterDelay(2 * time.Second))

// 		err = lambdaSvc.WaitUntilFunctionUpdatedWithContext(
// 			aws.BackgroundContext(),
// 			&lambda.GetFunctionConfigurationInput{
// 				FunctionName: aws.String(cfg.Name),
// 			},
// 			waiterCfg,
// 			waiterDelay,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to wait for function update (make sure you have lambda:GetFunctionConfiguration permission): %v", err)
// 		}

// 		_, err = lambdaSvc.UpdateFunctionConfiguration(&lambda.UpdateFunctionConfigurationInput{
// 			FunctionName: aws.String(cfg.Name),
// 			Environment: &lambda.Environment{
// 				Variables: awsEnvVariables,
// 			},
// 		})
// 		if err != nil {
// 			return fmt.Errorf("failed to update function configuration: %v", err)
// 		}

// 		fmt.Println("Successfully updated app on AWS Lambda")
// 	}

// 	err = addPublicAccessPermission(lambdaSvc, cfg.Name)
// 	if err != nil {
// 		log.Printf("Warning: Failed to add public access permission: %v", err)
// 	}

// 	createUrlConfig, err := lambdaSvc.CreateFunctionUrlConfig(&lambda.CreateFunctionUrlConfigInput{
// 		AuthType:     aws.String("NONE"),
// 		FunctionName: aws.String(cfg.Name),
// 	})

// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == lambda.ErrCodeResourceConflictException {
// 			getUrlConfig, err := lambdaSvc.GetFunctionUrlConfig(&lambda.GetFunctionUrlConfigInput{
// 				FunctionName: aws.String(cfg.Name),
// 			})
// 			if err != nil {
// 				return fmt.Errorf("failed to get function URL: %v", err)
// 			}
// 			fmt.Printf("\nExisting Function URL: %s\n\n", *getUrlConfig.FunctionUrl)
// 		} else {
// 			return fmt.Errorf("failed to create function URL: %v", err)
// 		}
// 	} else {
// 		fmt.Printf("\nNew Function URL: %s\n\n", *createUrlConfig.FunctionUrl)
// 	}

// 	return nil
// }

// func addPublicAccessPermission(lambdaSvc *lambda.Lambda, functionName string) error {
// 	_, err := lambdaSvc.AddPermission(&lambda.AddPermissionInput{
// 		Action:              aws.String("lambda:InvokeFunctionUrl"),
// 		FunctionName:        aws.String(functionName),
// 		Principal:           aws.String("*"),
// 		StatementId:         aws.String("FunctionURLAllowPublicAccess"),
// 		FunctionUrlAuthType: aws.String("NONE"),
// 	})

// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			case lambda.ErrCodeResourceConflictException:
// 				return nil
// 			default:
// 				return fmt.Errorf("failed to add permission: %v", err)
// 			}
// 		}
// 		return fmt.Errorf("failed to add permission: %v", err)
// 	}

// 	return nil
// }

// func validateAWSLambdaConfig(cfg *config.Config) error {
// 	if cfg.AWSLambda == nil {
// 		return fmt.Errorf("AWS Lambda configuration is missing")
// 	}
// 	if cfg.AWSLambda.Region == "" {
// 		return fmt.Errorf("AWS Lambda region is not specified")
// 	}
// 	if cfg.AWSLambda.RoleName == "" {
// 		return fmt.Errorf("AWS Lambda role name is not specified")
// 	}
// 	if cfg.AWSLambda.Runtime == "" {
// 		return fmt.Errorf("AWS Lambda runtime is not specified")
// 	}
// 	return nil
// }

// func Deploy(cfg *config.Config) error {
// 	if err := validateAWSLambdaConfig(cfg); err != nil {
// 		return err
// 	}

// 	tempDir, err := os.MkdirTemp("", "lambda_deployment_")
// 	if err != nil {
// 		return fmt.Errorf("failed to create temp directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	err = copyFilesToTempDir(".", tempDir)
// 	if err != nil {
// 		return fmt.Errorf("failed to copy files to temp directory: %v", err)
// 	}

// 	err = installRequirements(tempDir, cfg)
// 	if err != nil {
// 		return fmt.Errorf("failed to install requirements: %v", err)
// 	}

// 	zipBuffer, err := createZip(tempDir)
// 	if err != nil {
// 		return fmt.Errorf("failed to create zip: %v", err)
// 	}

// 	zipPath := filepath.Join(tempDir, "output.zip")
// 	os.WriteFile(zipPath, zipBuffer.Bytes(), 0644)

// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String(cfg.AWSLambda.Region),
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to create session: %v", err)
// 	}

// 	envVars, err := loadEnvVariables()
// 	if err != nil {
// 		return fmt.Errorf("failed to load environment variables: %v", err)
// 	}

// 	err = getOrCreateLambda(sess, cfg, zipPath, envVars)
// 	if err != nil {
// 		return fmt.Errorf("failed to get or create Lambda: %v", err)
// 	}

// 	return nil
// }
