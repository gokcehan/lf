package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

const luaPluginDirName = "plugins"

const luaGlobalNameApp = "app"
const luaGlobalNameDataStore = "data_store"

const luaMsgVariantMain = ""
const luaMsgMetaKeyIsAsync = "is_async"

const (
	registryKeyCommand       = "command"
	registryKeyEventHook     = "event_hook"
	registryKeyKeyMap        = "key_map"
	registryKeyLocalOption   = "local_option"
	registryKeyMisc          = "misc"
	registryKeyOption        = "option"
	registryKeyPreviewer     = "previewer"
	registryKeySortingMethod = "sorting_method"
	registryKeyUIFormatter   = "ui_formatter"
	registryKeyUIPrinter     = "ui_printer"
	registryKeyUIStyle       = "ui_style"
)

const (
	luaMiscMsgShell   = "shell"
	luaMiscMsgDupFile = "dupfile"

	luaUIFormatterCursorActive  = "cursoractive"
	luaUIFormatterCursorParent  = "cursorparent"
	luaUIFormatterCursorPreview = "cursorpreview"
	luaUIFormatterError         = "error"
	luaUIFormatterNumberCursor  = "numbercursor"
	luaUIFormatterNumber        = "number"
	luaUIFormatterTag           = "tag"

	luaUIPrinterDirEntry  = "dir_entry"
	luaUIPrinterDirectory = "directory"
	luaUIPrinterRuler     = "ruler"
	luaUIPrinterPrompt    = "prompt"

	luaUIStyleBorder     = "border"
	luaUIStyleCopy       = "copy"
	luaUIStyleCut        = "cut"
	luaUIStyleMenu       = "menu"
	luaUIStyleMenuheader = "menuheader"
	luaUIStyleMenuselect = "menuselect"
	luaUIStyleSelect     = "select"
	luaUIStyleVisual     = "visual"
)

const (
	luaCommandActionFuncKey     = "action"
	luaCommandCompletionFuncKey = "completion"

	luaEventHookActionFuncKey = "action"

	luaMiscActionFuncKey = "action"

	luaKeyMapActionFuncKey = "action"

	luaPreviewerActionFuncKey    = "action"
	luaPreviewerCleanFuncKey     = "clean"
	luaPreviewerConditionFuncKey = "condition"

	luaSortingMethodActionFuncKey = "action"

	luaUIFormatterActionFuncKey = "action"

	luaUIPrinterActionFuncKey = "action"
)

const (
	luaKeyMapTypeNormal  = "n"
	luaKeyMapTypeVisual  = "v"
	luaKeyMapTypeCommand = "c"
)

type luaDataStore struct {
	lock      sync.RWMutex
	dataStore map[string]any
}

func (store *luaDataStore) set(key string, value lua.LValue) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	if store.dataStore == nil {
		store.dataStore = make(map[string]any)
	}

	if value == lua.LNil {
		delete(store.dataStore, key)
		return nil
	}

	goValue, err := luaPlainValueToGoValue(value)
	if err != nil {
		return err
	}

	store.dataStore[key] = goValue

	return nil
}

func (store *luaDataStore) get(L *lua.LState, key string) (lua.LValue, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	if store.dataStore == nil {
		return lua.LNil, nil
	}

	goValue := store.dataStore[key]
	value, err := goValueToLuaValue(L, goValue)
	if err != nil {
		return lua.LNil, err
	}

	return value, nil
}

func (store *luaDataStore) clear() {
	store.lock.Lock()
	defer store.lock.Unlock()

	clear(store.dataStore)
}

func (store *luaDataStore) keysAsLuaTbl(L *lua.LState) *lua.LTable {
	store.lock.RLock()
	defer store.lock.RUnlock()

	tbl := L.NewTable()
	for k := range store.dataStore {
		tbl.Append(lua.LString(k))
	}

	return tbl
}

type lStatePool struct {
	lockPool  sync.Mutex
	saved     []*lua.LState
	allStates []*lua.LState

	lockLuaStateSync sync.RWMutex
	luaStateSync     *lua.LState

	isInitialized bool // Lua global registry has been updated by instanciate first Lua state
	isClosed      bool // indicating a shutdown has been called on the pool

	app             *app
	pluginRootDirs  []string             // plugin root directory list
	pluginByteCodes []*lua.FunctionProto // compiled Lua byte code of all plugin

	dataStore *luaDataStore
}

func newLStatePool(app *app) *lStatePool {
	return &lStatePool{
		app:       app,
		dataStore: new(luaDataStore),
	}
}

// ----------------------------------------------------------------------------
// lock-free APIs

// addPluginRoot adds given path to plugin root directory list.
func (pl *lStatePool) addPluginRoot(rootDir string) {
	if !slices.Contains(pl.pluginRootDirs, rootDir) {
		pl.pluginRootDirs = append(pl.pluginRootDirs, rootDir)
	}
}

// loadPluginScripts compiles and sotre plugin entrance Lua script founded under
// each plugin roots.
func (pl *lStatePool) loadPluginScripts() error {
	pluginByteCodes := make([]*lua.FunctionProto, 0)

	errorCnt := 0

	for _, pluginDir := range pl.pluginRootDirs {
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

// newWithRetAction creates a new Lua state and takes a `action` function that
// can do extra stuff with value returned by each plugin script.
func (pl *lStatePool) newWithRetAction(action func(sourceName string, L *lua.LState, tbl *lua.LTable)) (*lua.LState, error) {
	if pl.isClosed {
		return nil, fmt.Errorf("Lua State has been closed on app quit")
	}

	L := lua.NewState()

	if err := setupScripImportPath(L, pl.pluginRootDirs); err != nil {
		log.Printf("failed to setup Lua loader search path: %s", err)
	}
	setupLuaGlobals(pl, L)
	setupPreloadModules(L)

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

			dataTblMap, ok := gLuaRegistry.stateDataMap[L]
			if !ok {
				dataTblMap = make(map[string]*lua.LTable)
				gLuaRegistry.stateDataMap[L] = dataTblMap
			}

			dataTblMap[sourceName] = tbl

			if action != nil {
				action(sourceName, L, tbl)
			}
		} else {
			log.Println("plugin script", proto.SourceName, "did not return a table")
		}
	}

	return L, nil
}

// new creates a new Lua state.
func (pl *lStatePool) newSimple() (*lua.LState, error) {
	return pl.newWithRetAction(nil)
}

// newWithRegistryUpdate creates a new Lua state, and updates global Lua registry
// with value returned by each plugin script during Lua state initialization.
//
// P.S.: this function modify global option and registry, this function should
// be called on main goroutine.
func (pl *lStatePool) newWithRegistryUpdate() (*lua.LState, error) {
	return pl.newWithRetAction(func(sourceName string, L *lua.LState, tbl *lua.LTable) {
		if tbl == nil {
			return
		}

		log.Println("update Lua registry for plugin script:", sourceName)

		loadCommandRegistryFromTbl(sourceName, tbl)
		loadEventHookRegistryFromTbl(sourceName, tbl)
		loadKeyMapRegistryFromTbl(sourceName, tbl)
		loadPreviewerRegistryFromTbl(sourceName, tbl)
		loadMiscRegistryFromTbl(sourceName, tbl)
		loadSortingMethodRegistryFromTbl(sourceName, tbl)
		loadUIFormatterRegistryFromTbl(sourceName, tbl)
		loadUIPrinterRegistryFromTbl(sourceName, tbl)
		loadUIStyleRegistryFromTbl(L, sourceName, tbl)

		sortLuaPreviewers()
	})
}

// ----------------------------------------------------------------------------

// initializeState loads plugin script byte code and create synchronous Lua state
// object.
func (pl *lStatePool) initializeState(app *app) {
	pl.lockPool.Lock()
	defer pl.lockPool.Unlock()

	if pl.isInitialized {
		return
	}

	pl.isInitialized = true

	err := gLuaPool.loadPluginScripts()
	if err != nil {
		app.ui.echoerr(err.Error())
	}

	// initialize sycnhronous Lua state and Lua registry
	if L, err := gLuaPool.newWithRegistryUpdate(); err == nil {
		pl.luaStateSync = L
	} else {
		app.ui.echoerrf("failed to initialize synchronous Lua state: %s", err)
	}
}

