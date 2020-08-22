package internal

import (
	"dezhoupoker/conf"
	"dezhoupoker/game/internal/algorithm"
	"dezhoupoker/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"time"
)

type RoomStatus int32

const (
	RoomStatusNone RoomStatus = 1 // 房间等待状态
	RoomStatusRun  RoomStatus = 2 // 房间运行状态
	RoomStatusOver RoomStatus = 3 // 房间结束状态
)

type Room struct {
	roomId     string
	cfgId      string    // 房间配置ID
	PlayerList []*Player // 座位玩家列表，最高9人
	AllPlayer  []*Player // 房间，包括站起玩家座位号为-1

	activeSeat  int32        // 当前正在行动玩家座位号
	activeId    string       // 当前行动玩家Id
	minRaise    float64      // 加注最小值
	potMoney    float64      // 桌面注池金额
	publicCards []int32      // 桌面公牌
	RoomStat    RoomStatus   // 房间运行状态状态
	Status      msg.GameStep // 房间当前阶段  就直接判断是否在等待状态

	Cards      algorithm.Cards
	preChips   float64   // 当前回合, 上个玩家下注金额
	IsShowDown int32     // 0 为摊牌, 1 为不摊牌
	remain     int32     // 记录每个阶段玩家的下注的数量
	allin      int32     // allin玩家的数量
	Chips      []float64 // 所有玩家本局下的总筹码,奖池筹码数,第一项为主池，其他项(若存在)为边池
	Banker     int32     // 庄家座位号
	SB         float64   // 小盲注
	BB         float64   // 大盲注

	counter int32
	clock   *time.Ticker

	IsHaveAllin bool     // 是否有玩家allin
	UserLeave   []string // 用户是否在房间

	PiPeiList   []*Player // 匹配列表
	StandUpList []*Player // 站起列表
}

const (
	MaxPlayer = 9
)

const (
	ReadyTime      = 6  // 开始准备时间
	SettleTime     = 5  // 游戏结算时间
	ActionTime     = 15 // 玩家行动时间
	ActionWaitTime = 2  // 行动等待时间
)

var ReadyTimeChan chan bool

// 行动时间间隔
var ActionTimeChan chan bool

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
	r.activeId = ""
	r.minRaise = rd.BB
	r.potMoney = 0
	r.publicCards = nil
	r.RoomStat = RoomStatusNone
	r.Status = msg.GameStep_Waiting
	r.preChips = 0
	r.IsShowDown = 0
	r.remain = 0
	r.allin = 0
	r.Chips = make([]float64, MaxPlayer)
	r.Banker = 0
	r.SB = rd.SB
	r.BB = rd.BB

	r.counter = 0
	r.clock = time.NewTicker(time.Second)

	r.IsHaveAllin = false

	ReadyTimeChan = make(chan bool)
	ActionTimeChan = make(chan bool)

	r.PiPeiList = make([]*Player, 0)
	r.StandUpList = make([]*Player, 0)
}

//BroadCastExcept 向指定玩家之外的玩家广播
func (r *Room) BroadCastExcept(msg interface{}, except *Player) {
	for _, p := range r.AllPlayer {
		if p != nil && except.Id != p.Id && p.IsRobot == false {
			p.SendMsg(msg)
		}
	}
}

