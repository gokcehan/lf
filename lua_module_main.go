package main

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func LfMainModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"cmd":   luaRunColonCommand,
		"shell": luaRunShellCommand,

		"set_opt":       luaSetOptionValue,
		"set_local_opt": luaSetLocalOptionValue,

		"glob_match": luaGlobMatch,
		"match_word": luaLuaMatchWord,
	})

	setupModuleConstants(L, mod)

	L.Push(mod)

	return 1
}

func setupModuleConstants(L *lua.LState, mod *lua.LTable) {
	mod.RawSetString("REGISTRY_SORT_METHOD", lua.LString(registryKeySortMethod))
	mod.RawSetString("REGISTRY_COMMAND", lua.LString(registryKeyCommand))
	mod.RawSetString("REGISTRY_EVENT_HOOK", lua.LString(registryKeyEventHook))
	mod.RawSetString("REGISTRY_PREVIEWER", lua.LString(registryKeyPreviewer))

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
}

func luaRunColonCommand(L *lua.LState) int {
	cmd := L.CheckString(1)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	p := newParser(strings.NewReader(cmd))
	for p.parse() {
		p.expr.eval(app, nil)
	}
	if p.err != nil {
		app.ui.echoerrf("%s", p.err)
	}

	return 0
}

func luaRunShellCommand(L *lua.LState) int {
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

func luaSetOptionValue(L *lua.LState) int {
	opt := L.CheckString(1)
	val := L.CheckString(2)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	expr := &setExpr{opt, val}
	expr.eval(app, nil)

	return 0
}

func luaSetLocalOptionValue(L *lua.LState) int {
	path := L.CheckString(1)
	opt := L.CheckString(2)
	val := L.CheckString(3)

	app, err := getAppObjectFromLuaGlobals(L)
	if err != nil {
		L.RaiseError("failed to get app object: %s", err)
		return 0
	}

	expr := &setLocalExpr{path, opt, val}
	expr.eval(app, nil)

	return 0
}

func luaGlobMatch(L *lua.LState) int {
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

func luaLuaMatchWord(L *lua.LState) int {
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
		tbl.Append(LWrapCompMatch(L, &match))
	}

	L.Push(tbl)
	L.Push(lua.LString(longest))

	return 2
}
