package main

import (
	lua "github.com/yuin/gopher-lua"
)

func LfUIModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"print_length": lfUIModulePrintLength,
	})

	L.Push(mod)

	return 1
}

// lfUIModulePrintLength returns displayed width of string content.
func lfUIModulePrintLength(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LNumber(printLength(str)))
	return 1
}
