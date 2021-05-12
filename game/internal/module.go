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

	packageTax = make(map[uint16]float64)

	// 初始连接数据库
	InitMongoDB()

	//机器人初始化并开始
	gRobotCenter.Init()
	// 大厅初始化
	HallInit()

	c4c.Init()
	c4c.CreatConnect()

	go StartHttpServer()
}

func (m *Module) OnDestroy() {

}
