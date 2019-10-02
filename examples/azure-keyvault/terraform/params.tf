variable "client_id" {}
variable "client_secret" {}
variable "tenant_id" {}
variable "subscription_id" {}

variable "resource_group_name" {
    default = "azure-kvtest"
}

variable location {
    default = "East US"
}

variable "keyvault_name" {}