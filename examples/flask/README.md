# Flask

Create an `.upify` folder with a `config.yml` that describes your app:

```bash
upify init
```

## AWS Lambda

Add a `lambda_handler.py` and an `aws-lambda` section to the config with this command:

```bash
upify platform add aws-lambda
```

After running this command
1. Make sure that `lambda_handler.py` is pointing to the Flask app (follow the comments)
2. Make adjustments to  `config.yml` as needed

Deploy with:

```bash
upify deploy aws-lambda
```

## GCP Cloud Run
TODO