// get takes one Lua state from pool.
func (pl *lStatePool) get() (*lua.LState, error) {
	pl.lockPool.Lock()
	defer pl.lockPool.Unlock()

	if !pl.isInitialized {
		return nil, fmt.Errorf("Lua State has not been initialized")
	}

	if pl.isClosed {
		return nil, fmt.Errorf("Lua State has been closed on app quit")
	}

	n := len(pl.saved)
	if n > 0 {
		L := pl.saved[n-1]
		pl.saved = pl.saved[0 : n-1]
		return L, nil
	}

	L, err := pl.newSimple()
	if L != nil {
		pl.allStates = append(pl.allStates, L)
	}

	return L, err
}

// put returns a Lua state to pool.
func (pl *lStatePool) put(L *lua.LState) {
	pl.lockPool.Lock()
	defer pl.lockPool.Unlock()

	if pl.isClosed {
		L.Close()
	}
	pl.saved = append(pl.saved, L)
}

// acquireSyncState tries acquire synchronous Lua state's mutex, and returns
// synchronous Lua state object after successfully acquired lock.
func (pl *lStatePool) acquireSyncState() (*lua.LState, error) {
	pl.lockPool.Lock()

	if !pl.isInitialized {
		pl.lockPool.Unlock()
		return nil, fmt.Errorf("Lua State has not been initialized")
	}

	if pl.isClosed {
		pl.lockPool.Unlock()
		return nil, fmt.Errorf("Lua State has been closed on app quit")
	}

	pl.lockPool.Unlock()

	pl.lockLuaStateSync.Lock()

	if pl.luaStateSync == nil {
		return nil, fmt.Errorf("No Lua State is available")
	}

	return pl.luaStateSync, nil
}

// releaseSyncState releases synchronous Lua state's mutex.
func (pl *lStatePool) releaseSyncState() {
	pl.lockLuaStateSync.Unlock()
}

// shutdown closes all Lua states in pool.
func (pl *lStatePool) shutdown() {
	pl.lockPool.Lock()
	defer pl.lockPool.Unlock()

	pl.lockLuaStateSync.Lock()
	defer pl.lockLuaStateSync.Unlock()

	pl.isClosed = true

	if pl.luaStateSync != nil {
		pl.luaStateSync.Close()
	}

	for _, L := range pl.saved {
		L.Close()
	}
}

// resetLuaState closes and removes all Lua state in pool object, and reset Lua
// related data to uninitialized state.
func (pl *lStatePool) resetLuaState() error {
	pl.lockLuaStateSync.Lock()
	defer pl.lockLuaStateSync.Unlock()

	for range 10 {
		pl.lockPool.Lock()

		if len(pl.saved) != len(pl.allStates) {
			// some Lua states are occupied
			pl.lockPool.Unlock()
			<-time.After(10 * time.Millisecond)
			continue
		}

		if pl.luaStateSync != nil {
			pl.luaStateSync.Close()
		}

		for _, L := range pl.allStates {
			L.Close()
		}

		pl.saved = nil
		pl.allStates = nil
		pl.luaStateSync = nil

		pl.isInitialized = false

		pl.pluginByteCodes = nil

		pl.lockPool.Unlock()

		return nil
	}

	return fmt.Errorf("Lua State is busy")
}

// checkIsSyncState checks if given state is synchronous Lua state.
func (pl *lStatePool) checkIsSyncState(L *lua.LState) bool {
	return pl.luaStateSync == L
}

type luaPreviewerInfo struct {
	priority   int    // priority value for this previewer
	name       string // name of this previewer, takes the form `<plugin-source>.<previewer-key>`
	hasCleaner bool   // if a cleaner function is defined for this previewer
	msgexpr    luaMsgExpr
}

// luaFuncWriter implments io.Writer interface with a Lua function.
type luaFuncWriter struct {
	luaState *lua.LState
	fn       *lua.LFunction
}

func (writer *luaFuncWriter) Write(p []byte) (n int, err error) {
	luaErr := writer.luaState.CallByParam(
		lua.P{
			Fn:      writer.fn,
			NRet:    2,
			Protect: true,
		},
		lua.LString(string(p)),
	)
	if luaErr != nil {
		return 0, luaErr
	}

	defer writer.luaState.Pop(2)

	ret1 := writer.luaState.Get(-2)
	if ret1.Type() != lua.LTNumber {
		return 0, fmt.Errorf("return value #1 of Lua write function is not a number")
	}
	n = int(ret1.(lua.LNumber))

	var errStr string
	ret2 := writer.luaState.Get(-1)
	if lua.LVAsBool(ret2) {
		errStr = ret2.String()
	}

	if errStr != "" {
		err = fmt.Errorf("Lua writer function error: %s", err)
	}

	return
}

// ----------------------------------------------------------------------------
// Lua registry value operation

// Global LState pool, used for asynchronous execution
var gLuaPool *lStatePool

var gLuaRegistry struct {
	stateDataMap map[*lua.LState]map[string]*lua.LTable

	eventHooks    map[string][]*luaMsgExpr
	misc          map[string]*luaMsgExpr
	previewers    []luaPreviewerInfo
	sortingMethod map[string]*luaMsgExpr
	uiFormatter   map[string]*luaMsgExpr
	uiPrinter     map[string]*luaMsgExpr
	uiStyleMap    map[string]tcell.Style
}

func goReflectValueToLuaValue(L *lua.LState, rValue reflect.Value) (luaValue lua.LValue, err error) {
	switch rValue.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		luaValue = lua.LNumber(rValue.Convert(reflect.TypeOf(float64(0))).Float())
	case reflect.String:
		luaValue = lua.LString(rValue.String())
	case reflect.Bool:
		if rValue.Bool() {
			luaValue = lua.LTrue
		} else {
			luaValue = lua.LFalse
		}
	case reflect.Slice, reflect.Array:
		var elemValue lua.LValue
		tbl := L.NewTable()

		for i := 0; i < rValue.Len(); i++ {
			elem := rValue.Index(i)
			if elemValue, err = goReflectValueToLuaValue(L, elem); err == nil {
				tbl.Append(elemValue)
			} else {
				err = fmt.Errorf("failed to convert slice/array element: %s", err)
				break
			}
		}

		luaValue = tbl
	case reflect.Map:
		var lMapKey, lMapValue lua.LValue
		tbl := L.NewTable()
		keys := rValue.MapKeys()

		for _, mapKey := range keys {
			mapValue := rValue.MapIndex(mapKey)

			lMapKey, err = goReflectValueToLuaValue(L, mapKey)
			if err != nil {
				err = fmt.Errorf("failed to convert map key: %s", err)
			}

			lMapValue, err = goReflectValueToLuaValue(L, mapValue)
			if err != nil {
				err = fmt.Errorf("failed to convert map value: %s", err)
			}

			tbl.RawSet(lMapKey, lMapValue)
		}

		luaValue = tbl
	case reflect.Ptr:
		if rValue.IsNil() {
			luaValue = lua.LNil
		} else {
			ud := L.NewUserData()
			ud.Value = rValue.Pointer()
		}
	case reflect.Interface:
		if rValue.IsNil() {
			luaValue = lua.LNil
		} else if rValue.CanInterface() {
			ud := L.NewUserData()
			ud.Value = rValue.Interface()
		} else {
			err = fmt.Errorf("unaccessable interface value")
		}
	case reflect.Struct:
		if rValue.CanAddr() {
			ud := L.NewUserData()
			ud.Value = rValue.Addr().Pointer()
		} else if rValue.CanInterface() {
			ud := L.NewUserData()
			ud.Value = rValue.Interface()
		} else {
			err = fmt.Errorf("unaddressable struct value")
		}
	default:
		err = fmt.Errorf("unsupported value type: %s", rValue.Kind())
		luaValue = lua.LNil
	}

	return
}

// goValueToLuaValue converts Go value to lua.LValue.
func goValueToLuaValue(L *lua.LState, value any) (lua.LValue, error) {
	var err error

	lValue, ok := value.(lua.LValue)
	if ok {
		return lValue, err
	}

	return goReflectValueToLuaValue(L, reflect.ValueOf(value))
}

// luaPlainValueToGoValue converts simple Lua value to Go value.
func luaPlainValueToGoValue(value lua.LValue) (any, error) {
	switch value.Type() {
	case lua.LTNil:
		return nil, nil
	case lua.LTBool:
		if value == lua.LTrue {
			return true, nil
		} else {
			return false, nil
		}
	case lua.LTNumber:
		return float64(value.(lua.LNumber)), nil
	case lua.LTString:
		return string(value.(lua.LString)), nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", value.Type())
	}
}

// getAppObjectFromLuaGlobals fetchs app object from Lua state's global variable.
func getAppObjectFromLuaGlobals(L *lua.LState) (*app, error) {
	value := L.GetGlobal(luaGlobalNameApp)

	ud, ok := value.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("global variable `%s` is not a user data", luaGlobalNameApp)
	}

	app, ok := ud.Value.(*app)
	if !ok {
		return nil, fmt.Errorf("global variable `%s` is not `*app` value", ud.Value)
	}

	return app, nil
}

