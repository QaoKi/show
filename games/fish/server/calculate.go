package fish

import "bygame/common/utils"

// 从配置的数值上计算
func fishKilled(f *fish) bool {
	ret := false
	// 计算概率,先固定一个概率1/10, 命中 鱼的命中率 + 个人命中率 + 系统修正(杀控？活动)
	iskill := utils.Rate1w(2000)
	if iskill {
		ret = true
	}
	return ret
}
