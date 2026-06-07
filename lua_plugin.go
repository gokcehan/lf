package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

const pluginDirName = "plugins"

const (
	registryKeySortMethod = "sort_method"
	registryKeyCommand    = "command"
	registryKeyEventHook  = "event_hook"
	registryKeyPreviewer  = "previewer"
)

const (
	luaPreviewerConditionFuncKey = "condition"
	luaPreviewerActionFuncKey    = "action"
)

type luaStateBox struct {
	luaState *lua.LState
	lock     sync.Mutex
}

func (box *luaStateBox) acquire() *lua.LState {
	box.lock.Lock()
	return box.luaState
}

func (box *luaStateBox) release() {
	box.lock.Unlock()
}

type lStatePool struct {
	m     sync.Mutex
	saved []*lua.LState

	app             *app
	rootDirs        []string
	pluginByteCodes []*lua.FunctionProto
}

func newLStatePool(app *app) *lStatePool {
	return &lStatePool{
		saved: make([]*lua.LState, 0),
		app:   app,
	}
}

func (pl *lStatePool) setRootDirs(configRoots []string) {
	rootDirs := make([]string, len(configRoots))

	for i, configRoot := range configRoots {
		rootDirs[i] = filepath.Join(configRoot, pluginDirName)
	}

	pl.rootDirs = rootDirs
}

func (pl *lStatePool) addConfigRoot(configRoot string) {
	pl.rootDirs = append(pl.rootDirs, filepath.Join(configRoot, pluginDirName))
}

func (pl *lStatePool) loadPluginScripts() error {
	pluginByteCodes := make([]*lua.FunctionProto, 0)

	errorCnt := 0

	for _, pluginDir := range pl.rootDirs {
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(pluginDir)
		if err != nil {
			errorCnt++
			log.Printf("failed to read plugin directory %s: %s", pluginDir, err)
			continue
		}

		// only directories are treated as plugin entrance.
		// So that user can put Lua development config files under plugin root with ease.
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			scriptPath := filepath.Join(pluginDir, name, "init.lua")

			if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
				proto, err := compileLua(scriptPath)
				if err != nil {
					errorCnt++
					log.Printf("failed to compile plugin script: %s\n%s", scriptPath, err)
				} else {
					log.Printf("plugin script loaded: %s", scriptPath)
					pluginByteCodes = append(pluginByteCodes, proto)
				}
			}
		}
	}

	if errorCnt > 0 {
		return fmt.Errorf("%d error(s) occured while loading plugin script, see log for details", errorCnt)
	}

	pl.pluginByteCodes = pluginByteCodes

	return nil
}

func (pl *lStatePool) get() *lua.LState {
	pl.m.Lock()
	defer pl.m.Unlock()
	n := len(pl.saved)
	if n == 0 {
		return pl.new()
	}
	x := pl.saved[n-1]
	pl.saved = pl.saved[0 : n-1]
	return x
}

func (pl *lStatePool) newWithRetAction(action func(sourceName string, L *lua.LState, tbl *lua.LTable)) *lua.LState {
	L := lua.NewState()

	pl.setupLStateEnvironment(L)

	for _, proto := range pl.pluginByteCodes {
		err := doCompiledFile(L, proto)
		if err != nil {
			log.Printf("failed to execute plugin script: %s\n%s", proto.SourceName, err)
		}

		ret := L.Get(1)
		nRet := L.GetTop()
		L.Pop(nRet)

		if ret.Type() == lua.LTNil {
			if action != nil {
				action(proto.SourceName, L, nil)
			}
		} else if ret.Type() == lua.LTTable {
			sourceName := proto.SourceName
			tbl := ret.(*lua.LTable)

			if gLuaRegistry.stateDataMap == nil {
				gLuaRegistry.stateDataMap = make(map[*lua.LState]map[string]*lua.LTable)
			}

			registryMap, ok := gLuaRegistry.stateDataMap[L]
			if !ok {
				registryMap = make(map[string]*lua.LTable)
				gLuaRegistry.stateDataMap[L] = registryMap
			}

			registryMap[sourceName] = tbl

			if action != nil {
				action(sourceName, L, tbl)
			}
		} else {
			log.Println("plugin script", proto.SourceName, "did not return a table")
		}
	}

	return L
}

