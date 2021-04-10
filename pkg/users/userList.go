package users

import (
	"github.com/raf924/connector-api/pkg/gen"
	"strings"
)

type UserList struct {
	users       []*gen.User
	userIndexes map[string]int
}

func (l *UserList) All() []*gen.User {
	var list = make([]*gen.User, len(l.users))
	for _, u := range l.users {
		list = append(list, &gen.User{
			Nick:  u.GetNick(),
			Id:    u.GetId(),
			Mod:   u.GetMod(),
			Admin: u.GetAdmin(),
		})
	}
	return list
}

func (l *UserList) Get(i int) *gen.User {
	return l.users[i]
}

func (l *UserList) Find(nick string) *gen.User {
	i, ok := l.userIndexes[nick]
	if !ok {
		return nil
	}
	return l.users[i]
}

func (l *UserList) Add(user *gen.User) {
	if len(strings.TrimSpace(user.GetNick())) == 0 {
		return
	}
	l.users = append(l.users, user)
	l.userIndexes[user.GetNick()] = len(l.users) - 1
}

func (l *UserList) Remove(user *gen.User) {
	i, ok := l.userIndexes[user.GetNick()]
	if !ok {
		return
	}
	l.users = append(l.users[:i], l.users[i+1:]...)
}

func NewUserList(users ...*gen.User) *UserList {
	ul := &UserList{
		users:       []*gen.User{},
		userIndexes: map[string]int{},
	}
	for _, user := range users {
		ul.Add(user)
	}
	return ul
}
