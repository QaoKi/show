package main

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y float64
}

// 计算两点之间的距离
func distance(p1, p2 Point) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// 计算三次贝塞尔曲线上点的坐标
func bezierPoint(p0, p1, p2, p3 Point, t float64) Point {
	x := math.Pow(1-t, 3)*p0.X + 3*math.Pow(1-t, 2)*t*p1.X + 3*(1-t)*t*t*p2.X + math.Pow(t, 3)*p3.X
	y := math.Pow(1-t, 3)*p0.Y + 3*math.Pow(1-t, 2)*t*p1.Y + 3*(1-t)*t*t*p2.Y + math.Pow(t, 3)*p3.Y
	return Point{X: x, Y: y}
}

// 计算三次贝塞尔曲线中三个线段的长度
func bezierSegmentLengths(p0, p1, p2, p3 Point, numSegments int) []float64 {
	segmentLengths := make([]float64, numSegments)

	for i := 0; i < numSegments; i++ {
		t0 := float64(i) / float64(numSegments)
		t1 := float64(i+1) / float64(numSegments)
		segmentLength := 0.0

		// 使用离散采样计算线段长度
		numSamples := 100
		for j := 0; j < numSamples; j++ {
			t := t0 + (float64(j)/float64(numSamples))*(t1-t0)
			pStart := bezierPoint(p0, p1, p2, p3, t)
			t = t0 + (float64(j+1)/float64(numSamples))*(t1-t0)
			pEnd := bezierPoint(p0, p1, p2, p3, t)
			segmentLength += distance(pStart, pEnd)
		}

		segmentLengths[i] = segmentLength
	}

	return segmentLengths
}

func length(path [4]vec2) {
	p0 := Point{X: path[0].X, Y: path[0].Y}
	p1 := Point{X: path[1].X, Y: path[1].Y}
	p2 := Point{X: path[2].X, Y: path[2].Y}
	p3 := Point{X: path[3].X, Y: path[3].Y}

	numSegments := 3
	segmentLengths := bezierSegmentLengths(p0, p1, p2, p3, numSegments)

	for i, length := range segmentLengths {
		fmt.Printf("线段 %d 的长度: %.4f\n", i+1, length)
	}
}
