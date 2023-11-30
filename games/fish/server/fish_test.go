package fish

import (
	"bygame/common/conf"
	"bygame/common/rdb"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestSpecialFormat1Bezier(t *testing.T) {
	paths := genSpecialFormat1Bezier(10)
	fmt.Printf("%v\n", len(paths))

}

func TestSpecialFormat2Bezier(t *testing.T) {
	n := 1
	for i := 0; i < 20; i++ {
		path := genSpecialFormat2Bezier(int32(i), n)
		fmt.Printf("start [%.2f, %.2f],  end[%.2f, %.2f]\n", path[0].X, path[0].Y, path[3].X, path[3].Y)
	}

}

func TestSpecialFormat3Bezier(t *testing.T) {
	outerPaths, innerPaths, midPath := genSpecialFormat3Bezier(3, 3, true)
	for i := 0; i < len(outerPaths); i++ {
		fmt.Printf("out start [%.2f, %.2f],  end[%.2f, %.2f]\n", outerPaths[i][0].X, outerPaths[i][0].Y, outerPaths[i][3].X, outerPaths[i][3].Y)
	}

	for i := 0; i < len(innerPaths); i++ {
		fmt.Printf("inner start [%.2f, %.2f],  end[%.2f, %.2f]\n", innerPaths[i][0].X, innerPaths[i][0].Y, innerPaths[i][3].X, innerPaths[i][3].Y)
	}

	fmt.Printf("mid start [%.2f, %.2f],  end[%.2f, %.2f]\n", midPath[0].X, midPath[0].Y, midPath[3].X, midPath[3].Y)

}

func TestGetRoom(t *testing.T) {
	conf.Init()
	initRoomM()
	//roomList := []*room{}
	//for i := 0; i < 10; i++ {
	// r, _ := _roomM.getRoom(0)
	// roomList = append(roomList, r)
	// fmt.Printf("get room, room id[%s], curr user count[%d]\n", r.roomId, len(r.players))
	// r.players = append(r.players, "1")
	// fmt.Printf("len roomMap[%d]\n", len(_roomM.allRoomMapList[0]))
	// for count, list := range _roomM.allRoomMapList[0] {
	// 	fmt.Printf("count[%d]\n", count)
	// 	for _, r := range list {
	// 		fmt.Printf("房间id[%s], 人数[%d]\n", r.roomId, len(r.players))
	// 	}
	// 	fmt.Printf("\n")
	// }
	//}

	fmt.Println("=================================")

	for i := 0; i < 10; i++ {
		// fmt.Printf("return room, room id[%s], curr user count[%d]\n", roomList[i].roomId, len(roomList[i].players))
		// roomList[i].players = roomList[i].players[1:]
		// _roomM.returnRoom(roomList[i])

		// for count, list := range _roomM.allRoomMapList[0] {
		// 	fmt.Printf("count[%d]\n", count)
		// 	for _, r := range list {
		// 		fmt.Printf("房间id[%s], 人数[%d]\n", r.roomId, len(r.players))
		// 	}
		// 	fmt.Printf("\n")
		// }
	}

}

func TestGenFish(t *testing.T) {
	ctxStop, stopFunc := context.WithCancel(context.Background())
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	i := 0
	go func() {
		defer fmt.Printf("go func exit\n")

		for {
			select {
			case <-ctxStop.Done():
				fmt.Printf("gen fish stop\n")
				return
			case <-ticker.C:
				fmt.Printf("i[%d]\n", i)
				i++

				// 一轮鱼出完了，出鱼阵
				if i == 3 {
					f(ctxStop)
					i = 0
				}
			}
		}
	}()

	time.Sleep(15 * time.Second)
	fmt.Printf("stop\n")
	stopFunc()
	fmt.Printf("stop wait\n")
	time.Sleep(3 * time.Second)
}

func f(ctx context.Context) {
	fmt.Printf("f start\n")
	defer fmt.Printf("f exit\n")
	ticker := time.NewTicker(3 * time.Second)
	j := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("f stop\n")
			return
		case <-ticker.C:
			fmt.Printf("j[%d]\n", j)
			j++
			if j == 3 {
				return
			}
		}
	}
}

func TestTableConfig(t *testing.T) {
	conf.Init()
	server, ok := conf.Cf.ServerConf.GetServer("1")
	if !ok {
		fmt.Println("get server fail")
		return
	}

	conf.Cf.SetSelfServerConf(server)
	rdb.Init("center")
	initGameM()

	//time.Sleep(3 * time.Second)
	initGameM()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)

	<-sigCh
}

func TestRobot(t *testing.T) {
	conf.Init()

	robot, b := newRobot(0)
	if !b {
		fmt.Println("创建机器人失败")
		return
	}

	robot.startGame()
	time.Sleep(30 * time.Second)
	// sigCh := make(chan os.Signal, 1)
	// signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)

	// <-sigCh

	robot.stopGame()
}

func TestTime(t *testing.T) {
	tr := time.NewTimer(0)
	for {
		select {
		case <-tr.C:
		}

	}
}
