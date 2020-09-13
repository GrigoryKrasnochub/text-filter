package filters

import (
	"sort"
	"unicode"
)

var RuChars = map[string][]string{
	"a": {"а", "a", "@",},
	"б": {"б", "6", "b",},
	"в": {"в", "b", "v",},
	"г": {"г", "r", "g",},
	"д": {"д", "d", "g",},
	"e": {"е", "e",},
	"ё": {"ё", "е", "e",},
	"ж": {"ж", "zh", "*",},
	"з": {"з", "3", "z",},
	"и": {"и", "u", "i",},
	"й": {"й", "u", "y", "i",},
	"к": {"к", "k", "i{", "|{",},
	"л": {"л", "l", "ji", "|\\", "/\\",},
	"м": {"м", "m",},
	"н": {"н", "h", "n",},
	"о": {"о", "o", "0",},
	"п": {"п", "n", "p",},
	"р": {"р", "r", "p",},
	"с": {"с", "c", "s",},
	"т": {"т", "m", "t",},
	"у": {"у", "y", "u",},
	"ф": {"ф", "f",},
	"х": {"х", "x", "h", "к", "k", "}{", "][",},
	"ц": {"ц", "с", "u",},
	"ч": {"ч", "ch",},
	"ш": {"ш", "sh",},
	"щ": {"щ", "sch",},
	"ь": {"ь", "b",},
	"ы": {"ы", "bi", "b|", "ьi", "ь|",},
	"ъ": {"ъ",},
	"э": {"э", "е", "e",},
	"ю": {"ю", "io",},
	"я": {"я", "ya",},
}

type CharsComparer struct {
	charsMap map[rune][][]rune
}

func NewCharsComparer() CharsComparer {
	convertedMap := convertCompareChars(RuChars)
	charsComparer := CharsComparer{charsMap: make(map[rune][][]rune, len(convertedMap))}
	for key, val := range convertedMap {
		charsComparer.charsMap[key] = val
	}
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

func (cc *CharsComparer) getFirstLettersPossibleChars(words []Word) (result map[rune]struct{}) {
	for _, word := range words {
		if len(word.Word) < 1 {
			continue
		}

		result = cc.getLetterPossibleChars(word.Word[0], result)
	}

	return result
}

func (cc *CharsComparer) getLetterPossibleChars(letter rune, resultMap map[rune]struct{}) map[rune]struct{} {
	if resultMap == nil {
		resultMap = make(map[rune]struct{})
	}

	LetterChrs := cc.charsMap[letter]
	for _, chr := range LetterChrs {
		if len(chr) < 1 {
			continue
		}
		if _, ok := resultMap[chr[0]]; !ok {
			resultMap[chr[0]] = struct{}{}
		}
	}

	return resultMap
}
