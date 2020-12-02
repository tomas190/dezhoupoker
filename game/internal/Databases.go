package internal

import (
	"C"
	"dezhoupoker/conf"
	"dezhoupoker/msg"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	session *mgo.Session
)

const (
	dbName          = "dezhoupoker-Game"
	playerInfo      = "playerInfo"
	roomSettle      = "roomSettle"
	settleWinMoney  = "settleWinMoney"
	settleLoseMoney = "settleLoseMoney"
	accessDB        = "accessData"
	surPlusDB       = "surPlusDB"
	surPool         = "surplus-pool"
	playerGameData  = "playerGameData"
)

// 连接数据库集合的函数 传入集合 默认连接IM数据库
func InitMongoDB() {
	// 此处连接正式线上数据库  下面是模拟的直接连接
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{conf.Server.MongoDBAddr},
		Timeout:  60 * time.Second,
		Database: conf.Server.MongoDBAuth,
		Username: conf.Server.MongoDBUser,
		Password: conf.Server.MongoDBPwd,
	}

	var err error
	session, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatal("Connect DataBase 数据库连接ERROR: %v ", err)
	}
	log.Debug("Connect DataBase 数据库连接SUCCESS~")

	//打开数据库
	session.SetMode(mgo.Monotonic, true)
}

func connect(dbName, cName string) (*mgo.Session, *mgo.Collection) {
	s := session.Copy()
	c := s.DB(dbName).C(cName)
	return s, c
}

func (p *Player) FindPlayerInfo() {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	player := &msg.PlayerInfo{}
	player.Id = p.Id
	player.NickName = p.NickName
	player.HeadImg = p.HeadImg
	player.Account = p.Account

	err := c.Find(bson.M{"id": player.Id}).One(player)
	if err != nil {
		err2 := InsertPlayerInfo(player)
		if err2 != nil {
			log.Error("<----- 插入用户信息数据失败 ~ ----->:%v", err)
			return
		}
		log.Debug("<----- 插入用户信息数据成功 ~ ----->")
	}
}

func InsertPlayerInfo(player *msg.PlayerInfo) error {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	err := c.Insert(player)
	return err
}

//LoadPlayerCount 获取玩家数量
func LoadPlayerCount() int32 {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	n, err := c.Find(nil).Count()
	if err != nil {
		log.Debug("not Found Player Count, Maybe don't have Player")
		return 0
	}
	return int32(n)
}

func (r *Room) InsertRoomData() error {
	s, c := connect(dbName, roomSettle)
	defer s.Close()

	err := c.Insert(r)
	return err
}

