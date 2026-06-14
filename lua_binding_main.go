package main

import (
	"fmt"
	"slices"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type app

const luaAppTypeName = "lf.app"

func lRegisterAppType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaAppTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"ui":  luaAppUI,
		"nav": luaAppNav,
	}))

	return mt
}

func lCheckApp(L *lua.LState, index int) *app {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*app); ok {
		return v
	}

	L.ArgError(index, "value of type `App` expected")

	return nil
}

func lWrapApp(L *lua.LState, data *app) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaAppTypeName))

	return ud
}

func lAddAppToState(L *lua.LState, data *app) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapApp(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaAppUI returns `ui` object hold by app
func luaAppUI(L *lua.LState) int {
	app := lCheckApp(L, 1)
	return lAddUIToState(L, app.ui)
}

// luaAppNav returns `nav` object hold by app
func luaAppNav(L *lua.LState) int {
	app := lCheckApp(L, 1)
	return lAddNavToState(L, app.nav)
}

// ----------------------------------------------------------------------------
// Type compMatch

const luaCompMatchTypeName = "lf.comp_match"

func lRegisterCompMatchType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaCompMatchTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaCompMatchNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":   luaCompMatchName,
		"result": luaCompMatchResult,
	}))

	return mt
}

func lCheckCompMatch(L *lua.LState, index int) *compMatch {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*compMatch); ok {
		return v
	}

	L.ArgError(index, "value of type `CompMatch` expected")

	return nil
}

func lWrapCompMatch(L *lua.LState, data *compMatch) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaCompMatchTypeName))

	return ud
}

func lAddCompMatchToState(L *lua.LState, data *compMatch) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapCompMatch(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaCompMatchNew(L *lua.LState) int {
	name := L.CheckString(1)
	result := L.CheckString(2)
	return lAddCompMatchToState(L, &compMatch{name: name, result: result})
}

// ----------------------------------------------------------------------------

// luaCompMatchName is getter & setter for name field. It's displayed text for
// this completion entry.
func luaCompMatchName(L *lua.LState) int {
	cm := lCheckCompMatch(L, 1)

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
	cm := lCheckCompMatch(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		cm.result = value
	}

	L.Push(lua.LString(cm.result))

	return 1
}

// ----------------------------------------------------------------------------
// Type file

const FileTypeName = "lf.file"

func lRegisterFileTypeMt(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(FileTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaFileNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":     luaFileName,
		"size":     luaFileSize,
		"mode":     luaFileMode,
		"mod_time": luaFileModTime,
		"is_dir":   luaFileIsDir,

		"link_state":  luaFileLinkState,
		"link_target": luaFileLinkTarget,
		"path":        luaFilePath,

		"dir_count": luaFileDirCount,
		"dir_size":  luaFileDirSize,

		"access_time": luaFileAccessTime,
		"birth_time":  luaFileBirthTime,
		"change_time": luaFileChangeTime,

		"custom_info": luaFileCustomInfo,
		"ext":         luaFileExt,

		"extra_data": luaFileExtraData,

		"is_previewable": luaFileIsPreviewable,
	}))

	addLinkStateConstantToMt(L, mt)

	return mt
}

func addLinkStateConstantToMt(L *lua.LState, tbl *lua.LTable) {
	L.SetField(tbl, "LinkStateNotLink", lua.LNumber(notLink))
	L.SetField(tbl, "LinkStateWorking", lua.LNumber(working))
	L.SetField(tbl, "LinkStateBroken", lua.LNumber(broken))
}

func lCheckFile(L *lua.LState, index int) *file {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*file); ok {
		return v
	}

	L.ArgError(index, "value of type `File` expected")

	return nil
}

func lWrapFile(L *lua.LState, data *file) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(FileTypeName))

	return ud
}

func lAddFileToState(L *lua.LState, data *file) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapFile(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaFileNew(L *lua.LState) int {
	path := L.CheckString(1)
	file := newFile(path)
	return lAddFileToState(L, file)
}

// ----------------------------------------------------------------------------

func luaFileName(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(file.Name()))
	return 1
}

func luaFileSize(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LNumber(file.Size()))
	return 1
}

func luaFileMode(L *lua.LState) int {
	file := lCheckFile(L, 1)
	return lAddFileModeToState(L, file.Mode())
}

func luaFileModTime(L *lua.LState) int {
	file := lCheckFile(L, 1)
	modTime := file.ModTime()
	return lAddTimeToState(L, &modTime)
}

func luaFileIsDir(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LBool(file.IsDir()))
	return 1
}

func luaFileLinkState(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LNumber(file.linkState))
	return 1
}

func luaFileLinkTarget(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(file.linkTarget))
	return 1
}

func luaFilePath(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(file.path))
	return 1
}

// luaFileDirCount returns number items of a directory.
func luaFileDirCount(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LNumber(file.dirCount))
	return 1
}

// luaFileDirSize return directory's total content size.
func luaFileDirSize(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LNumber(file.dirSize))
	return 1
}

func luaFileAccessTime(L *lua.LState) int {
	file := lCheckFile(L, 1)
	return lAddTimeToState(L, &file.accessTime)
}

func luaFileBirthTime(L *lua.LState) int {
	file := lCheckFile(L, 1)
	return lAddTimeToState(L, &file.birthTime)
}

func luaFileChangeTime(L *lua.LState) int {
	file := lCheckFile(L, 1)
	return lAddTimeToState(L, &file.changeTime)
}

// luaFileCustomInfo returns custom info string add to this file by `addcustominfo`
// command.
func luaFileCustomInfo(L *lua.LState) int {
	file := lCheckFile(L, 1)

	if L.GetTop() >= 2 {
		tryRaiseNonSyncLuaStateError(L)
		value := L.CheckString(2)
		file.customInfo = value
	}

	L.Push(lua.LString(file.customInfo))

	return 1
}

func luaFileExt(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(file.ext))
	return 1
}

