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
	"add_hook":             luaAppAddHook,
	"hook_pre_cd":          makeAppHookAdder("pre-cd"),
	"hook_on_cd":           makeAppHookAdder("on-cd"),
	"hook_on_load":         makeAppHookAdder("on-load"),
	"hook_on_focus_gained": makeAppHookAdder("on-focus-gained"),
	"hook_on_focus_lost":   makeAppHookAdder("on-focus-lost"),
	"hook_on_init":         makeAppHookAdder("on-init"),
	"hook_on_redraw":       makeAppHookAdder("on-redraw"),
	"hook_on_select":       makeAppHookAdder("on-select"),
	"hook_on_quit":         makeAppHookAdder("on-quit"),
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

	if gLuaRegistry.sortMethod == nil {
		gLuaRegistry.sortMethod = make(map[string]*lua.LFunction)
	}
	gLuaRegistry.sortMethod[name] = sortFunc

	return 0
}

func luaAppAddHook(L *lua.LState) int {
	// app := LCheckApp(L, 1)
	cmdName := L.CheckString(2)
	hookFunc := L.CheckFunction(3)

	if gLuaRegistry.eventHooks == nil {
		gLuaRegistry.eventHooks = make(map[string][]*lua.LFunction)
	}

	list := gLuaRegistry.eventHooks[cmdName]
	gLuaRegistry.eventHooks[cmdName] = append(list, hookFunc)

	return 0
}

func makeAppHookAdder(cmdName string) lua.LGFunction {
	return func(L *lua.LState) int {
		// app := LCheckApp(L, 1)
		hookFunc := L.CheckFunction(2)

		if gLuaRegistry.eventHooks == nil {
			gLuaRegistry.eventHooks = make(map[string][]*lua.LFunction)
		}

		list := gLuaRegistry.eventHooks[cmdName]
		gLuaRegistry.eventHooks[cmdName] = append(list, hookFunc)

		return 0
	}
}
