module get.porter.sh/mixin/terraform

go 1.13

require (
	get.porter.sh/porter v0.27.2
	github.com/cnabio/cnab-go v0.13.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	github.com/xeipuuv/gojsonschema v1.2.0
	gopkg.in/yaml.v2 v2.2.4
)

replace github.com/hashicorp/go-plugin => github.com/carolynvs/go-plugin v1.0.1-acceptstdin