// luaFileExtraData can get & stores value to a map associated with this file.
// Only number, string, boolean, nil value are supported.
func luaFileExtraData(L *lua.LState) int {
	file := lCheckFile(L, 1)
	key := L.CheckString(2)

	nargs := L.GetTop()
	if nargs >= 3 {
		tryRaiseNonSyncLuaStateError(L)
		value := L.Get(3)

		if file.extraLuaData == nil {
			file.extraLuaData = make(map[string]any)
		}

		goValue, err := luaValueToGoValue(value)
		if err != nil {
			L.Push(value)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		file.extraLuaData[key] = goValue
		L.Push(value)

		return 1
	}

	if file.extraLuaData == nil {
		L.Push(lua.LNil)
		return 1
	}

	goValue := file.extraLuaData[key]
	value, err := goValueToLuaValue(L, goValue)

	L.Push(value)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// luaFileIsPreviewable returns true if this file requires a preview call.
func luaFileIsPreviewable(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LBool(file.isPreviewable()))
	return 1
}

// ----------------------------------------------------------------------------
// Type dir

const luaDirTypeName = "lf.dir"

func lRegisterDirType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaDirTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaDirNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"loading":        luaDirLoading,
		"load_time":      luaDirLoadTime,
		"ind":            luaDirInd,
		"pos":            luaDirPos,
		"path":           luaDirPath,
		"files":          luaDirFiles,
		"all_files":      luaDirAllFiles,
		"sortby":         luaDirSortby,
		"dircounts":      luaDirDircounts,
		"dirfirst":       luaLuaDirfirst,
		"dironly":        luaDirDironly,
		"hidden":         luaDirHidden,
		"reverse":        luaDirReverse,
		"visual_anchor":  luaDirVisualAnchor,
		"visual_wrap":    luaDirVisualWrap,
		"hiddenfiles":    luaDirHiddenFiles,
		"filter":         luaDirFilter,
		"sortignorecase": luaDirSortignorecase,
		"sortignoredia":  luaDirSortignoredia,
		"no_perm":        luaDirNoPerm,

		"sort":              luaDirSort,
		"name":              luaDirName,
		"visual_selections": luaDirVisualSelectioins,
		"sel":               luaDirSel,

		"iter_files":     luaDirIterFiles,
		"iter_all_files": luaDirIterAllFiles,

		"extra_data": luaDirExtraData,
	}))

	return mt
}

func lCheckDir(L *lua.LState, index int) *dir {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*dir); ok {
		return v
	}

	L.ArgError(index, "value of type `Dir` expected")

	return nil
}

func lWrapDir(L *lua.LState, data *dir) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaDirTypeName))

	return ud
}

func lAddDirToState(L *lua.LState, data *dir) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapDir(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaDirNew(L *lua.LState) int {
	path := L.CheckString(1)
	dir := newDir(path)
	return lAddDirToState(L, dir)
}

// ----------------------------------------------------------------------------

func luaDirLoading(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.loading))
	return 1
}

func luaDirLoadTime(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	return lAddTimeToState(L, &dir.loadTime)
}

// luaDirInd returns a 0-based index of current entry in directory.
func luaDirInd(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LNumber(dir.ind))
	return 1
}

// luaDirPos returns a 0-based row index indicating position of cursor.
func luaDirPos(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LNumber(dir.pos))
	return 1
}

func luaDirPath(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LString(dir.path))
	return 1
}

// luaDirFiles returns a list of displayed file.
func luaDirFiles(L *lua.LState) int {
	dir := lCheckDir(L, 1)

	filesTable := L.NewTable()
	for _, file := range dir.files {
		filesTable.Append(lWrapFile(L, file))
	}

	L.Push(filesTable)

	return 1
}

// luaDirAllFiles returns a list of file including non-displayed ones.
func luaDirAllFiles(L *lua.LState) int {
	dir := lCheckDir(L, 1)

	filesTable := L.NewTable()
	for _, file := range dir.allFiles {
		filesTable.Append(lWrapFile(L, file))
	}

	L.Push(filesTable)

	return 1
}

// luaDirSortby is getter & setter for directory sort method
func luaDirSortby(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LString(dir.sortby))
	return 1
}

func luaDirDircounts(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.dircounts))
	return 1
}

func luaLuaDirfirst(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.dirfirst))
	return 1
}

func luaDirDironly(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.dironly))
	return 1
}

func luaDirHidden(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.hidden))
	return 1
}

func luaDirReverse(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.reverse))
	return 1
}

// luaDirVisualAnchor returns anchor position of visual mode selection range.
func luaDirVisualAnchor(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LNumber(dir.visualAnchor))
	return 1
}

// luaDirVisualWrap returns wrap method of visual mode.
func luaDirVisualWrap(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LNumber(dir.visualWrap))
	return 1
}

func luaDirHiddenFiles(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	hiddenFilesTable := L.NewTable()
	for _, file := range dir.hiddenfiles {
		hiddenFilesTable.Append(lua.LString(file))
	}

	L.Push(hiddenFilesTable)

	return 1
}

func luaDirFilter(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	filterTable := L.NewTable()
	for _, file := range dir.filter {
		filterTable.Append(lua.LString(file))
	}

	L.Push(filterTable)

	return 1
}

func luaDirSortignorecase(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.sortignorecase))
	return 1
}

func luaDirSortignoredia(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.sortignoredia))
	return 1
}

// luaDirNoPerm returns true if progm doesn't have permission to open this directory.
func luaDirNoPerm(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LBool(dir.noPerm))
	return 1
}

// luaDirSort runs sorting for current directory
func luaDirSort(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	dir.sort()
	return 0
}

func luaDirName(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	L.Push(lua.LString(dir.name()))
	return 1
}

// luaDirVisualSelectioins returns a list of path selected in visual mode.
func luaDirVisualSelectioins(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	tbl := L.NewTable()

	paths := dir.visualSelections()
	for _, path := range paths {
		tbl.Append(lua.LString(path))
	}

	L.Push(tbl)

	return 1
}

