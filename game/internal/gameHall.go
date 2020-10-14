package internal

import (
	"dezhoupoker/msg"
	"errors"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type GameHall struct {
	UserRecord sync.Map          // 用户记录
	RoomRecord sync.Map          // 房间记录
	roomList   []*Room           // 房间列表
	UserRoom   map[string]string // 用户房间
	PiPeiList0 []*Player         // 匹配列表
	PiPeiList1 []*Player         // 匹配列表
	PiPeiList2 []*Player         // 匹配列表
	PiPeiList3 []*Player         // 匹配列表
}

func NewHall() *GameHall {
	return &GameHall{
		UserRecord: sync.Map{},
		RoomRecord: sync.Map{},
		roomList:   make([]*Room, 0),
		UserRoom:   make(map[string]string),
		PiPeiList0: make([]*Player, 0),
		PiPeiList1: make([]*Player, 0),
		PiPeiList2: make([]*Player, 0),
		PiPeiList3: make([]*Player, 0),
	}
}

func HallInit() { // 大厅初始化增加一个房间
	for i := 0; i < 4; i++ {
		r := &Room{}
		roomCfg := strconv.Itoa(i)
		r.Init(roomCfg)
		hall.roomList = append(hall.roomList, r)
		hall.RoomRecord.Store(r.roomId, r)
		log.Debug("CreateRoom 创建新的房间:%v", r.roomId)

		robot := gRobotCenter.CreateRobot()
		r.PlayerJoinRoom(robot)
		robot.StandUpTable()

	}
}

//ReplacePlayerAgent 替换用户链接
func (hall *GameHall) ReplacePlayerAgent(Id string, agent gate.Agent) error {
	log.Debug("用户重连或顶替，正在替换agent %+v", Id)
	// tip 这里会拷贝一份数据，需要替换的是记录中的，而非拷贝数据中的，还要注意替换连接之后要把数据绑定到新连接上
	if v, ok := hall.UserRecord.Load(Id); ok {
		//ErrorResp(agent, msg.ErrorMsg_UserRemoteLogin, "异地登录")
		user := v.(*Player)
		user.ConnAgent = agent
		user.ConnAgent.SetUserData(v)
		return nil
	} else {
		return errors.New("用户不在记录中~")
	}
}

//agentExist 链接是否已经存在 (是否开销过大？后续可通过新增记录解决)
func (hall *GameHall) agentExist(a gate.Agent) bool {
	var exist bool
	hall.UserRecord.Range(func(key, value interface{}) bool {
		u := value.(*Player)
		if u.ConnAgent == a {
			exist = true
		}
		return true
	})
	return exist
}

//PlayerChangeTable 玩家进行换桌
func (hall *GameHall) PlayerChangeTable(r *Room, p *Player) {
	data := SetRoomConfig(r.cfgId)
	if p.Account < data.MinTakeIn {
		//ErrorResp(p.ConnAgent, msg.ErrorMsg_ChipsInsufficient, "玩家金币不足")
		return
	}

	// 玩家退出当前房间
	p.PlayerExitRoom()

	time.Sleep(time.Millisecond * 1500)

	// 延时5秒，重新开始游戏
	for _, room := range hall.roomList {
		if room.cfgId == r.cfgId && room.IsCanJoin() && room.roomId != r.roomId {
			if room.RealPlayerLength() <= 1 && room.RobotsLength() < 1 {
				// 装载房间机器人
				room.LoadRoomRobots()
			}
			r.PlayerJoinRoom(p)
			return
		}
	}

	hall.PlayerCreateRoom(r.cfgId, p)
	return
}

//PlayerQuickStart 快速匹配房间
func (hall *GameHall) PlayerQuickStart(cfgId string, p *Player) {
	data := SetRoomConfig(cfgId)
	if p.Account < data.MinTakeIn {
		//ErrorResp(p.ConnAgent, msg.ErrorMsg_ChipsInsufficient, "玩家金币不足")
		return
	}

	roomId := hall.UserRoom[p.Id]
	rm, _ := hall.RoomRecord.Load(roomId)
	if rm != nil {
		// 玩家如果已在游戏中，则返回房间数据
		room := rm.(*Room)
		for i, userId := range room.UserLeave {
			// 把玩家从掉线列表中移除
			if userId == p.Id {
				log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
				room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
				log.Debug("AllocateUser 清除玩家记录~:%v", userId)
				log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
				break
			}
		}
	}

	// 处理重连
	for _, r := range hall.roomList {
		for _, v := range r.PlayerList {
			if v != nil && v.Id == p.Id {
				roomData := r.RespRoomData()
				enter := &msg.EnterRoom_S2C{}
				enter.RoomData = roomData
				p.SendMsg(enter)
				return
			}
		}
	}

	if cfgId == "0" {
		if len(hall.PiPeiList0) >= 1 && len(hall.PiPeiList0) <= 3 {
			//for _, r := range hall.roomList {
			//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
			//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
			//			// 装载房间机器人
			//			r.LoadRoomRobots()
			//		}
			//		hall.DeleteWaitList(p)
			//		r.PlayerJoinRoom(p)
			//		return
			//	}
			//}
			hall.DeleteWaitList(p)
			hall.PlayerCreateRoom(cfgId, p)
			return
		} else if len(hall.PiPeiList0) >= 4 && len(hall.PiPeiList0) <= 6 {
			sliceNum := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			rand.Seed(time.Now().UnixNano())
			randNum := rand.Intn(len(sliceNum))
			if sliceNum[randNum] >= 9 {
				room := hall.CreatPiPeiRoom(cfgId)
				for _, v := range hall.PiPeiList0 {
					data := &msg.WaitPlayerList_S2C{}
					data.WaitStatus = 1
					v.SendMsg(data)
					room.PlayerJoinRoom(v)
					time.Sleep(time.Millisecond)
				}
				hall.PiPeiList0 = []*Player{}
				return
			} else {
				//for _, r := range hall.roomList {
				//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
				//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
				//			// 装载房间机器人
				//			r.LoadRoomRobots()
				//		}
				//		hall.DeleteWaitList(p)
				//		r.PlayerJoinRoom(p)
				//		return
				//	}
				//}
				hall.DeleteWaitList(p)
				hall.PlayerCreateRoom(cfgId, p)
				return
			}
		} else if len(hall.PiPeiList0) >= 7 && len(hall.PiPeiList0) <= 9 {
			room := hall.CreatPiPeiRoom(cfgId)
			for _, v := range hall.PiPeiList0 {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				v.SendMsg(data)
				room.PlayerJoinRoom(v)
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList0 = []*Player{}
			return
		} else if len(hall.PiPeiList0) >= 10 {
			room1 := hall.CreatPiPeiRoom(cfgId)
			for i := 0; i < 9; i++ {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				hall.PiPeiList0[i].SendMsg(data)
				room1.PlayerJoinRoom(hall.PiPeiList0[i])
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList0 = hall.PiPeiList0[9:]
		}
	} else if cfgId == "1" {
		if len(hall.PiPeiList1) >= 1 && len(hall.PiPeiList1) <= 3 {
			//for _, r := range hall.roomList {
			//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
			//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
			//			// 装载房间机器人
			//			r.LoadRoomRobots()
			//		}
			//		hall.DeleteWaitList(p)
			//		r.PlayerJoinRoom(p)
			//		return
			//	}
			//}
			hall.DeleteWaitList(p)
			hall.PlayerCreateRoom(cfgId, p)
			return
		} else if len(hall.PiPeiList1) >= 4 && len(hall.PiPeiList1) <= 6 {
			sliceNum := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			rand.Seed(time.Now().UnixNano())
			randNum := rand.Intn(len(sliceNum))
			if sliceNum[randNum] >= 9 {
				room := hall.CreatPiPeiRoom(cfgId)
				for _, v := range hall.PiPeiList1 {
					data := &msg.WaitPlayerList_S2C{}
					data.WaitStatus = 1
					v.SendMsg(data)
					room.PlayerJoinRoom(v)
					time.Sleep(time.Millisecond)
				}
				hall.PiPeiList1 = []*Player{}
				return
			} else {
				//for _, r := range hall.roomList {
				//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
				//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
				//			// 装载房间机器人
				//			r.LoadRoomRobots()
				//		}
				//		hall.DeleteWaitList(p)
				//		r.PlayerJoinRoom(p)
				//		return
				//	}
				//}
				hall.DeleteWaitList(p)
				hall.PlayerCreateRoom(cfgId, p)
				return
			}
		} else if len(hall.PiPeiList1) >= 7 && len(hall.PiPeiList1) <= 9 {
			room := hall.CreatPiPeiRoom(cfgId)
			for _, v := range hall.PiPeiList1 {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				v.SendMsg(data)
				room.PlayerJoinRoom(v)
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList1 = []*Player{}
			return
		} else if len(hall.PiPeiList1) >= 10 {
			room1 := hall.CreatPiPeiRoom(cfgId)
			for i := 0; i < 9; i++ {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				hall.PiPeiList1[i].SendMsg(data)
				room1.PlayerJoinRoom(hall.PiPeiList1[i])
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList1 = hall.PiPeiList1[9:]
		}
	} else if cfgId == "2" {
		if len(hall.PiPeiList2) >= 1 && len(hall.PiPeiList2) <= 3 {
			//for _, r := range hall.roomList {
			//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
			//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
			//			// 装载房间机器人
			//			r.LoadRoomRobots()
			//		}
			//		hall.DeleteWaitList(p)
			//		r.PlayerJoinRoom(p)
			//		return
			//	}
			//}
			hall.DeleteWaitList(p)
			hall.PlayerCreateRoom(cfgId, p)
			return
		} else if len(hall.PiPeiList2) >= 4 && len(hall.PiPeiList2) <= 6 {
			sliceNum := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			rand.Seed(time.Now().UnixNano())
			randNum := rand.Intn(len(sliceNum))
			if sliceNum[randNum] >= 9 {
				room := hall.CreatPiPeiRoom(cfgId)
				for _, v := range hall.PiPeiList2 {
					data := &msg.WaitPlayerList_S2C{}
					data.WaitStatus = 1
					v.SendMsg(data)
					room.PlayerJoinRoom(v)
					time.Sleep(time.Millisecond)
				}
				hall.PiPeiList2 = []*Player{}
				return
			} else {
				//for _, r := range hall.roomList {
				//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
				//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
				//			// 装载房间机器人
				//			r.LoadRoomRobots()
				//		}
				//		hall.DeleteWaitList(p)
				//		r.PlayerJoinRoom(p)
				//		return
				//	}
				//}
				hall.DeleteWaitList(p)
				hall.PlayerCreateRoom(cfgId, p)
				return
			}
		} else if len(hall.PiPeiList2) >= 7 && len(hall.PiPeiList2) <= 9 {
			room := hall.CreatPiPeiRoom(cfgId)
			for _, v := range hall.PiPeiList2 {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				v.SendMsg(data)
				room.PlayerJoinRoom(v)
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList2 = []*Player{}
			return
		} else if len(hall.PiPeiList2) >= 10 {
			room1 := hall.CreatPiPeiRoom(cfgId)
			for i := 0; i < 9; i++ {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				hall.PiPeiList2[i].SendMsg(data)
				room1.PlayerJoinRoom(hall.PiPeiList2[i])
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList2 = hall.PiPeiList2[9:]
		}
	} else if cfgId == "3" {
		if len(hall.PiPeiList3) >= 1 && len(hall.PiPeiList3) <= 3 {
			//for _, r := range hall.roomList {
			//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
			//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
			//			// 装载房间机器人
			//			r.LoadRoomRobots()
			//		}
			//		hall.DeleteWaitList(p)
			//		r.PlayerJoinRoom(p)
			//		return
			//	}
			//}
			hall.DeleteWaitList(p)
			hall.PlayerCreateRoom(cfgId, p)
			return
		} else if len(hall.PiPeiList3) >= 4 && len(hall.PiPeiList3) <= 6 {
			sliceNum := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			rand.Seed(time.Now().UnixNano())
			randNum := rand.Intn(len(sliceNum))
			if sliceNum[randNum] >= 9 {
				room := hall.CreatPiPeiRoom(cfgId)
				for _, v := range hall.PiPeiList3 {
					data := &msg.WaitPlayerList_S2C{}
					data.WaitStatus = 1
					v.SendMsg(data)
					room.PlayerJoinRoom(v)
					time.Sleep(time.Millisecond)
				}
				hall.PiPeiList3 = []*Player{}
				return
			} else {
				//for _, r := range hall.roomList {
				//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
				//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
				//			// 装载房间机器人
				//			r.LoadRoomRobots()
				//		}
				//		hall.DeleteWaitList(p)
				//		r.PlayerJoinRoom(p)
				//		return
				//	}
				//}
				hall.DeleteWaitList(p)
				hall.PlayerCreateRoom(cfgId, p)
				return
			}
		} else if len(hall.PiPeiList3) >= 7 && len(hall.PiPeiList3) <= 9 {
			room := hall.CreatPiPeiRoom(cfgId)
			for _, v := range hall.PiPeiList3 {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				v.SendMsg(data)
				room.PlayerJoinRoom(v)
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList3 = []*Player{}
			return
		} else if len(hall.PiPeiList3) >= 10 {
			room1 := hall.CreatPiPeiRoom(cfgId)
			for i := 0; i < 9; i++ {
				data := &msg.WaitPlayerList_S2C{}
				data.WaitStatus = 1
				hall.PiPeiList3[i].SendMsg(data)
				room1.PlayerJoinRoom(hall.PiPeiList3[i])
				time.Sleep(time.Millisecond)
			}
			hall.PiPeiList3 = hall.PiPeiList3[9:]
		}
	}

	//for _, r := range hall.roomList {
	//	if r.cfgId == cfgId && r.IsCanJoin() && p.PreRoomId != r.roomId {
	//		if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
	//			// 装载房间机器人
	//			r.LoadRoomRobots()
	//		}
	//		r.PlayerJoinRoom(p)
	//		return
	//	}
	//}
	//
	//hall.PlayerCreateRoom(cfgId, p)
	//return
}

//PlayerCreateRoom 创建游戏房间
func (hall *GameHall) PlayerCreateRoom(cfgId string, p *Player) {
	r := &Room{}
	r.Init(cfgId)

	hall.roomList = append(hall.roomList, r)
	hall.RoomRecord.Store(r.roomId, r)

	log.Debug("CreateRoom 创建新的房间:%v,当前房间数量:%v", r.roomId, len(hall.roomList))

	if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
		// 装载房间机器人
		r.LoadRoomRobots()
	}
	r.PlayerJoinRoom(p)

}

func (hall *GameHall) DeleteWaitList(p *Player) {
	if p.cfgId == "0" {
		for k, v := range hall.PiPeiList0 {
			if v.Id == p.Id {
				hall.PiPeiList0 = append(hall.PiPeiList0[:k], hall.PiPeiList0[k+1:]...)
			}
		}
	} else if p.cfgId == "1" {
		for k, v := range hall.PiPeiList1 {
			if v.Id == p.Id {
				hall.PiPeiList1 = append(hall.PiPeiList1[:k], hall.PiPeiList1[k+1:]...)
			}
		}
	} else if p.cfgId == "2" {
		for k, v := range hall.PiPeiList2 {
			if v.Id == p.Id {
				hall.PiPeiList2 = append(hall.PiPeiList2[:k], hall.PiPeiList2[k+1:]...)
			}
		}
	} else if p.cfgId == "3" {
		for k, v := range hall.PiPeiList3 {
			if v.Id == p.Id {
				hall.PiPeiList3 = append(hall.PiPeiList3[:k], hall.PiPeiList3[k+1:]...)
			}
		}
	}
}

func (hall *GameHall) CreatPiPeiRoom(cfgId string) *Room {
	r := &Room{}
	r.Init(cfgId)

	hall.roomList = append(hall.roomList, r)
	hall.RoomRecord.Store(r.roomId, r)
	return r
}
