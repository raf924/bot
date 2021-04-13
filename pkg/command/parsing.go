package command

import (
	"github.com/raf924/connector-api/pkg/gen"
	"sort"
)

func Is(possibleCommand string, cmd *gen.Command) bool {
	if possibleCommand == cmd.GetName() {
		return true
	}
	aliases := cmd.GetAliases()
	if len(aliases) == 0 {
		return false
	}
	sort.Strings(aliases)
	index := sort.SearchStrings(aliases, possibleCommand)
	if index < len(aliases) && aliases[index] == possibleCommand {
		return true
	}
	return false
}

func Find(commands []*gen.Command, possibleCommand string) *gen.Command {
	for _, cmd := range commands {
		if Is(possibleCommand, cmd) {
			return cmd
		}
	}
	return nil
}