func (pl *lStatePool) new() *lua.LState {
	return pl.newWithRetAction(nil)
}

func (pl *lStatePool) newWithRegistryUpdate() *lua.LState {
	return pl.newWithRetAction(func(sourceName string, L *lua.LState, tbl *lua.LTable) {
		if tbl == nil {
			return
		}

		log.Println("update Lua registry for plugin script:", sourceName)

		loadEventHookRegistryFromTbl(sourceName, tbl)
		loadSortMethodRegistryFromTbl(sourceName, tbl)
		loadCommandRegistryFromTbl(sourceName, tbl)

		loadPreviewerRegistryFromTbl(sourceName, tbl)
		sortLuaPreviewers()
	})
}

func (pl *lStatePool) setupLStateEnvironment(L *lua.LState) {
	setupLuaTypeBindings(L)

	// setup modules
	err := setupScripImportPath(L, pl.rootDirs)
	if err != nil {
		log.Printf("failed to setup Lua loader search path: %s", err)
	}

	setupLuaGlobals(pl.app, L)

	L.PreloadModule("lf", LfMainModuleLoader)
	L.PreloadModule("lf.fs", LfFsModuleLoader)
	L.PreloadModule("lf.utf8", LfUtf8ModuleLoader)
}

func (pl *lStatePool) put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.saved = append(pl.saved, L)
}

func (pl *lStatePool) shutdown() {
	for _, L := range pl.saved {
		L.Close()
	}
}

// Global LState pool, used for asynchronous execution
var gLuaPool *lStatePool

// Global LState instance with lock, used for synchronous execution
var gLuaStateSync *luaStateBox

type luaPreviewerInfo struct {
	priority int
	name     string
	msgexpr  luaMsgExpr
}

var gLuaRegistry struct {
	stateDataMap map[*lua.LState]map[string]*lua.LTable

	sortMethod map[string]luaMsgExpr
	eventHooks map[string][]luaMsgExpr
	previewers []luaPreviewerInfo
}

// lgFuncUnconditional is a Lua function that always returns true.
var lgFuncUnconditional = lua.LGFunction(func(L *lua.LState) int {
	L.Push(lua.LTrue)
	return 1
})

// ----------------------------------------------------------------------------

// goValueToLuaValue converts Go value to lua.LValue.
func goValueToLuaValue(value any) (lua.LValue, error) {
	var err error

	lValue, ok := value.(lua.LValue)
	if ok {
		return lValue, err
	}

	switch v := value.(type) {
	case int, int16, int32, int64, float32, float64:
		lValue = lua.LNumber(reflect.ValueOf(v).Convert(reflect.TypeOf(float64(0))).Float())
	case string:
		lValue = lua.LString(v)
	case bool:
		if v {
			lValue = lua.LTrue
		} else {
			lValue = lua.LFalse
		}
	default:
		err = fmt.Errorf("unsupported element value type: %T", value)
		lValue = lua.LNil
	}

	return lValue, err
}

// compileLua reads the passed lua file from disk and compiles it.
func compileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

// doCompiledFile takes a FunctionProto, as returned by CompileLua, and runs it in the LState. It is equivalent
// to calling DoFile on the LState with the original source file.
func doCompiledFile(L *lua.LState, proto *lua.FunctionProto) error {
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)
	return L.PCall(0, lua.MultRet, nil)
}

// setupLuaTypeBindings adds `lf_types` global table as entrance of accessing
// Go type binding meta tables.
func setupLuaTypeBindings(L *lua.LState) {
	lfTypes := L.NewTable()

	lfTypes.RawSetString("App", LRegisterAppType(L))
	lfTypes.RawSetString("UI", LRegisterUIType(L))

	lfTypes.RawSetString("File", LRegisterFileTypeMt(L))
	lfTypes.RawSetString("Dir", LRegisterDirType(L))
	lfTypes.RawSetString("Nav", LRegisterNavType(L))

	lfTypes.RawSetString("FileInfo", LRegisterFileInfoType(L))

	lfTypes.RawSetString("BufWriter", LRegisterBufWriterType(L))

	L.SetGlobal("lf_types", lfTypes)
}

