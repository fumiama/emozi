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

// 从表 从部首表
type 部首表 struct {
	R rune   // R 该部首
	E string `db:"E,UNIQUE"` // E 该部首对应的颜文字
}

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

func (z *字表) String() string {
	sb := strings.Builder{}
	sb.WriteString("#")
	sb.WriteString(strconv.FormatInt(z.ID, 10))
	sb.WriteByte(' ')
	sb.WriteRune(z.W)
	sb.WriteString(" [")
	sb.WriteString(z.读音())
	sb.WriteString("] 从")
	sb.WriteRune(z.R)
	sb.WriteByte(' ')
	sb.WriteString(z.P)
	sb.WriteByte(' ')
	sb.WriteString(z.F)
	return sb.String()
}

func (z *字表) 读音() string {
	sb := strings.Builder{}
	sb.WriteString(z.S.String())
	sb.WriteString(", ")
	sb.WriteString(z.Y.String())
	sb.WriteString(", ")
	sb.WriteString(z.T.String())
	return sb.String()
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
	}
	err = c.db.FindFor(主字表名, &x, q, func() error {
		lstbuf = append(lstbuf, x)
		return nil
	})
	if len(lstbuf) == 0 {
		c.字表缓存[ch] = nil
		if err == nil {
			err = ErrNoSuchChar
		}
		return nil, lstbuf, err
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

func (c *Coder) 添加字到表(table, w, r, p, f string) (int64, string, error) {
	if p == "" {
		p = 去调(f)
	}
	s, y, t, rw, rr, err := 拆音识字(w, r, p, f)
	if err != nil {
		return 0, "", err
	}
	id := 字表ID(rw, s, y, t)
	revid := 逆字ID(s, y, t, rr)
	x := &字表{
		ID: id,
		W:  rw, S: s, Y: y, T: t,
		R: rr, P: p, F: f,
	}
	c.mu.Lock()
	err = c.db.InsertUnique(table, x)
	if err == nil {
		c.字表缓存[rw] = append(c.字表缓存[rw], *x)
		c.逆字表缓存[revid] = append(c.逆字表缓存[revid], rw)
	}
	c.mu.Unlock()
	if err != nil {
		return 0, "", errors.New("已有同音同形的字 '" + w + "'")
	}
	return id, x.读音(), nil
}

// AddChar 向主库添加一个新字
//
// w: 字, r: 部首, p: 不带声调的拼音(可空), f: 带声调的拼音
func (c *Coder) AddChar(w, r, p, f string) (int64, string, error) {
	return c.添加字到表(主字表名, w, r, p, f)
}

// AddCharOverlay 向附加库添加一个新字, 覆盖在主库之上
//
// w: 字, r: 部首, p: 不带声调的拼音(可空), f: 带声调的拼音
// 返回: 字表ID, 文字描述, error
func (c *Coder) AddCharOverlay(w, r, p, f string) (int64, string, error) {
	return c.添加字到表(附字表名, w, r, p, f)
}

// StabilizeCharFromOverlay 将附加库中的一项固定到主库
func (c *Coder) StabilizeCharFromOverlay(id int64) (string, error) {
	x := 字表{}
	q := "WHERE ID=" + strconv.FormatInt(id, 10)
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.Find(附字表名, &x, q)
	if err != nil {
		return "", err
	}
	err = c.db.Insert(主字表名, &x)
	if err != nil {
		return x.String(), err
	}
	return x.String(), c.db.Del(附字表名, q)
}

// 清除缓存字 未加锁, 必须在写锁内调用
func (c *Coder) 清除缓存字(x *字表) {
	for i, ch := range c.字表缓存[x.W] {
		if ch.ID == x.ID {
			switch {
			case i == 0:
				c.字表缓存[x.W] = c.字表缓存[x.W][1:]
			case i == len(c.字表缓存[x.W])-1:
				c.字表缓存[x.W] = c.字表缓存[x.W][:i-1]
			default:
				c.字表缓存[x.W] = append(c.字表缓存[x.W][:i], c.字表缓存[x.W][i+1:]...)
			}
			break
		}
	}
	revid := 逆字ID(x.S, x.Y, x.T, x.R)
	for i, ch := range c.逆字表缓存[revid] {
		if ch == x.W {
			switch {
			case i == 0:
				c.逆字表缓存[revid] = c.逆字表缓存[revid][1:]
			case i == len(c.逆字表缓存[revid])-1:
				c.逆字表缓存[revid] = c.逆字表缓存[revid][:i-1]
			default:
				c.逆字表缓存[revid] = append(c.逆字表缓存[revid][:i], c.逆字表缓存[revid][i+1:]...)
			}
			break
		}
	}
}

func (c *Coder) 删除表中字(table string, id int64) error {
	q := "WHERE ID=" + strconv.FormatInt(id, 10)
	x := 字表{}
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.Find(table, &x, q)
	if err != nil {
		return err
	}
	c.清除缓存字(&x)
	return c.db.Del(table, q)
}

// DelChar 删除主库的一个字
func (c *Coder) DelChar(id int64) error {
	return c.删除表中字(主字表名, id)
}

// DelCharOverlay 删除附加库的一个字
func (c *Coder) DelCharOverlay(id int64) error {
	return c.删除表中字(附字表名, id)
}

// AddRadical 添加一个部首
func (c *Coder) AddRadical(r rune, e string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.InsertUnique(部首表名, &部首表{R: r, E: e})
	if err == nil {
		c.部首缓存[r] = e
		if 无此字符(c.逆部首缓存[e], r) {
			c.逆部首缓存[e] = append(c.逆部首缓存[e], r)
		}
	}
	return err
}

// DelRadical 删除一个部首
func (c *Coder) DelRadical(r rune) error {
	x := 部首表{}
	q := "WHERE R=" + strconv.Itoa(int(r))
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.Find(部首表名, &x, q)
	if err != nil {
		return err
	}
	delete(c.部首缓存, r)
	for i, item := range c.逆部首缓存[x.E] {
		if item == r {
			switch {
			case i == 0:
				c.逆部首缓存[x.E] = c.逆部首缓存[x.E][1:]
			case i == len(c.逆部首缓存[x.E])-1:
				c.逆部首缓存[x.E] = c.逆部首缓存[x.E][:i-1]
			default:
				c.逆部首缓存[x.E] = append(c.逆部首缓存[x.E][:i], c.逆部首缓存[x.E][i+1:]...)
			}
			break
		}
	}
	return c.db.Del(部首表名, q)
}
