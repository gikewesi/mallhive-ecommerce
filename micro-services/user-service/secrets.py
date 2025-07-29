# secrets.py

import os
import json
import boto3
from functools import lru_cache
from botocore.exceptions import ClientError
from fastapi.security import OAuth2PasswordBearer

# OAuth2 scheme for dependency injection
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="login")

@lru_cache()
def get_secret(secret_name: str = None, region_name: str = "us-east-1") -> dict:
    """
    Load secrets from AWS Secrets Manager. Caches result using lru_cache.
    """
    secret_name = secret_name or os.getenv("USER_SERVICE_SECRET_NAME", "prod/user-service")

    client = boto3.client("secretsmanager", region_name=region_name)
    try:
        response = client.get_secret_value(SecretId=secret_name)
        secret_string = response.get("SecretString")
        if not secret_string:
            raise ValueError("Empty secret string.")
        return json.loads(secret_string)
    except ClientError as e:
        raise RuntimeError(f"Unable to retrieve secrets: {e}")
