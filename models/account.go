package models

import (
	"bygame/common/conf"
	"bygame/common/data"
	"bygame/common/log"
	"bygame/common/mdb"
	"bygame/common/proto/common_proto"
	"bygame/common/rdb"
	"bygame/common/utils"

	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	Guest = "101" // 游客登录
	Zy    = "103" // 之艺登录
	Pop   = "104" // pop
)

const (
	ZyApiUserInfo = "/game-api/user-info"
	ZyApiAddCoin  = "/game-api/add-coin"
)

type ZyBase struct {
	Status string `json:"status"`
	ErrMsg string `json:"err_msg"`
}

type ZyUser struct {
	ZyBase
	Uid      string `json:"uid"`
	NickName string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Coins    string `json:"coins"`
	Exp      string `json:"exp"`
	Gender   string `json:"gender"`
}

func userInfoZy(sign string) (*ZyUser, error) {
	bts, err := utils.Reqzy(ZyApiUserInfo, sign, nil)
	if err != nil {
		return nil, err
	}
	var zu ZyUser
	err = json.Unmarshal(bts, &zu)
	if err != nil {
		return nil, err
	}

	if zu.Status == "err" {
		return nil, fmt.Errorf(zu.ErrMsg)
	}
	return &zu, nil
}

func AddCoinZy(sign string, coin int) error {
	bts, err := utils.Reqzy(ZyApiAddCoin, sign, map[string]string{"add_coin": fmt.Sprint(coin)})
	if err != nil {
		return err
	}
	var zu ZyBase
	err = json.Unmarshal(bts, &zu)
	if err != nil {
		return err
	}

	if zu.Status == "err" {
		return fmt.Errorf(zu.ErrMsg)
	}
	return nil
}

func RegistZy(account *data.Account, zu ZyUser) (*data.User, error) {
	var userInfo data.UserInfo
	userInfo.Avatar = zu.Avatar

	coins, _ := strconv.Atoi(zu.Coins)
	gender, _ := strconv.Atoi(zu.Gender)
	userInfo.Coin = int64(coins)
	userInfo.Gender = int32(gender)
	userInfo.NickName = zu.NickName
	userInfo.Uid = zu.Uid

	db := mdb.GetMdb()
	now := int(time.Now().Unix())
	var user data.User
	user.Bid = bson.NewObjectId()
	user.Mid = user.Bid.Hex()
	user.LastTime = now
	user.RegisterTime = now
	user.LoginDay = 1
	user.UserInfo = userInfo
	user.UserInfo.Mid = user.Mid
	account.Mid = user.Mid
	db.C(mdb.DB_USER).Insert(user)
	db.C(mdb.DB_ACCOUNT).Insert(account)
	return &user, nil
}

func AddCoin(userInfo *common_proto.UserInfo, increment int64) bool {
	if increment > 0 {
		atomic.AddInt64(&userInfo.CoinOffset, increment)
		atomic.AddInt64(&userInfo.Coin, increment)
		return true
	}

	if increment < 0 {
		if atomic.LoadInt64(&userInfo.Coin) < increment {
			return false
		}
		atomic.AddInt64(&userInfo.CoinOffset, increment)
		atomic.AddInt64(&userInfo.Coin, increment)
		return true
	}
	return true
}

func GetUserInfoCoin(userInfo *common_proto.UserInfo) int64 {
	return atomic.LoadInt64(&userInfo.Coin)
}

// 扣金币 余额暂时没有
func UpdateCoin(account *data.Account, increment int) (coin int, ok bool) {
	switch account.Platform {
	case Zy:
		err := AddCoinZy(account.Sign, increment)
		if err != nil {
			return 0, false
		} else {
			return 0, true
		}
	default:
		db := mdb.GetMdb()
		err := db.C(mdb.DB_USER).Update(bson.M{"_id": bson.ObjectIdHex(account.Mid), "userinfo.coin": bson.M{"$gte": -increment}}, bson.M{"$inc": bson.M{"userinfo.coin": increment}})
		if err != nil {
			return 0, false
		} else {
			return 0, true
		}
	}
}

type reqPopUserToken struct {
	AppKey  string `json:"k"`
	Time    string `json:"t"` // s
	IdToken string `json:"id"`
	Sign    string `json:"sign"`
}

