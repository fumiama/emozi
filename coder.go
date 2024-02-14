package emozi

import (
	"errors"
	"strconv"
	"strings"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// Coder encoder/decoder
type Coder struct {
	db       sql.Sqlite
	isRandom bool
}

// NewCoder israndom 随机挑选声母韵母的颜文字, 否则固定使用第一个
func NewCoder(israndom bool, cachettl time.Duration) (c Coder, err error) {
	c.db.DBPath = EmoziDatabasePath
	c.isRandom = israndom
	err = c.db.Open(cachettl)
	if err != nil {
		return
	}
	err = c.db.Create(主字表名, &字表{})
	if err != nil {
		return
	}
	err = c.db.Create(附字表名, &字表{})
	if err != nil {
		return
	}
	err = c.db.Create(部首表名, &部首表{})
	if err != nil {
		return
	}
	_ = c.db.Query("CREATE INDEX IF NOT EXISTS IE ON "+部首表名+" (E);", nil)
	return
}

// Close ...
func (c *Coder) Close() error {
	return c.db.Close()
}

// Encode 从汉字序列生成 EmoziString
func (c *Coder) Encode(s string, selections ...int) (EmoziString, error) {
	sb := strings.Builder{}
	x := &字表{}
	lst := []字表{}
	write := func(x *字表) {
		sb.WriteString(c.查声母(x.S))
		sb.WriteString(c.查韵母(x.Y))
		sb.WriteString(c.查声调(x.T))
		sb.WriteString(c.查部首(x.R))
	}
	多音字计数 := 0
	for _, ch := range s { // nolint: go-staticcheck
		lst = lst[:0]
		err := c.db.FindFor(附字表名, x, "WHERE W="+strconv.Itoa(int(ch)), func() error {
			lst = append(lst, *x)
			return nil
		})
		if err != nil {
			lst = lst[:0]
			err = c.db.FindFor(主字表名, x, "WHERE W="+strconv.Itoa(int(ch)), func() error {
				lst = append(lst, *x)
				return nil
			})
		}
		if err != nil || len(lst) == 0 {
			sb.WriteRune(ch)
			continue
		}
		if len(lst) == 1 {
			write(x)
			continue
		}
		if len(selections) > 多音字计数 {
			idx := selections[多音字计数]
			多音字计数++
			if idx >= 0 && idx < len(lst) {
				write(&lst[idx])
				continue
			}
		}
		// 多音字
		sb.WriteString("[")
		write(&lst[0])
		for _, x := range lst[1:] {
			sb.WriteString("|")
			write(&x)
		}
		sb.WriteString("]")
	}
	return WrapRawEmoziString(sb.String()), nil
}

// Add 向主库添加一个新字
//
// w: 字, r: 部首, p: 不带声调的拼音(可空), f: 带声调的拼音
func (c *Coder) Add(w, r, p, f string) error {
	if p == "" {
		p = 去调(f)
	}
	s, y, t, rw, rr, err := 拆音识字(w, r, p, f)
	if err != nil {
		return err
	}
	err = c.db.InsertUnique(主字表名, &字表{
		ID: 颜表ID(rw, s, y, t),
		W:  rw, S: s, Y: y, T: t,
		R: rr, P: p, F: f,
	})
	if err != nil {
		return errors.New("已有同音同形的字 '" + w + "'")
	}
	return nil
}

// Overlay 向附加库添加一个新字, 覆盖在主库之上
//
// w: 字, r: 部首, p: 不带声调的拼音(可空), f: 带声调的拼音
func (c *Coder) Overlay(w, r, p, f string) error {
	if p == "" {
		p = 去调(f)
	}
	s, y, t, rw, rr, err := 拆音识字(w, r, p, f)
	if err != nil {
		return err
	}
	return c.overlay(w, p, f, s, y, t, rw, rr)
}

func (c *Coder) overlay(w, p, f string, s 声母枚举, y 韵母枚举, t 声调枚举, rw rune, rr rune) error {
	err := c.db.InsertUnique(附字表名, &字表{
		ID: 颜表ID(rw, s, y, t),
		W:  rw, S: s, Y: y, T: t,
		R: rr, P: p, F: f,
	})
	if err != nil {
		return errors.New("已有同音同形的字 '" + w + "'")
	}
	return nil
}

// ChangeOverlay 更改附加库的一项
func (c *Coder) ChangeOverlay(oldw, oldr, oldf, neww, newr, newf string) error {
	s, y, t, rw, rr, err := 拆音识字(oldw, oldr, 去调(oldf), oldf)
	if err != nil {
		return err
	}
	newp := 去调(newf)
	ns, ny, nt, nrw, nrr, err := 拆音识字(neww, newr, newp, newf)
	if err != nil {
		return err
	}
	q := "WHERE ID=" + strconv.FormatInt(颜表ID(rw, s, y, t), 10)
	x := 字表{}
	err = c.db.Find(附字表名, &x, q)
	if err != nil {
		return err
	}
	if x.R != rr {
		return errors.New("提供的旧部首 '" + string(rr) + "' 与记载的 '" + string(x.R) + "' 不符")
	}
	err = c.db.Del(附字表名, q)
	if err != nil {
		return err
	}
	return c.overlay(neww, newp, newf, ns, ny, nt, nrw, nrr)
}

// OverlayRadical 添加一个部首
func (c *Coder) OverlayRadical(r rune, e string) error {
	return c.db.InsertUnique(部首表名, &部首表{R: r, E: e})
}
