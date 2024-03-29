package emozi

import (
	"math/rand"
	"strconv"
	"strings"
)

// EmptyMark 代表 "空" 的颜文字
const EmptyMark = '🈳'

// LookupChar 查一个汉字 (可能是多音字)
func (c *Coder) LookupChar(ch rune) (explains []string, err error) {
	lst, _, err := c.查字(ch, make([]字表, 0, 8))
	if err != nil || len(lst) == 0 {
		return
	}
	explains = make([]string, len(lst))
	for i, x := range lst {
		explains[i] = x.String()
	}
	return
}

// LookupRadical 查一个部首
func (c *Coder) LookupRadical(r rune) string {
	return c.部首(r)
}

// GetCharByID ...
func (c *Coder) GetCharByID(id int64) (w, r rune, p, f, desc string, err error) {
	x := 字表{}
	q := "WHERE ID=" + strconv.FormatInt(id, 10)
	c.mu.RLock()
	err = c.db.Find(附字表名, &x, q)
	if err != nil {
		err = c.db.Find(主字表名, &x, q)
	}
	c.mu.RUnlock()
	if err != nil {
		return
	}
	w = x.W
	r = x.R
	p = x.P
	f = x.F
	desc = x.String()
	return
}

func 随机正查(m [][]string, isRandom bool, i uint8) string {
	lst := m[i]
	if len(lst) == 0 {
		return string(EmptyMark)
	}
	if len(lst) == 1 || !isRandom {
		return lst[0]
	}
	return lst[rand.Intn(len(lst))]
}

func (c *Coder) 声母(isRandom bool, s 声母枚举) string {
	return 随机正查(声母[:], isRandom, uint8(s))
}

func (c *Coder) 韵母(isRandom bool, y 韵母枚举) string {
	return 随机正查(韵母[:], isRandom, uint8(y))
}

func (c *Coder) 声调(isRandom bool, t 声调枚举) string {
	return 随机正查(声调[:], isRandom, uint8(t))
}

func (c *Coder) 部首(r rune) string {
	c.mu.RLock()
	e, ok := c.部首缓存[r]
	c.mu.RUnlock()
	if ok {
		return e
	}
	x := &部首表{}
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.Find(部首表名, x, "WHERE R="+strconv.Itoa(int(r)))
	if err == nil && len(x.E) > 0 && x.E != string(EmptyMark) {
		c.部首缓存[r] = x.E
		return x.E
	}
	if e, ok := 部首后备[r]; ok {
		c.部首缓存[r] = e
		return e
	}
	c.部首缓存[r] = string(EmptyMark)
	return string(EmptyMark)
}

func 二阶逆查[E ~uint8](lowm map[rune][]string, m map[string]E, s string) (enum E, n int) {
	lowk := rune(0)
	lows := s
	if len(lows) > 12 {
		lows = lows[:12]
	}
	r := []rune(lows)
	if len(r) == 0 {
		return
	}
	lowk = r[0]
	ks := lowm[lowk]
	if len(ks) == 0 {
		return
	}
	// 寻找最长匹配 T
	matchp := -1
	matchl := 0
	for i, k := range ks {
		if strings.HasPrefix(s, k) {
			if len(k) > matchl {
				matchl = len(k)
				matchp = i
			}
		}
	}
	if matchp < 0 {
		return
	}
	enum, ok := m[ks[matchp]]
	if !ok {
		return
	}
	n = matchl
	return
}

func (c *Coder) 逆声母(s string) (声母枚举, int) {
	return 二阶逆查[声母枚举](低阶逆声母, 逆声母, s)
}

func (c *Coder) 逆韵母(s string) (韵母枚举, int) {
	return 二阶逆查[韵母枚举](低阶逆韵母, 逆韵母, s)
}
func (c *Coder) 逆声调(s string) (声调枚举, int) {
	return 二阶逆查[声调枚举](低阶逆声调, 逆声调, s)
}

func (c *Coder) 逆部首(s string) (rs []rune, n int) {
	lim := len(s)
	if lim > 32 {
		lim = 32
	}
	// fmt.Println("逆部首: recv", s, "len", len(s), "lim", lim)
	c.mu.RLock()
	for i := 1; i <= lim; i++ {
		l := c.逆部首缓存[s[:i]]
		if len(l) > 0 {
			rs = l
			n = i
		}
	}
	c.mu.RUnlock()
	if n > 0 && len(rs) > 0 {
		return
	}
	x := &部首表{}
	sb := strings.Builder{}
	sb.WriteString("WHERE ")
	for i := 1; i <= lim; i++ {
		sb.WriteString("E='")
		sb.WriteString(s[:i])
		sb.WriteString("' OR ")
	}
	q := sb.String()[:sb.Len()-4]
	n = 0
	e := ""
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.db.FindFor(部首表名, x, q, func() error {
		if len(x.E) > n {
			n = len(x.E)
			rs = rs[:0]
			e = x.E
		}
		if len(x.E) == n && 无此字符(rs, x.R) {
			rs = append(rs, x.R)
		}
		return nil
	})
	if err == nil && len(rs) > 0 && n > 0 {
		c.逆部首缓存[e] = rs
		return
	}
	for i := 1; i <= lim; i++ {
		k := s[:i]
		innerrs, ok := 逆部首后备[k]
		c.逆部首缓存[k] = innerrs
		// fmt.Println(k, innerrs)
		if ok && len(innerrs) > 0 {
			n = i
			rs = innerrs
		}
	}
	return
}

func 无此字符(runes []rune, ch rune) bool {
	for _, r := range runes {
		if ch == r {
			return false
		}
	}
	return true
}
