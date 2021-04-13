package connector

type Config struct {
	Name       string                 `yaml:"name"`
	Bot        map[string]interface{} `yaml:"bot"`
	Connection map[string]interface{} `yaml:"connection"`
	Trigger    string                 `yaml:"trigger"`
}
