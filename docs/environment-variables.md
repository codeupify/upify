---
layout: default
title: Environment Variables
permalink: /environment-variables
nav_order: 6
---

# Environment Variables

Environment variables for your application should be placed in `.upify/.env`. These variables will be automatically loaded and set during deployment.

Example `.upify/.env`:
```bash
API_KEY=your-api-key
DATABASE_URL=your-database-url
NODE_ENV=production
```

All variables in this file will be available to your application at runtime on the cloud platform.