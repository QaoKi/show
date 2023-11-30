package rdb

type CloseReq struct {
	BaseRpc
}

type CloseRet struct {
	BaseRpc
	Time int `json:"time"`
}
