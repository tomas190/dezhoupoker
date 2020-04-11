package internal

import (
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"time"
)

type Room struct {
	roomId     string
	cfgId      string    // 房间配置ID
	PlayerList []*Player // 座位玩家列表，最高9人
	AllPlayer  []*Player // 房间所有玩家，包括站起玩家座位号为-1

	activeSeat  int32        // 当前正在行动玩家座位号
	minRaise    float64      // 加注最小值
	potMoney    float64      // 桌面注池金额
	publicCards []int32      // 桌面公牌
	Status      msg.GameStep // 房间当前阶段  就直接判断是否在等待状态

	Cards    algorithm.Cards
	preChips float64   // 当前回合，上个玩家下注金额
	remain   int32     // 记录每个阶段玩家的下注的数量
	allin    int32     // allin玩家的数量
	Pot      []float64 // 奖池筹码数, 第一项为主池，其他项(若存在)为边池
	Chips    []float64 // 所有玩家本局下的总筹码,奖池筹码数,第一项为主池，其他项(若存在)为边池
	Banker   int32     // 庄家座位号
	SB       float64   // 小盲注
	BB       float64   // 大盲注

	counter int32
	clock   *time.Ticker
}

const (
	MaxPlayer = 9
)

const (
	ReadyTime  = 5  // 开始准备时间
	SettleTime = 5  // 游戏结算时间
	ActionTime = 15 // 玩家行动时间
)

var ReadyTimeChan chan bool

func (r *Room) Init(cfgId string) {
	roomId := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	r.roomId = roomId
	r.cfgId = cfgId
	r.AllPlayer = nil
	r.PlayerList = make([]*Player, MaxPlayer)
	for i := 0; i < len(r.PlayerList); i++ {
		r.PlayerList[i] = nil
	}

	rd := SetRoomConfig(cfgId)

	r.activeSeat = -1
	r.minRaise = rd.BB
	r.potMoney = 0
	r.publicCards = nil
	r.Status = msg.GameStep_Waiting
	r.preChips = 0
	r.remain = 0
	r.allin = 0
	r.Pot = make([]float64, 0, MaxPlayer)
	r.Chips = make([]float64, MaxPlayer)
	r.Banker = 0
	r.SB = rd.SB
	r.BB = rd.BB

	r.counter = 0
	r.clock = time.NewTicker(time.Second)

	ReadyTimeChan = make(chan bool)
}

//BroadCastExcept 向指定玩家之外的玩家广播
func (r *Room) BroadCastExcept(msg interface{}, except *Player) {
	for _, p := range r.AllPlayer {
		if p != nil && except.chair != p.chair {
			p.SendMsg(msg)
		}
	}
}

//Broadcast 广播消息
func (r *Room) Broadcast(msg interface{}) {
	for _, v := range r.AllPlayer {
		if v != nil {
			v.SendMsg(msg)
		}
	}
}

//IsCanJoin 房间是否还能加入
func (r *Room) IsCanJoin() bool {
	return r.PlayerLength() < MaxPlayer
}

//PlayerLength 房间玩家人数
func (r *Room) PlayerLength() int32 {
	var num int32
	for _, v := range r.PlayerList {
		if v != nil {
			num++
		}
	}
	log.Debug("房间号:%v  玩家人数:%v", r.roomId, num)
	return num
}

// 房间庄家座位号
func (r *Room) RoomBanker(banker int) *Player {
	i := banker + 1
	for ; i < len(r.PlayerList); i = (i + 1) % MaxPlayer {
		if r.PlayerList[i] != nil {
			log.Debug("玩家信息:%v", r.PlayerList[i])
			return r.PlayerList[i]
		}
	}
	return nil
}

//TakeInRoomChips 玩家带入筹码
func (r *Room) TakeInRoomChips(p *Player) {
	//1、如果玩家余额 大于房间最大设定金额 MaxTakeIn，则带入金额就设为 房间最大设定金额
	//2、如果玩家余额 小于房间最大设定金额 MaxTakeIn，则带入金额就设为 玩家的所有余额
	data := SetRoomConfig(r.cfgId)
	if p.Account > data.MaxTakeIn {
		p.Account = p.Account - data.MaxTakeIn
		p.chips = data.MinTakeIn
		p.roomChips = data.MaxTakeIn - p.chips
	} else {
		p.roomChips = p.Account
		p.Account = p.Account - p.Account
		p.chips = data.MinTakeIn
		p.roomChips -= p.chips
	}

}

