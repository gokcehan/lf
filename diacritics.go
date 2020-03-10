package main

import (
	"strconv"
	"unicode"
)

var normMap map[rune]rune

func init() {
	normMap = make(map[rune]rune)

	// (not only) european
	appendTransliterate(
		"ěřůøĉĝĥĵŝŭèùÿėįųāēīūļķņģőűëïąćęłńśźżõșțčďĺľňŕšťýžéíñóúüåäöçîşûğăâđêôơưáàãảạ",
		"eruocghjsueuyeiuaeiulkngoueiacelnszzostcdllnrstyzeinouuaaocisugaadeoouaaaaa",
	)

	// Vietnamese
	appendTransliterate(
		"áạàảãăắặằẳẵâấậầẩẫéẹèẻẽêếệềểễiíịìỉĩoóọòỏõôốộồổỗơớợờởỡúụùủũưứựừửữyýỵỳỷỹđ",
		"aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiioooooooooooooooooouuuuuuuuuuuyyyyyyd",
	)
}

func appendTransliterate(base, norm string) {
	normRunes := []rune(norm)
	baseRunes := []rune(base)

	lenNorm := len(normRunes)
	lenBase := len(baseRunes)
	if lenNorm != lenBase {
		panic("Base and normalized strings have differend length: base=" + strconv.Itoa(lenBase) + ", norm=" + strconv.Itoa(lenNorm)) // programmer error in constant length
	}

	for i := 0; i < lenBase; i++ {
		normMap[baseRunes[i]] = normRunes[i]

		baseUpper := unicode.ToUpper(baseRunes[i])
		normUpper := unicode.ToUpper(normRunes[i])

		normMap[baseUpper] = normUpper
	}
}

// Remove diacritics and make lowercase.
func removeDiacritics(baseString string) string {
	var normalizedRunes []rune
	for _, baseRune := range baseString {
		if normRune, ok := normMap[baseRune]; ok {
			normalizedRunes = append(normalizedRunes, normRune)
		} else {
			normalizedRunes = append(normalizedRunes, baseRune)
		}
	}
	return string(normalizedRunes)
}