func luaDirSel(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	dir := lCheckDir(L, 1)
	name := L.CheckString(2)

	height := int(L.CheckNumber(3))
	dir.sel(name, height)

	return 0
}

// luaDirIterFiles returns iterator over displayed files.
func luaDirIterFiles(L *lua.LState) int {
	dir := lCheckDir(L, 1)

	L.Push(L.NewFunction(func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		index := L.CheckInt(2)

		list, ok := ud.Value.([]*file)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}

		if index >= len(list) {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(lua.LNumber(index + 1))
		L.Push(lWrapFile(L, list[index]))

		return 2
	}))

	ud := L.NewUserData()
	ud.Value = dir.files

	L.Push(ud)
	L.Push(lua.LNumber(0))

	return 3
}

// luaDirIterAllFiles returns iterator over all files.
func luaDirIterAllFiles(L *lua.LState) int {
	dir := lCheckDir(L, 1)

	L.Push(L.NewFunction(func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		index := L.CheckInt(2)

		list, ok := ud.Value.([]*file)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}

		if index >= len(list) {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(lua.LNumber(index + 1))
		L.Push(lWrapFile(L, list[index]))

		return 2

	}))

	ud := L.NewUserData()
	ud.Value = dir.allFiles

	L.Push(ud)
	L.Push(lua.LNumber(0))

	return 3
}

// luaDirExtraData can get & stores value to a map associated with this file.
// Only number, string, boolean, nil value are supported.
func luaDirExtraData(L *lua.LState) int {
	dir := lCheckDir(L, 1)
	key := L.CheckString(2)

	nargs := L.GetTop()
	if nargs >= 3 {
		tryRaiseNonSyncLuaStateError(L)
		value := L.Get(3)

		if dir.extraLuaData == nil {
			dir.extraLuaData = make(map[string]any)
		}

		goValue, err := luaValueToGoValue(value)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		dir.extraLuaData[key] = goValue
		L.Push(value)

		return 1
	}

	if dir.extraLuaData == nil {
		L.Push(lua.LNil)
		return 1
	}

	goValue := dir.extraLuaData[key]
	value, err := goValueToLuaValue(L, goValue)

	L.Push(value)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// ----------------------------------------------------------------------------
// Type nav

const luaNavTypeName = "lf.nav"

func lRegisterNavType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaNavTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get_dir":           luaNavGetDir,
		"cd_jump_list_prev": luaNavCdJumpListPrev,
		"cd_jump_list_next": luaNavCdJumpListNext,
		"renew":             luaNavRenew,
		"reload":            luaNavReload,
		"sort":              luaNavSort,
		"set_filter":        luaNavSetFilter,
		"up":                luaNavUp,
		"down":              luaNavDown,
		"scroll_up":         luaNavScrollUp,
		"scroll_down":       luaNavScrollDown,
		"updir":             luaNavUpDir,
		"open":              luaNavOpen,
		"top":               luaNavTop,
		"bottom":            luaNavBottom,
		"high":              luaNavHigh,
		"middle":            luaNavMiddle,
		"low":               luaNavLow,
		"move":              luaNavMove,

		"select":               luaNavSelect,
		"toggle_selection":     luaNavToggleSelection,
		"toggle":               luaNavToggle,
		"tag_toggle_selection": luaNavTagToggleSelection,
		"tag_toggle":           luaNavTagToggle,
		"tag":                  luaNavTag,
		"invert":               luaNavInvert,
		"unselect":             luaNavUnselect,
		"unselect_one":         luaNavUnselectOne,
		"cd":                   luaNavCd,
		"glob_sel":             luaNavGlobSel,

		"read_marks":  luaNavReadMarks,
		"write_marks": luaNavWriteMarks,
		"read_tags":   luaNavReadTags,
		"write_tags":  luaNavWriteTags,

		"curr_dir":               luaNavCurrDir,
		"curr_file":              luaNavCurrFile,
		"curr_selections":        luaNavCurrSelections,
		"curr_file_or_selection": luaNavCurrFileOrSelection,

		"get_tag": luaNavGetTag,
	}))

	return mt
}

func lCheckNav(L *lua.LState, index int) *nav {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*nav); ok {
		return v
	}

	L.ArgError(index, "value of type `Nav` expected")

	return nil
}

func lWrapNav(L *lua.LState, data *nav) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaNavTypeName))

	return ud
}

