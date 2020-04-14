package msg

import "github.com/name5566/leaf/network/protobuf"

// 使用默认的 Json 消息处理器 (默认还提供了 ProtoBuf 消息处理器)
var Processor = protobuf.NewProcessor()

func init() {
	Processor.Register(&Ping{})                   //--0
	Processor.Register(&Pong{})                   //--1
	Processor.Register(&Error_S2C{})              //--2
	Processor.Register(&Login_C2S{})              //--3
	Processor.Register(&Login_S2C{})              //--4
	Processor.Register(&Logout_C2S{})             //--5
	Processor.Register(&Logout_S2C{})             //--6
	Processor.Register(&QuickStart_C2S{})         //--7
	Processor.Register(&ChangeTable_C2S{})        //--8
	Processor.Register(&JoinRoom_S2C{})           //--9
	Processor.Register(&EnterRoom_S2C{})          //--10
	Processor.Register(&NoticeJoin_S2C{})         //--11
	Processor.Register(&LeaveRoom_C2S{})          //--12
	Processor.Register(&LeaveRoom_S2C{})          //--13
	Processor.Register(&SitDown_C2S{})            //--14
	Processor.Register(&SitDown_S2C{})            //--15
	Processor.Register(&StandUp_C2S{})            //--16
	Processor.Register(&StandUp_S2C{})            //--17
	Processor.Register(&PlayerAction_C2S{})       //--18
	Processor.Register(&PlayerAction_S2C{})       //--19
	Processor.Register(&PlayerActionChange_S2C{}) //--20
	Processor.Register(&AddChips_C2S{})           //--21
	Processor.Register(&AddChips_S2C{})           //--22
	Processor.Register(&GameStepChange_S2C{})     //--23
	Processor.Register(&ResultGameData_S2C{})     //--24
	Processor.Register(&ReadyTime_S2C{})          //--25
	Processor.Register(&SettleTime_S2C{})         //--26
	Processor.Register(&PushCardTime_S2C{})       //--27
}
