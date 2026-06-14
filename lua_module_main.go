package main

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func lfMainModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"print":      luaMainModulePrint,
		"list_ext":   luaMainModuleListExtend,
		"tbl_extend": luaMainModuleTblExtend,

		"cmd":    luaMainModuleRunColonCommand,
		"shell":  luaMainModuleRunShellCommand,
		"call":   luaMainModuleCallCommand,
		"call_n": luaMainModuleCallCommandN,

		"set_opt":       luaMainModuleSetOptionValue,
		"set_local_opt": luaMainModuleSetLocalOptionValue,
		"get_opt":       luaMainModuleGetOptionValue,
		"get_local_opt": luaMainModuleGetLocalOptionValue,

		"call_msg_expr": luaMainModuleCallMsgExpr,

		"glob_match":          luaMainModuleGlobMatch,
		"match_word":          luaMainModuleMatchWord,
		"str_fill":            luaMainModuleStrFill,
		"str_fill_right":      luaMainModuleStrFillRight,
		"to_perm_string":      luaMainModuleToPermString,
		"make_link_count_str": luaMainModuleMakeLinkCountStr,
		"make_user_name_str":  luaMainModuleMakeUserNameStr,
		"make_group_name_str": luaMainModuleMakeGroupNameStr,
		"sanitize_name":       luaMainModuleSanitizeName,
		"file_size_humanize":  luaMainModuleFileSizeHumanize,
		"disk_free_space":     luaMainModuleDiskFreeSpace,
	})

	setupModuleConstants(L, mod)

	L.Push(mod)

	return 1
}

func setupModuleConstants(L *lua.LState, mod *lua.LTable) {
	// constant
	mod.RawSetString("PREVIEW_LOADING_DELAY", lua.LNumber(previewLoadingDelay))

	// clipboard mode
	clipboardMode := L.NewTable()
	clipboardMode.RawSetString("Copy", lua.LNumber(clipboardCopy))
	clipboardMode.RawSetString("Cut", lua.LNumber(clipboardCut))
	mod.RawSetString("ClipboardMode", clipboardMode)

	// dir role
	dirRole := L.NewTable()
	dirRole.RawSetString("Active", lua.LNumber(Active))
	dirRole.RawSetString("Parent", lua.LNumber(Parent))
	dirRole.RawSetString("Preview", lua.LNumber(Preview))
	mod.RawSetString("DirRole", dirRole)

	// event type
	eventType := L.NewTable()
	eventType.RawSetString("PreCd", lua.LString("pre-cd"))
	eventType.RawSetString("OnCd", lua.LString("on-cd"))
	eventType.RawSetString("OnLoad", lua.LString("on-load"))
	eventType.RawSetString("OnFocus-gained", lua.LString("on-focus-gained"))
	eventType.RawSetString("OnFocus-lost", lua.LString("on-focus-lost"))
	eventType.RawSetString("OnInit", lua.LString("on-init"))
	eventType.RawSetString("OnRedraw", lua.LString("on-redraw"))
	eventType.RawSetString("OnSelect", lua.LString("on-select"))
	eventType.RawSetString("OnQuit", lua.LString("on-quit"))
	mod.RawSetString("EventType", eventType)

	// shell command type
	shellCmdType := L.NewTable()
	shellCmdType.RawSetString("Normal", lua.LString("$"))
	shellCmdType.RawSetString("Pipe", lua.LString("%"))
	shellCmdType.RawSetString("Wait", lua.LString("!"))
	shellCmdType.RawSetString("Async", lua.LString("&"))
	mod.RawSetString("ShellCmdType", shellCmdType)

	// key map type
	keyMapType := L.NewTable()
	keyMapType.RawSetString("Normal", lua.LString(luaKeyMapTypeNormal))
	keyMapType.RawSetString("Visual", lua.LString(luaKeyMapTypeVisual))
	keyMapType.RawSetString("Command", lua.LString(luaKeyMapTypeCommand))
	mod.RawSetString("KeyMapType", keyMapType)

	// ui formatter type
	uiFormatterType := L.NewTable()
	uiFormatterType.RawSetString("cursoractive", lua.LString(luaUIFormatterCursorActive))
	uiFormatterType.RawSetString("cursorparent", lua.LString(luaUIFormatterCursorParent))
	uiFormatterType.RawSetString("cursorpreview", lua.LString(luaUIFormatterCursorPreview))
	uiFormatterType.RawSetString("error", lua.LString(luaUIFormatterError))
	uiFormatterType.RawSetString("numbercursor", lua.LString(luaUIFormatterNumberCursor))
	uiFormatterType.RawSetString("number", lua.LString(luaUIFormatterNumber))
	uiFormatterType.RawSetString("tag", lua.LString(luaUIFormatterTag))
	mod.RawSetString("UIFormatterType", uiFormatterType)

	// ui printer type
	uiPrinterType := L.NewTable()
	uiPrinterType.RawSetString("file", lua.LString(luaUIPrinterFile))
	uiPrinterType.RawSetString("ruler", lua.LString(luaUIPrinterRuler))
	uiPrinterType.RawSetString("prompt", lua.LString(luaUIPrinterPrompt))
	mod.RawSetString("UIPrinterType", uiPrinterType)

	// ui style type
	uiStyleType := L.NewTable()
	uiStyleType.RawSetString("border", lua.LString(luaUIStyleBorder))
	uiStyleType.RawSetString("copy", lua.LString(luaUIStyleCopy))
	uiStyleType.RawSetString("cut", lua.LString(luaUIStyleCut))
	uiStyleType.RawSetString("menu", lua.LString(luaUIStyleMenu))
	uiStyleType.RawSetString("menuheader", lua.LString(luaUIStyleMenuheader))
	uiStyleType.RawSetString("menuselect", lua.LString(luaUIStyleMenuselect))
	uiStyleType.RawSetString("select", lua.LString(luaUIStyleSelect))
	uiStyleType.RawSetString("visual", lua.LString(luaUIStyleVisual))
	mod.RawSetString("UIStyleType", uiStyleType)
}