func lAddNavToState(L *lua.LState, data *nav) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapNav(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaNavGetDir(L *lua.LState) int {
	nav := lCheckNav(L, 1)
	path := L.CheckString(2)
	dir := nav.getDir(path)
	return lAddDirToState(L, dir)
}

func luaNavCdJumpListPrev(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.cdJumpListPrev()

	return 0
}

func luaNavCdJumpListNext(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.cdJumpListNext()

	return 0
}

func luaNavRenew(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.renew()

	return 0
}

func luaNavReload(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.reload()

	return 0
}

func luaNavSort(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.sort()

	return 0
}

func luaNavSetFilter(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	tbl := L.CheckTable(2)

	nPatt := tbl.Len()
	if nPatt <= 0 {
		return 0
	}

	patterns := make([]string, nPatt)
	for i := 1; i <= nPatt; i++ {
		value := tbl.RawGetInt(i)
		pattern, ok := value.(lua.LString)
		if ok {
			patterns = append(patterns, string(pattern))
		}
	}

	nav.setFilter(patterns)

	return 0
}

func luaNavUp(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	dist := L.CheckNumber(2)

	nav.up(int(dist))

	return 0
}

func luaNavDown(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	dist := L.CheckNumber(2)

	nav.down(int(dist))

	return 0
}

func luaNavScrollUp(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	dist := L.CheckNumber(2)

	nav.scrollUp(int(dist))

	return 0
}

func luaNavScrollDown(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	dist := L.CheckNumber(2)

	nav.scrollDown(int(dist))

	return 0
}

func luaNavUpDir(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.updir()

	return 0
}

func luaNavOpen(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.open()

	return 0
}

func luaNavTop(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	L.Push(lua.LBool(nav.top()))

	return 1
}

func luaNavBottom(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	L.Push(lua.LBool(nav.bottom()))

	return 1
}

func luaNavHigh(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	L.Push(lua.LBool(nav.high()))

	return 1
}

func luaNavMiddle(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	L.Push(lua.LBool(nav.middle()))

	return 1
}

func luaNavLow(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	L.Push(lua.LBool(nav.low()))

	return 1
}

func luaNavMove(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	index := L.CheckNumber(2)

	nav.move(int(index))

	return 0
}

func luaNavSelect(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	path := L.CheckString(2)

	nav.selections[path] = nav.selectionInd
	nav.selectionInd++

	return 0
}

func luaNavToggleSelection(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	path := L.CheckString(2)

	nav.toggleSelection(path)

	return 0
}

func luaNavToggle(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.toggle()

	return 0
}

func luaNavTagToggleSelection(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	path := L.CheckString(2)
	tag := L.CheckString(3)

	nav.tagToggleSelection(path, tag)

	return 0
}

func luaNavTagToggle(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	tag := L.CheckString(2)

	if err := nav.tagToggle(tag); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavTag(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	tag := L.CheckString(2)

	if err := nav.tag(tag); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavInvert(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.invert()

	return 0
}

func luaNavUnselect(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	nav.unselect()

	return 0
}

func luaNavUnselectOne(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	path := L.CheckString(2)

	if _, ok := nav.selections[path]; ok {
		delete(nav.selections, path)
		if len(nav.selections) == 0 {
			nav.selectionInd = 0
		}
	}

	return 0
}

func luaNavCd(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	path := L.CheckString(2)

	if err := nav.cd(path); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavGlobSel(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	pattern := L.CheckString(2)
	invert := L.CheckBool(3)

	if err := nav.globSel(pattern, invert); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavCurrDir(L *lua.LState) int {
	nav := lCheckNav(L, 1)
	return lAddDirToState(L, nav.currDir())
}

func luaNavReadMarks(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	err := nav.readMarks()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func luaNavWriteMarks(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	err := nav.writeMarks()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func luaNavReadTags(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	err := nav.readTags()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func luaNavWriteTags(L *lua.LState) int {
	tryRaiseNonSyncLuaStateError(L)

	nav := lCheckNav(L, 1)
	err := nav.writeTags()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func luaNavCurrFile(L *lua.LState) int {
	nav := lCheckNav(L, 1)
	return lAddFileInfoToState(L, nav.currFile())
}

func luaNavCurrSelections(L *lua.LState) int {
	nav := lCheckNav(L, 1)

	tbl := L.NewTable()
	selections := nav.currSelections()
	for _, path := range selections {
		tbl.Append(lua.LString(path))
	}

	L.Push(tbl)

	return 1
}

func luaNavCurrFileOrSelection(L *lua.LState) int {
	nav := lCheckNav(L, 1)

	results, err := nav.currFileOrSelections()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	tbl := L.NewTable()
	for _, path := range results {
		tbl.Append(lua.LString(path))
	}

	L.Push(tbl)

	return 1
}

func luaNavGetTag(L *lua.LState) int {
	nav := lCheckNav(L, 1)
	path := L.CheckString(2)

	value, exists := nav.tags[path]
	if !exists {
		return 0
	}

	L.Push(lua.LString(value))

	return 1
}

// ----------------------------------------------------------------------------
// Type ui

const luaUITypeName = "lf.ui"

func lRegisterUIType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaUITypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"screen": luaUiScreen,

		"echo":     luaUIEcho,
		"echomsg":  luaUIEchoMsg,
		"echoerr":  luaUIEchhoErr,
		"echoerrf": luaUIEchhoErrf,
	}))

	return mt
}

func lCheckUI(L *lua.LState, index int) *ui {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*ui); ok {
		return v
	}

	L.ArgError(index, "value of type `UI` expected")

	return nil
}

func lWrapUI(L *lua.LState, data *ui) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaUITypeName))

	return ud
}

func lAddUIToState(L *lua.LState, data *ui) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapUI(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaUiScreen(L *lua.LState) int {
	ui := lCheckUI(L, 1)
	return lAddTcellScreenToState(L, ui.screen)
}

// luaUIEcho prints content to lf message bar.
func luaUIEcho(L *lua.LState) int {
	ui := lCheckUI(L, 1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	ui.exprChan <- &callExpr{"echo", args, 1}

	return 0
}

// luaUIEcho prints content to both lf message bar and log.
func luaUIEchoMsg(L *lua.LState) int {
	ui := lCheckUI(L, 1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	ui.exprChan <- &callExpr{"echomsg", args, 1}

	return 0
}

// luaUIEcho prints error message to both lf message bar and log.
func luaUIEchhoErr(L *lua.LState) int {
	ui := lCheckUI(L, 1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	ui.exprChan <- &callExpr{"echoerr", args, 1}

	return 0
}

// luaUIEcho prints error message with formatting string.
func luaUIEchhoErrf(L *lua.LState) int {
	ui := lCheckUI(L, 1)
	fmtStr := L.ToString(2)

	st := 3
	nArgs := L.GetTop()
	args := make([]any, nArgs-st+1)
	for i := 3; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	msg := fmt.Sprintf(fmtStr, args...)
	ui.exprChan <- &callExpr{"echoerr", []string{msg}, 1}

	return 0
}

// ----------------------------------------------------------------------------
// type win

const LuaWinTypeName = "lf.win"

func lRegisterWinType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaWinTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaWinNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"w": luaWinW,
		"h": luaWinH,
		"x": luaWinX,
		"y": luaWinY,

		"renew": luaWinRenew,

		"print":     luaWinPrint,
		"print_msg": luaWinPrintMsg,
	}))

	return mt
}

func lCheckWin(L *lua.LState, index int) *win {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*win); ok {
		return v
	}

	L.ArgError(index, "value of type `Win` expected")

	return nil
}

