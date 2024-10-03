# Flask

## AWS Lambda

```bash
upify platform aws-lambda
```

This command will create the `.upify` folder with a `handler.py` and add an `aws-lambda` section to `config.yml`. You will also have to adjust `handler.py` to point to the Flask app in your module.

Deploy with
```bash
upify deploy aws-lambda
```

## GCP Cloud Run
TODO