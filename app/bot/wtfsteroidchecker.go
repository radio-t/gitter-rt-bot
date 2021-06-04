package bot

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// WTFSteroidChecker check if command wtf{!,?} is written with additional characters
// "𝀥tf!" should be recognized as "wtf?" and so on
type WTFSteroidChecker struct {
	message string
}

// WTFUnicodeLibrary contains unicode characters and strings that looks like "w","t","f","!","?"
func (w *WTFSteroidChecker) WTFUnicodeLibrary() map[string][]string {
	repl := make(map[string][]string)
	repl["w"] = []string{
		"ᘺ",
		"ய",
		"ʍ",
		"Ⱳ",
		"ⱳ",
		"ᴡ",
		"🅆",
		"🆆",
		"ᵂ",
		"ʷ",
		"🅦",
		"Ⓦ",
		"𝓦",
		"𝙒",
		"𝖂",
		"Ｗ",
		"ⓦ",
		"𝑤",
		"𝕨",
		"𝖜",
		"ᷱ",
		"ｗ",
		"ꙍ",
		"в",
		"₩",
		"𝀥",
		"⨈",
		"🇼",
		"\\/\\/",
		"🄦",
		"⒲",
		"ᐯᐯ",
		"ᏙᏙ",
		"ᜠᜠ",
		"ⴸⴸ",
		"ᶺᶺ",
		"ɅɅ",
		"ʌʌ",
		"ⱴⱴ",
		"ⱱⱱ",
		"ƲƲ",
		"ʋʋ",
		"ᶌᶌ",
		"ꝞꝞ",
		"ꝟꝟ",
		"ᴠᴠ",
		"ṽṽ",
		"🅅🅅",
		"🆅🆅",
		"ⱽⱽ",
		"ᵥᵥ",
		"ᵛᵛ",
		"🅥🅥",
		"ⓋⓋ",
		"𝖁𝖁",
		"^^",
		"𝘝𝘝",
		"𝕍𝕍",
		"𝚅𝚅",
		"𝖵𝖵",
		"ⅤⅤ",
		"ＶＶ",
		"VV",
		"ⓥⓥ",
		"𝖛𝖛",
		"𝕧𝕧",
		"𝘷𝘷",
		"𝚟𝚟",
		"𝗏𝗏",
		"ⅴⅴ",
		"ｖｖ",
		"vv",
		"ѴѴ",
		"ѵѵ",
		"𝈍𝈍",
		"🇻 🇻",
		"⋁⋁",
		"√√",
		"ˇˇ",
		"🄥🄥",
		"⒱⒱"}
	repl["t"] = []string{
		"丅",
		"𐤯",
		"𐊗",
		"ナ",
		"ߠ",
		"Ϯ",
		"ϯ",
		"Ʇ",
		"ʇ",
		"ȶ",
		"ᵀ",
		"🅃",
		"🆃",
		"ᵗ",
		"🅣",
		"Ⓣ",
		"𝕿",
		"𝕋",
		"Ｔ",
		"ⓣ",
		"𝖙",
		"ɫ",
		"ꝉ",
		"т",
		"⥡",
		"🇹",
		"╩",
		"╨",
		"╦",
		"╥",
		"┼",
		"┴",
		"┭",
		"┬",
		"∤",
		"⸷",
		"‡",
		"†",
		"🄣",
		"⒯"}
	repl["f"] = []string{
		"𐌅",
		"𖨝",
		"ϝ",
		"ʄ",
		"ꟻ",
		"Ⅎ",
		"ⅎ",
		"Ƒ",
		"ƒ",
		"ᵮ",
		"Ꞙ",
		"ꞙ",
		"ꬵ",
		"Ꝼ",
		"ꝼ",
		"🄵",
		"🅵",
		"🅕",
		"Ⓕ",
		"ℱ",
		"𝕱",
		"Ｆ",
		"ⓕ",
		"𝕗",
		"𝔣",
		"𝓯",
		"𝖋",
		"ｆ",
		"ф",
		"Ғ",
		"£",
		"⨚",
		"⨑",
		"⨍",
		"🇫",
		"℉",
		"🄕",
		"⒡"}
	repl["!"] = []string{
		"i",
		"1",
		"❗",
		"❕",
		"║",
		"|",
		"ꜟ",
		"ꜞ",
		"ꜝ",
		"¡",
		"︕",
		"﹗",
		"︁！",
		"⁉",
		"‼"}
	repl["?"] = []string{
		"7",
		"❔",
		"❓",
		"⍰",
		"؟",
		"⸮",
		"¿",
		"︖",
		"﹖",
		"？",
		"?",
		"⁇",
		"⁈"}
	return repl
}

// removeDiacretic smart remove diacritic marks
// isMn check rune is in Unicode Mn category nonspacing marks
// Example ẃŧḟ! -> wtf!
// https://blog.golang.org/normalization#TOC_10.
// https://pkg.go.dev/golang.org/x/text/runes#Remove
func (w *WTFSteroidChecker) removeDiacretic() {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	w.message, _, _ = transform.String(t, w.message)
}

// removeUnicodeAnalog replace characters that looks like "w","t","f","!", "?" with their ASCII representation
func (w *WTFSteroidChecker) removeUnicodeAnalog() {
	replaceMap := w.WTFUnicodeLibrary()
	for mainLetter, listOfUnicodes := range replaceMap {
		for _, unicodeSymbol := range listOfUnicodes {
			w.message = strings.ReplaceAll(w.message, unicodeSymbol, mainLetter)
		}
	}
}

// removeNotASCIIAndNotRussian delete all non-unicode characters except russian unicode characters
// Example: W؈T؈F؈! → WTF!
// "Вот фон!" ↛ "wtf!" correct is "Вот фон!" → "wоt fон!"
func (w *WTFSteroidChecker) removeNotASCIIAndNotRussian() {
	w.message = strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII && (r < 0x0400 && r > 0x04ff) {
			return -1
		}
		return r
	}, w.message)
}

// removeNotLetters delete all non-letter characters
// Example w_t_f_!, w-t-f-! → wtf!
func (w *WTFSteroidChecker) removeNotLetters() {
	w.message = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || r == '!' || r == '?' {
			return r
		}
		return -1
	}, w.message)
}

// Contains remove all bad symbols from message and check if the message contained the commands
func (w *WTFSteroidChecker) Contains() bool {

	w.message = strings.ToLower(w.message)
	w.removeDiacretic()
	w.removeUnicodeAnalog()
	w.removeNotASCIIAndNotRussian()
	w.removeNotLetters()

	return contains([]string{"wtf!", "wtf?"}, w.message)

}
