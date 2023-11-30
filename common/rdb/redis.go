package rdb

import (
	"bygame/common/conf"
	"bygame/common/log"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gopkg.in/mgo.v2/bson"
)

const (
	rpcReq = 1
	rpcRet = 0
)

type BaseRpc struct {
	ErrCode int    `json:"errCode"`  // 错误码
	ErrMsg  string `json:"errMsg"`   // 错误信息
	RId     string `json:"_reqId"`   // 请求id
	FS      string `json:"_from"`    // 发起服务
	Type    int    `json:"_method"`  // 1 请求响应 2 响应
	Cmd     string `json:"_reqType"` // 消息号
}

var rmq *redisMQ

const prefix = "game:inner_rpc:"

type redisMQ struct {
	Client *redis.Client       // redis 客户端
	ch     chan *redis.Message // redis 订阅信道

	mu      sync.RWMutex
	retm    map[string]chan []byte       // 响应map
	reqm    map[string]func(data []byte) // 请求map,这个map读写在逻辑上是分离的所以不用锁
	sid     string                       // 服务当前id
	channel string                       // 服务监听的redis通道
}

func Init(sid string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Cf.EnvConf.Redis.Host + ":" + conf.Cf.EnvConf.Redis.Port,
		Password: conf.Cf.EnvConf.Redis.Password,
		DB:       conf.Cf.EnvConf.Redis.Db,
		PoolSize: 25,
	})

	rmq = &redisMQ{
		Client:  client,
		ch:      make(chan *redis.Message, 100000),
		retm:    make(map[string]chan []byte),
		reqm:    make(map[string]func(data []byte)),
		sid:     sid,
		channel: prefix + sid,
	}
	c := rmq.Client.Subscribe(context.Background(), rmq.channel)
	_, err := c.Receive(context.Background())
	if err != nil {
		return err
	}
	go func() {
		ch := c.Channel()
		for msg := range ch {
			rmq.ch <- msg
		}
	}()
	onMessage()
	return nil
}

func Pub(channel string, message string) error {
	return rmq.Client.Publish(context.Background(), channel, message).Err()
}

/*
	exp 过期时间设置为0即不过期
*/
func Set(key string, value any, exp time.Duration) {
	rmq.Client.Set(context.Background(), key, value, exp)
}

func Get(key string) *redis.StringCmd {
	return rmq.Client.Get(context.Background(), key)
}

func Del(key string) *redis.IntCmd {
	return rmq.Client.Del(context.Background(), key)
}

func Client() *redis.Client {
	return rmq.Client
}

func Request(sid string, req any, ret any) error {
	rid := bson.NewObjectId().Hex()
	m := make(map[string]any)
	b, _ := json.Marshal(req)
	json.Unmarshal(b, &m)
	m["_reqId"] = rid
	m["_from"] = rmq.sid
	m["_method"] = rpcReq
	m["_reqType"] = strings.Replace(strings.Replace(fmt.Sprintf("%v", reflect.TypeOf(req)), "*", "", 1), "rdb.", "", 1)

	ch := make(chan []byte)
	rmq.mu.Lock()
	rmq.retm[rid] = ch
	rmq.mu.Unlock()
	pub(sid, m)
	t := time.After(time.Second * 5)
	select {
	case b := <-ch:
		err := json.Unmarshal(b, ret)
		if err != nil {
			return err
		}
		m := make(map[string]any)
		json.Unmarshal(b, &m)
		if msg, ok := m["ErrMsg"]; ok && msg != "" {
			return fmt.Errorf("%s", msg)
		}
	case <-t:
		return errors.New("time out")
	}
	return nil
}

type reqFunc[TReq, TRet any] func(req *TReq, ret *TRet)

func Router[TReq any, TRet any](fn reqFunc[TReq, TRet]) {
	var req TReq
	f := func(data []byte) {
		err := json.Unmarshal(data, &req)
		var baseReq BaseRpc
		json.Unmarshal(data, &baseReq)
		if err != nil {
			fmt.Printf("反序列化错误：%s\n", data)
		}
		var ret TRet
		fn(&req, &ret)
		response(baseReq.RId, baseReq.FS, &ret)
	}
	fmt.Printf("%v\n", reflect.TypeOf(req))
	rmq.reqm[fmt.Sprintf("%v", reflect.TypeOf(req))] = f
}

// **************inner****************

func onMessage() error {
	go func() {
		for msg := range rmq.ch {
			var data BaseRpc
			bytes := []byte(msg.Payload)
			err := json.Unmarshal(bytes, &data)
			log.Dbg("redis 收到消息 %v", string(bytes))
			if err != nil {
				fmt.Printf("redis mq message unmarshal err. %v\n", err)
				continue
			}

			if data.Type == rpcRet {
				rmq.mu.RLock()
				ch, ok := rmq.retm[data.RId]
				rmq.mu.RUnlock()
				if ok {
					ch <- bytes
				}
				continue
			}

			if data.Type == rpcReq {
				if f, ok := rmq.reqm[data.Cmd]; ok {
					f(bytes)
				}
				continue
			}
		}
	}()
	return nil
}

// 响应一个消息
func response(requestId string, fromServer string, ret any) {
	val := reflect.ValueOf(ret).Elem()
	uidField := val.FieldByName("RId")
	if uidField.IsValid() && uidField.CanSet() {
		uidField.SetString(requestId)
	}
	pub(fromServer, ret)
}

func pub(sid string, ret any) {
	bts, _ := json.Marshal(ret)
	log.Dbg("请求, %v,%v", prefix+sid, string(bts))
	Pub(prefix+sid, string(bts))
}
