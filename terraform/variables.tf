variable "project_name" {
  description = "Project name prefix"
  type        = string
  default     = "complaintportal"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "Poland Central"
}

variable "jwt_secret" {
  description = "JWT secret key for token signing (minimum 32 characters - keep this secret!)"
  type        = string
  sensitive   = true
  validation {
    condition     = length(var.jwt_secret) >= 32
    error_message = "JWT secret must be at least 32 characters long."
  }
}

variable "custom_domain" {
  description = "Custom domain for the app (e.g., complaints.yourdomain.com). Leave empty to skip."
  type        = string
  default     = ""
}

variable "app_service_sku" {
  description = "App Service Plan SKU (B1=small/dev, B2=medium, B3=large, P1V2=premium)"
  type        = string
  default     = "B1"

  validation {
    condition     = contains(["B1", "B2", "B3", "P1V2", "P2V2", "P3V2"], var.app_service_sku)
    error_message = "SKU must be B1, B2, B3, or a Premium tier (P1V2, P2V2, P3V2)."
  }
}
