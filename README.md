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
Or
```yaml
mixins:
- terraform:
    clientVersion: 1.0.3
    workingDirs:
      - infra1
      - infra2
    initFile: providers.tf
```

### clientVersion
The Terraform client version can be specified via the `clientVersion` configuration when declaring this mixin.

### workingDir
The `workingDir` configuration setting is the relative path to your terraform files. Defaults to "terraform".

### workingDirs
The `workingDirs` configuraiton setting is used when multiple terraform plans are part of a single bundle. When the `workingDirs` setting is specified then the `workingDir` setting is ignored.

### initFile
Terraform providers are installed into the bundle during porter build. 
We recommend that you put your provider declarations into a single file, e.g. "terraform/providers.tf".
Then use `initFile` to specify the relative path to this file within workingDir.
This will dramatically improve Docker image layer caching and performance when building, publishing and installing the bundle.
If `workingDirs` is specified instead of `workingDir` then the `initFile` must be the same in all of the terraform plans for the bundle.
> Note: this approach isn't suitable when using terraform modules as those need to be "initilized" as well but aren't specified in the `initFile`. You shouldn't specifiy an `initFile` in this situation.

### User Agent Opt Out

When you declare the mixin, you can disable the mixin from customizing the azure user agent string

```yaml
mixins:
- terraform:
    userAgentOptOut: true
```

By default, the terraform mixin adds the porter and mixin version to the user agent string used by the azure provider.
We use this to understand which version of porter and the mixin are being used by a bundle, and assist with troubleshooting.
Below is an example of what the user agent string looks like:

```
AZURE_HTTP_USER_AGENT="getporter/porter/v1.0.0 getporter/terraform/v1.2.3"
```

You can add your own custom strings to the user agent string by editing your [template Dockerfile] and setting the AZURE_HTTP_USER_AGENT environment variable.

[template Dockerfile]: https://getporter.org/bundle/custom-dockerfile/

## Terraform state

### Let Porter do the heavy lifting

The simplest way to use this mixin with Porter is to let Porter track the Terraform [state](https://www.terraform.io/docs/state/index.html) as actions are executed.  This can be done via the state section:

```yaml
state:
  - name: tfstate
    path: terraform/terraform.tfstate
  - name: tfvars
    path: terraform/terraform.tfvars.json
```

The [TabbyCats Tracker bundle](https://github.com/carolynvs/tabbycat-demo) is a good example of how to use the terraform mixin with Porter v1.

The specified path inside the installer (`/cnab/app/terraform/terraform.tfstate`) should be where Terraform will be looking to read/write its state.  For a full example bundle using this approach, see the [basic-tf-example](examples/basic-tf-example).

Any arbitrary file can be added to the state including any files created by terraform during install or upgrade.

When working with multiple different terraform plans in the same bundle make sure to specify the path to the corresponding plans state:

```yaml
state:
  - name: infra1-tfstate
    path: infra1/terraform.tfstate
  - name: infra1-tfvars
    path: infra1/terraform.tfvars.json
  - name: infra1-file
    path: infra1/infra1-file
  - name: infra2-tfstate
    path: infra2/terraform.tfstate
  - name: infra2-tfvars
    path: infra2/terraform.tfvars.json
  - name: infra2-file
    path: infra2/infra2-file
```

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


## Multiple Terraform Plans In A Single Bundle

Multiple terraform plans can be specified for a single bundle. When using the mixin with this configuration then every step **MUST** include a `workingDir` configuration setting so that porter can resolve the corresponding plan for that step at runtime. 

The `workingDir` and `workingDirs` configuration settings are mutally exclusive. If the `workingDirs` configuration setting is provided then anything set for `workingDir` will be ignored at bundle build time.