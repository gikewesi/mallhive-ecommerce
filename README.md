# 🛍️ Cloud-Native E-commerce Platform

A robust, scalable, and cloud-native e-commerce platform built using **Python (Django/Flask)**, **Go**, **AWS (EKS, RDS, S3, CloudFront, Fargate)**, **Docker**, **Kubernetes**, **Terraform**, **GitHub Actions**, and **ArgoCD**. The platform is designed with a microservices architecture and GitOps-driven CI/CD workflows to ensure high availability, scalability, and ease of maintenance.

---

## 🚀 Features

- 🧩 **Microservices-based architecture** for modularity and scalability
- ⚙️ **GitOps** with ArgoCD for continuous delivery
- 📦 **CI/CD Pipelines** using GitHub Actions
- 🐳 Containerized deployment with Docker and Kubernetes (EKS)
- ☁️ AWS-native resources via Terraform (IaC)
- 🔐 Secure API gateway and RBAC
- 🛒 Product catalog, cart, checkout, and order tracking
- 📈 Observability with Prometheus, Grafana, OpenTelemetry
- 🔐 Secrets management via AWS Secrets Manager

---

## 🛠️ Tech Stack

| Layer | Technologies |
|------|--------------|
| **Frontend** | React (future), CloudFront |
| **Backend** | Django or Flask (Python), Go |
| **Microservices** | Cart, Checkout, Payment, Auth, Inventory |
| **CI/CD** | GitHub Actions, ArgoCD |
| **Infrastructure** | AWS EKS, RDS, S3, Route 53, IAM, Fargate |
| **IaC** | Terraform, Terragrunt |
| **Containerization** | Docker, Kubernetes, Helm |
| **Monitoring** | Prometheus, Grafana, OpenTelemetry |
| **Secrets** | AWS Secrets Manager |
| **Networking** | ALB, Ingress, API Gateway |

---

## 🧱 System Architecture

```mermaid
graph TD
  subgraph Frontend
    A[User] --> B[CloudFront -> React App]
  end

  subgraph API Gateway & Auth
    B --> C[API Gateway]
    C --> D[Auth Service (JWT)]
  end

  subgraph Services
    C --> E[Cart Service]
    C --> F[Checkout Service]
    C --> G[Payment Service]
    C --> H[Product Catalog Service]
    C --> I[User Profile Service]
  end

  subgraph Data Layer
    E --> J[(RDS - PostgreSQL)]
    F --> J
    G --> J
    H --> J
    I --> J
  end

  subgraph DevOps
    K[GitHub] --> L[GitHub Actions -> Build & Push]
    L --> M[ArgoCD -> Deploy to EKS]
    M --> EKS[Kubernetes (EKS)]
    EKS --> Services
  end

  subgraph Monitoring
    Services --> N[Prometheus / Grafana / OTEL]
  end

