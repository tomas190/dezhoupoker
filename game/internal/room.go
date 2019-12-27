package internal

import (
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

	preChips float64   // 当前回合，上个玩家下注金额
	remain   int32     // 记录每个阶段玩家的下注的数量
	allin    int32     // allin玩家的数量
	Chips    []float64 // 所有玩家本局下的总筹码,对应player玩家
	Pot      []float64 // 奖池筹码数,第一项为主池，其他项(若存在)为边池
	Button   int32     // 庄家座位号
	SB       float64   // 小盲注
	BB       float64   // 大盲注

	clock *time.Ticker
}

const (
	MaxPlayer = 9
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
	r.Chips = make([]float64, MaxPlayer)
	r.Pot = []float64{}
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
	return len(r.PlayerList) < MaxPlayer
}

//PlayerLength 房间玩家人数
func (r *Room) PlayerLength() int32 {
	var num int32
	for _, v := range r.PlayerList {
		if v != nil {
			num++
		}
	}
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

//KickPlayer 踢掉筹码小于大盲玩家
func (r *Room) KickPlayer() {
	for _, v := range r.PlayerList {
		if v != nil && v.chips < r.BB {
			ErrorResp(v.ConnAgent, msg.ErrorMsg_ChipsInsufficient, "玩家筹码不足")
			v.PlayerExitRoom()
		}
	}
}

//RespRoomData 返回房间数据
func (r *Room) RespRoomData(p *Player) *msg.RoomData {
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
			if p.Id == v.Id {
				pd.CardSuitData = new(msg.CardSuitData)
				pd.CardSuitData.HandCardKeys = v.cardData.HandCardKeys
				pd.CardSuitData.PublicCardKeys = v.cardData.PublicCardKeys
				pd.CardSuitData.SuitPattern = v.cardData.SuitPattern
			}
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
