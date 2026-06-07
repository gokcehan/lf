package main

import lua "github.com/yuin/gopher-lua"

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
