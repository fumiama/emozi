package emozi

import (
	"errors"
	"strings"
)

var smap = map[string]声母枚举{
	"b": 声母b,
	"p": 声母p,
	"m": 声母m,
	"f": 声母f,
	"d": 声母d,
	"t": 声母t,
	"n": 声母n,
	"l": 声母l,
	"g": 声母g,
	"k": 声母k,
	"h": 声母h,
	"j": 声母j,
	"q": 声母q,
	"x": 声母x,
	"r": 声母r,
	"w": 声母w,
}

const 双字声母 = "zcs"

var amap = map[string]韵母枚举{
	"a":   韵母a,
	"ai":  韵母ai,
	"ao":  韵母ao,
	"an":  韵母an,
	"ang": 韵母ang,
}

var wamap = map[string]韵母枚举{
	"a":   韵母ua,
	"ai":  韵母uai,
	"an":  韵母uan,
	"ang": 韵母uang,
}

var omap = map[string]韵母枚举{
	"o":   韵母o,
	"ou":  韵母ou,
	"ong": 韵母ong,
}

var emap = map[string]韵母枚举{
	"e":   韵母e,
	"er":  韵母er,
	"ei":  韵母ei,
	"en":  韵母en,
	"eng": 韵母eng,
}

var wemap = map[string]韵母枚举{
	"ei":  韵母ei,
	"en":  韵母en,
	"eng": 韵母ueng,
}

var imap = map[string]韵母枚举{
	"i":    韵母yi,
	"iao":  韵母iao,
	"iu":   韵母iu,
	"ia":   韵母ia,
	"ie":   韵母ie,
	"ian":  韵母ian,
	"in":   韵母in,
	"iang": 韵母iang,
	"ing":  韵母ing,
	"iong": 韵母iong,
}

var umap = map[string]韵母枚举{
	"u":    韵母wu,
	"uai":  韵母uai,
	"ui":   韵母ui,
	"ua":   韵母ua,
	"uo":   韵母uo,
	"uan":  韵母uan,
	"un":   韵母un,
	"uang": 韵母uang,
	"ueng": 韵母ueng,
}

var yumap = map[string]韵母枚举{
	"u":   韵母yu,
	"ue":  韵母yue,
	"uan": 韵母yuan,
	"un":  韵母yun,
	"ü":   韵母yu,
	"üe":  韵母yue,
	"üan": 韵母yuan,
	"ün":  韵母yun,
	"v":   韵母yu,
	"ve":  韵母yue,
	"van": 韵母yuan,
	"vn":  韵母yun,
}

func combine(maps ...map[string]韵母枚举) map[string]韵母枚举 {
	newmap := make(map[string]韵母枚举, 128)
	for _, m := range maps {
		for k, v := range m {
			newmap[k] = v
		}
	}
	return newmap
}

var aoeiu = combine(amap, omap, emap, imap, umap)

var aoeu = combine(amap, omap, emap, umap)

const (
	阴平字母 = "āōēīūǖ"
	阳平字母 = "áóéíúǘń"
	上声字母 = "ǎǒěǐǔǚň"
	去声字母 = "àòèìùǜ"
	G    = 'ɡ'
	A    = 'ɑ'
)

var notonemap = map[rune]string{
	'ā': "a", 'á': "a", 'ǎ': "a", 'à': "a",
	'ō': "o", 'ó': "o", 'ǒ': "o", 'ò': "o",
	'ē': "e", 'é': "e", 'ě': "e", 'è': "e",
	'ī': "i", 'í': "i", 'ǐ': "i", 'ì': "i",
	'ū': "u", 'ú': "u", 'ǔ': "u", 'ù': "u",
	'ǖ': "ü", 'ǘ': "ü", 'ǚ': "ü", 'ǜ': "ü",
	G: "g", A: "a", 'ń': "n", 'ň': "n",
}

