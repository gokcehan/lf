package main

import (
	"strings"

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
	"ui":  luaAppUI,
	"nav": luaAppNav,

	"create_cmd":           luaAppCreateCmd,
	"register_sort_method": luaAppRegisterSortMethod,
}

func luaAppUI(L *lua.LState) int {
	app := LCheckApp(L, 1)
	return LAddUIToState(L, app.ui)
}

func luaAppNav(L *lua.LState) int {
	app := LCheckApp(L, 1)
	return LAddNavToState(L, app.nav)
}

func luaAppCreateCmd(L *lua.LState) int {
	app := LCheckApp(L, 1)
	name := L.CheckString(2)
	action := L.Get(3)

	switch action.Type() {
	case lua.LTString:
		text := action.String()
		p := newParser(strings.NewReader(text))
		expr := p.parseExpr()
		if expr == nil {
			app.ui.echoerrf("failed to parse Lua command: %s", text)
		} else {
			gOpts.cmds[name] = &luaCmdExpr{
				name: name,
				expr: expr,
			}
		}
	case lua.LTFunction:
		gOpts.cmds[name] = &luaCmdExpr{
			name:    name,
			luaFunc: action.(*lua.LFunction),
		}
	default:
		L.ArgError(2, "string or function expected")
	}

	return 0
}

func luaAppRegisterSortMethod(L *lua.LState) int {
	// app := LCheckApp(L, 1)
	name := L.CheckString(2)
	sortFunc := L.CheckFunction(3)

	if gOpts.luaSortMethod == nil {
		gOpts.luaSortMethod = make(map[string]*lua.LFunction)
	}
	gOpts.luaSortMethod[name] = sortFunc

	return 0
}
