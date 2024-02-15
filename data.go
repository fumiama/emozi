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
	sb.WriteString(z.S.String())
	sb.WriteString(z.Y.String())
	sb.WriteString(z.T.String())
	sb.WriteString("] 从")
	sb.WriteRune(z.R)
	sb.WriteByte(' ')
	sb.WriteString(z.P)
	sb.WriteByte(' ')
	sb.WriteString(z.F)
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

// AddChar 向主库添加一个新字
//
// w: 字, r: 部首, p: 不带声调的拼音(可空), f: 带声调的拼音
func (c *Coder) AddChar(w, r, p, f string) error {
	if p == "" {
		p = 去调(f)
	}
	s, y, t, rw, rr, err := 拆音识字(w, r, p, f)
	if err != nil {
		return err
	}
	c.mu.Lock()
	err = c.db.InsertUnique(主字表名, &字表{
		ID: 字表ID(rw, s, y, t),
		W:  rw, S: s, Y: y, T: t,
		R: rr, P: p, F: f,
	})
	c.mu.Unlock()
	if err != nil {
		return errors.New("已有同音同形的字 '" + w + "'")
	}
	return nil
}

// AddCharOverlay 向附加库添加一个新字, 覆盖在主库之上
//
// w: 字, r: 部首, p: 不带声调的拼音(可空), f: 带声调的拼音
// 返回: 字表ID, 文字描述, error
func (c *Coder) AddCharOverlay(w, r, p, f string) (int64, string, error) {
	if p == "" {
		p = 去调(f)
	}
	s, y, t, rw, rr, err := 拆音识字(w, r, p, f)
	if err != nil {
		return 0, "", err
	}
	return c.addcharoverlay(w, p, f, s, y, t, rw, rr)
}

func (c *Coder) addcharoverlay(w, p, f string, s 声母枚举, y 韵母枚举, t 声调枚举, rw rune, rr rune) (int64, string, error) {
	id := 字表ID(rw, s, y, t)
	c.mu.Lock()
	err := c.db.InsertUnique(附字表名, &字表{
		ID: id,
		W:  rw, S: s, Y: y, T: t,
		R: rr, P: p, F: f,
	})
	c.mu.Unlock()
	if err != nil {
		return 0, "", errors.New("已有同音同形的字 '" + w + "'")
	}
	sb := strings.Builder{}
	sb.WriteString(s.String())
	sb.WriteString(", ")
	sb.WriteString(y.String())
	sb.WriteString(", ")
	sb.WriteString(t.String())
	return id, sb.String(), nil
}

// ChangeCharOverlay 更改附加库的一项
func (c *Coder) ChangeCharOverlay(oldw, oldr, oldf, neww, newr, newf string) (int64, string, error) {
	s, y, t, rw, rr, err := 拆音识字(oldw, oldr, 去调(oldf), oldf)
	if err != nil {
		return 0, "", err
	}
	newp := 去调(newf)
	ns, ny, nt, nrw, nrr, err := 拆音识字(neww, newr, newp, newf)
	if err != nil {
		return 0, "", err
	}
	q := "WHERE ID=" + strconv.FormatInt(字表ID(rw, s, y, t), 10)
	x := 字表{}
	c.mu.RLock()
	err = c.db.Find(附字表名, &x, q)
	c.mu.RUnlock()
	if err != nil {
		return 0, "", err
	}
	if x.R != rr {
		return 0, "", errors.New("提供的旧部首 '" + string(rr) + "' 与记载的 '" + string(x.R) + "' 不符")
	}
	c.mu.Lock()
	err = c.db.Del(附字表名, q)
	c.mu.Unlock()
	if err != nil {
		return 0, "", err
	}
	return c.addcharoverlay(neww, newp, newf, ns, ny, nt, nrw, nrr)
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

// DelChar 删除主库的一个字
func (c *Coder) DelChar(id int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Del(主字表名, "WHERE ID="+strconv.FormatInt(id, 10))
}

// DelCharOverlay 删除附加库的一个字
func (c *Coder) DelCharOverlay(id int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Del(附字表名, "WHERE ID="+strconv.FormatInt(id, 10))
}

// AddRadical 添加一个部首
func (c *Coder) AddRadical(r rune, e string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.InsertUnique(部首表名, &部首表{R: r, E: e})
}

// DelRadical 删除一个部首
func (c *Coder) DelRadical(r rune) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Del(部首表名, "WHERE R="+strconv.Itoa(int(r)))
}
