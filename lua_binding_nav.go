package main

import (
	lua "github.com/yuin/gopher-lua"
)

const LuaNavTypeName = "lf.nav"

func LRegisterNavType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaNavTypeName)

	L.SetFuncs(mt, luaNavStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaNavMethods))

	return mt
}

func LCheckNav(L *lua.LState, index int) *nav {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*nav); ok {
		return v
	}

	L.ArgError(index, "value of type `Nav` expected")

	return nil
}

func LWrapNav(L *lua.LState, data *nav) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaNavTypeName))

	return ud
}

func LAddNavToState(L *lua.LState, data *nav) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapNav(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaNavStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaNavMethods = map[string]lua.LGFunction{
	"get_tag": luaNavGetTag,

	"select":               luaNavSelect,
	"toggle_selection":     luaNavToggleSelection,
	"toggle":               luaNavToggle,
	"tag_toggle_selection": luaNavTagToggleSelection,
	"tag_toggle":           luaNavTagToggle,
	"invert":               luaNavInvert,
	"unselect":             luaNavUnselect,

	"curr_dir": luaNavCurrDir,
}

func luaNavGetTag(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)

	value, exists := nav.tags[path]
	if !exists {
		return 0
	}

	L.Push(lua.LString(value))

	return 1
}

func luaNavSelect(L *lua.LState) int {
	nav := LCheckNav(L, 1)

	nargs := L.GetTop()
	for i := 2; i <= nargs; i++ {
		path := L.CheckString(i)
		nav.selections[path] = nav.selectionInd
		nav.selectionInd++
	}

	return 0
}

func luaNavToggleSelection(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)
	nav.toggleSelection(path)
	return 0
}

func luaNavToggle(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.toggle()
	return 0
}

func luaNavTagToggleSelection(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)
	tag := L.CheckString(3)
	nav.tagToggleSelection(path, tag)
	return 0
}

func luaNavTagToggle(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	tag := L.CheckString(2)
	nav.tagToggle(tag)
	return 0
}

func luaNavInvert(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.invert()
	return 0
}

func luaNavUnselect(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.unselect()
	return 0
}

func luaNavCurrDir(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	return LAddDirToState(L, nav.currDir())
}
