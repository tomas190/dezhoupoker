package internal

import (
	"time"
)

//盈余池数据存入数据库
type SurplusPoolDB struct {
	UpdateTime     time.Time
	TimeNow        string  //记录时间（分为时间戳/字符串显示）
	Rid            string  //房间ID
	TotalWinMoney  float64 //玩家当局总赢
	TotalLoseMoney float64 //玩家当局总输
	PoolMoney      float64 //盈余池
	HistoryWin     float64 //玩家历史总赢
	HistoryLose    float64 //玩家历史总输
	PlayerNum      int32   //历史玩家人数
}

const (
	taxRate    float64 = 0.05 // 税率
	SurplusTax float64 = 0.2  // 指定盈余池的百分随机数
)

//盈余池
var SurplusPool float64 = 0

func SetPackageTaxM(packageT uint16, tax uint8) {
	packageTax[packageT] = tax
}
