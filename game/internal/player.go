package internal

import (
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"math/rand"
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
	RoundId  string

	cards           algorithm.Cards  // 牌型数据
	chips           float64          // 玩家筹码
	roomChips       float64          // 玩家房间筹码
	chair           int32            // 座位号(站起为-1)
	standUPNum      int32            // 站起玩家局数(站起5局直接踢出)
	actStatus       msg.ActionStatus // 玩家行动状态
	gameStep        GameStatus       // 玩家游戏状态
	downBets        float64          // 下注金额
	lunDownBets     float64          // 每轮总下注
	totalDownBet    float64          // 下注总金额
	cardData        msg.CardSuitData // 卡牌数据和类型
	resultMoney     float64          // 结算金额
	WinResultMoney  float64          // 本局赢钱金额
	LoseResultMoney float64          // 本局输钱金额
	blindType       msg.BlindType    // 盲注类型
	IsAllIn         bool             // 是否全压
	IsButton        bool             // 是否庄家
	IsWinner        bool             // 是否赢家
	IsOnline        bool             // 是否在线
	actTime         int32            // 当前行动时间
	IsTimeOutFold   bool             // 是否超时弃牌
	timerCount      int32            // 玩家行动计时

	HandValue uint32
	action    chan msg.ActionStatus // 玩家行动命令

	IsRobot bool // 是否机器人
}

