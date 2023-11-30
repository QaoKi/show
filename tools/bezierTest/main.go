package main

import (
	"bygame/common/utils"
)

func main() {
	path := getRandBezier()
	gen(path)
	length(path)
}

type vec2 struct {
	X float64
	Y float64
}

func getRandBezier() (path [4]vec2) {
	// 横屏 1920 1080 服务器屏幕扩大 3920 * 3080 贝塞尔曲线的起始点要在客户端屏幕外
	// 起点在屏幕外左侧
	clientX, clientY := 860, 540
	serverX, serverY := 900, 600
	path[0] = vec2{X: float64(utils.RandInt(-serverX, -clientX)), Y: float64(utils.RandInt(-serverY, serverY))}
	// 第一个控制点在屏幕左侧并且范围略小于屏幕范围防止擦边出界
	path[1] = vec2{X: float64(utils.RandInt(-clientX, (-clientX/10)*8)) * 0.8, Y: float64(utils.RandInt(-clientY, clientY)) * 0.8}
	// 第二个控制点在屏幕右侧并且范围略小于屏幕范围防止擦边出界
	path[2] = vec2{X: float64(utils.RandInt((clientX/10)*2, clientX)) * 0.8, Y: float64(utils.RandInt(-clientY, clientY)) * 0.8}
	// 终点在屏幕外右侧
	path[3] = vec2{X: float64(utils.RandInt(clientX, serverX)), Y: float64(utils.RandInt(-serverY, serverY))}
	return
}
