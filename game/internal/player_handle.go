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
		room := v.(*Room)
		if p.IsInGame == true{
			var exist bool
			for _, v := range room.UserLeave {
				if v == p.Id {
					exist = true
				}
			}
			if exist == false {
				log.Debug("添加离线玩家UserLeave:%v",p.Id)
				room.UserLeave = append(room.UserLeave, p.Id)
			}

			leave := &msg.LeaveRoom_S2C{}
			leave.PlayerData = p.RespPlayerData()
			p.SendMsg(leave)

		} else {
			room.ExitFromRoom(p)
		}
	} else {
		log.Debug("Player Exit Room, But Not Found Player Room~")
	}
}

//ClearPlayerData 清除玩家数据
func (p *Player) ClearPlayerData() {
	p.chips = 0
	p.roomChips = 0
	p.chair = 0
	p.standUPNum = 0
	p.actStatus = msg.ActionStatus_WAITING
	p.gameStep = emNotGaming
	p.downBets = 0
	p.lunDownBets = 0
	p.totalDownBet = 0
	p.cardData = msg.CardSuitData{}
	p.resultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0
	p.blindType = msg.BlindType_No_Blind
	p.IsAllIn = false
	p.IsButton = false
	p.IsWinner = false
	p.IsTimeOutFold = false
	p.IsInGame = false
	p.timerCount = 0
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
			//ErrorResp(p.ConnAgent, msg.ErrorMsg_ChairAlreadyFull, "桌面位置已满")
			return
		}
		// 玩家坐下筹码重置为房间最少带入金额
		data := SetRoomConfig(r.cfgId)
		p.roomChips += p.chips
		p.chips = data.MinTakeIn
		p.roomChips -= p.chips

		p.chair = r.FindAbleChair()
		r.PlayerList[p.chair] = p
		p.standUPNum = 0
		p.IsStandUp = false
		p.IsTimeOutFold = false

		sitDown := &msg.SitDown_S2C{}
		sitDown.PlayerData = p.RespPlayerData()
		sitDown.RoomData = r.RespRoomData()
		r.Broadcast(sitDown)

		if r.RoomStat != RoomStatusRun {
			log.Debug("SitDownTable 开始运行游戏~")
			r.StartGameRun()
		}
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

		//if r.Status != msg.GameStep_Waiting {
		//	//ErrorResp(p.ConnAgent, msg.ErrorMsg_UserInGameNotStandUp, "玩家正在游戏中,不能站起")
		//	return
		//}

		if p.chair == -1 { // 防止客戶端重复点击多次
			return
		}

		if r.activeId == p.Id {
			p.actStatus = msg.ActionStatus_FOLD
		}

		r.PlayerList[p.chair] = nil

		//站起改变状态，座位为 -1，视为观战
		p.gameStep = emNotGaming
		p.IsStandUp = true

		standUp := &msg.StandUp_S2C{}
		standUp.PlayerData = p.RespPlayerData()
		r.Broadcast(standUp)

		// 这里发送数据给前端不能发送已经改变的位置
		p.chair = -1
	}
}
