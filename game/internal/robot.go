package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"time"
)

//机器人问题:
//1、机器人没钱怎么充值,不能再房间就直接充值,不然可以被其他用户看见
//2、机器人怎么下注，如果在桌面6个位置上，是否设置机器的下注速度和选择注池
//3、机器人选择注池的输赢,都要进行计算，只是不和盈余池牵扯，主要是前端做展示
//4、如果机器人金额如果小于50或不能参加游戏,则踢出房间删除机器人，在生成新的机器人加入该房间。

//机器人下标
var RobotIndex uint32

//Init 初始机器人控制中心
func (rc *RobotsCenter) Init() {
	log.Debug("-------------- RobotsCenter Init~! ---------------")
	rc.mapRobotList = make(map[uint32]*Player)
}

//CreateRobot 创建一个机器人
func (rc *RobotsCenter) CreateRobot() *Player {
	r := &Player{}
	r.Init()

	r.IsRobot = true
	//生成随机ID
	r.Id = RandomID()
	//生成随机头像IMG
	r.HeadImg = RandomIMG()
	//生成随机机器人NickName
	r.NickName = RandomName()
	//生成机器人金币随机数
	rand.Intn(int(time.Now().Unix()))
	//money := rand.Intn(6000) + 1000
	money := rand.Intn(1182) + 523
	r.Account = float64(money)

	RobotIndex++
	return r
}

//生成随机机器人ID
func RandomID() string {
	RobotId := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
	return RobotId
}

//生成随机机器人头像IMG
func RandomIMG() string {
	//slice := []string{
	//	"1.png", "2.png", "3.png", "4.png", "5.png", "6.png", "7.png", "8.png", "9.png", "10.png",
	//	"11.png", "12.png", "13.png", "14.png", "15.png", "16.png", "17.png", "18.png", "19.png", "20.png",
	//}
	slice := []string{
		"1.png", "2.png", "3.png", "4.png", "5.png", "6.png", "7.png", "8.png", "9.png",
	}
	rand.Seed(int64(time.Now().UnixNano()))
	num := rand.Intn(len(slice))

	return slice[num]
}

//生成随机机器人NickName
func RandomName() string {
	randNum := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))
	RobotName := randNum
	return RobotName
}
