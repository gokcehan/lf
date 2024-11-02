package main

import (
	"strconv"
	"testing"
	"unicode"
)

func Test_normMap(t *testing.T) {
	t.Run("european", func(t *testing.T) {
		europeanBase := []rune("ěřůøĉĝĥĵŝŭèùÿėįųāēīūļķņģőűëïąćęłńśźżõșțčďĺľňŕšťýžéíñóúüåäöçîşûğăâđêôơưáàãảạ")
		europeanNorm := []rune("eruocghjsueuyeiuaeiulkngoueiacelnszzostcdllnrstyzeinouuaaocisugaadeoouaaaaa")

		t.Run("lowercase", func(t *testing.T) {
			for i := range europeanBase {
				if normMap[europeanBase[i]] != europeanNorm[i] {
					t.Errorf("at input '%c' expected '%c' but got '%c'", europeanBase[i], europeanNorm[i], normMap[europeanBase[i]])
				}
			}
		})

		t.Run("uppercase", func(t *testing.T) {
			for i := range europeanBase {
				if normMap[unicode.ToUpper(europeanBase[i])] != unicode.ToUpper(europeanNorm[i]) {
					t.Errorf("at input '%c' expected '%c' but got '%c'",
						unicode.ToUpper(europeanBase[i]), unicode.ToUpper(europeanNorm[i]), unicode.ToUpper(normMap[europeanBase[i]]))
				}
			}
		})
	})

	t.Run("vietnamese", func(t *testing.T) {
		vietnameseBase := []rune("áạàảãăắặằẳẵâấậầẩẫéẹèẻẽêếệềểễiíịìỉĩoóọòỏõôốộồổỗơớợờởỡúụùủũưứựừửữyýỵỳỷỹđ")
		vietnameseNorm := []rune("aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiioooooooooooooooooouuuuuuuuuuuyyyyyyd")

		t.Run("lowercase", func(t *testing.T) {
			for i := range vietnameseBase {
				if normMap[vietnameseBase[i]] != vietnameseNorm[i] {
					t.Errorf("at input '%c' expected '%c' but got '%c'", vietnameseBase[i], vietnameseNorm[i], normMap[vietnameseBase[i]])
				}
			}
		})

		t.Run("uppercase", func(t *testing.T) {
			for i := range vietnameseBase {
				if unicode.ToUpper(normMap[vietnameseBase[i]]) != unicode.ToUpper(vietnameseNorm[i]) {
					t.Errorf("at input '%c' expected '%c' but got '%c'",
						unicode.ToUpper(vietnameseBase[i]), unicode.ToUpper(vietnameseNorm[i]), unicode.ToUpper(normMap[vietnameseBase[i]]))
				}
			}
		})
	})
}

// typical czech test sentence ;-)
const baseTestString = "Příliš žluťoučký kůň příšerně úpěl ďábelské ódy"

