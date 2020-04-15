package internal

import (
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"time"
)

type GameStatus int32

const (
	emNotGaming GameStatus = 0 // 没有在游戏中
	emInGaming  GameStatus = 1 // 正在游戏中
)

type Player struct {
	// 玩家代理链接
	ConnAgent gate.Agent

	Id       string
	NickName string
	HeadImg  string
	Account  float64
	Password string
	Token    string

	cards         algorithm.Cards  // 牌型数据
	chips         float64          // 玩家筹码
	roomChips     float64          // 玩家房间筹码
	chair         int32            // 座位号(站起为-1)
	standUPNum    int32            // 站起玩家局数(站起5局直接踢出)
	actStatus     msg.ActionStatus // 玩家行动状态
	gameStep      GameStatus       // 玩家游戏状态
	downBets      float64          // 下注金额
	totalDownBet  float64          // 下注总金额
	cardData      msg.CardSuitData // 卡牌数据和类型
	resultMoney   float64          // 结算金额
	blindType     msg.BlindType    // 盲注类型
	IsAllIn       bool             // 是否全压
	IsButton      bool             // 是否庄家
	IsWinner      bool             // 是否赢家
	IsOnline      bool             // 是否在线
	actTime       int32            // 当前行动时间
	IsTimeOutFold bool             // 是否超时弃牌
	timerCount    int32            // 玩家行动计时

	HandValue uint32
	action    chan msg.ActionStatus // 玩家行动命令
}

func NewPlayer() *Player {
	return &Player{
		chips:         0,
		roomChips:     0,
		chair:         0,
		standUPNum:    0,
		actStatus:     msg.ActionStatus_WAITING,
		gameStep:      emNotGaming,
		downBets:      0,
		totalDownBet:  0,
		cardData:      msg.CardSuitData{},
		resultMoney:   0,
		blindType:     msg.BlindType_No_Blind,
		IsAllIn:       false,
		IsButton:      false,
		IsWinner:      false,
		HandValue:     0,
		IsOnline:      true,
		IsTimeOutFold: false,
		timerCount:    0,
		action:        make(chan msg.ActionStatus),
	}
}

//SendMsg 玩家向客户端发送消息
func (p *Player) SendMsg(msg interface{}) {
	if p.ConnAgent != nil {
		p.ConnAgent.WriteMsg(msg)
	}
}

//RespPlayerData 返回玩家数据
func (p *Player) RespPlayerData() *msg.PlayerData {
	data := &msg.PlayerData{}
	data.PlayerInfo = new(msg.PlayerInfo)
	data.PlayerInfo.Id = p.Id
	data.PlayerInfo.NickName = p.NickName
	data.PlayerInfo.HeadImg = p.HeadImg
	data.PlayerInfo.Account = p.Account
	data.Chair = p.chair
	data.StandUPNum = p.standUPNum
	data.Chips = p.chips
	data.RoomChips = p.roomChips
	data.ActionStatus = p.actStatus
	data.GameStep = int32(p.gameStep)
	data.DownBets = p.downBets
	data.TotalDownBet = p.totalDownBet
	data.CardSuitData = new(msg.CardSuitData)
	data.CardSuitData.HandCardKeys = p.cardData.HandCardKeys
	data.CardSuitData.PublicCardKeys = p.cardData.PublicCardKeys
	data.CardSuitData.SuitPattern = p.cardData.SuitPattern
	data.ResultMoney = p.resultMoney
	data.BlindType = p.blindType
	data.IsButton = p.IsButton
	data.IsAllIn = p.IsAllIn
	data.IsWinner = p.IsWinner
	data.TimerCount = p.timerCount
	return data
}

func (p *Player) GetAction(r *Room, timeout time.Duration) bool {

	log.Debug("玩家行动时间: %v", time.Now().Format("2006-01-02 15:04:05"))

	p.timerCount = 0 // todo

	after := time.NewTicker(timeout)
	var IsRaised bool
	for {
		select {
		case x := <-p.action:
			switch x {
			case msg.ActionStatus_RAISE:
				p.actStatus = msg.ActionStatus_RAISE
				// 玩家筹码
				p.chips -= p.downBets
				// 本局玩家下注总金额
				p.totalDownBet += p.downBets
				// 房间上个玩家下注金额
				r.preChips = p.downBets
				// 总筹码
				r.potMoney += p.downBets
				IsRaised = true
			case msg.ActionStatus_CALL:
				p.actStatus = msg.ActionStatus_CALL
				p.chips -= p.downBets
				p.totalDownBet += p.downBets
				r.potMoney += p.downBets
			case msg.ActionStatus_CHECK:
				p.actStatus = msg.ActionStatus_CHECK
			case msg.ActionStatus_FOLD:
				p.actStatus = msg.ActionStatus_FOLD
				p.gameStep = emNotGaming
				r.remain--
			}
			r.Chips[p.chair] += p.chips

			if p.chips == 0 {
				p.IsAllIn = true
				r.allin++
			}
			//玩家本局下注的总筹码数
			//r.Chips[p.chair] += uint32(r.preChips)
			return IsRaised


		case <-after.C:
			log.Debug("超时行动弃牌: %v", time.Now().Format("2006-01-02 15:04:05"))

			ErrorResp(p.ConnAgent, msg.ErrorMsg_UserTimeOutFoldCard, "玩家超时弃牌")

			p.gameStep = emNotGaming
			p.actStatus = msg.ActionStatus_FOLD
			p.IsTimeOutFold = true
			r.remain--
			return IsRaised
		}
	}
}