func lWrapWin(L *lua.LState, data *win) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaWinTypeName))

	return ud
}

func lAddWinToState(L *lua.LState, data *win) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapWin(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaWinNew(L *lua.LState) int {
	w := L.CheckInt(1)
	h := L.CheckInt(1)
	x := L.CheckInt(1)
	y := L.CheckInt(1)

	return lAddWinToState(L, newWin(w, h, x, y))
}

// ----------------------------------------------------------------------------

func luaWinW(L *lua.LState) int {
	win := lCheckWin(L, 1)
	L.Push(lua.LNumber(win.w))
	return 1
}

func luaWinH(L *lua.LState) int {
	win := lCheckWin(L, 1)
	L.Push(lua.LNumber(win.h))
	return 1
}

func luaWinX(L *lua.LState) int {
	win := lCheckWin(L, 1)
	L.Push(lua.LNumber(win.x))
	return 1
}

func luaWinY(L *lua.LState) int {
	win := lCheckWin(L, 1)
	L.Push(lua.LNumber(win.y))
	return 1
}

func luaWinRenew(L *lua.LState) int {
	win := lCheckWin(L, 1)
	w := L.CheckInt(1)
	h := L.CheckInt(1)
	x := L.CheckInt(1)
	y := L.CheckInt(1)

	win.renew(w, h, x, y)

	return 0
}

func luaWinPrint(L *lua.LState) int {
	win := lCheckWin(L, 1)
	screen := lCheckTcellScreen(L, 2)
	x := L.CheckInt(3)
	y := L.CheckInt(4)
	st := lCheckTcellStyle(L, 5)
	str := L.CheckString(6)

	result := win.print(screen, x, y, *st, str)

	return lAddTcellStyleToState(L, &result)
}

func luaWinPrintMsg(L *lua.LState) int {
	win := lCheckWin(L, 1)
	screen := lCheckTcellScreen(L, 2)
	msg := L.CheckString(3)

	win.printMsg(screen, msg)

	return 0
}

// ----------------------------------------------------------------------------
// type dirStyle

const LuaDirStyleTypeName = "lf.dirStyle"

func lRegisterDirStyleType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaDirStyleTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"colors": luaDirStyleColors,
		"icons":  luaDirStyleIcons,
		"role":   luaDirStyleRole,
	}))

	return mt
}

func lCheckDirStyle(L *lua.LState, index int) *dirStyle {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*dirStyle); ok {
		return v
	}

	L.ArgError(index, "value of type `DirStyle` expected")

	return nil
}

func lWrapDirStyle(L *lua.LState, data *dirStyle) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaDirStyleTypeName))

	return ud
}

func lAddDirStyleToState(L *lua.LState, data *dirStyle) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapDirStyle(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaDirStyleColors(L *lua.LState) int {
	dirSt := lCheckDirStyle(L, 1)
	return lAddStyleMapToState(L, &dirSt.colors)
}

func luaDirStyleIcons(L *lua.LState) int {
	dirSt := lCheckDirStyle(L, 1)
	return lAddIconMapToState(L, &dirSt.icons)
}

func luaDirStyleRole(L *lua.LState) int {
	dirSt := lCheckDirStyle(L, 1)
	L.Push(lua.LNumber(dirSt.role))
	return 1
}

// ----------------------------------------------------------------------------
// type styleMap

const LuaStyleMapTypeName = "lf.styleMap"

func lRegisterStyleMapType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaStyleMapTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get": luaStyleMapGet,
	}))

	return mt
}

func lCheckStyleMap(L *lua.LState, index int) *styleMap {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*styleMap); ok {
		return v
	}

	L.ArgError(index, "value of type `StyleMap` expected")

	return nil
}

func lWrapStyleMap(L *lua.LState, data *styleMap) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaStyleMapTypeName))

	return ud
}

func lAddStyleMapToState(L *lua.LState, data *styleMap) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapStyleMap(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaStyleMapGet(L *lua.LState) int {
	stMap := lCheckStyleMap(L, 1)
	file := lCheckFile(L, 2)
	st := stMap.get(file)
	return lAddTcellStyleToState(L, &st)
}

// ----------------------------------------------------------------------------
// type iconDef

const LuaIconDefTypeName = "lf.iconDef"

func lRegisterIconDefType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaIconDefTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"icon":      luaIconDefIcon,
		"has_style": luaIconDefHasStyle,
		"style":     luaIconDefStyle,
	}))

	return mt
}

func lCheckIconDef(L *lua.LState, index int) *iconDef {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*iconDef); ok {
		return v
	}

	L.ArgError(index, "value of type `IconDef` expected")

	return nil
}

func lWrapIconDef(L *lua.LState, data *iconDef) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaIconDefTypeName))

	return ud
}

func lAddIconDefToState(L *lua.LState, data *iconDef) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapIconDef(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaIconDefIcon gets icon string of a file.
func luaIconDefIcon(L *lua.LState) int {
	def := lCheckIconDef(L, 1)
	L.Push(lua.LString(def.icon))
	return 1
}

// luaIconDefHasStyle returns if this icon has style.
func luaIconDefHasStyle(L *lua.LState) int {
	def := lCheckIconDef(L, 1)
	L.Push(lua.LBool(def.hasStyle))
	return 1
}

// luaIconDefStyle returns style object binded with this icon.
func luaIconDefStyle(L *lua.LState) int {
	def := lCheckIconDef(L, 1)
	return lAddTcellStyleToState(L, &def.style)
}

// ----------------------------------------------------------------------------
// type iconMap

const LuaIconMapTypeName = "lf.iconMap"

func lRegisterIconMapType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaIconMapTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get": luaIconMapGet,
	}))

	return mt
}

