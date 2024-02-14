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
	es, err := c.Encode("你好，世界！看看多音字：行。")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(es.String())
	if es.String() != "🥛👔🐴👤🌹🐱🐴👩，💦🌞😨🌍➕✌😨👨‍🌾！😭🔐🍉👁️😭🔐🍉👁️🔪🌀🍉🪩🐑🎵🍉🎵👈🌞😨🚼：[👇🦅🧗⛕|🌹👍🧗⛕]。" {
		t.Fatal("got", es.String())
	}
	es, err = c.Encode("你好，世界！指定多音字：银行行。", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(es.String())
	if es.String() != "🥛👔🐴👤🌹🐱🐴👩，💦🌞😨🌍➕✌😨👨‍🌾！🐽🌞🐴✋🔪🦅😨🏠🔪🌀🍉🪩🐑🎵🍉🎵👈🌞😨🚼：🐑🎵🧗💰🌹👍🧗⛕👇🦅🧗⛕。" {
		t.Fatal("got", es.String())
	}
}
