# upify

A CLI tool that streamlines cloud deployment. Designed to be platform, language, and framework agnostic.

- **Quick and simple cloud deployment**
- **Multi-platform support**: Deploy to a single platform or multiple platforms simultaneously. Currently supports AWS Lambda and GCP Cloud Run
- **Multi-framework support**: Currently supports Flask and Express
- **Multi-runtime support**: Currently supports Python and Node.js
- **Adapts to your project structure**: Wraps your existing project
- **Generates version-controlled and adjustable configs**


## Installation

You can install the latest version of upify from GitHub by following these steps:

### Option 1: Install via `go install`
1. Ensure you have Go installed on your system (version 1.16 or later).
2. Run the following command to install upify:

   ```bash
   go install github.com/codeupify/upify@latest
   ```
3. Verify the installation:

    ```bash
    upify --help
    ```

### Option 2: Install pre-built binaries

#### For Linux and macOS:
1. Download the latest release for your operating system from the [GitHub releases page](https://github.com/codeupify/upify/releases).
2. Unpack the binary for your operating system.
3. Move the binary to a directory included in your system's `PATH`:
    ```bash
    mv upify /usr/local/bin/
    ```
4. Make the binary executable:
    ```bash
    chmod +x /usr/local/bin/upify
    ```
5. Verify the installation:
    ```
    upify --help
    ```

#### For Windows:
1. Download the latest release for your operating system from the [GitHub releases page](https://github.com/codeupify/upify/releases).
2. Unpack the binary.
3. Add the directory containing the binary to your system's `PATH`.
4. Verify the installation by opening a new Command Prompt and running `upify --help`.


## Usage

### Initialize your project

Run the following command at the base of your project to initialize it:

```bash
upify init
```

This command will generate config and wrapper files. Depending on the options selected, you may need to adjust the generated code and config files. Follow the instructions provided in the command output.

### Add a platform

To add cloud platform support, run:

```bash
upify platform add [platform]
```

### Deploy to the cloud

To deploy your project, use the following command:

```bash
upify deploy aws-lambda
```

*Note: You must have your cloud credentials set up before deploying. See the [Authentication](#authentication) section for more details.*

## Example projects

- [Flask](/examples/flask/README.md)
- [Express](/examples/express/README.md)
- [Python without a framework](/examples/python/README.md)
- [Javascript without a framework](/examples/javascript/README.md)
- [Typescript without a framework](/examples/typescript/README.md)


## [Authentication](#authentication)

### AWS - Lambda

#### Setting up AWS Credentials

1. Log into your AWS Console
2. Go to IAM (Identity and Access Management)
3. Create a new IAM user or select an existing one
4. Attach permissions (Required permissions can be found [here](#aws-lambda-required-permissions))
5. Under "Security credentials", create a new access key and save those credentials

#### Configuring Credentials

##### Option 1: AWS CLI Configure

First, install AWS CLI:

- macOS: ```brew install awscli```
- Windows: Download the official MSI installer
- Linux: ```apt install awscli``` or ```yum install awscli```

Then configure:
```bash
aws configure
```

This will prompt you to enter:
- AWS Access Key ID
- AWS Secret Access Key
- Default region (e.g., `us-east-1`)
- Default output format (json)

#### Option 2: Manual Credentials File

```bash
mkdir ~/.aws
cat > ~/.aws/credentials << EOF
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
EOF
```

#### Option 3: Environment Variables

```bash
export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_KEY"
export AWS_DEFAULT_REGION="us-east-1"
```


### [Required Permissions](#aws-lambda-required-permissions)
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "lambda:CreateFunction",
                "lambda:UpdateFunctionCode",
                "lambda:GetFunction",
                "lambda:CreateFunctionUrlConfig",
                "lambda:GetFunctionUrlConfig",
                "lambda:AddPermission",
                "lambda:UpdateFunctionConfiguration",
                "lambda:GetFunctionConfiguration"
            ],
            "Resource": "arn:aws:lambda:*:*:function:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "iam:GetRole",
                "iam:CreateRole",
                "iam:AttachRolePolicy",
                "iam:PassRole"
            ],
            "Resource": "arn:aws:iam::*:role/*"
        }
    ]
}
```

### GCP - Cloud Run

#### Setting up GCP Project
1. Log into GCP Console
2. Enable required APIs (Required APIs can be found [here](#gcp-cloudrun-required-permissions))

#### Configuring Credentials

##### Option 1: User Account
First, install Google Cloud SDK - https://cloud.google.com/sdk/docs/install

Then authenticate:
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```
##### Option 2: Service Account
1. Create a service account in GCP Console:
    1. Go to IAM & Admin > Service Accounts
    2. Click "Create Service Account"
    3. Add required roles (Required roles can be found [here](#gcp-cloudrun-required-permissions))
2. Download service account key (JSON format)
3. Set the credentials:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

### [Required Permissions](#gcp-cloudrun-required-permissions)
Required Roles:

- Cloud Functions Admin (`roles/cloudfunctions.admin`)
- Cloud Functions Developer (`roles/cloudfunctions.developer`)
- Service Account User (`roles/iam.serviceAccountUser`)
- Storage Admin (`roles/storage.admin`)
- Viewer (`roles/viewer`)

Required APIs:

- Cloud Functions API
- Cloud Run API
- Cloud Build API
- Artifact Registry API
- Cloud Resource Manager API
- Cloud Storage API