// setupScripImportPath appends plugin root directory paths to Lua loader search
// list.
func setupScripImportPath(L *lua.LState, runtimeDirs []string) error {
	pack, ok := L.GetGlobal("package").(*lua.LTable)
	if !ok {
		return fmt.Errorf("failed to retrive global variable `package`")
	}

	pathVal, ok := L.GetField(pack, "path").(lua.LString)
	if !ok {
		return fmt.Errorf("`path` field of `package` table is not a string")
	}

	path := string(pathVal)

	var builder strings.Builder
	builder.WriteString(path)
	for _, dir := range runtimeDirs {
		builder.WriteString(";")
		builder.WriteString(dir)
		builder.WriteString("/?.lua")

		builder.WriteString(";")
		builder.WriteString(dir)
		builder.WriteString("/?/init.lua")
	}

	path = builder.String()

	L.SetField(pack, "path", lua.LString(path))

	return nil
}

// setupLuaGlobals setup global variables.
func setupLuaGlobals(app *app, L *lua.LState) {
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		nargs := L.GetTop()
		values := make([]any, nargs)

		for i := range nargs {
			value := L.Get(i + 1)
			values[i] = value.String()
		}

		log.Println(values...)

		return 0
	}))

	L.SetGlobal("app", LWrapApp(L, app))
}

// ----------------------------------------------------------------------------

func loadEventHookRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyEventHook

	value := tbl.RawGetString(registryKey)
	switch value.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	eventHookTbl := value.(*lua.LTable)
	eventHookTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("event hook registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		if value.Type() != lua.LTFunction {
			log.Printf("event hook registry value is expected to be function, found %s: %s", value.Type(), value)
			return
		}

		if gLuaRegistry.eventHooks == nil {
			gLuaRegistry.eventHooks = make(map[string][]luaMsgExpr)
		}

		name := key.String()
		gLuaRegistry.eventHooks[name] = append(
			gLuaRegistry.eventHooks[name],
			luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        name,
				// isSync:      false,
			},
		)

		log.Printf("add event hook: %s", name)
	})
}

func loadSortMethodRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeySortMethod

	value := tbl.RawGetString(registryKey)
	switch value.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	sortMethodTbl := value.(*lua.LTable)
	sortMethodTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("sort method registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		if value.Type() != lua.LTFunction {
			log.Printf("sort method registry value is expected to be function, found %s: %s", value.Type(), value)
			return
		}

		if gLuaRegistry.sortMethod == nil {
			gLuaRegistry.sortMethod = make(map[string]luaMsgExpr)
		}

		name := key.String()
		gLuaRegistry.sortMethod[name] = luaMsgExpr{
			sourceName: sourceName,
			registry:   registryKey,
			msg:        name,
			// isSync:      false,
		}

		log.Printf("add sort method: %s", name)
	})
}

func loadCommandRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyCommand

	value := tbl.RawGetString(registryKey)
	switch value.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	cmdTbl := value.(*lua.LTable)
	cmdTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("command registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		name := key.String()

		log.Printf("add command: %s", name)
		switch value.Type() {
		case lua.LTString:
			text := value.String()
			p := newParser(strings.NewReader(text))
			expr := p.parseExpr()
			if expr == nil {
				log.Printf("failed to parse Lua command: %s", text)
			} else {
				gOpts.cmds[name] = expr
			}
		case lua.LTFunction:
			gOpts.cmds[name] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        name,
			}
		default:
			log.Printf("command registry value is expected to be string or function, get %s: %s", value.Type(), value)
		}
	})
}

func loadPreviewerRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyPreviewer

	value := tbl.RawGetString(registryKey)
	switch value.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	registryTbl := value.(*lua.LTable)
	registryTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("previewer registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		switch value.Type() {
		case lua.LTFunction, lua.LTTable:
			// ok
		default:
			log.Printf("previewer registry value is expected to be function or table, found %s: %s", value.Type(), value)
			return
		}

		msg := key.String()
		name := filepath.Base(filepath.Dir(sourceName)) + "." + msg
		gLuaRegistry.previewers = append(gLuaRegistry.previewers, luaPreviewerInfo{
			priority: 0,
			name:     name,
			msgexpr: luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        msg,
				// isSync:      false,
			},
		})

		log.Printf("add previewer: %s", name)
	})
}

func sortLuaPreviewers() {
	slices.SortStableFunc(gLuaRegistry.previewers, func(a, b luaPreviewerInfo) int {
		if a.priority < b.priority {
			return 1
		} else if a.priority > b.priority {
			return -1
		}

		if a.name < b.name {
			return -1
		} else if a.name > b.name {
			return 1
		}

		return 0
	})
}

func setLuaPreviewerPriority(name string, priority int, withSort bool) bool {
	changed := false

	for i := range gLuaRegistry.previewers {
		if gLuaRegistry.previewers[i].name == name {
			if gLuaRegistry.previewers[i].priority != priority {
				changed = true
				gLuaRegistry.previewers[i].priority = priority
			}
			break
		}
	}

	if changed && withSort {
		sortLuaPreviewers()
	}

	return changed
}

// ----------------------------------------------------------------------------

func getLuaMsgEntry(L *lua.LState, sourceName string, registryKey string, msg string) (lua.LValue, error) {
	registryMap := gLuaRegistry.stateDataMap[L]
	if registryMap == nil {
		return nil, fmt.Errorf("no registry data found for current Lua state")
	}

	tbl := registryMap[sourceName]
	if tbl == nil {
		return nil, fmt.Errorf("invalid msg source name: %s", sourceName)
	}

	value := tbl.RawGetString(registryKey)
	switch value.Type() {
	case lua.LTNil:
		return nil, fmt.Errorf("no handler table found for registry key `%s`", registryKey)
	case lua.LTTable:
		// ok
	default:
		return nil, fmt.Errorf("unexpected type of registry value for key `%s`: %s", registryKey, value.Type())
	}

	handlerTbl := value.(*lua.LTable)
	handler := handlerTbl.RawGetString(msg)

	return handler, nil
}

type luaMsgHandlerExtractor func(L *lua.LState, msgEntry lua.LValue) *lua.LFunction
type luaMsgArgsMaker func(L *lua.LState) []lua.LValue

type luaMsgCallArgs struct {
	sourceName, registryKey, msg string
	handlerExtractor             luaMsgHandlerExtractor
	getArgs                      luaMsgArgsMaker
}

var handlerExtractorMap = map[string]luaMsgHandlerExtractor{
	registryKeyPreviewer: func(L *lua.LState, msgEntry lua.LValue) *lua.LFunction {
		switch msgEntry.Type() {
		case lua.LTFunction:
			return msgEntry.(*lua.LFunction)
		case lua.LTTable:
			actionFunc := msgEntry.(*lua.LTable).RawGetString(luaPreviewerActionFuncKey)
			if actionFunc.Type() == lua.LTFunction {
				return actionFunc.(*lua.LFunction)
			}
		}
		// invalid previewer
		return nil
	},
}

func callLuaCmdOnState(L *lua.LState, callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	entry, err := getLuaMsgEntry(L, callArgs.sourceName, callArgs.registryKey, callArgs.msg)
	if err != nil {
		return nil, err
	}

	var handler *lua.LFunction
	extractor := callArgs.handlerExtractor
	if extractor == nil {
		extractor = handlerExtractorMap[callArgs.registryKey]
	}

	if extractor != nil {
		handler = extractor(L, entry)
	} else {
		handler, _ = entry.(*lua.LFunction)
	}
	if handler == nil {
		return nil, fmt.Errorf("can't get valid msg handler function")
	}

	var luaArgs []lua.LValue
	if callArgs.getArgs != nil {
		luaArgs = callArgs.getArgs(L)
	}

	L.Push(handler)
	for _, arg := range luaArgs {
		L.Push(arg)
	}

	log.Printf("call Lua msg: (%s, %s, %s): %q", callArgs.sourceName, callArgs.registryKey, callArgs.msg, luaArgs)
	err = L.PCall(len(luaArgs), lua.MultRet, nil)
	if err != nil {
		return nil, err
	}

	nRet := L.GetTop()
	defer L.Pop(nRet)
	if nRet <= 0 {
		return nil, nil
	}

	ret := make([]lua.LValue, nRet)
	for i := 0; i < nRet; i++ {
		ret[i] = L.Get(i + 1)
	}

	return ret, nil
}