// getPluginNameForSourcePath return plugin name for given plugin script path
func getPluginNameForSourcePath(sourceName string) string {
	return filepath.Base(filepath.Dir(sourceName))
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
func setupLuaGlobals(pl *lStatePool, L *lua.LState) {
	// Lua meta table registering must happens before any user data gets pushed
	// onto Lua state.
	setupLuaTypeBindings(L)

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

	L.SetGlobal(luaGlobalNameApp, lWrapApp(L, pl.app))
	L.SetGlobal(luaGlobalNameDataStore, lWrapLuaDataStore(L, pl.dataStore))
}

// setupLuaTypeBindings adds `lf_types` global table as entrance of accessing
// Go type binding meta tables.
func setupLuaTypeBindings(L *lua.LState) {
	lfTypes := L.NewTable()

	// bufio
	lfTypes.RawSetString("BufWriter", lRegisterBufWriterType(L))
	lfTypes.RawSetString("BufReader", lRegisterBufReaderType(L))
	// exec
	lfTypes.RawSetString("Cmd", lRegisterCmdType(L))
	// fs
	lfTypes.RawSetString("DirEntry", lRegisterDirEntryType(L))
	lfTypes.RawSetString("FileInfo", lRegisterFileInfoType(L))
	lfTypes.RawSetString("FileMode", lRegisterFileModeType(L))
	// main
	lfTypes.RawSetString("App", lRegisterAppType(L))
	lfTypes.RawSetString("CompMatch", lRegisterCompMatchType(L))
	lfTypes.RawSetString("Clipboard", lRegisterClipboardType(L))
	lfTypes.RawSetString("Dir", lRegisterDirType(L))
	lfTypes.RawSetString("DirContext", lRegisterDirContextType(L))
	lfTypes.RawSetString("DirStyle", lRegisterDirStyleType(L))
	lfTypes.RawSetString("File", lRegisterFileTypeMt(L))
	lfTypes.RawSetString("FuncWriter", lRegisterFuncWriterType(L))
	lfTypes.RawSetString("IconDef", lRegisterIconDefType(L))
	lfTypes.RawSetString("IconMap", lRegisterIconMapType(L))
	lfTypes.RawSetString("LuaDataStore", lRegisterLuaDataStoreType(L))
	lfTypes.RawSetString("LuaMsgExpr", lRegisterLuaMsgExprType(L))
	lfTypes.RawSetString("Nav", lRegisterNavType(L))
	lfTypes.RawSetString("PrintDirEntryContext", lRegisterPrintDirEntryContextType(L))
	lfTypes.RawSetString("StyleMap", lRegisterStyleMapType(L))
	lfTypes.RawSetString("Win", lRegisterWinType(L))
	lfTypes.RawSetString("UI", lRegisterUIType(L))
	// tcell
	lfTypes.RawSetString("TcellColor", lRegisterTcellColorType(L))
	lfTypes.RawSetString("TcellScreen", lRegisterTcellScreenType(L))
	lfTypes.RawSetString("TcellStyle", lRegisterTcellStyleType(L))
	// time
	lfTypes.RawSetString("Duration", lRegisterDurationType(L))
	lfTypes.RawSetString("Month", lRegisterMonthType(L))
	lfTypes.RawSetString("Time", lRegisterTimeType(L))
	lfTypes.RawSetString("Timer", lRegisterTimerType(L))
	lfTypes.RawSetString("Weekday", lRegisterWeekdayType(L))

	L.SetGlobal("lf_types", lfTypes)
}

// setupPreloadModules register load functions for preload modules.
func setupPreloadModules(L *lua.LState) {
	L.PreloadModule("lf", lfMainModuleLoader)
	L.PreloadModule("lf.fs", lfFsModuleLoader)
	L.PreloadModule("lf.utf8", lfUtf8ModuleLoader)
	L.PreloadModule("lf.ui", lfUIModuleLoader)
}

// tryRaiseNonSyncLuaStateError raises an error if `L` is not synchronous Lua state.
// This function is used to enforce a Lua API to be called on synchronous Lua state.
func tryRaiseNonSyncLuaStateError(L *lua.LState) {
	if !gLuaPool.checkIsSyncState(L) {
		app, _ := getAppObjectFromLuaGlobals(L)
		if app != nil {
			app.ui.exprChan <- &callExpr{"echoerr", []string{"synchronous Lua function is called under async mode"}, 1}
		}
		L.RaiseError("this func should be called with synchronous mode")
	}
}

// tryRaiseSyncLuaStateError raises an error if `L` is synchronous Lua state.
// This function is used to enforce a Lua API to be called with asynchronous mode,
// such as a Lua API that calls other Lua mesages in it.
func tryRaiseSyncLuaStateError(L *lua.LState) {
	if gLuaPool.checkIsSyncState(L) {
		app, _ := getAppObjectFromLuaGlobals(L)
		if app != nil {
			app.ui.exprChan <- &callExpr{"echoerr", []string{"async Lua API is called under synchronous mode"}, 1}
		}
		L.RaiseError("this func should be called with asynchronous mode")
	}
}

// loadLuaPluginOptionValue loads option and local option defined in Lua script.
// This process may trigger other Lua message call, it has to be seperated from
// other registry load.
func loadLuaPluginOptionValue(app *app) {
	L, err := gLuaPool.get()
	defer gLuaPool.put(L)

	if err != nil {
		app.ui.echoerrf("failed to initialize async Lua state")
		return
	}

	registryMap, ok := gLuaRegistry.stateDataMap[L]
	if ok {
		for _, tbl := range registryMap {
			loadLocalOptionRegistryFromTbl(L, app, tbl)
			loadOptionRegistryFromTbl(L, app, tbl)
		}
	}
}

// initializeLua load plugin scripts, and initialize Lua state.
//
// P.S.: this function modify global option and registry, this function should
// be called on main goroutine.
func initializeLua(app *app) {
	gLuaPool = newLStatePool(app)

	if gPluginDir != "" {
		gLuaPool.addPluginRoot(gPluginDir)
	} else if gConfigPath != "" {
		pluginRoot := filepath.Join(filepath.Dir(gConfigPath), luaPluginDirName)
		gLuaPool.addPluginRoot(pluginRoot)
	} else {
		for _, path := range gConfigPaths {
			pluginRoot := filepath.Join(filepath.Dir(path), luaPluginDirName)
			gLuaPool.addPluginRoot(pluginRoot)
		}
	}

	gLuaPool.initializeState(app)
	loadLuaPluginOptionValue(app)
}

// luaPluginReload reset Lua state, Lua registry, then reload Lua script again.
//
// P.S.: this function modify global option and registry, this function should
// be called on main goroutine.
func luaPluginReload(app *app) {
	err := gLuaPool.resetLuaState()
	if err != nil {
		app.ui.echoerrf("Lua plugin reload failed: %s", err)
		return
	}

	gLuaRegistry.stateDataMap = make(map[*lua.LState]map[string]*lua.LTable)

	gLuaRegistry.sortingMethod = make(map[string]*luaMsgExpr)
	gLuaRegistry.eventHooks = make(map[string][]*luaMsgExpr)
	gLuaRegistry.previewers = nil
	gLuaRegistry.misc = make(map[string]*luaMsgExpr)
	gLuaRegistry.uiFormatter = make(map[string]*luaMsgExpr)
	gLuaRegistry.uiPrinter = make(map[string]*luaMsgExpr)
	gLuaRegistry.uiStyleMap = make(map[string]tcell.Style)

	gLuaPool.initializeState(app)
	loadLuaPluginOptionValue(app)

	app.ui.echo("Lua plugins reloaded")
}

// ----------------------------------------------------------------------------

// loadCommandRegistryFromTbl registers commands defined in table returned from
// plugin script.
func loadCommandRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyCommand

	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	cnt := 0
	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("command registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		msg := key.String()

		switch value.Type() {
		case lua.LTString:
			text := value.String()
			p := newParser(strings.NewReader(text))
			expr := p.parseExpr()
			if expr == nil {
				log.Printf("failed to parse Lua command: %s", text)
				return
			} else {
				gOpts.cmds[msg] = expr
			}
		case lua.LTFunction:
			gOpts.cmds[msg] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        msg,
				variant:    luaMsgVariantMain,
			}
		case lua.LTTable:
			actionValue := value.(*lua.LTable).RawGetString(luaCommandActionFuncKey)
			if actionValue.Type() == lua.LTFunction {
				gOpts.cmds[msg] = &luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        msg,
					variant:    luaMsgVariantMain,
					isAsync:    lua.LVAsBool(value.(*lua.LTable).RawGetString(luaMsgMetaKeyIsAsync)),
				}
			} else {
				log.Printf("invalid command action value: %s", value)
				return
			}
		default:
			log.Printf("invalid command registry value of type %s: %s", value.Type(), value)
			return
		}

		cnt++
	})

	if cnt > 0 {
		log.Printf("%d command(s) added", cnt)
	}
}

