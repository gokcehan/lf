package main

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

const LuaDirTypeName = ""

func LRegisterDirType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaDirTypeName)

	L.SetFuncs(mt, luaDirStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaDirMethods))

	return mt
}

func LCheckDir(L *lua.LState, index int) *dir {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*dir); ok {
		return v
	}

	L.ArgError(index, "value of type `Dir` expected")

	return nil
}

func LWrapDir(L *lua.LState, data *dir) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaDirTypeName))

	return ud
}

func LAddDirToState(L *lua.LState, data *dir) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapDir(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaDirStaticMethod = map[string]lua.LGFunction{
	"new": luaDirNew,
}

func luaDirNew(L *lua.LState) int {
	path := L.CheckString(1)
	dir := newDir(path)
	return LAddDirToState(L, dir)
}

// ----------------------------------------------------------------------------

var luaDirMethods = map[string]lua.LGFunction{
	"files_for_each":     luaDirFilesForEach,
	"all_files_for_each": luaDirAllFilesForEach,
}

func luaDirFilesForEach(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	fn := L.CheckFunction(2)

	for i, file := range dir.files {
		err := L.CallByParam(
			lua.P{
				Fn:      fn,
				NRet:    0,
				Protect: true,
			},
			lua.LNumber(i),
			LWrapFile(L, file),
		)
		if err != nil {
			log.Printf("error during iteration : %s", err)
		}
	}

	return 0
}

func luaDirAllFilesForEach(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	fn := L.CheckFunction(2)

	for i, file := range dir.allFiles {
		err := L.CallByParam(
			lua.P{
				Fn:      fn,
				NRet:    0,
				Protect: true,
			},
			lua.LNumber(i),
			LWrapFile(L, file),
		)
		if err != nil {
			log.Printf("error during iteration : %s", err)
		}
	}

	return 0
}

func luaDirSort(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	dir.sort()
	return 0
}

func luaDirName(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LString(dir.name()))
	return 1
}
