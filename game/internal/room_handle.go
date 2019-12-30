package internal

import (
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"github.com/name5566/leaf/log"
	"time"
)

//PlayerJoinRoom 玩家加入房间
func (r *Room) PlayerJoinRoom(p *Player) {

	log.Debug("Player Join Game Room ~")

	hall.UserRoom[p.Id] = r.roomId

	// 查找用户是否存在，如果存在就插入数据库
	p.FindPlayerID()

	// 玩家带入筹码
	p.chips = r.TakeInRoomChips(p)

	p.chair = r.FindAbleChair(p.historyChair)
	r.PlayerList[p.chair] = p
	p.historyChair = p.chair
	// 房间总人数
	r.AllPlayer = append(r.AllPlayer, p)

	if r.Status == msg.GameStep_Waiting {
		// 返回房间数据
		enter := &msg.EnterRoom_S2C{}
		data := r.RespRoomData()
		enter.RoomData = data
		p.SendMsg(enter)

		r.StartGameRun()
	} else {
		// 如果玩家中途加入游戏，则玩家视为弃牌状态
		p.actStatus = msg.ActionStatus_FOLD
		// 返回房间数据
		enter := &msg.EnterRoom_S2C{}
		data := r.RespRoomData()
		enter.RoomData = data
		p.SendMsg(enter)
	}
}

//StartGameRun 游戏开始运行
func (r *Room) StartGameRun() {

	// 当前房间人数存在两人及两人以上才开始游戏
	n := r.PlayerLength()
	if n < 2 {
		log.Debug("房间人数少于2人，不能开始游戏~")
		return
	}

	// 设置玩家状态
	r.SetPlayerStatus()

	//1、洗牌
	r.Cards.Shuffle()

	//2、产生庄家
	var dealer *Player
	button := r.Button - 1
	r.Each(int(button+1)%MaxPlayer, func(p *Player) bool {
		r.Button = p.chair
		dealer = p
		dealer.IsButton = true
		log.Debug("玩家信息:%v", dealer)
		return false
	})

	log.Debug("庄家的座位号为 :%v", dealer.chair)

	//3、产生小盲注
	sb := r.Blind(dealer.chair)
	sb.blindType = msg.BlindType_Small_Blind
	log.Debug("小盲注座位号为 :%v", sb.chair)

	//4、小盲注下注
	r.betting(sb, r.SB)

	//5、产生大盲注
	bb := r.Blind(sb.chair)
	bb.blindType = msg.BlindType_Big_Blind
	log.Debug("大盲注座位号为 :%v", bb.chair)

	//6、大盲注下注
	r.betting(bb, r.BB)

	// 定义公共牌
	var pubCards algorithm.Cards

	//Round 1：preFlop 开始发手牌,下注
	r.readyPlay()
	r.Status = msg.GameStep_PreFlop

	r.Each(0, func(p *Player) bool {
		//2、生成玩家手牌,获取的是对应牌型生成二进制的数
		p.cards = algorithm.Cards{r.Cards.Take(), r.Cards.Take()}
		p.cardData.HandCardKeys = p.cards.HexInt()

		kind, _ := algorithm.De(p.cards.GetType())
		p.cardData.SuitPattern = msg.CardSuit(kind)
		log.Debug("preFlop玩家手牌和类型 ~ :%v, %v", p.cards.HexInt(), kind)

		game := &msg.GameStepChange_S2C{}
		game.RoomData = r.RespRoomData()
		p.SendMsg(game)
		return true
	})
	//3、行动、下注
	r.action(0)

	// 如果玩家全部摊牌直接比牌
	if r.remain <= 1 {
		// 直接摊牌
		goto showdown
	}

	//4、设置桌面筹码池
	r.calc()

	//Round 2：Flop 翻牌圈,牌桌上发3张公牌
	//1、准备阶段
	r.readyPlay()
	r.Status = msg.GameStep_Flop

	//2、生成桌面工牌赋值
	pubCards = algorithm.Cards{r.Cards.Take(), r.Cards.Take(), r.Cards.Take()}
	log.Debug("Flop桌面工牌数字 ~ :%v", pubCards.HexInt())

	r.publicCards = pubCards.HexInt()
	r.Each(0, func(p *Player) bool {
		cs := pubCards.Append(p.cards...)
		kind, _ := algorithm.De(cs.GetType())
		p.cardData.SuitPattern = msg.CardSuit(kind)

		// 游戏阶段变更
		game := &msg.GameStepChange_S2C{}
		game.RoomData = r.RespRoomData()
		p.SendMsg(game)
		return true
	})

	//3、行动、下注
	r.action(0)

	// 如果玩家全部摊牌直接比牌
	if r.remain <= 1 {
		// 直接摊牌
		goto showdown
	}

	//4、设置桌面筹码池
	r.calc()

	//Round 3：Turn 转牌圈,牌桌上发第4张公共牌
	//1、准备阶段
	r.readyPlay()
	r.Status = msg.GameStep_Turn

	//2、生成桌面第四张公牌
	pubCards = pubCards.Append(r.Cards.Take())
	log.Debug("Turn桌面工牌数字 ~ :%v", pubCards.HexInt())

	r.publicCards = pubCards.HexInt()
	r.Each(0, func(p *Player) bool {
		cs := pubCards.Append(p.cards...)
		kind, _ := algorithm.De(cs.GetType())
		p.cardData.SuitPattern = msg.CardSuit(kind)

		// 游戏阶段变更
		game := &msg.GameStepChange_S2C{}
		game.RoomData = r.RespRoomData()
		p.SendMsg(game)
		return true
	})

	//3、行动、下注
	r.action(0)

	// 如果玩家全部摊牌直接比牌
	if r.remain <= 1 {
		// 直接摊牌
		goto showdown
	}

	//4、设置桌面筹码池
	r.calc()

	//Round 4：River 河牌圈,牌桌上发第5张公共牌
	//1、准备阶段
	r.readyPlay()
	r.Status = msg.GameStep_River

	//2、生成桌面第五张公牌
	pubCards = pubCards.Append(r.Cards.Take())
	log.Debug("River桌面工牌数字 ~ :%v", pubCards.HexInt())

	r.publicCards = pubCards.HexInt()
	r.Each(0, func(p *Player) bool {
		cs := pubCards.Append(p.cards...)
		p.HandValue = cs.GetType()

		kind, _ := algorithm.De(cs.GetType())
		log.Debug("玩家手牌最后牌型：%v , %v", p.Id, kind)

		// 游戏阶段变更
		game := &msg.GameStepChange_S2C{}
		game.RoomData = r.RespRoomData()
		p.SendMsg(game)
		return true
	})

	//3、行动、下注
	r.action(0)

	// showdown 摊开底牌,开牌比大小
showdown:
	log.Debug("开始摊牌，开牌比大小 ~")
	r.showdown()

	// 打印数据
	r.PlantData()

	r.Status = msg.GameStep_ShowDown

	res := &msg.ResultGameData_S2C{}
	res.RoomData = r.RespRoomData()
	r.Broadcast(res)

	err := r.InsertRoomData()
	if err != nil {
		log.Debug("插入房间数据失败: %v", err)
	}

	// 剔除房间玩家
	r.KickPlayer()

	// 清除房间数据
	r.ClearRoomData()

	// 延时5秒，重新开始游戏
	time.AfterFunc(time.Second*5, func() {
		r.StartGameRun()
	})
}