//FindAbleChair 寻找可用座位
func (r *Room) FindAbleChair() int32 {
	// 先判断玩家历史座位是否已存在其他玩家，如果没有还是坐下历史座位

	for chair, p := range r.PlayerList {
		if p == nil {
			log.Debug("座位号下标为~ :%v", chair)
			return int32(chair)
		}
	}
	panic("ERROR: Don't find able chair, Should check canJoin first please")
}

//KickPlayer 剔除房间玩家
func (r *Room) KickPlayer() {
	// 遍历桌面玩家，踢掉玩家筹码和房间小于房间最小带入金额
	//data := SetRoomConfig(r.cfgId)
	for _, v := range r.PlayerList { // 玩家筹码为0怎么办
		if v != nil {
			if v.chips+v.roomChips < 3 {
				ErrorResp(v.ConnAgent, msg.ErrorMsg_ChipsInsufficient, "玩家筹码不足")
				v.PlayerExitRoom()
			}
		}
	}

	// 遍历站起玩家，是否在该房间站起超时
	for _, v := range r.AllPlayer {
		if v != nil && v.chair == -1 {
			v.standUPNum++
			if v.standUPNum >= 6 {
				ErrorResp(v.ConnAgent, msg.ErrorMsg_UserStandUpTimeOut, "玩家站起超时")
				v.PlayerExitRoom()
			}
		}
	}

	// 清理断线玩家
}

// 玩家补充筹码
func (r *Room) PlayerAddChips() {
	for _, v := range r.PlayerList {
		if v != nil && v.chips < 1 {
			if v.roomChips > 10 {
				v.roomChips -= 10
				v.chips += 10
				addChips := &msg.AddChips_S2C{}
				addChips.Chair = v.chair
				addChips.AddChips = 10
				addChips.Chips = v.chips
				addChips.RoomChips = v.roomChips
				addChips.SysBuyChips = 1
				v.SendMsg(addChips)
			} else {
				// 自动补充筹码
				money := v.roomChips
				v.roomChips = 0
				v.chips = v.chips + money
				addChips := &msg.AddChips_S2C{}
				addChips.Chair = v.chair
				addChips.AddChips = money
				addChips.Chips = v.chips
				addChips.RoomChips = v.roomChips
				addChips.SysBuyChips = 1
				v.SendMsg(addChips)
			}
		}
	}
}

// 超时玩家站起
func (r *Room) TimeOutStandUp() {
	for _, v := range r.PlayerList {
		if v != nil && v.IsTimeOutFold == true {
			v.StandUpTable()
		}
	}
}

//ClearRoomData 清除房间数据
func (r *Room) ClearRoomData() {
	r.activeSeat = -1
	r.potMoney = 0
	r.publicCards = nil
	r.Status = msg.GameStep_Waiting
	r.preChips = 0
	r.remain = 0
	r.allin = 0
	r.Chips = make([]float64, MaxPlayer)

	for _, v := range r.AllPlayer {
		if v != nil {
			v.actStatus = msg.ActionStatus_WAITING
			v.gameStep = emNotGaming
			v.downBets = 0
			v.totalDownBet = 0
			v.cardData = msg.CardSuitData{}
			v.resultMoney = 0
			v.blindType = msg.BlindType_No_Blind
			v.IsAllIn = false
			v.IsWinner = false
			v.IsButton = false
			v.HandValue = 0
		}
	}

}

