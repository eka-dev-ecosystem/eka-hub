output "namespace" {
  description = "The namespace where Eka-Hub is deployed"
  value       = kubernetes_namespace.eka_hub.metadata[0].name
}

output "deployment_name" {
  description = "The name of the Kubernetes Deployment"
  value       = kubernetes_deployment.backend.metadata[0].name
}

output "service_name" {
  description = "The internal cluster Service name"
  value       = kubernetes_service.backend.metadata[0].name
}

output "service_port" {
  description = "The internal port exposed by the service"
  value       = var.server_port
}

output "ingress_host" {
  description = "The public ingress host domain"
  value       = var.domain_name
}

output "api_endpoint" {
  description = "Public API URL endpoint for Eka-Hub backend"
  value       = "https://${var.domain_name}"
}

