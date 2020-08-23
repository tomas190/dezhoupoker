package internal

import (
	"dezhoupoker/msg"
	"errors"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"strconv"
	"sync"
	"time"
)

type GameHall struct {
	UserRecord  sync.Map          // 用户记录
	RoomRecord  sync.Map          // 房间记录
	roomList    []*Room           // 房间列表
	UserRoom    map[string]string // 用户房间
}

func NewHall() *GameHall {
	return &GameHall{
		UserRecord:  sync.Map{},
		RoomRecord:  sync.Map{},
		roomList:    make([]*Room, 0),
		UserRoom:    make(map[string]string),
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

		//r.LoadRoomRobots()

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

	// 延时5秒，重新开始游戏
	for _, room := range hall.roomList {
		if room.cfgId == r.cfgId && room.IsCanJoin() && room.roomId != r.roomId {
			room.PlayerJoinRoom(p)
			time.Sleep(time.Millisecond * 2500)
			if room.RealPlayerLength() <= 1 && room.RobotsLength() < 1 {
				// 装载房间机器人
				room.LoadRoomRobots()
			}

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

	for _, r := range hall.roomList {
		if r.cfgId == cfgId && r.IsCanJoin() {
			r.PlayerJoinRoom(p)
			time.Sleep(time.Millisecond * 2500)
			if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
				// 装载房间机器人
				r.LoadRoomRobots()
			}

			return
		}
	}

	hall.PlayerCreateRoom(cfgId, p)
	return
}

//PlayerCreateRoom 创建游戏房间
func (hall *GameHall) PlayerCreateRoom(cfgId string, p *Player) {
	r := &Room{}
	r.Init(cfgId)

	hall.roomList = append(hall.roomList, r)
	hall.RoomRecord.Store(r.roomId, r)

	log.Debug("CreateRoom 创建新的房间:%v,当前房间数量:%v", r.roomId, len(hall.roomList))
	r.PlayerJoinRoom(p)
	time.Sleep(time.Millisecond * 2500)

	if r.RealPlayerLength() <= 1 && r.RobotsLength() < 1 {
		// 装载房间机器人
		r.LoadRoomRobots()
	}

}
