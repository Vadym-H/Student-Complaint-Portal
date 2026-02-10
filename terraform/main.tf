terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

provider "azurerm" {
  features {}
}

# Resource Group
resource "azurerm_resource_group" "main" {
  name     = "${var.project_name}-${var.environment}-rg"
  location = var.location

  tags = {
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "Terraform"
  }
}

# Random suffix for unique naming
resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

# Cosmos DB Account
resource "azurerm_cosmosdb_account" "main" {
  name                = "${var.project_name}-db-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  offer_type          = "Standard"
  kind                = "GlobalDocumentDB"

  # Serverless mode (cheaper for development)
  capabilities {
    name = "EnableServerless"
  }

  consistency_policy {
    consistency_level = "Session"
  }

  geo_location {
    location          = azurerm_resource_group.main.location
    failover_priority = 0
  }

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Cosmos DB Database
resource "azurerm_cosmosdb_sql_database" "main" {
  name                = "complaintportal"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main.name
}

# Container: users
resource "azurerm_cosmosdb_sql_container" "users" {
  name                = "users"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = azurerm_cosmosdb_sql_database.main.name
  partition_key_paths = ["/id"]
}

# Container: complaints
resource "azurerm_cosmosdb_sql_container" "complaints" {
  name                = "complaints"
  resource_group_name = azurerm_resource_group.main.name
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = azurerm_cosmosdb_sql_database.main.name
  partition_key_paths = ["/userId"]
}

# Service Bus Namespace
resource "azurerm_servicebus_namespace" "main" {
  name                = "${var.project_name}-bus-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Basic"

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Queue 1: For new complaints
resource "azurerm_servicebus_queue" "new_complaints" {
  name         = "new-complaints"
  namespace_id = azurerm_servicebus_namespace.main.id

  default_message_ttl = "P14D"
  dead_lettering_on_message_expiration = true
  max_size_in_megabytes = 1024
}

# Queue 2: For complaint status changes
resource "azurerm_servicebus_queue" "status_changed" {
  name         = "complaint-status-changed"
  namespace_id = azurerm_servicebus_namespace.main.id

  default_message_ttl                  = "P14D"
  dead_lettering_on_message_expiration = true
  max_size_in_megabytes                = 1024
}

# ============================================================================
# Azure Container Registry - for storing Docker images
# ============================================================================
resource "azurerm_container_registry" "main" {
  name                = "${replace(var.project_name, "-", "")}acr${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Basic"

  admin_enabled = true  # CHANGED: Enable admin user for simple authentication

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# ============================================================================
# App Service Plan - compute resources for hosting the application
# ============================================================================
resource "azurerm_service_plan" "main" {
  name                = "${var.project_name}-plan-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  os_type             = "Linux"
  sku_name            = var.app_service_sku

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# ============================================================================
# Linux Web App - the actual application running in a container
# ============================================================================
resource "azurerm_linux_web_app" "main" {
  name                = "${var.project_name}-app-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  service_plan_id     = azurerm_service_plan.main.id

  # Application settings and Docker configuration
  app_settings = {
    "WEBSITES_PORT" = "8080"
    "WEBSITES_CONTAINER_START_TIME_LIMIT" = "600"

    # Docker settings
    "WEBSITES_ENABLE_APP_SERVICE_STORAGE" = "false"
    "DOCKER_REGISTRY_SERVER_URL"          = "https://${azurerm_container_registry.main.login_server}"
    "DOCKER_ENABLE_CI"                    = "true"

    # CHANGED: Using admin credentials instead of managed identity
    "DOCKER_REGISTRY_SERVER_USERNAME"     = azurerm_container_registry.main.admin_username
    "DOCKER_REGISTRY_SERVER_PASSWORD"     = azurerm_container_registry.main.admin_password

    # Application configuration
    "ENV"                    = var.environment
    "HTTP_PORT"              = "8080"
    "COSMOS_ENDPOINT"        = azurerm_cosmosdb_account.main.endpoint
    "COSMOS_KEY"             = azurerm_cosmosdb_account.main.primary_key
    "COSMOS_DATABASE"        = azurerm_cosmosdb_sql_database.main.name
    "SERVICE_BUS_CONNECTION" = azurerm_servicebus_namespace.main.default_primary_connection_string
    "JWT_SECRET"             = var.jwt_secret
  }

  site_config {
    # CHANGED: Disabled managed identity
    container_registry_use_managed_identity = false

    # Docker image configuration
    application_stack {
      docker_image     = "${azurerm_container_registry.main.login_server}/complaint-portal"
      docker_image_tag = "latest"
    }

    # Health check
    health_check_path = "/health"

    # CORS Configuration
    cors {
      allowed_origins     = ["http://localhost:4200", "https://${var.project_name}-app-${random_string.suffix.result}.azurewebsites.net"]
      support_credentials = true
    }
  }

  # HTTPS-only enforcement
  https_only = true

  # Logging
  logs {
    application_logs {
      file_system_level = "Information"
    }
    http_logs {
      file_system {
        retention_in_days = 7
        retention_in_mb   = 35
      }
    }
  }

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
  }

  depends_on = [
    azurerm_cosmosdb_account.main,
    azurerm_servicebus_namespace.main,
    azurerm_container_registry.main
  ]
}

# REMOVED: Role assignment block deleted entirely

# ============================================================================
# Custom Domain Binding (Optional)
# ============================================================================
resource "azurerm_app_service_custom_hostname_binding" "main" {
  count               = var.custom_domain != "" ? 1 : 0
  hostname            = var.custom_domain
  app_service_name    = azurerm_linux_web_app.main.name
  resource_group_name = azurerm_resource_group.main.name

  depends_on = [azurerm_linux_web_app.main]
}

# ============================================================================
# Azure-Managed SSL Certificate (Optional)
# ============================================================================
resource "azurerm_app_service_managed_certificate" "main" {
  count               = var.custom_domain != "" ? 1 : 0
  custom_hostname_binding_id = azurerm_app_service_custom_hostname_binding.main[0].id
}

# ============================================================================
# SSL Certificate Binding
# ============================================================================
resource "azurerm_app_service_certificate_binding" "main" {
  count               = var.custom_domain != "" ? 1 : 0
  hostname_binding_id = azurerm_app_service_custom_hostname_binding.main[0].id
  certificate_id      = azurerm_app_service_managed_certificate.main[0].id
  ssl_state           = "SniEnabled"
}