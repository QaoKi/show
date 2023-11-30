package rdb_test

import (
	"bygame/common/log"
	"bygame/common/rdb"
	"testing"
)

func TestMain(m *testing.M) {
	rdb.Init("center")
	var req rdb.CloseReq
	var ret rdb.CloseRet
	rdb.Request("slots:s2", &req, &ret)
	log.Dbg("ret %+v", ret)
}
