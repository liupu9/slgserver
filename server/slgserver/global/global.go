package global

import "github.com/llr104/slgserver/config"

// 定义全局变量

var MapWith = 200
var MapHeight = 200

func ToPosition(x, y int) int {
	return x + MapHeight*y
}

func IsDev() bool {
	return config.File.MustBool("slgserver", "is_dev", false)
}
