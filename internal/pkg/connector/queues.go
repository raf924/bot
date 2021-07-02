package connector

import "github.com/raf924/bot/pkg/queue"

var ToConnectionQueue = queue.NewQueue()
var FromConnectionQueue = queue.NewQueue()
var FromBotQueue = queue.NewQueue()
var ToBotQueue = queue.NewQueue()

func WithConnectionExchange() (*queue.Exchange, error) {
	return queue.NewExchange(ToConnectionQueue, FromConnectionQueue)
}

var Cr2CnExchange = WithConnectionExchange

func WithBotExchange() (*queue.Exchange, error) {
	return queue.NewExchange(ToBotQueue, FromBotQueue)
}

var Cr2BExchange = WithBotExchange

func ConnectionExchange() (*queue.Exchange, error) {
	return queue.NewExchange(FromConnectionQueue, ToConnectionQueue)
}

var Cn2CrExchange = ConnectionExchange

func BotExchange() (*queue.Exchange, error) {
	return queue.NewExchange(FromBotQueue, ToBotQueue)
}

var B2CrExchange = BotExchange
