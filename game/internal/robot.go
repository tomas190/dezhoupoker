package internal

import (
	"dezhoupoker/msg"
	"fmt"
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
	//log.Debug("-------------- RobotsCenter Init~! ---------------")
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

func (p *Player) RobotDownBet(r *Room) {
	//log.Debug("机器人开始下注~")
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
			var downBet []float64
			if r.cfgId == "0" {
				downBet = []float64{0.4, 0.5, 0.6}
			}
			if r.cfgId == "1" {
				downBet = []float64{4, 5, 6}
			}
			if r.cfgId == "2" {
				downBet = []float64{20, 25, 30}
			}
			if r.cfgId == "3" {
				downBet = []float64{100, 125, 150}
			}
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
		timerSlice = []int32{3, 6, 4, 3, 5, 8, 4}
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

	go func() {
		for range r.clock.C {
			r.counter++
			if r.counter == timerSlice[num] {
				r.counter = 0
				p.action <- actionType
				return
			}
		}
	}()
}

func (r *Room) AddRobot() {
	// 机器人处理
	robotRand := []int32{0, 1, 0, 1}
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(robotRand))
	if robotRand[num] == 1 {
		robot := gRobotCenter.CreateRobot()
		r.PlayerJoinRoom(robot)
	}
}

func (r *Room) DelRobot() {
	// 机器人处理
	robotRand := []int32{0, 1, 0, 1}
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(robotRand))
	if robotRand[num] == 1 {
		for _, v := range r.PlayerList {
			if v != nil && v.IsRobot == true {
				v.PlayerExitRoom()
			}
		}
	}
}

func (r *Room) AdjustRobot() {
	if r.RobotsLength() <= 3 {
		robot := gRobotCenter.CreateRobot()
		r.PlayerJoinRoom(robot)
	} else if r.RobotsLength() >= 6 {
		for _, v := range r.PlayerList {
			if v != nil && v.IsRobot == true {
				v.PlayerExitRoom()
			}
		}
	}
}

