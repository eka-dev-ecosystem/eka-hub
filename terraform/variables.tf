variable "kubeconfig_path" {
  description = "Path to the kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kubeconfig_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = null
}

variable "namespace" {
  description = "Kubernetes namespace for Eka-Hub"
  type        = string
  default     = "eka-hub"
}

variable "environment" {
  description = "Deployment environment (production, staging, development)"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "Application name for the backend workload"
  type        = string
  default     = "eka-hub-backend"
}

variable "image_name" {
  description = "Container image repository"
  type        = string
  default     = "ghcr.io/exa31/eka-hub-backend"
}

variable "image_tag" {
  description = "Container image tag"
  type        = string
  default     = "latest"
}

variable "domain_name" {
  description = "Domain name for the backend API"
  type        = string
  default     = "hub-api.eka-dev.cloud"
}

variable "replicas" {
  description = "Number of backend pod replicas"
  type        = number
  default     = 1
}

variable "server_port" {
  description = "HTTP port exposed by the backend container"
  type        = number
  default     = 3000
}

variable "database_host" {
  description = "PostgreSQL database hostname"
  type        = string
  default     = "postgres.default.svc.cluster.local"
}

variable "database_port" {
  description = "PostgreSQL database port"
  type        = number
  default     = 5432
}

variable "database_user" {
  description = "PostgreSQL username"
  type        = string
  default     = "postgres"
}

variable "database_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "eka_hub_db"
}

variable "database_password" {
  description = "PostgreSQL database password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "database_sslmode" {
  description = "PostgreSQL SSL mode (disable, require, verify-ca, verify-full)"
  type        = string
  default     = "disable"
}

variable "redis_host" {
  description = "Redis cache hostname"
  type        = string
  default     = "redis.default.svc.cluster.local"
}

variable "redis_port" {
  description = "Redis port"
  type        = number
  default     = 6379
}

variable "redis_password" {
  description = "Redis password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "jwt_secret" {
  description = "Secret key for JWT signature and token verification"
  type        = string
  default     = ""
  sensitive   = true
}

variable "gemini_api_key" {
  description = "Google Gemini API key for AI Briefing & Function Calling"
  type        = string
  default     = ""
  sensitive   = true
}

variable "ftracker_api_base_url" {
  description = "Base URL for F-Tracker API integration"
  type        = string
  default     = "https://f-tracker.eka-dev.cloud"
}

variable "task_management_api_base_url" {
  description = "Base URL for Task Management API integration"
  type        = string
  default     = "https://task-management.eka-dev.cloud"
}