func TestRemoveDiacritics(t *testing.T) {
	testStr := baseTestString
	expStr := "Prilis zlutoucky kun priserne upel dabelske ody"
	checkRemoveDiacritics(testStr, expStr, t)

	// other accents (non comlete, but all I founded)
	testStr = "áéíóúýčďěňřšťžůåøĉĝĥĵŝŭšžõäöüàâçéèêëîïôùûüÿžščćđáéíóúąęėįųūčšžāēīūčšžļķņģáéíóúöüőűäöüëïąćęłńóśźżáàãâçéêíóõôăâîșțáäčďéíĺľňóôŕšťúýžáéíñóúüåäöâçîşûğăâđêôơưáàãảạ"
	expStr = "aeiouycdenrstzuaocghjsuszoaouaaceeeeiiouuuyzsccdaeiouaeeiuucszaeiucszlkngaeiouououaoueiacelnoszzaaaaceeioooaaistaacdeillnoorstuyzaeinouuaaoacisugaadeoouaaaaa"
	checkRemoveDiacritics(testStr, expStr, t)

	testStr = "ÁÉÍÓÚÝČĎĚŇŘŠŤŽŮÅØĈĜĤĴŜŬŠŽÕÄÖÜÀÂÇÉÈÊËÎÏÔÙÛÜŸŽŠČĆĐÁÉÍÓÚĄĘĖĮŲŪČŠŽĀĒĪŪČŠŽĻĶŅĢÁÉÍÓÚÖÜŐŰÄÖÜËÏĄĆĘŁŃÓŚŹŻÁÀÃÂÇÉÊÍÓÕÔĂÂÎȘȚÁÄČĎÉÍĹĽŇÓÔŔŠŤÚÝŽÁÉÍÑÓÚÜÅÄÖÂÇÎŞÛĞĂÂĐÊÔƠƯÁÀÃẢẠ"
	expStr = "AEIOUYCDENRSTZUAOCGHJSUSZOAOUAACEEEEIIOUUUYZSCCDAEIOUAEEIUUCSZAEIUCSZLKNGAEIOUOUOUAOUEIACELNOSZZAAAACEEIOOOAAISTAACDEILLNOORSTUYZAEINOUUAAOACISUGAADEOOUAAAAA"
	checkRemoveDiacritics(testStr, expStr, t)

	testStr = "áạàảãăắặằẳẵâấậầẩẫéẹèẻẽêếệềểễiíịìỉĩoóọòỏõôốộồổỗơớợờởỡúụùủũưứựừửữyýỵỳỷỹđ"
	expStr = "aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiioooooooooooooooooouuuuuuuuuuuyyyyyyd"
	checkRemoveDiacritics(testStr, expStr, t)

	testStr = "ÁẠÀẢÃĂẮẶẰẲẴÂẤẬẦẨẪÉẸÈẺẼÊẾỆỀỂỄÍỊÌỈĨÓỌÒỎÕÔỐỘỒỔỖƠỚỢỜỞỠÚỤÙỦŨƯỨỰỪỬỮÝỴỲỶỸĐ"
	expStr = "AAAAAAAAAAAAAAAAAEEEEEEEEEEEIIIIIOOOOOOOOOOOOOOOOOUUUUUUUUUUUYYYYYD"
	checkRemoveDiacritics(testStr, expStr, t)
}

func checkRemoveDiacritics(testStr string, expStr string, t *testing.T) {
	resultStr := removeDiacritics(testStr)
	if resultStr != expStr {
		t.Errorf("at input '%v' expected '%v' but got '%v'", testStr, expStr, resultStr)
	}
}

func TestSearchSettings(t *testing.T) {
	runSearch(t, true, false, true, true, "Veřejný", "vere", true)

	runSearch(t, true, false, true, false, baseTestString, "Zlutoucky", true)
	runSearch(t, true, false, true, false, baseTestString, "zlutoucky", true)
	runSearch(t, true, true, true, false, baseTestString, "Zlutoucky", false)
	runSearch(t, true, true, true, true, baseTestString, "zlutoucky", true)

	runSearch(t, false, false, true, false, baseTestString, "žlutoucky", true)
	runSearch(t, false, false, true, false, baseTestString, "Žlutoucky", false)

	runSearch(t, false, false, true, true, baseTestString, "žluťoučký", true)
	runSearch(t, false, false, true, false, baseTestString, "žluťoučký", true)
	runSearch(t, false, false, false, false, baseTestString, "žluťoučký", true)
	runSearch(t, false, false, false, false, baseTestString, "zlutoucky", false)
	runSearch(t, false, false, true, true, baseTestString, "zlutoucky", true)
}

func runSearch(t *testing.T, ignorecase, smartcase, ignorediacritics, smartdiacritics bool, base, pattern string, expected bool) {
	gOpts.ignorecase = ignorecase
	gOpts.smartcase = smartcase
	gOpts.ignoredia = ignorediacritics
	gOpts.smartdia = smartdiacritics
	matched, _ := searchMatch(base, pattern, false)
	if matched != expected {
		t.Errorf("False search for" +
			" ignorecase = " + strconv.FormatBool(gOpts.ignorecase) + ", " +
			" smartcase = " + strconv.FormatBool(gOpts.smartcase) + ", " +
			" ignoredia = " + strconv.FormatBool(gOpts.ignoredia) + ", " +
			" smartdia = " + strconv.FormatBool(gOpts.smartdia) + ", ")
	}
}
