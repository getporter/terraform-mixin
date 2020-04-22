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

### Invoke

An invoke step is used for any custom action (not one of `install`, `upgrade` or `uninstall`).

By default, the command given to `terraform` will be the step name.  Here it is `show`,
resulting in `terraform show` with the provided configuration.

```yaml
show:
  - terraform:
      description: "Invoke 'terraform show'"
      backendConfig:
        key: "mybundle.tfstate"
        storage_account_name: "mystorageacct"
        container_name: "mycontainer"
        access_key: "myaccesskey"
```

Or, if the step name does not match the intended terraform command, the command
can be supplied via the `arguments:` section, like so:

```yaml
printVersion:
  - terraform:
      description: "Invoke 'terraform version'"
      arguments:
        - version
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

## Outputs

As seen above, outputs can be declared for a step.  All that is needed is the name of the output.

For each output listed, `terraform output <output name>` is invoked to fetch the output value
from the state file for use by Porter.

See the Porter [Outputs documentation](https://porter.sh/wiring/#outputs) on how to wire up
outputs for use in a bundle.