// loadEventHookRegistryFromTbl registers event hooks defined in table returned
// from plugin script.
func loadEventHookRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyEventHook
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	if gLuaRegistry.eventHooks == nil {
		gLuaRegistry.eventHooks = make(map[string][]*luaMsgExpr)
	}

	cnt := 0
	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("event hook registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		msg := key.String()

		switch value.Type() {
		case lua.LTFunction:
			gLuaRegistry.eventHooks[msg] = append(
				gLuaRegistry.eventHooks[msg],
				&luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        msg,
					variant:    luaMsgVariantMain,
					isAsync:    false,
				},
			)
		case lua.LTTable:
			gLuaRegistry.eventHooks[msg] = append(
				gLuaRegistry.eventHooks[msg],
				&luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        msg,
					variant:    luaMsgVariantMain,
					isAsync:    lua.LVAsBool(value.(*lua.LTable).RawGetString(luaMsgMetaKeyIsAsync)),
				},
			)
		default:
			log.Printf("unsupported event hook registry value type for key: %s", msg)
			return
		}

		cnt++
	})

	if cnt > 0 {
		log.Printf("%d event hook(s) added", cnt)
	}
}

// loadKeyMapRegistryFromTbl registers key maps defined in table returned from plugin
// script.
func loadKeyMapRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyKeyMap

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

	addKeyMapForRegistryValue(registryTbl, sourceName, luaKeyMapTypeNormal, "normal", gOpts.nkeys)
	addKeyMapForRegistryValue(registryTbl, sourceName, luaKeyMapTypeVisual, "visual", gOpts.vkeys)
	addKeyMapForRegistryValue(registryTbl, sourceName, luaKeyMapTypeCommand, "command", gOpts.cmdkeys)
}

// addKeyMapForRegistryValue loads one type of key map from registry table.
func addKeyMapForRegistryValue(registryTbl *lua.LTable, sourceName, keyMapType, displayName string, keys map[string]expr) {
	tbl := registryTbl.RawGetString(keyMapType)
	switch tbl.Type() {
	case lua.LTTable:
		// ok
	case lua.LTNil:
		return
	default:
		log.Printf("key map group %s is not a table: %s", keyMapType, tbl)
		return
	}

	keyMapCnt := 0

	tbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("map registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		mapKey := key.String()
		isAsync := false

		mapAction := value
		if value.Type() == lua.LTTable {
			tbl := value.(*lua.LTable)

			mapAction = tbl.RawGetString(luaKeyMapActionFuncKey)
			isAsync = lua.LVAsBool(tbl.RawGetString(luaMsgMetaKeyIsAsync))
		}

		switch mapAction.Type() {
		case lua.LTString:
			text := mapAction.String()
			if text == "" {
				delete(keys, mapKey)
			} else {
				p := newParser(strings.NewReader(text))
				expr := p.parseExpr()
				if expr == nil {
					log.Printf("failed to parse Lua key map %s.%s: %s", keyMapType, mapKey, p.err)
				} else {
					keys[mapKey] = expr
				}
			}
		case lua.LTFunction:
			keys[mapKey] = &luaKeyMapExpr{
				sourceName: sourceName,
				keyMapType: keyMapType,
				key:        mapKey,
				count:      1,
				isAsync:    isAsync,
			}
		default:
			log.Printf("unsupported key map registry value for %s.%s", keyMapType, mapKey)
			return
		}

		keyMapCnt++
	})

	if keyMapCnt > 0 {
		log.Printf("%d %s key map(s) added", keyMapCnt, displayName)
	}
}

// loadLocalPreviewerRegistryFromTbl loads option value from table returned from plugin script.
func loadLocalOptionRegistryFromTbl(L *lua.LState, app *app, tbl *lua.LTable) {
	registryKey := registryKeyLocalOption
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	registryTbl.(*lua.LTable).ForEach(func(pathKey, optionTbl lua.LValue) {
		if pathKey.Type() != lua.LTString {
			log.Printf("local option registry key is expected to be string, found %s: %s", pathKey.Type(), pathKey)
			return
		}

		path := string(pathKey.(lua.LString))

		if optionTbl.Type() != lua.LTTable {
			app.ui.echoerrf("registry value for local option group `%s` is not a table", path)
			return
		}

		optionTbl.(*lua.LTable).ForEach(func(optionKey, optionValue lua.LValue) {
			if optionKey.Type() != lua.LTString {
				app.ui.echoerrf("local option group option key is expected to be string, found %s: %s", optionKey.Type(), optionKey)
				return
			}

			option := string(optionKey.(lua.LString))

			switch optionValue.Type() {
			case lua.LTString:
				expr := &setLocalExpr{path: path, opt: option, val: string(optionValue.(lua.LString))}
				expr.eval(app, nil)
			case lua.LTFunction:
				err := L.CallByParam(lua.P{
					Fn:      optionValue.(*lua.LFunction),
					NRet:    1,
					Protect: true,
				})
				if err != nil {
					app.ui.echoerrf("failed to evaluate local option `%s`, see log for more detail", option)
					log.Printf("failed to run function for local option `%s` `%s`: %s", path, option, err)
				}

				defer L.Pop(1)
				ret := L.Get(-1)
				if ret.Type() == lua.LTString {
					expr := &setLocalExpr{path: path, opt: option, val: string(ret.(lua.LString))}
					expr.eval(app, nil)
				} else {
					app.ui.echoerrf("Lua function for local option `%s` `%s` does not return string value", path, option)
				}
			default:
				log.Printf("unsupported local option registry value type for %s %s", path, option)
				return
			}
		})
	})
}

// loadMiscRegistryFromTbl loads shell relative registry entry.
func loadMiscRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyMisc
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	if gLuaRegistry.misc == nil {
		gLuaRegistry.misc = make(map[string]*luaMsgExpr)
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("misc registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		msg := key.String()
		switch msg {
		case luaMiscMsgDupFile,
			luaMiscMsgShell:
			// ok
		default:
			log.Println("unsupported misc registry entry key:", msg)
			return
		}

		switch value.Type() {
		case lua.LTFunction:
			gLuaRegistry.misc[msg] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        msg,
				variant:    luaMsgVariantMain,
				isAsync:    false,
			}
		case lua.LTTable:
			gLuaRegistry.misc[msg] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        msg,
				variant:    luaMsgVariantMain,
				isAsync:    lua.LVAsBool(value.(*lua.LTable).RawGetString(luaMsgMetaKeyIsAsync)),
			}
		default:
			log.Printf("unsupported misc registry value for key: %s", msg)
			return
		}
	})
}

// loadPreviewerRegistryFromTbl loads option value from table returned from plugin script.
func loadOptionRegistryFromTbl(L *lua.LState, app *app, tbl *lua.LTable) {
	registryKey := registryKeyOption
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("option registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		option := key.String()

		switch value.Type() {
		case lua.LTString:
			expr := &setExpr{opt: option, val: string(value.(lua.LString))}
			expr.eval(app, nil)
		case lua.LTFunction:
			err := L.CallByParam(lua.P{
				Fn:      value.(*lua.LFunction),
				NRet:    1,
				Protect: true,
			})
			if err != nil {
				app.ui.echoerrf("failed to evaluate option `%s`, see log for more detail", option)
				log.Printf("failed to run function for option `%s`: %s", option, err)
			}

			defer L.Pop(1)
			ret := L.Get(-1)
			if ret.Type() == lua.LTString {
				expr := &setExpr{opt: option, val: string(ret.(lua.LString))}
				expr.eval(app, nil)
			} else {
				app.ui.echoerrf("Lua function for option `%s` does not return string value", option)
			}
		default:
			log.Printf("unsupported option registry value type for key: %s", option)
			return
		}
	})
}

