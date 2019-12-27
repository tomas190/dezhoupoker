package internal

import (
	"dezhoupoker/msg"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

type RoomData struct {
	Id        string
	SB        float64
	BB        float64
	MinTakeIn float64
	MaxTakeIn float64
}

//SetRoomConfig 房间配置限定金额
func SetRoomConfig(cfgId string) RoomData {
	roomData := []RoomData{
		{Id: "0", SB: 0.1, BB: 0.2, MinTakeIn: 10, MaxTakeIn: 200},
		{Id: "1", SB: 1, BB: 2, MinTakeIn: 50, MaxTakeIn: 1000},
		{Id: "2", SB: 5, BB: 10, MinTakeIn: 300, MaxTakeIn: 5000},
		{Id: "3", SB: 25, BB: 50, MinTakeIn: 1000, MaxTakeIn: 1000000},
	}
	for _, v := range roomData {
		if v.Id == cfgId {
			return v
		}
	}
	return RoomData{}
}

//ErrorResp 错误消息返回
func ErrorResp(a gate.Agent, err msg.ErrorMsg, data string) {
	log.Debug("<--------ErrorResp错误: %v--------->", err)
	a.WriteMsg(&msg.Error_S2C{Error: err, Data: data})
}
