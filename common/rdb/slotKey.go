package rdb

import "fmt"

//slot 系统库存 hash{lastTime:int,stop:bool,coin:int}
//lastTime 最后一次杀控结束时间
//stop 当前杀控是否暂停
//当获取到杀分状态为暂停的时候随机一个时间看到没到如果到了就是设置为true 和最后时间，这个方法被重入也没关系只有coin需要原子操作
func KeySlotRepertory(tid string) string {
	return fmt.Sprintf("%v:repertory:%v", preFixSlot, tid)
}

func KeySlotUser(mid string) string {
	return fmt.Sprintf("%v:user:%v", preFixSlot, mid)
}
