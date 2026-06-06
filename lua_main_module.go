package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func LfMainModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), LfMainModuleExports)

	L.Push(mod)

	return 1
}

var LfMainModuleExports = map[string]lua.LGFunction{
	"glob_match": luaGlobMatch,
	"readdir":    luaReadDir,

	"create_cmd":           luaCreateCmd,
	"register_sort_method": luaRegisterSortMethod,
	"add_hook":             luaAddHook,
	"hook_pre_cd":          makeHookAdder("pre-cd"),
	"hook_on_cd":           makeHookAdder("on-cd"),
	"hook_on_load":         makeHookAdder("on-load"),
	"hook_on_focus_gained": makeHookAdder("on-focus-gained"),
	"hook_on_focus_lost":   makeHookAdder("on-focus-lost"),
	"hook_on_init":         makeHookAdder("on-init"),
	"hook_on_redraw":       makeHookAdder("on-redraw"),
	"hook_on_select":       makeHookAdder("on-select"),
	"hook_on_quit":         makeHookAdder("on-quit"),
}

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

func luaCreateCmd(L *lua.LState) int {
	name := L.CheckString(1)
	action := L.Get(2)

	switch action.Type() {
	case lua.LTString:
		text := action.String()
		p := newParser(strings.NewReader(text))
		expr := p.parseExpr()
		if expr == nil {
			log.Printf("failed to parse Lua command: %s", text)
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

func luaRegisterSortMethod(L *lua.LState) int {
	name := L.CheckString(1)
	sortFunc := L.CheckFunction(2)

	if gLuaRegistry.sortMethod == nil {
		gLuaRegistry.sortMethod = make(map[string]*lua.LFunction)
	}
	gLuaRegistry.sortMethod[name] = sortFunc

	return 0
}

func luaAddHook(L *lua.LState) int {
	cmdName := L.CheckString(1)
	hookFunc := L.CheckFunction(2)

	if gLuaRegistry.eventHooks == nil {
		gLuaRegistry.eventHooks = make(map[string][]*lua.LFunction)
	}

	list := gLuaRegistry.eventHooks[cmdName]
	gLuaRegistry.eventHooks[cmdName] = append(list, hookFunc)

	return 0
}

func makeHookAdder(cmdName string) lua.LGFunction {
	return func(L *lua.LState) int {
		hookFunc := L.CheckFunction(1)

		if gLuaRegistry.eventHooks == nil {
			gLuaRegistry.eventHooks = make(map[string][]*lua.LFunction)
		}

		list := gLuaRegistry.eventHooks[cmdName]
		gLuaRegistry.eventHooks[cmdName] = append(list, hookFunc)

		return 0
	}
}
