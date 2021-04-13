package connector

import "github.com/raf924/bot/pkg/queue"

var ToConnectionQueue = queue.NewQueue(queue.WithMaxProducers(1))
var FromConnectionQueue = queue.NewQueue()
var FromBotQueue = queue.NewQueue(queue.WithMaxConsumers(1))
var ToBotQueue = queue.NewQueue(queue.WithMaxProducers(1))

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
