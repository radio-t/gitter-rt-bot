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

// WTFUnicodeDiacreticLibrary contains diacretic unicode symbols that looks like "w","t","f","!","?"
// All symbols that removes by removeDiacretic function
func (w *WTFSteroidChecker) WTFUnicodeDiacreticLibrary() map[string][]string {
	repl := make(map[string][]string)
	repl["w"] = []string{"ᷱ"}
	repl["t"] = []string{"∤"}
	repl["f"] = []string{"ᷥ", "ᷫ"}
	repl["!"] = []string{"︁！"}
	repl["?"] = []string{}
	return repl
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
		"ｗ",
		"ꙍ",
		"в",
		"ʚ",
		"₩",
		"𝀥",
		"⨈",
		"🇼",
		"Ꮃ",
		"Ꮚ",
		"Ꮤ",
		"ᠠ",
		"ᠢ",
		"ᡅ",
		"ᡞ",
		"ᡳ",
		"ᱦ",
		"♆",
		"♕",
		"♛",
		"⟱",
		"⨄",
		"ʬ",
		"ѡ",
		"Ѿ",
		"ѿ",
		"Ԝ",
		"W",
		"ꔲ",
		"ꛃ",
		"ꝡ",
		"ꟽ",
		"ꤿ",
		"ꪟ",
		"ꮗ",
		"ꮚ",
		"ꮤ",
		"ꮿ",
		"௰",
		"ฝ",
		"ฟ",
		"ພ",
		"ຟ",
		"ཡ",
		"￦",
		"Ꮙ",
		"ᐫ",
		"ᔑ",
		"ᗯ",
		"ᗻ",
		"ᘈ",
		"ᙔ",
		"ᙛ",
		"ᙧ",
		"Ѡ",
		"Ѽ",
		"ѽ",
		"ש",
		"𝕎",
		"𝚆",
		"𝐖",
		"𝐰",
		"𝑊",
		"𝑾",
		"𝒘",
		"𝒲",
		"𝓌",
		"𝔀",
		"𝔚",
		"𝔴",
		"𝖶",
		"𝗐",
		"𝗪",
		"𝘄",
		"𝘞",
		"𝘸",
		"𝙬",
		"𝚠",
		"𝛚",
		"𝛡",
		"𝞈",
		"𝟂",
		"𝟉",
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
		"⒱⒱",
		"ᐁᐁ",
		"∀∀",
		"∇∇",
		"⊽⊽",
		"⋎⋎"}
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
		"ɯ",
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
		"⸷",
		"‡",
		"†",
		"🄣",
		"⒯",
		"ቲ",
		"ፐ",
		"ፒ",
		"ፔ",
		"Ꭲ",
		"Ꮏ",
		"ᝨ",
		"ƫ",
		"Ƭ",
		"ᴛ",
		"₸",
		"ℸ",
		"⍑",
		"⍡",
		"Ⱦ",
		"⤒",
		"⫟",
		"⫪",
		"Ⲧ",
		"ⲧ",
		"ⴕ",
		"ㅜ",
		"ͳ",
		"Ҭ",
		"ҭ",
		"T",
		"ד",
		"ߟ",
		"फ",
		"ꃌ",
		"꓅",
		"ꓔ",
		"ꔋ",
		"ꕛ",
		"Ꚍ",
		"Ꚑ",
		"ꛙ",
		"ꭲ",
		"ꮦ",
		"ﬢ",
		"ｔ",
		"ｾ",
		"ﾃ",
		"ﾅ",
		"ﾓ",
		"ￓ",
		"ቸ",
		"𝚃",
		"𝐓",
		"𝐭",
		"𝑇",
		"𝑡",
		"𝑻",
		"𝒕",
		"𝒯",
		"𝓉",
		"𝓣",
		"𝓽",
		"𝔗",
		"𝔱",
		"𝕥",
		"𝖳",
		"𝗍",
		"𝗧",
		"𝘁",
		"𝘛",
		"𝘵",
		"𝙏",
		"𝙩",
		"𝚝",
		"𝚻",
		"𝛕",
		"𝛵",
		"𝜏",
		"𝜯",
		"𝝉",
		"𝝩",
		"𝞃",
		"𝞣",
		"𝞽"}
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
		"ȸ",
		"Ғ",
		"£",
		"⨚",
		"⨑",
		"⨍",
		"🇫",
		"℉",
		"🄕",
		"⒡",
		"ɟ",
		"ᖴ",
		"F",
		"ᶲ",
		"ፑ",
		"Ŧ",
		"ғ",
		"ߓ",
		"ꈭ",
		"ꊰ",
		"ꓝ",
		"ꘘ",
		"ꝭ",
		"ቀ",
		"𝔽",
		"𝙵",
		"𝐅",
		"𝐟",
		"𝐹",
		"𝑓",
		"𝑭",
		"𝒇",
		"𝒥",
		"𝒻",
		"𝓕",
		"𝔉",
		"𝖥",
		"𝖿",
		"𝗙",
		"𝗳",
		"𝘍",
		"𝘧",
		"𝙁",
		"𝙛",
		"𝚏",
		"𝛗",
		"𝚽",
		"𝛟",
		"𝜑",
		"𝛷",
		"𝜙",
		"𝝋",
		"𝜱",
		"𝝓",
		"𝞅",
		"𝝫",
		"𝞍",
		"𝞿",
		"𝞥",
		"𝟇",
		"𝟊",
		"𝟋"}
	repl["!"] = []string{
		"i",
		"1",
		"１",
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
		"⁉",
		"‼"}
	repl["?"] = []string{
		"7",
		"７",
		"❔",
		"❓",
		"⍰",
		"؟",
		"⸮",
		"¿",
		"︖",
		"﹖",
		"？",
		"⁇",
		"⁈",
		"‽",
		"ʔ",
		"ʡ",
		"܊",
		"ॽ",
		"ɂ",
		"⫀",
		"⫂",
		"ꛫ",
		"꜅"}
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

// removeUnicodeDiacreticAnalog replace diacretic characters that looks like "w","t","f","!","?" with their ASCII representation
// replace only characters that removes by removeUnicodeDiacreticAnalog function
func (w *WTFSteroidChecker) removeUnicodeDiacreticAnalog() {
	replaceMap := w.WTFUnicodeDiacreticLibrary()
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
	w.removeUnicodeDiacreticAnalog()
	w.removeDiacretic()
	w.removeUnicodeAnalog()
	w.removeNotASCIIAndNotRussian()
	w.removeNotLetters()

	// Straight and reverse order
	return contains([]string{"wtf!", "wtf?"}, w.message) || contains([]string{"!ftw", "?ftw"}, w.message)

}
