package internal

import (
	"dezhoupoker/msg"
	"github.com/name5566/leaf/gate"
	"reflect"
)

func init() {
	handlerReg(&msg.Ping{}, handlePing)

	handlerReg(&msg.Login_C2S{}, handleLogin)
	handlerReg(&msg.Logout_C2S{}, handleLogout)

	handlerReg(&msg.QuickStart_C2S{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom_C2S{}, handleJoinRoom)

	handlerReg(&msg.SitDown_C2S{}, handleRoomEvent)
	handlerReg(&msg.StandUp_C2S{}, handleRoomEvent)
	handlerReg(&msg.PlayerAction_C2S{}, handleRoomEvent)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	a := args[1].(gate.Agent)

}