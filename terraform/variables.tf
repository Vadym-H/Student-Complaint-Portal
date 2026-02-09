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