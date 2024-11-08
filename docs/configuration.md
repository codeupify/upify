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

### Reference

| Field | Description |
|-------|-------------|
| name | Project name |
| framework | Web framework being used |
| language | Programming language |
| package_manager | Package management tool |
| entrypoint | Main application file |
| app_var | App variable name in entrypoint |

# Terraform

Upify leverages Terraform for infrastructure management, terraform files are written to:

- `.upify/environments`
- `.upify/modules`

At this time Upify only supports a single `prod` environment, but in the future we will add support for additional environments