func lCheckIconMap(L *lua.LState, index int) *iconMap {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*iconMap); ok {
		return v
	}

	L.ArgError(index, "value of type `IconMap` expected")

	return nil
}

func lWrapIconMap(L *lua.LState, data *iconMap) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaIconMapTypeName))

	return ud
}

func lAddIconMapToState(L *lua.LState, data *iconMap) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapIconMap(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaIconMapGet gets icon definition of a file.
func luaIconMapGet(L *lua.LState) int {
	im := lCheckIconMap(L, 1)
	file := lCheckFile(L, 2)
	def := im.get(file)
	return lAddIconDefToState(L, &def)
}

// ----------------------------------------------------------------------------
// type dirContext

const LuaDirContextTypeName = "lf.dirContext"

func lRegisterDirContextType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaDirContextTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"selections": luaDirContextSelections,
		"clipboard":  luaDirContextClipboard,
		"tags":       luaDirContextTags,

		"get_selection_index": luaDirContextGetSelectionIndex,
		"get_tag":             luaDirContextGetTag,
	}))

	return mt
}

func lCheckDirContext(L *lua.LState, index int) *dirContext {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*dirContext); ok {
		return v
	}

	L.ArgError(index, "value of type `DirContext` expected")

	return nil
}

func lWrapDirContext(L *lua.LState, data *dirContext) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaDirContextTypeName))

	return ud
}

func lAddDirContextToState(L *lua.LState, data *dirContext) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapDirContext(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaDirContextNew(L *lua.LState) int {
	tbl := L.CheckTable(1)

	selectionsTbl, ok := tbl.RawGetString("selections").(*lua.LTable)
	if !ok {
		L.RaiseError("key `selections` should be a table")
	}
	selections := map[string]int{}
	selectionsTbl.ForEach(func(kValue, vValue lua.LValue) {
		key, keyOk := kValue.(lua.LString)
		value, valueOk := vValue.(lua.LNumber)
		if keyOk && valueOk {
			selections[string(key)] = int(value)
		}
	})

	clipboardValue, ok := tbl.RawGetString("clipboard").(*lua.LUserData)
	if !ok {
		L.RaiseError("key `clipboard` should be a userdata")
	}
	clipboard, ok := clipboardValue.Value.(*clipboard)
	if !ok {
		L.RaiseError("key `clipboard` should be a clipboard object")
	}

	tagTbl, ok := tbl.RawGetString("tags").(*lua.LTable)
	if !ok {
		L.RaiseError("key `tags` should be a table")
	}
	tags := map[string]string{}
	tagTbl.ForEach(func(kValue, vValue lua.LValue) {
		key, keyOk := kValue.(lua.LString)
		value, valueOk := vValue.(lua.LString)
		if keyOk && valueOk {
			tags[string(key)] = string(value)
		}
	})

	visualSelectionTbl, ok := tbl.RawGetString("visual_selections").(*lua.LTable)
	if !ok {
		L.RaiseError("key `visual_selections` should be a table")
	}
	visualSelections := []string{}
	nVisualSelection := visualSelectionTbl.Len()
	for i := 1; i <= nVisualSelection; i++ {
		value := visualSelectionTbl.RawGetInt(i)
		if path, ok := value.(lua.LString); ok {
			visualSelections = append(visualSelections, string(path))
		}
	}

	context := &dirContext{
		selections: selections,
		clipboard:  *clipboard,
		tags:       tags,
	}

	return lAddDirContextToState(L, context)
}

// ----------------------------------------------------------------------------

func luaDirContextSelections(L *lua.LState) int {
	context := lCheckDirContext(L, 1)

	tbl := L.NewTable()
	for k, v := range context.selections {
		tbl.RawSetString(k, lua.LNumber(v))
	}

	L.Push(tbl)

	return 1
}

func luaDirContextClipboard(L *lua.LState) int {
	context := lCheckDirContext(L, 1)
	return lAddClipboardToState(L, &context.clipboard)
}

func luaDirContextTags(L *lua.LState) int {
	context := lCheckDirContext(L, 1)

	tbl := L.NewTable()
	for k, v := range context.tags {
		tbl.RawSetString(k, lua.LString(v))
	}

	L.Push(tbl)

	return 1
}

// luaDirContextGetSelectionIndex returns 1-based selection index of
// given path, returns 0 when that path is not selected.
func luaDirContextGetSelectionIndex(L *lua.LState) int {
	context := lCheckDirContext(L, 1)
	path := L.CheckString(2)

	index, found := context.selections[path]
	if found {
		L.Push(lua.LNumber(index + 1))
	} else {
		L.Push(lua.LNumber(0))
	}

	return 1
}

// luaDirContextGetTag returns tag of given path, returns `nil` when
// no tag is set for target path.
func luaDirContextGetTag(L *lua.LState) int {
	context := lCheckDirContext(L, 1)
	path := L.CheckString(2)

	tag, ok := context.tags[path]
	if ok {
		L.Push(lua.LString(tag))
	} else {
		L.Push(lua.LNil)
	}

	return 1
}

// ----------------------------------------------------------------------------
// type printDirEntryContext

const LuaPrintDirEntryContextTypeName = "lf.printDirEntryContext"

func lRegisterPrintDirEntryContextType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaPrintDirEntryContextTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaPrintDirEntryContextNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"dir":       luaPrintDirEntryContextDir,
		"dir_beg":   luaPrintDirEntryContextDirBeg,
		"dir_end":   luaPrintDirEntryContextDirEnd,
		"dir_style": luaPrintDirEntryContextDirStyle,

		"lnwidth":      luaPrintDirEntryContextLnwidth,
		"user_width":   luaPrintDirEntryContextUserWidth,
		"group_width":  luaPrintDirEntryContextGroupWidth,
		"custom_width": luaPrintDirEntryContextCustomWidth,

		"selections":         luaPrintDirEntryContextSelections,
		"clipboard":          luaPrintDirEntryContextClipboard,
		"tags":               luaPrintDirEntryContextTags,
		"visual_selectioins": luaPrintDirEntryContextVisualSelections,

		"get_selection_index":      luaPrintDirEntryContextGetSelectionIndex,
		"visual_selection_contain": luaPrintDirEntryContextVisualSelectionsContain,
		"get_tag":                  luaPrintDirEntryContextGetTag,
	}))

	return mt
}

