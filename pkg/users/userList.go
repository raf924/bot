package users

import (
	"github.com/raf924/connector-api/pkg/gen"
	"strings"
	"sync"
)

type UserList struct {
	rwm         *sync.RWMutex
	users       []*gen.User
	userIndexes map[string]int
}

func (l *UserList) Copy() *UserList {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	return NewUserList(l.All()...)
}

func (l *UserList) All() []*gen.User {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
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
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	return l.users[i]
}

func (l *UserList) Find(nick string) *gen.User {
	l.rwm.RLock()
	defer l.rwm.RUnlock()
	i, ok := l.userIndexes[nick]
	if !ok {
		return nil
	}
	return l.users[i]
}

func (l *UserList) Add(user *gen.User) {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	if len(strings.TrimSpace(user.GetNick())) == 0 {
		return
	}
	l.users = append(l.users, user)
	l.userIndexes[user.GetNick()] = len(l.users) - 1
}

func (l *UserList) Remove(user *gen.User) {
	l.rwm.Lock()
	defer l.rwm.Unlock()
	i, ok := l.userIndexes[user.GetNick()]
	if !ok {
		return
	}
	l.users = append(l.users[:i], l.users[i+1:]...)
	delete(l.userIndexes, user.GetNick())
	for j, user := range l.users[i:] {
		l.userIndexes[user.GetNick()] = i + j
	}
}

func NewUserList(users ...*gen.User) *UserList {
	ul := &UserList{
		users:       []*gen.User{},
		userIndexes: map[string]int{},
		rwm:         &sync.RWMutex{},
	}
	for _, user := range users {
		ul.Add(user)
	}
	return ul
}
