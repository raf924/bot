package domain

type Command struct {
	name    string
	aliases []string
	usage   string
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Aliases() []string {
	return c.aliases
}

func (c *Command) Usage() string {
	return c.usage
}

func NewCommand(name string, aliases []string, usage string) *Command {
	return &Command{name: name, aliases: aliases, usage: usage}
}