//InsertWinMoney 插入房间数据
func InsertWinMoney(base interface{}) {
	s, c := connect(dbName, settleWinMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 赢钱结算数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 赢钱结算数据插入成功 ~ ----->")

}

//InsertLoseMoney 插入房间数据
func InsertLoseMoney(base interface{}) {
	s, c := connect(dbName, settleLoseMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 输钱结算数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 输钱结算数据插入成功 ~ ----->")
}

func FindSurplusPool() *SurplusPoolDB {
	s, c := connect(dbName, surPlusDB)
	defer s.Close()

	//c.RemoveAll(nil) // todo

	sur := &SurplusPoolDB{}
	err := c.Find(nil).Sort("-updatetime").One(sur)
	if err != nil {
		log.Error("<----- 查找SurplusPool数据失败 ~ ----->:%v", err)
		return nil
	}

	return sur
}

//InsertSurplusPool 插入盈余池数据
func InsertSurplusPool(sur *SurplusPoolDB) {
	s, c := connect(dbName, surPlusDB)
	defer s.Close()

	sur.PoolMoney = (sur.HistoryLose - (sur.HistoryWin * 1)) * 0.5
	log.Debug("surplusPoolDB 数据: %v", sur)
	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurplusPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurplusPool数据成功 ~ ----->")

	SurPool := &SurPool{}
	SurPool.GameId = conf.Server.GameID
	SurPool.SurplusPool = Decimal(sur.PoolMoney)
	SurPool.PlayerTotalLoseWin = Decimal(sur.HistoryLose - sur.HistoryWin)
	SurPool.PlayerTotalLose = Decimal(sur.HistoryLose)
	SurPool.PlayerTotalWin = Decimal(sur.HistoryWin)
	SurPool.TotalPlayer = sur.PlayerNum
	SurPool.FinalPercentage = 0.5
	SurPool.PercentageToTotalWin = 1
	SurPool.CoefficientToTotalPlayer = sur.PlayerNum * 0
	SurPool.PlayerLoseRateAfterSurplusPool = 0.7
	SurPool.DataCorrection = 0
	FindSurPool(SurPool)
}

func FindSurPool(SurP *SurPool) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	//c.RemoveAll(nil) // todo

	sur := &SurPool{}
	err := c.Find(nil).One(sur)
	if err != nil {
		InsertSurPool(SurP)
	} else {
		SurP.SurplusPool = (SurP.PlayerTotalLose - (SurP.PlayerTotalWin * sur.PercentageToTotalWin) - float64(SurP.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage
		SurP.FinalPercentage = sur.FinalPercentage
		SurP.PercentageToTotalWin = sur.PercentageToTotalWin
		SurP.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
		SurP.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
		SurP.DataCorrection = sur.DataCorrection
		UpdateSurPool(SurP)
	}
}

func GetSurPlus() float64 {
	s, c := connect(dbName, surPool)
	defer s.Close()

	//c.RemoveAll(nil) // todo

	sur := &SurPool{}
	err := c.Find(nil).One(sur)
	if err != nil {
		log.Debug("获取GetSurP数据失败:%v", err)
		return 0
	}
	return sur.SurplusPool
}

//插入盈余池统一字段
func InsertSurPool(sur *SurPool) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	log.Debug("SurPool 数据: %v", sur)

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurPool数据成功 ~ ----->")
}

func UpdateSurPool(sur *SurPool) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	err := c.Update(bson.M{}, sur)
	if err != nil {
		log.Error("<----- 更新 SurPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 更新SurPool数据成功 ~ ----->")
}

// 玩家的记录
type PlayerDownBetRecode struct {
	GameId          string        `json:"game_id" bson:"game_id"`                   // gameId
	RoundId         string        `json:"round_id" bson:"round_id"`                 // 随机Id
	Id              string        `json:"id" bson:"id"`                             // 玩家Id
	RoomId          string        `json:"room_id" bson:"room_id"`                   // 所在房间
	CfgID           string        `json:"cfg_id" bson:"cfg_id"`                     // 房间类型
	SmallBlind      string        `json:"small_blind" bson:"small_blind"`           // 小盲注Id
	BigBlind        string        `json:"big_blind" bson:"big_blind"`               // 大盲注Id
	SmallMoney      float64       `json:"small_money" bson:"small_money"`           // 小盲注金额
	BigMoney        float64       `json:"big_money" bson:"big_money"`               // 大盲注金额
	PublicCard      []int32       `json:"public_card" bson:"public_card"`           // 桌面公牌
	ResultInfo      []*ResultData `json:"result_info" bson:"result_info"`           // 玩家结算信息
	DownBetTime     int64         `json:"down_bet_time" bson:"down_bet_time"`       // 下注时间
	PotMoney        float64       `json:"pot_money" bson:"pot_money"`               // 当局房间底池金额
	TaxRate         float64       `json:"tax_rate" bson:"tax_rate"`                 // 税率
	SettlementFunds float64       `json:"settlement_funds" bson:"settlement_funds"` // 结算信息
	SpareCash       float64       `json:"spare_cash" bson:"spare_cash"`             // 剩余金额
	StartTime       int64         `json:"start_time" bson:"start_time"`             // 开始时间
	EndTime         int64         `json:"end_time" bson:"end_time"`                 // 结束时间
}

type ResultData struct {
	PlayerId        string  `json:"player_id" bson:"player_id"`               // 玩家ID
	Chair           int32   `json:"chair" bson:"chair"`                       // 玩家座位
	HandCard        []int32 `json:"hand_card" bson:"hand_card"`               // 玩家手牌
	DownBet         float64 `json:"down_bet" bson:"down_bet"`                 // 下注金币
	SettlementFunds float64 `json:"settlement_funds" bson:"settlement_funds"` // 结算金币(未税)
	SpareCash       float64 `json:"spare_cash" bson:"spare_cash"`             // 剩余金额
	IsRobot         bool    `json:"is_robot" bson:"is_robot"`                 // 是否机器人
}

//InsertAccessData 插入运营数据接入
func InsertAccessData(data *PlayerDownBetRecode) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	//log.Debug("AccessData 数据: %v", data)
	err := c.Insert(data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->: %v", err)
		return
	}
	log.Debug("<----- 运营接入数据插入成功 ~ ----->")
}

//GetDownRecodeList 获取运营数据接入
func GetDownRecodeList(page, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	var wts []PlayerDownBetRecode

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	log.Debug("获取 %v 条数据,limit:%v", n, limit)
	skip := (page - 1) * limit
	err = c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}

type SurPool struct {
	GameId                         string  `json:"game_id" bson:"game_id"`
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
}

//GetDownRecodeList 获取盈余池数据
func GetSurPoolData(selector bson.M) (SurPool, error) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	var wts SurPool

	err := c.Find(selector).One(&wts)
	if err != nil {
		return wts, err
	}
	return wts, nil
}

type PlayerGameData struct {
	Id          string  `json:"id" bson:"id"`                       // 玩家ID
	RoomType    string  `json:"room_type" bson:"room_type"`         // 房间类型
	ResultMoney float64 `json:"result_money" bson:"result_money"`   // 税前结算
	DownBetTime int64   `json:"down_bet_time" bson:"down_bet_time"` // 下注时间
}

//InsertPlayerInfoData 插入玩家信息数据
func InsertPlayerGameData(data *PlayerGameData) {
	s, c := connect(dbName, playerGameData)
	defer s.Close()

	err := c.Insert(data)
	if err != nil {
		log.Error("<----- 玩家信息数据插入失败 ~ ----->: %v", err)
		return
	}
	log.Debug("<----- 玩家信息数据插入成功 ~ ----->")
}

//GetPlayerInfoData 获取玩家信息
func GetPlayerGameData(selector bson.M, sortBy string) ([]PlayerGameData, int, error) {
	s, c := connect(dbName, playerGameData)
	defer s.Close()

	var wts []PlayerGameData

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	err = c.Find(selector).Sort(sortBy).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}
