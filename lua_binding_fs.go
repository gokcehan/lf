package main

import (
	"io/fs"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type fs.FileInfo

const luaFileInfoTypeName = "fs.FileInfo"

func lRegisterFileInfoType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaFileInfoTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":     luaFileInfoName,
		"size":     luaFileInfoSize,
		"mode":     luaFileInfoMode,
		"mod_time": luaFileInfoModTime,
		"is_dir":   luaFileInfoIsDir,
	}))

	return mt
}

func lCheckFileInfo(L *lua.LState, index int) fs.FileInfo {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(fs.FileInfo); ok {
		return v
	}

	L.ArgError(index, "value of type `FileInfo` expected")

	return nil
}

func lWrapFileInfo(L *lua.LState, data fs.FileInfo) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaFileInfoTypeName))

	return ud
}

func lAddFileInfoToState(L *lua.LState, data fs.FileInfo) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapFileInfo(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaFileInfoName(L *lua.LState) int {
	info := lCheckFileInfo(L, 1)
	L.Push(lua.LString(info.Name()))
	return 1
}

func luaFileInfoSize(L *lua.LState) int {
	info := lCheckFileInfo(L, 1)
	L.Push(lua.LNumber(info.Size()))
	return 1
}

func luaFileInfoMode(L *lua.LState) int {
	info := lCheckFileInfo(L, 1)
	L.Push(lua.LNumber(info.Mode()))
	return 1
}

func luaFileInfoModTime(L *lua.LState) int {
	info := lCheckFileInfo(L, 1)
	t := info.ModTime()
	return lAddTimeToState(L, &t)
}

func luaFileInfoIsDir(L *lua.LState) int {
	info := lCheckFileInfo(L, 1)
	L.Push(lua.LBool(info.IsDir()))
	return 1
}
