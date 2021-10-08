package command

import (
	"github.com/raf924/bot/pkg/domain"
	"sort"
)

func Is(possibleCommand string, cmd *domain.Command) bool {
	if possibleCommand == cmd.Name() {
		return true
	}
	aliases := cmd.Aliases()
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

func Find(commands []*domain.Command, possibleCommand string) *domain.Command {
	for _, cmd := range commands {
		if Is(possibleCommand, cmd) {
			return cmd
		}
	}
	return nil
}
