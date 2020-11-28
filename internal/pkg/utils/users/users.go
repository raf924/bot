package users

import "github.com/raf924/bot/api/messages"

func Equal(user *messages.User, user2 *messages.User) bool {
	sameNick := user.Nick == user2.Nick
	sameId := user.Id == user2.Id
	return sameId && sameNick
}

func Same(user *messages.User, user2 *messages.User) bool {
	if user.Id == "" && user2.Id == "" {
		return user.Nick == user2.Nick
	}
	return user.Id == user2.Id
}