//Broadcast 广播消息
func (r *Room) Broadcast(msg interface{}) {
	for _, v := range r.AllPlayer {
		if v != nil && v.IsRobot == false {
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
		if v != nil && v.chair != -1 { //todo
			num++
		}
	}
	log.Debug("房间号:%v  玩家人数:%v", r.roomId, num)
	return num
}

//PlayerLength 房间玩家人数
func (r *Room) AllPlayerLength() int32 {
	var num int32
	for _, v := range r.AllPlayer {
		if v != nil {
			num++
		}
	}
	log.Debug("当前房间所有玩家人数: %v", num)
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
		p.roomChips = data.MaxTakeIn
		p.roomChips -= p.chips
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
	// 清理断线玩家
	for _, uid := range r.UserLeave {
		for _, v := range r.PlayerList {
			if v != nil && v.Id == uid {
				//玩家断线的话，退出房间信息，也要断开链接
				if v.IsOnline == true {
					v.PlayerExitRoom()
					log.Debug("踢出断线玩家~")
				} else {
					v.PlayerExitRoom()
					hall.UserRecord.Delete(v.Id)
					c4c.UserLogoutCenter(v.Id, v.Password, v.Token)
					leaveHall := &msg.Logout_S2C{}
					v.ConnAgent.WriteMsg(leaveHall)
					v.IsOnline = false
					log.Debug("踢出房间断线玩家 : %v", v.Id)
				}
			}
		}
	}

	// 遍历桌面玩家，踢掉玩家筹码和房间小于房间最小带入金额
	for _, v := range r.PlayerList { // 玩家筹码为0怎么办
		if v != nil {
			rd := SetRoomConfig(r.cfgId)
			if v.chips+v.roomChips < rd.BB {
				//ErrorResp(v.ConnAgent, msg.ErrorMsg_ChipsInsufficient, "玩家筹码不足")
				v.PlayerExitRoom()
				log.Debug("踢掉玩家筹码和房间小于房间最小带入金额:%v", v)
			}
		}
	}

	// 遍历站起玩家，是否在该房间站起超时
	for _, v := range r.AllPlayer {
		if v != nil && v.IsRobot == false && v.chair == -1 {
			v.standUPNum++
			if v.standUPNum >= 6 {
				//ErrorResp(v.ConnAgent, msg.ErrorMsg_UserStandUpTimeOut, "玩家站起超时")
				v.PlayerExitRoom()
				log.Debug("玩家站起次数6次:%v", v)
			}
		}
	}

	for _, v := range r.PlayerList {
		if v != nil && v.chair == -1 {
			r.PlayerList[v.chair] = nil
		}
	}
}

// 玩家补充筹码
func (r *Room) PlayerAddChips() {
	var limit float64
	var limitMoney float64
	if r.cfgId == "0" {
		limit = 1
		limitMoney = 10
	} else if r.cfgId == "1" {
		limit = 5
		limitMoney = 50
	} else if r.cfgId == "2" {
		limit = 30
		limitMoney = 300
	} else if r.cfgId == "3" {
		limit = 100
		limitMoney = 1000
	}

	for _, v := range r.PlayerList {
		if v != nil && v.chips < limit {
			if v.roomChips > limitMoney {
				v.roomChips -= limitMoney
				v.chips += limitMoney
				addChips := &msg.AddChips_S2C{}
				addChips.Chair = v.chair
				addChips.AddChips = limitMoney
				addChips.Chips = v.chips
				addChips.RoomChips = v.roomChips
				addChips.SysBuyChips = 1
				v.SendMsg(addChips)
			} else {
				// 自动补充筹码
				money := v.roomChips
				v.roomChips = 0
				v.chips += money
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
			log.Debug("行动超时玩家站起:%v", v.Id)
			v.StandUpTable()
		}
	}
}

//ClearRoomData 清除房间数据
func (r *Room) ClearRoomData() {
	r.activeSeat = -1
	r.activeId = ""
	r.potMoney = 0
	r.publicCards = nil
	r.preChips = 0
	r.remain = 0
	r.allin = 0
	r.IsShowDown = 0
	r.IsHaveAllin = false
	r.Chips = make([]float64, MaxPlayer)

	for _, v := range r.AllPlayer {
		if v != nil {
			v.actStatus = msg.ActionStatus_WAITING
			v.gameStep = emNotGaming
			v.downBets = 0
			v.lunDownBets = 0
			v.totalDownBet = 0
			v.cardData = msg.CardSuitData{}
			v.resultMoney = 0
			v.WinResultMoney = 0
			v.LoseResultMoney = 0
			v.blindType = msg.BlindType_No_Blind
			v.IsAllIn = false
			v.IsWinner = false
			v.IsButton = false
			v.IsInGame = false
			v.IsLeaveR = true
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
	rd.IsShowDown = r.IsShowDown
	rd.IsHaveAllin = r.IsHaveAllin
	rd.ActiveId = r.activeId
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
			pd.LunDownBets = v.lunDownBets
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
			pd.IsInGame = v.IsInGame
			pd.IsStandUp = v.IsStandUp
			pd.IsLeaveR = v.IsLeaveR
			pd.TimerCount = v.timerCount
			rd.PlayerData = append(rd.PlayerData, pd)
		}
	}
	for _, v := range r.AllPlayer {
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
			pd.LunDownBets = v.lunDownBets
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
			pd.IsInGame = v.IsInGame
			pd.IsStandUp = v.IsStandUp
			pd.IsLeaveR = v.IsLeaveR
			pd.TimerCount = v.timerCount
			rd.AllPlayer = append(rd.AllPlayer, pd)
		}
	}
	return rd
}

//SetPlayerStatus 设置玩家状态
func (r *Room) SetPlayerStatus() {
	for _, v := range r.PlayerList {
		if v != nil {
			v.gameStep = emInGaming
			v.IsInGame = true
			//log.Debug("设置玩家状态:%v,%v", v.Id, v.gameStep)
		}
	}
}

func (r *Room) CalBet() {
	for i, v := range r.PlayerList {
		if v != nil {
			r.Chips[i] = v.totalDownBet
		} else {
			r.Chips[i] = 0
		}
		//fmt.Printf("i: %d bet:%d\n",i,this.Bets[i])
	}
}

func (r *Room) PrintPots(pots []PotNode) {
	for k, v := range pots {
		fmt.Printf("分池%d:(%f) ", k, v.Bet)
		//fmt.Print("参与玩家(座位号):")
		for _, pos := range v.Pos {
			fmt.Printf("%d ", pos)
		}
		fmt.Println()
	}
}

func (r *Room) Each(pos int, f func(p *Player) bool) {
	num := 0
	i := pos
	for ; i < MaxPlayer; i = (i + 1) % MaxPlayer {
		if i == pos {
			num++
			if num == 2 {
				return
			}
		}
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
		if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming {
			return r.PlayerList[i]
		}
	}
	return nil
}

//betting 小大盲下注
func (r *Room) betting(p *Player, blind float64) {
	log.Debug("玩家盲注下注金额:%v", blind)
	//当前行动玩家
	r.activeSeat = p.chair
	r.activeId = p.Id
	//玩家筹码变动
	p.chips = p.chips - blind
	//本轮玩家下注额
	p.downBets = blind
	// 本轮游戏总下注金额
	p.lunDownBets += blind
	//玩家本局总下注额
	p.totalDownBet = p.totalDownBet + blind
	//总筹码变动
	r.potMoney += blind
	r.preChips = p.lunDownBets

	action := &msg.PlayerAction_S2C{}
	action.Id = p.Id
	action.Chair = p.chair
	action.Chips = p.chips // 这里传入房间筹码金额
	action.DownBet = p.lunDownBets
	action.PreChips = r.preChips
	action.PotMoney = r.potMoney
	action.ActionType = p.actStatus
	r.Broadcast(action)
}

//readyPlay 准备阶段
func (r *Room) readyPlay() {
	r.preChips = 0
	r.remain = 0
	r.IsHaveAllin = false
	r.Each(0, func(p *Player) bool {
		p.downBets = 0
		p.lunDownBets = 0
		p.HandValue = 0
		p.IsAction = false
		r.remain++
		return true
	})
	for i := 0; i < len(r.AllPlayer); i++ {
		if r.AllPlayer[i] != nil {
			r.AllPlayer[i].downBets = 0
			r.AllPlayer[i].lunDownBets = 0
		}
	}
}

func (r *Room) Action(pos int) {
	if r.allin+1 >= r.remain {
		return
	}

	actionPos := pos
	if actionPos >= MaxPlayer {
		actionPos = actionPos % MaxPlayer
	}

	for {
		var IsRaised bool
		i := actionPos
		var num int
		for ; i < len(r.PlayerList); i = (i + 1) % MaxPlayer {
			if i == actionPos {
				num++
			}
			if num == 2 {
				break
			}
			if r.PlayerList[i] != nil && r.PlayerList[i].gameStep == emInGaming && r.PlayerList[i].IsAction == false {
				p := r.PlayerList[i]

				if r.remain <= 1 {
					return
				}

				//log.Debug("当前行动玩家金额为:%v", p.chips)
				if p.chips == 0 {
					p.IsAction = true
					continue
				}

				//玩家行动
				waitTime := ActionTime
				ticker := time.Second * time.Duration(waitTime)

				r.activeSeat = p.chair
				r.activeId = p.Id
				p.timerCount = 0 // todo

				//log.Debug("行动玩家 ~ :%v", r.activeSeat)

				changed := &msg.PlayerActionChange_S2C{}
				room := r.RespRoomData()
				changed.RoomData = room
				r.Broadcast(changed)

				IsRaised = p.GetAction(r, ticker)
				//if p.IsRobot == false {
				//	log.Debug("真实玩家行动:%v,%v", p.Id, p.lunDownBets)
				//}

				action := &msg.PlayerAction_S2C{}
				action.Id = p.Id
				action.Chair = p.chair
				action.Chips = p.chips // 这里传入房间筹码金额
				action.DownBet = p.lunDownBets
				action.PreChips = r.preChips
				action.PotMoney = r.potMoney
				action.ActionType = p.actStatus
				r.Broadcast(action)

				if IsRaised == true {
					actionPos = int(r.activeSeat)
					break
				}

				if r.allin >= r.remain {
					return
				}
				if r.remain <= 1 {
					return
				}
			}
		}
		if IsRaised == true {
			for _, v := range r.PlayerList {
				if v != nil && v.chair != int32(actionPos) {
					v.IsAction = false
				}
			}
		} else {
			return
		}
	}
}

//showdown 玩家摊牌结算
func (r *Room) ShowDown() {
	//1.统计玩家下注情况
	r.CalBet()

	//2.计算分池
	pots := CalPots(r.Chips)

	//r.PrintPots(pots)

	for i, _ := range r.Chips {
		r.Chips[i] = 0
	}

	for _, pot := range pots {

		var maxPlayer *Player
		for _, pos := range pot.Pos {
			player := r.PlayerList[pos]
			if player != nil && player.gameStep == emInGaming {
				if maxPlayer == nil {
					maxPlayer = player
					continue
				}
				if player.HandValue > maxPlayer.HandValue {
					maxPlayer = player
				}
			}
		}
		var winners []int
		for _, pos := range pot.Pos {
			player := r.PlayerList[pos]
			if player != nil && player.gameStep == emInGaming && player.HandValue == maxPlayer.HandValue {
				//log.Debug("比牌手值:%v,%v", player.HandValue, maxPlayer.HandValue)
				winners = append(winners, pos)
			}
		}
		if len(winners) == 0 {
			fmt.Println("no winners")
			return
		}
		//多个玩家组合牌相等平分奖池
		for _, pos := range winners {
			r.Chips[pos] += pot.Bet / float64(len(winners))
		}

	}
	//fmt.Println("比牌结果:")
	for i, v := range r.Chips {
		player := r.PlayerList[i]
		if player == nil {
			continue
		}
		if v > 0 {
			player := r.PlayerList[i]
			player.WinResultMoney = v
			player.resultMoney += v
			if v-player.totalDownBet > 0 {
				player.IsWinner = true
			}
		}
		//fmt.Printf("uid:%s seat:%d result:%s win:%f chips:%f\n", player.Id, player.chair, player.cardData.SuitPattern, v, player.chips)
	}
}

func (r *Room) ResultMoney() {
	sur := &SurplusPoolDB{}
	sur.UpdateTime = time.Now()
	sur.TimeNow = time.Now().Format("2006-01-02 15:04:05")
	sur.Rid = r.roomId
	sur.PlayerNum = LoadPlayerCount()

	surPool := FindSurplusPool()
	if surPool != nil {
		sur.HistoryWin = surPool.HistoryWin
		sur.HistoryLose = surPool.HistoryLose
	}

	for i := 0; i < len(r.PlayerList); i++ {
		if r.PlayerList[i] != nil && r.PlayerList[i].totalDownBet > 0 {
			p := r.PlayerList[i]
			if r.PlayerList[i].IsRobot == false {
				p.resultMoney -= p.totalDownBet

				nowTime := time.Now().Unix()
				p.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.roomId)
				var taxMoney float64
				if p.resultMoney > 0 {
					taxMoney = p.resultMoney * taxRate
					p.WinResultMoney = p.resultMoney
					winReason := "德州扑克赢钱"
					c4c.UserSyncWinScore(p, nowTime, p.RoundId, winReason)
					sur.HistoryWin += p.WinResultMoney
					sur.TotalWinMoney += p.WinResultMoney
				}
				if p.resultMoney < 0 {
					p.LoseResultMoney = p.resultMoney
					loseReason := "德州扑克输钱"
					c4c.UserSyncLoseScore(p, nowTime, p.RoundId, loseReason)
					sur.HistoryLose -= p.LoseResultMoney // -- = +
					sur.TotalLoseMoney -= p.LoseResultMoney
				}

				// 这里是玩家金额扣税
				p.resultMoney -= taxMoney

				if p.resultMoney > 0 {
					p.chips += p.totalDownBet
					p.chips += p.resultMoney
				}

				// 插入盈余池数据
				if sur.TotalWinMoney != 0 || sur.TotalLoseMoney != 0 {
					InsertSurplusPool(sur)
				}
				// 跑马灯
				if p.resultMoney > PaoMaDeng {
					c4c.NoticeWinMoreThan(p.Id, p.NickName, p.resultMoney)
				}
				// 插入运营数据
				if sur.TotalWinMoney != 0 || sur.TotalLoseMoney != 0 {
					data := &PlayerDownBetRecode{}
					data.Id = p.Id
					data.GameId = conf.Server.GameID
					data.RoundId = p.RoundId
					data.RoomId = r.roomId
					data.DownBetInfo = p.totalDownBet
					data.DownBetTime = nowTime
					data.CardResult = msg.CardSuitData{}
					data.CardResult.SuitPattern = p.cardData.SuitPattern
					data.CardResult.HandCardKeys = p.cardData.HandCardKeys
					data.CardResult.PublicCardKeys = p.cardData.PublicCardKeys
					data.ResultMoney = p.resultMoney
					data.TaxRate = taxRate
					InsertAccessData(data)
				}
			} else {
				p.resultMoney -= p.totalDownBet
				var taxMoney float64
				if p.resultMoney > 0 {
					taxMoney = p.resultMoney * taxRate
					p.WinResultMoney = p.resultMoney
				}
				if p.resultMoney < 0 {
					p.LoseResultMoney = p.resultMoney
				}

				// 这里是玩家金额扣税
				p.resultMoney -= taxMoney

				if p.resultMoney > 0 {
					p.chips += p.totalDownBet
					p.chips += p.resultMoney
				}
			}
		}
	}
}

