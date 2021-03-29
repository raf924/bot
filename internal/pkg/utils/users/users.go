package users

import messages "github.com/raf924/connector-api/pkg/gen"

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
