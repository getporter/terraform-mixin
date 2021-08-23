# terraform Mixin for Porter

This is a terraform mixin for [Porter](https://github.com/getporter/porter).

[![Build Status](https://dev.azure.com/getporter/porter/_apis/build/status/terraform-mixin?branchName=main)](https://dev.azure.com/getporter/porter/_build/latest?definitionId=10&branchName=main)

<img src="https://porter.sh/images/mixins/terraform.svg" align="right" width="150px"/>

## Install via Porter

This will install the latest mixin release via the Porter CLI.

```
porter mixin install terraform --feed-url https://cdn.porter.sh/mixins/atom.xml
```

## Build from source

Following commands build the terraform mixin.
1. `go get -u get.porter.sh/mixin/terraform`
1. `cd $(go env GOPATH)/src/get.porter.sh/mixin/terraform`
1. `make build`

Then, to install the resulting mixin into PORTER_HOME, execute
`make install`

## Mixin Configuration

The Terraform client version can be specified via the `clientVersion` configuration when declaring this mixin.

```yaml
- terraform:
    clientVersion: 1.0.3
```

## Terraform state

### Let Porter do the heavy lifting

The simplest way to use this mixin with Porter is to let Porter track the Terraform [state](https://www.terraform.io/docs/state/index.html) as actions are executed.  This can be done via a parameter of type `file` that has a source of a corresponding output (of the same `file` type).  Each time the bundle is executed, the output will capture the updated state file and inject it into the next action via its parameter correlate.

Here is an example setup:

```yaml
parameters:
  - name: tfstate
    type: file
    # This designates the path within the installer to place the parameter value
    path: /cnab/app/terraform/terraform.tfstate
    # Here we tell Porter that the value for this parameter should come from the 'tfstate' output
    source:
      output: tfstate

outputs:
  - name: tfstate
    type: file
    # This designates the path within the installer to read the output from
    path: /cnab/app/terraform/terraform.tfstate
```

The specified path inside the installer (`/cnab/app/terraform/terraform.tfstate`) should be where Terraform will be looking to read/write its state.  For a full example bundle using this approach, see the [basic-tf-example](examples/basic-tf-example).

### Remote Backends

Alternatively, state can be managed by a remote backend.  When doing so, each action step needs to supply the remote backend config via `backendConfig`.  In the step examples below, the configuration has key/value pairs according to the [Azurerm](https://www.terraform.io/docs/backends/types/azurerm.html) backend.


## Terraform variables file

By default the mixin will create a default 
[`terraform.tfvars.json`](https://www.terraform.io/docs/language/values/variables.html#variable-definitions-tfvars-files)
file from the `vars` block during during the install step.

To use this file, a `tfvars` file parameter and output must be added to persist it for subsequent steps.

This can be disabled by setting `disableVarFile` to `true` during install.

Here is an example setup using the tfvar file:

```yaml
parameters:
  - name: tfvars
    type: file
    # This designates the path within the installer to place the parameter value
    path: /cnab/app/terraform/terraform.tfvars.json
    # Here we tell Porter that the value for this parameter should come from the 'tfvars' output
    source:
      output: tfvars
  - name: foo
    type: string
    applyTo:
      - install 
  - name: baz
    type: string
    default: blaz
    applyTo:
      - install 

outputs:
  - name: tfvars
    type: file
    # This designates the path within the installer to read the output from
    path: /cnab/app/terraform/terraform.tfvars.json
    
install:
  - terraform:
      description: "Install Azure Key Vault"
      vars:
        foo: bar
        baz: biz
      outputs:
      - name: vault_uri
upgrade: # No var block required
  - terraform:
      description: "Install Azure Key Vault"
      outputs:
      - name: vault_uri
uninstall: # No var block required
  - terraform:
      description: "Install Azure Key Vault"
      outputs:
      - name: vault_uri
```

and with var file disabled

```yaml
parameters:
  - name: foo
    type: string
    applyTo:
      - install 
  - name: baz
    type: string
    default: blaz
    applyTo:
      - install 

install:
  - terraform:
      description: "Install Azure Key Vault"
      disableVarFile: true
      vars:
        foo: bar
        baz: biz
      outputs:
      - name: vault_uri
uninstall: # Var block required
  - terraform:
      description: "Install Azure Key Vault"
      vars:
        foo: bar
        baz: biz
```

## Examples

### Install

```yaml
install:
  - terraform:
      description: "Install Azure Key Vault"
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

## Step Outputs

As seen above, outputs can be declared for a step.  All that is needed is the name of the output.

For each output listed, `terraform output <output name>` is invoked to fetch the output value
from the state file for use by Porter.

See the Porter [Outputs documentation](https://porter.sh/wiring/#outputs) on how to wire up
outputs for use in a bundle.
