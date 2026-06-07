package main

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type file

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

var luaFileStaticMethod = map[string]lua.LGFunction{
	"new": luaFileNew,
}

func luaFileNew(L *lua.LState) int {
	path := L.CheckString(1)
	file := newFile(path)
	return LAddFileToState(L, file)
}

// ----------------------------------------------------------------------------

var luaFileMethods = map[string]lua.LGFunction{
	"name":   luaFileName,
	"size":   luaFileSize,
	"mode":   luaFileMode,
	"is_dir": luaFileIsDir,

	"link_state":  luaFileLinkState,
	"link_target": luaFileLinkTarget,
	"path":        luaFilePath,

	"dir_count": luaFileDirCount,
	"dir_size":  luaFileDirSize,

	"custom_info": luaFileCustomInfo,
	"ext":         luaFileExt,

	"extra_info": luaFileExtraInfo,
}

func luaFileName(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LString(file.Name()))
	return 1
}

func luaFileSize(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LNumber(file.Size()))
	return 1
}

func luaFileMode(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LNumber(file.Mode()))
	return 1
}

func luaFileIsDir(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LBool(file.IsDir()))
	return 1
}

func luaFileLinkState(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckNumber(2)
		file.linkState = linkState(value)
	}

	L.Push(lua.LNumber(file.linkState))

	return 1
}

func luaFileLinkTarget(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.linkTarget = value
	}

	L.Push(lua.LString(file.linkTarget))

	return 1
}

func luaFilePath(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.path = value
	}

	L.Push(lua.LString(file.path))

	return 1
}

func luaFileDirCount(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckNumber(2)
		file.dirCount = int(value)
	}

	L.Push(lua.LNumber(file.dirCount))

	return 1
}

func luaFileDirSize(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckNumber(2)
		file.dirSize = int64(value)
	}

	L.Push(lua.LNumber(file.dirSize))

	return 1
}

func luaFileCustomInfo(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.customInfo = value
	}

	L.Push(lua.LString(file.customInfo))

	return 1
}

func luaFileExt(L *lua.LState) int {
	file := LCheckFile(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		file.ext = value
	}

	L.Push(lua.LString(file.ext))

	return 1
}

func luaFileIsPreviewable(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LBool(file.isPreviewable()))
	return 1
}

func luaFileExtraInfo(L *lua.LState) int {
	file := LCheckFile(L, 1)
	key := L.Get(2)

	value := lua.LNil
	nargs := L.GetTop()
	if nargs >= 3 {
		value := L.Get(3)

		if file.luaExtraInfo == nil {
			file.luaExtraInfo = L.NewTable()
		}
		file.luaExtraInfo.RawSet(key, value)
	} else {
		if file.luaExtraInfo != nil {
			value = file.luaExtraInfo.RawGet(key)
		}
	}

	L.Push(value)

	return 1
}

// ----------------------------------------------------------------------------
// Type dir

const LuaDirTypeName = "lf.dir"

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

// ----------------------------------------------------------------------------
// Type nav

const LuaNavTypeName = "lf.nav"

func LRegisterNavType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaNavTypeName)

	L.SetFuncs(mt, luaNavStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaNavMethods))

	return mt
}

func LCheckNav(L *lua.LState, index int) *nav {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*nav); ok {
		return v
	}

	L.ArgError(index, "value of type `Nav` expected")

	return nil
}

func LWrapNav(L *lua.LState, data *nav) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaNavTypeName))

	return ud
}

func LAddNavToState(L *lua.LState, data *nav) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapNav(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

var luaNavStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaNavMethods = map[string]lua.LGFunction{
	"get_tag": luaNavGetTag,

	"select":               luaNavSelect,
	"toggle_selection":     luaNavToggleSelection,
	"toggle":               luaNavToggle,
	"tag_toggle_selection": luaNavTagToggleSelection,
	"tag_toggle":           luaNavTagToggle,
	"invert":               luaNavInvert,
	"unselect":             luaNavUnselect,
	"glob_sel":             luaNavGlobSel,

	"curr_dir": luaNavCurrDir,
}

func luaNavGetTag(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)

	value, exists := nav.tags[path]
	if !exists {
		return 0
	}

	L.Push(lua.LString(value))

	return 1
}

func luaNavSelect(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)

	nav.selections[path] = nav.selectionInd
	nav.selectionInd++

	return 0
}

func luaNavToggleSelection(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)
	nav.toggleSelection(path)
	return 0
}

func luaNavToggle(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.toggle()
	return 0
}

func luaNavTagToggleSelection(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)
	tag := L.CheckString(3)
	nav.tagToggleSelection(path, tag)
	return 0
}

func luaNavTagToggle(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	tag := L.CheckString(2)
	nav.tagToggle(tag)
	return 0
}

func luaNavInvert(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.invert()
	return 0
}

func luaNavUnselect(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.unselect()
	return 0
}

func luaNavUnselectOne(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)

	if _, ok := nav.selections[path]; ok {
		delete(nav.selections, path)
		if len(nav.selections) == 0 {
			nav.selectionInd = 0
		}
	}

	return 0
}

func luaNavGlobSel(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	pattern := L.CheckString(2)
	invert := L.CheckBool(3)

	nav.globSel(pattern, invert)

	return 0
}

func luaNavCurrDir(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	return LAddDirToState(L, nav.currDir())
}
