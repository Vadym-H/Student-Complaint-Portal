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
  partition_key_paths = ["/id"]  // partition_key_paths = ["/email"]
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
  sku                 = "Basic"  # Cheapest tier, perfect for your needs

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Queue 1: For new complaints
resource "azurerm_servicebus_queue" "new_complaints" {
  name         = "new-complaints"
  namespace_id = azurerm_servicebus_namespace.main.id

  # Messages stay in queue for 14 days if not processed
  default_message_ttl = "P14D"

  # Dead letter queue enabled (for failed messages)
  dead_lettering_on_message_expiration = true

  # Max size: 1GB
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