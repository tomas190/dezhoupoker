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
	dbName          = "dezhoupoker"
	playerInfo      = "playerInfo"
	roomSettle      = "roomSettle"
	settleWinMoney  = "settleWinMoney"
	settleLoseMoney = "settleLoseMoney"
	accessDB        = "accessData"
	surPlusDB       = "surPlusDB"
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

func (p *Player) InsertPlayerID() error {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	err := c.Insert(p.Id)
	return err
}

func (p *Player) FindPlayerID() {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	err := c.Find(bson.M{"id": p.Id})
	if err != nil {
		err2 := p.InsertPlayerID()
		if err2 != nil {
			log.Error("<----- 数据库用户ID数据失败 ~ ----->:%v", err)
			return
		}
		SurplusPool -= 6
		log.Debug("<----- 数据库用户ID数据成功 ~ ----->")
	}
}

func (r *Room) InsertRoomData() error{
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

//InsertSurplusPool 插入盈余池数据
func InsertSurplusPool(sur *SurplusPoolDB) {
	s, c := connect(dbName, surPlusDB)
	defer s.Close()

	log.Debug("surplusPoolDB 数据: %v", sur)
	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurplusPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurplusPool数据成功 ~ ----->")
}

// 玩家的记录
type PlayerDownBetRecode struct {
	Id          string           `json:"id" bson:"id"`                       // 玩家Id
	RandId      string           `json:"rand_id" bson:"rand_id"`             // 随机Id
	RoomId      string           `json:"room_id" bson:"room_id"`             // 所在房间
	DownBetInfo float64          `json:"down_bet_info" bson:"down_bet_info"` // 玩家各注池下注的金额
	DownBetTime int64            `json:"down_bet_time" bson:"down_bet_time"` // 下注时间
	CardResult  msg.CardSuitData `json:"card_result" bson:"card_result"`     // 当局开牌结果
	ResultMoney float64          `json:"result_money" bson:"result_money"`   // 当局输赢结果(税后)
	TaxRate     float64          `json:"tax_rate" bson:"tax_rate"`           // 税率
}

//InsertAccessData 插入运营数据接入
func InsertAccessData(data *PlayerDownBetRecode) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	log.Debug("AccessData 数据: %v", data)
	err := c.Insert(data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 运营接入数据插入成功 ~ ----->")
}

//GetDownRecodeList 获取运营数据接入
func GetDownRecodeList(skip, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	var wts []PlayerDownBetRecode

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	err = c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}
