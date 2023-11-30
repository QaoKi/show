package data

import "bygame/common/proto/slot_proto"

type SlotData struct {
	MiniGame     []*MiniGame           // 小游戏会嵌套，所以要保存一个栈
	CtrMap       map[string]*PersonCtr // 个人杀控
	Round        int                   // 当前回合数
	ActiveCtr    int                   // 当前回合生效的杀控种类
	Ante         int64                 // 投注 //记得传值
	BetId        string                // 唯一id
	Win          int64                 // 总输赢
	PayModel     slot_proto.Stage      // 默认值是0
	PayModelAnte int64                 // 购买的底注

	MachineId   int // 机台Id，默认值为 当前用户数 + 100 (uid 中间的值)
	NotWinRound int // 未中大奖的轮次
}

type MiniGame struct {
	BetId     string           // 获得小游戏的那一次投注的id
	Buy       bool             // 是否是购买的小游戏
	Stage     slot_proto.Stage // 小游戏
	FromStage slot_proto.Stage // 从哪个状态转过来
	RoundL    int32            // 剩余回合数
	// TotalBonus int64            // 获得的金额
	TotalOdds int32 // 总回报

	// 钻石夺宝
	Result      [][]int32           // 结果
	Diamonds    []*slot_proto.Point // 钻石
	AddDiamonds []*slot_proto.Point // 待添加的钻石
}

// 个人杀控数据，每组杀控数据成对出现并且指定下一轮的初始状态
// 杀分因为和投注有关，所以不太容易杀超，送分可能会出现大奖容易送超
type PersonCtr struct {
	N          int              // 杀控基数，用来确认当前放水倍率，放水倍率低于30的时候不能出现大奖，固定采用layout7滚轮
	StartRound int              // 杀控开始回合
	Info       []*PersonCtrInfo // 一组杀控控制数据，只有第一个会生效
}

type PersonCtrInfo struct {
	Hard   bool
	Target int64
}

// 游戏记录 一次投注记一次，免费游戏合并到获得免费游戏的那一次投注上

type SlotRecord struct {
	Mid      string `json:"mid"`      // 玩家mid
	BetId    string `json:"BetId"`    // 投注id
	Vid      string `json:"vid"`      // 年月日时分秒毫秒时间戳组成的用来前端显示
	Ante     int64  `json:"ante"`     // 投注值
	Bonus    int64  `json:"bonus"`    // 赢钱
	Time     int64  `json:"time"`     // 时间戳
	MiniGame int    `json:"miniGame"` // 获得的小游戏类型
	TotalWin int64  `json:"totalWin"` // 总输赢
	Ctr      int    `json:"-"`        // 生效的杀控状态
}

// 游戏订单表

type SlotOrder struct {
	Mid string `json:"mid"` // 玩家mid
	//GameId  string `json:"gameId"`  // 游戏id
	BetId   string `json:"betId"`   // 投注id
	OrderId string `json:"orderId"` // 订单id
	Coin    int64  `json:"coin"`    // 交易金额
	Time    int64  `json:"time"`    // 时间戳
	Status  int    `json:"status"`  // 订单状态
}
