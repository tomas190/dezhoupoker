package internal

import (
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"github.com/name5566/leaf/log"
	"time"
)

//PlayerJoinRoom 玩家加入房间
func (r *Room) PlayerJoinRoom(p *Player) {
	// 查找用户是否存在，如果存在就插入数据库
	if p.IsRobot == false {
		p.FindPlayerInfo()
	}

	//log.Debug("Player Join Game Room ~")

	hall.UserRoom[p.Id] = r.roomId

	p.PreRoomId = r.roomId
	// 玩家带入筹码
	r.TakeInRoomChips(p)

	p.chair = r.FindAbleChair()
	r.PlayerList[p.chair] = p

	// 房间总人数
	r.AllPlayer = append(r.AllPlayer, p)

	//log.Debug("玩家加入房间: %v,房间状态: %v", p.Id, r.Status)
	if r.RoomStat != RoomStatusRun {
		// 返回房间数据
		roomData := r.RespRoomData()

		enter := &msg.JoinRoom_S2C{}
		enter.RoomData = roomData
		p.SendMsg(enter)

		if r.AllPlayerLength() > 1 { // 广播其他玩家进入游戏
			notice := &msg.NoticeJoin_S2C{}
			notice.PlayerData = roomData.PlayerData[p.chair]
			r.BroadCastExcept(notice, p)
		}

		log.Debug("PlayerJoinRoom 开始运行游戏~")
		r.StartGameRun()
	} else {
		// 如果玩家中途加入游戏，则玩家视为弃牌状态
		p.actStatus = msg.ActionStatus_WAITING
		p.gameStep = emNotGaming
		// 返回房间数据
		roomData := r.RespRoomData()

		enter := &msg.JoinRoom_S2C{}
		enter.RoomData = roomData
		p.SendMsg(enter)

		if r.AllPlayerLength() > 1 { // 广播其他玩家进入游戏
			notice := &msg.NoticeJoin_S2C{}
			notice.PlayerData = roomData.PlayerData[p.chair]
			r.BroadCastExcept(notice, p)
		}
	}
}

//StartGameRun 游戏开始运行
func (r *Room) StartGameRun() {

	// 当前房间人数存在两人及两人以上才开始游戏
	if r.PlayerLength() < 2 {
		r.RoomStat = RoomStatusNone

		log.Debug("房间人数少于2人，不能开始游戏~")
		return
	}
	log.Debug("%v房间 游戏开始，玩家开始行动~", r.cfgId)

	// 准备阶段定时任务
	r.ReadyTimer()

	// 游戏开始定时器任务
	r.GameRunTask()
}

