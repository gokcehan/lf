package main

import (
	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type ui

const LuaUITypeName = "lf.ui"

func LRegisterUIType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaUITypeName)

	L.SetFuncs(mt, luaUIStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaUIMethods))

	return mt
}

func LCheckUI(L *lua.LState, index int) *ui {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*ui); ok {
		return v
	}

	L.ArgError(index, "value of type `UI` expected")

	return nil
}

func LWrapUI(L *lua.LState, data *ui) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaUITypeName))

	return ud
}

func LAddUIToState(L *lua.LState, data *ui) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapUI(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaUIStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaUIMethods = map[string]lua.LGFunction{
	"echo":     luaUIEcho,
	"echomsg":  luaUIEchoMsg,
	"echoerr":  luaUIEchhoErr,
	"echoerrf": luaUIEchhoErrf,
}

func luaUIEcho(L *lua.LState) int {
	ui := LCheckUI(L, 1)
	msg := L.ToString(2)

	ui.echo(msg)

	return 0
}

func luaUIEchoMsg(L *lua.LState) int {
	ui := LCheckUI(L, 1)
	msg := L.ToString(2)

	ui.echomsg(msg)

	return 0
}

func luaUIEchhoErr(L *lua.LState) int {
	ui := LCheckUI(L, 1)
	msg := L.ToString(2)

	ui.echoerr(msg)

	return 0
}

func luaUIEchhoErrf(L *lua.LState) int {
	ui := LCheckUI(L, 1)
	msg := L.ToString(2)

	args := []any{}
	nargs := L.GetTop()
	for i := 3; i <= nargs; i++ {
		value := L.Get(i)
		args = append(args, value.String())
	}

	ui.echoerrf(msg, args...)

	return 0
}