func (p *Player) Init() {
	p.RoundId = ""
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
	p.HandValue = 0
	p.IsOnline = true
	p.IsTimeOutFold = false
	p.timerCount = 0
	p.action = make(chan msg.ActionStatus)
	p.IsRobot = false
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
	data.LunDownBets = p.lunDownBets
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
	if p.IsRobot == false {
		for {
			select {
			case x := <-p.action:
				switch x {
				case msg.ActionStatus_RAISE:
					p.actStatus = msg.ActionStatus_RAISE
					p.chips -= p.downBets
					r.preChips = p.lunDownBets
					r.potMoney += p.downBets
					IsRaised = true
				case msg.ActionStatus_CALL:
					p.actStatus = msg.ActionStatus_CALL
					p.chips -= p.downBets
					r.preChips = p.lunDownBets
					r.potMoney += p.downBets
				case msg.ActionStatus_CHECK:
					p.actStatus = msg.ActionStatus_CHECK
				case msg.ActionStatus_FOLD:
					p.actStatus = msg.ActionStatus_FOLD
					p.gameStep = emNotGaming
					r.remain--
				case msg.ActionStatus_ALLIN:
					p.actStatus = msg.ActionStatus_ALLIN
					p.chips -= p.downBets
					r.preChips = p.lunDownBets
					r.potMoney += p.downBets
				}

				r.Chips[p.chair] += p.chips

				if p.chips == 0 {
					p.actStatus = msg.ActionStatus_ALLIN
					p.IsAllIn = true
					r.allin++
					r.IsHaveAllin = true
				}
				//玩家本局下注的总筹码数
				//r.Chips[p.chair] += uint32(r.preChips)
				return IsRaised


			case <-after.C:
				log.Debug("超时行动弃牌: %v", time.Now().Format("2006-01-02 15:04:05"))

				//ErrorResp(p.ConnAgent, msg.ErrorMsg_UserTimeOutFoldCard, "玩家超时弃牌")

				p.gameStep = emNotGaming
				p.actStatus = msg.ActionStatus_FOLD
				p.IsTimeOutFold = true
				r.remain--
				return IsRaised
			}
		}
	} else {
		log.Debug("机器人开始下注~")
		var actionType msg.ActionStatus
		callMoney := r.preChips - p.lunDownBets
		if callMoney > 0 {
			// 当跟注金额 大于筹码时
			if callMoney > p.chips {
				callBets := []int32{1, 2, 1, 1, 1} // 1为弃牌,2 全压
				rand.Seed(time.Now().UnixNano())
				callNum := rand.Intn(len(callBets))
				if callBets[callNum] == 1 {
					actionType = msg.ActionStatus_FOLD
				}
				if callBets[callNum] == 2 {
					actionType = msg.ActionStatus_ALLIN
					p.downBets = p.chips
					p.lunDownBets += p.chips
					p.totalDownBet += p.chips
				}
			} else {
				callBets := []int32{1, 1, 3, 1, 1,} // 1跟注,2加注,3弃牌,4全压
				rand.Seed(time.Now().UnixNano())
				callNum := rand.Intn(len(callBets))
				if r.Status == msg.GameStep_PreFlop {
					callBets[callNum] = 1
				}
				if callBets[callNum] == 1 {
					actionType = msg.ActionStatus_CALL
					p.downBets = callMoney
					p.lunDownBets += callMoney
					p.totalDownBet += callMoney
				}
				if callBets[callNum] == 3 {
					actionType = msg.ActionStatus_FOLD
				}
			}
		} else {
			callBets := []int32{1, 2, 1, 1, 1} // 1为让牌,2为加注
			rand.Seed(time.Now().UnixNano())
			callNum := rand.Intn(len(callBets))
			if callBets[callNum] == 1 {
				actionType = msg.ActionStatus_CHECK
			}
			if callBets[callNum] == 2 {
				downBet := []float64{0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1}
				rand.Seed(time.Now().UnixNano())
				num := rand.Intn(len(downBet))
				if p.chips > downBet[num] && r.Status != msg.GameStep_PreFlop {
					actionType = msg.ActionStatus_RAISE
					p.downBets = downBet[num]
					p.lunDownBets += downBet[num]
					p.totalDownBet += downBet[num]
				} else {
					actionType = msg.ActionStatus_CHECK
				}
			}
		}
		var timerSlice []int32
		if actionType == 1 {
			timerSlice = []int32{4, 6, 8, 5, 6}
		}
		if actionType == 2 {
			timerSlice = []int32{3, 6, 4, 6, 5, 8, 4}
			if r.Status != msg.GameStep_PreFlop {
				timerSlice = []int32{3, 6, 4, 6, 15, 5, 8, 4}
			}
		}
		if actionType == 3 {
			timerSlice = []int32{3, 5, 4, 3, 2, 4, 6, 3}
		}
		if actionType == 4 {
			timerSlice = []int32{4, 6, 8, 5, 6, 4}
		}
		if actionType == 5 {
			timerSlice = []int32{6, 8, 7, 5, 6, 9}
		}
		rand.Seed(time.Now().UnixNano())
		num := rand.Intn(len(timerSlice))
		time.Sleep(time.Second * time.Duration(timerSlice[num]))

		if timerSlice[num] == 15 {
			actionType = msg.ActionStatus_CHECK
			p.IsTimeOutFold = true
			log.Debug("超时行动弃牌: %v", time.Now().Format("2006-01-02 15:04:05"))
		}

		if actionType == msg.ActionStatus_RAISE {
			p.actStatus = msg.ActionStatus_RAISE
			p.chips -= p.downBets
			r.preChips = p.lunDownBets
			r.potMoney += p.downBets
			IsRaised = true
		}
		if actionType == msg.ActionStatus_CALL {
			p.actStatus = msg.ActionStatus_CALL
			p.chips -= p.downBets
			r.preChips = p.lunDownBets
			r.potMoney += p.downBets
		}
		if actionType == msg.ActionStatus_CHECK {
			p.actStatus = msg.ActionStatus_CHECK
		}
		if actionType == msg.ActionStatus_FOLD {
			p.actStatus = msg.ActionStatus_FOLD
			p.gameStep = emNotGaming
			r.remain--
		}
		if actionType == msg.ActionStatus_ALLIN {
			p.actStatus = msg.ActionStatus_ALLIN
			p.chips -= p.downBets
			r.preChips = p.lunDownBets
			r.potMoney += p.downBets
		}
		r.Chips[p.chair] += p.chips

		if p.chips == 0 {
			p.actStatus = msg.ActionStatus_ALLIN
			p.IsAllIn = true
			r.allin++
			r.IsHaveAllin = true
		}
		return IsRaised
	}
}
