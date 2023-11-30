package fish

import (
	"bygame/common/utils"
	"math"
)

var (
	_MaxX float64 = 860 // 客户端屏幕右上角的x轴
	_MaxY float64 = 540 // 客户端屏幕右上角的y轴

	_BigFishLength float64 = 650 // 大鱼的长度
	_BigFishWide   float64 = 480 // 大鱼的宽度

	_MidFishLength float64 = 300 // 中鱼的长度
	_MidFishWide   float64 = 200 // 中鱼的宽度

	_SmallFishLength float64 = 100 // 小鱼的长度
	_SmallFishWide   float64 = 50  // 小鱼的宽度

	// 终点的右边 x 轴
	_EndXRight float64 = 1400.0
	// 终点的左边 x 轴
	_EndXLeft float64 = -1400.0
)

// 生成随机三次贝塞尔曲线
func getRandBezier() (paths [][4]vec2) {
	// 横屏 1920 1080 服务器屏幕扩大 3920 * 3080 贝塞尔曲线的起始点要在客户端屏幕外
	// 起点在屏幕外左侧
	clientX, clientY := 860, 540
	serverX, serverY := 900, 600
	var path [4]vec2
	path[0] = vec2{X: float64(utils.RandInt(-serverX, -clientX)), Y: float64(utils.RandInt(-serverY, serverY))}
	// 第一个控制点在屏幕左侧并且范围略小于屏幕范围防止擦边出界
	path[1] = vec2{X: float64(utils.RandInt(-clientX, (-clientX/10)*8)) * 0.8, Y: float64(utils.RandInt(-clientY, clientY)) * 0.8}
	// 第二个控制点在屏幕右侧并且范围略小于屏幕范围防止擦边出界
	path[2] = vec2{X: float64(utils.RandInt((clientX/10)*2, clientX)) * 0.8, Y: float64(utils.RandInt(-clientY, clientY)) * 0.8}
	// 终点在屏幕外右侧
	path[3] = vec2{X: float64(utils.RandInt(clientX, serverX)), Y: float64(utils.RandInt(-serverY, serverY))}

	paths = append(paths, path)
	return
}

func genSpecialFormat1Bezier(pathNum int32) (paths [][4]vec2) {

	// 起点在原点
	startX, startY := 0.0, 0.0
	// 终点
	endX, endY := _EndXRight, 0.0
	ends := getSpinCoord(pathNum, startX, startY, endX, endY)

	for i := 0; i < len(ends); i++ {
		path := [4]vec2{}
		path[0] = vec2{startX, startY}
		path[3] = vec2{ends[i].X, ends[i].Y}
		path1, path2 := genTwoXY(startX, startY, ends[i].X, ends[i].Y)
		path[1] = path1
		path[2] = path2

		paths = append(paths, path)
	}

	return
}

// 创建第二鱼阵的位置，lot:第几批次，wathPath:第几线路
func genSpecialFormat2Bezier(lot int32, wathPath int) (path [4]vec2) {
	// 第一线路的y轴
	firstPathY := _MaxY - _SmallFishWide/2
	// 第二线路的y轴
	secondPathY := 0.0
	// 第三线路的y轴
	thirdPathY := -(_MaxY - _SmallFishWide/2)

	// 第一线路和第三线路，每条鱼之间的间隔
	firstInteral := _SmallFishLength + 40
	// 第二线路，每条鱼之间的间隔
	secondInteral := _BigFishWide + 40

	// y轴
	y := firstPathY
	if wathPath == 2 {
		y = secondPathY
	} else if wathPath == 3 {
		y = thirdPathY
	}

	// 鱼之间的间隔
	interval := firstInteral
	if wathPath == 2 {
		interval = secondInteral
	}

	startX := -_MaxX - interval/2 - float64(lot)*interval

	path[0] = vec2{startX, y}
	path[3] = vec2{_EndXRight, y}
	path1, path2 := genTwoXY(startX, y, _EndXRight, y)
	path[1] = path1
	path[2] = path2
	return
}

