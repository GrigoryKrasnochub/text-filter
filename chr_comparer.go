package filters

import (
	"sort"
	"unicode"
)

type CharsComparer struct {
	charsMap map[rune][][]rune
}

func NewCharsComparer(charsMap map[string][]string) CharsComparer {
	charsComparer := CharsComparer{
		charsMap: make(map[rune][][]rune, len(charsMap)),
	}
	charsComparer.AddCharsMap(charsMap)
	return charsComparer
}

func (cc *CharsComparer) AddCharsMap(charsMap map[string][]string) {
	convertedMap := convertCompareChars(charsMap)
	for key, val := range convertedMap {
		cc.charsMap[key] = val
	}
}

func (cc *CharsComparer) SetCharsMap(charsMap map[string][]string) {
	cc.charsMap = convertCompareChars(charsMap)
}

func convertCompareChars(charsMap map[string][]string) map[rune][][]rune {
	compChars := make(map[rune][][]rune, len(charsMap))
	for key, charMap := range charsMap {
		charMapRune := make([][]rune, 0, len(charMap))
		for _, chars := range charMap {
			charMapRune = append(charMapRune, []rune(chars))
		}
		sort.SliceStable(charMapRune, func(i, j int) bool {
			return len(charMapRune[i]) < len(charMapRune[j])
		})
		compChars[[]rune(key)[0]] = charMapRune
	}
	return compChars
}

func (cc *CharsComparer) compareChars(sample, compareTo rune, getNextChar func() rune) bool {
	compareTo = unicode.ToLower(compareTo)
	sample = unicode.ToLower(sample)
	if _, ok := cc.charsMap[sample]; !ok {
		return sample == compareTo
	}
	compareToChars := []rune{compareTo}
	for _, variant := range cc.charsMap[sample] {
		if lenDif := len(variant) - len(compareToChars); variant[0] == compareToChars[0] {
			for i := 0; i < lenDif; i++ {
				compareToChars = append(compareToChars, getNextChar())
			}
			for i := 0; i < len(variant); i++ {
				if compareToChars[i] != variant[i] {
					break
				}
				if len(compareToChars) == i+1 {
					return true
				}
			}
		}
	}

	return false
}

func (cc *CharsComparer) fillLettersPossibleChars(words []Word, result map[rune]struct{}) map[rune]struct{} {
	for _, word := range words {
		if len(word.Word) < 1 {
			continue
		}
		result = cc.fillLetterPossibleChars(word.Word[0], result)
	}

	return result
}

func (cc *CharsComparer) fillLetterPossibleChars(letter rune, resultMap map[rune]struct{}) map[rune]struct{} {
	letterChrs := cc.charsMap[letter]
	for _, chr := range letterChrs {
		if len(chr) < 1 {
			continue
		}
		resultMap[chr[0]] = struct{}{}
	}

	return resultMap
}
