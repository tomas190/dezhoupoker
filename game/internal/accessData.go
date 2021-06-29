package internal

import (
	"dezhoupoker/conf"
	"dezhoupoker/msg"
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

type GameDataReq struct {
	Id        string `form:"id" json:"id"`
	GameId    string `form:"game_id" json:"game_id"`
	RoundId   string `form:"round_id" json:"round_id"`
	RoomId    string `form:"room_id" json:"room_id"`
	CfgID     string `form:"cfg_id" json:"cfg_id"`
	StartTime int64  `form:"start_time" json:"start_time"`
	EndTime   int64  `form:"end_time" json:"end_time"`
	Skip      int    `form:"skip" json:"skip"`
	Limit     int    `form:"limit" json:"limit"`
}

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GameData struct {
	Time            int64         `json:"time"`
	TimeFmt         string        `json:"time_fmt"`
	RoundId         string        `json:"round_id"`
	RoomId          string        `json:"room_id"`
	CfgID           string        `json:"cfg_id"`
	SmallBlind      string        `json:"small_blind"`
	BigBlind        string        `json:"big_blind"`
	SmallMoney      float64       `json:"small_money"`
	BigMoney        float64       `json:"big_money"`
	PublicCard      []int32       `json:"public_card"`
	PlayerInfo      []*ResultData `json:"player_info"`
	DownBetTime     int64         `json:"down_bet_time"`
	PotMoney        float64       `json:"pot_money"`
	TaxRate         float64       `json:"tax_rate"`
	PlayerId        string        `json:"player_id"`
	SettlementFunds interface{}   `json:"settlement_funds"` // 结算信息 输赢结果
	SpareCash       interface{}   `json:"spare_cash"`       // 剩余金额
	CreatedAt       int64         `json:"created_at"`
	StartTime       int64         `json:"start_time"`
	EndTime         int64         `json:"end_time"`
}

type pageData struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

type GetSurPool struct {
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
}

type UpSurPool struct {
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
}

type PlayerInfoData struct {
	GameFlow float64 `json:"game_flow" bson:"game_flow"` // 流水
	WinNum   int64   `json:"win_num" bson:"win_num"`     // 总赢局数
	WinGold  float64 `json:"win_gold" bson:"win_gold"`   // 总赢金币
	LoseNum  int64   `json:"lose_num" bson:"lose_num"`   // 总输局数
	LoseGold float64 `json:"lose_gold" bson:"lose_gold"` // 总输金币
}

const (
	SuccCode = 0
	ErrCode  = -1
)

// HTTP端口监听
func StartHttpServer() {
	// 运营后台数据接口
	http.HandleFunc("/api/accessData", getAccessData)
	// 获取游戏数据接口
	http.HandleFunc("/api/getGameData", getAccessData)
	// 查询子游戏盈余池数据
	http.HandleFunc("/api/getSurplusOne", getSurplusOne)
	// 修改盈余池数据
	http.HandleFunc("/api/uptSurplusConf", uptSurplusOne)
	// 请求玩家退出
	http.HandleFunc("/api/reqPlayerLeave", reqPlayerLeave)
	// 获取玩家信息
	http.HandleFunc("/api/getPlayInfo", getPlayInfo)
	// 解锁玩家资金
	http.HandleFunc("/api/unLockUserMoney", unLockUserMoney)

	err := http.ListenAndServe(":"+conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

func getAccessData(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq

	req.Id = r.FormValue("id")
	req.GameId = r.FormValue("game_id")
	req.RoomId = r.FormValue("room_id")
	req.CfgID = r.FormValue("cfg_id")
	req.RoundId = r.FormValue("round_id")
	startTime := r.FormValue("start_time")
	endTime := r.FormValue("end_time")
	page := r.FormValue("page")
	limit := r.FormValue("limit")

	selector := bson.M{}

	if req.Id != "" {
		selector["result_info.player_id"] = req.Id
	}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.RoomId != "" {
		selector["room_id"] = req.RoomId
	}

	if req.CfgID != "" {
		selector["cfg_id"] = req.CfgID
	}

	if req.RoundId != "" {
		selector["round_id"] = req.RoundId
	}

	sTime, _ := strconv.Atoi(startTime)

	eTime, _ := strconv.Atoi(endTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime != 0 && eTime == 0 {
		selector["down_bet_time"] = bson.M{"$gt": sTime}
	}

	if eTime != 0 && sTime == 0 {
		selector["down_bet_time"] = bson.M{"$lt": eTime}
	}

	pages, _ := strconv.Atoi(page)

	limits, _ := strconv.Atoi(limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}

	recodes, count, err := GetDownRecodeList(pages, limits, selector, "-down_bet_time")
	if err != nil {
		return
	}

	var gameData []GameData
	for i := 0; i < len(recodes); i++ {
		var gd GameData
		pr := recodes[i]
		gd.Time = pr.DownBetTime
		gd.TimeFmt = FormatTime(pr.DownBetTime, "2006-01-02 15:04:05")
		gd.RoundId = pr.RoundId
		gd.RoomId = pr.RoomId
		gd.CfgID = pr.CfgID
		gd.SmallBlind = pr.SmallBlind
		gd.BigBlind = pr.BigBlind
		gd.SmallMoney = pr.SmallMoney
		gd.BigMoney = pr.BigMoney
		gd.PublicCard = pr.PublicCard
		gd.PlayerInfo = pr.ResultInfo
		gd.DownBetTime = pr.DownBetTime
		gd.PotMoney = pr.PotMoney
		gd.TaxRate = pr.TaxRate
		gd.PlayerId = pr.Id
		gd.SettlementFunds = pr.SettlementFunds
		gd.SpareCash = pr.SpareCash
		gd.CreatedAt = pr.DownBetTime
		gd.StartTime = pr.StartTime
		gd.EndTime = pr.EndTime
		gameData = append(gameData, gd)
	}

	var result pageData
	result.Total = count
	result.List = gameData

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func FormatTime(timeUnix int64, layout string) string {
	if timeUnix == 0 {
		return ""
	}
	format := time.Unix(timeUnix, 0).Format(layout)
	return format
}

func NewResp(code int, msg string, data interface{}) ApiResp {
	return ApiResp{Code: code, Msg: msg, Data: data}
}

// 查询子游戏盈余池数据
func getSurplusOne(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq
	req.GameId = r.FormValue("game_id")
	log.Debug("game_id :%v", req.GameId)

	selector := bson.M{}
	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	result, err := GetSurPoolData(selector)
	if err != nil {
		return
	}

	var getSur GetSurPool
	getSur.PlayerTotalLose = result.PlayerTotalLose
	getSur.PlayerTotalWin = result.PlayerTotalWin
	getSur.PercentageToTotalWin = result.PercentageToTotalWin
	getSur.TotalPlayer = result.TotalPlayer
	getSur.CoefficientToTotalPlayer = result.CoefficientToTotalPlayer
	getSur.FinalPercentage = result.FinalPercentage
	getSur.PlayerTotalLoseWin = result.PlayerTotalLoseWin
	getSur.SurplusPool = result.SurplusPool
	getSur.PlayerLoseRateAfterSurplusPool = result.PlayerLoseRateAfterSurplusPool
	getSur.DataCorrection = result.DataCorrection

	js, err := json.Marshal(NewResp(SuccCode, "", getSur))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func uptSurplusOne(w http.ResponseWriter, r *http.Request) {

	rateSur := r.PostFormValue("player_lose_rate_after_surplus_pool")
	percentage := r.PostFormValue("percentage_to_total_win")
	coefficient := r.PostFormValue("coefficient_to_total_player")
	final := r.PostFormValue("final_percentage")
	correction := r.PostFormValue("data_correction")

	var req GameDataReq
	req.GameId = r.FormValue("game_id")
	log.Debug("game_id :%v", req.GameId)

	selector := bson.M{}
	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}
	sur, err := GetSurPoolData(selector)
	if err != nil {
		return
	}

	var upt UpSurPool
	upt.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
	upt.PercentageToTotalWin = sur.PercentageToTotalWin
	upt.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
	upt.FinalPercentage = sur.FinalPercentage
	upt.DataCorrection = sur.DataCorrection

	if rateSur != "" {
		upt.PlayerLoseRateAfterSurplusPool, _ = strconv.ParseFloat(rateSur, 64)
		sur.PlayerLoseRateAfterSurplusPool = upt.PlayerLoseRateAfterSurplusPool
	}
	if percentage != "" {
		upt.PercentageToTotalWin, _ = strconv.ParseFloat(percentage, 64)
		sur.PercentageToTotalWin = upt.PercentageToTotalWin
	}
	if coefficient != "" {
		data, _ := strconv.ParseInt(coefficient, 10, 32)
		upt.CoefficientToTotalPlayer = int32(data)
		sur.CoefficientToTotalPlayer = upt.CoefficientToTotalPlayer
	}
	if final != "" {
		upt.FinalPercentage, _ = strconv.ParseFloat(final, 64)
		sur.FinalPercentage = upt.FinalPercentage
	}
	if correction != "" {
		upt.DataCorrection, _ = strconv.ParseFloat(correction, 64)
		sur.DataCorrection = upt.DataCorrection
	}

	sur.SurplusPool = Decimal((sur.PlayerTotalLose - (sur.PlayerTotalWin * sur.PercentageToTotalWin) - float64(sur.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage)
	// 更新盈余池数据
	UpdateSurPool(&sur)

	js, err := json.Marshal(NewResp(SuccCode, "", upt))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func reqPlayerLeave(w http.ResponseWriter, r *http.Request) {
	Id := r.FormValue("id")
	user, _ := hall.UserRecord.Load(Id)
	if user != nil {
		u := user.(*Player)
		u.gameStep = emNotGaming
		u.totalDownBet = 0
		u.PlayerExitRoom()
		js, err := json.Marshal(NewResp(SuccCode, "", "已成功T出房间!"))
		if err != nil {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
			return
		}
		w.Write(js)
	}
}

func getPlayInfo(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	startTime := r.FormValue("start_time")
	endTime := r.FormValue("end_time")
	roomType := r.FormValue("room_type")
	log.Debug("id为:%v", id)
	log.Debug("roomType为:%v", roomType)

	selector := bson.M{}

	if id != "" {
		selector["id"] = id
	}

	if roomType != "" && roomType != "0" {
		selector["room_type"] = roomType
	}

	sTime, _ := strconv.Atoi(startTime)

	eTime, _ := strconv.Atoi(endTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime != 0 && eTime == 0 {
		selector["down_bet_time"] = bson.M{"$gt": sTime}
	}

	if eTime != 0 && sTime == 0 {
		selector["down_bet_time"] = bson.M{"$lt": eTime}
	}

	recodes, count, err := GetPlayerGameData(selector, "-down_bet_time")
	if err != nil {
		return
	}
	log.Debug("当前获取的数量为:%v", count)

	var playerInfo PlayerInfoData
	for i := 0; i < len(recodes); i++ {
		pr := recodes[i]
		if pr.ResultMoney > 0 {
			playerInfo.GameFlow += pr.ResultMoney
			playerInfo.WinNum += 1
			playerInfo.WinGold += pr.ResultMoney
		} else if pr.ResultMoney < 0 {
			playerInfo.GameFlow -= pr.ResultMoney
			playerInfo.LoseNum += 1
			playerInfo.LoseGold -= pr.ResultMoney
		}
	}
	playerInfo.GameFlow = Decimal(playerInfo.GameFlow)
	playerInfo.WinGold = Decimal(playerInfo.WinGold)
	playerInfo.LoseGold = Decimal(playerInfo.LoseGold)

	js, err := json.Marshal(NewResp(SuccCode, "ok", playerInfo))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "ok", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func unLockUserMoney(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	lockMoney := r.FormValue("lock_money")

	user, _ := hall.UserRecord.Load(id)
	if user != nil {
		u := user.(*Player)
		money, _ := strconv.ParseFloat(lockMoney, 64)
		u.LockMoney = money
		//c4c.UnlockSettlement(u, 0)

		time.Sleep(time.Second)

		c4c.UserLogoutCenter(u.Id, u.Password, u.Token)
		u.IsOnline = false
		hall.UserRecord.Delete(u.Id)
		leaveHall := &msg.Logout_S2C{}
		u.SendMsg(leaveHall)
		u.ConnAgent.Close()

		js, err := json.Marshal(NewResp(SuccCode, "ok", "解锁成功！"))
		if err != nil {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "ok", Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
