package slugify

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

var (
	defaultSlugger = New(Configuration{})
)

func Slugify(value string) string {
	return defaultSlugger.Slugify(value)
}

func validCharacter(c rune) bool {

	if c >= 'a' && c <= 'z' {
		return true
	}

	if c >= '0' && c <= '9' {
		return true
	}

	return false
}

type Slugifier struct {
	isValidCharacter func(c rune) bool
	replaceCharacter rune
	replacementMap   map[rune]string
}

func (s Slugifier) Slugify(value string) string {

	value = strings.ToLower(value)
	var buffer bytes.Buffer
	lastCharacterWasInvalid := false

	for len(value) > 0 {
		c, size := utf8.DecodeRuneInString(value)
		value = value[size:]
		if newCharacter, ok := s.replacementMap[c]; ok {
			buffer.WriteString(newCharacter)
			lastCharacterWasInvalid = false
			continue
		}
		if s.isValidCharacter(c) {
			buffer.WriteRune(c)
			lastCharacterWasInvalid = false
		} else if lastCharacterWasInvalid == false {
			buffer.WriteRune(s.replaceCharacter)
			lastCharacterWasInvalid = true
		}
	}

	return strings.Trim(buffer.String(), string(s.replaceCharacter))
}

type Configuration struct {
	IsValidCharacterChecker func(rune) bool
	ReplaceCharacter        rune
	ReplacementMap          map[rune]string
}

func New(config Configuration) *Slugifier {
	if config.IsValidCharacterChecker == nil {
		config.IsValidCharacterChecker = validCharacter
	}

	if config.ReplaceCharacter == 0 {
		config.ReplaceCharacter = '-'
	}

	if config.ReplacementMap == nil {
		config.ReplacementMap = map[rune]string{
			'&': "and",
			'@': "at",
			'©': "c",
			'®': "r",
			'Æ': "ae",
			'ß': "ss",
			'à': "a",
			'á': "a",
			'â': "a",
			'ä': "a",
			'å': "a",
			'æ': "ae",
			'ç': "c",
			'è': "e",
			'é': "e",
			'ê': "e",
			'ë': "e",
			'ì': "i",
			'í': "i",
			'î': "i",
			'ï': "i",
			'ò': "o",
			'ó': "o",
			'ô': "o",
			'õ': "o",
			'ö': "o",
			'ø': "o",
			'ù': "u",
			'ú': "u",
			'û': "u",
			'ü': "u",
			'ý': "y",
			'þ': "p",
			'ÿ': "y",
			'ā': "a",
			'ă': "a",
			'Ą': "a",
			'ą': "a",
			'ć': "c",
			'ĉ': "c",
			'ċ': "c",
			'č': "c",
			'ď': "d",
			'đ': "d",
			'ē': "e",
			'ĕ': "e",
			'ė': "e",
			'ę': "e",
			'ě': "e",
			'ĝ': "g",
			'ğ': "g",
			'ġ': "g",
			'ģ': "g",
			'ĥ': "h",
			'ħ': "h",
			'ĩ': "i",
			'ī': "i",
			'ĭ': "i",
			'į': "i",
			'ı': "i",
			'ĳ': "ij",
			'ĵ': "j",
			'ķ': "k",
			'ĸ': "k",
			'Ĺ': "l",
			'ĺ': "l",
			'ļ': "l",
			'ľ': "l",
			'ŀ': "l",
			'ł': "l",
			'ń': "n",
			'ņ': "n",
			'ň': "n",
			'ŉ': "n",
			'ŋ': "n",
			'ō': "o",
			'ŏ': "o",
			'ő': "o",
			'Œ': "oe",
			'œ': "oe",
			'ŕ': "r",
			'ŗ': "r",
			'ř': "r",
			'ś': "s",
			'ŝ': "s",
			'ş': "s",
			'š': "s",
			'ţ': "t",
			'ť': "t",
			'ŧ': "t",
			'ũ': "u",
			'ū': "u",
			'ŭ': "u",
			'ů': "u",
			'ű': "u",
			'ų': "u",
			'ŵ': "w",
			'ŷ': "y",
			'ź': "z",
			'ż': "z",
			'ž': "z",
			'ſ': "z",
			'Ə': "e",
			'ƒ': "f",
			'Ơ': "o",
			'ơ': "o",
			'Ư': "u",
			'ư': "u",
			'ǎ': "a",
			'ǐ': "i",
			'ǒ': "o",
			'ǔ': "u",
			'ǖ': "u",
			'ǘ': "u",
			'ǚ': "u",
			'ǜ': "u",
			'ǻ': "a",
			'Ǽ': "ae",
			'ǽ': "ae",
			'Ǿ': "o",
			'ǿ': "o",
			'ə': "e",
			'Є': "e",
			'Б': "b",
			'Г': "g",
			'Д': "d",
			'Ж': "zh",
			'З': "z",
			'У': "u",
			'Ф': "f",
			'Х': "h",
			'Ц': "c",
			'Ч': "ch",
			'Ш': "sh",
			'Щ': "sch",
			'Ъ': "-",
			'Ы': "y",
			'Ь': "-",
			'Э': "je",
			'Ю': "ju",
			'Я': "ja",
			'а': "a",
			'б': "b",
			'в': "v",
			'г': "g",
			'д': "d",
			'е': "e",
			'ж': "zh",
			'з': "z",
			'и': "i",
			'й': "j",
			'к': "k",
			'л': "l",
			'м': "m",
			'н': "n",
			'о': "o",
			'п': "p",
			'р': "r",
			'с': "s",
			'т': "t",
			'у': "u",
			'ф': "f",
			'х': "h",
			'ц': "c",
			'ч': "ch",
			'ш': "sh",
			'щ': "sch",
			'ъ': "-",
			'ы': "y",
			'ь': "-",
			'э': "je",
			'ю': "ju",
			'я': "ja",
			'ё': "jo",
			'є': "e",
			'і': "i",
			'ї': "i",
			'Ґ': "g",
			'ґ': "g",
			'א': "a",
			'ב': "b",
			'ג': "g",
			'ד': "d",
			'ה': "h",
			'ו': "v",
			'ז': "z",
			'ח': "h",
			'ט': "t",
			'י': "i",
			'ך': "k",
			'כ': "k",
			'ל': "l",
			'ם': "m",
			'מ': "m",
			'ן': "n",
			'נ': "n",
			'ס': "s",
			'ע': "e",
			'ף': "p",
			'פ': "p",
			'ץ': "C",
			'צ': "c",
			'ק': "q",
			'ר': "r",
			'ש': "w",
			'ת': "t",
			'™': "tm",
		}
	}

	return &Slugifier{
		isValidCharacter: config.IsValidCharacterChecker,
		replaceCharacter: config.ReplaceCharacter,
		replacementMap:   config.ReplacementMap,
	}
}
