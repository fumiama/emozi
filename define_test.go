package emozi

import "testing"

func TestFirstEmojiSingle(t *testing.T) {
	for i, lst := range 声母 {
		if len([]rune(lst[0])) != 1 {
			t.Fatal("声母", i, "长度", len([]rune(lst[0])), "字", lst[0])
		}
	}
	for i, lst := range 韵母 {
		if len([]rune(lst[0])) != 1 {
			t.Fatal("韵母", i, "长度", len([]rune(lst[0])), "字", lst[0])
		}
	}
	for i, lst := range 声调 {
		if len([]rune(lst[0])) != 1 {
			t.Fatal("声调", i, "长度", len([]rune(lst[0])), "字", lst[0])
		}
	}
}
