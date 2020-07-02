package internal

import (
	"dezhoupoker/msg"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	"time"
)

func init() {
	handlerReg(&msg.Ping{}, handlePing)

	handlerReg(&msg.Login_C2S{}, handleLogin)
	handlerReg(&msg.Logout_C2S{}, handleLogout)

	handlerReg(&msg.QuickStart_C2S{}, handleQuickStart)
	handlerReg(&msg.ChangeTable_C2S{}, handleChangeTable)
	handlerReg(&msg.LeaveRoom_C2S{}, handleLeaveRoom)

	handlerReg(&msg.SitDown_C2S{}, handleSitDown)
	handlerReg(&msg.StandUp_C2S{}, handleStandUp)
	handlerReg(&msg.PlayerAction_C2S{}, handleAction)
	handlerReg(&msg.AddChips_C2S{}, handleAddChips)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	a := args[1].(gate.Agent)

	pingTime := time.Now().UnixNano() / 1e6
	pong := &msg.Pong{
		ServerTime: pingTime,
	}
	a.WriteMsg(pong)
}

func handleLogin(args []interface{}) {
	m := args[0].(*msg.Login_C2S)
	a := args[1].(gate.Agent)

	//pl := a.UserData().(*Player)

	log.Debug("handleLogin 用户登入游戏~ :%v", m.Id)
	v, ok := hall.UserRecord.Load(m.Id)
	if ok { // 说明用户已存在
		p := v.(*Player)
		if p.ConnAgent == a { // 用户和链接都相同
			log.Debug("同一用户相同连接重复登录~")
			//ErrorResp(a, msg.ErrorMsg_UserRepeatLogin, "重复登录")
			return
		} else { // 用户相同，链接不相同
			err := hall.ReplacePlayerAgent(p.Id, a)
			if err != nil {
				log.Error("用户链接替换错误", err)
			}

			rId := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rId)
			if v != nil {
				// 玩家如果已在游戏中，则返回房间数据
				room := v.(*Room)
				for i, userId := range room.UserLeave {
					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
					// 把玩家从掉线列表中移除
					if userId == p.Id {
						room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
						break
					}
					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
				}
			}

			user, _ := hall.UserRecord.Load(p.Id)
			if user != nil {
				u := user.(*Player)
				login := &msg.Login_S2C{}
				login.PlayerInfo = new(msg.PlayerInfo)
				login.PlayerInfo.Id = u.Id
				login.PlayerInfo.NickName = u.NickName
				login.PlayerInfo.HeadImg = u.HeadImg
				login.PlayerInfo.Account = u.Account
				a.WriteMsg(login)

				//p.ConnAgent.Destroy()
				p.ConnAgent = a
				p.ConnAgent.SetUserData(user) //p
				p.IsOnline = true
			}
		}
	} else if !hall.agentExist(a) { // 玩家首次登入
		c4c.UserLoginCenter(m.GetId(), m.GetPassWord(), m.GetToken(), func(u *Player) {
			log.Debug("玩家首次登陆:%v", u.Id)
			login := &msg.Login_S2C{}
			login.PlayerInfo = new(msg.PlayerInfo)
			login.PlayerInfo.Id = u.Id
			login.PlayerInfo.NickName = u.NickName
			login.PlayerInfo.HeadImg = u.HeadImg
			login.PlayerInfo.Account = u.Account
			a.WriteMsg(login)

			u.Init()
			// 重新绑定信息
			a.SetUserData(u)
			u.ConnAgent = a

			u.Password = m.GetPassWord()
			u.Token = m.GetToken()

			hall.UserRecord.Store(u.Id, u)
		})
	}
}

func handleLogout(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLeaveHall 玩家退出大厅~ : %v", p.Id)

	if ok {
		if p.gameStep == emInGaming {
			var exist bool
			rid := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rid)
			if v != nil {
				room := v.(*Room)
				for _, v := range room.UserLeave {
					if v == p.Id {
						exist = true
					}
				}
				if exist == false {
					room.UserLeave = append(room.UserLeave, p.Id)
				}
				p.IsOnline = false
				leaveHall := &msg.Logout_S2C{}
				a.WriteMsg(leaveHall)
			}
		} else {
			c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
			p.IsOnline = false
			hall.UserRecord.Delete(p.Id)
			leaveHall := &msg.Logout_S2C{}
			a.WriteMsg(leaveHall)
			p.ConnAgent.Close()
		}
	}
}

func handleQuickStart(args []interface{}) {
	m := args[0].(*msg.QuickStart_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleQuickStart 快速匹配房间~ :%v", p.Id)

	if ok {
		hall.PlayerQuickStart(m.CfgId, p)
	}
}

func handleChangeTable(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleChangeTable 玩家更换房间~ :%v", p.Id)

	if ok {
		// 判断玩家当前状态是否正在游戏
		if p.gameStep == emInGaming {
			//ErrorResp(a, msg.ErrorMsg_UserNotChangeTable, "玩家正在游戏,不能换桌")
			return
		}
		rId := hall.UserRoom[p.Id]
		v, _ := hall.RoomRecord.Load(rId)
		if v != nil {
			r := v.(*Room)
			hall.PlayerChangeTable(r, p)
		}
	}
}

func handleLeaveRoom(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLeaveRoom 玩家离开房间~ :%v", p.Id)

	if ok {
		p.PlayerExitRoom()
	}
}

func handleSitDown(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleSitDown 玩家坐下座位~ :%v", p.Id)

	if ok {
		p.SitDownTable()
	}
}

func handleStandUp(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleStandUp 玩家站起观战~ :%v", p.Id)

	if ok {
		p.StandUpTable()
	}
}

func handleAction(args []interface{}) {
	m := args[0].(*msg.PlayerAction_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleAction 玩家开始行动~ :%v", p.Id)

	if ok {
		p.action <- m.Action
		p.downBets = m.BetAmount
		p.lunDownBets += m.BetAmount
		p.totalDownBet += m.BetAmount
	}
}

func handleAddChips(args []interface{}) {
	m := args[0].(*msg.AddChips_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleAction 玩家添加筹码~ :%v", p.Id)

	if ok {
		p.chips += m.AddChips
		p.roomChips -= m.AddChips

		data := &msg.AddChips_S2C{}
		data.Chair = p.chair
		data.AddChips = m.AddChips
		data.Chips = p.chips
		data.RoomChips = p.roomChips
		data.SysBuyChips = m.SysBuyChips

		p.SendMsg(data)
	}
}