// loadPreviewerRegistryFromTbl registers previewers defined in table returned
// rom plugin script.
func loadPreviewerRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyPreviewer
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("previewer registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		msg := key.String()
		name := getPluginNameForSourcePath(sourceName) + "." + msg

		switch value.Type() {
		case lua.LTFunction:
			gLuaRegistry.previewers = append(gLuaRegistry.previewers, luaPreviewerInfo{
				priority: 0,
				name:     name,
				msgexpr: luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        msg,
					variant:    luaMsgVariantMain,
					isAsync:    false,
				},
			})
		case lua.LTTable:
			previewerTbl := value.(*lua.LTable)
			hasCleaner := previewerTbl.RawGetString(luaPreviewerCleanFuncKey).Type() == lua.LTFunction

			gLuaRegistry.previewers = append(gLuaRegistry.previewers, luaPreviewerInfo{
				priority:   0,
				name:       name,
				hasCleaner: hasCleaner,
				msgexpr: luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        msg,
					variant:    luaMsgVariantMain,
					isAsync:    lua.LVAsBool(previewerTbl.RawGetString(luaMsgMetaKeyIsAsync)),
				},
			})
		default:
			log.Printf("unsupported previewer registry value type for key: %s", msg)
			return
		}

		log.Printf("add previewer: %s", name)
	})
}

// sortLuaPreviewers sorts Lua previewers by their priority and name. Previewers
// with higher priority takes precedence.
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

// setLuaPreviewerPriority updates priority value for previewer with given name.
// When `withSort` is true, this function will sort previewer list when previewer
// priority is actually changed.
// If no previewer with given name is found, this function does nothing.
// This function returns true if previewer priority is actually changed, otherwise
// false.
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

// loadSortingMethodRegistryFromTbl registers sort methods defined in table returned
// from plugin script.
func loadSortingMethodRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeySortingMethod
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	if gLuaRegistry.sortingMethod == nil {
		gLuaRegistry.sortingMethod = make(map[string]*luaMsgExpr)
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("sort method registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		msg := key.String()
		name := getPluginNameForSourcePath(sourceName) + "." + msg

		switch value.Type() {
		case lua.LTFunction:
			gLuaRegistry.sortingMethod[name] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        msg,
				variant:    luaMsgVariantMain,
				isAsync:    false,
			}
		case lua.LTTable:
			gLuaRegistry.sortingMethod[name] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        msg,
				variant:    luaMsgVariantMain,
				isAsync:    lua.LVAsBool(value.(*lua.LTable).RawGetString(luaMsgMetaKeyIsAsync)),
			}
		default:
			log.Printf("unsupported sort method registry value for key: %s", msg)
			return
		}

		log.Printf("add sort method: %s", name)
	})
}

// loadUIFormatterRegistryFromTbl registers UI formatters defined in table returned
// from plugin script.
func loadUIFormatterRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyUIFormatter
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	if gLuaRegistry.uiFormatter == nil {
		gLuaRegistry.uiFormatter = make(map[string]*luaMsgExpr)
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("UI formatter key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		option := key.String()
		switch option {
		case luaUIFormatterCursorActive,
			luaUIFormatterCursorParent,
			luaUIFormatterCursorPreview,
			luaUIFormatterError,
			luaUIFormatterNumberCursor,
			luaUIFormatterNumber,
			luaUIFormatterTag:
			// ok
		default:
			log.Println("unsupported UI formatter registry key:", option)
			return
		}

		switch value.Type() {
		case lua.LTFunction:
			gLuaRegistry.uiFormatter[option] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        option,
				variant:    luaMsgVariantMain,
			}
		case lua.LTTable:
			actionValue := value.(*lua.LTable).RawGetString(luaCommandActionFuncKey)
			if actionValue.Type() == lua.LTFunction {
				gLuaRegistry.uiFormatter[option] = &luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        option,
					variant:    luaMsgVariantMain,
					isAsync:    lua.LVAsBool(value.(*lua.LTable).RawGetString(luaMsgMetaKeyIsAsync)),
				}
			} else {
				log.Println("invalid UI formatter value:", value)
			}
		default:
			log.Printf("invalid UI formatter registry value of type %s: %s", value.Type(), value)
		}
	})
}

// loadUIPrinterRegistryFromTbl registers UI printer defined in table returned
// from plugin script.
func loadUIPrinterRegistryFromTbl(sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyUIPrinter
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	if gLuaRegistry.uiPrinter == nil {
		gLuaRegistry.uiPrinter = make(map[string]*luaMsgExpr)
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("UI pinter key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		option := key.String()
		switch option {
		case luaUIPrinterDirEntry,
			luaUIPrinterDirectory,
			luaUIPrinterRuler,
			luaUIPrinterPrompt:
			// ok
		default:
			log.Println("unsupported UI pinter registry key:", option)
			return
		}

		switch value.Type() {
		case lua.LTFunction:
			gLuaRegistry.uiPrinter[option] = &luaMsgExpr{
				sourceName: sourceName,
				registry:   registryKey,
				msg:        option,
				variant:    luaMsgVariantMain,
			}
		case lua.LTTable:
			actionValue := value.(*lua.LTable).RawGetString(luaCommandActionFuncKey)
			if actionValue.Type() == lua.LTFunction {
				gLuaRegistry.uiPrinter[option] = &luaMsgExpr{
					sourceName: sourceName,
					registry:   registryKey,
					msg:        option,
					variant:    luaMsgVariantMain,
					isAsync:    lua.LVAsBool(value.(*lua.LTable).RawGetString(luaMsgMetaKeyIsAsync)),
				}
			} else {
				log.Println("invalid UI pinter value:", value)
			}
		default:
			log.Printf("invalid UI printer registry value of type %s: %s", value.Type(), value)
		}
	})
}

// loadUIStyleRegistryFromTbl loads UI styles defined in table into Go map.
func loadUIStyleRegistryFromTbl(L *lua.LState, sourceName string, tbl *lua.LTable) {
	registryKey := registryKeyUIStyle
	registryTbl := tbl.RawGetString(registryKey)
	switch registryTbl.Type() {
	case lua.LTNil:
		return
	case lua.LTTable:
		// ok
	default:
		log.Printf("registry field `%s` is not a table", registryKey)
		return
	}

	if gLuaRegistry.uiStyleMap == nil {
		gLuaRegistry.uiStyleMap = make(map[string]tcell.Style)
	}

	registryTbl.(*lua.LTable).ForEach(func(key, value lua.LValue) {
		if key.Type() != lua.LTString {
			log.Printf("ui style registry key is expected to be string, found %s: %s", key.Type(), key)
			return
		}

		settingName := key.String()
		option := value

		switch settingName {
		case luaUIStyleBorder,
			luaUIStyleCopy,
			luaUIStyleCut,
			luaUIStyleMenu,
			luaUIStyleMenuheader,
			luaUIStyleMenuselect,
			luaUIStyleSelect,
			luaUIStyleVisual:
			// ok
		default:
			log.Println("unsupported UI style registry key:", settingName)
			return
		}

		if option.Type() == lua.LTFunction {
			if err := L.CallByParam(lua.P{
				Fn:      value.(*lua.LFunction),
				NRet:    1,
				Protect: true,
			}); err == nil {
				option = L.Get(-1)
				L.Pop(1)
			} else {
				log.Printf("failed to evaluate UI style registry function for key `%s`: %s", settingName, err)
				return
			}
		}

		if option.Type() == lua.LTUserData {
			ud := value.(*lua.LUserData)
			if style, ok := ud.Value.(*tcell.Style); ok {
				gLuaRegistry.uiStyleMap[settingName] = *style
			} else {
				log.Printf("invalid UI style registry user data for key: %s", settingName)
			}
		} else {
			log.Printf("unsupported UI style registry value type for key: %s", settingName)
		}
	})
}

// ----------------------------------------------------------------------------
// message call operation

// luaMsgActionExtractor takes Lua state and a message entry, and should return
// a Lua function pointer as action function of this message entry. When no valid
// action can be made for given message entry, this function returns nil.
type luaMsgActionExtractor func(L *lua.LState, msgEntry lua.LValue) *lua.LFunction

// luaMsgArgsMaker is message argument type converter function, it takes Lua state
// and returns a slice of Lua values as arguments for message action. This will
// be used for Go value conversion after determining actual Lua State used for
// running message call.
type luaMsgArgsMaker func(L *lua.LState) []lua.LValue

type luaMsgCallArgs struct {
	sourceName, registryKey, msg, variant string
	isAsync                               bool // if this message should be called on asynchronous Lua state
	getArgs                               luaMsgArgsMaker
}

