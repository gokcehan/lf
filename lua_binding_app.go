package main

import (
	lua "github.com/yuin/gopher-lua"
)

const LuaAppTypeName = "lf.app"

func LRegisterAppType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaAppTypeName)

	L.SetFuncs(mt, luaAppStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaAppMethods))

	return mt
}

func LCheckApp(L *lua.LState, index int) *app {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*app); ok {
		return v
	}

	L.ArgError(index, "value of type `App` expected")

	return nil
}

func LWrapApp(L *lua.LState, data *app) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaAppTypeName))

	return ud
}

func LAddAppToState(L *lua.LState, data *app) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapApp(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaAppStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaAppMethods = map[string]lua.LGFunction{
	"ui": luaAppUI,
}

func luaAppUI(L *lua.LState) int {
	app := LCheckApp(L, 1)
	return LAddUIToState(L, app.ui)
}
