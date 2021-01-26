# go-pbbot

使用方法

```go
// 创建http服务器，以下代码放在合适位置
// pbbot.UpgradeWebsocket(w http.ResponseWriter, r *http.Request)

pbbot.HandleGroupMessage = func (bot *pbbot.Bot, event *onebot.GroupMessageEvent){
    msg := pbbot.NewMsg().Text("hello").At(event.UserId)
    bot.SendGroupMesssage(event.GroupId, msg,false)
}
```