func 去调(pf string) string {
	sb := strings.Builder{}
	for _, c := range pf {
		if nc, ok := notonemap[c]; ok {
			sb.WriteString(nc)
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// 识调 从拼音获得声调
func 识调(pf string) 声调枚举 {
	switch {
	case strings.ContainsAny(pf, 阴平字母):
		return 阴平
	case strings.ContainsAny(pf, 阳平字母):
		return 阳平
	case strings.ContainsAny(pf, 上声字母):
		return 上声
	case strings.ContainsAny(pf, 去声字母):
		return 去声
	}
	return 轻声
}

// 拆音 拆分拼音为声母韵母
func 拆音(p string) (sm 声母枚举, ym 韵母枚举, err error) {
	if len(p) == 1 {
		sm = 声母0
		// 韵母ê, 因为文字稀少, 并入 ie
		if p == "ê" {
			ym = 韵母ie
			return
		}
		if y, ok := aoeiu[p]; ok {
			ym = y
			return
		}
		err = errors.New("无法识别零声母拼音" + p)
		return
	}
	if s, ok := smap[p[:1]]; ok {
		sm = s
		ok = false
		switch p[1:2] {
		case "a":
			if p[:1] == "w" {
				ym, ok = wamap[p[1:]]
			} else {
				ym, ok = amap[p[1:]]
			}
		case "o":
			if p[:1] == "w" {
				ym = 韵母uo
				ok = true
			} else {
				ym, ok = omap[p[1:]]
			}
		case "e":
			if p[:1] == "w" {
				ym, ok = wemap[p[1:]]
			} else {
				ym, ok = emap[p[1:]]
			}
		case "i":
			if p[:1] == "r" {
				ym = 韵母ri
				ok = true
			} else {
				ym, ok = imap[p[1:]]
			}
		case "u":
			if strings.Contains("jqx", p[:1]) {
				ym, ok = yumap[p[1:]]
			} else {
				ym, ok = umap[p[1:]]
				if !ok {
					ym, ok = yumap[p[1:]]
				}
			}
		case "v", "ü"[:1]:
			ym, ok = yumap[p[1:]]
		}
		if !ok {
			err = errors.New("无法识别拼音" + p + "的韵母部分" + p[1:])
		}
		return
	}
	ok := false
	if strings.Contains(双字声母, p[:1]) {
		if p[1:2] == "h" { // zh ch sh
			switch p[:1] {
			case "z":
				sm = 声母zh
			case "c":
				sm = 声母ch
			case "s":
				sm = 声母sh
			}
			ym, ok = aoeu[p[2:]]
			if !ok {
				if p[2:] == "i" {
					ym = 韵母ri
				} else {
					err = errors.New("无法识别拼音" + p)
				}
			}
			return
		}
		switch p[:1] {
		case "z":
			sm = 声母z
		case "c":
			sm = 声母c
		case "s":
			sm = 声母s
		}
		ym, ok = aoeu[p[1:]]
		if !ok {
			if p[1:] == "i" {
				ym = 韵母ri
			} else {
				err = errors.New("无法识别拼音" + p)
			}
		}
		return
	}
	if p[:1] == "y" { // /j/ or /y/
		if strings.Contains("uvü"[:3], p[1:2]) { // /y/
			sm = 声母yu
			ym, ok = yumap[p[1:]]
			if !ok {
				err = errors.New("无法识别拼音" + p)
			}
			return
		}
		if p[1:] == "ong" { // yong
			sm = 声母yu
			ym = 韵母iong
			ok = true
			return
		}
		sm = 声母yi
		ym, ok = aoeiu[p[1:]]
		if !ok {
			err = errors.New("无法识别拼音" + p)
		}
		return
	}
	sm = 声母0
	ym, ok = aoeiu[p]
	if !ok {
		err = errors.New("无法识别拼音" + p)
	}
	return
}

func 拆音识字(w, r, p, f string) (s 声母枚举, y 韵母枚举, t 声调枚举, rw, rr rune, err error) {
	myp := 去调(f)
	s, y, err = 拆音(p)
	if err != nil {
		return
	}
	mys, myy, err := 拆音(myp)
	if err != nil {
		return
	}
	if mys != s || myy != y {
		err = errors.New("无声调拼音" + p + "与有声调拼音" + f + "不符")
		return
	}
	t = 识调(f)
	rws := []rune(w)
	if len(rws) != 1 {
		err = errors.New("无法正确识别文字" + w)
		return
	}
	rw = rws[0]
	rrs := []rune(r)
	if len(rrs) != 1 {
		err = errors.New("无法正确识别部首" + w)
		return
	}
	rr = rrs[0]
	return
}

// SplitPinyin 拆分带声调拼音
func SplitPinyin(pf string) (sm, ym, sd string, err error) {
	p := 去调(pf)
	s, y, err := 拆音(p)
	if err != nil {
		return
	}
	t := 识调(pf)
	println(s, y, t)
	return s.String(), y.String(), t.String(), nil
}
