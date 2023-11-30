package rdb

// redis 用到的key都放这里统一管理，命名规则  项目:游戏:表「:id1_id2_id3...」 有的key可能不需要索引
// 例： bygame:slot:user:1010082191    用户个人数据去要每个用户区分
//     bygame:center:serverusers:1_2  某个游戏的某个服的玩家 所以需要两个索引字段 gameId,服务器id
//     bygame:slot:repertory          slot的各区间库存内部使用hash存储所以这个key不需要索引

// 前缀
const (
	preFixCenter = "bygame:center"
	preFixGate   = "bygame:gate" // 负载相关的用这个
	preFixSlot   = "bygame:slot"
	preFixDomino = "bygame:domino"
	preFixFish   = "bygame:fish"
)
