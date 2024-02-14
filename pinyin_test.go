package emozi

import "testing"

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