//TimerTask 游戏准备阶段定时器任务
func (r *Room) ReadyTimer() {

	r.RoomStat = RoomStatusRun

	// 广播游戏准备时间
	ready := &msg.ReadyTime_S2C{}
	ready.ReadyTime = ReadyTime
	r.Broadcast(ready)

	// 玩家补充筹码
	r.PlayerAddChips()

	go func() {
		for range r.clock.C {
			r.counter++
			//log.Debug("readyTime clock : %v ", r.counter)
			if r.counter == 2 {
				// 洗牌
				r.Cards.Shuffle()

				// 设置玩家状态
				r.SetPlayerStatus()

				// 产生庄家
				var dealer *Player
				banker := r.Banker
				r.Each(int(banker+1)%MaxPlayer, func(p *Player) bool {
					dealer = p
					r.Banker = dealer.chair
					dealer.IsButton = true
					return false
				})
				if dealer == nil {
					return
				}
				log.Debug("庄家的座位号为 :%v", dealer.chair)

				r.remain = 0
				r.allin = 0

				//Round 1：preFlop 开始发手牌,下注
				r.readyPlay()
				r.Status = msg.GameStep_PreFlop
				log.Debug("GameStep_PreFlop 阶段: %v", r.Status)
				r.Each(0, func(p *Player) bool {
					// 生成玩家手牌,获取的是对应牌型生成二进制的数
					p.cards = algorithm.Cards{r.Cards.Take(), r.Cards.Take()}
					p.cardData.HandCardKeys = p.cards.HexInt()

					kind, _ := algorithm.De(p.cards.GetType())
					p.cardData.SuitPattern = msg.CardSuit(kind)
					//log.Debug("preFlop玩家手牌和类型 ~ :%v, %v", p.cards.HexInt(), kind)

					game := &msg.GameStepChange_S2C{}
					game.RoomData = r.RespRoomData()
					p.SendMsg(game)
					return true
				})
			}
			if r.counter == 4 {
				push := &msg.PushCardTime_S2C{}
				push.RoomData = r.RespRoomData()
				r.Broadcast(push)
			}
			if r.counter >= ReadyTime {
				r.counter = 0
				ReadyTimeChan <- true
				return
			}
		}
	}()
}

