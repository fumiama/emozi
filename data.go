package emozi

import _ "embed"

// 字数据库 数据来自 https://github.com/shuowenjiezi/shuowen
//
//var 字数据库 []byte

// DatabasePath 字数据库的路径 如找不到会向对应路径写入内嵌的字数据库
var EmoziDatabasePath = "字.db"

const (
	主字表名 = "emozi"
	附字表名 = "altzi"
	部首表名 = "radcl"
)

// 字表 emozi表 定义
type 字表 struct {
	ID int64 // ID 高 32 位 W 的 rune, 低 32 位 保留8 S8 Y8 T8
	W  rune
	S  声母枚举
	Y  韵母枚举
	T  声调枚举
	R  rune
	P  string
	F  string
}

func 颜表ID(w rune, s 声母枚举, y 韵母枚举, t 声调枚举) int64 {
	return int64((uint64(w) << 32) | (uint64(s) << 16) | (uint64(y) << 8) | (uint64(t)))
}

// 从表 从部首表
type 部首表 struct {
	R rune   // R 该部首
	E string `db:"E,UNIQUE"` // E 该部首对应的颜文字
}
