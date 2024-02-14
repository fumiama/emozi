package main

// 数据来自 https://github.com/shuowenjiezi/shuowen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/liuzl/gocc"

	"github.com/fumiama/emozi"
)

type data struct {
	W string   `json:"wordhead"`
	R string   `json:"radical"`
	P string   `json:"pinyin"`
	F string   `json:"pinyin_full"`
	A []string `json:"pinyin_alternative"`
}

func main() {
	x := data{}
	t2s, err := gocc.New("t2s")
	if err != nil {
		panic(fmt.Sprintf("ERROR: creating gocc: %v", err))
	}
	_ = os.RemoveAll(emozi.EmoziDatabasePath)
	c, err := emozi.NewCoder(false, time.Minute)
	if err != nil {
		panic(fmt.Sprintf("ERROR: creating emozi coder: %v", err))
	}
	for i := 1; i <= 9833; i++ {
		f, err := os.Open(fmt.Sprintf("./data/%d.json", i))
		if err != nil {
			panic(fmt.Sprintf("ERROR: opening data/%d.json: %v", i, err))
		}
		x.A = x.A[:0]
		err = json.NewDecoder(f).Decode(&x)
		if err != nil {
			panic(fmt.Sprintf("ERROR: decoding data/%d.json: %v", i, err))
		}
		_ = f.Close()
		x.P = strings.ReplaceAll(x.P, string(emozi.G), "g")
		if len(x.P) == 0 {
			panic(fmt.Sprintf("ERROR: decoding data/%d.json: p: %s, f: %s", i, x.P, x.F))
		}
		insert := func(w string) error {
			err = c.Add(w, x.R, x.P, x.F)
			if err != nil {
				return fmt.Errorf("inserting table emozi of data/%d.json: %v", i, err)
			}
			for _, a := range x.A {
				err = c.Add(w, x.R, "", a)
				if err != nil {
					return fmt.Errorf("inserting table emozi of data/%d.json, alter %s: %v", i, a, err)
				}
			}
			return nil
		}
		fmt.Print("\r[", i, "/9833] Insert char: ", x.W, "                             ")
		err = insert(x.W)
		if err != nil {
			fmt.Println("\n\t\tWARN:", err)
		}
		sc, err := t2s.Convert(x.W)
		if err != nil {
			continue
		}
		if sc != x.W {
			fmt.Print("\r[", i, "/9833] insert simplified char: ", sc, "                             ")
			err = insert(sc)
			if err != nil {
				fmt.Println("\n\t\tWARN:", err)
			}
		}
	}
}
