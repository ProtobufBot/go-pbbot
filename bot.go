package pbbot

import (
	"encoding/json"
	"errors"

	"github.com/ProtobufBot/go-pbbot/proto_gen/onebot"
	"github.com/ProtobufBot/go-pbbot/util"
	"github.com/fanliao/go-promise"
	"github.com/golang/groupcache/lru"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var Bots = make(map[int64]*Bot)

type Bot struct {
	BotId         int64
	Session       *SafeWebSocket
	WaitingFrames *lru.Cache
}

func NewBot(botId int64, conn *websocket.Conn) *Bot {
	messageHandler := func(messageType int, data []byte) {
		var frame onebot.Frame
		if messageType == websocket.BinaryMessage {
			err := proto.Unmarshal(data, &frame)
			if err != nil {
				log.Errorf("failed to unmarshal websocket binary message, err: %+v", err)
				return
			}
		} else if messageType == websocket.TextMessage {
			err := json.Unmarshal(data, &frame)
			if err != nil {
				log.Errorf("failed to unmarshal websocket text message, err: %+v", err)
				return
			}
		} else {
			log.Errorf("invalid websocket messageType: %+v", messageType)
			return
		}

		bot, ok := Bots[botId]
		if !ok {
			_ = conn.Close()
			return
		}
		util.SafeGo(func() {
			bot.handleFrame(&frame)
		})
	}
	closeHandler := func(code int, message string) {
		HandleDisconnect(Bots[botId])
		delete(Bots, botId)
	}
	safeWs := NewSafeWebSocket(conn, messageHandler, closeHandler)
	bot := &Bot{
		BotId:         botId,
		Session:       safeWs,
		WaitingFrames: lru.New(128),
	}
	bot.WaitingFrames.OnEvicted = func(key lru.Key, value interface{}) {
		p, ok := value.(*promise.Promise)
		if !ok {
			log.Errorf("failed to clean expired waiting frame")
			return
		}
		_ = p.Reject(errors.New("cleaned by lru"))
	}
	Bots[botId] = bot
	HandleConnect(bot)
	return bot
}

func (bot *Bot) handleFrame(frame *onebot.Frame) {
	if event := frame.GetPrivateMessageEvent(); event != nil {
		HandlePrivateMessage(bot, event)
		return
	}
	if event := frame.GetGroupMessageEvent(); event != nil {
		HandleGroupMessage(bot, event)
		return
	}
	// TODO 剩余Event
	if frame.FrameType < 300 {
		log.Errorf("unknown frame type: %+v", frame.FrameType)
		return
	}
	v, ok := bot.WaitingFrames.Get(frame.Echo)
	if !ok {
		log.Errorf("failed to find waiting frame")
		return
	}
	p, ok := v.(*promise.Promise)
	if !ok {
		log.Errorf("failed to convert waiting frame promise")
		return
	}
	if err := p.Resolve(frame); err != nil {
		log.Errorf("failed to resolve waiting frame promise")
		return
	}
}

func (bot *Bot) sendFrameAndWait(frame *onebot.Frame) (*onebot.Frame, error) {
	frame.BotId = bot.BotId
	frame.Echo = util.GenerateIdStr()
	frame.Ok = true
	data, err := proto.Marshal(frame)
	if err != nil {
		return nil, err
	}
	bot.Session.Send(websocket.BinaryMessage, data)
	p := promise.NewPromise()
	bot.WaitingFrames.Add(frame.Echo, p)
	resp, err := p.Get()
	if err != nil {
		return nil, err
	}
	respFrame, ok := resp.(*onebot.Frame)
	if !ok {
		return nil, errors.New("failed to convert promise result to resp frame")
	}
	return respFrame, nil
}

func (bot *Bot) SendPrivateMessage(userId int64, msg *Msg, autoEscape bool) *onebot.SendPrivateMsgResp {
	resp, err := bot.sendFrameAndWait(&onebot.Frame{
		FrameType: onebot.Frame_TSendPrivateMsgReq,
		Data: &onebot.Frame_SendPrivateMsgReq{
			SendPrivateMsgReq: &onebot.SendPrivateMsgReq{
				UserId:     userId,
				Message:    msg.MessageList,
				AutoEscape: autoEscape,
			},
		},
	})
	if err != nil {
		return nil
	}
	return resp.GetSendPrivateMsgResp()
}

func (bot *Bot) SendGroupMessage(groupId int64, msg *Msg, autoEscape bool) *onebot.SendGroupMsgResp {
	resp, err := bot.sendFrameAndWait(&onebot.Frame{
		FrameType: onebot.Frame_TSendGroupMsgReq,
		Data: &onebot.Frame_SendGroupMsgReq{
			SendGroupMsgReq: &onebot.SendGroupMsgReq{
				GroupId:    groupId,
				Message:    msg.MessageList,
				AutoEscape: autoEscape,
			},
		},
	})
	if err != nil {
		return nil
	}
	return resp.GetSendGroupMsgResp()
}

// TODO 剩余API
