package data

import (
	"bygame/common/conf"
	"bygame/common/log"
	"bygame/common/mdb"
	"bygame/common/proto/common_proto"
	"bygame/common/utils"
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type GameData struct {
	Mid      string
	FishData FishData
	SlotData SlotData
}

type User struct {
	Bid          bson.ObjectId `json:"-" bson:"_id"` // 数据库id
	Mid          string        `json:"mid"`          // bid 对应的字符串 所有逻辑操作使用,token中会解析出来
	RegisterTime int           `json:"registerTime"` // 注册时间
	LastTime     int           `json:"LastTime"`     // 最后登录时间
	LoginDay     int           `json:"loginDay"`     // 登录天数
	UserInfo     UserInfo      `json:"userInfo"`     // 用户基础信息
}

// 和 common_proto 中是相同的
type UserInfo struct {
	Mid        string  `json:"mid"`      // mid
	Uid        string  `json:"uid"`      // 玩家uid
	NickName   string  `json:"nickName"` // 玩家昵称
	Coin       int64   `json:"coin"`     // 玩家持有金币
	Avatar     string  `json:"avatar"`   // 玩家头像
	Gender     int32   `json:"gender"`   // 0 未知 1 男 2 女
	Exp        int32   `json:"exp"`      // 经验值
	CoinOffset int64   `json:"-"`        // 金币变动
	Account    Account `json:"-"`        // 账户相关的
	Index      int     `json:"-"`        // 注册序号
}

type Account struct {
	Platform string // 平台
	Code     string // 各个平台的唯一编码 游客:deviceId  之艺: uid  pop uid
	Mid      string // 用户mid

	// 之艺
	Sign string

	// pop
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
}

func (userInfo *UserInfo) UserInfo2ProtoUserInfo() *common_proto.UserInfo {
	var pui common_proto.UserInfo
	pui.Mid = userInfo.Mid
	pui.Uid = userInfo.Uid
	pui.NickName = userInfo.NickName
	pui.Coin = userInfo.Coin
	pui.Avatar = userInfo.Avatar
	pui.Gender = userInfo.Gender
	pui.Exp = userInfo.Exp
	return &pui
}

// 加金币
func (userInfo *UserInfo) AddCoin(coin int64) (ok bool) {
	now := time.Now().Unix()
	if userInfo.Account.Platform == "104" {
		// token 可能会过期
		if userInfo.Account.ExpiresIn+600 < int(now) {
			// 刷新
			refreshToken(userInfo)
		}
		err := PopAddCoin(coin, userInfo.Account.AccessToken)
		log.Inf("pop 用户更新金币 mid: %v uid: %v coin: %v err:%v", userInfo.Mid, userInfo.Uid, coin, err)
		if err == nil {
			userInfo.Coin += coin
		} else {
			userInfo.Coin = 0
		}
		return err == nil
	}

	db := mdb.GetMdb()
	err := db.C(mdb.DB_USER).Update(bson.M{"_id": bson.ObjectIdHex(userInfo.Mid), "userinfo.coin": bson.M{"$gte": -coin}}, bson.M{"$inc": bson.M{"userinfo.coin": coin}})
	if err != nil {
		return false
	} else {
		userInfo.Coin += coin
		return true
	}
}

type reqMoneyAdd struct {
	Money     int64  `json:"money"`
	RequestId string `json:"request_id"`
}

type retMoneyAdd struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 加钱 /user/money_add?access_token=ACCESS_TOKEN
func PopAddCoin(coin int64, token string) error {
	var addr string
	if coin > 0 {
		addr = fmt.Sprintf("http://%v:%v/user/money_add?access_token=%v", conf.Cf.EnvConf.Pop.Host, conf.Cf.EnvConf.Pop.Port, token)
	} else {
		coin = -coin
		addr = fmt.Sprintf("http://%v:%v/user/money_deduct?access_token=%v", conf.Cf.EnvConf.Pop.Host, conf.Cf.EnvConf.Pop.Port, token)
	}
	var req reqMoneyAdd
	req.Money = coin
	req.RequestId = bson.NewObjectId().Hex()
	bts, _ := json.Marshal(req)
	b, err := utils.HttpPostJson(addr, bts)
	if err != nil {
		return err
	}
	var ret retMoneyAdd
	err2 := json.Unmarshal(b, &ret)
	if err2 != nil {
		return fmt.Errorf("%v,%v", err2, string(b))
	}
	if ret.Code != 0 {
		return fmt.Errorf("%v,%v", ret.Message, ret.Code)
	}
	return nil
}

type reqPopRefreshToken struct {
	AppKey       string `json:"k"`
	RefreshToken string `json:"refresh_token"`
}

type retPopRefreshToken struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshToken     string `json:"refresh_token"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	} `json:"data"`
}

func refreshToken(userInfo *UserInfo) error {
	var req reqPopRefreshToken
	req.AppKey = conf.Cf.EnvConf.Pop.AppKey
	req.RefreshToken = userInfo.Account.RefreshToken
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("http://%v:%v/user/refresh_token", conf.Cf.EnvConf.Pop.Host, conf.Cf.EnvConf.Pop.Port)

	b2, err := utils.HttpPostJson(addr, b)
	if err != nil {
		return err
	}
	var ret retPopRefreshToken
	err = json.Unmarshal(b2, &ret)
	if err != nil {
		return err
	}
	if ret.Code != 0 {
		return fmt.Errorf("pop api err. code: %v msg:%v", ret.Code, ret.Message)
	}
	userInfo.Account.AccessToken = ret.Data.AccessToken
	userInfo.Account.ExpiresIn = ret.Data.ExpiresIn
	userInfo.Account.RefreshToken = ret.Data.RefreshToken
	userInfo.Account.RefreshExpiresIn = ret.Data.RefreshExpiresIn
	db := mdb.GetMdb()
	db.C(mdb.DB_USER).Update(bson.M{"_id": bson.ObjectIdHex(userInfo.Mid)}, bson.M{"$set": bson.M{"userinfo": userInfo}})
	return nil
}

// 扣金币
func (UserInfo *UserInfo) ReduceCoin() (ok bool) {
	return
}

func Test() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRJZCI6MSwiaWF0IjoxNzAwMjAyMTE1LCJleHAiOjE3MDAyMDkzMTV9.kMtvoy0wFLhkUCCFXi0JNKMST2dfWNRrLwKg02srg2o"
	fmt.Println(PopAddCoin(1, token))
}
