package domain

import (
	"strings"
	"sync"
)

type CommandList interface {
	Copy() CommandList
	All() []*Command
	Get(i int) *Command
	Find(command string) *Command
	Add(command *Command)
}

var _ CommandList = (*immutableCommandList)(nil)

var _ CommandList = (*commandList)(nil)

type immutableCommandList struct {
	CommandList
}

func (i *immutableCommandList) Add(command *Command) {
	panic("cannot modify list")
}

func ImmutableCommandList(cl CommandList) CommandList {
	return &immutableCommandList{CommandList: cl}
}

type commandList struct {
	rwm            *sync.RWMutex
	commands       []*Command
	commandIndexes map[string]int
}

func (l *commandList) Copy() CommandList {
	return NewCommandList(l.All()...)
}

func (l *commandList) All() []*Command {
	l.rwm.RLock()
	var list = make([]*Command, len(l.commands))
	for i, c := range l.commands {
		list[i] = NewCommand(c.Name(), c.Aliases(), c.Usage())
	}
	l.rwm.RUnlock()
	return list
}

func (l *commandList) Get(i int) *Command {
	l.rwm.RLock()
	command := l.commands[i]
	l.rwm.RUnlock()
	return command
}

func (l *commandList) Find(command string) *Command {
	l.rwm.RLock()
	actualCommand := func() *Command {
		i, ok := l.commandIndexes[command]
		if !ok {
			return nil
		}
		return l.commands[i]
	}()
	l.rwm.RUnlock()
	return actualCommand
}

func (l *commandList) Add(command *Command) {
	aliases := append(command.Aliases(), command.Name())
	if len(strings.TrimSpace(command.Name())) == 0 {
		return
	}
	l.rwm.Lock()
	func() {
		l.commands = append(l.commands, command)
		for _, alias := range aliases {
			if len(strings.TrimSpace(alias)) == 0 {
				return
			}
			l.commandIndexes[alias] = len(l.commands) - 1
		}
	}()
	l.rwm.Unlock()
}

func NewCommandList(commands ...*Command) CommandList {
	ul := &commandList{
		commands:       []*Command{},
		commandIndexes: map[string]int{},
		rwm:            &sync.RWMutex{},
	}
	for _, command := range commands {
		ul.Add(command)
	}
	return ul
}
