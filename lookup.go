package emozi

import (
	"math/rand"
	"strconv"
)

var 空 = "🈳️"

func (c *Coder) 查声母(s 声母枚举) string {
	lst := 声母[s]
	if len(lst) == 0 {
		return 空
	}
	if len(lst) == 1 || !c.isRandom {
		return lst[0]
	}
	return lst[rand.Intn(len(lst))]
}

func (c *Coder) 查韵母(y 韵母枚举) string {
	lst := 韵母[y]
	if len(lst) == 0 {
		return 空
	}
	if len(lst) == 1 || !c.isRandom {
		return lst[0]
	}
	return lst[rand.Intn(len(lst))]
}

func (c *Coder) 查声调(t 声调枚举) string {
	lst := 声调[t]
	if len(lst) == 0 {
		return 空
	}
	if len(lst) == 1 || !c.isRandom {
		return lst[0]
	}
	return lst[rand.Intn(len(lst))]
}

func (c *Coder) 查部首(r rune) string {
	x := &部首表{}
	err := c.db.Find(部首表名, x, "WHERE R="+strconv.Itoa(int(r)))
	if err == nil && len(x.E) > 0 && x.E != 空 {
		return x.E
	}
	if e, ok := 部首后备[r]; ok {
		return e
	}
	return 空
}
