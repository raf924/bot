package bot

import "github.com/raf924/bot/pkg/queue"

var toRelayQueue = queue.NewQueue()
var fromBotQueue = toRelayQueue
var fromRelayQueue = queue.NewQueue()
var toBotQueue = fromRelayQueue

var WithRelayExchange, _ = queue.NewExchange(toRelayQueue, fromRelayQueue)
var WithBotExchange, _ = queue.NewExchange(toBotQueue, fromBotQueue)
