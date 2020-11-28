package bot

type PermissionConfig struct {
	Format   string `yaml:"format"`
	Location string `yaml:"location"`
}

type Config struct {
	Connector map[string]interface{} `yaml:"connector"`
	Bot       struct {
		Trigger string            `yaml:"trigger"`
		ApiKeys map[string]string `yaml:"apiKeys"`
		Users   struct {
			Owner       string           `yaml:"owner"`
			Permissions PermissionConfig `yaml:"permissions"`
		} `yaml:"users"`
		Commands struct {
			Disabled    map[string]bool  `yaml:"disabled"`
			Permissions PermissionConfig `yaml:"permissions"`
		} `yaml:"commands"`
	} `yaml:"bot"`
}
