package main

import (
	"strconv"
	"testing"
)

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
	matched, _ := searchMatch(base, pattern)
	if matched != expected {
		t.Errorf("False search for" +
			" ignorecase = " + strconv.FormatBool(gOpts.ignorecase) + ", " +
			" smartcase = " + strconv.FormatBool(gOpts.smartcase) + ", " +
			" ignoredia = " + strconv.FormatBool(gOpts.ignoredia) + ", " +
			" smartdia = " + strconv.FormatBool(gOpts.smartdia) + ", ")
	}
}
