package main

import (
	lua "github.com/yuin/gopher-lua"
)

func LfMainModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)

	setupModuleConstants(L, mod)

	L.Push(mod)

	return 1
}

var exports = map[string]lua.LGFunction{}

func setupModuleConstants(L *lua.LState, mod *lua.LTable) {}
