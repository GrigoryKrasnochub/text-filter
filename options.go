package filters

type options struct {
	lettersBetweenKeyLetters int // сколько букв может быть между ключевыми символами хпупй
	symbolsBetweenKeyLetters int // сколько символов может быть между ключевыми символами х.у.й
}

func getDefaultOptions() *options {
	return &options{
		lettersBetweenKeyLetters: 1,
		symbolsBetweenKeyLetters: 3,
	}
}

type OptionFunc func(*options)

// WithLettersBetweenKeyLetters сколько букв может быть между ключевыми символами хпупй
func WithLettersBetweenKeyLetters(lettersBetweenKeyLetters int) OptionFunc {
	return func(o *options) {
		o.lettersBetweenKeyLetters = lettersBetweenKeyLetters
	}
}

// WithSymbolsBetweenKeyLetters сколько символов может быть между ключевыми символами х.у.й
func WithSymbolsBetweenKeyLetters(symbolsBetweenKeyLetters int) OptionFunc {
	return func(o *options) {
		o.symbolsBetweenKeyLetters = symbolsBetweenKeyLetters
	}
}
