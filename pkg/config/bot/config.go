package bot

type PermissionConfig struct {
	Format   string `yaml:"format"`
	Location string `yaml:"location"`
}

type UserConfig struct {
	AllowAll    bool             `yaml:"all"`
	Permissions PermissionConfig `yaml:"permissions"`
}

type CommandConfig struct {
	Disabled    map[string]bool  `yaml:"disabled"`
	Permissions PermissionConfig `yaml:"permissions"`
}

type Config struct {
	Connector map[string]interface{} `yaml:"connector"`
	Trigger   string                 `yaml:"trigger"`
	ApiKeys   map[string]string      `yaml:"apiKeys"`
	Users     UserConfig             `yaml:"users"`
	Commands  CommandConfig          `yaml:"commands"`
}
