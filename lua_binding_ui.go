package main

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type ui

const LuaUITypeName = "lf.ui"

func LRegisterUIType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaUITypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"echo":     luaUIEcho,
		"echomsg":  luaUIEchoMsg,
		"echoerr":  luaUIEchhoErr,
		"echoerrf": luaUIEchhoErrf,
	}))

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

// luaUIEcho prints content to lf message bar.
func luaUIEcho(L *lua.LState) int {
	ui := LCheckUI(L, 1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	ui.exprChan <- &callExpr{"echo", args, 1}

	return 0
}

// luaUIEcho prints content to both lf message bar and log.
func luaUIEchoMsg(L *lua.LState) int {
	ui := LCheckUI(L, 1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	ui.exprChan <- &callExpr{"echomsg", args, 1}

	return 0
}

// luaUIEcho prints error message to both lf message bar and log.
func luaUIEchhoErr(L *lua.LState) int {
	ui := LCheckUI(L, 1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	ui.exprChan <- &callExpr{"echoerr", args, 1}

	return 0
}

// luaUIEcho prints error message with formatting string.
func luaUIEchhoErrf(L *lua.LState) int {
	ui := LCheckUI(L, 1)
	fmtStr := L.ToString(2)

	st := 3
	nArgs := L.GetTop()
	args := make([]any, nArgs-st+1)
	for i := 3; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	msg := fmt.Sprintf(fmtStr, args...)
	ui.exprChan <- &callExpr{"echoerr", []string{msg}, 1}

	return 0
}
