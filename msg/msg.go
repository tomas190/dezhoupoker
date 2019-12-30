package msg

import "github.com/name5566/leaf/network/protobuf"

// 使用默认的 Json 消息处理器 (默认还提供了 ProtoBuf 消息处理器)
var Processor = protobuf.NewProcessor()

func init() {
	Processor.Register(&Ping{})               //--0
	Processor.Register(&Pong{})               //--1
	Processor.Register(&Error_S2C{})          //--2
	Processor.Register(&Login_C2S{})          //--3
	Processor.Register(&Login_S2C{})          //--4
	Processor.Register(&Logout_C2S{})         //--5
	Processor.Register(&Logout_S2C{})         //--6
	Processor.Register(&QuickStart_C2S{})     //--7
	Processor.Register(&ChangeTable_C2S{})    //--8
	Processor.Register(&EnterRoom_S2C{})      //--9
	Processor.Register(&LeaveRoom_C2S{})      //--10
	Processor.Register(&LeaveRoom_S2C{})      //--11
	Processor.Register(&SitDown_C2S{})        //--12
	Processor.Register(&SitDown_S2C{})        //--13
	Processor.Register(&StandUp_C2S{})        //--14
	Processor.Register(&StandUp_S2C{})        //--15
	Processor.Register(&PlayerAction_C2S{})   //--16
	Processor.Register(&PlayerAction_S2C{})   //--17
	Processor.Register(&GameStepChange_S2C{}) //--18
	Processor.Register(&ResultGameData_S2C{}) //--19
}
