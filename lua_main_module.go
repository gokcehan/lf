package main

import (
	"fmt"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

func LfMainModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)

	setupModuleConstants(L, mod)

	L.Push(mod)

	return 1
}

var exports = map[string]lua.LGFunction{
	"glob_match": luaGlobMatch,
}

func setupModuleConstants(L *lua.LState, mod *lua.LTable) {}

func luaGlobMatch(L *lua.LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)

	match, err := filepath.Match(pattern, str)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(fmt.Sprintf("glob match error: %s", err)))
		return 2
	}

	L.Push(lua.LBool(match))

	return 1
}
