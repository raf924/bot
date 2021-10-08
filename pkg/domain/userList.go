package domain

import (
	"strings"
	"sync"
)

var _ UserList = (*immutableUserList)(nil)
var _ UserList = (*userList)(nil)

type immutableUserList struct {
	UserList
}

func (il *immutableUserList) Get(i int) *User {
	user := il.UserList.Get(i)
	return &User{
		nick:     user.nick,
		id:       user.id,
		role:     user.role,
		joinedAt: user.joinedAt,
	}
}

func (il *immutableUserList) Add(user *User) {
	panic("cannot modify list")
}

func (il *immutableUserList) Remove(user *User) {
	panic("cannot modify list")
}

func ImmutableUserList(ul UserList) UserList {
	return &immutableUserList{UserList: ul}
}

type UserList interface {
	Copy() UserList
	All() []*User
	Get(i int) *User
	Find(nick string) *User
	Add(user *User)
	Remove(user *User)
}

type userList struct {
	rwm         *sync.RWMutex
	users       []*User
	userIndexes map[string]int
}

func (l *userList) Copy() UserList {
	return NewUserList(l.All()...)
}

func (l *userList) All() []*User {
	l.rwm.RLock()
	var list = make([]*User, 0, len(l.userIndexes))
	for _, u := range l.users {
		list = append(list, &User{
			nick:     u.nick,
			id:       u.id,
			role:     u.role,
			joinedAt: u.joinedAt,
		})
	}
	l.rwm.RUnlock()
	return list
}

func (l *userList) Get(i int) *User {
	l.rwm.RLock()
	user := l.users[i]
	l.rwm.RUnlock()
	return &User{
		nick:     user.nick,
		id:       user.id,
		role:     user.role,
		joinedAt: user.joinedAt,
	}
}

func (l *userList) Find(nick string) *User {
	l.rwm.RLock()
	user := func() *User {
		i, ok := l.userIndexes[nick]
		if !ok {
			return nil
		}
		return l.users[i]
	}()
	l.rwm.RUnlock()
	return user
}

func (l *userList) Add(user *User) {
	l.rwm.Lock()
	func() {
		if len(strings.TrimSpace(user.Nick())) == 0 {
			return
		}
		l.users = append(l.users, user)
		l.userIndexes[user.Nick()] = len(l.users) - 1
	}()
	l.rwm.Unlock()
}

func (l *userList) Remove(user *User) {
	l.rwm.Lock()
	i, ok := l.userIndexes[user.Nick()]
	func() {
		if !ok {
			return
		}
		l.users = append(l.users[:i], l.users[i+1:]...)
		delete(l.userIndexes, user.Nick())
		for j, user := range l.users[i:] {
			l.userIndexes[user.Nick()] = i + j
		}
	}()
	l.rwm.Unlock()
}

func NewUserList(users ...*User) UserList {
	ul := &userList{
		users:       make([]*User, 0, len(users)),
		userIndexes: map[string]int{},
		rwm:         &sync.RWMutex{},
	}
	for _, user := range users {
		ul.Add(user)
	}
	return ul
}
