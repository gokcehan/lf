package main

import (
	"github.com/clipperhouse/displaywidth"
	lua "github.com/yuin/gopher-lua"
)

func lfUIModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"print_length":  lfUIModulePrintLength,
		"display_width": lfUIModuleDisplayWidth,
	})

	L.Push(mod)

	return 1
}

// lfUIModulePrintLength returns displayed width of string content in terminal cells.
//
// It ignores supported terminal control sequences and accounts for tab
// expansions using the `tabstop` option.
func lfUIModulePrintLength(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LNumber(printLength(str)))
	return 1
}

// lfUIModuleDisplayWidth calculates the display width of a string, by iterating
// over grapheme clusters in the string and summing their widths.
func lfUIModuleDisplayWidth(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LNumber(displaywidth.String(str)))
	return 1
}
