package gate

import (
	"dezhoupoker/game"
	"dezhoupoker/msg"
)

func init() {
	msg.Processor.SetRouter(&msg.Ping{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.Login_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Logout_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.QuickStart_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.ChangeTable_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.LeaveRoom_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.SitDown_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.StandUp_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.PlayerAction_C2S{}, game.ChanRPC)
}
