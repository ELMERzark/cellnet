package rpc

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/gamedef"
	"github.com/davyxu/cellnet/socket"
)

type Response interface {
	Feedback(interface{})
}

type response struct {
	ses cellnet.Session
	req *gamedef.RemoteCallREQ
}

func (self *response) Feedback(msg interface{}) {

	pkt, _ := cellnet.BuildPacket(msg)

	self.ses.Send(&gamedef.RemoteCallACK{
		MsgID:  pkt.MsgID,
		Data:   pkt.Data,
		CallID: self.req.CallID,
	})
}

func (self *response) ContextID() uint32 {
	return self.req.MsgID
}

func InstallServer(p cellnet.Peer) {

	// 服务端
	socket.RegisterSessionMessage(p, "gamedef.RemoteCallREQ", func(content interface{}, ses cellnet.Session) {
		msg := content.(*gamedef.RemoteCallREQ)

		p.CallData(&response{
			ses: ses,
			req: msg,
		})

	})

}

// 注册连接消息
func RegisterMessage(eq cellnet.EventQueue, msgName string, userHandler func(Response, interface{})) {

	msgMeta := cellnet.MessageMetaByName(msgName)

	eq.RegisterCallback(msgMeta.ID, func(data interface{}) {

		if ev, ok := data.(*response); ok {

			rawMsg, err := cellnet.ParsePacket(&cellnet.Packet{
				MsgID: ev.req.MsgID,
				Data:  ev.req.Data,
			}, msgMeta.Type)

			if err != nil {
				log.Errorln("unmarshaling error:", err)
				return
			}

			userHandler(ev, rawMsg)

		}

	})

}
