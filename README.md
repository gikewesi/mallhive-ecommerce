# 🛍️ MallHive: Cloud-Native E-Commerce Platform

MallHive is a scalable, microservices-based e-commerce platform designed to connect buyers and sellers with personalized shopping experiences. It leverages modern cloud-native technologies to ensure high availability, performance, and ease of deployment.

---

## 📖 Table of Contents

* [Overview](#-overview)
* [Architecture](#-architecture)
* [Tech Stack](#-tech-stack)
* [Microservices](#-microservices)
* [Supporting Repositories](#-supporting-repositories)
* [Infrastructure](#-infrastructure)
* [CI/CD Pipeline](#-cicd-pipeline)
* [Security](#-security)
* [Monitoring & Observability](#-monitoring--observability)
* [Micro Frontends](#-micro-frontends)
* [Deployment Roadmap](#-deployment-roadmap)
* [Contributing](#-contributing)
* [License](#-license)

---

## 📄 Overview

MallHive allows sellers to list products and buyers to shop with ease, backed by AI-powered personalized recommendations. It is built using microservices to enable independent development, deployment, and scaling of individual components.

---

## 🏗️ Architecture

The platform consists of multiple microservices communicating via REST and gRPC, supported by event-driven messaging for asynchronous workflows. The system runs on AWS EKS with serverless components for scalability.

*(architecture diagram here)*

---

## 🛠️ Tech Stack Highlights

* Backend: FastAPI (Python), Go, NestJS (Node.js)
* Frontend: React.js (micro frontends)
* Databases: PostgreSQL (AWS RDS), Redis
* Search: OpenSearch
* AI/ML: AWS SageMaker
* Containerization: Docker, Kubernetes (EKS + Fargate)
* Infrastructure: Terraform
* CI/CD: Jenkins, AWS CodeBuild, CodePipeline, CodeCommit, GitHub Actions, ArgoCD
* Monitoring: Prometheus, Grafana, AWS CloudWatch
* Messaging: AWS SNS, SQS, EventBridge
* Storage: AWS S3, CloudFront
* Secrets & Encryption: AWS Secrets Manager, KMS

---

## 🧩 Microservices Overview

Each service is containerized and deployed independently on Kubernetes (EKS), with CI/CD automations for testing, building, and deployment using multiple tools.

### 1. User Authentication Service
**Purpose**: Handles user registration, login, JWT authentication, and profile management.  
**Tech Stack**:
- FastAPI (Python)
- PostgreSQL (AWS RDS)
- JWT, OAuth2
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD
- Prometheus, Grafana, CloudWatch
- AWS Secrets Manager

---

### 2. Product Catalog Service
**Purpose**: Manages product data, categories, pricing, and availability.  
**Tech Stack**:
- Go
- PostgreSQL (AWS RDS)
- OpenSearch
- gRPC (internal), REST (external)
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD
- Prometheus, CloudWatch

---

### 3. Shopping Cart & Inventory Service
**Purpose**: Handles cart operations and tracks product availability.  
**Tech Stack**:
- Go
- Redis (for cart operations)
- PostgreSQL (for inventory tracking)
- REST API
- AWS SNS for checkout events
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD

---

### 4. Order Management Service
**Purpose**: Handles order placement, tracking, and fulfillment.  
**Tech Stack**:
- Go
- PostgreSQL (AWS RDS)
- Kafka or AWS SQS
- AWS EventBridge
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD

---

### 5. Payment Gateway Service
**Purpose**: Handles payment processing via Stripe or PayPal.  
**Tech Stack**:
- FastAPI (Python)
- Stripe, PayPal
- AWS KMS (Encryption)
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD

---

### 6. Notification Service
**Purpose**: Sends emails, SMS, and push notifications.  
**Tech Stack**:
- Node.js (NestJS)
- AWS SNS/SQS
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD

---

### 7. Search & Recommendations Service
**Purpose**: Implements product search and AI-driven recommendations.  
**Tech Stack**:
- Go
- OpenSearch
- AWS SageMaker
- gRPC (internal), REST (external)
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD

---

### 8. Analytics & Reporting Service
**Purpose**: Logs user behavior, sales metrics, and trends.  
**Tech Stack**:
- Go
- AWS Kinesis, AWS Lambda
- AWS Redshift
- Grafana, AWS QuickSight
- Docker, Kubernetes (EKS)
- GitHub Actions, ArgoCD

---

## 📂 Supporting Repositories

MallHive is split into several repositories to keep concerns separate and manageable:

1. **[mallhive-infra](https://github.com/gikewesi/mallhive-infra)**
   Contains all Infrastructure as Code (IaC) configurations, primarily Terraform scripts, that provision and manage cloud resources including VPCs, EKS clusters, RDS instances, networking, container registries, and CI/CD pipelines. It also contains containerization and deployment automation setups for Jenkins, AWS CodeBuild, CodePipeline, and CodeCommit alongside GitHub Actions and ArgoCD.

2. **[mallhive-security](https://github.com/gikewesi/mallhive-security)**
   Focuses on security tooling and policies such as Kubernetes RBAC, network policies, IAM roles and policies, AWS Security Hub integrations, secrets management, and security scanning automation. This repo centralizes all security best practices applied across the platform.

3. **[mallhive-monitoring](https://github.com/gikewesi/mallhive-monitoring)**
   Houses observability components including Prometheus configurations, Grafana dashboards, alerting rules, AWS CloudWatch integration, OpenTelemetry instrumentation, and log aggregation setups. It covers application and infrastructure monitoring and alerting.

---

## 🏗️ Infrastructure

* AWS EKS with Fargate profiles for serverless container management
* Managed PostgreSQL via AWS RDS
* AWS S3 and CloudFront for static assets and CDN
* Application Load Balancers for ingress
* AWS Secrets Manager and KMS for secure secrets and encryption
* Terraform as IaC for reproducible provisioning

---

## 🔄 CI/CD Pipeline

MallHive’s CI/CD workflows are powered by multiple tools working together for reliability and flexibility:

* **Jenkins**: Orchestrates build and deployment pipelines, integrates with testing tools, and handles complex workflow automation.
* **AWS CodeBuild, CodePipeline, CodeCommit**: AWS-native tools for source control, build, and deployment automation integrated with the rest of the AWS ecosystem.
* **GitHub Actions**: Runs unit and integration tests, builds Docker images tagged by git commit hashes, and pushes to Amazon ECR.
* **ArgoCD**: Monitors Git repositories and deploys Kubernetes manifests to EKS clusters with automated rollbacks and health checks.

---

## 🔐 Security

* Least-privilege IAM roles with scoped permissions
* Kubernetes RBAC and Network Policies for pod-level isolation
* Encrypted secrets storage with automatic rotation
* Mutual TLS for service-to-service authentication
* Continuous security validation via tooling in the mallhive-security repo

---

## 📊 Monitoring & Observability

* Centralized logs collected with Fluent Bit to CloudWatch
* Metrics aggregated by Prometheus and visualized with Grafana
* Distributed tracing with OpenTelemetry
* Alerts configured for critical service failures and performance degradation

---

## 🧩 Micro Frontends

Frontend is broken into independently deployable  micro frontends covering key user flows like homepage, product browsing, cart, checkout, and user profile, enabling agile front-end development.

---

## 🚀 Deployment Roadmap

Phased rollout starting with core services (authentication, product catalog, cart), followed by order processing, payments, notifications, search, analytics, and finally frontend microservices with end-to-end testing.

---

## 🤝 Contributing

Contributions are welcome!

* Fork the repo
* Create feature branches
* Submit pull requests with clear descriptions and tests