func (r *Room) GameRunning() {

	// 定义公共牌
	var pubCards algorithm.Cards

	//1、产生小盲注
	sb := r.Blind(r.Banker) //dealer.chair
	r.SBId = sb.Id
	sb.blindType = msg.BlindType_Small_Blind
	log.Debug("小盲注座位号为 :%v", sb.chair)

	//2、产生大盲注
	bb := r.Blind(sb.chair)
	r.BBId = bb.Id
	bb.blindType = msg.BlindType_Big_Blind
	log.Debug("大盲注座位号为 :%v", bb.chair)

	//3、小盲注下注
	r.betting(sb, r.SB)
	//4、大盲注下注
	r.betting(bb, r.BB)

	//5、行动、下注 (这里应该大盲下一位开始下注)
	r.Action(int(bb.chair) + 1)

	// 如果玩家全部摊牌直接比牌
	if r.remain <= 1 {
		r.IsShowDown = 1
		// 直接摊牌
		goto showdown
	}

	time.Sleep(time.Millisecond * 1000)

	//Round 2：Flop 翻牌圈,牌桌上发3张公牌
	r.Status = msg.GameStep_Flop
	log.Debug("GameStep_Flop 阶段: %v", r.Status)
	//2、生成桌面工牌赋值
	pubCards = algorithm.Cards{r.Cards.Take(), r.Cards.Take(), r.Cards.Take()}
	//log.Debug("Flop桌面工牌数字 ~ :%v", pubCards.HexInt())

	r.publicCards = pubCards.HexInt()
	for i := 0; i < len(r.PlayerList); i++ {
		if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming {
			p := r.PlayerList[i]
			cs := pubCards.Append(p.cards...)
			kind, _ := algorithm.De(cs.GetType())
			p.cardData.SuitPattern = msg.CardSuit(kind)

			// 游戏阶段变更
			game := &msg.GameStepChange_S2C{}
			game.RoomData = r.RespRoomData()
			r.Broadcast(game)
		}
	}
	//1、准备阶段
	r.readyPlay()

	// 随机添加机器人
	//r.AddRobot()

	time.Sleep(time.Millisecond * 1000)

	//3、行动、下注
	r.Action(int(r.Banker + 1))

	// 如果玩家全部摊牌直接比牌
	if r.remain <= 1 {
		r.IsShowDown = 1
		// 直接摊牌
		goto showdown
	}

	time.Sleep(time.Millisecond * 1000)

	//Round 3：Turn 转牌圈,牌桌上发第4张公共牌
	r.Status = msg.GameStep_Turn
	log.Debug("GameStep_Turn 阶段: %v", r.Status)

	//2、生成桌面第四张公牌
	pubCards = pubCards.Append(r.Cards.Take())
	//log.Debug("Turn桌面工牌数字 ~ :%v", pubCards.HexInt())

	r.publicCards = pubCards.HexInt()
	for i := 0; i < len(r.PlayerList); i++ {
		if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming {
			p := r.PlayerList[i]
			cs := pubCards.Append(p.cards...)
			kind, _ := algorithm.De(cs.GetType())
			p.cardData.SuitPattern = msg.CardSuit(kind)

			// 游戏阶段变更
			game := &msg.GameStepChange_S2C{}
			game.RoomData = r.RespRoomData()
			r.Broadcast(game)
		}
	}
	//1、准备阶段
	r.readyPlay()

	time.Sleep(time.Millisecond * 1000)

	//3、行动、下注
	r.Action(int(r.Banker + 1))

	// 如果玩家全部摊牌直接比牌
	if r.remain <= 1 {
		r.IsShowDown = 1
		// 直接摊牌
		goto showdown
	}

	time.Sleep(time.Millisecond * 1000)

	//Round 4：River 河牌圈,牌桌上发第5张公共牌
	r.Status = msg.GameStep_River
	log.Debug("GameStep_River 阶段: %v", r.Status)

	//2、生成桌面第五张公牌
	pubCards = pubCards.Append(r.Cards.Take())
	//log.Debug("River桌面工牌数字 ~ :%v", pubCards.HexInt())

	r.publicCards = pubCards.HexInt()
	for i := 0; i < len(r.PlayerList); i++ {
		if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming {
			p := r.PlayerList[i]
			cs := pubCards.Append(p.cards...)
			p.HandValue = cs.GetType()

			kind, _ := algorithm.De(cs.GetType())
			p.cardData.SuitPattern = msg.CardSuit(kind)

			cardSlice := cs.GetCardHexInt()
			if kind == 5 {
				p.cardData.PublicCardKeys = cardSlice[:len(cardSlice)-2]
			} else if kind == 6 {
				cardSlice = algorithm.ShowCards(kind, cardSlice)
				p.cardData.PublicCardKeys = cardSlice[:len(cardSlice)-2]
			} else {
				p.cardData.PublicCardKeys = cardSlice[2:]
			}
			//algorithm.ShowCards(kind, cardSlice)
			//log.Debug("玩家手牌最后牌型: %v , 类型: %v, 牌值: %v ", p.Id, kind, p.cardData.PublicCardKeys)

			// 游戏阶段变更
			game := &msg.GameStepChange_S2C{}
			game.RoomData = r.RespRoomData()
			r.Broadcast(game)
		}
	}

	time.Sleep(time.Millisecond * 1000)
	//3、行动、下注
	r.Action(int(r.Banker + 1))

showdown:
	log.Debug("开始摊牌，开牌比大小 ~")
	r.ShowDown()

	r.ResultMoney()

	time.Sleep(time.Millisecond * 1000)

	//Round 5: ShowDown 摊开底牌,开牌比大小
	r.Status = msg.GameStep_ShowDown
	log.Debug("GameStep_ShowDown 阶段: %v", r.Status)

	result := &msg.ResultGameData_S2C{}
	result.RoomData = r.RespRoomData()
	r.Broadcast(result)

	// 打印数据
	//r.PlantData()

	//err := r.InsertRoomData()
	//if err != nil {
	//	log.Debug("插入房间数据失败: %v", err)
	//}

	// 清除房间数据
	r.ClearRoomData()

	// 广播游戏结算时间
	settle := &msg.SettleTime_S2C{}
	settle.SettleTime = SettleTime
	r.Broadcast(settle)

	// 重新开始游戏
	r.RestartGame()

}

//ExitFromRoom 退出房间处理
func (r *Room) ExitFromRoom(p *Player) {

	if p.chair != -1 {
		r.PlayerList[p.chair] = nil
	}

	for k, v := range r.AllPlayer {
		if v != nil && v.Id == p.Id {
			r.AllPlayer = append(r.AllPlayer[:k], r.AllPlayer[k+1:]...)
		}
	}

	p.Account += p.chips
	p.Account += p.roomChips

	leave := &msg.LeaveRoom_S2C{}
	leave.PlayerData = p.RespPlayerData()
	p.SendMsg(leave)
	r.BroadCastExcept(leave, p)
	log.Debug("玩家退出房间成功！:%v", p)

	// 如果房间总人数为0，删除房间缓存
	if len(r.AllPlayer) == 0 {
		for k, v := range hall.roomList {
			if v.roomId == r.roomId {
				hall.roomList = append(hall.roomList[:k], hall.roomList[k+1:]...)
				hall.RoomRecord.Delete(r.roomId)
				log.Debug("Room Player Number is 0，so Delete this Room~")
			}
		}
	}

	delete(hall.UserRoom, p.Id)

	// 清除用户数据
	p.ClearPlayerData()
}

func (r *Room) PlantData() {
	for _, v := range r.AllPlayer {
		if v != nil && v.IsRobot == false {
			log.Debug("玩家的ID: %v,玩家name:%v, 金额为: %v, 筹码为: %v", v.Id, v.NickName, v.Account, v.chips)
		}
	}
}
