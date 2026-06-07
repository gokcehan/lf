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

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaFileNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":   luaFileName,
		"size":   luaFileSize,
		"mode":   luaFileMode,
		"is_dir": luaFileIsDir,

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

		"extra_info": luaFileExtraInfo,

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

func luaFileNew(L *lua.LState) int {
	path := L.CheckString(1)
	file := newFile(path)
	return LAddFileToState(L, file)
}

// ----------------------------------------------------------------------------

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
	L.Push(lua.LNumber(file.linkState))
	return 1
}

func luaFileLinkTarget(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LString(file.linkTarget))
	return 1
}

func luaFilePath(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LString(file.path))
	return 1
}

func luaFileDirCount(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LNumber(file.dirCount))
	return 1
}

func luaFileDirSize(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LNumber(file.dirSize))
	return 1
}

func luaFileAccessTime(L *lua.LState) int {
	file := LCheckFile(L, 1)
	return LAddTimeToState(L, &file.accessTime)
}

func luaFileBirthTime(L *lua.LState) int {
	file := LCheckFile(L, 1)
	return LAddTimeToState(L, &file.birthTime)
}

func luaFileChangeTime(L *lua.LState) int {
	file := LCheckFile(L, 1)
	return LAddTimeToState(L, &file.changeTime)
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
	L.Push(lua.LString(file.ext))
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

func luaFileIsPreviewable(L *lua.LState) int {
	file := LCheckFile(L, 1)
	L.Push(lua.LBool(file.isPreviewable()))
	return 1
}

// ----------------------------------------------------------------------------
// Type dir

const LuaDirTypeName = "lf.dir"

func LRegisterDirType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaDirTypeName)

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

		"files_for_each":     luaDirFilesForEach,
		"all_files_for_each": luaDirAllFilesForEach,
	}))

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

func luaDirNew(L *lua.LState) int {
	path := L.CheckString(1)
	dir := newDir(path)
	return LAddDirToState(L, dir)
}

// ----------------------------------------------------------------------------

func luaDirLoading(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.loading))
	return 1
}

func luaDirLoadTime(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	return LAddTimeToState(L, &dir.loadTime)
}

func luaDirInd(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LNumber(dir.ind))
	return 1
}

func luaDirPos(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LNumber(dir.pos))
	return 1
}

func luaDirPath(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LString(dir.path))
	return 1
}

func luaDirFiles(L *lua.LState) int {
	dir := LCheckDir(L, 1)

	filesTable := L.NewTable()
	for _, file := range dir.files {
		filesTable.Append(LWrapFile(L, file))
	}

	L.Push(filesTable)

	return 1
}

func luaDirAllFiles(L *lua.LState) int {
	dir := LCheckDir(L, 1)

	filesTable := L.NewTable()
	for _, file := range dir.allFiles {
		filesTable.Append(LWrapFile(L, file))
	}

	L.Push(filesTable)

	return 1
}

func luaDirSortby(L *lua.LState) int {
	dir := LCheckDir(L, 1)

	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		dir.sortby = sortMethod(value)
	}

	L.Push(lua.LString(dir.sortby))

	return 0
}

func luaDirDircounts(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.dircounts))
	return 1
}

func luaLuaDirfirst(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.dirfirst))
	return 1
}

func luaDirDironly(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.dironly))
	return 1
}

func luaDirHidden(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.hidden))
	return 1
}

func luaDirReverse(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.reverse))
	return 1
}

func luaDirVisualAnchor(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LNumber(dir.visualAnchor))
	return 1
}

func luaDirVisualWrap(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LNumber(dir.visualWrap))
	return 1
}

func luaDirHiddenFiles(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	hiddenFilesTable := L.NewTable()
	for _, file := range dir.hiddenfiles {
		hiddenFilesTable.Append(lua.LString(file))
	}

	L.Push(hiddenFilesTable)

	return 1
}

