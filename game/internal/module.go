package internal

import (
	"dezhoupoker/base"
	"github.com/name5566/leaf/module"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	hall = NewHall()
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	// 初始连接数据库
	InitMongoDB()
	HallInit()
}

func (m *Module) OnDestroy() {

}
