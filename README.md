# terraform Mixin for Porter

This is a terraform mixin for [Porter](https://github.com/deislabs/porter).

## Build from source

This will get the terraform mixin and install it from source.

1. `go get -u get.porter.sh/mixin/terraform`
1. `cd $(go env GOPATH)/src/get.porter.sh/mixin/terraform`
1. `make build install`


## Examples

### Install

```yaml
install:
  - terraform:
      description: "Install Azure Key Vault"
      input: false
      backendConfig:
        key: "mybundle.tfstate"
        storage_account_name: "mystorageacct"
        container_name: "mycontainer"
        access_key: "myaccesskey"
      outputs:
      - name: vault_uri
```

### Upgrade

```yaml
upgrade:
  - terraform:
      description: "Upgrade Azure Key Vault"
      input: false
      backendConfig:
        key: "mybundle.tfstate"
        storage_account_name: "mystorageacct"
        container_name: "mycontainer"
        access_key: "myaccesskey"
      outputs:
      - name: vault_uri
```

### Uninstall

```yaml
uninstall:
  - terraform:
      description: "Uninstall Azure Key Vault"
      backendConfig:
        key: "mybundle.tfstate"
        storage_account_name: "mystorageacct"
        container_name: "mycontainer"
        access_key: "myaccesskey"
```

See further examples in the [Examples](examples) directory