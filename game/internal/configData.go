package internal

import (
	"dezhoupoker/msg"
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"math/rand"
	"strconv"
	"time"
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
		{Id: "9", SB: 0, BB: 0, MinTakeIn: 0, MaxTakeIn: 0},
	}
	for _, v := range roomData {
		if v.Id == cfgId {
			return v
		}
	}
	return RoomData{}
}

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", value), 64)
	return value
}

func RandInRange(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(1 * time.Nanosecond)
	return rand.Intn(max-min) + min
}

//ErrorResp 错误消息返回
func ErrorResp(a gate.Agent, err msg.ErrorMsg, data string) {
	log.Debug("<--------ErrorResp错误: %v--------->", err)
	a.WriteMsg(&msg.Error_S2C{Error: err, Data: data})
}
