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

// ----------------------------------------------------------------------------
// type fs.DirEntry

const LuaDirEntryTypeName = "fs.DirEntry"

func lRegisterDirEntryType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaDirEntryTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":   luaDirEntryName,
		"info":   luaDirEntryInfo,
		"is_dir": luaDirEntryIsDir,
		"type":   luaDirEntryType,
	}))

	return mt
}

func lCheckDirEntry(L *lua.LState, index int) fs.DirEntry {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(fs.DirEntry); ok {
		return v
	}

	L.ArgError(index, "value of type `DirEntry` expected")

	return nil
}

func lWrapDirEntry(L *lua.LState, data fs.DirEntry) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaDirEntryTypeName))

	return ud
}

func lAddDirEntryToState(L *lua.LState, data fs.DirEntry) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapDirEntry(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaDirEntryName returns name of this entry.
func luaDirEntryName(L *lua.LState) int {
	entry := lCheckDirEntry(L, 1)
	L.Push(lua.LString(entry.Name()))
	return 1
}

// luaDirEntryInfo returns FileInfo of this entry.
func luaDirEntryInfo(L *lua.LState) int {
	entry := lCheckDirEntry(L, 1)
	info, err := entry.Info()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	return lAddFileInfoToState(L, info)
}

// luaDirEntryIsDir returns true if this entry is directory.
func luaDirEntryIsDir(L *lua.LState) int {
	entry := lCheckDirEntry(L, 1)
	L.Push(lua.LBool(entry.IsDir()))
	return 1
}

func luaDirEntryType(L *lua.LState) int {
	entry := lCheckDirEntry(L, 1)
	return lAddFileModeToState(L, entry.Type())
}

// ----------------------------------------------------------------------------
// type fs.FileMode

const LuaFileModeTypeName = "fs.FileMode"

func lRegisterFileModeType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaFileModeTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new":        luaFileModeNew,
		"__tostring": luaFileModeMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"is_dir":     luaFileModeIsDir,
		"is_regular": luaFileModeIsRegular,
		"perm":       luaFileModePerm,
		"type":       luaFileModeType,
	}))

	return mt
}

func lCheckFileMode(L *lua.LState, index int) fs.FileMode {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(fs.FileMode); ok {
		return v
	}

	L.ArgError(index, "value of type `FileMode` expected")

	return 0
}

func lWrapFileMode(L *lua.LState, data fs.FileMode) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaFileModeTypeName))

	return ud
}

func lAddFileModeToState(L *lua.LState, data fs.FileMode) int {
	ud := lWrapFileMode(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaFileModeNew(L *lua.LState) int {
	value := L.CheckInt(1)
	return lAddFileModeToState(L, fs.FileMode(value))
}

func luaFileModeMetaTostring(L *lua.LState) int {
	mode := lCheckFileMode(L, 1)
	L.Push(lua.LString(mode.String()))
	return 1
}

// ----------------------------------------------------------------------------

// luaFileModeIsDir returns true if current file mode is directory.
func luaFileModeIsDir(L *lua.LState) int {
	mode := lCheckFileMode(L, 1)
	L.Push(lua.LBool(mode.IsDir()))
	return 1
}

// luaFileModeIsRegular returns true if current file mode is regular file.
func luaFileModeIsRegular(L *lua.LState) int {
	mode := lCheckFileMode(L, 1)
	L.Push(lua.LBool(mode.IsRegular()))
	return 1
}

// luaFileModePerm returns unix permission bits in mode.
func luaFileModePerm(L *lua.LState) int {
	mode := lCheckFileMode(L, 1)
	return lAddFileModeToState(L, mode.Perm())
}

// luaFileModeType returns type bits in mode.
func luaFileModeType(L *lua.LState) int {
	mode := lCheckFileMode(L, 1)
	return lAddFileModeToState(L, mode.Type())
}
