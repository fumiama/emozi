package emozi

import (
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	c, err := NewCoder(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	es, lst, err := c.Encode(false, "你好，世界！看看多音字：行。")
	if err != nil {
		t.Fatal(err)
	}
	if es.String() != "🥛👔🐴👤🌹🐱🐴👩，💦🌞😨🌍➕👴😨👨‍🌾！😭🔐🍉👁️😭🔐🍉👁️🔪🌀🍉🪩🐑🎵🍉🎵👈🌞😨🚼：[👇🦅🧗⛕|🌹👍🧗⛕]。" {
		t.Fatal("got", es.String())
	}
	if len(lst) != 1 && lst[0] != 2 {
		t.Fail()
	}
	es, lst, err = c.Encode(false, "你好，世界！指定多音字：银行行。", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if es.String() != "🥛👔🐴👤🌹🐱🐴👩，💦🌞😨🌍➕👴😨👨‍🌾！🐽🌞🐴✋🔪🦅😨🏠🔪🌀🍉🪩🐑🎵🍉🎵👈🌞😨🚼：🐑🎵🧗💰🌹👍🧗⛕👇🦅🧗⛕。" {
		t.Fatal("got", es.String())
	}
	if len(lst) != 2 && lst[0] != 2 && lst[1] != 2 {
		t.Fail()
	}
	es, _, err = c.Encode(false, "的")
	if err != nil {
		t.Fatal(err)
	}
	if es.String() != "的🈳🈳🈳" {
		t.Fatal("got", es.String())
	}
}

func TestDecode(t *testing.T) {
	c, err := NewCoder(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	s := "你好，世界！看看多音字：行。"
	es, _, err := c.Encode(false, s)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := c.Decode(es, false)
	if err != nil {
		t.Fatal(err)
	}
	if ds != "[你|儗]好，世[界|畍]！看看多音字：[行|行]。" {
		t.Fatal("got", ds)
	}
	es, _, err = c.Encode(false, "你好，世界！指定多音字：银行行。", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	ds, err = c.Decode(es, false)
	if err != nil {
		t.Fatal(err)
	}
	if ds != "[你|儗]好，世[界|畍]！[指|抧|扺]定多音字：[銀|银]行行。" {
		t.Fatal("got", ds)
	}
	es, _, err = c.Encode(false, "好啊")
	if err != nil {
		t.Fatal(err)
	}
	ds, err = c.Decode(es, false)
	if err != nil {
		t.Fatal(err)
	}
	if ds != "好啊" {
		t.Fatal("got", ds)
	}
	es = EmoziString("🌹‪🐱⁢🐴‭👩") // nolint: go-staticcheck
	ds, err = c.Decode(es, false)
	if err != nil {
		t.Fatal(err)
	}
	if ds != "好" {
		t.Fatal("got", ds)
	}
}