var gLuaMsgActionExtractorMap = map[string]map[string]luaMsgActionExtractor{
	registryKeyCommand: {
		luaMsgVariantMain: extractLuaMsgActionWithTblKey(luaCommandActionFuncKey),
		luaCommandCompletionFuncKey: extractLuaMsgActionWithDefaultAction(
			luaCommandCompletionFuncKey,
			func(L *lua.LState) int {
				return 0
			},
		),
	},
	registryKeyEventHook: {
		luaMsgVariantMain: extractLuaMsgActionWithTblKey(luaEventHookActionFuncKey),
	},
	registryKeyMisc: {
		luaMsgVariantMain: extractLuaMsgActionWithTblKey(luaMiscActionFuncKey),
	},
	registryKeyPreviewer: {
		luaMsgVariantMain:        extractLuaMsgActionWithTblKey(luaPreviewerActionFuncKey),
		luaPreviewerCleanFuncKey: extractLuaMsgActionWithTblKey(luaPreviewerCleanFuncKey),
		luaPreviewerConditionFuncKey: extractLuaMsgActionWithDefaultAction(
			luaPreviewerConditionFuncKey,
			func(L *lua.LState) int {
				L.Push(lua.LTrue)
				return 1
			},
		),
	},
	registryKeySortingMethod: {
		luaMsgVariantMain: extractLuaMsgActionWithTblKey(luaSortingMethodActionFuncKey),
	},
	registryKeyUIFormatter: {
		luaMsgVariantMain: extractLuaMsgActionWithTblKey(luaUIFormatterActionFuncKey),
	},
	registryKeyUIPrinter: {
		luaMsgVariantMain: extractLuaMsgActionWithTblKey(luaUIPrinterActionFuncKey),
	},
}

// extractLuaMsgActionWithTblKey returns a action extractor function, returned
// function will retruns message entry as a Lua function if it is one, if that
// entry is a table, extractor will check the value for given key, and returns
// that value if it is a function.
// Otherwise extractor returns nil.
func extractLuaMsgActionWithTblKey(key string) luaMsgActionExtractor {
	return func(L *lua.LState, msgEntry lua.LValue) *lua.LFunction {
		switch msgEntry.Type() {
		case lua.LTFunction:
			//  if entry is a function, return it as is.
			return msgEntry.(*lua.LFunction)
		case lua.LTTable:
			// if entry is a table, try to get actioin function with given key
			actionFunc := msgEntry.(*lua.LTable).RawGetString(key)
			if actionFunc.Type() == lua.LTFunction {
				return actionFunc.(*lua.LFunction)
			}
		}
		// invalid entry
		return nil
	}
}

// extractLuaMsgActionWithDefaultAction returns a action extractor that works
// pretty much like extractLuaMsgActionWithTblKey ones, but will return a Lua
// function made by wrapping `defualtAction` when message entry is defined as
// a function itself, or is defined as a table but does not contains specified key.
func extractLuaMsgActionWithDefaultAction(key string, defaultAction lua.LGFunction) luaMsgActionExtractor {
	return func(L *lua.LState, msgEntry lua.LValue) *lua.LFunction {
		switch msgEntry.Type() {
		case lua.LTFunction:
			return L.NewFunction(defaultAction)
		case lua.LTTable:
			actionFunc := msgEntry.(*lua.LTable).RawGetString(key)
			switch actionFunc.Type() {
			case lua.LTFunction:
				return actionFunc.(*lua.LFunction)
			case lua.LTNil:
				return L.NewFunction(defaultAction)
			}
		}
		// invalid entry
		return nil
	}
}

// getLuaMsgEntry looks up Lua registry table for target message entry.
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

// getLuaMsgAction finds message action of specified message.
func getLuaMsgAction(L *lua.LState, sourceName, registryKey, msg, variant string) (*lua.LFunction, error) {
	entry, err := getLuaMsgEntry(L, sourceName, registryKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get msg entry: %s", err)
	}

	var action *lua.LFunction
	var extractor luaMsgActionExtractor
	if extractorMap, ok := gLuaMsgActionExtractorMap[registryKey]; ok {
		extractor = extractorMap[variant]
	}

	if extractor != nil {
		action = extractor(L, entry)
	} else {
		action, _ = entry.(*lua.LFunction)
	}
	if action == nil {
		return nil, fmt.Errorf("failed to get valid msg action function")
	}

	return action, nil
}

// makeLuaMsgArgsWrapper returns a luaMsgArgsMaker function that converts all
// of passed arguments into Lua value slices.
func makeLuaMsgArgsWrapper(args ...any) luaMsgArgsMaker {
	return func(L *lua.LState) []lua.LValue {
		result := make([]lua.LValue, len(args))
		for i, arg := range args {
			if value, err := goValueToLuaValue(L, arg); err == nil {
				result[i] = value
			} else {
				log.Printf("Lua message argument wrapper error at value %v: %s", arg, err)
				result[i] = lua.LNil
			}
		}
		return result
	}
}

