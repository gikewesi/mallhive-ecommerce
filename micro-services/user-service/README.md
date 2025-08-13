# Mallhive User Microservice

This is the `user-service` component of the Mallhive platform. It handles user registration, login, email verification, password resets, and token-based authentication.

---

## Features

* ✅ Secure JWT-based authentication
* ✅ Email verification and password reset flows
* ✅ Internal communication via private DNS (e.g., `notification.internal.mallhive.com`)
* ✅ Secrets managed via AWS Secrets Manager
* ✅ Structured logging and metrics sent to centralized monitoring repo
* ✅ Built for AWS EKS (Fargate-compatible)

---

## Tech Stack

* **FastAPI** (web framework)
* **SQLAlchemy** + **PostgreSQL** (ORM + DB)
* **AWS Secrets Manager** (for DB credentials)
* **Prometheus** (metrics)
* **Datadog** or **CloudWatch** (logging)
* **Private Route53 DNS** (internal service discovery)

---

## Folder Structure

```bash
user-service/
├── main.py                 # FastAPI app entrypoint
├── auth.py                 # Auth logic (JWT, verification, reset)
├── database.py             # SQLAlchemy models and session logic
├── secrets.py              # Secure secrets fetch from AWS
├── logging.py              # Pre-configured logger + Datadog events
├── metrics.py              # Prometheus-compatible metrics export
├── tests.py                # Basic test suite (Pytest-ready)
├── requirements.txt
└── README.md
```

---

## Secrets Management

Credentials and sensitive config are pulled from **AWS Secrets Manager**.

### Example Secret Format (`prod/user-service/db-creds`):

```json
{
  "username": "mallhive_user",
  "password": "secure_password",
  "host": "user-db.internal.mallhive.com",
  "port": 5432,
  "dbname": "user_service"
}
```

---

## Environment Variables

| Variable             | Description                       |
| -------------------- | --------------------------------- |
| `AWS_REGION`         | AWS region (default: `us-east-1`) |
| `USER_SERVICE_SECRET_NAME`     | Secrets Manager ID for DB creds   |
| `JWT_SECRET_KEY`     | Secret for signing JWTs           |
| `JWT_EXPIRATION_MIN` | JWT expiration time (in minutes)  |
| `DATADOG_API_KEY`    | (Optional) API key for Datadog    |
| `DATADOG_APP_KEY`    | (Optional) App key for Datadog    |

---

## Internal Service Calls

Uses internal Route 53 DNS for communication:

* `notification.internal.mallhive.com` → Sends email/SMS via internal notification service

---

## Metrics & Observability

* **Metrics**: Exposed at `/metrics` (Prometheus format)
* **Logs**: Structured logs + optional push to Datadog or CloudWatch
* **Events**: Custom log events sent via `logging.py` to your monitoring repo

---

## Running Locally

```bash
pip install -r requirements.txt
uvicorn main:app --reload
```

Set AWS credentials in your environment (or use `aws configure`) to allow fetching secrets.

---

## Tests

```bash
pytest tests.py
```

---

## Deployment

This service is meant to run in AWS EKS with Fargate profiles. It assumes:

* Secrets are stored in AWS Secrets Manager
* The pod uses IAM roles for service accounts (IRSA) to access AWS APIs
* Logging and metrics forwarders (e.g., FluentBit, OpenTelemetry Collector) are configured

---