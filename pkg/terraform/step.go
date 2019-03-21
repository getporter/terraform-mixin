package terraform

type Step struct {
	Description string       `yaml:"description"`
	Outputs     []terraformOutput `yaml:"outputs"`
}

type terraformOutput struct {
	Name   string `yaml:"name"`
	Secret string `yaml:"secret"`
	Key    string `yaml:"key"`
}
