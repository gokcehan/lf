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
	"readdir":    luaReadDir,
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

func luaReadDir(L *lua.LState) int {
	path := L.CheckString(1)

	files, err := readdir(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	tbl := L.NewTable()
	for _, f := range files {
		tbl.Append(LWrapFile(L, f))
	}

	L.Push(tbl)

	return 1
}
