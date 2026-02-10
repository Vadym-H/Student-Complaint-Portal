output "resource_group_name" {
  value       = azurerm_resource_group.main.name
  description = "Resource group name"
}

output "cosmos_endpoint" {
  value       = azurerm_cosmosdb_account.main.endpoint
  description = "Cosmos DB endpoint URL"
}

output "cosmos_primary_key" {
  value       = azurerm_cosmosdb_account.main.primary_key
  description = "Cosmos DB primary key"
  sensitive   = true
}

output "cosmos_database" {
  value       = azurerm_cosmosdb_sql_database.main.name
  description = "Cosmos DB database name"
}

output "servicebus_namespace" {
  value       = azurerm_servicebus_namespace.main.name
  description = "Service Bus namespace name"
}

output "servicebus_connection_string" {
  value       = azurerm_servicebus_namespace.main.default_primary_connection_string
  description = "Service Bus connection string"
  sensitive   = true
}

output "queue_new_complaints" {
  value       = azurerm_servicebus_queue.new_complaints.name
  description = "Queue for new complaints"
}

output "queue_status_changed" {
  value       = azurerm_servicebus_queue.status_changed.name
  description = "Queue for status changes"
}

# ============================================================================
# App Service Outputs
# ============================================================================
output "app_service_name" {
  value       = azurerm_linux_web_app.main.name
  description = "App Service name (used for monitoring, scaling, etc.)"
}

output "app_service_default_hostname" {
  value       = azurerm_linux_web_app.main.default_hostname
  description = "Default App Service hostname (format: <name>.azurewebsites.net)"
}

output "app_service_default_url" {
  value       = "https://${azurerm_linux_web_app.main.default_hostname}"
  description = "Full HTTPS URL to the application (default hostname)"
}

output "app_service_custom_domain_url" {
  value       = var.custom_domain != "" ? "https://${var.custom_domain}" : "N/A (custom domain not configured)"
  description = "Full HTTPS URL to the application (custom domain if configured)"
}

# ============================================================================
# Container Registry Outputs
# ============================================================================
output "container_registry_login_server" {
  value       = azurerm_container_registry.main.login_server
  description = "Container Registry login server (used to push images)"
}

output "container_registry_name" {
  value       = azurerm_container_registry.main.name
  description = "Container Registry name"
}

output "container_registry_username" {
  value       = azurerm_container_registry.main.admin_username
  description = "Container Registry admin username (for docker login)"
  sensitive   = false  # Username is not secret, but shown separately
}

output "container_registry_password" {
  value       = azurerm_container_registry.main.admin_password
  description = "Container Registry admin password (KEEP THIS SECRET)"
  sensitive   = true
}
