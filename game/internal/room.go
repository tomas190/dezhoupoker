package internal

import (
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"fmt"
	"github.com/golang/glog"
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
	preChips float64  // 当前回合，上个玩家下注金额
	remain   int32    // 记录每个阶段玩家的下注的数量
	allin    int32    // allin玩家的数量
	Chips    []uint32 // 所有玩家本局下的总筹码,对应player玩家
	Pot      []uint32 // 奖池筹码数,第一项为主池，其他项(若存在)为边池
	Button   int32    // 庄家座位号
	SB       float64  // 小盲注
	BB       float64  // 大盲注

	clock *time.Ticker
}

const (
	MaxPlayer = 9
)

const (
	ActionTime = 15
)

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
	r.Chips = make([]uint32, MaxPlayer)
	r.Pot = []uint32{}
	r.Button = 0
	r.SB = rd.SB
	r.BB = rd.BB
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
	for _, p := range r.AllPlayer {
		if p != nil {
			p.SendMsg(msg)
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
	log.Debug("num :%v", num)
	return num
}

//TakeInRoomChips 玩家带入筹码
func (r *Room) TakeInRoomChips(p *Player) float64 {

	//1、如果玩家余额 大于房间最大设定金额 MaxTakeIn，则带入金额就设为 房间最大设定金额
	//2、如果玩家余额 小于房间最大设定金额 MaxTakeIn，则带入金额就设为 玩家的所有余额
	data := SetRoomConfig(r.cfgId)
	if p.Account > data.MaxTakeIn {
		p.Account = p.Account - data.MaxTakeIn
		return data.MaxTakeIn
	}
	Balance := p.Account
	p.Account = p.Account - p.Account
	return Balance
}

//FindAbleChair 寻找可用座位
func (r *Room) FindAbleChair(seatNum int32) int32 {
	// 先判断玩家历史座位是否已存在其他玩家，如果没有还是坐下历史座位
	if r.PlayerList[seatNum] == nil {
		return seatNum
	}

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
	// 遍历桌面玩家，踢掉筹码小于大盲玩家
	for _, v := range r.PlayerList {
		if v != nil && v.chips < r.BB {
			ErrorResp(v.ConnAgent, msg.ErrorMsg_ChipsInsufficient, "玩家筹码不足")
			v.PlayerExitRoom()
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

//ClearRoomData 清除房间数据
func (r *Room) ClearRoomData() {
	r.activeSeat = -1
	r.potMoney = 0
	r.publicCards = nil
	r.Status = msg.GameStep_Waiting
	r.preChips = 0
	r.remain = 0
	r.allin = 0
	r.Chips = make([]uint32, MaxPlayer)
	r.Pot = nil

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
	rd.ActionSeat = r.activeSeat
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
			pd.ActionStatus = v.actStatus
			pd.DownBets = v.downBets
			pd.CardSuitData = new(msg.CardSuitData)
			pd.CardSuitData.HandCardKeys = v.cardData.HandCardKeys
			pd.CardSuitData.PublicCardKeys = v.cardData.PublicCardKeys
			pd.CardSuitData.SuitPattern = v.cardData.SuitPattern
			pd.ResultMoney = v.resultMoney
			pd.BlindType = v.blindType
			pd.IsButton = v.IsButton
			pd.IsAllIn = v.IsAllIn
			pd.IsWinner = v.IsWinner
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

func (r *Room) Each(pos int, f func(p *Player) bool) {
	//房间最大限定人数
	volume := MaxPlayer
	end := (volume + pos - 1) % volume
	i := pos
	for ; i != end; i = (i + 1) % volume {
		if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming && !f(r.PlayerList[i]) {
			return
		}
	}

	// end
	if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming {
		log.Debug("Player : %v", r.PlayerList[i])
		f(r.PlayerList[i])
	}
}

//Blind 小盲注和大盲注
func (r *Room) Blind(pos int32) *Player {
	max := MaxPlayer
	start := int(pos+1) % max
	for i := start; i < max; i = (i + 1) % max {
		if r.PlayerList[i] != nil && r.PlayerList[pos] != r.PlayerList[i] {
			return r.PlayerList[i]
		}
	}
	return nil
}

//betting 小大盲下注
func (r *Room) betting(p *Player, blind float64) {
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
	//玩家筹码池
	r.Chips[p.chair] += uint32(blind)
}

//calc 筹码池
func (r *Room) calc() (pots []handPot) {
	pots = calcPot(r.Chips)
	r.Pot = r.Pot[:]
	var ps []uint32
	for _, pot := range pots {
		r.Pot = append(r.Pot, pot.Pot)
		ps = append(ps, pot.Pot)
	}
	return
}

//readyPlay 准备阶段
func (r *Room) readyPlay() {
	r.preChips = 0
	r.Each(0, func(p *Player) bool {
		p.HandValue = 0
		p.downBets = 0
		r.remain++
		return true
	})
}

//action 玩家行动
func (r *Room) action(pos int) {

	if r.allin+1 >= r.remain {
		return
	}
	//var skip uint32
	//从庄家的下家开始下注
	if pos == 0 {
		pos = int(r.Button%MaxPlayer + 1)
	}

	r.Each(pos, func(p *Player) bool {
		//3、行动玩家是根据庄家的下一位玩家
		r.activeSeat = p.chair
		log.Debug("行动玩家 ~ :%v", r.activeSeat)

		//changed := &pb_msg.ActionPlayerChangedS2C{}
		//room := p.RspRoomData()
		//changed.RoomData = room
		//gr.Broadcast(changed)

		if r.remain <= 1 {
			return false
		}
		if p.chips == 0 { // p.chair == int32(skip) ||
			return true
		}

		//玩家行动
		waitTime := ActionTime
		ticker := time.Second * time.Duration(waitTime)

		p.GetAction(r, ticker)

		if r.remain <= 1 {
			return false
		}

		action := &msg.PlayerAction_S2C{}
		action.Id = p.Id
		action.Account = p.Account
		action.PotMoney = r.potMoney
		p.SendMsg(action)

		return true
	})
}

//showdown 玩家摊牌结算
func (r *Room) showdown() {
	pots := r.calc()

	for i := range r.Chips {
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

		var winners []uint8

		for _, pos := range pot.OPos {
			o := r.PlayerList[pos]
			if o != nil && o.HandValue == maxO.HandValue && o.gameStep == emInGaming { //&& o.IsGameing()
				winners = append(winners, uint8(o.chair))
			}
		}

		if len(winners) == 0 {
			glog.Errorln("!!!no winners!!!")
			return
		}

		for _, winner := range winners {
			r.Chips[winner] += pot.Pot / uint32(len(winners))
		}
		r.Chips[winners[0]] += pot.Pot % uint32(len(winners)) // odd chips
	}

	for i := range r.Chips {
		if r.PlayerList[i] != nil {
			r.PlayerList[i].chips += float64(r.Chips[i])
		}
	}
}
