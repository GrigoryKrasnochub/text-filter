package filters

import (
	"regexp"
	"sort"
	"unicode"
)

var linksRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
var emailRegex = regexp.MustCompile(`(?:[a-zA-Z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-zA-Z0-9-]*[a-zA-Z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)
var repSym = regexp.MustCompile(`(?:[^a-zA-Z0-9а-яА-ЯёЁ\s]{3,})+`)
var repNLine = regexp.MustCompile(`(?:\n(.{0,10})\n)+`)
var repWhiteSpc = regexp.MustCompile(`(?:\s{2,})`)
var chainSymAndNum = regexp.MustCompile(`((?:([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*\d{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*[^a-zA-Z0-9а-яА-ЯёЁ\s]{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*)|(?:([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*[^a-zA-Z0-9а-яА-ЯёЁ\s]{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*\d{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*))+`)
var notLetter = regexp.MustCompile(`[^a-zA-Zа-яА-ЯёЁ]`)

// Filter links
func FilterLinks(str, replacing string) string {
	if replacing != "" {
		return linksRegex.ReplaceAllString(str, replacing)
	}
	return linksRegex.ReplaceAllString(str, "")
}

// Filter emails
func FilterEmails(str, replacing string) string {
	if replacing != "" {
		return emailRegex.ReplaceAllString(str, replacing)
	}
	return emailRegex.ReplaceAllString(str, "")
}

// Filter same character to one (first of them). Case insensitive
func FilterRepeatedCharsToOne(str string) string {
	result := make([]int32, 0, len(str))
	counter := 0
	maxCount := 3
	charBuffer := make([]int32, 0, maxCount)
	var lastChar int32
	for i, chr := range str {
		// init
		if i == 0 {
			result = append(result, chr)
			lastChar = unicode.ToLower(chr)
			continue
		}

		if unicode.ToLower(chr) == lastChar {
			counter++
			if counter < maxCount {
				charBuffer = append(charBuffer, chr)
			}
		} else {
			if counter < maxCount && len(charBuffer) > 0 {
				result = append(result, charBuffer...)
			}
			result = append(result, chr)
			lastChar = unicode.ToLower(chr)
			charBuffer = charBuffer[:0]
			counter = 0
		}
	}
	return string(result)
}

// Filter character chain excluding a-zA-Z0-9а-яА-ЯёЁ\s
func FilterRepeatedSymbols(str string) string {
	return repSym.ReplaceAllString(str, "")
}

func FilterSymbolsAndNumbersChain(str string) string {
	return chainSymAndNum.ReplaceAllString(str, "")
}

// Filter repeated whitespaces. replace them to one
func FilterRepeatedWhiteSpaces(str string) string {
	return repWhiteSpc.ReplaceAllString(str, " ")
}

// Filter repeated newLines. replace their content to
func FilterRepeatedNewLines(str string) string {
	return repNLine.ReplaceAllString(str, " $1")
}

//TODO ExcludePast
type Word struct {
	ExcludePrev           [][]rune
	Word                  []rune
	lastActiveChar        int
	skipCheckIteration    int
	symbolCounter         int
	lettersBetweenSymbols int
	charsBetweenSymbols   int
	startSymbol           int
	status                WordCompareStatus
}

func (w *Word) reset() {
	w.lastActiveChar = 0
	w.skipCheckIteration = 0
	w.symbolCounter = 0
	w.status = InProgress
	w.resetBetweenWordLetters()
}

func (w *Word) resetBetweenWordLetters() {
	w.lettersBetweenSymbols = 0
	w.charsBetweenSymbols = 0
}

type WordCompareStatus int

const (
	InProgress WordCompareStatus = iota
	Failed
	Success
)

const (
	WordLettersBetweenSymbols = 1
	//TODO 3 символа маловато, если буквы отделить 3 точками слово все равно вполне себе читается. возможно, имеет смысл снять ограничение
	WordCharsBetweenSymbols = 3
)

func (w *Word) compareChar(chr rune, chrComparer charsComparer, getNextChar func() rune) WordCompareStatus {
	w.symbolCounter++
	if w.skipCheckIteration > 0 {
		w.skipCheckIteration--
		return w.status
	}

	result := chrComparer.compareChars(w.Word[w.lastActiveChar], chr, func() rune {
		w.skipCheckIteration++
		return getNextChar()
	})
	if result {
		w.resetBetweenWordLetters()
		w.lastActiveChar++
		if len(w.Word) == w.lastActiveChar {
			w.lastActiveChar = 0
			w.status = Success
			return w.status
		}
	} else {
		if w.symbolCounter > 1 {
			if w.lastActiveChar > 0 && chrComparer.compareChars(w.Word[w.lastActiveChar-1], chr, func() rune {
				w.skipCheckIteration++
				return getNextChar()
			}) {
				w.status = InProgress
				return w.status
			}
			if notLetter.FindAllString(string(chr), -1) != nil && w.charsBetweenSymbols < WordCharsBetweenSymbols && w.lettersBetweenSymbols == 0 {
				w.charsBetweenSymbols++
				w.status = InProgress
				return w.status
			}
			if w.lettersBetweenSymbols < WordLettersBetweenSymbols && w.charsBetweenSymbols == 0 {
				w.lettersBetweenSymbols++
				w.status = InProgress
				return w.status
			}
		}
		w.reset()
		w.status = Failed
		return w.status
	}
	w.status = InProgress
	return w.status
}

func (w *Word) compareWithExcludePrev(str []rune) bool {
	if len(w.ExcludePrev) < 1 {
		return false
	}

	strStartChr := 0
	for _, excludedPrevRuns := range w.ExcludePrev {
		strStartChr = len(str) - len(excludedPrevRuns)
		if strStartChr >= 0 {
			break
		}
	}
	if strStartChr < 0 {
		return false
	}

	skipExcludePrev := make([]bool, len(w.ExcludePrev))
	str = str[strStartChr:]
	strLen := len(str)
	for i := 0; i < len(str); i++ {
		strLen--
		for j, excludedPrevRuns := range w.ExcludePrev {
			if skipExcludePrev[j] == true {
				continue
			}
			if len(excludedPrevRuns) < strLen {
				break
			}
			if excludedPrevRuns[i] != str[i] {
				skipExcludePrev[j] = true
				continue
			}
			if strLen == 0 {
				return true
			}
		}
	}
	return false
}

type WordFilter struct {
	words             []Word
	CharsComparer     charsComparer
	wordsFirstChrsMap map[rune]struct{}
}

func NewWordFilter() WordFilter {
	return WordFilter{
		CharsComparer: newCharsComparer(),
	}
}

//TODO возможно лучше принимать строку?
func (wf *WordFilter) addWord(word []rune, excludedPrev [][]rune) {
	sort.SliceStable(excludedPrev, func(i, j int) bool {
		return len(excludedPrev[i]) > len(excludedPrev[j])
	})
	wf.words = append(wf.words, Word{
		ExcludePrev: excludedPrev,
		Word:        word,
	})
	wf.wordsFirstChrsMap = wf.CharsComparer.getLetterPossibleChars(word[0], wf.wordsFirstChrsMap)
}

func (wf *WordFilter) FilterWords(str string) string {
	if len(wf.words) == 0 {
		return str
	}

	result := ""
	var chrBuf []rune
	runeStr := []rune(str)
	wordStartSymb := 0
	filterWordEnding := false
	for strChrNumb, chr := range runeStr {
		if _, ok := wf.wordsFirstChrsMap[chr]; !ok && notLetter.MatchString(string(chr)) {
			wordStartSymb = strChrNumb + 1
			filterWordEnding = false
		}
		if filterWordEnding {
			continue
		}
		chrBuf = append(chrBuf, chr)
		chrBufToResult := true
		lastIter := len([]rune(str))-1 == strChrNumb
		for i := range wf.words {
			result := wf.words[i].compareChar(chr, wf.CharsComparer, func() rune {
				return runeStr[strChrNumb+1]
			})
			switch result {
			case Success:
				if !wf.words[i].compareWithExcludePrev(chrBuf[:len(chrBuf)-wf.words[i].symbolCounter]) {
					wordLen := wf.words[i].symbolCounter
					wordLenFromWordStart := strChrNumb - wf.words[i].startSymbol + 1
					if len(chrBuf) >= wordLenFromWordStart && wordLen < wordLenFromWordStart {
						wordLen = wordLenFromWordStart
					}
					chrBuf = chrBuf[:len(chrBuf)-wordLen]
					chrBufToResult = true
					filterWordEnding = true
				}
				//Избавиться от этого сброса
				wf.words[i].reset()
			case InProgress:
				chrBufToResult = false
				if wf.words[i].symbolCounter == 1 {
					wf.words[i].startSymbol = wordStartSymb
				}
			}
			if result == Success {
				break
			}
		}
		if (chrBufToResult && wordStartSymb > strChrNumb) || lastIter {
			result += string(chrBuf)
			chrBuf = chrBuf[:0]
		}
	}

	//Избавиться от этого сброса
	for i := range wf.words {
		wf.words[i].reset()
	}

	return result
}
