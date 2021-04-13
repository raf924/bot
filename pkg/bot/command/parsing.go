package command

import (
	"sort"
)

func Is(possibleCommand string, cmd Command) bool {
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