type retPopUserToken struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshToken     string `json:"refresh_token"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		User             struct {
			Uid      string `json:"uid"`
			NickName string `json:"nickName"`
			Avatar   string `json:"avatar"`
			Money    string `json:"money"`
		} `json:"user"`
	} `json:"data"`
}

func PopLogin(idToken string) (*data.User, error) {
	addr := fmt.Sprintf("http://%v:%v/user/token", conf.Cf.EnvConf.Pop.Host, conf.Cf.EnvConf.Pop.Port)
	now := time.Now().Unix()
	var req reqPopUserToken
	req.AppKey = conf.Cf.EnvConf.Pop.AppKey
	req.Time = fmt.Sprint(now)
	req.IdToken = idToken
	sign := utils.PopSign(req, conf.Cf.EnvConf.Pop.AppSecret)
	req.Sign = sign
	bts, _ := json.Marshal(req)
	log.Dbg("pop登录请求数据 %+v", req)

	b, err := utils.HttpPostJson(addr, bts)
	if err != nil {
		return nil, err
	}

	var ret retPopUserToken
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	}

	if ret.Code != 0 {
		return nil, fmt.Errorf(ret.Message)
	}
	log.Inf("pop login ret %+v", ret)

	db := mdb.GetMdb()
	var u data.User
	db.C(mdb.DB_USER).Find(bson.M{"userinfo.account.code": ret.Data.User.Uid}).One(&u)
	if u.Mid == "" {
		// 需要新建用户
		u.Bid = bson.NewObjectId()
		u.Mid = u.Bid.Hex()
		u.LastTime = int(now)
		u.RegisterTime = int(now)
		u.LoginDay = 1
		var userInfo data.UserInfo
		userInfo.Mid = u.Mid
		userInfo.Uid = ret.Data.User.Uid
		result := rdb.Client().Incr(context.TODO(), "bygame:uid")
		if result.Err() != nil {
			return nil, result.Err()
		}
		userInfo.Index = int(result.Val())
		userInfo.NickName = ret.Data.User.NickName
		coin, _ := strconv.Atoi(ret.Data.User.Money)
		userInfo.Coin = int64(coin)
		userInfo.Avatar = ret.Data.User.Avatar
		userInfo.Gender = 0
		var account data.Account
		account.Platform = Pop
		account.Code = ret.Data.User.Uid
		account.AccessToken = ret.Data.AccessToken
		account.RefreshToken = ret.Data.RefreshToken
		account.ExpiresIn = ret.Data.ExpiresIn
		account.RefreshExpiresIn = ret.Data.RefreshExpiresIn
		userInfo.Account = account
		u.UserInfo = userInfo
		db.C(mdb.DB_USER).Insert(u)
	} else {
		// 更新数据
		u.LastTime = int(now)
		u.LoginDay = 1
		u.UserInfo.NickName = ret.Data.User.NickName
		coin, _ := strconv.Atoi(ret.Data.User.Money)
		u.UserInfo.Coin = int64(coin)
		u.UserInfo.Avatar = ret.Data.User.Avatar
		u.UserInfo.Account.AccessToken = ret.Data.AccessToken
		u.UserInfo.Account.RefreshToken = ret.Data.RefreshToken
		u.UserInfo.Account.ExpiresIn = ret.Data.ExpiresIn
		u.UserInfo.Account.RefreshExpiresIn = ret.Data.RefreshExpiresIn
		db.C(mdb.DB_USER).Upsert(bson.M{"_id": u.Bid}, bson.M{"$set": u})
	}
	return &u, nil
}

func GuestLogin(deviceId string) (*data.User, error) {
	db := mdb.GetMdb()
	var u data.User
	db.C(mdb.DB_USER).Find(bson.M{"userinfo.account.code": deviceId}).One(&u)
	if u.Mid == "" {
		// 需要新建用户
		u.Bid = bson.NewObjectId()
		u.Mid = u.Bid.Hex()
		u.LastTime = int(time.Now().Unix())
		u.RegisterTime = int(time.Now().Unix())
		u.LoginDay = 1
		var userInfo data.UserInfo
		userInfo.Mid = u.Mid
		result := rdb.Client().Incr(context.TODO(), "bygame:uid")
		if result.Err() != nil {
			return nil, result.Err()
		}
		userInfo.Index = int(result.Val())
		userInfo.Uid = fmt.Sprintf("101%03d%03d", result.Val(), utils.RandInt(0, 999))
		userInfo.NickName = utils.RandEasyIntCode(4)
		userInfo.Coin = 1000 * 1000 * 200
		userInfo.Gender = int32(utils.RandInt(1, 2))
		if userInfo.Gender == 1 {
			userInfo.Avatar = fmt.Sprintf("https://aimo-test.oss-cn-shanghai.aliyuncs.com/assets/avatar/%s%d.jpg", "boy", utils.RandInt(1, 4))
		} else {
			userInfo.Avatar = fmt.Sprintf("https://aimo-test.oss-cn-shanghai.aliyuncs.com/assets/avatar/%s%d.jpg", "girl", utils.RandInt(1, 4))
		}
		var account data.Account
		account.Platform = Guest
		account.Code = deviceId
		userInfo.Account = account
		u.UserInfo = userInfo
		db.C(mdb.DB_USER).Insert(u)
	}
	return &u, nil
}

func ZyLogin(sign string) (*data.User, error) {
	zu, err := userInfoZy(sign)
	if err != nil {
		return nil, err
	}

	db := mdb.GetMdb()
	var u data.User
	db.C(mdb.DB_USER).Find(bson.M{"userinfo.account.code": zu.Uid}).One(&u)
	if u.Mid == "" {
		// 需要新建用户
		u.Bid = bson.NewObjectId()
		u.Mid = u.Bid.Hex()
		u.LastTime = int(time.Now().Unix())
		u.RegisterTime = int(time.Now().Unix())
		u.LoginDay = 1
		var userInfo data.UserInfo
		u.UserInfo = userInfo
		userInfo.Mid = u.Mid
		userInfo.Uid = zu.Uid
		result := rdb.Client().Incr(context.TODO(), "bygame:uid")
		if result.Err() != nil {
			return nil, result.Err()
		}
		userInfo.Index = int(result.Val())
		userInfo.NickName = zu.NickName
		coin, _ := strconv.Atoi(zu.Coins)
		userInfo.Coin = int64(coin)
		userInfo.Avatar = zu.Avatar
		gender, _ := strconv.Atoi(zu.Gender)
		userInfo.Gender = int32(gender)
		var account data.Account
		account.Platform = Zy
		account.Code = zu.Uid
		account.Sign = sign
		userInfo.Account = account
		db.C(mdb.DB_USER).Insert(u)
	}
	return &u, nil
}