//RespRoomData 返回房间数据
func (r *Room) RespRoomData() *msg.RoomData {
	rd := &msg.RoomData{}
	rd.RoomId = r.roomId
	rd.CfgId = r.cfgId
	rd.GameStep = r.Status
	rd.MinRaise = r.minRaise
	rd.PreChips = r.preChips
	rd.ActionSeat = r.activeSeat
	rd.BigBlind = r.BB
	rd.Banker = r.Banker
	rd.PotMoney = r.potMoney
	rd.PublicCards = r.publicCards
	// 这里只需要遍历桌面玩家，站起玩家不显示出来
	for _, v := range r.PlayerList {
		if v != nil {
			pd := &msg.PlayerData{}
			pd.PlayerInfo = new(msg.PlayerInfo)
			pd.PlayerInfo.Id = v.Id
			pd.PlayerInfo.NickName = v.NickName
			pd.PlayerInfo.HeadImg = v.HeadImg
			pd.PlayerInfo.Account = v.Account
			pd.Chair = v.chair
			pd.StandUPNum = v.standUPNum
			pd.Chips = v.chips
			pd.RoomChips = v.roomChips
			pd.ActionStatus = v.actStatus
			pd.GameStep = int32(v.gameStep)
			pd.DownBets = v.downBets
			pd.TotalDownBet = v.totalDownBet
			pd.CardSuitData = new(msg.CardSuitData)
			pd.CardSuitData.HandCardKeys = v.cardData.HandCardKeys
			pd.CardSuitData.PublicCardKeys = v.cardData.PublicCardKeys
			pd.CardSuitData.SuitPattern = v.cardData.SuitPattern
			pd.ResultMoney = v.resultMoney
			pd.BlindType = v.blindType
			pd.IsButton = v.IsButton
			pd.IsAllIn = v.IsAllIn
			pd.IsWinner = v.IsWinner
			pd.TimerCount = v.timerCount
			rd.PlayerData = append(rd.PlayerData, pd)
		}
	}
	return rd
}

//SetPlayerStatus 设置玩家状态
func (r *Room) SetPlayerStatus() {
	for _, v := range r.PlayerList {
		if v != nil {
			v.gameStep = emInGaming
		}
	}
}

func (r *Room) calc() (pots []handPot) {
	pots = calcPot(r.Chips)
	r.Pot = r.Pot[:]
	var ps []float64
	for _, pot := range pots {
		r.Pot = append(r.Pot, pot.Pot)
		ps = append(ps, pot.Pot)
	}
	return
}

func (r *Room) Each(start int, f func(p *Player) bool) {
	if start >= 9 {
		start = start % MaxPlayer
	}
	//房间最大限定人数
	end := (MaxPlayer + start - 1) % MaxPlayer
	i := start
	for ; i != end; i = (i + 1) % MaxPlayer {
		if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming && !f(r.PlayerList[i]) {
			return
		}
	}

	// end
	if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming {
		f(r.PlayerList[i])
	}
}

//Blind 小盲注和大盲注
func (r *Room) Blind(pos int32) *Player {

	i := int(pos) + 1
	for ; i < len(r.PlayerList); i = (i + 1) % MaxPlayer {
		if r.PlayerList[i] != nil {
			return r.PlayerList[i]
		}
	}
	return nil
}

//betting 小大盲下注
func (r *Room) betting(p *Player, blind float64) {
	log.Debug("玩家下注金额:%v", blind)
	//当前行动玩家
	r.activeSeat = p.chair
	//玩家筹码变动
	p.chips = p.chips - blind
	//本轮玩家下注额
	p.downBets = blind
	//玩家本局总下注额
	p.totalDownBet = p.totalDownBet + blind
	//总筹码变动
	r.potMoney = r.potMoney + blind

	action := &msg.PlayerAction_S2C{}
	action.Id = p.Id
	action.Chair = p.chair
	action.Chips = p.chips // 这里传入房间筹码金额
	action.DownBet = p.downBets
	action.PotMoney = r.potMoney
	action.ActionType = p.actStatus
	r.Broadcast(action)
}

//readyPlay 准备阶段
func (r *Room) readyPlay() {
	r.preChips = 0
	r.Each(0, func(p *Player) bool {
		p.downBets = 0
		p.HandValue = 0
		r.remain++
		return true
	})
}

