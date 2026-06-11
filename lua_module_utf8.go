package main

import (
	"unicode/utf8"

	lua "github.com/yuin/gopher-lua"
)

func LfUtf8ModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"to_rune_tbl": luaUtf8ToRuneTbl,
		"len":         luaUtf8Len,
		"get_rune":    luaUtf8GetRune,
	})

	L.Push(mod)

	return 1
}

// luaUtf8ToRuneTbl converts given string into a list of UTF-8 runes.
func luaUtf8ToRuneTbl(L *lua.LState) int {
	str := L.CheckString(1)
	runes := []rune(str)

	tbl := L.NewTable()
	for _, r := range runes {
		tbl.Append(lua.LString(string(r)))
	}

	L.Push(tbl)

	return 1
}

// luaUtf8Len returns length of a string counted in UTF-8 rune.
func luaUtf8Len(L *lua.LState) int {
	str := L.CheckString(1)
	length := utf8.RuneCountInString(str)

	L.Push(lua.LNumber(length))

	return 1
}

// luaUtf8GetRune returns UTF-8 rune at given index.
func luaUtf8GetRune(L *lua.LState) int {
	str := L.CheckString(1)
	index := L.CheckInt(2)

	runes := []rune(str)
	L.Push(lua.LString(string(runes[index-1])))

	return 1
}
