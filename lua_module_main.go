package main

import (
	"fmt"
	"path/filepath"
	"slices"

	lua "github.com/yuin/gopher-lua"
)

func LfMainModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), LfMainModuleExports)

	setupModuleConstants(L, mod)

	L.Push(mod)

	return 1
}

var LfMainModuleExports = map[string]lua.LGFunction{
	"glob_match": luaGlobMatch,

	"match_word": luaLuaMatchWord,
}

func setupModuleConstants(L *lua.LState, mod *lua.LTable) {
	mod.RawSetString("REGISTRY_SORT_METHOD", lua.LString(registryKeySortMethod))
	mod.RawSetString("REGISTRY_COMMAND", lua.LString(registryKeyCommand))
	mod.RawSetString("REGISTRY_EVENT_HOOK", lua.LString(registryKeyEventHook))
	mod.RawSetString("REGISTRY_PREVIEWER", lua.LString(registryKeyPreviewer))

	eventType := L.NewTable()
	eventType.RawSetString("PreCd", lua.LString("pre-cd"))
	eventType.RawSetString("OnCd", lua.LString("on-cd"))
	eventType.RawSetString("OnLoad", lua.LString("on-load"))
	eventType.RawSetString("OnFocus-gained", lua.LString("on-focus-gained"))
	eventType.RawSetString("OnFocus-lost", lua.LString("on-focus-lost"))
	eventType.RawSetString("OnInit", lua.LString("on-init"))
	eventType.RawSetString("OnRedraw", lua.LString("on-redraw"))
	eventType.RawSetString("OnSelect", lua.LString("on-select"))
	eventType.RawSetString("OnQuit", lua.LString("on-quit"))
	mod.RawSetString("EventType", eventType)
}

// ----------------------------------------------------------------------------

const LuaCompMatchTypeName = "lf.comp_match"

func LRegisterCompMatchType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaCompMatchTypeName)

	L.SetFuncs(mt, luaCompMatchStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaCompMatchMethods))

	return mt
}

func LCheckCompMatch(L *lua.LState, index int) *compMatch {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*compMatch); ok {
		return v
	}

	L.ArgError(index, "value of type `CompMatch` expected")

	return nil
}

func LWrapCompMatch(L *lua.LState, data *compMatch) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaCompMatchTypeName))

	return ud
}

func LAddCompMatchToState(L *lua.LState, data *compMatch) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapCompMatch(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaCompMatchStaticMethod = map[string]lua.LGFunction{
	"new": luaCompMatchNew,
}

func luaCompMatchNew(L *lua.LState) int {
	name := L.CheckString(1)
	result := L.CheckString(2)
	return LAddCompMatchToState(L, &compMatch{name: name, result: result})
}

// ----------------------------------------------------------------------------

var luaCompMatchMethods = map[string]lua.LGFunction{
	"name":   luaCompMatchName,
	"result": luaCompMatchResult,
}

func luaCompMatchName(L *lua.LState) int {
	cm := LCheckCompMatch(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		cm.name = value
	}

	L.Push(lua.LString(cm.name))

	return 1
}

func luaCompMatchResult(L *lua.LState) int {
	cm := LCheckCompMatch(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		cm.result = value
	}

	L.Push(lua.LString(cm.result))

	return 1
}

// ----------------------------------------------------------------------------

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

func luaLuaMatchWord(L *lua.LState) int {
	longest := L.CheckString(1)
	wordTbl := L.CheckTable(2)

	nWord := wordTbl.Len()
	words := make([]string, nWord)

	for i := 1; i <= nWord; i++ {
		word := wordTbl.RawGetInt(i)
		words[i-1] = word.String()
	}

	slices.Sort(words)
	matches, longest := matchWord(longest, slices.Compact(words))

	tbl := L.NewTable()
	for _, match := range matches {
		tbl.Append(LWrapCompMatch(L, &match))
	}

	L.Push(tbl)
	L.Push(lua.LString(longest))

	return 2
}
