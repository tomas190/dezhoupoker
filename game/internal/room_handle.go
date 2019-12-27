package internal

import (
	"dezhoupoker/msg"
	"github.com/name5566/leaf/log"
)

//PlayerJoinRoom 玩家加入房间
func (r *Room) PlayerJoinRoom(p *Player) {

	log.Debug("Player Join Game Room ~")

	// 查找用户是否存在，如果存在就插入数据库
	p.FindPlayerID()

	// 玩家带入筹码
	p.chips = r.TakeInRoomChips(p)

	p.chair = r.FindAbleChair(p.historyChair)
	r.PlayerList[p.chair] = p
	p.historyChair = p.chair
	// 房间总人数
	r.AllPlayer = append(r.AllPlayer, p)

	if r.Status == msg.GameStep_Waiting {
		// 返回房间数据
		enter := &msg.EnterRoom_S2C{}
		data := r.RespRoomData(p)
		enter.RoomData = data
		p.SendMsg(enter)

		r.StartGameRun()
	} else {
		// 如果玩家中途加入游戏，则玩家视为弃牌状态
		p.actStatus = msg.ActionStatus_FOLD
		// 返回房间数据
		enter := &msg.EnterRoom_S2C{}
		data := r.RespRoomData(p)
		enter.RoomData = data
		p.SendMsg(enter)
	}
}

//StartGameRun 游戏开始运行
func (r *Room) StartGameRun() {
	// 踢掉筹码小于大盲的玩家
	r.KickPlayer()

	// 当前房间人数存在两人及两人以上才开始游戏
	n := r.PlayerLength()
	if n < 2 {
		log.Debug("房间人数少于2人，不能开始游戏~")
		return
	}

}

//ExitFromRoom 退出房间处理
func (r *Room) ExitFromRoom(p *Player) {
	p.Account += p.chips
	// 清除用户数据
	p.ClearPlayerData()

	if p.chair != -1 {
		r.PlayerList[p.chair] = nil
	}

	for k, v := range r.AllPlayer {
		if v != nil {
			r.AllPlayer = append(r.AllPlayer[:k], r.AllPlayer[k+1:]...)
		}
	}

	delete(hall.UserRoom, p.Id)
	hall.UserRecord.Delete(p.Id)

	// 如果房间总人数为0，删除房间缓存
	if len(r.AllPlayer) == 0 {
		hall.RoomRecord.Delete(r.roomId)
		log.Debug("Room Player Number is 0，so Delete this Room~")
	}

	leave := &msg.LeaveRoom_S2C{}
	leave.PlayerInfo = new(msg.PlayerInfo)
	leave.PlayerInfo.Id = p.Id
	leave.PlayerInfo.NickName = p.NickName
	leave.PlayerInfo.HeadImg = p.HeadImg
	leave.PlayerInfo.Account = p.Account
	p.SendMsg(leave)
}
