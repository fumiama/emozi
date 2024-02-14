package emozi

import "testing"

func TestWrapUnwrap(t *testing.T) {
	t.Log(校验表长度)
	if !WrapRawEmoziString("😨😨😨😨").IsValid() {
		t.Fail()
	}
	if EmoziString("😨😨😨😨😨😨😨").IsValid() {
		t.Fail()
	}
}

func TestString(t *testing.T) {
	t.Log(校验表长度)
	if WrapRawEmoziString("😨😨😨😨").String() != "😨😨😨😨" {
		t.Fail()
	}
}