// ----------------------------------------------------------------------------

func prettyPrintLuaValue(builder *strings.Builder, val lua.LValue, visited map[lua.LValue]int, tableCnt, indentLevel int) int {
	switch val.Type() {
	case lua.LTNil, lua.LTBool, lua.LTNumber, lua.LTFunction, lua.LTUserData, lua.LTThread, lua.LTChannel:
		builder.WriteString(val.String())
	case lua.LTString:
		fmt.Fprintf(builder, "%q", val.String())
	case lua.LTTable:
		mark := visited[val]
		if mark > 0 {
			builder.WriteString("table<")
			builder.WriteString(strconv.Itoa(mark))
			builder.WriteString(">")
			return tableCnt
		}

		tableCnt++
		visited[val] = tableCnt

		builder.WriteString("{")

		isEmpty := true
		val.(*lua.LTable).ForEach(func(key, value lua.LValue) {
			isEmpty = false

			builder.WriteString("\n")
			for range indentLevel + 1 {
				builder.WriteString("  ")
			}
			tableCnt = prettyPrintLuaValue(builder, key, visited, tableCnt, indentLevel+1)
			builder.WriteString(" = ")
			tableCnt = prettyPrintLuaValue(builder, value, visited, tableCnt, indentLevel+1)
			builder.WriteString(",")
		})

		if !isEmpty {
			builder.WriteString("\n")
			for range indentLevel {
				builder.WriteString("  ")
			}
		}
		builder.WriteString("}")
	}

	return tableCnt
}

// luaMainModulePrint prints a Lua value, this can be used for debugging.
func luaMainModulePrint(L *lua.LState) int {
	value := L.CheckAny(1)
	var builder strings.Builder
	prettyPrintLuaValue(&builder, value, make(map[lua.LValue]int), 0, 0)
	log.Println(builder.String())
	return 0
}

// luaMainModuleListExtend takes 2 table, append all elements in `src` table into
// `dst` table.
func luaMainModuleListExtend(L *lua.LState) int {
	dst := L.CheckTable(1)
	src := L.CheckTable(2)
	stValue := L.Get(3)
	edValue := L.Get(4)

	nElem := src.Len()
	st := 1
	ed := nElem

	if stValue.Type() == lua.LTNumber {
		st = int(stValue.(lua.LNumber))
		if st < 1 {
			st = 0
		}
	}
	if edValue.Type() == lua.LTNumber {
		ed = int(edValue.(lua.LNumber))
		if ed > nElem {
			ed = nElem
		}
	}

	for i := st; i <= ed; i++ {
		dst.Append(src.RawGetInt(i))
	}

	L.Push(dst)

	return 1
}

