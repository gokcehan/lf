package main

import lua "github.com/yuin/gopher-lua"

// ----------------------------------------------------------------------------
// Type compMatch

const LuaCompMatchTypeName = "lf.comp_match"

func LRegisterCompMatchType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaCompMatchTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaCompMatchNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":   luaCompMatchName,
		"result": luaCompMatchResult,
	}))

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

func luaCompMatchNew(L *lua.LState) int {
	name := L.CheckString(1)
	result := L.CheckString(2)
	return LAddCompMatchToState(L, &compMatch{name: name, result: result})
}

// ----------------------------------------------------------------------------

// luaCompMatchName is getter & setter for name field. It's displayed text for
// this completion entry.
func luaCompMatchName(L *lua.LState) int {
	cm := LCheckCompMatch(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		cm.name = value
	}

	L.Push(lua.LString(cm.name))

	return 1
}

// luaCompMatchResult is getter & setter for result field. It's applied text used
// when this completion entry is picked.
func luaCompMatchResult(L *lua.LState) int {
	cm := LCheckCompMatch(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		cm.result = value
	}

	L.Push(lua.LString(cm.result))

	return 1
}