func lCheckPrintDirEntryContext(L *lua.LState, index int) *printDirEntryContext {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*printDirEntryContext); ok {
		return v
	}

	L.ArgError(index, "value of type `PrintDirEntryContext` expected")

	return nil
}

func lWrapPrintDirEntryContext(L *lua.LState, data *printDirEntryContext) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaPrintDirEntryContextTypeName))

	return ud
}

func lAddPrintDirEntryContextToState(L *lua.LState, data *printDirEntryContext) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapPrintDirEntryContext(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaPrintDirEntryContextNew(L *lua.LState) int {
	tbl := L.CheckTable(1)

	dirUd, ok := tbl.RawGetString("dir").(*lua.LUserData)
	if !ok {
		L.RaiseError("key `dir` should be userdata")
	}
	dir, ok := dirUd.Value.(*dir)
	if !ok {
		L.RaiseError("key `dir` should be a dir object")
	}

	dirBegValue, ok := tbl.RawGetString("dir_beg").(lua.LNumber)
	if !ok {
		L.RaiseError("key `dir_beg` should be a number")
	}
	dirBeg := int(dirBegValue)

	dirEndValue, ok := tbl.RawGetString("dir_end").(lua.LNumber)
	if !ok {
		L.RaiseError("key `dir_end` should be a number")
	}
	dirEnd := int(dirEndValue)

	dirStyleUd, ok := tbl.RawGetString("dir_style").(*lua.LUserData)
	if !ok {
		L.RaiseError("key `dir_style` should be userdata")
	}
	dirStyle, ok := dirStyleUd.Value.(*dirStyle)
	if !ok {
		L.RaiseError("key `dir_style should be dirStyle object")
	}

	lnwidthValue, ok := tbl.RawGetString("lnwidth").(lua.LNumber)
	if !ok {
		L.RaiseError("key `lnwidth` should be a number")
	}
	lnwidth := int(lnwidthValue)

	userWidthValue, ok := tbl.RawGetString("user_width").(lua.LNumber)
	if !ok {
		L.RaiseError("key `lnwidth` should be a number")
	}
	userWidth := int(userWidthValue)

	groupWidthValue, ok := tbl.RawGetString("group_width").(lua.LNumber)
	if !ok {
		L.RaiseError("key `lnwidth` should be a number")
	}
	groupWidth := int(groupWidthValue)

	customWidthValue, ok := tbl.RawGetString("custom_width").(lua.LNumber)
	if !ok {
		L.RaiseError("key `lnwidth` should be a number")
	}
	customWidth := int(customWidthValue)

	selectionsTbl, ok := tbl.RawGetString("selections").(*lua.LTable)
	if !ok {
		L.RaiseError("key `selections` should be a table")
	}
	selections := map[string]int{}
	selectionsTbl.ForEach(func(kValue, vValue lua.LValue) {
		key, keyOk := kValue.(lua.LString)
		value, valueOk := vValue.(lua.LNumber)
		if keyOk && valueOk {
			selections[string(key)] = int(value)
		}
	})

	clipboardValue, ok := tbl.RawGetString("clipboard").(*lua.LUserData)
	if !ok {
		L.RaiseError("key `clipboard` should be a userdata")
	}
	clipboard, ok := clipboardValue.Value.(*clipboard)
	if !ok {
		L.RaiseError("key `clipboard` should be a clipboard object")
	}

	tagTbl, ok := tbl.RawGetString("tags").(*lua.LTable)
	if !ok {
		L.RaiseError("key `tags` should be a table")
	}
	tags := map[string]string{}
	tagTbl.ForEach(func(kValue, vValue lua.LValue) {
		key, keyOk := kValue.(lua.LString)
		value, valueOk := vValue.(lua.LString)
		if keyOk && valueOk {
			tags[string(key)] = string(value)
		}
	})

	visualSelectionTbl, ok := tbl.RawGetString("visual_selections").(*lua.LTable)
	if !ok {
		L.RaiseError("key `visual_selections` should be a table")
	}
	visualSelections := []string{}
	nVisualSelection := visualSelectionTbl.Len()
	for i := 1; i <= nVisualSelection; i++ {
		value := visualSelectionTbl.RawGetInt(i)
		if path, ok := value.(lua.LString); ok {
			visualSelections = append(visualSelections, string(path))
		}
	}

	context := &printDirEntryContext{
		dir:      dir,
		dirBeg:   dirBeg,
		dirEnd:   dirEnd,
		dirStyle: dirStyle,

		lnwidth:     lnwidth,
		userWidth:   userWidth,
		groupWidth:  groupWidth,
		customWidth: customWidth,

		selections:       selections,
		clipboard:        *clipboard,
		tags:             tags,
		visualSelections: visualSelections,
	}

	return lAddPrintDirEntryContextToState(L, context)
}

// ----------------------------------------------------------------------------

func luaPrintDirEntryContextDir(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	return lAddDirToState(L, context.dir)
}

func luaPrintDirEntryContextDirBeg(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	L.Push(lua.LNumber(context.dirBeg))
	return 1
}

func luaPrintDirEntryContextDirEnd(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	L.Push(lua.LNumber(context.dirEnd))
	return 1
}

func luaPrintDirEntryContextDirStyle(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	return lAddDirStyleToState(L, context.dirStyle)
}

func luaPrintDirEntryContextLnwidth(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	L.Push(lua.LNumber(context.lnwidth))
	return 1
}

func luaPrintDirEntryContextUserWidth(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	L.Push(lua.LNumber(context.userWidth))
	return 1
}

func luaPrintDirEntryContextGroupWidth(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	L.Push(lua.LNumber(context.groupWidth))
	return 1
}

func luaPrintDirEntryContextCustomWidth(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	L.Push(lua.LNumber(context.customWidth))
	return 1
}

func luaPrintDirEntryContextSelections(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)

	tbl := L.NewTable()
	for k, v := range context.selections {
		tbl.RawSetString(k, lua.LNumber(v))
	}

	L.Push(tbl)

	return 1
}

