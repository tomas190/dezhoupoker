package internal

import (
	"dezhoupoker/msg"
	"github.com/name5566/leaf/gate"
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

	chips        float64          // 玩家筹码
	chair        int32            // 座位号
	historyChair int32            // 历史座位号
	standUPNum   int32            // 站起玩家局数(站起5局直接踢出)
	actStatus    msg.ActionStatus // 玩家行动状态
	gameStep     GameStatus       // 玩家游戏状态  (发牌InGaming,弃牌NotGaming)
	downBets     float64          // 下注金额
	totalDownBet float64          // 下注总金额
	cardData     msg.CardSuitData // 卡牌数据和类型
	resultMoney  float64          // 结算金额
	blindType    msg.BlindType    // 盲注类型
	IsAllIn      bool             // 是否全压
	IsButton     bool             // 是否庄家
	IsWinner     bool             // 是否赢家
	actTime      int32            // 当前行动时间
}

//SendMsg 玩家向客户端发送消息
func (p *Player) SendMsg(msg interface{}) {
	if p.ConnAgent != nil {
		p.ConnAgent.WriteMsg(msg)
	}
}
