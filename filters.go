package filters

import (
	"regexp"
	"sort"
	"unicode"
	"unicode/utf8"
)

var (
	linksRegex     = regexp.MustCompile(`https?://(www\.)?[-a-zA-Zа-яА-ЯёЁ0-9@:%._+~#=]{1,256}\.[a-zA-Zа-яА-ЯёЁ0-9()]{1,6}([-a-zA-Zа-яА-ЯёЁ0-9()@:%_+.~#?&/=]*)`)
	emailRegex     = regexp.MustCompile(`(?:[a-zA-Z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-zA-Z0-9-]*[a-zA-Z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)
	repSym         = regexp.MustCompile(`(?:[^a-zA-Z0-9а-яА-ЯёЁ\s]{3,})+`)
	repNLine       = regexp.MustCompile(`(?:\n(.{0,10})\n)+`)
	repWhiteSpc    = regexp.MustCompile(`(?:\s{2,})`)
	chainSymAndNum = regexp.MustCompile(`((?:([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*\d{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*[^a-zA-Z0-9а-яА-ЯёЁ\s]{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*)|(?:([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*[^a-zA-Z0-9а-яА-ЯёЁ\s]{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*\d{2,}([^a-zA-Z0-9а-яА-ЯёЁ\s]*\d*)*))+`)
)

// Filter links
func FilterLinks(str, replacing string) string {
	return linksRegex.ReplaceAllString(str, replacing)
}

// Filter emails
func FilterEmails(str, replacing string) string {
	return emailRegex.ReplaceAllString(str, replacing)
}

// Filter same character to one (first of them). Case insensitive
func FilterRepeatedCharsToOne(str string, maxCount int) string {
	result := make([]int32, 0, utf8.RuneCountInString(str))
	counter := 0
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
			if counter < maxCount {
				result = append(result, charBuffer...)
			}
			result = append(result, chr)
			lastChar = unicode.ToLower(chr)
			charBuffer = charBuffer[:0]
			counter = 0
		}
	}
	if counter < maxCount {
		result = append(result, charBuffer...)
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

type excludedPast struct {
	excludedPart []rune
	status       bool
}

type preparedWord struct {
	excludePrev  [][]rune
	excludePast  []excludedPast
	searchedWord []rune
}

type wordProcessor struct {
	options *options
	preparedWord
	lastActiveChar        int
	skipCheckIteration    int
	symbolCounter         int
	lettersBetweenSymbols int
	charsBetweenSymbols   int
	startSymbol           int
	status                wordCompareStatus
}

func (w *wordProcessor) reset() {
	w.lastActiveChar = 0
	w.skipCheckIteration = 0
	w.symbolCounter = 0
	w.status = inProgress
	w.resetBetweenWordLetters()
}

func (w *wordProcessor) resetBetweenWordLetters() {
	w.lettersBetweenSymbols = 0
	w.charsBetweenSymbols = 0
}

func resetAllWords(words []wordProcessor) {
	for i, _ := range words {
		words[i].reset()
	}
}

type wordCompareStatus int

const (
	inProgress wordCompareStatus = iota
	failed
	success
)

const (
	wordLettersBetweenSymbols = 1
	//TODO 3 символа маловато, если буквы отделить 3 точками слово все равно вполне себе читается. возможно, имеет смысл снять ограничение
	wordCharsBetweenSymbols = 3
)

func (w *wordProcessor) compareChar(chr rune, chrComparer CharsComparer, getNextChar func() rune) wordCompareStatus {
	w.symbolCounter++
	if w.skipCheckIteration > 0 {
		w.skipCheckIteration--
		return w.status
	}

	result := chrComparer.compareChars(w.searchedWord[w.lastActiveChar], chr, func() rune {
		w.skipCheckIteration++
		return getNextChar()
	})
	if result {
		w.resetBetweenWordLetters()
		w.lastActiveChar++
		if len(w.searchedWord) == w.lastActiveChar {
			w.status = success
			return w.status
		}
	} else {
		if w.symbolCounter > 1 {
			if w.lastActiveChar > 0 && chrComparer.compareChars(w.searchedWord[w.lastActiveChar-1], chr, func() rune {
				w.skipCheckIteration++
				return getNextChar()
			}) {
				w.status = inProgress
				return w.status
			}
			if !unicode.IsLetter(chr) && w.charsBetweenSymbols < w.options.symbolsBetweenKeyLetters && w.lettersBetweenSymbols == 0 {
				w.charsBetweenSymbols++
				w.status = inProgress
				return w.status
			}
			if w.lettersBetweenSymbols < w.options.lettersBetweenKeyLetters && w.charsBetweenSymbols == 0 {
				w.lettersBetweenSymbols++
				w.status = inProgress
				return w.status
			}
		}
		w.reset()
		w.status = failed
		return w.status
	}
	w.status = inProgress
	return w.status
}

func (w *wordProcessor) compareWithExcludePrev(str []rune) bool {
	if len(w.excludePrev) < 1 {
		return false
	}

	strStartChr := 0
	for _, excludedPrevRuns := range w.excludePrev {
		strStartChr = len(str) - len(excludedPrevRuns)
		if strStartChr >= 0 {
			break
		}
	}
	if strStartChr < 0 {
		return false
	}

	skipExcludePrev := make([]bool, len(w.excludePrev))
	str = str[strStartChr:]
	strLen := len(str)
	for i := 0; i < len(str); i++ {
		strLen--
		for j, excludedPrevRuns := range w.excludePrev {
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

//func (w *wordProcessor) compareWithExcludedPast(chr rune, symbCount int) {
//	for _,
//}

type WordFilter struct {
	options           *options
	words             []preparedWord
	CharsComparer     CharsComparer
	wordsFirstChrsMap map[rune]struct{}
}

func NewWordFilter(charsMap map[string][]string, optionFuncs ...OptionFunc) WordFilter {
	opts := getDefaultOptions()

	for _, option := range optionFuncs {
		option(opts)
	}

	return WordFilter{
		options:           opts,
		CharsComparer:     NewCharsComparer(charsMap),
		wordsFirstChrsMap: make(map[rune]struct{}, 0),
	}
}

func (wf *WordFilter) ResetWords() {
	wf.words = nil
	wf.wordsFirstChrsMap = nil
}

type UserWord struct {
	Word         string
	ExcludedPrev []string
	ExcludedPast []string
}

func (wf *WordFilter) AddWords(userWords []UserWord) {
	wordsSlice := make([]preparedWord, len(userWords))
	for i, userWord := range userWords {
		formattedExcludePrev := make([][]rune, len(userWord.ExcludedPrev))
		for j, prev := range userWord.ExcludedPrev {
			formattedExcludePrev[j] = []rune(prev)
		}
		excludedPastParts := make([]excludedPast, 0, len(userWords))
		for _, excludedPart := range userWords[i].ExcludedPast {
			excludedPastParts = append(excludedPastParts, excludedPast{
				excludedPart: []rune(excludedPart),
			})
		}
		wordsSlice[i] = preparedWord{
			excludePrev:  formattedExcludePrev,
			excludePast:  excludedPastParts,
			searchedWord: []rune(userWord.Word),
		}
	}
	wf.words = append(wf.words, wordsSlice...)
	wf.wordsFirstChrsMap = wf.CharsComparer.fillLettersPossibleChars(wordsSlice, wf.wordsFirstChrsMap)
}

func (wf *WordFilter) AddWord(word string, excludedPrev []string) {
	formattedExcludePrev := make([][]rune, len(excludedPrev))
	for i, prev := range excludedPrev {
		formattedExcludePrev[i] = []rune(prev)
	}
	wf.addWord([]rune(word), formattedExcludePrev)
}

func (wf *WordFilter) addWord(searchedWord []rune, excludedPrev [][]rune) {
	sort.SliceStable(excludedPrev, func(i, j int) bool {
		return len(excludedPrev[i]) > len(excludedPrev[j])
	})
	wf.words = append(wf.words, preparedWord{
		excludePrev:  excludedPrev,
		searchedWord: searchedWord,
	})
	wf.wordsFirstChrsMap = wf.CharsComparer.fillLetterPossibleChars(searchedWord[0], wf.wordsFirstChrsMap)
}

type DetectedWord struct {
	OriginalWord string
	Beginning    string
	Word         string
	Ending       string
}

func (wf *WordFilter) FilterWords(str string, replaceWord func(DetectedWord) string) string {
	if len(wf.words) == 0 {
		return str
	}

	words := make([]wordProcessor, 0, len(wf.words))
	for _, prepearedWord := range wf.words {
		words = append(words, wordProcessor{
			options:      wf.options,
			preparedWord: prepearedWord,
		})
	}

	var chrBuf []rune
	runeStr := []rune(str)
	wordStartSymb := 0
	detectWord := false
	detectedWord := DetectedWord{}
	detectedWordEnding := make([]rune, 0)
	result := make([]rune, 0, utf8.RuneCountInString(str))
	for strChrNumb, chr := range runeStr {
		positionBetweenWords := false
		if _, ok := wf.wordsFirstChrsMap[chr]; !ok && !unicode.IsLetter(chr) {
			wordStartSymb = strChrNumb + 1
			positionBetweenWords = true
			if detectWord {
				detectedWord.Ending = string(detectedWordEnding)
				result = append(result, []rune(replaceWord(detectedWord))...)
				detectWord = false
			}
		}
		if detectWord {
			detectedWordEnding = append(detectedWordEnding, chr)
			continue
		}

		wordsNotInProgress := true
		chrBuf = append(chrBuf, chr)
		for i := range words {
			compareStatus := words[i].compareChar(chr, wf.CharsComparer, func() rune {
				return runeStr[strChrNumb+1]
			})
			switch compareStatus {
			case success:
				if !words[i].compareWithExcludePrev(chrBuf[:len(chrBuf)-words[i].symbolCounter]) {
					wordLenFromWordStart := strChrNumb - words[i].startSymbol + 1

					detectedWord.Beginning = string(chrBuf[len(chrBuf)-wordLenFromWordStart : len(chrBuf)-words[i].symbolCounter])
					detectedWord.Word = string(chrBuf[len(chrBuf)-words[i].symbolCounter:])
					detectedWord.OriginalWord = string(words[i].searchedWord)

					chrBuf = chrBuf[:len(chrBuf)-wordLenFromWordStart]
					detectedWordEnding = make([]rune, 0)
					wordsNotInProgress = true
					detectWord = true
				}
				resetAllWords(words)
			case inProgress:
				wordsNotInProgress = false
				if words[i].symbolCounter == 1 {
					words[i].startSymbol = wordStartSymb
				}
			}
			if compareStatus == success {
				break
			}
		}
		if wordsNotInProgress && positionBetweenWords {
			result = append(result, chrBuf...)
			chrBuf = chrBuf[:0]
		}
	}

	result = append(result, chrBuf...)
	if detectWord {
		detectedWord.Ending = string(detectedWordEnding)
		result = append(result, []rune(replaceWord(detectedWord))...)
	}

	return string(result)
}