//action 玩家行动
func (r *Room) action(pos int) {

	if r.allin+1 >= r.remain {
		return
	}

	if pos == 0 {
		pos = int((r.Banker)%MaxPlayer) + 1
	}
	//玩家行动
	waitTime := ActionTime
	ticker := time.Second * time.Duration(waitTime)

	for {
		var IsMove bool
		r.Each(pos, func(p *Player) bool {
			if r.remain <= 1 {
				return false
			}
			if p.chips == 0 { // p.chair == int32(skip) ||
				return true
			}
			if (r.preChips - p.downBets) == 0 {
				IsMove = true
				return true
			} else {
				IsMove = false
			}
			//3、行动玩家是根据庄家的下一位玩家
			r.activeSeat = p.chair
			log.Debug("行动玩家 ~ :%v", r.activeSeat)

			changed := &msg.PlayerActionChange_S2C{}
			room := r.RespRoomData()
			changed.RoomData = room
			r.Broadcast(changed)

			p.GetAction(r, ticker)

			if r.remain <= 1 {
				return false
			}

			action := &msg.PlayerAction_S2C{}
			action.Id = p.Id
			action.Chair = p.chair
			action.Chips = p.chips // 这里传入房间筹码金额
			action.DownBet = p.downBets
			action.PotMoney = r.potMoney
			action.ActionType = p.actStatus
			r.Broadcast(action)

			return true
		})
		if IsMove == true {
			break
		}

	}
}

//showdown 玩家摊牌结算
func (r *Room) ShowDown() {
	//1.统计玩家下注情况
	pots := r.calc()

	//2.计算分池
	for i, _ := range r.Chips {
		r.Chips[i] = 0
	}

	for _, pot := range pots {
		var maxO *Player
		for _, pos := range pot.OPos {
			o := r.PlayerList[pos]
			if o != nil && len(o.cards) > 0 {
				if maxO == nil {
					maxO = o
					continue
				}
				if o.HandValue > maxO.HandValue {
					maxO = o
				}
			}
		}

		var winners []int32

		for _, pos := range pot.OPos {
			o := r.PlayerList[pos]
			if o != nil && o.HandValue == maxO.HandValue && o.gameStep == emInGaming {
				winners = append(winners, o.chair)
			}
		}

		if len(winners) == 0 {
			log.Debug("no winner")
			return
		}

		for _, winner := range winners {
			r.Chips[winner] += pot.Pot / float64(len(winners))
		}
		log.Debug("ShowDown:%v,%v", pot.Pot, len(winners))
		win := float64(len(winners))
		if pot.Pot > win {
			r.Chips[winners[0]] += pot.Pot - win
		} else {
			r.Chips[winners[0]] = pot.Pot
		}
		//r.Chips[winners[0]] += pot.Pot % float64(len(winners)) // odd chips
	}

	for i, _ := range r.Chips {
		if r.PlayerList[i] != nil {
			r.PlayerList[i].chips += r.Chips[i]
		}
	}
}

//TimerTask 游戏准备阶段定时器任务
func (r *Room) ReadyTimerTask() {
	// 广播游戏准备时间
	ready := &msg.ReadyTime_S2C{}
	ready.ReadyTime = ReadyTime
	r.Broadcast(ready)

	// 玩家补充筹码
	r.PlayerAddChips()

	go func() {
		for range r.clock.C {
			r.counter++
			log.Debug("readyTime clock : %v ", r.counter)
			if r.counter == 3 {
				push := &msg.PushCardTime_S2C{}
				r.Broadcast(push)
			}
			if r.counter == ReadyTime {
				r.counter = 0
				ReadyTimeChan <- true
				return
			}
		}
	}()
}

//TimerTask 游戏开始定时器任务
func (r *Room) GameRunTimerTask() {
	go func() {
		select {
		case t := <-ReadyTimeChan:
			if t == true {
				// 游戏开始
				r.GameRunning()
				return
			}
		}
	}()
}

//TimerTask 重新开始定时器
func (r *Room) RestartGame() {
	go func() {
		for range r.clock.C {
			r.counter++
			log.Debug("settleTime clock : %v ", r.counter)
			if r.counter == SettleTime {
				r.counter = 0
				//开始新一轮游戏,重复调用StartGameRun函数
				r.StartGameRun()
				return
			}
		}
	}()
}