func luaDirFilter(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	filterTable := L.NewTable()
	for _, file := range dir.hiddenfiles {
		filterTable.Append(lua.LString(file))
	}

	L.Push(filterTable)

	return 1
}

func luaDirSortignorecase(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.sortignorecase))
	return 1
}

func luaDirSortignoredia(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.sortignoredia))
	return 1
}

func luaDirNoPerm(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	L.Push(lua.LBool(dir.noPerm))
	return 1
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

func luaDirVisualSelectioins(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	tbl := L.NewTable()

	paths := dir.visualSelections()
	for _, path := range paths {
		tbl.Append(lua.LString(path))
	}

	L.Push(tbl)

	return 1
}

func luaDirSel(L *lua.LState) int {
	dir := LCheckDir(L, 1)
	name := L.CheckString(2)
	height := int(L.CheckNumber(3))
	dir.sel(name, height)
	return 0
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

// ----------------------------------------------------------------------------
// Type nav

const LuaNavTypeName = "lf.nav"

func LRegisterNavType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaNavTypeName)

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

		"curr_dir":               luaNavCurrDir,
		"curr_file":              luaNavCurrFile,
		"curr_selections":        luaNavCurrSelections,
		"curr_file_or_selection": luaNavCurrFileOrSelection,

		"get_tag": luaNavGetTag,
	}))

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

func luaNavGetDir(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)
	dir := nav.getDir(path)
	return LAddDirToState(L, dir)
}

func luaNavCdJumpListPrev(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.cdJumpListPrev()
	return 0
}

func luaNavCdJumpListNext(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.cdJumpListNext()
	return 0
}

func luaNavRenew(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.renew()
	return 0
}

func luaNavReload(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.reload()
	return 0
}

func luaNavSort(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.sort()
	return 0
}

func luaNavSetFilter(L *lua.LState) int {
	nav := LCheckNav(L, 1)
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
	nav := LCheckNav(L, 1)
	dist := L.CheckNumber(2)
	nav.up(int(dist))
	return 0
}

func luaNavDown(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	dist := L.CheckNumber(2)
	nav.down(int(dist))
	return 0
}

func luaNavScrollUp(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	dist := L.CheckNumber(2)
	nav.scrollUp(int(dist))
	return 0
}

func luaNavScrollDown(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	dist := L.CheckNumber(2)
	nav.scrollDown(int(dist))
	return 0
}

func luaNavUpDir(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.updir()
	return 0
}

func luaNavOpen(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.open()
	return 0
}

func luaNavTop(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.top()
	return 0
}

func luaNavBottom(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.bottom()
	return 0
}

func luaNavHigh(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.high()
	return 0
}

func luaNavMiddle(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.middle()
	return 0
}

func luaNavLow(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	nav.low()
	return 0
}

func luaNavMove(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	index := L.CheckNumber(2)
	nav.move(int(index))
	return 0
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

	if err := nav.tagToggle(tag); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavTag(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	tag := L.CheckString(2)

	if err := nav.tag(tag); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

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

func luaNavCd(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)

	if err := nav.cd(path); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavGlobSel(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	pattern := L.CheckString(2)
	invert := L.CheckBool(3)

	if err := nav.globSel(pattern, invert); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaNavCurrDir(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	return LAddDirToState(L, nav.currDir())
}

func luaNavCurrFile(L *lua.LState) int {
	nav := LCheckNav(L, 1)
	return LAddFileInfoToState(L, nav.currFile())
}

func luaNavCurrSelections(L *lua.LState) int {
	nav := LCheckNav(L, 1)

	tbl := L.NewTable()
	selections := nav.currSelections()
	for _, path := range selections {
		tbl.Append(lua.LString(path))
	}

	L.Push(tbl)

	return 1
}

func luaNavCurrFileOrSelection(L *lua.LState) int {
	nav := LCheckNav(L, 1)

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
	nav := LCheckNav(L, 1)
	path := L.CheckString(2)

	value, exists := nav.tags[path]
	if !exists {
		return 0
	}

	L.Push(lua.LString(value))

	return 1
}
