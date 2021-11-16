package domain

import "time"

type UserRole string

const (
	RegularUser UserRole = "RegularUser"
	Admin       UserRole = "Admin"
	Moderator   UserRole = "Mod"
)

type User struct {
	nick     string
	id       string
	role     UserRole
	joinedAt *time.Time
}

func (u *User) Is(user *User) bool {
	return u.nick == user.nick && u.id == user.id
}

type Users []*User

func (u Users) Len() int {
	return len(u)
}

func (u Users) Less(i, j int) bool {
	if u[i].joinedAt == nil || u[j].joinedAt == nil {
		return false
	}
	return u[i].joinedAt.Before(*u[j].joinedAt)
}

func (u Users) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func NewUser(nick string, id string, role UserRole) *User {
	return &User{nick: nick, id: id, role: role}
}

func NewOnlineUser(nick string, id string, role UserRole, joinedAt time.Time) *User {
	return &User{nick: nick, id: id, role: role, joinedAt: &joinedAt}
}

func (u *User) Nick() string {
	return u.nick
}

func (u *User) Id() string {
	return u.id
}

func (u *User) Role() UserRole {
	return u.role
}

func (u *User) JoinedAt() *time.Time {
	return u.joinedAt
}
