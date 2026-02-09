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