// 创建第三鱼阵的位置,outerNum, innerNum:外圈鱼和内圈鱼的数量, bleft:是否是屏幕的左侧
func genSpecialFormat3Bezier(outerNum, innerNum int32, bLeft bool) (outerPaths [][4]vec2, innerPaths [][4]vec2, midPath [4]vec2) {
	// 外圈的半径
	outerR := _MaxY - _SmallFishWide/2
	// 内圈的半径
	innerR := outerR - _SmallFishWide

	{
		startX := -_MaxX - outerR
		endX := _EndXRight
		if !bLeft {
			// 右侧出鱼
			startX = -startX
			endX = -endX
		}

		y := 0.0
		midPath[0] = vec2{startX, y}
		midPath[3] = vec2{endX, y}
		path1, path2 := genTwoXY(startX, y, endX, y)
		midPath[1] = path1
		midPath[2] = path2

	}

	{
		// 外圈鱼
		startX, startY := -_MaxX-_SmallFishWide/2, 0.0
		coordX, coordY := startX-outerR, 0.0
		endX := _EndXRight
		if !bLeft {
			startX = -startX
			coordX = -coordX
			endX = _EndXLeft
		}

		starts := getSpinCoord(outerNum, coordX, coordY, startX, startY)

		for i := 0; i < len(starts); i++ {
			path := [4]vec2{}
			path[0] = vec2{starts[i].X, starts[i].Y}
			path[3] = vec2{endX, starts[i].Y}
			path1, path2 := genTwoXY(starts[i].X, starts[i].Y, endX, starts[i].Y)
			path[1] = path1
			path[2] = path2

			outerPaths = append(outerPaths, path)
		}
	}

	{
		// 内圈鱼
		startX, startY := -_MaxX-_SmallFishWide/2-_SmallFishWide, 0.0
		coordX, coordY := startX-innerR, 0.0
		endX := _EndXRight
		if !bLeft {
			startX = -startX
			coordX = -coordX
			endX = _EndXLeft
		}

		starts := getSpinCoord(innerNum, coordX, coordY, startX, startY)

		for i := 0; i < len(starts); i++ {
			path := [4]vec2{}
			path[0] = vec2{starts[i].X, starts[i].Y}
			path[3] = vec2{endX, starts[i].Y}
			path1, path2 := genTwoXY(starts[i].X, starts[i].Y, endX, starts[i].Y)
			path[1] = path1
			path[2] = path2

			innerPaths = append(outerPaths, path)
		}
	}

	return
}

// 创建第四鱼阵的位置，index:当前线路中的第几条鱼, y:这条鱼的y轴, bLeft:是否是屏幕左边出来的鱼
func genSpecialFormat4Bezier(index int32, y float64, bLeft bool) (path [4]vec2) {

	// 每条鱼的间隔
	interval := _MidFishLength + 40

	startX := -_MaxX - interval/2 - float64(index)*interval
	endX := _EndXRight
	if !bLeft {
		startX = -startX
		endX = -endX
	}

	path[0] = vec2{startX, y}
	path[3] = vec2{endX, y}
	path1, path2 := genTwoXY(startX, y, endX, y)
	path[1] = path1
	path[2] = path2
	return
}

// 一条线段，以起点为旋转点，旋转一定角度后的终点坐标
func getSpinCoord(pathNum int32, startX, startY float64, endX, endY float64) (ends []vec2) {
	// 角度
	degrees := float64(360.0) / float64(pathNum)
	// 弧度
	radians := degrees * (math.Pi / 180.0)

	for i := 0; i < int(pathNum); i++ {

		ends = append(ends, vec2{endX, endY})

		//旋转
		// 计算另一端相对于中心的坐标差
		dx := endX - startX
		dy := endY - startY

		// 对坐标差应用旋转变换
		x2Prime := dx*math.Cos(radians) - dy*math.Sin(radians)
		y2Prime := dx*math.Sin(radians) + dy*math.Cos(radians)

		// 计算旋转后的端点坐标
		endX = startX + x2Prime
		endY = startY + y2Prime
	}

	return
}

// 在一条直线上取两个点
func genTwoXY(startX, startY float64, endX, endY float64) (path1 vec2, path2 vec2) {
	// 取一个中间点
	midX := (startX + endX) / 2
	midY := (startY + endY) / 2
	path1 = vec2{midX, midY}

	// 取mid和终点之间的中间点
	mid2X := (midX + endX) / 2
	mid2Y := (midY + endY) / 2
	path2 = vec2{mid2X, mid2Y}
	return
}

func bezierLength(path [4]vec2) float64 {
	steps := 1000
	length := 0.0
	tIncrement := 1.0 / float64(steps)

	for i := 0; i < steps; i++ {
		t1 := float64(i) * tIncrement
		t2 := float64(i+1) * tIncrement

		p1 := bezierPoint(t1, path)
		p2 := bezierPoint(t2, path)

		dx := p2.X - p1.X
		dy := p2.Y - p1.Y

		length += math.Sqrt(dx*dx + dy*dy)
	}

	return length
}

func bezierPoint(t float64, path [4]vec2) vec2 {
	u := 1 - t
	u2 := u * u
	u3 := u2 * u
	t2 := t * t
	t3 := t2 * t

	x := u3*path[0].X + 3*u2*t*path[1].X + 3*u*t2*path[2].X + t3*path[3].X
	y := u3*path[0].Y + 3*u2*t*path[1].Y + 3*u*t2*path[2].Y + t3*path[3].Y

	return vec2{X: x, Y: y}
}
