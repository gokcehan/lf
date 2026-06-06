package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

const pluginDirName = "plugins"

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
			log.Println("failed to read plugin directory %s: %s", pluginDir, err)
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
				proto, err := CompileLua(scriptPath)
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

func (pl *lStatePool) new() *lua.LState {
	L := lua.NewState()

	setupLuaTypeBindings(L)

	// setup modules
	err := setupScripImportPath(L, pl.rootDirs)
	if err != nil {
		log.Printf("failed to setup Lua loader search path: %s", err)
	}

	setupLuaGlobals(pl.app, L)

	L.PreloadModule("lf", LfMainModuleLoader)
	L.PreloadModule("lf.utf8", LfUtf8ModuleLoader)

	for _, proto := range pl.pluginByteCodes {
		err = DoCompiledFile(L, proto)
		if err != nil {
			log.Printf("failed to execute plugin script: %s\n%s", proto.SourceName, err)
		}
	}

	return L
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

var gLuaRegistry struct {
	sortMethod map[string]*lua.LFunction
	eventHooks map[string][]*lua.LFunction
}

// CompileLua reads the passed lua file from disk and compiles it.
func CompileLua(filePath string) (*lua.FunctionProto, error) {
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

// DoCompiledFile takes a FunctionProto, as returned by CompileLua, and runs it in the LState. It is equivalent
// to calling DoFile on the LState with the original source file.
func DoCompiledFile(L *lua.LState, proto *lua.FunctionProto) error {
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

// callLuaFuncWithArgList calls a Lua function with list of string arguments.
func callLuaFuncWithArgList(fn *lua.LFunction, args []string) error {
	L := gLuaStateSync.acquire()
	defer gLuaStateSync.release()

	L.Push(fn)
	for _, arg := range args {
		L.Push(lua.LString(arg))
	}

	return L.PCall(len(args), 0, nil)
}

// batchCallLuaFuncWithArgList calls a list of Lua functions with given argument
// list.
func batchCallLuaFuncWithArgList(app *app, fnList []*lua.LFunction, args []string) {
	for _, fn := range fnList {
		err := callLuaFuncWithArgList(fn, args)
		if err != nil {
			app.ui.echoerrf("error during Lua event hook call: %s", err)
		}
	}
}
