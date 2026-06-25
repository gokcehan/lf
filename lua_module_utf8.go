package main

import (
	"unicode/utf8"

	lua "github.com/yuin/gopher-lua"
)

func lfUtf8ModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"to_rune_tbl": luaUtf8ModuleToRuneTbl,
		"len":         luaUtf8ModuleLen,
		"get_rune":    luaUtf8ModuleGetRune,
	})

	L.Push(mod)

	return 1
}

// luaUtf8ModuleToRuneTbl converts given string into a list of UTF-8 runes.
func luaUtf8ModuleToRuneTbl(L *lua.LState) int {
	str := L.CheckString(1)

	tbl := L.NewTable()
	for _, r := range str {
		tbl.Append(lua.LString(string(r)))
	}

	L.Push(tbl)

	return 1
}

// luaUtf8ModuleLen returns length of a string counted in UTF-8 rune.
func luaUtf8ModuleLen(L *lua.LState) int {
	str := L.CheckString(1)
	length := utf8.RuneCountInString(str)

	L.Push(lua.LNumber(length))

	return 1
}

// luaUtf8ModuleGetRune returns UTF-8 rune at given index.
func luaUtf8ModuleGetRune(L *lua.LState) int {
	str := L.CheckString(1)
	index := L.CheckInt(2)

	runes := []rune(str)
	L.Push(lua.LString(string(runes[index-1])))

	return 1
}
