package emozi

import (
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	c, err := NewCoder(false, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	es, lst, err := c.Encode("你好，世界！看看多音字：行。")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(es.String(), lst)
	if es.String() != "🥛👔🐴👤🌹🐱🐴👩，💦🌞😨🌍➕✌😨👨‍🌾！😭🔐🍉👁️😭🔐🍉👁️🔪🌀🍉🪩🐑🎵🍉🎵👈🌞😨🚼：[👇🦅🧗⛕|🌹👍🧗⛕]。" {
		t.Fatal("got", es.String())
	}
	if len(lst) != 1 && lst[0] != 2 {
		t.Fail()
	}
	es, lst, err = c.Encode("你好，世界！指定多音字：银行行。", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(es.String(), lst)
	if es.String() != "🥛👔🐴👤🌹🐱🐴👩，💦🌞😨🌍➕✌😨👨‍🌾！🐽🌞🐴✋🔪🦅😨🏠🔪🌀🍉🪩🐑🎵🍉🎵👈🌞😨🚼：🐑🎵🧗💰🌹👍🧗⛕👇🦅🧗⛕。" {
		t.Fatal("got", es.String())
	}
	if len(lst) != 2 && lst[0] != 2 && lst[1] != 2 {
		t.Fail()
	}
}
