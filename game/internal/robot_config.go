package internal

//RobotsCenter 机器人控制中心
type RobotsCenter struct {
	RobotsNumRoom int32              //每个房间放入机器数量
	mapRobotList  map[uint32]*Player //机器人列表
}

var gRobotCenter RobotsCenter
