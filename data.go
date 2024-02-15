package emozi

import (
	_ "embed"
	"errors"
	"strconv"
	"strings"
)

// 字数据库 数据来自 https://github.com/shuowenjiezi/shuowen
//
//go:embed 字.db
var 字数据库 []byte

// DatabasePath 字数据库的路径 如找不到会向对应路径写入内嵌的字数据库
var EmoziDatabasePath = "字.db"

const (
	主字表名 = "emozi"
	附字表名 = "altzi"
	部首表名 = "radcl"
)

var (
	ErrNoSuchChar = errors.New("no such char")
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

// CharGlobalID 计算全局唯一字表ID
func CharGlobalID(w rune, f string) (int64, error) {
	p := 去调(f)
	s, y, err := 拆音(p)
	if err != nil {
		return 0, err
	}
	t := 识调(f)
	return 字表ID(w, s, y, t), nil
}

func 字表ID(w rune, s 声母枚举, y 韵母枚举, t 声调枚举) int64 {
	return int64((uint64(w) << 32) | (uint64(s) << 16) | (uint64(y) << 8) | (uint64(t)))
}

// 逆字ID 同声母 韵母 声调 部首的字的集合
func 逆字ID(s 声母枚举, y 韵母枚举, t 声调枚举, r rune) int64 {
	return int64((uint64(r) << 32) | (uint64(s) << 16) | (uint64(y) << 8) | (uint64(t)))
}

// 查字 返回 lst lstbuf error
func (c *Coder) 查字(ch rune, lstbuf []字表) ([]字表, []字表, error) {
	c.mu.RLock()
	lst, ok := c.字表缓存[ch]
	c.mu.RUnlock()
	if ok {
		if len(lst) == 0 {
			return nil, lstbuf, ErrNoSuchChar
		}
		return lst, lstbuf, nil
	}
	lstbuf = lstbuf[:0]
	x := 字表{}
	q := "WHERE W=" + strconv.Itoa(int(ch))
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.FindFor(附字表名, &x, q, func() error {
		lstbuf = append(lstbuf, x)
		return nil
	})
	if err != nil {
		lstbuf = lstbuf[:0]
		err = c.db.FindFor(主字表名, &x, q, func() error {
			lstbuf = append(lstbuf, x)
			return nil
		})
	}
	if err != nil {
		c.字表缓存[ch] = nil
		return nil, lstbuf, err
	}
	if len(lstbuf) == 0 {
		c.字表缓存[ch] = nil
		return nil, lstbuf, ErrNoSuchChar
	}
	lstsave := make([]字表, len(lstbuf))
	copy(lstsave, lstbuf)
	c.字表缓存[ch] = lstsave
	return lstbuf, lstbuf, nil
}

// 逆字 逆查匹配的字
func (c *Coder) 逆字(s 声母枚举, y 韵母枚举, t 声调枚举, r rune, lstbuf []字表) ([]rune, []字表, error) {
	id := 逆字ID(s, y, t, r)
	c.mu.RLock()
	matches, ok := c.逆字表缓存[id]
	c.mu.RUnlock()
	if ok {
		if len(matches) == 0 {
			return nil, lstbuf, ErrNoSuchChar
		}
		return matches, lstbuf, nil
	}
	lstbuf = lstbuf[:0]
	x := 字表{}
	sb := strings.Builder{}
	sb.WriteString("WHERE S=")
	sb.WriteString(strconv.Itoa(int(s)))
	sb.WriteString(" AND Y=")
	sb.WriteString(strconv.Itoa(int(y)))
	sb.WriteString(" AND T=")
	sb.WriteString(strconv.Itoa(int(t)))
	if r != 0 {
		sb.WriteString(" AND R=")
		sb.WriteString(strconv.Itoa(int(r)))
	}
	q := sb.String()
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.FindFor(附字表名, &x, q, func() error {
		lstbuf = append(lstbuf, x)
		return nil
	})
	if err != nil {
		lstbuf = lstbuf[:0]
		err = c.db.FindFor(主字表名, &x, q, func() error {
			lstbuf = append(lstbuf, x)
			return nil
		})
	}
	if err != nil {
		c.逆字表缓存[id] = nil
		return nil, lstbuf, err
	}
	if len(lstbuf) == 0 {
		c.逆字表缓存[id] = nil
		return nil, lstbuf, ErrNoSuchChar
	}
	rs := make([]rune, len(lstbuf))
	for i, x := range lstbuf {
		rs[i] = x.W
	}
	c.逆字表缓存[id] = rs
	return rs, lstbuf, nil
}

// 从表 从部首表
type 部首表 struct {
	R rune   // R 该部首
	E string `db:"E,UNIQUE"` // E 该部首对应的颜文字
}
