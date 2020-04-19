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

	c4c.Init()
	c4c.CreatConnect()
}

func (m *Module) OnDestroy() {

}
