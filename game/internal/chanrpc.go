package internal

import (
	"dezhoupoker/msg"
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

	p := &Player{}
	p.Init()
	p.ConnAgent = a
	a.SetUserData(p)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	p, ok := a.UserData().(*Player)
	if ok && p.ConnAgent == a {
		log.Debug("<-------------%v 主动断开链接--------------->", p.Id)

		p.IsOnline = false
		if p.totalDownBet > 0 {
			rid := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rid)
			if v != nil {
				room := v.(*Room)
				var exist bool
				for _, v := range room.UserLeave {
					if v == p.Id {
						exist = true
					}
				}
				if exist == false {
					room.UserLeave = append(room.UserLeave, p.Id)
				}
			}
		} else {
			hall.UserRecord.Delete(p.Id)
			p.PlayerExitRoom()
		}
		c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
		leaveHall := &msg.Logout_S2C{}
		a.WriteMsg(leaveHall)
		a.Close()
	}
}
