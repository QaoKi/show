package main

import (
	"bygame/common/proto/common_proto"
	"bygame/common/proto/fish_proto"
	"fmt"
	"os"
	"syscall"
	"time"

	"os/signal"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	address := "127.0.0.1:8888"
	gameId := "5"
	protoc := ""
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDc5OTE4NTksImp0aSI6IjY1NGEwYTAwMjE4Y2I3NjkxNTA4NmIzMiIsInN1YiI6IjEwMTAwMjAwMiJ9.zh3f5BcrUhg7rlt8QnLtcpTSels0vCyA628WQmlMNhA"

	url := fmt.Sprintf("ws://%s/wss?gameId=%s&protoc=%s&token=%s", address, gameId, protoc, token)

	fmt.Println("url: ", url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}

	defer conn.Close()

	go writeMessage(conn)
	go readMessage(conn)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)

	<-sigCh
}

func writeMessage(conn *websocket.Conn) {

	// 加入房间
	req := &fish_proto.ReqUserJoinRoom{}
	req.RoomType = 1
	send(conn, int32(common_proto.CmdId_UserJoinRoom), req)

	ticker := time.NewTicker(2 * time.Second)
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			fmt.Println("退出房间")
			// 退出房间
			req := &fish_proto.ReqUserQuitRoom{}
			req.RoomType = 1
			send(conn, int32(common_proto.CmdId_UserQuitRoom), req)
		}
	}
}

func send(conn *websocket.Conn, cmdId int32, m proto.Message) {
	baseCmd := &common_proto.BaseCmd{}
	baseCmd.ReqId = 123

	baseCmd.CmdId = cmdId
	baseCmd.Data, _ = proto.Marshal(m)

	message, _ := proto.Marshal(baseCmd)
	err := conn.WriteMessage(websocket.BinaryMessage, message)
	if err != nil {
		fmt.Println("写数据错误：", err)
		return
	}
}

func readMessage(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("读数据出错[%v]\n", err)
			return
		}

		baseCmd := &common_proto.BaseCmd{}
		err = proto.Unmarshal(message, baseCmd)
		if err != nil {
			fmt.Printf("读数据解析出错[%v]\n", err)
			return
		}

		fmt.Printf("接受到数据，消息号[%d]\n", baseCmd.CmdId)
		analyData(baseCmd.CmdId, baseCmd.Data)
	}
}

func analyData(cmdId int32, b []byte) {
	switch cmdId {
	case int32(common_proto.CmdId_UserJoinRoom):
		ret := fish_proto.RetUserJoinRoom{}
		err := proto.Unmarshal(b, &ret)
		if err != nil {
			fmt.Printf("加入房间数据解析出错[%v]\n", err)
			return
		}

		fmt.Printf("加入房间回复\n")

		if ret.RoomInfo != nil {
			fmt.Printf("现有的鱼：\n")
			for _, f := range ret.RoomInfo.Fishs {
				fmt.Printf("	[%s]\n", f.FishId)
			}
			fmt.Printf("现有的玩家：\n")
			for _, u := range ret.RoomInfo.UserInfos {
				fmt.Printf("	[%s]\n", u.UserInfo.Mid)
			}
		} else {
			fmt.Println("房间信息是空")
		}
	}
}