// luaMainModuleTableExtend merges two or more tables.
func luaMainModuleTblExtend(L *lua.LState) int {
	behaviorValue := L.CheckAny(1)

	var behavior string
	var checker *lua.LFunction

	switch behaviorValue.Type() {
	case lua.LTString:
		behavior = string(behaviorValue.(lua.LString))
		switch behavior {
		case "error", "keep", "force":
			// ok
		default:
			L.ArgError(1, "unsupport behavior value")
		}
	case lua.LTFunction:
		checker = behaviorValue.(*lua.LFunction)
	default:
		L.ArgError(1, "expected string or function")
		return 0
	}

	dst := L.NewTable()

	offset := 2
	nArgs := L.GetTop()
	for i := offset; i <= nArgs; i++ {
		tbl := L.CheckTable(i)
		tbl.ForEach(func(key, value lua.LValue) {
			oldValue := dst.RawGet(key)

			if checker != nil {
				L.CallByParam(
					lua.P{
						Fn:   checker,
						NRet: 1,
					},
					key,
					oldValue,
					value,
				)

				newValue := L.Get(-1)
				L.Pop(1)

				dst.RawSet(key, newValue)
			} else if oldValue == lua.LNil {
				dst.RawSet(key, value)
			} else {
				switch behavior {
				case "error":
					L.RaiseError("key duplicated: %s", key)
				case "keep":
					// pass
				case "force":
					dst.RawSet(key, value)
				}
			}
		})
	}

	L.Push(dst)

	return 1
}

// ----------------------------------------------------------------------------

// luaMainModuleRunColonCommand runs a lf command string just like calling command with `:`.
func luaMainModuleRunColonCommand(L *lua.LState) int {
	cmd := L.CheckString(1)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	p := newParser(strings.NewReader(cmd))
	for p.parse() {
		app.ui.exprChan <- p.expr
	}
	if p.err != nil {
		app.ui.echoerrf("%s", p.err)
	}

	return 0
}

// luaMainModuleRunShellCommand takes execution type prefix, command name and variable length
// argument list, and asks lf to execute given shell command.
func luaMainModuleRunShellCommand(L *lua.LState) int {
	prefix := L.CheckString(1)
	cmd := L.CheckString(2)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	st := 3
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		arg := L.Get(i)
		args[i-st] = arg.String()
	}

	switch prefix {
	case "$":
		log.Printf("shell: %s -- %q", cmd, args)
		app.runShell(cmd, args, prefix)
	case "%":
		log.Printf("shell-pipe: %s -- %q", cmd, args)
		app.runShell(cmd, args, prefix)
	case "!":
		log.Printf("shell-wait: %s -- %q", cmd, args)
		app.runShell(cmd, args, prefix)
	case "&":
		log.Printf("shell-async: %s -- %q", cmd, args)
		app.runShell(cmd, args, prefix)
	default:
		log.Printf("unknown execution prefix: %q", prefix)
	}

	return 0
}

// luaMainModuleCallCommand runs lf command.
func luaMainModuleCallCommand(L *lua.LState) int {
	name := L.CheckString(1)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		arg := L.Get(i)
		args[i-st] = arg.String()
	}

	app.ui.exprChan <- &callExpr{name, args, 1}

	return 0
}

// luaMainModuleCallCommandN runs lf command with repetition argument `n`.
func luaMainModuleCallCommandN(L *lua.LState) int {
	count := L.CheckInt(1)
	name := L.CheckString(2)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	st := 3
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		arg := L.Get(i)
		args[i-st] = arg.String()
	}

	app.ui.exprChan <- &callExpr{name, args, count}

	return 0
}

func luaMainModuleSetOptionValue(L *lua.LState) int {
	opt := L.CheckString(1)
	val := L.CheckString(2)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	app.ui.exprChan <- &setExpr{opt, val}

	return 0
}

func luaMainModuleSetLocalOptionValue(L *lua.LState) int {
	path := L.CheckString(1)
	opt := L.CheckString(2)
	val := L.CheckString(3)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	app.ui.exprChan <- &setLocalExpr{path, opt, val}

	return 0
}

func luaMainModuleGetOptionValue(L *lua.LState) int {
	opt := L.CheckString(1)

	rValue := reflect.ValueOf(gOpts)
	field := rValue.FieldByName(opt)

	if !field.IsValid() {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("option %q does not exist", opt)))
		return 2
	}

	luaValue, err := goReflectValueToLuaValue(L, field)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("error converting option value: %s", err)))
		return 2
	}

	L.Push(luaValue)

	return 1
}

func luaMainModuleGetLocalOptionValue(L *lua.LState) int {
	path := L.CheckString(1)
	opt := L.CheckString(2)

	rValue := reflect.ValueOf(gLocalOpts)
	field := rValue.FieldByName(opt)

	if !field.IsValid() {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("option %q does not exist", opt)))
		return 2
	}

	if field.Kind() != reflect.Map {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("option %q is not a map field", opt)))
		return 2
	}

	value := field.MapIndex(reflect.ValueOf(path))
	if !value.IsValid() {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}

	luaValue, err := goReflectValueToLuaValue(L, value)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("error converting option value: %s", err)))
		return 2
	}

	L.Push(luaValue)

	return 1
}