//TimerTask 游戏准备阶段定时器任务
func (r *Room) ActionWaitTimer() {
	go func() {
		for range r.clock.C {
			r.counter++
			if r.counter == ActionWaitTime {
				r.counter = 0
				ActionTimeChan <- true
				return
			}
		}
	}()
}

//TimerTask 游戏开始定时器任务
func (r *Room) GameRunTask() {
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
			//log.Debug("settleTime clock : %v ", r.counter)
			if r.counter >= SettleTime {
				r.counter = 0
				// 剔除房间玩家
				r.KickPlayer()
				// 根据房间机器数量来调整机器
				//r.AdjustRobot()
				// 超时弃牌站起,这里要设置房间为等待状态,不然不能站起玩家
				r.TimeOutStandUp()

				r.RoomStat = RoomStatusOver
				r.Status = msg.GameStep_Waiting
				// 游戏阶段变更
				game := &msg.GameStepChange_S2C{}
				game.RoomData = r.RespRoomData()
				r.Broadcast(game)

				//开始新一轮游戏,重复调用StartGameRun函数
				log.Debug("RestartGame 开始运行游戏~")

				r.PiPeiHandle()



				r.StartGameRun()
				return
			}
		}
	}()
}

//RealPlayerLength 真实房间玩家人数
func (r *Room) RealPlayerLength() int32 {
	var num int32
	for _, v := range r.AllPlayer {
		if v != nil && v.IsRobot == false {
			num++
		}
	}
	log.Debug("当前房间所有玩家人数: %v", num)
	return num
}

