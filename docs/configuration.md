---
layout: default
title: Configuration
permalink: /configuration
nav_order: 4
---

# Configuration

Configuration is stored in `.upify/config.yaml`.

## Basic Configuration
```yaml
name: project-name
framework: flask | express | none
language: python | nodejs
package_manager: pip | npm
entrypoint: main.py
app_var: app
```

## AWS Lambda Configuration
```yaml
aws-lambda:
  region: us-east-1
  role_name: lambda-role
  runtime: python3.12 | nodejs20.x
```

## GCP Cloud Run Configuration
```yaml
gcp-cloudrun:
  region: us-central1
  project_id: your-project-id
  runtime: python312 | nodejs20
```

## Reference

| Field | Description |
|-------|-------------|
| name | Project name |
| framework | Web framework being used |
| language | Programming language |
| package_manager | Package management tool |
| entrypoint | Main application file |
| app_var | App variable name in entrypoint |