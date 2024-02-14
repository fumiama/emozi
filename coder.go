package emozi

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// Coder encoder/decoder
type Coder struct {
	mu       sync.RWMutex
	db       sql.Sqlite
	字表缓存     map[rune][]字表
	部首缓存     map[rune]string
	isRandom bool
}

// NewCoder israndom 随机挑选声母韵母的颜文字, 否则固定使用第一个
func NewCoder(israndom bool, cachettl time.Duration) (c Coder, err error) {
	if _, err = os.Stat(EmoziDatabasePath); err != nil {
		err = os.WriteFile(EmoziDatabasePath, 字数据库, 0644)
		if err != nil {
			return
		}
	}
	c.db.DBPath = EmoziDatabasePath
	c.字表缓存 = make(map[rune][]字表, 4096)
	c.部首缓存 = make(map[rune]string, 4096)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Close()
}

// Encode 从汉字序列生成 EmoziString 返回 EmoziString 多音字选择数列表
func (c *Coder) Encode(s string, selections ...int) (EmoziString, []int, error) {
	sb := strings.Builder{}
	lstbuf := make([]字表, 0, len(s)/2)
	var lst []字表
	write := func(x *字表) {
		sb.WriteString(c.查声母(x.S))
		sb.WriteString(c.查韵母(x.Y))
		sb.WriteString(c.查声调(x.T))
		sb.WriteString(c.查部首(x.R))
	}
	多音字计数 := 0
	多音字数表 := []int{}
	var err error
	for _, ch := range s { // nolint: go-staticcheck
		lst, lstbuf, err = c.查字(ch, lstbuf)
		if err != nil || len(lst) == 0 {
			sb.WriteRune(ch)
			continue
		}
		if len(lst) == 1 {
			write(&lst[0])
			continue
		}
		多音字数表 = append(多音字数表, len(lst))
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
	return WrapRawEmoziString(sb.String()), 多音字数表, nil
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
func (c *Coder) AddCharOverlay(w, r, p, f string) error {
	if p == "" {
		p = 去调(f)
	}
	s, y, t, rw, rr, err := 拆音识字(w, r, p, f)
	if err != nil {
		return err
	}
	return c.addcharoverlay(w, p, f, s, y, t, rw, rr)
}

func (c *Coder) addcharoverlay(w, p, f string, s 声母枚举, y 韵母枚举, t 声调枚举, rw rune, rr rune) error {
	c.mu.Lock()
	err := c.db.InsertUnique(附字表名, &字表{
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

// ChangeCharOverlay 更改附加库的一项
func (c *Coder) ChangeCharOverlay(oldw, oldr, oldf, neww, newr, newf string) error {
	s, y, t, rw, rr, err := 拆音识字(oldw, oldr, 去调(oldf), oldf)
	if err != nil {
		return err
	}
	newp := 去调(newf)
	ns, ny, nt, nrw, nrr, err := 拆音识字(neww, newr, newp, newf)
	if err != nil {
		return err
	}
	q := "WHERE ID=" + strconv.FormatInt(字表ID(rw, s, y, t), 10)
	x := 字表{}
	c.mu.RLock()
	err = c.db.Find(附字表名, &x, q)
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	if x.R != rr {
		return errors.New("提供的旧部首 '" + string(rr) + "' 与记载的 '" + string(x.R) + "' 不符")
	}
	c.mu.Lock()
	err = c.db.Del(附字表名, q)
	c.mu.Unlock()
	if err != nil {
		return err
	}
	return c.addcharoverlay(neww, newp, newf, ns, ny, nt, nrw, nrr)
}

// StabilizeCharFromOverlay 将附加库中的一项固定到主库
func (c *Coder) StabilizeCharFromOverlay(id int64) error {
	x := 字表{}
	q := "WHERE ID=" + strconv.FormatInt(id, 10)
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.Find(附字表名, &x, q)
	if err != nil {
		return err
	}
	err = c.db.Insert(主字表名, &x)
	if err != nil {
		return err
	}
	return c.db.Del(附字表名, q)
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

// AddRadicalOverlay 添加一个部首
func (c *Coder) AddRadicalOverlay(r rune, e string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.InsertUnique(部首表名, &部首表{R: r, E: e})
}

// DelRadicalOverlay 删除一个部首
func (c *Coder) DelRadicalOverlay(r rune) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Del(部首表名, "WHERE R="+strconv.Itoa(int(r)))
}
