# MallHive Cloud-Native E-Commerce Platform

## 🏗 Architecture Overview

MallHive is a cloud-native e-commerce platform built using a **microservices architecture** with **microfrontends**, the **Backend-for-Frontend (BFF)** pattern, and **event-driven communication**.

MallHive helps buyers to shop with ease, backed by **AI-powered** personalized recommendations.

Each major feature (e.g., product catalog, cart, user profile, checkout) is implemented as an isolated microservice and corresponding microfrontend. Each service is designed with strong boundaries and minimal backend dependencies, enabling independent development and deployment.

---

## 🎯 Why Microservices & Microfrontends Architecture?

I chose this architecture for the following reasons:

- **Scalability**: Each microservice and microfrontend can scale independently.
- **Modularity**: Teams can own vertical slices (UI, logic, data) without cross-team friction.
- **Cloud-native principles**: Designed to leverage AWS-native services, containers, and eventing.
- **Resilience**: Service failures are isolated; event-driven design enables graceful fallbacks.
- **Agility**: Faster deployments, CI/CD pipelines per microservice/frontend, and safer changes.

---

## 🧱 Architecture Components

### Frontend Layer – Microfrontends (hosted via S3 + CloudFront)

| Microfrontend           | Purpose                                | Communicates with (Microservice)     |
|-------------------------|----------------------------------------|---------------------------------------|
| `homepage.mallhive.com` | Product discovery and recommendations  | `product`, `recommendation`, `notification` |
| `products.mallhive.com` | Product listings and filters           | `product`, `recommendation`          |
| `user-profile.mallhive.com` | User settings and past orders     | `user`, `order`, `notification`      |
| `cart.mallhive.com`     | Manage shopping cart                   | `cart`, `product`                    |
| `checkout.mallhive.com` | Finalize purchase                      | `cart`, `order`, `payment`, `user`, `product` |

- Each microfrontend communicates **directly** with its backend counterpart (BFF pattern).
- All microfrontends operate statelessly; user sessions are authenticated via JWTs passed on each request.

---

### Backend Layer – Microservices (deployed via EKS on Fargate)

| Microservice     | Role                                         | Communicates with                |
|------------------|----------------------------------------------|----------------------------------|
| `user`           | Auth, registration, profile management       | `notification`, `analytics`     |
| `product`        | Product metadata, categories, pricing        | `analytics`, `recommendation`   |
| `cart`           | User shopping carts                          | `product`, `analytics`          |
| `order`          | Order creation and tracking                  | `payment`, `product`, `notification`, `analytics` |
| `payment`        | Payment handling and verification            | `order`, `notification`, `analytics` |
| `notification`   | Sends email, SMS, push messages              | —                                |
| `analytics`      | Aggregates metrics and events from services  | `recommendation`                |
| `recommendation` | Suggests related or personalized products    | `product`, `analytics`, `cart` |

> Most communication between services is **event-driven**, using AWS EventBridge/ SQS.

---

## 🔁 Communication Patterns

### Frontend-to-Backend (Direct via HTTPS):
Each microfrontend sends requests directly to its respective backend. For example:

````text
products.mallhive.com → product.mallhive.com/api/products
cart.mallhive.com → cart.mallhive.com/api/items
````

### Backend-to-Backend (Only Where Needed):

* **Async communication preferred** (via EventBridge or SQS)
* Minimal direct REST/gRPC calls
* Examples:

  * `order` emits `OrderCreated` → consumed by `notification`, `payment`, `analytics`
  * `user` emits `UserSignedUp` → consumed by `notification`, `analytics`
  * `cart` sends event to `product` for stock verification at checkout

---

## ⚙️ Tools & Technologies Used

### 🛠 Microservices:

* **Language**: Go (preferred), Python (for some services)
* **Containers**: Docker
* **Orchestration**: Kubernetes (AWS EKS with Fargate profiles per service)
* **Networking**: Internal ALB with host-based routing
* **API Communication**: REST and gRPC (internal)
* **Eventing**: AWS EventBridge, SNS/SQS
* **Secrets Management**: AWS Secrets Manager
* **Databases**:

  * `product`, `user`, `order`: PostgreSQL (Amazon RDS)
  * `cart`: DynamoDB (optional)
  * `analytics`: ClickHouse or S3 (batch ingestion)
  * `recommendation`: Uses OpenSearch or ML embeddings

### 📦 Microfrontends:

* **Framework**: React or Next.js
* **Deployment**: S3 + CloudFront per frontend
* **Routing**: Independent per frontend
* **Authentication**: JWT tokens, stored in localStorage or cookies (short-lived)

### 📡 Observability:

* **Metrics**: Prometheus + Grafana
* **Logs**: CloudWatch Logs
* **Tracing**: OpenTelemetry

### 🔐 Security:

* **IAM Roles for Service Accounts (IRSA)**
* **Security Groups + Network Policies**
* **HTTPS via ACM**
* **Private Route 53 for internal DNS**
* **Public/Private subnets split via NAT Gateway**

### 🚀 CI/CD:

* **Build**: GitHub Actions
* **Deploy**: Helm + Argo CD (GitOps)
* **IaC**: Terraform (modular structure)

---

## 🧩 Design Considerations

* **No tight service-to-service chains**: Communication flows through frontends or events
* **Autonomy**: Teams can deploy, scale, and develop services independently
* **Testability**: Services can be tested in isolation
* **Resilience**: One service going down doesn't cascade into others
* **Cost visibility**: Each service has isolated cost and usage metrics

---

## 📂 Supporting Repository

MallHive separates concerns by maintaining a dedicated infrastructure repository:

**[mallhive-infra](https://github.com/gikewesi/mallhive-infra)**
This repository manages all Infrastructure as Code (IaC) configurations, primarily using Terraform, to provision and maintain cloud resources such as VPCs, EKS clusters, RDS instances, networking components, and container registries. It also handles CI/CD pipelines and deployment automation using Jenkins, AWS CodeBuild, CodePipeline, CodeCommit, GitHub Actions, and ArgoCD.

In addition, it includes observability setups with Prometheus, Grafana dashboards, alerting rules, AWS CloudWatch integration, OpenTelemetry instrumentation, and log aggregation to cover both application and infrastructure monitoring.

---

## 📄 License

[MIT](LICENSE)

---

## ⚠️ Disclaimer

> **Note**: The microservices and microfrontends in this project are **not fully implemented** in terms of business logic. They serve primarily as **mock services and UI placeholders** to demonstrate cloud-native design patterns, deployment automation, and DevOps practices.
>
> The main purpose of this project is to showcase skills in **infrastructure provisioning**, **CI/CD**, **Kubernetes orchestration**, **GitOps**, **observability**, and **secure cloud-native deployments**.
>
> Developers and contributors are welcome to fork, improve, or expand the application logic to make it more production-ready.

---
