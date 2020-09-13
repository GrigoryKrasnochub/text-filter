package filters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareChars(t *testing.T) {
	a := assert.New(t)
	type testCompareChars struct {
		char      string
		compareTo string
		result    bool
	}
	cases := []testCompareChars{
		{char: "х", compareTo: "}{", result: true},
		{char: "д", compareTo: "d", result: true},
		{char: "д", compareTo: "д", result: true},
		{char: "д", compareTo: "Д", result: true},
		{char: "d", compareTo: "d", result: true},
		{char: "l", compareTo: "d", result: false},
		{char: "д", compareTo: "b", result: false},
		{char: "ы", compareTo: "b|", result: true},
	}
	cc := NewCharsComparer()
	for i, tcase := range cases {
		a.Equal(tcase.result, cc.compareChars([]rune(tcase.char)[0], []rune(tcase.compareTo)[0], func() rune {
			return []rune(tcase.compareTo)[1]
		}), fmt.Sprintf("err in %d test", i))
	}
}
