package unicode

import "testing"

func testEastAsianWidth(t *testing.T) {
	cv("基本功能", func() {
		lines := []string{
			"0123456789",
			"一二三四五",
			//lint:ignore ST1018 intend to do this to check emoji
			"👦👧👨👩👨‍👩‍👧‍👧",
		}

		t.Log("lines in console:")
		for _, line := range lines {
			t.Log(line)
		}

		for _, line := range lines {
			t.Logf("|%v|", ActualEastAsianWidth(line, 30, WithAlign(AlignLeft), WithBlank("-")))
		}
		for _, line := range lines {
			t.Logf("|%v|", ActualEastAsianWidth(line, 30, WithAlign(AlignCenter), WithBlank("二")))
		}
		for _, line := range lines {
			t.Logf("|%v|", ActualEastAsianWidth(line, 30, WithAlign(AlignRight), WithBlank("=")))
		}
	})
}
