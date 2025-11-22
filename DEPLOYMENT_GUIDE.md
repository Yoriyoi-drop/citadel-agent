# Production Deployment Guide

This document provides instructions for deploying Citadel Agent in a production environment.

## Architecture Overview

Citadel Agent production deployment consists of:

1. **API Server Cluster** - Multiple instances behind load balancer
2. **Temporal Cluster** - Workflow orchestration and state management
3. **Database** - Persistent storage (PostgreSQL)
4. **Cache** - Redis for caching and session storage
5. **Load Balancer** - Distribution of traffic
6. **Monitoring** - Metrics and logging infrastructure

## Prerequisites

- Docker and Docker Compose
- Kubernetes cluster (for K8s deployment)
- Domain name and SSL certificate
- PostgreSQL database (production-grade)
- Redis instance
- Temporal server cluster

## Environment Variables

Create a `.env` file with the following variables:

```bash
# Database
DB_PASSWORD=your_secure_db_password
DB_HOST=your_postgres_host
DB_PORT=5432

# Redis
REDIS_PASSWORD=your_secure_redis_password

# JWT
JWT_SECRET=your_very_secure_jwt_secret

# Temporal (if not using internal discovery)
TEMPORAL_ADDRESS=your_temporal_host:7233
TEMPORAL_NAMESPACE=default

# API Server
PORT=3000
HOST=0.0.0.0
DEBUG=false

# Security
CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

## Docker Compose Deployment

### 1. Quick Start (Development/Testing)

```bash
cp .env.example .env
# Edit .env with your configuration
docker-compose -f docker-compose.prod.yml up -d
```

### 2. Production Deployment

```bash
# Using the production compose file
docker-compose -f docker-compose.prod.yml up -d --scale citadel-agent=3
```

## Kubernetes Deployment

### 1. Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured
- Helm (optional but recommended)

### 2. Deploy with kubectl

```bash
# Create namespace
kubectl create namespace citadel-agent

# Apply deployment
kubectl apply -f deploy/k8s/citadel-agent.yaml

# Verify deployment
kubectl get pods -n citadel-agent
kubectl get services -n citadel-agent
```

### 3. Deploy with Helm (if available)

```bash
# Add Helm repository
helm repo add citadel-agent https://your-helm-repo.com

# Install
helm install citadel-agent citadel-agent/citadel-agent \
  --namespace citadel-agent \
  --set server.replicas=3 \
  --set temporal.address=your-temporal-host:7233
```

## Configuration

### 1. Production Configuration

The production configuration is located at `config/production.yaml`. It includes:

- Server settings (timeouts, CORS, etc.)
- Temporal configuration
- Database connection
- Redis configuration
- Security settings
- Plugin manager limits
- Monitoring configuration

### 2. Environment-specific overrides

You can override configuration values using environment variables that match the YAML path:

```
server.port=4000        # overrides server.port
temporal.address=temporal.example.com:7233  # overrides temporal.address
```

## Security Considerations

### 1. Network Security

- Use SSL/TLS for all connections
- Implement proper firewall rules
- Use private networks for internal services
- Enable authentication for all services

### 2. Container Security

- Run containers as non-root user
- Use read-only root filesystem where possible
- Implement resource limits
- Enable seccomp and AppArmor profiles

### 3. Data Security

- Encrypt data at rest
- Encrypt data in transit
- Implement proper backup procedures
- Use secrets management

## Monitoring and Observability

### 1. Health Checks

- `/health` - Basic health check
- `/api/v1/engine/status` - Engine status
- `/api/v1/engine/stats` - Engine statistics
- `/metrics` - Prometheus metrics (if enabled)

### 2. Logging

- Structured JSON logs
- Request/response logging
- Error tracking
- Audit logging

### 3. Performance Monitoring

- Response times
- Throughput metrics
- Resource utilization
- Error rates

## Scaling

### 1. Horizontal Scaling

- Scale API servers based on load
- Use Kubernetes HPA or Docker Swarm scaling
- Monitor resource utilization

### 2. Vertical Scaling

- Adjust container resources as needed
- Monitor CPU and memory usage
- Use appropriate instance types

## Backup and Recovery

### 1. Database Backup

```bash
# PostgreSQL backup
pg_dump -h your-postgres-host -U citadel citadel_agent > backup.sql
```

### 2. Temporal Data

- Temporal provides its own backup mechanisms
- Follow Temporal's backup best practices
- Export workflow history for critical workflows

## Troubleshooting

### 1. Common Issues

- **API server not starting**: Check environment variables and Temporal connection
- **Workflow failures**: Check Temporal server logs
- **Performance issues**: Monitor resource usage and adjust configuration

### 2. Debug Commands

```bash
# Check container logs
docker-compose logs -f citadel-agent

# Check Kubernetes logs
kubectl logs -n citadel-agent -l app=citadel-agent-api -f

# Health check
curl http://your-domain/health
```

## Environment-specific Deployments

### 1. AWS Deployment

```bash
# Using AWS ECS/Fargate
aws ecs create-cluster --cluster-name citadel-agent
# Further AWS-specific configuration
```

### 2. Google Cloud Deployment

```bash
# Using Google Cloud Run or GKE
gcloud run deploy citadel-agent --image=your-image
# Further GCP-specific configuration
```

### 3. Azure Deployment

```bash
# Using Azure Container Instances or AKS
az container create --name citadel-agent --image=your-image
# Further Azure-specific configuration
```

## Deployment Best Practices

1. **Use Infrastructure as Code**: Terraform, CloudFormation, etc.
2. **Implement CI/CD**: Automated testing and deployment
3. **Monitor Resource Usage**: Set up alerts and monitoring
4. **Backup Regularly**: Implement automated backup schedules
5. **Security Scanning**: Scan containers for vulnerabilities
6. **Performance Testing**: Load test before production deployment
7. **Rolling Updates**: Use zero-downtime deployment strategies

## Post-Deployment Steps

1. Configure domain name and SSL certificate
2. Set up monitoring and alerting
3. Perform load testing
4. Review security configurations
5. Document the deployment process
6. Create runbooks for operations

This deployment guide provides a comprehensive approach to deploying Citadel Agent in production with security, scalability, and maintainability in mind.