package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const pluginDirName = "plugins"

func setupLuaTypeBindings(L *lua.LState) {
	lfTypes := L.NewTable()

	lfTypes.RawSetString("File", LRegisterFileTypeMt(L))
	lfTypes.RawSetString("App", LRegisterAppType(L))
	lfTypes.RawSetString("UI", LRegisterUIType(L))
}

// setupScripImportPath appends plugin root directory paths to Lua loader search
// list.
func setupScripImportPath(L *lua.LState, rootDirs []string) error {
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
	for _, rootDir := range rootDirs {
		pluginDir := filepath.Join(rootDir, pluginDirName)

		builder.WriteString(";")
		builder.WriteString(pluginDir)
		builder.WriteString("/?.lua")

		builder.WriteString(";")
		builder.WriteString(pluginDir)
		builder.WriteString("/?/init.lua")
	}

	path = builder.String()

	L.SetField(pack, "path", lua.LString(path))
	log.Println("Lua loader search path:", path)

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

// luaStateInit create and initialize a new Lua state.
func luaStateInit(app *app, rootDirs []string) (*lua.LState, error) {
	log.Println("initializing Lua state")
	L := lua.NewState()

	setupLuaTypeBindings(L)

	// setup modules
	err := setupScripImportPath(L, rootDirs)
	if err != nil {
		err = fmt.Errorf("failed to setup Lua loader search path: %s", err)
	}

	setupLuaGlobals(app, L)

	log.Println("Lua state initialized")

	return L, err
}

// loadLuaPluginFromDir loads all plugins under given directory.
func loadLuaPluginFromDir(L *lua.LState, root string) error {
	pluginDir := filepath.Join(root, pluginDirName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %s", err)
	}

	failed := []string{}

	// only directories are treated as plugin entrance.
	// So that user can put Lua development config files under plugin root with ease.
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		scriptPath := filepath.Join(pluginDir, name, "init.lua")

		if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
			if err = L.DoFile(scriptPath); err != nil {
				failed = append(failed, name)
				log.Println("failed to execute plugin script:", scriptPath)
				log.Println(err)
			} else {
				log.Println("plugin loaded:", scriptPath)
			}
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to load plugins (more details in log): %s", strings.Join(failed, ", "))
	}

	return nil
}