// ----------------------------------------------------------------------------

func luaMainModuleCallMsgExpr(L *lua.LState) int {
	expr := lCheckLuaMsgExpr(L, 1)

	action, err := getLuaMsgAction(L, expr.sourceName, expr.registry, expr.msg, expr.variant)
	if err != nil {
		L.RaiseError("%s", err)
		return 0
	}

	L.Replace(1, action)
	L.Call(L.GetTop()-1, lua.MultRet)

	return L.GetTop()
}

// ----------------------------------------------------------------------------

// luaMainModuleGlobMatch checks if a pattern matches certain string.
func luaMainModuleGlobMatch(L *lua.LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)

	match, err := filepath.Match(pattern, str)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(fmt.Sprintf("glob match error: %s", err)))
		return 2
	}

	L.Push(lua.LBool(match))

	return 1
}

// luaMainModuleMatchWord takes a source string, and a list of candidate string, and
// returns a list of match object and longest common matched string.
func luaMainModuleMatchWord(L *lua.LState) int {
	longest := L.CheckString(1)
	wordTbl := L.CheckTable(2)

	nWord := wordTbl.Len()
	words := make([]string, nWord)

	for i := 1; i <= nWord; i++ {
		word := wordTbl.RawGetInt(i)
		words[i-1] = word.String()
	}

	slices.Sort(words)
	matches, longest := matchWord(longest, slices.Compact(words))

	tbl := L.NewTable()
	for _, match := range matches {
		tbl.Append(lWrapCompMatch(L, &match))
	}

	L.Push(tbl)
	L.Push(lua.LString(longest))

	return 2
}

func luaMainModuleStrFill(L *lua.LState) int {
	base := L.CheckString(1)
	width := L.CheckInt(2)
	fillStrValue := L.Get(3)

	fillStr := " "
	switch fillStrValue.Type() {
	case lua.LTString:
		fillStr = string(fillStrValue.(lua.LString))
	case lua.LTNil:
		// pass
	default:
		L.ArgError(3, "a string is expected")
	}

	baseLen := len(base)
	fillLen := len(fillStr)
	repeatCnt := (width - baseLen) / fillLen
	if repeatCnt <= 0 {
		L.Push(lua.LString(base))
		return 1
	}

	L.Push(lua.LString(strings.Repeat(fillStr, repeatCnt) + base))

	return 1
}

func luaMainModuleStrFillRight(L *lua.LState) int {
	base := L.CheckString(1)
	width := L.CheckInt(2)
	fillStrValue := L.Get(3)

	fillStr := " "
	switch fillStrValue.Type() {
	case lua.LTString:
		fillStr = string(fillStrValue.(lua.LString))
	case lua.LTNil:
		// pass
	default:
		L.ArgError(3, "a string is expected")
	}

	baseLen := len(base)
	fillLen := len(fillStr)
	repeatCnt := (width - baseLen) / fillLen
	if repeatCnt <= 0 {
		L.Push(lua.LString(base))
		return 1
	}

	L.Push(lua.LString(base + strings.Repeat(fillStr, repeatCnt)))

	return 1
}

func luaMainModuleToPermString(L *lua.LState) int {
	mod := lCheckFileMode(L, 1)
	L.Push(lua.LString(permString(mod)))
	return 1
}

func luaMainModuleMakeLinkCountStr(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(linkCount(file)))
	return 1
}

func luaMainModuleMakeUserNameStr(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(userName(file)))
	return 1
}

func luaMainModuleMakeGroupNameStr(L *lua.LState) int {
	file := lCheckFile(L, 1)
	L.Push(lua.LString(groupName(file)))
	return 1
}

func luaMainModuleSanitizeName(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LString(sanitizeName(str)))
	return 1
}

func luaMainModuleFileSizeHumanize(L *lua.LState) int {
	size := L.CheckInt64(1)
	L.Push(lua.LString(humanize(size)))
	return 1
}

func luaMainModuleDiskFreeSpace(L *lua.LState) int {
	pathValue := L.Get(1)

	path := "."
	switch pathValue.Type() {
	case lua.LTString:
		path = string(pathValue.(lua.LString))
	case lua.LTNil:
		// pass
	default:
		L.ArgError(1, "string expected")
	}

	L.Push(lua.LString(diskFree(path)))

	return 1
}
