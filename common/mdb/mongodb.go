package mdb

import (
	"bygame/common/conf"
	"bygame/common/log"
	"time"

	"gopkg.in/mgo.v2"
)

var _mdb *Database
var dbSession *mgo.Session

const (
	DB_USER       = "user"
	DB_ACCOUNT    = "account"
	DB_GAMEDATA   = "gamedata"
	DB_SLOTRECORD = "slotrecord"
)

func GetMdb() *Database {
	return &Database{Database: dbSession.DB(conf.Cf.EnvConf.Mongo.Db)}
}

func Init() bool {

	authDB := conf.Cf.EnvConf.Mongo.Db

	dialInfo := &mgo.DialInfo{
		Addrs:     []string{conf.Cf.EnvConf.Mongo.Host},
		PoolLimit: 25,
		Timeout:   time.Second * time.Duration(3),
		Database:  authDB,
		Username:  "",
		Password:  "",
	}
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return false
	}
	log.Inf("mongo connect %v success", dialInfo.Addrs)
	dbSession = session
	session.SetMode(mgo.Monotonic, true)
	_mdb = &Database{Database: session.DB(conf.Cf.EnvConf.Mongo.Db)}
	initSysIndexes()
	return true
}

/*
 * SetIndex_details("users", []string{"uid", "name"}, false)
 * SetIndex("adminUser", "uid", false)
**/
func initSysIndexes() {
	SetIndex(DB_USER, "userinfo.uid", true)
}

func SetIndex(coll, key string, unique bool) {
	setIndex_details(coll, "", []string{key}, unique, false, 0)
}

func SetIndex_details(coll string, keys []string, unique bool) {
	setIndex_details(coll, "", keys, unique, false, 0)
}

func setIndex_details(coll, name string, keys []string, unique, sparse bool, dur time.Duration) {
	index := mgo.Index{
		Name:        name,
		Key:         keys,
		Unique:      unique,
		Sparse:      sparse,
		ExpireAfter: dur,
	}
	nano := time.Now().UnixNano()
	err := _mdb.C(coll).EnsureIndex(index)
	log.Inf("MongoDB set collection %v index %v, time:%v, err:%v", coll, keys, time.Duration(time.Now().UnixNano()-nano), err)
}
