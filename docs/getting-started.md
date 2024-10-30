---
layout: default
title: Getting Started
permalink: /getting-started
nav_order: 2
---

# Getting Started

## Installation

### Go Install
```bash
go install github.com/codeupify/upify@latest
```

### Pre-built Binaries

#### Linux and macOS
1. Download from [releases page](https://github.com/codeupify/upify/releases)
2. Move to PATH:
```bash
mv upify /usr/local/bin/
chmod +x /usr/local/bin/upify
```

#### Windows
1. Download from [releases page](https://github.com/codeupify/upify/releases)
2. Add binary location to system PATH

## Quick Start

1. Initialize project:
```bash
upify init
```

2. Set environment variables in `.upify/.env` (see [Environment Variables](./environment-variables))

3. If you aren't using a web framework, you'll need to add code to your wrappers (see [Wrappers](./wrappers))

4. Add cloud platform:
```bash
upify platform add aws-lambda
```

5. Deploy:
```bash
upify deploy aws-lambda
```
