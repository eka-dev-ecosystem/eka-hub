# ==============================================================================
# Eka-Hub Kubernetes Infrastructure as Code (IaC) - Terraform Module
# ==============================================================================

# 1. Namespace
resource "kubernetes_namespace" "eka_hub" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/name"    = "eka-hub"
      "app.kubernetes.io/part-of" = "eka-ecosystem"
    }
  }
}

# 2. ConfigMap
resource "kubernetes_config_map" "backend_config" {
  metadata {
    name      = "${var.app_name}-config"
    namespace = kubernetes_namespace.eka_hub.metadata[0].name
    labels = {
      "app.kubernetes.io/name" = var.app_name
    }
  }

  data = {
    APP_ENVIRONMENT         = var.environment
    SERVER_HOST             = "0.0.0.0"
    SERVER_PORT             = tostring(var.server_port)
    LOG_LEVEL               = "info"
    FEATURE_REDIS           = "true"
    FEATURE_RABBITMQ        = "false"
    FEATURE_REALTIME        = "true"
    FEATURE_REALTIME_SSE    = "true"
    FEATURE_REALTIME_WS     = "true"
    FEATURE_REALTIME_BRIDGE = "redis"

    DATABASE_HOST    = var.database_host
    DATABASE_PORT    = tostring(var.database_port)
    DATABASE_USER    = var.database_user
    DATABASE_DBNAME  = var.database_name
    DATABASE_SSLMODE = var.database_sslmode

    REDIS_HOST = var.redis_host
    REDIS_PORT = tostring(var.redis_port)
    REDIS_DB   = "0"

    FTRACKER_API_BASE_URL        = var.ftracker_api_base_url
    TASK_MANAGEMENT_API_BASE_URL = var.task_management_api_base_url
  }
}

# 3. Secret
resource "kubernetes_secret" "backend_secret" {
  metadata {
    name      = "${var.app_name}-secret"
    namespace = kubernetes_namespace.eka_hub.metadata[0].name
    labels = {
      "app.kubernetes.io/name" = var.app_name
    }
  }

  type = "Opaque"

  data = {
    DATABASE_PASSWORD = var.database_password
    REDIS_PASSWORD    = var.redis_password
    JWT_SECRET        = var.jwt_secret
    GEMINI_API_KEY    = var.gemini_api_key
  }
}

# 4. Deployment
resource "kubernetes_deployment" "backend" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.eka_hub.metadata[0].name
    labels = {
      "app.kubernetes.io/name"      = var.app_name
      "app.kubernetes.io/component" = "backend"
      "app.kubernetes.io/part-of"   = "eka-ecosystem"
    }
  }

  spec {
    replicas               = var.replicas
    revision_history_limit = 5

    strategy {
      type = "RollingUpdate"
      rolling_update {
        max_surge       = "1"
        max_unavailable = "0"
      }
    }

    selector {
      match_labels = {
        "app.kubernetes.io/name" = var.app_name
      }
    }

    template {
      metadata {
        labels = {
          "app.kubernetes.io/name"      = var.app_name
          "app.kubernetes.io/component" = "backend"
        }
      }

      spec {
        container {
          name              = "backend"
          image             = "${var.image_name}:${var.image_tag}"
          image_pull_policy = "Always"

          port {
            name           = "http"
            container_port = var.server_port
            protocol       = "TCP"
          }

          env_from {
            config_map_ref {
              name = kubernetes_config_map.backend_config.metadata[0].name
            }
          }

          env_from {
            secret_ref {
              name = kubernetes_secret.backend_secret.metadata[0].name
            }
          }

          liveness_probe {
            http_get {
              path = "/health"
              port = "http"
            }
            initial_delay_seconds = 15
            period_seconds        = 10
            timeout_seconds       = 3
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/health"
              port = "http"
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 2
          }

          resources {
            requests = {
              cpu    = "100m"
              memory = "128Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "512Mi"
            }
          }
        }
      }
    }
  }
}

# 5. Service
resource "kubernetes_service" "backend" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.eka_hub.metadata[0].name
    labels = {
      "app.kubernetes.io/name"      = var.app_name
      "app.kubernetes.io/component" = "backend"
    }
  }

  spec {
    type = "ClusterIP"

    port {
      name        = "http"
      port        = var.server_port
      target_port = "http"
      protocol    = "TCP"
    }

    selector = {
      "app.kubernetes.io/name" = var.app_name
    }
  }
}

# 6. Ingress
resource "kubernetes_ingress_v1" "backend" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.eka_hub.metadata[0].name
    labels = {
      "app.kubernetes.io/name" = var.app_name
    }
    annotations = {
      "cert-manager.io/cluster-issuer"                   = "letsencrypt"
      "traefik.ingress.kubernetes.io/router.entrypoints" = "web,websecure"
    }
  }

  spec {
    ingress_class_name = "traefik"

    tls {
      hosts       = [var.domain_name]
      secret_name = "${var.app_name}-tls"
    }

    rule {
      host = var.domain_name

      http {
        path {
          path      = "/"
          path_type = "Prefix"

          backend {
            service {
              name = kubernetes_service.backend.metadata[0].name
              port {
                number = var.server_port
              }
            }
          }
        }
      }
    }
  }
}

# 7. Cloudflare DNS Record
resource "cloudflare_dns_record" "backend" {
  count   = var.cloudflare_enabled && var.cloudflare_api_token != null ? 1 : 0
  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  type    = "A"
  content = var.cloudflare_ip
  proxied = var.cloudflare_proxied
  ttl     = 1
}

