package internal

import (
	"dezhoupoker/msg"
	"github.com/name5566/leaf/log"
)

//PlayerExitRoom 玩家退出房间
func (p *Player) PlayerExitRoom() {
	rId := hall.UserRoom[p.Id]
	v, _ := hall.RoomRecord.Load(rId)
	if v != nil {
		r := v.(*Room)
		if p.gameStep == emInGaming {
			leave := &msg.LeaveRoom_S2C{}
			leave.PlayerInfo = new(msg.PlayerInfo)
			leave.PlayerInfo.Id = p.Id
			leave.PlayerInfo.NickName = p.NickName
			leave.PlayerInfo.HeadImg = p.HeadImg
			leave.PlayerInfo.Account = p.Account
			p.SendMsg(leave)

		} else {
			r.ExitFromRoom(p)
		}
	} else {
		log.Debug("Player Exit Room, But Not Found Player Room~")
	}
}

//ClearPlayerData 清除玩家数据
func (p *Player) ClearPlayerData() {
	p.chips = 0
	p.chair = 0
	p.historyChair = 0
	p.standUPNum = 0
	p.actStatus = msg.ActionStatus_WAITING
	p.gameStep = emNotGaming
	p.downBets = 0
	p.totalDownBet = 0
	p.cardData = msg.CardSuitData{}
	p.resultMoney = 0
	p.blindType = msg.BlindType_No_Blind
	p.IsAllIn = false
	p.IsButton = false
	p.IsWinner = false
	p.HandValue = 0
}

//SitDownTable 玩家坐下座位
func (p *Player) SitDownTable() {

	rId := hall.UserRoom[p.Id]
	v, _ := hall.RoomRecord.Load(rId)
	if v != nil {
		// 玩家如果已在游戏中，则返回房间数据
		r := v.(*Room)
		if r.PlayerLength() >= MaxPlayer {
			ErrorResp(p.ConnAgent, msg.ErrorMsg_ChairAlreadyFull, "桌面位置已满")
			return
		}
		p.chair = r.FindAbleChair(p.historyChair)
		r.PlayerList[p.chair] = p

		sitDown := &msg.SitDown_S2C{}
		sitDown.RoomData = r.RespRoomData()
		p.SendMsg(sitDown)
	}
}

//StandUpTable 玩家站起观战
func (p *Player) StandUpTable() {
	// 判断玩家是否是当前行动玩家，如果是则直接弃牌站起
	rId := hall.UserRoom[p.Id]
	v, _ := hall.RoomRecord.Load(rId)
	if v != nil {
		// 玩家如果已在游戏中，则返回房间数据
		r := v.(*Room)
		if r.activeSeat == p.chair {
			p.actStatus = msg.ActionStatus_FOLD
		}
		r.PlayerList[p.chair] = nil

		//站起改变状态，座位为 -1，视为观战
		p.gameStep = emNotGaming
		p.chair = -1

		standUp := &msg.StandUp_S2C{}
		standUp.RoomData = r.RespRoomData()
		p.SendMsg(standUp)
	}
}

//SetPlayerAction 设置玩家行动
func (p *Player) SetPlayerAction(m *msg.PlayerAction_C2S) {

}
