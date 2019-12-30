package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	log.Debug("<-------------新链接请求连接--------------->")

	p := NewPlayer()
	p.ConnAgent = a
	a.SetUserData(p)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	p, ok := a.UserData().(*Player)
	if ok && p.ConnAgent == a {
		log.Debug("<-------------%v主动断开链接--------------->", p.Id)
		rid := hall.UserRoom[p.Id]
		v, _ := hall.RoomRecord.Load(rid)
		if v != nil {
			if p.gameStep == emInGaming {
				p.IsOnline = false
			} else {
				p.PlayerExitRoom()
			}
		}
		a.Close()
		// c4c.Logout()
	}
}
