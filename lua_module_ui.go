package main

import (
	"fmt"

	"github.com/clipperhouse/displaywidth"
	lua "github.com/yuin/gopher-lua"
)

func lfUIModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"print_length":  lfUIModulePrintLength,
		"display_width": lfUIModuleDisplayWidth,

		"get_formatter":                   luaUIModuleGetUIFormatter,
		"get_printer":                     luaUIModuleGetUIPrinter,
		"call_formatter_with_default_str": luaUIModuleCallFormatterWithDefaultStr,
		"get_style_with_default_str":      luaUIModuleGetStyleWithDefaultStr,
		"format_option_str":               luaUIModuleFormatUIOptionStr,
		"get_file_display_info":           luaUIModuleGetFileDisplayInfo,
		"truncate_filename":               luaUIModuleTruncateFilename,
		"option_to_fmtstr":                luaUIModuleOptionToFmtstr,
		"strip_term_sequence":             luaUIModuleStripTermSequence,

		"print_dir_entries": luaUIModulePrintDirEntries,
	})

	L.Push(mod)

	return 1
}

// lfUIModulePrintLength returns displayed width of string content in terminal cells.
//
// It ignores supported terminal control sequences and accounts for tab
// expansions using the `tabstop` option.
func lfUIModulePrintLength(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LNumber(printLength(str)))
	return 1
}

// lfUIModuleDisplayWidth calculates the display width of a string, by iterating
// over grapheme clusters in the string and summing their widths.
func lfUIModuleDisplayWidth(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LNumber(displaywidth.String(str)))
	return 1
}

func luaUIModuleGetUIFormatter(L *lua.LState) int {
	name := L.CheckString(1)
	formatter := getLuaUIFormatter(name)
	return lAddLuaMsgExprToState(L, formatter)
}

func luaUIModuleGetUIPrinter(L *lua.LState) int {
	name := L.CheckString(1)
	expr := getLuaUIPrinter(name)
	return lAddLuaMsgExprToState(L, expr)
}

func luaUIModuleCallFormatterWithDefaultStr(L *lua.LState) int {
	name := L.CheckString(1)
	defaultFmtStr := L.CheckString(2)
	nArgs := L.GetTop()

	expr := getLuaUIFormatter(name)
	if expr == nil {
		offset := 3
		param := make([]any, nArgs-offset+1)
		for i := offset; i <= nArgs; i++ {
			param[i-offset] = L.Get(i).String()
		}

		result := fmt.Sprintf(optionToFmtstr(defaultFmtStr), param...)
		L.Push(lua.LString(result))

		return 1
	}

	action, err := getLuaMsgAction(L, expr.sourceName, expr.registry, expr.msg, expr.variant)
	if err != nil {
		L.RaiseError("%s", err)
		return 0
	}

	L.Replace(1, action)
	for i := 2; i <= nArgs; i++ {
		L.Replace(i, L.Get(i+1))
	}
	L.Call(L.GetTop()-1, lua.MultRet)

	return L.GetTop()
}

func luaUIModuleGetStyleWithDefaultStr(L *lua.LState) int {
	name := L.CheckString(1)
	defaultFmtStr := L.CheckString(2)
	st := getLuaUIStyleWithDefaultStr(name, defaultFmtStr)
	return lAddTcellStyleToState(L, &st)
}

func luaUIModuleFormatUIOptionStr(L *lua.LState) int {
	fmtStr := L.CheckString(1)

	offset := 2
	nArg := L.GetTop()
	param := make([]any, nArg-offset+1)
	for i := offset; i <= nArg; i++ {
		param[i-offset] = L.Get(i).String()
	}

	result := fmt.Sprintf(optionToFmtstr(fmtStr), param...)
	L.Push(lua.LString(result))

	return 1
}

func luaUIModuleGetFileDisplayInfo(L *lua.LState) int {
	file := lCheckFile(L, 1)
	dir := lCheckDir(L, 2)
	userWidth := L.CheckInt(3)
	groupWidth := L.CheckInt(4)
	customWidth := L.CheckInt(5)

	info, custom, customOff := fileInfo(file, dir, userWidth, groupWidth, customWidth)
	L.Push(lua.LString(info))
	L.Push(lua.LString(custom))
	L.Push(lua.LNumber(customOff))

	return 3
}

func luaUIModuleTruncateFilename(L *lua.LState) int {
	file := lCheckFile(L, 1)
	maxWidth := L.CheckInt(2)
	filename := truncateFilename(file, maxWidth, gOpts.truncatepct, gOpts.truncatechar)
	L.Push(lua.LString(filename))
	return 1
}

func luaUIModuleOptionToFmtstr(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LString(optionToFmtstr(str)))
	return 1
}

func luaUIModuleStripTermSequence(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LString(stripTermSequence(str)))
	return 1
}

// luaUIModulePrintDirEntries is default implmenetation of printing given list
// of directory entries.
func luaUIModulePrintDirEntries(L *lua.LState) int {
	tryRaiseSyncLuaStateError(L)

	win := lCheckWin(L, 1)
	screen := lCheckTcellScreen(L, 2)
	context := lCheckPrintDirEntryContext(L, 3)
	fileTbl := L.CheckTable(4)

	nFile := fileTbl.Len()
	files := make([]*file, nFile)
	for i := 1; i <= nFile; i++ {
		value := fileTbl.RawGetInt(i)
		if ud, ok := value.(*lua.LUserData); ok {
			if file, ok := ud.Value.(*file); ok {
				files[i-1] = file
			} else {
				L.ArgError(4, fmt.Sprintf("element #%d is not a file object", i))
			}
		} else {
			L.ArgError(4, fmt.Sprintf("element #%d is not userdata", i))
		}
	}

	if !tryPrintDirEntriesWithLua(win, screen, context, files) {
		for i, f := range files {
			printDirEntry(win, screen, context, i, f)
		}
	}

	return 0
}
