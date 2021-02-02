package pbbot

import (
	"github.com/ProtobufBot/go-pbbot/proto_gen/onebot"
)

// HandleConnect 机器人连接
var HandleConnect = func(bot *Bot) {

}

// HandleDisconnect 机器人断开
var HandleDisconnect = func(bot *Bot) {

}

// HandlePrivateMessage 收到私聊消息
var HandlePrivateMessage = func(bot *Bot, event *onebot.PrivateMessageEvent) {

}

// HandleGroupMessage 收到群聊消息
var HandleGroupMessage = func(bot *Bot, event *onebot.GroupMessageEvent) {

}
