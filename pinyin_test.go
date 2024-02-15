package emozi

import (
	"strings"
	"testing"
)

func TestShengmuString(t *testing.T) {
	for i := 0; i < len(声母); i++ {
		t.Log(声母枚举(i).String())
		if 声母枚举(i).String() != strings.TrimSpace(声母枚举(i).String()) {
			t.Fatal("声母: '", 声母枚举(i), "'")
		}
	}
}

func TestSplitPinyin(t *testing.T) {
	s, y, tone, err := SplitPinyin("yōng")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s, y, tone)
	if s+y+tone != "ɥi̯ʊŋ阴平" {
		t.Fail()
	}
	s, y, tone, err = SplitPinyin("hóng")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s, y, tone)
	if s+y+tone != "xʊŋ阳平" {
		t.Fail()
	}
	s, y, tone, err = SplitPinyin("yǜn")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s, y, tone)
	if s+y+tone != "ɥyn去声" {
		t.Fail()
	}
	s, y, tone, err = SplitPinyin("jiǒng")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s, y, tone)
	if s+y+tone != "tɕi̯ʊŋ上声" {
		t.Fail()
	}
	s, y, tone, err = SplitPinyin("e")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s, y, tone)
	if s+y+tone != "0ɤ轻声" {
		t.Fail()
	}
}
