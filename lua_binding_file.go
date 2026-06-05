package main

import (
	lua "github.com/yuin/gopher-lua"
)

const FileTypeName = "lf.file"

func LRegisterFileTypeMt(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(FileTypeName)

	L.SetFuncs(mt, luaFileStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaFileMethods))

	addLinkStateConstantToMt(L, mt)

	return mt
}

func addLinkStateConstantToMt(L *lua.LState, tbl *lua.LTable) {
	L.SetField(tbl, "LinkStateNotLink", lua.LNumber(notLink))
	L.SetField(tbl, "LinkStateWorking", lua.LNumber(working))
	L.SetField(tbl, "LinkStateBroken", lua.LNumber(broken))
}

func LCheckFile(L *lua.LState, index int) *file {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*file); ok {
		return v
	}

	L.ArgError(index, "value of type `File` expected")

	return nil
}

func LWrapFile(L *lua.LState, data *file) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(FileTypeName))

	return ud
}

func LAddFileToState(L *lua.LState, data *file) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapFile(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaFileStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaFileMethods = map[string]lua.LGFunction{}

func luaFileLinkState(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckNumber(2)
		file.linkState = linkState(value)
		return 0
	}

	L.Push(lua.LNumber(file.linkState))

	return 1
}

func luaFileLinkTarget(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.linkTarget = value
		return 0
	}

	L.Push(lua.LString(file.linkTarget))

	return 1
}

func luaFilePath(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.path = value
		return 0
	}

	L.Push(lua.LString(file.path))

	return 1
}

func luaFileDirCount(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckNumber(2)
		file.dirCount = int(value)
		return 0
	}

	L.Push(lua.LNumber(file.dirCount))

	return 1
}

func luaFileDirSize(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckNumber(2)
		file.dirSize = int64(value)
		return 0
	}

	L.Push(lua.LNumber(file.dirSize))

	return 1
}

func luaFileCustomInfo(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.customInfo = value
		return 0
	}

	L.Push(lua.LString(file.customInfo))

	return 1
}

func luaFileExt(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.ext = value
		return 0
	}

	L.Push(lua.LString(file.ext))

	return 1
}
