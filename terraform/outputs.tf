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