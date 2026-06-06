package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
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

type luaMsgTarget struct {
	sourceName  string
	registryKey string
	msg         string
	isSync      bool
}

var gLuaRegistry struct {
	sortMethod map[string]luaMsgExpr
	eventHooks map[string][]luaMsgExpr

	stateDataMap map[*lua.LState]map[string]*lua.LTable
}

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

func loadEventHookRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyEventHook

	value := tbl.RawGetString(registryKey)
	switch value.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", sourceName, registryKey)
		return
	}

	eventHookTbl := value.(*lua.LTable)
	eventHookTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("event hook registry key is expected to be string, found %s: %s", sourceName, key.Type(), key)
			return
		}

		if value.Type() != lua.LTFunction {
			log.Printf("event hook registry value is expected to be function, found %s: %s", sourceName, value.Type(), value)
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
		log.Printf("registry field `%s` is not a table", sourceName, registryKey)
		return
	}

	sortMethodTbl := value.(*lua.LTable)
	sortMethodTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("sort method registry key is expected to be string, found %s: %s", sourceName, key.Type(), key)
			return
		}

		if value.Type() != lua.LTFunction {
			log.Printf("sort method registry value is expected to be function, found %s: %s", sourceName, value.Type(), value)
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
		log.Printf("registry field `%s` is not a table", sourceName, registryKey)
		return
	}

	cmdTbl := value.(*lua.LTable)
	cmdTbl.ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("command registry key is expected to be string, found %s: %s", sourceName, key.Type(), key)
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
			log.Printf("command registry value is expected to be string or function, get %s: %s", sourceName, value.Type(), value)
		}
	})
}

func callLuaCmdOnState(L *lua.LState, sourceName string, registryKey string, msg string, getargs func(L *lua.LState) []lua.LValue) ([]lua.LValue, error) {
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
	if handler.Type() != lua.LTFunction {
		return nil, fmt.Errorf("msg handler is not a function")
	}

	var luaArgs []lua.LValue
	if getargs != nil {
		luaArgs = getargs(L)
	}

	L.Push(handler.(*lua.LFunction))
	for _, arg := range luaArgs {
		L.Push(arg)
	}

	log.Printf("call Lua msg: (%s, %s, %s): %q", sourceName, registryKey, msg, luaArgs)
	err := L.PCall(len(luaArgs), lua.MultRet, nil)
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

func callLuaMsg(sourceName string, registryKey string, msg string, getArgs func(L *lua.LState) []lua.LValue) ([]lua.LValue, error) {
	L := gLuaPool.get()
	defer gLuaPool.put(L)
	return callLuaCmdOnState(L, sourceName, registryKey, msg, getArgs)
}

func callLuaMsgSync(sourceName string, registryKey string, msg string, getArgs func(L *lua.LState) []lua.LValue) ([]lua.LValue, error) {
	L := gLuaStateSync.acquire()
	defer gLuaStateSync.release()
	return callLuaCmdOnState(L, sourceName, registryKey, msg, getArgs)
}

func callLuaMsgExpr(expr *luaMsgExpr, getArgs func(L *lua.LState) []lua.LValue) ([]lua.LValue, error) {
	if expr.isSync {
		return callLuaMsgSync(expr.sourceName, expr.registry, expr.msg, getArgs)
	} else {
		return callLuaMsg(expr.sourceName, expr.registry, expr.msg, getArgs)
	}
}

func callLuaEventHooks(cmdName string, getArgs func(L *lua.LState) []lua.LValue) error {
	exprList, ok := gLuaRegistry.eventHooks[cmdName]
	if !ok {
		return nil
	}

	errCnt := 0
	for _, expr := range exprList {
		_, err := callLuaMsgExpr(&expr, getArgs)
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
