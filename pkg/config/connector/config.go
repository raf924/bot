package connector

type Config struct {
	Connector struct {
		Name       string                 `yaml:"name"`
		Bot        map[string]interface{} `yaml:"bot"`
		Connection map[string]interface{} `yaml:"connection"`
	} `yaml:"connector"`
}
