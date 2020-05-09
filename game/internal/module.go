package internal

import (
	"dezhoupoker/base"
	"github.com/name5566/leaf/module"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	hall = NewHall()

	c4c = &Conn4Center{}
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	// 初始连接数据库
	InitMongoDB()
	HallInit()

	//机器人初始化并开始
	gRobotCenter.Init()

	c4c.Init()
	c4c.CreatConnect()
}

func (m *Module) OnDestroy() {

}
