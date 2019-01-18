package main

import (
	"strconv"
	"strings"
)

var normMap map[rune]rune

func init() {
	normMap = make(map[rune]rune)

	// (not only) european
	appendTransliterate(
		"ěřůøĉĝĥĵŝŭèùÿėįųāēīūļķņģőűëïąćęłńśźżõșțčďĺľňŕšťýžéíñóúüåäöçîşûğăâđêôơưáàãảạ",
		"eruocghjsueuyeiuaeiulkngoueiacelnszzostcdllnrstyzeinouuaaocisugaadeoouaaaaa",
		true,
	)

	// Vietnamese
	appendTransliterate(
		"áạàảãăắặằẳẵâấậầẩẫéẹèẻẽêếệềểễiíịìỉĩoóọòỏõôốộồổỗơớợờởỡúụùủũưứựừửữyýỵỳỷỹđ",
		"aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiioooooooooooooooooouuuuuuuuuuuyyyyyyd",
		true,
	)

}

func appendTransliterate(base, norm string, uppercase bool) {
	normRunes := []rune(norm)
	baseRunes := []rune(base)

	lenNorm := len(normRunes)
	lenBase := len(baseRunes)
	if lenNorm != lenBase {
		panic("Base and normalized strings have differend lenght: base=" + strconv.Itoa(lenBase) + ", norm=" + strconv.Itoa(lenNorm)) // programmer error in constant lenght
	}

	for i, baseRune := range baseRunes {
		normMap[baseRune] = normRunes[i]
	}

	if uppercase {
		upperBase := strings.ToUpper(base)
		upperNorm := strings.ToUpper(norm)
		normRune := []rune(upperNorm)
		for i, baseRune := range []rune(upperBase) {
			normMap[baseRune] = normRune[i]
		}
	}
}

// Remove diacritics and make lowercase.
func removeDiacritics(baseString string) string {
	var normalizedRunes []rune
	for _, baseRune := range []rune(baseString) {
		if normRune, ok := normMap[baseRune]; ok {
			normalizedRunes = append(normalizedRunes, normRune)
		} else {
			normalizedRunes = append(normalizedRunes, baseRune)
		}
	}
	return string(normalizedRunes)
}