// callLuaMsgOnState finds and runs target Lua message action on given Lua state.
func callLuaMsgOnState(L *lua.LState, callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	if !gOpts.luamsglog {
		// pass
	} else if callArgs.variant != "" {
		log.Printf("call Lua msg: (%s, %s, %s)@%s", callArgs.sourceName, callArgs.registryKey, callArgs.msg, callArgs.variant)
	} else {
		log.Printf("call Lua msg: (%s, %s, %s)", callArgs.sourceName, callArgs.registryKey, callArgs.msg)
	}

	entry, err := getLuaMsgEntry(L, callArgs.sourceName, callArgs.registryKey, callArgs.msg)
	if err != nil {
		return nil, err
	}

	var action *lua.LFunction
	var extractor luaMsgActionExtractor
	if extractorMap, ok := gLuaMsgActionExtractorMap[callArgs.registryKey]; ok {
		extractor = extractorMap[callArgs.variant]
	}

	if extractor != nil {
		action = extractor(L, entry)
	} else {
		action, _ = entry.(*lua.LFunction)
	}
	if action == nil {
		return nil, fmt.Errorf("can't get valid msg action function")
	}

	var luaArgs []lua.LValue
	if callArgs.getArgs != nil {
		luaArgs = callArgs.getArgs(L)
	}

	oldTop := L.GetTop()

	L.Push(action)
	for _, arg := range luaArgs {
		L.Push(arg)
	}

	err = L.PCall(len(luaArgs), lua.MultRet, nil)
	if err != nil {
		return nil, err
	}

	nRet := L.GetTop() - oldTop
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

// callLuaMsgAsync gets a Lua state from pool and runs target Lua message on it.
func callLuaMsgAsync(callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	L, err := gLuaPool.get()
	if err != nil {
		return nil, err
	}
	defer gLuaPool.put(L)
	return callLuaMsgOnState(L, callArgs)
}

// callLuaMsgSync acquires global synchronous Lua state and runs target Lua message
// on it.
func callLuaMsgSync(callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	L, err := gLuaPool.acquireSyncState()
	defer gLuaPool.releaseSyncState()

	if err != nil {
		return nil, err
	}

	return callLuaMsgOnState(L, callArgs)
}

// callLuaMsg calls Lua message specified in call argument, this function will use
// the `isAsync` flag in argument to determine which Lua state source to use.
func callLuaMsg(callArgs luaMsgCallArgs) ([]lua.LValue, error) {
	if callArgs.isAsync {
		return callLuaMsgAsync(callArgs)
	} else {
		return callLuaMsgSync(callArgs)
	}
}

// callLuaMsgExpr calls Lua message specified by Lua message expression.
func callLuaMsgExpr(expr *luaMsgExpr, getArgs luaMsgArgsMaker) ([]lua.LValue, error) {
	return callLuaMsg(luaMsgCallArgs{
		sourceName:  expr.sourceName,
		registryKey: expr.registry,
		msg:         expr.msg,
		variant:     expr.variant,
		isAsync:     expr.isAsync,
		getArgs:     getArgs,
	})
}

// callLuaCommandCompletion calls completion message for given Lua command.
func callLuaCommandCompletion(expr *luaMsgExpr, args []string, longest string) ([]compMatch, string) {
	ret, err := callLuaMsg(
		luaMsgCallArgs{
			sourceName:  expr.sourceName,
			registryKey: expr.registry,
			msg:         expr.msg,
			variant:     luaCommandCompletionFuncKey,
			isAsync:     expr.isAsync,
			getArgs: func(L *lua.LState) []lua.LValue {
				tbl := L.NewTable()
				for i, arg := range args {
					tbl.RawSetInt(i+1, lua.LString(arg))
				}
				return []lua.LValue{tbl, lua.LString(longest)}
			},
		},
	)

	var matches []compMatch

	if err != nil {
		log.Printf("failed to call Lua command completion function: %s", err)
		return matches, longest
	}

	nRet := len(ret)
	if nRet == 0 {
		return matches, longest
	}

	ret1 := ret[0]
	switch ret1.Type() {
	case lua.LTTable:
		matchesTbl := ret1.(*lua.LTable)

		cnt := matchesTbl.Len()
		for i := 1; i <= cnt; i++ {
			value := matchesTbl.RawGetInt(i)
			switch value.Type() {
			case lua.LTString:
				str := value.String()
				matches = append(matches, compMatch{str, str})
			case lua.LTUserData:
				if v, ok := value.(*lua.LUserData).Value.(*compMatch); ok {
					matches = append(matches, *v)
				} else {
					log.Printf("matches list element #%d is not a valid user data", i)
				}
			default:
				log.Printf("matches list element #%d be string or user data, found %s: %s", i, value.Type(), value)
			}
		}
	case lua.LTNil:
		// pass
	default:
		log.Println("return value #1 of completion function should be a table")
		return matches, longest
	}

	if len(ret) >= 2 {
		ret2 := ret[1]
		switch ret2.Type() {
		case lua.LTString:
			longest = ret2.String()
		case lua.LTNil:
			// pass
		default:
			log.Println("return value #2 of completion function should be a string")
			return nil, longest
		}
	}

	return matches, longest
}

// callLuaEventHooks calls all Lua event hooks under given command name.
func callLuaEventHooks(cmdName string, getArgs luaMsgArgsMaker) error {
	exprList, ok := gLuaRegistry.eventHooks[cmdName]
	if !ok {
		return nil
	}

	errCnt := 0
	for _, expr := range exprList {
		_, err := callLuaMsgExpr(expr, getArgs)
		if err != nil {
			errCnt++
			log.Printf("failed to run hook %s: %s", expr, err)
		}
	}

	if errCnt > 0 {
		return fmt.Errorf("%d error(s) occured during event hook call, see log for more detail", errCnt)
	}

	return nil
}

type luaPreviewerPipe struct {
	m   sync.Mutex
	buf *bytes.Buffer

	ticker *time.Ticker
	wake   chan struct{}

	closed bool
	done   chan struct{}

	volatile bool

	previewErr error
}

func newLuaPreviewerPipe() *luaPreviewerPipe {
	lp := &luaPreviewerPipe{
		buf: new(bytes.Buffer),

		ticker: time.NewTicker(10 * time.Millisecond),
		wake:   make(chan struct{}, 1),
		done:   make(chan struct{}),
	}

	go lp.wakeupLoop()

	return lp
}

// wakeupLoop runs in background goroutine, wakes blocked `Read` call from time
// to time for checking new content in pipe buffer.
func (lp *luaPreviewerPipe) wakeupLoop() {
	for {
		select {
		case <-lp.ticker.C:
			select {
			case lp.wake <- struct{}{}:
			default:
			}
		case <-lp.done:
			return
		}
	}
}

// Write puts data to pipe buffer. Returns `io.ErrClosedPipe` when pipe is closed,
// so that writing side can known no more data is required on reading side.
func (lp *luaPreviewerPipe) Write(p []byte) (n int, err error) {
	lp.m.Lock()
	defer lp.m.Unlock()

	if lp.closed {
		return 0, io.ErrClosedPipe
	}

	return lp.buf.Write(p)
}

// Read tries to read from pipe buffer, and if there is currently nothing to read,
// this function will wait for small amount of time and then try reading again.
func (lp *luaPreviewerPipe) Read(p []byte) (n int, err error) {
	maxTry := 100

	for range maxTry {
		lp.m.Lock()

		n, _ = lp.buf.Read(p)
		if n > 0 {
			lp.m.Unlock()
			return n, nil
		}

		if lp.closed {
			lp.m.Unlock()
			return 0, io.EOF
		}

		lp.m.Unlock()

		select {
		case <-lp.wake:
			continue
		case <-lp.done:
			continue
		case <-time.After(10 * time.Second):
			return 0, errors.New("previewer pipe read timeout")
		}
	}

	return 0, fmt.Errorf("previewer reads nothing after %d attempt", maxTry)
}

func (lp *luaPreviewerPipe) Close() error {
	lp.m.Lock()
	defer lp.m.Unlock()

	if lp.closed {
		return nil
	}

	lp.closed = true

	close(lp.done)
	lp.ticker.Stop()

	return nil
}

// setVolatile updates volatile indicator on pipe.
func (lp *luaPreviewerPipe) setVolatile(isVolatile bool) {
	lp.m.Lock()
	defer lp.m.Unlock()

	lp.volatile = isVolatile
}

// isVolatile returns flag indicating if preview result should be marked as volatile.
func (lp *luaPreviewerPipe) isVolatile() bool {
	lp.m.Lock()
	defer lp.m.Unlock()

	return lp.volatile
}

// wait blocks goroutine until previewer pip is closed.
func (lp *luaPreviewerPipe) wait() {
	<-lp.done
}

// setPreviewError bind Lua execution error to pipe.
func (lp *luaPreviewerPipe) setPreviewError(err error) {
	lp.m.Lock()
	defer lp.m.Unlock()

	lp.previewErr = err
}

// checkPreviewError retruns Lua execution error binded to pipe.
func (lp *luaPreviewerPipe) checkPreviewError() error {
	lp.m.Lock()
	defer lp.m.Unlock()
	return lp.previewErr
}

// callLuaPreviewerConditionChecker calls condition message for given Lua previewer.
// And returns a bool flag indicating if this previewer is active for given argument.
func callLuaPreviewerConditionChecker(expr *luaMsgExpr, path string) (bool, error) {
	ret, err := callLuaMsg(luaMsgCallArgs{
		sourceName:  expr.sourceName,
		registryKey: expr.registry,
		msg:         expr.msg,
		variant:     luaPreviewerConditionFuncKey,
		isAsync:     expr.isAsync,
		getArgs: func(L *lua.LState) []lua.LValue {
			return []lua.LValue{lua.LString(path)}
		},
	})

	if err != nil {
		return false, err
	}

	if len(ret) == 0 {
		return false, nil
	}

	return lua.LVAsBool(ret[0]), nil
}

// callLuaPreviewerAction calls action message for given Lua previewer. when this
// function returns true, preview content should be marked as volatile, just link
// an non-zero exit code returned by previewer command.
func callLuaPreviewerAction(expr *luaMsgExpr, path string, w, h, x, y int, mode string) *luaPreviewerPipe {
	pipe := newLuaPreviewerPipe()

	go func() {
		writer := bufio.NewWriter(pipe)
		defer pipe.Close()
		defer writer.Flush()

		ret, err := callLuaMsgExpr(expr, func(L *lua.LState) []lua.LValue {
			return []lua.LValue{
				lWrapBufWriter(L, writer),
				lua.LString(path),
				lua.LNumber(w),
				lua.LNumber(h),
				lua.LNumber(x),
				lua.LNumber(y),
				lua.LString(mode),
			}
		})

		if err != nil {
			pipe.setPreviewError(err)
			pipe.setVolatile(false)
			return
		}

		nRet := len(ret)
		if nRet > 0 {
			pipe.setVolatile(lua.LVAsBool(ret[0]))
		}

		if nRet > 1 {
			luaErr := ret[1]
			if luaErr.Type() != lua.LTNil {
				pipe.setPreviewError(errors.New(luaErr.String()))
			}
		}
	}()

	return pipe
}

// callLuaPreviewerCleaning calls clean message for given Lua previewer.
func callLuaPreviewerCleaning(expr *luaMsgExpr, previousFile string, w, h, x, y int, nextFile string) error {
	_, err := callLuaMsg(luaMsgCallArgs{
		sourceName:  expr.sourceName,
		registryKey: expr.registry,
		msg:         expr.msg,
		variant:     luaPreviewerCleanFuncKey,
		isAsync:     expr.isAsync,
		getArgs: func(L *lua.LState) []lua.LValue {
			return []lua.LValue{
				lua.LString(previousFile),
				lua.LNumber(w),
				lua.LNumber(h),
				lua.LNumber(x),
				lua.LNumber(y),
				lua.LString(nextFile),
			}
		},
	})

	return err
}

// getLuaPreviewerForPath search for active previewer for certain path.
func getLuaPreviewerForPath(path string) *luaPreviewerInfo {
	var result *luaPreviewerInfo
	for i := range gLuaRegistry.previewers {
		previewer := &gLuaRegistry.previewers[i]
		ok, err := callLuaPreviewerConditionChecker(&previewer.msgexpr, path)
		if err != nil {
			log.Printf("failed to check condition for previewer %s: %s", previewer.name, err)
		} else if ok {
			result = previewer
			break
		}
	}

	return result
}

// getLuaPreviewerNames returns name list of all registered Lua previewers.
func getLuaPreviewerNames() []string {
	names := make([]string, len(gLuaRegistry.previewers))
	for i, previewer := range gLuaRegistry.previewers {
		names[i] = previewer.name
	}
	return names
}

// callLuaKeyMapMsgOnState calls Lua key map message on given Lua state.
func callLuaKeyMapMsgOnState(L *lua.LState, expr *luaKeyMapExpr) error {
	if gOpts.luamsglog {
		log.Printf("call Lua key map: %s - %s.%s", expr.sourceName, expr.keyMapType, expr.key)
	}

	keyMapGroup, err := getLuaMsgEntry(L, expr.sourceName, registryKeyKeyMap, expr.keyMapType)
	if err != nil {
		return err
	}

	groupTbl, ok := keyMapGroup.(*lua.LTable)
	if !ok {
		return fmt.Errorf("key map group is not table value")
	}

	var action *lua.LFunction

	entry := groupTbl.RawGetString(expr.key)
	switch entry.Type() {
	case lua.LTFunction:
		action = entry.(*lua.LFunction)
	case lua.LTTable:
		value := entry.(*lua.LTable).RawGetString(luaKeyMapActionFuncKey)
		action, _ = value.(*lua.LFunction)
	case lua.LTNil:
		return fmt.Errorf("no action found")
	default:
		return fmt.Errorf("not supported action value type")
	}

	if action == nil {
		return fmt.Errorf("no action found")
	}

	L.Push(action)
	L.Push(lua.LNumber(expr.count))

	err = L.PCall(1, 0, nil)
	if err != nil {
		return err
	}

	return nil
}

// callLuaKeyMapMsg calls Lua key map message, pick proper Lua state source according
// to `isAsync` flag in message expression.
func callLuaKeyMapMsg(expr *luaKeyMapExpr) error {
	if expr.isAsync {
		L, err := gLuaPool.get()
		if err != nil {
			return err
		}
		defer gLuaPool.put(L)
		return callLuaKeyMapMsgOnState(L, expr)
	} else {
		L, err := gLuaPool.acquireSyncState()
		defer gLuaPool.releaseSyncState()

		if err != nil {
			return err
		}

		return callLuaKeyMapMsgOnState(L, expr)
	}
}

// getLuaSortingMethodNames returns name list of all registered Lua sort method.
func getLuaSortingMethodNames() []string {
	return slices.Collect(maps.Keys(gLuaRegistry.sortingMethod))
}

// getLuaSortingMethod returns Lua message expression for sort method with given name.
func getLuaSortingMethod(name string) *luaMsgExpr {
	return gLuaRegistry.sortingMethod[name]
}

// sortByLuaMsg pass given file list to Lua sort method and update file list order
// in place.
func sortByLuaMsg(expr *luaMsgExpr, files []*file, isReverse bool) error {
	retList, err := callLuaMsgExpr(expr, func(L *lua.LState) []lua.LValue {
		udTbl := L.NewTable()
		for _, file := range files {
			udTbl.Append(lWrapFile(L, file))
		}
		return []lua.LValue{udTbl}
	})

	if err != nil {
		return fmt.Errorf("%s", err)
	} else if len(retList) <= 0 {
		return fmt.Errorf("Lua sort method returns nothing")
	}

	ret := retList[0]
	if ret.Type() != lua.LTTable {
		return fmt.Errorf("return value of Lua function is not a table")
	}

	retTbl := ret.(*lua.LTable)
	nElem := retTbl.Len()
	if nElem != len(files) {
		return fmt.Errorf("number of elements in returned table does not match number of files")
	}

	result := make([]*file, nElem)
	for i := 1; i <= nElem; i++ {
		value := retTbl.RawGetInt(i)
		if value.Type() != lua.LTUserData {
			return fmt.Errorf("element %d in returned table is not userdata", i)
		}

		file, ok := value.(*lua.LUserData).Value.(*file)
		if !ok {
			return fmt.Errorf("element %d is not a *file data", i)
		}

		result[i-1] = file
	}

	if isReverse {
		for i := range files {
			files[i] = result[nElem-i-1]
		}
	} else {
		for i := range files {
			files[i] = result[i]
		}
	}

	return nil
}

// getLuaUIFormatter finds Lua UI formatter message with given name.
func getLuaUIFormatter(name string) *luaMsgExpr {
	return gLuaRegistry.uiFormatter[name]
}

// callLuaUIFormatter calls a Lua UI formatter message and returns the string
// build by formatter.
func callLuaUIFormatter(expr *luaMsgExpr, getArgs luaMsgArgsMaker) (string, error) {
	ret, err := callLuaMsgExpr(expr, getArgs)
	if err != nil {
		return "", err
	}

	if len(ret) == 0 {
		return "", fmt.Errorf("Lua UI formatter does not return a string")
	}

	value := ret[0]
	if value.Type() != lua.LTString {
		return "", fmt.Errorf("Lua UI formatter does not return a string")
	}

	return string(value.(lua.LString)), nil
}

// callLuaUIFormatterIgnoreError is a wrapper of callLuaUIFormatter which logs
// any error returned by callLuaUIFormatter without returning it.
func callLuaUIFormatterIgnoreError(expr *luaMsgExpr, getArgs luaMsgArgsMaker) string {
	ret, err := callLuaUIFormatter(expr, getArgs)
	if err != nil {
		log.Printf("failed to execute Lua UI formatter %s: %s", expr, err)
	}
	return ret
}

// callLuaUIFormatterWithSingleParam calls Lua UI formatter with single string
// parameter, if such formatter does not exists, a string build with given default
// format string will be returned.
func callLuaUIFormatterWithSingleParam(formatterName, defaultFmtStr, param string) string {
	luaFormatter := getLuaUIFormatter(formatterName)
	if luaFormatter != nil {
		return callLuaUIFormatterIgnoreError(luaFormatter, makeLuaMsgArgsWrapper(param))
	}
	return fmt.Sprintf(optionToFmtstr(defaultFmtStr), param)
}

// getLuaUIPrinter finds Lua UI printer message with given name.
func getLuaUIPrinter(name string) *luaMsgExpr {
	return gLuaRegistry.uiPrinter[name]
}

// getLuaUIStyle looks up Lua UI style registry with given name
func getLuaUIStyle(name string, defaultFmtStr string) (tcell.Style, bool) {
	style, ok := gLuaRegistry.uiStyleMap[name]
	return style, ok
}

// getLuaUIStyleWithDefaultStr looks up Lua UI style registry with given name.
// When target key does not exists, make a new style object with default format
// string.
func getLuaUIStyleWithDefaultStr(name string, defaultFmtStr string) tcell.Style {
	style, ok := gLuaRegistry.uiStyleMap[name]
	if !ok {
		style = parseEscapeSequence(defaultFmtStr)
	}
	return style
}

// getLuaMiscMsg returns misc message with given name. If no such message is registered,
// this function returns nil.
func getLuaMiscMsg(name string) *luaMsgExpr {
	return gLuaRegistry.misc[name]
}

func formatDuplicatedFilenameWithLuaMsg(expr *luaMsgExpr, basename, ext string, dupIndex int) (string, error) {
	ret, err := callLuaMsgExpr(expr, makeLuaMsgArgsWrapper(basename, ext, dupIndex))
	if err != nil {
		return "", err
	}

	if len(ret) <= 0 {
		return "", fmt.Errorf("Lua message returns nonthing")
	}

	strValue, ok := ret[0].(lua.LString)
	if !ok {
		return "", fmt.Errorf("return value #1 is not a string")
	}

	return string(strValue), nil
}

// makeShellCmdWithLuaMsg creates `exec.Cmd` object by calling Lua message.
func makeShellCmdWithLuaMsg(expr *luaMsgExpr, cmd_name string, args []string) (*exec.Cmd, error) {
	ret, err := callLuaMsgExpr(expr, makeLuaMsgArgsWrapper(cmd_name, args))
	if err != nil {
		return nil, err
	}

	if len(ret) <= 0 {
		return nil, fmt.Errorf("Lua shell command maker returns nonthing")
	}

	ret1 := ret[0]
	ud, ok := ret1.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("return value #1 of Lua shell command maker is not a userdata")
	}

	cmd, ok := ud.Value.(*exec.Cmd)
	if !ok {
		return nil, fmt.Errorf("return value #1 of Lua shell command is not a Cmd object")
	}

	return cmd, nil
}