func luaPrintDirEntryContextClipboard(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	return lAddClipboardToState(L, &context.clipboard)
}

func luaPrintDirEntryContextTags(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)

	tbl := L.NewTable()
	for k, v := range context.tags {
		tbl.RawSetString(k, lua.LString(v))
	}

	L.Push(tbl)

	return 1
}

func luaPrintDirEntryContextVisualSelections(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)

	tbl := L.NewTable()
	for _, v := range context.visualSelections {
		tbl.Append(lua.LString(v))
	}

	L.Push(tbl)

	return 1
}

// luaPrintDirEntryContextGetSelectionIndex returns 1-based selection index of
// given path, returns 0 when that path is not selected.
func luaPrintDirEntryContextGetSelectionIndex(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	path := L.CheckString(2)

	index, found := context.selections[path]
	if found {
		L.Push(lua.LNumber(index + 1))
	} else {
		L.Push(lua.LNumber(0))
	}

	return 1
}

// luaPrintDirEntryContextVisualSelectionsContain checks if visual selection
// contains given path.
func luaPrintDirEntryContextVisualSelectionsContain(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	path := L.CheckString(2)

	found := slices.Contains(context.visualSelections, path)
	L.Push(lua.LBool(found))

	return 1
}

// luaPrintDirEntryContextGetTag returns tag of given path, returns `nil` when
// no tag is set for target path.
func luaPrintDirEntryContextGetTag(L *lua.LState) int {
	context := lCheckPrintDirEntryContext(L, 1)
	path := L.CheckString(2)

	tag, ok := context.tags[path]
	if ok {
		L.Push(lua.LString(tag))
	} else {
		L.Push(lua.LNil)
	}

	return 1
}

// ----------------------------------------------------------------------------
// type clipboard

const LuaClipboardTypeName = "lf.clipboard"

func lRegisterClipboardType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaClipboardTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"paths":         luaClipboardPaths,
		"mode":          luaClipboardMode,
		"iter_path":     luaClipboardIterPath,
		"contains_path": luaClipboardPathsContain,
	}))

	return mt
}

func lCheckClipboard(L *lua.LState, index int) *clipboard {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*clipboard); ok {
		return v
	}

	L.ArgError(index, "value of type `Clipboard` expected")

	return nil
}

func lWrapClipboard(L *lua.LState, data *clipboard) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaClipboardTypeName))

	return ud
}

func lAddClipboardToState(L *lua.LState, data *clipboard) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapClipboard(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaClipboardPaths(L *lua.LState) int {
	board := lCheckClipboard(L, 1)

	tbl := L.NewTable()
	for _, path := range board.paths {
		tbl.Append(lua.LString(path))
	}

	L.Push(tbl)

	return 1
}

func luaClipboardMode(L *lua.LState) int {
	board := lCheckClipboard(L, 1)
	L.Push(lua.LNumber(board.mode))
	return 1
}

func luaClipboardIterPath(L *lua.LState) int {
	board := lCheckClipboard(L, 1)

	L.Push(L.NewFunction(func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		index := L.CheckInt(2)

		list, ok := ud.Value.([]string)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}

		if index >= len(list) {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(lua.LNumber(index + 1))
		L.Push(lua.LString(list[index]))

		return 2
	}))

	ud := L.NewUserData()
	ud.Value = board.paths

	L.Push(ud)
	L.Push(lua.LNumber(0))

	return 3
}

func luaClipboardPathsContain(L *lua.LState) int {
	board := lCheckClipboard(L, 1)
	path := L.CheckString(2)
	found := slices.Contains(board.paths, path)
	L.Push(lua.LBool(found))
	return 1
}

// ----------------------------------------------------------------------------
// type luaMsgExpr

const LuaLuaMsgExprTypeName = "lf.luaMsgExpr"

func lRegisterLuaMsgExprType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaLuaMsgExprTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func lCheckLuaMsgExpr(L *lua.LState, index int) *luaMsgExpr {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*luaMsgExpr); ok {
		return v
	}

	L.ArgError(index, "value of type `LuaMsgExpr` expected")

	return nil
}

func lWrapLuaMsgExpr(L *lua.LState, data *luaMsgExpr) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaLuaMsgExprTypeName))

	return ud
}

func lAddLuaMsgExprToState(L *lua.LState, data *luaMsgExpr) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapLuaMsgExpr(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------
// type luaFuncWriter

const luaFuncWriterTypeName = "lf.FuncWriter"

func lRegisterFuncWriterType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaFuncWriterTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaFuncWriterNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"write": luaFuncWriterWrite,
	}))

	return mt
}

func lCheckFuncWriter(L *lua.LState, index int) *luaFuncWriter {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*luaFuncWriter); ok {
		return v
	}

	L.ArgError(index, "value of type `FuncWriter` expected")

	return nil
}

func lWrapFuncWriter(L *lua.LState, data *luaFuncWriter) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaFuncWriterTypeName))

	return ud
}

func lAddFuncWriterToState(L *lua.LState, data *luaFuncWriter) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapFuncWriter(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaFuncWriterNew(L *lua.LState) int {
	fn := L.CheckFunction(1)
	writer := &luaFuncWriter{
		luaState: L,
		fn:       fn,
	}
	return lAddFuncWriterToState(L, writer)
}

// ----------------------------------------------------------------------------

func luaFuncWriterWrite(L *lua.LState) int {
	writer := lCheckFuncWriter(L, 1)
	content := L.CheckString(2)

	n, err := writer.Write([]byte(content))
	L.Push(lua.LNumber(n))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}
