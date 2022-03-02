# Terraform Mixin for Porter

This is a Terraform mixin for [Porter](https://porter.sh).

[![Build Status](https://dev.azure.com/getporter/porter/_apis/build/status/terraform-mixin?branchName=main)](https://dev.azure.com/getporter/porter/_build/latest?definitionId=10&branchName=main)

<img src="https://porter.sh/images/mixins/terraform.svg" align="right" width="150px"/>

## Install via Porter

This will install the latest mixin release via the Porter CLI.

```
porter mixin install terraform
```

## Build from source

Following commands build the terraform mixin.
```bash
git clone https://github.com/getporter/terraform-mixin.git
cd terraform-mixin
# Learn about Mage in our CONTRIBUTING.md
go run mage.go EnsureMage
mage build
```

Then, to install the resulting mixin into PORTER_HOME, execute
`mage install`

## Mixin Configuration

```yaml
mixins:
- terraform:
    clientVersion: 1.0.3
    workingDir: myinfra
    initFile: providers.tf
```

### clientVersion
The Terraform client version can be specified via the `clientVersion` configuration when declaring this mixin.

### workingDir
The `workingDir` configuration setting is the relative path to your terraform files. Defaults to "terraform".

### initFile
Terraform providers are installed into the bundle during porter build. 
We recommend that you put your provider declarations into a single file, e.g. "terraform/providers.tf".
Then use `initFile` to specify the relative path to this file within workingDir.
This will dramatically improve Docker image layer caching and performance when building, publishing and installing the bundle.

## Terraform state

### Let Porter do the heavy lifting

The simplest way to use this mixin with Porter is to let Porter track the Terraform [state](https://www.terraform.io/docs/state/index.html) as actions are executed.  This can be done via a parameter of type `file` that has a source of a corresponding output (of the same `file` type).  Each time the bundle is executed, the output will capture the updated state file and inject it into the next action via its parameter correlate.

Here is an example setup that works with Porter v0.38:

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

If you are working with the Porter v1 prerelease, use the new state section:

```yaml
state:
  - name: tfstate
    path: terraform/terraform.tfstate
  - name: tfvars
    path: terraform/terraform.tfvars.json
```

The [TabbyCats Tracker bundle](https://github.com/carolynvs/tabbycat-demo) is a good example of how to use the terraform mixin with the Porter v1 prerelease.

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
from the state file for use by Porter. Outputs can be saved to the filesystem so that subsequent
steps can use the file by specifying the `destinationFile` field. This is particularly useful
when your terraform module creates a Kubernetes cluster. In the example below, the module
creates a cluster, and then writes the kubeconfig to /root/.kube/config so that the rest of the
bundle can immediately use the cluster.

```yaml
install:
  - terraform:
      description: "Create a Kubernetes cluster"
      outputs:
      - name: kubeconfig
        destinationFile: /root/.kube/config
```

See the Porter [Outputs documentation](https://porter.sh/wiring/#outputs) on how to wire up
outputs for use in a bundle.
