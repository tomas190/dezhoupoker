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

	pl := a.UserData().(*Player)

	log.Debug("handleLogin 用户登入游戏~ :%v", m.Id)
	v, ok := hall.UserRecord.Load(m.Id)
	if ok { // 说明用户已存在
		p := v.(*Player)
		if p.ConnAgent == a { // 用户和链接都相同
			log.Debug("同一用户相同连接重复登录~")
			ErrorResp(a, msg.ErrorMsg_UserRepeatLogin, "重复登录")
			return
		} else { // 用户相同，链接不相同
			err := hall.ReplacePlayerAgent(p.Id, a)
			if err != nil {
				log.Error("用户链接替换错误", err)
			}

			login := &msg.Login_S2C{}
			login.PlayerInfo = new(msg.PlayerInfo)
			login.PlayerInfo.Id = p.Id
			login.PlayerInfo.NickName = p.NickName
			login.PlayerInfo.HeadImg = p.HeadImg
			login.PlayerInfo.Account = p.Account
			a.WriteMsg(login)

			rId := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rId)
			if v != nil {
				// 玩家如果已在游戏中，则返回房间数据
				r := v.(*Room)
				enter := &msg.EnterRoom_S2C{}
				enter.RoomData = r.RespRoomData()
				a.WriteMsg(enter)
			}
		}
	} else if !hall.agentExist(a) { // 玩家首次登入
		//p := v.(*Player)
		// 中心服登入
		//c4c.UserLogin()
		pl.Id = m.Id
		pl.Account = 4000

		// 重新绑定信息
		pl.ConnAgent = a
		a.SetUserData(pl)

		hall.UserRecord.Store(pl.Id, pl)

		login := &msg.Login_S2C{}
		login.PlayerInfo = new(msg.PlayerInfo)
		login.PlayerInfo.Id = pl.Id
		login.PlayerInfo.NickName = pl.NickName
		login.PlayerInfo.HeadImg = pl.HeadImg
		login.PlayerInfo.Account = pl.Account
		a.WriteMsg(login)
	}

}

func handleLogout(args []interface{}) {

}

func handleQuickStart(args []interface{}) {
	m := args[0].(*msg.QuickStart_C2S)
	a := args[1].(gate.Agent)

	p := a.UserData().(*Player)

	//p.Id = strconv.Itoa(int(time.Now().UnixNano()))
	//p.Account = 4000

	log.Debug("handleQuickStart 快速匹配房间~ :%v", p.Id)

	rId := hall.UserRoom[p.Id]
	v, _ := hall.RoomRecord.Load(rId)
	if v != nil {
		// 玩家如果已在游戏中，则返回房间数据
		r := v.(*Room)
		data := r.RespRoomData()

		enter := &msg.EnterRoom_S2C{}
		enter.RoomData = data
		a.WriteMsg(enter)
		return
	}

	hall.PlayerQuickStart(m.CfgId, p)
}

func handleChangeTable(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleChangeTable 玩家更换房间~ :%v", p.Id)

	if ok {
		// 判断玩家当前状态是否正在游戏
		if p.gameStep == emInGaming {
			ErrorResp(a, msg.ErrorMsg_UserNotChangeTable, "玩家正在游戏,不能换桌")
			return
		}
		rId := hall.UserRoom[p.Id]
		v, _ := hall.RoomRecord.Load(rId)
		if v != nil {
			r := v.(*Room)
			hall.PlayerQuickStart(r.cfgId, p)
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
		p.downBets = m.BetAmount
		p.action <- m.Action
	}
}
