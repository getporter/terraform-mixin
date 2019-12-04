resource "azurerm_resource_group" "test" {
    name     = var.resource_group_name
    location = var.location
}

resource "azurerm_key_vault" "test" {
    name                = var.keyvault_name
    location            = azurerm_resource_group.test.location
    resource_group_name = azurerm_resource_group.test.name

    enabled_for_disk_encryption = true
    tenant_id                   = var.tenant_id

    sku_name = "standard"

    access_policy {
        tenant_id = var.tenant_id
        object_id = var.client_id

        key_permissions = [
        "get",
        ]

        secret_permissions = [
        "get",
        ]

        storage_permissions = [
        "get",
        ]
    }

    network_acls {
        default_action = "Deny"
        bypass         = "AzureServices"
    }

    tags = {
        environment = "Production"
    }
}