func callLuaMsg(callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	L := gLuaPool.get()
	defer gLuaPool.put(L)
	return callLuaCmdOnState(L, callArgs)
}

func callLuaMsgSync(callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	L := gLuaStateSync.acquire()
	defer gLuaStateSync.release()
	return callLuaCmdOnState(L, callArgs)
}

func callLuaMsgExpr(expr *luaMsgExpr, handlerExtractor luaMsgHandlerExtractor, getArgs luaMsgArgsMaker) ([]lua.LValue, error) {
	callArgs := luaMsgCallArgs{
		sourceName:       expr.sourceName,
		registryKey:      expr.registry,
		msg:              expr.msg,
		handlerExtractor: handlerExtractor,
		getArgs:          getArgs,
	}

	if expr.isSync {
		return callLuaMsgSync(callArgs)
	} else {
		return callLuaMsg(callArgs)
	}
}

func callLuaEventHooks(cmdName string, getArgs func(L *lua.LState) []lua.LValue) error {
	exprList, ok := gLuaRegistry.eventHooks[cmdName]
	if !ok {
		return nil
	}

	errCnt := 0
	for _, expr := range exprList {
		_, err := callLuaMsgExpr(&expr, nil, getArgs)
		if err != nil {
			errCnt++
			log.Printf("failed to run hook %s: %s", &expr, err)
		}
	}

	if errCnt > 0 {
		return fmt.Errorf("%d error(s) occured during event hook call, see log for more detail", errCnt)
	}

	return nil
}

func callLuaPreviewerConditionChecker(expr *luaMsgExpr, getArgs luaMsgArgsMaker) (bool, error) {
	ret, err := callLuaMsgExpr(
		expr,
		func(L *lua.LState, msgEntry lua.LValue) *lua.LFunction {
			switch msgEntry.Type() {
			case lua.LTFunction:
				// previewer defined as a simple function, activated unconditionally.
				return L.NewFunction(lgFuncUnconditional)
			case lua.LTTable:
				// previewer defined as a table, try find condition function from it.
				condFunc := msgEntry.(*lua.LTable).RawGetString(luaPreviewerConditionFuncKey)
				if condFunc.Type() == lua.LTFunction {
					return condFunc.(*lua.LFunction)
				} else {
					return L.NewFunction(lgFuncUnconditional)
				}
			}
			// invalid previewer
			return nil
		},
		getArgs,
	)

	if err != nil {
		return false, fmt.Errorf("previewer condition check failed: %s", err)
	}

	if len(ret) == 0 {
		return false, nil
	}

	return !lua.LVIsFalse(ret[0]), nil
}

func callLuaPreviewerAction(expr *luaMsgExpr, getArgs luaMsgArgsMaker) (bool, error) {
	ret, err := callLuaMsgExpr(expr, nil, getArgs)
	if err != nil {
		return true, fmt.Errorf("previewer action failed: %s", err)
	}

	if len(ret) == 0 {
		return false, nil
	}

	return !lua.LVIsFalse(ret[0]), nil
}

// ----------------------------------------------------------------------------

func getLuaPreviewerForPath(path string) *luaPreviewerInfo {
	getArgs := func(L *lua.LState) []lua.LValue {
		return []lua.LValue{lua.LString(path)}
	}

	var result *luaPreviewerInfo
	for i := range gLuaRegistry.previewers {
		previewer := &gLuaRegistry.previewers[i]
		ok, _ := callLuaPreviewerConditionChecker(&previewer.msgexpr, getArgs)
		if ok {
			result = previewer
			break
		}
	}

	return result
}
