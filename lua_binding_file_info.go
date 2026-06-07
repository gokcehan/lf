package main

import (
	"io/fs"

	lua "github.com/yuin/gopher-lua"
)

const LuaFileInfoTypeName = "lf.file_info"

func LRegisterFileInfoType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaFileInfoTypeName)

	L.SetFuncs(mt, luaFileInfoStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaFileInfoMethods))

	return mt
}

func LCheckFileInfo(L *lua.LState, index int) fs.FileInfo {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(fs.FileInfo); ok {
		return v
	}

	L.ArgError(index, "value of type `FileInfo` expected")

	return nil
}

func LWrapFileInfo(L *lua.LState, data fs.FileInfo) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaFileInfoTypeName))

	return ud
}

func LAddFileInfoToState(L *lua.LState, data fs.FileInfo) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapFileInfo(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaFileInfoStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaFileInfoMethods = map[string]lua.LGFunction{
	"name":   luaFileInfoName,
	"size":   luaFileInfoSize,
	"mode":   luaFileInfoMode,
	"is_dir": luaFileInfoIsDir,
}

func luaFileInfoName(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LString(info.Name()))
	return 1
}

func luaFileInfoSize(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LNumber(info.Size()))
	return 1
}

func luaFileInfoMode(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LNumber(info.Mode()))
	return 1
}

func luaFileInfoIsDir(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LBool(info.IsDir()))
	return 1
}
