package main

import (
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func gen(path [4]vec2) {

	p := plot.New()

	p0 := path[0]
	p1 := path[1]
	p2 := path[2]
	p3 := path[3]

	// 计算贝塞尔曲线上的点
	numPoints := 1000
	var points plotter.XYs
	for i := 0; i <= numPoints; i++ {
		t := float64(i) / float64(numPoints)
		x := bezier(p0.X, p1.X, p2.X, p3.X, t)
		y := bezier(p0.Y, p1.Y, p2.Y, p3.Y, t)
		points = append(points, plotter.XY{X: x, Y: y})
	}

	// 创建曲线图形
	line, err := plotter.NewLine(points)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // 设置曲线颜色

	// 将曲线添加到图表中
	p.Add(line)

	// 设置图表标题和坐标轴标签
	p.Title.Text = "贝塞尔曲线示例"
	p.X.Label.Text = "X轴"
	p.Y.Label.Text = "Y轴"

	// 保存图表到文件
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "bezier_curve.png"); err != nil {
		panic(err)
	}
}

// 计算贝塞尔曲线上的点
func bezier(p0, p1, p2, p3, t float64) float64 {
	b0 := math.Pow(1-t, 3)
	b1 := 3 * t * math.Pow(1-t, 2)
	b2 := 3 * math.Pow(t, 2) * (1 - t)
	b3 := math.Pow(t, 3)
	return b0*p0 + b1*p1 + b2*p2 + b3*p3
}
