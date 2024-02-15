package emozi

import (
	"os"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// Coder encoder/decoder
type Coder struct {
	mu    sync.RWMutex
	db    sql.Sqlite
	字表缓存  map[rune][]字表
	逆字表缓存 map[int64][]rune
	部首缓存  map[rune]string
	逆部首缓存 map[string][]rune
}

// NewCoder israndom 随机挑选声母韵母的颜文字, 否则固定使用第一个
func NewCoder(cachettl time.Duration) (c Coder, err error) {
	if _, err = os.Stat(EmoziDatabasePath); err != nil {
		err = os.WriteFile(EmoziDatabasePath, 字数据库, 0644)
		if err != nil {
			return
		}
	}
	c.db.DBPath = EmoziDatabasePath
	c.字表缓存 = make(map[rune][]字表, 4096)
	c.逆字表缓存 = make(map[int64][]rune, 4096)
	c.部首缓存 = make(map[rune]string, 4096)
	c.逆部首缓存 = make(map[string][]rune, 4096)
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
	c.字表缓存 = nil
	c.逆字表缓存 = nil
	c.部首缓存 = nil
	c.逆部首缓存 = nil
	return c.db.Close()
}

// Encode 从汉字序列生成 EmoziString 返回 EmoziString 多音字选择数列表
func (c *Coder) Encode(enableRandom bool, s string, selections ...int) (EmoziString, []int, error) {
	sb := strings.Builder{}
	buflen := len(s) / 2
	if buflen < 4 {
		buflen = 4
	}
	lstbuf := make([]字表, 0, buflen)
	var lst []字表
	var write func(x *字表)
	randomwrite := func(x *字表) {
		sb.WriteString(c.声母(enableRandom, x.S))
		sb.WriteString(c.韵母(enableRandom, x.Y))
		sb.WriteString(c.声调(enableRandom, x.T))
		sb.WriteString(c.部首(x.R))
	}
	norandomwrite := func(x *字表) {
		sb.WriteString(c.声母(false, x.S))
		sb.WriteString(c.韵母(false, x.Y))
		sb.WriteString(c.声调(false, x.T))
		sb.WriteString(c.部首(x.R))
		write = randomwrite
	}
	write = norandomwrite
	多音字计数 := 0
	多音字数表 := []int{}
	var err error
	for _, ch := range s { // nolint: go-staticcheck
		lst, lstbuf, err = c.查字(ch, lstbuf)
		if err != nil || len(lst) == 0 {
			//fmt.Println("写入未知字:", string(ch), ch)
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

// Decode 从 EmoziString 解码得到可能的文字序列
func (c *Coder) Decode(es EmoziString, forcedecode bool) (string, error) {
	if !es.IsValid() && !forcedecode {
		return "", ErrInvalidEmoziString
	}
	s := ""
	if forcedecode {
		s = string(es)
	} else {
		s = es.String()
	}
	// fmt.Println(len(s), s)
	lstbuf := make([]字表, 0, len(s)/8)
	read := func(s string) (string, int) {
		sum := 0
		sm, n := c.逆声母(s)
		if n == 0 {
			return "", 0
		}
		sum += n
		// fmt.Println(n, sm, s[0:sum])
		ym, n := c.逆韵母(s[sum:])
		if n == 0 {
			return "", 0
		}
		sum += n
		// fmt.Println(n, ym, s[sum-n:sum])
		t, n := c.逆声调(s[sum:])
		if n == 0 {
			return "", 0
		}
		sum += n
		// fmt.Println(n, t, s[sum-n:sum])
		rs, n := c.逆部首(s[sum:])
		if n == 0 {
			return "", 0
		}
		sum += n
		// fmt.Println(n, rs, s[sum-n:sum])
		var possibles []rune
		var err error
		if len(rs) == 0 { // 意符为空
			possibles, lstbuf, err = c.逆字(sm, ym, t, 0, lstbuf)
			if err != nil {
				return "[]", sum
			}
		} else {
			var revr []rune
			for i := 0; i < len(rs); i++ {
				revr, lstbuf, err = c.逆字(sm, ym, t, rs[i], lstbuf)
				if err != nil || len(revr) == 0 {
					continue
				}
				if len(possibles) == 0 {
					possibles = revr
				} else {
					possibles = append(possibles, revr...)
				}
			}
		}
		if len(possibles) == 0 {
			return "[]", sum
		}
		if len(possibles) == 1 {
			return string(possibles[0]), sum
		}
		sb := strings.Builder{}
		sb.WriteString("[")
		sb.WriteRune(possibles[0])
		for _, r := range possibles[1:] {
			sb.WriteString("|")
			sb.WriteRune(r)
		}
		sb.WriteString("]")
		return sb.String(), sum
	}
	sb := strings.Builder{}
	sum := 0
	for sum < len(s) {
		ch, n := read(s[sum:])
		if n <= 0 {
			sb.WriteByte(s[sum])
			sum++
			continue
		}
		sum += n
		sb.WriteString(ch)
	}
	return sb.String(), nil
}
