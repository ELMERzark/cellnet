package socket

import (
	"net"

	"github.com/davyxu/cellnet"
)

type socketAcceptor struct {
	*peerProfile
	*sessionMgr

	listener net.Listener

	running bool
}

func (self *socketAcceptor) Start(address string) cellnet.Peer {

	ln, err := net.Listen("tcp", address)

	self.listener = ln

	if err != nil {

		log.Errorf("#listen failed(%s) %v", self.name, err.Error())
		return self
	}

	self.running = true

	log.Debugf("#listen(%s) %s ", self.name, address)

	// 接受线程
	go func() {
		for self.running {
			conn, err := ln.Accept()

			if err != nil {
				log.Errorln(err)
				break
			}

			// 处理连接进入独立线程, 防止accept无法响应
			go func() {

				ses := newSession(NewPacketStream(conn), self.EventQueue, self)

				// 添加到管理器
				self.sessionMgr.Add(ses)

				// 断开后从管理器移除
				ses.OnClose = func() {
					self.sessionMgr.Remove(ses)
				}

				log.Debugf("#accepted(%s) sid: %d", self.name, ses.ID())

				// 通知逻辑
				self.PostData(NewSessionEvent(Event_SessionAccepted, ses, nil))
			}()

		}

	}()

	self.PostData(NewPeerEvent(Event_PeerStart, self))

	return self
}

func (self *socketAcceptor) Stop() {

	if !self.running {
		return
	}

	self.PostData(NewPeerEvent(Event_PeerStop, self))

	self.running = false

	self.listener.Close()
}

func NewAcceptor(pipe cellnet.EventPipe) cellnet.Peer {

	self := &socketAcceptor{
		sessionMgr:  newSessionManager(),
		peerProfile: newPeerProfile(pipe.AddQueue()),
	}

	self.PostData(NewPeerEvent(Event_PeerInit, self))

	return self
}
