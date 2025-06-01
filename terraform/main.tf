terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.0"
    }
  }

  backend "azurerm" {
    resource_group_name  = "rg-stratocraft-tfstate"
    storage_account_name = "ststratocrafttfstate"
    container_name       = "tfstate"
    key                  = "stratocraft.terraform.tfstate"
  }
}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
}

provider "azuread" {}

# Data sources
data "azurerm_client_config" "current" {}

# Variables
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "East US 2"
}

variable "domain_name" {
  description = "Custom domain name"
  type        = string
  default     = "stratocraft.dev"
}

# Local values
locals {
  project_name = "stratocraft"
  common_tags = {
    Project     = local.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
    Owner       = "StratoCraft"
  }
}

# Resource Group
resource "azurerm_resource_group" "main" {
  name     = "rg-${local.project_name}-${var.environment}"
  location = var.location
  tags     = local.common_tags
}

# Container Registry
resource "azurerm_container_registry" "main" {
  name                = "cr${local.project_name}${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  location           = azurerm_resource_group.main.location
  sku                = "Basic"
  admin_enabled      = true

  tags = local.common_tags
}

# Log Analytics Workspace
resource "azurerm_log_analytics_workspace" "main" {
  name                = "log-${local.project_name}-${var.environment}"
  location           = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                = "PerGB2018"
  retention_in_days  = 30

  tags = local.common_tags
}

# Application Insights
resource "azurerm_application_insights" "main" {
  name                = "appi-${local.project_name}-${var.environment}"
  location           = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  workspace_id       = azurerm_log_analytics_workspace.main.id
  application_type   = "web"

  tags = local.common_tags
}

# Container App Environment
resource "azurerm_container_app_environment" "main" {
  name                       = "cae-${local.project_name}-${var.environment}"
  location                  = azurerm_resource_group.main.location
  resource_group_name       = azurerm_resource_group.main.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id

  tags = local.common_tags
}

# Container App
resource "azurerm_container_app" "main" {
  name                         = "ca-${local.project_name}-${var.environment}"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name         = azurerm_resource_group.main.name
  revision_mode               = "Single"

  template {
    min_replicas = 1
    max_replicas = 3

    container {
      name   = "stratocraft-web"
      image  = "${azurerm_container_registry.main.login_server}/stratocraft:latest"
      cpu    = 0.25
      memory = "0.5Gi"

      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "PORT"
        value = "8080"
      }

      env {
        name        = "APPINSIGHTS_INSTRUMENTATIONKEY"
        secret_name = "appinsights-key"
      }
    }

    http_scale_rule {
      name                = "http-rule"
      concurrent_requests = 100
    }
  }

  secret {
    name  = "appinsights-key"
    value = azurerm_application_insights.main.instrumentation_key
  }

  ingress {
    allow_insecure_connections = false
    external_enabled          = true
    target_port              = 8080

    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  identity {
    type = "SystemAssigned"
  }

  tags = local.common_tags
}

# Role assignment for Container App to pull images from ACR
resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.main.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_container_app.main.identity[0].principal_id
}

# CDN Profile
resource "azurerm_cdn_profile" "main" {
  name                = "cdn-${local.project_name}-${var.environment}"
  location           = "Global"
  resource_group_name = azurerm_resource_group.main.name
  sku                = "Standard_Microsoft"

  tags = local.common_tags
}

# CDN Endpoint
resource "azurerm_cdn_endpoint" "main" {
  name                = "cdn-${local.project_name}-${var.environment}"
  profile_name        = azurerm_cdn_profile.main.name
  location           = azurerm_cdn_profile.main.location
  resource_group_name = azurerm_resource_group.main.name

  origin {
    name      = "stratocraft-origin"
    host_name = azurerm_container_app.main.latest_revision_fqdn
  }

  delivery_rule {
    name  = "EnforceHTTPS"
    order = 1

    request_scheme_condition {
      operator     = "Equal"
      match_values = ["HTTP"]
    }

    url_redirect_action {
      redirect_type = "Found"
      protocol      = "Https"
    }
  }

  delivery_rule {
    name  = "CacheStaticAssets"
    order = 2

    url_path_condition {
      operator     = "BeginsWith"
      match_values = ["/static/"]
    }

    cache_expiration_action {
      behavior = "Override"
      duration = "1.00:00:00"
    }
  }

  global_delivery_rule {
    cache_expiration_action {
      behavior = "Override"
      duration = "01:00:00"
    }

    cache_key_query_string_action {
      behavior = "ExcludeAll"
    }
  }

  tags = local.common_tags
}

# DNS Zone
resource "azurerm_dns_zone" "main" {
  name                = var.domain_name
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

# DNS CNAME record for CDN
resource "azurerm_dns_cname_record" "www" {
  name                = "www"
  zone_name          = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.main.name
  ttl                = 300
  record             = azurerm_cdn_endpoint.main.fqdn

  tags = local.common_tags
}

# DNS A record for apex domain (requires custom domain setup on CDN)
resource "azurerm_dns_a_record" "apex" {
  name                = "@"
  zone_name          = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.main.name
  ttl                = 300
  records            = ["104.21.0.0"] # Placeholder - will be updated after CDN custom domain setup

  tags = local.common_tags
}

# Key Vault for secrets
resource "azurerm_key_vault" "main" {
  name                = "kv-${local.project_name}-${var.environment}"
  location           = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id          = data.azurerm_client_config.current.tenant_id
  sku_name           = "standard"

  purge_protection_enabled = false

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    secret_permissions = [
      "Get",
      "List",
      "Set",
      "Delete",
      "Purge"
    ]
  }

  # Access policy for Container App
  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = azurerm_container_app.main.identity[0].principal_id

    secret_permissions = [
      "Get",
      "List"
    ]
  }

  tags = local.common_tags
}

# Monitoring - Action Group for alerts
resource "azurerm_monitor_action_group" "main" {
  name                = "ag-${local.project_name}-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  short_name          = "stratocraft"

  email_receiver {
    name          = "admin"
    email_address = "admin@stratocraft.dev"
  }

  tags = local.common_tags
}

# Metric Alert for high response time
resource "azurerm_monitor_metric_alert" "response_time" {
  name                = "High Response Time"
  resource_group_name = azurerm_resource_group.main.name
  scopes              = [azurerm_container_app.main.id]
  description         = "Alert when average response time exceeds 2 seconds"

  criteria {
    metric_namespace = "Microsoft.App/containerApps"
    metric_name      = "RequestDuration"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 2000
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  tags = local.common_tags
}

# Outputs
output "container_app_fqdn" {
  value = azurerm_container_app.main.latest_revision_fqdn
}

output "cdn_endpoint_fqdn" {
  value = azurerm_cdn_endpoint.main.fqdn
}

output "container_registry_login_server" {
  value = azurerm_container_registry.main.login_server
}

output "dns_zone_name_servers" {
  value = azurerm_dns_zone.main.name_servers
}