//ExitFromRoom 退出房间处理
func (r *Room) ExitFromRoom(p *Player) {

	if p.chair != -1 {
		r.PlayerList[p.chair] = nil
	}

	for k, v := range r.AllPlayer {
		if v != nil {
			r.AllPlayer = append(r.AllPlayer[:k], r.AllPlayer[k+1:]...)
		}
	}

	delete(hall.UserRoom, p.Id)
	hall.UserRecord.Delete(p.Id)

	// 如果房间总人数为0，删除房间缓存
	if len(r.AllPlayer) == 0 {
		hall.RoomRecord.Delete(r.roomId)
		log.Debug("Room Player Number is 0，so Delete this Room~")
	}

	p.Account += p.chips
	// 清除用户数据
	p.ClearPlayerData()

	leave := &msg.LeaveRoom_S2C{}
	leave.PlayerInfo = new(msg.PlayerInfo)
	leave.PlayerInfo.Id = p.Id
	leave.PlayerInfo.NickName = p.NickName
	leave.PlayerInfo.HeadImg = p.HeadImg
	leave.PlayerInfo.Account = p.Account
	p.SendMsg(leave)
}

func (r *Room) PlantData() {

	for _, v := range r.PlayerList {
		if v != nil {
			log.Debug("玩家的ID: %v, 金额为: %v, 筹码为: %v", v.Id, v.Account, v.chips)
		}
	}
}