//RealPlayerLength 真实房间玩家人数
func (r *Room) ListRealPlayerLen() int32 {
	var num int32
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == false {
			num++
		}
	}
	return num
}

//RealPlayerLength 真实房间机器人数
func (r *Room) RobotsLength() int32 {
	var num int32
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == true {
			num++
		}
	}
	log.Debug("当前房间机器人数: %v", num)
	return num
}

// 房间装载1-6机器人
func (r *Room) LoadRoomRobots() {
	// 当玩家创建新房间时,则安排随机2-4机器人
	sliceNum := []int{1, 2, 3, 4, 5, 6}
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(len(sliceNum))
	for i := 0; i < sliceNum[randNum]; i++ {
		time.Sleep(time.Millisecond)
		robot := gRobotCenter.CreateRobot()
		r.PlayerJoinRoom(robot)
	}
}

// 清除房间所有机器人
func (r *Room) ClearRoomRobots() {
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == true {
			r.ExitFromRoom(v)
		}
	}
}

func (r *Room) PiPeiHandle() {
	if r.ListRealPlayerLen() <= 3 {
		for _, v := range r.AllPlayer {
			if v != nil && v.IsRobot == false {
				data := &msg.PiPeiPlayer_S2C{}
				v.SendMsg(data)
				r.ClearPiPeiData(v)
				if v.chair == -1 {
					v.IsLeaveR = true
					r.StandUpList = append(r.StandUpList, v)
				} else {
					v.IsLeaveR = false
					r.PiPeiList = append(r.PiPeiList, v)
				}
			}
		}
	}

	go func() {
		sliceNum := []int{3, 4}
		rand.Seed(time.Now().UnixNano())
		randNum := rand.Intn(len(sliceNum))
		time.Sleep(time.Second * time.Duration(sliceNum[randNum]))
		for k, v := range r.StandUpList {
			if v != nil {
				leave := &msg.LeaveRoom_S2C{}
				leave.PlayerData = v.RespPlayerData()
				v.SendMsg(leave)
				r.StandUpList = append(r.StandUpList[:k], r.StandUpList[k+1:]...)
			}
		}
		for k, v := range r.PiPeiList {
			if v != nil {
				v.PiPeiRoom(r.cfgId)
				r.PiPeiList = append(r.PiPeiList[:k], r.PiPeiList[k+1:]...)
			}
		}
		return
	}()


	//if r.ListRealPlayerLen() <= 2 {
	//	for _, v := range r.AllPlayer {
	//		if v != nil && v.IsRobot == false {
	//			data := &msg.PiPeiPlayer_S2C{}
	//			v.SendMsg(data)
	//			if v.chair == -1 && v.IsStandUp == true {
	//				//log.Debug("玩家id,玩家座位:%v,%v,%v", v.Id, v.chair,v.IsStandUp)
	//				go func() {
	//					time.Sleep(time.Second * 3)
	//					v.PlayerExitRoom()
	//					return
	//				}()
	//			} else {
	//				r.ClearPiPeiData(v)
	//				go func() {
	//					time.Sleep(time.Second * 3)
	//					v.PiPeiRoom(r.cfgId)
	//					return
	//				}()
	//			}
	//		}
	//	}
	//}
	//if r.ListRealPlayerLen() <= 3 && r.ListRealPlayerLen() >= 6 {
	//	for _, v := range r.AllPlayer {
	//		if v != nil && v.IsRobot == false {
	//			data := &msg.PiPeiPlayer_S2C{}
	//			v.SendMsg(data)
	//			if v.chair == -1 && v.IsStandUp == true {
	//				r.ClearPiPeiData(v)
	//				go func() {
	//					time.Sleep(time.Second * 3)
	//					leave := &msg.LeaveRoom_S2C{}
	//					leave.PlayerData = v.RespPlayerData()
	//					v.SendMsg(leave)
	//					return
	//				}()
	//			} else {
	//				r.ClearPiPeiData(v)
	//				go func() {
	//					time.Sleep(time.Second * 3)
	//					hall.PlayerChangeTable(r, v)
	//					return
	//				}()
	//			}
	//		}
	//	}
	//}
}

func (r *Room) ClearPiPeiData(p *Player) {
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

	//log.Debug("匹配玩家退出房间成功！:%v", p)

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

func (p *Player) PiPeiRoom(cfgId string) {
	r := &Room{}
	r.Init(cfgId)

	hall.roomList = append(hall.roomList, r)
	hall.RoomRecord.Store(r.roomId, r)

	log.Debug("PiPeiRoom 创建新的房间:%v,当前房间数量:%v", r.roomId, len(hall.roomList))

	// 查找用户是否存在，如果存在就插入数据库
	if p.IsRobot == false {
		p.FindPlayerInfo()
	}

	hall.UserRoom[p.Id] = r.roomId

	// 玩家带入筹码
	r.TakeInRoomChips(p)

	p.chair = r.FindAbleChair()
	r.PlayerList[p.chair] = p

	// 房间总人数
	r.AllPlayer = append(r.AllPlayer, p)

	data := &msg.PiPeiData_S2C{}
	data.RoomData = r.RespRoomData()
	p.SendMsg(data)

	// 装载机器人
	r.LoadRoomRobots()
}
