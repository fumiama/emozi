package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/fumiama/emozi"
)

func main() {
	dbpath := flag.String("db", emozi.EmoziDatabasePath, "符合规范的查询数据库位置, 不存在则会自动释放到该路径.")
	isencode := flag.String("e", "", "编码汉字序列为颜文字")
	isdecode := flag.String("d", "", "解码颜文字为汉字序列")
	getglobalid := flag.Bool("i", false, "指定汉字-a和带声调的拼音-p以计算其全局唯一ID")
	addoverlay := flag.String("a", "", "添加一个汉字到附加库")
	pinyinfull := flag.String("p", "", "带声调的拼音")
	radical := flag.String("r", "", "指定欲编辑的部首")
	radicalemozi := flag.String("re", "", "指定部首对应的颜文字")
	noRandom := flag.Bool("nr", false, "不随机选取所有读音相近的颜文字")
	showhelp := flag.Bool("h", false, "显示帮助信息")
	forcedecode := flag.Bool("f", false, "强制解码并非由本程序生成的颜文字序列")
	flag.Parse()
	if *showhelp {
		fmt.Println("用法: [-h|f|nr] [-db 字.db] [-d 🌹⁪😺‎🐴‫👩] [-e 好] 形声字选择1 形声字选择2 ...")
		flag.PrintDefaults()
		return
	}
	emozi.EmoziDatabasePath = *dbpath
	coder, err := emozi.NewCoder(time.Minute)
	if err != nil {
		panic(err)
	}
	defer coder.Close()
	if *isencode != "" {
		rem := flag.Args()
		lst := make([]int, len(rem))
		for i, ns := range rem {
			n, err := strconv.Atoi(ns)
			if err != nil {
				panic("第" + strconv.Itoa(i+1) + "个形声字参数 '" + ns + "' 非法")
			}
			lst[i] = n
		}
		es, lst, err := coder.Encode(!*noRandom, *isencode, lst...)
		if err != nil {
			panic(err)
		}
		fmt.Println("编码结果:", string(es))
		if len(lst) > 0 && len(rem) == 0 {
			fmt.Println("可选形声:", lst)
			fmt.Println("在参数中指定形声字编号(从0开始)以生成不带中括号的编码结果")
		}
	}
	if *isdecode != "" {
		s, err := coder.Decode(emozi.EmoziString(*isdecode), *forcedecode)
		switch {
		case err == emozi.ErrInvalidEmoziString:
			fmt.Println("解码结果: 非本程序直接生成的颜文字序列或序列经过人为修改")
		case err != nil:
			panic(err)
		default:
			fmt.Println("解码结果:", s)
		}
	}
	if *addoverlay != "" && *pinyinfull != "" && *radical != "" {
		id, desc, err := coder.AddCharOverlay(*addoverlay, *radical, "", *pinyinfull)
		if err != nil {
			panic(err)
		}
		fmt.Println("已添加汉字:", *addoverlay, "读音:", desc, "部首:", *radical, "ID:", id)
	}
	if *radical != "" && *radicalemozi != "" {
		rr := []rune(*radical)
		if len(rr) != 1 {
			panic("非法部首 '" + *radical + "': 长度为" + strconv.Itoa(len(rr)))
		}
		err = coder.AddRadicalOverlay(rr[0], *radicalemozi)
		if err != nil {
			panic(err)
		}
		fmt.Println("已添加部首:", *radical, "颜文字:", *radicalemozi)
	}
	if *getglobalid && *addoverlay != "" && *pinyinfull != "" {
		sm, ym, sd, err := emozi.SplitPinyin(*pinyinfull)
		if err != nil {
			panic(err)
		}
		r := []rune(*addoverlay)[0]
		id, _ := emozi.CharGlobalID(r, *pinyinfull)
		fmt.Println("文字:", string(r), "拼音IPA:", sm, ym, sd, "ID:", id)
	}
}
