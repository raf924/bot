package domain

import (
	"fmt"
	"time"
)

type ServerMessage interface {
	Timestamp() time.Time
}

type ChatMessage struct {
	recipients            []*User
	timestamp             time.Time
	message               string
	sender                *User
	mentionsConnectorUser bool
	private               bool
	incoming              bool
}

func (s *ChatMessage) Incoming() bool {
	return s.incoming
}

func (s *ChatMessage) MentionsConnectorUser() bool {
	return s.mentionsConnectorUser
}

func (s *ChatMessage) Sender() *User {
	return s.sender
}

func (s *ChatMessage) Timestamp() time.Time {
	return s.timestamp
}

func NewChatMessage(message string, sender *User, recipients []*User, mentionsConnectorUser bool, private bool, timestamp time.Time, incoming bool) *ChatMessage {
	return &ChatMessage{message: message, sender: sender, mentionsConnectorUser: mentionsConnectorUser, recipients: recipients, private: private, timestamp: timestamp, incoming: incoming}
}

func (s *ChatMessage) Message() string {
	return s.message
}

func (s *ChatMessage) Recipients() []*User {
	return s.recipients
}

func (s *ChatMessage) Private() bool {
	return s.private
}

type UserEventType string

const (
	UserJoined UserEventType = "JOINED"
	UserLeft   UserEventType = "LEFT"
)

type UserEvent struct {
	user      *User
	eventType UserEventType
	timestamp time.Time
}

func (u *UserEvent) User() *User {
	return u.user
}

func (u *UserEvent) EventType() UserEventType {
	return u.eventType
}

func (u *UserEvent) Timestamp() time.Time {
	return u.timestamp
}

func NewUserEvent(user *User, eventType UserEventType, timestamp time.Time) *UserEvent {
	return &UserEvent{user: user, eventType: eventType, timestamp: timestamp}
}

type CommandMessage struct {
	command   string
	args      []string
	argString string
	sender    *User
	private   bool
	timestamp time.Time
}

func NewCommandMessage(command string, args []string, argString string, sender *User, private bool, timestamp time.Time) *CommandMessage {
	return &CommandMessage{command: command, args: args, argString: argString, sender: sender, private: private, timestamp: timestamp}
}

func (c *CommandMessage) ToChatMessage() *ChatMessage {
	return &ChatMessage{
		message:   fmt.Sprintf("%s %s", c.command, c.argString),
		sender:    c.sender,
		private:   c.private,
		timestamp: c.timestamp,
	}
}

func (c *CommandMessage) Private() bool {
	return c.private
}

func (c *CommandMessage) Command() string {
	return c.command
}

func (c *CommandMessage) Args() []string {
	return c.args
}

func (c *CommandMessage) ArgString() string {
	return c.argString
}

func (c *CommandMessage) Sender() *User {
	return c.sender
}

func (c *CommandMessage) Timestamp() time.Time {
	return c.timestamp
}

type ClientMessage struct {
	message   string
	recipient *User
	private   bool
}

func (c *ClientMessage) Message() string {
	return c.message
}

func (c *ClientMessage) Recipient() *User {
	return c.recipient
}

func (c *ClientMessage) Private() bool {
	return c.private
}

func NewClientMessage(message string, recipient *User, private bool) *ClientMessage {
	return &ClientMessage{message: message, recipient: recipient, private: private}
}

type RegistrationMessage struct {
	commands []*Command
}

func (r *RegistrationMessage) Commands() []*Command {
	return r.commands
}

func NewRegistrationMessage(commands []*Command) *RegistrationMessage {
	return &RegistrationMessage{commands: commands}
}

type ConfirmationMessage struct {
	currentUser *User
	trigger     string
	users       UserList
}

func (c *ConfirmationMessage) CurrentUser() *User {
	return c.currentUser
}

func (c *ConfirmationMessage) Trigger() string {
	return c.trigger
}

func (c *ConfirmationMessage) Users() UserList {
	return c.users.Copy()
}

func NewConfirmationMessage(currentUser *User, trigger string, users []*User) *ConfirmationMessage {
	return &ConfirmationMessage{currentUser: currentUser, trigger: trigger, users: ImmutableUserList(NewUserList(users...))}
}
