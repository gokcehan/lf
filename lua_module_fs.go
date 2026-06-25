package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

func lfFsModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"mkdir":     luaFsMkdir,
		"mkdir_all": luaFsModuleMkdirAll,
		"link":      luaFsModuleLink,
		"symlink":   luaFsModuleSymlink,
		"copy":      luaFsModuleCopyFile,

		"join":      luaFsModuleJoin,
		"split":     luaFsModuleSplit,
		"split_ext": luaFsModuleSplitExt,
		"dirname":   luaFsModuleDirname,
		"basename":  luaFsModuleBasename,
		"ext":       luaFsModuleExt,

		"stat":    luaFsModuleStat,
		"readdir": luaFsModuleReadDir,
	})

	L.Push(mod)

	return 1
}

func luaFsMkdir(L *lua.LState) int {
	path := L.CheckString(1)

	if err := os.Mkdir(path, 0o777); err == nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(err.Error()))
	}

	return 1
}

func luaFsModuleMkdirAll(L *lua.LState) int {
	path := L.CheckString(1)

	if err := os.MkdirAll(path, 0o777); err == nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(err.Error()))
	}

	return 1
}

func luaFsModuleLink(L *lua.LState) int {
	oldname := L.CheckString(1)
	newname := L.CheckString(2)
	force := L.OptBool(3, false)

	if stat, err := os.Lstat(newname); err == nil {
		if stat.Mode()&os.ModeSymlink == 0 {
			L.Push(lua.LString(fmt.Sprintf("%s already exists and is not a symlink", newname)))
			return 1
		}

		if force {
			err := os.Remove(newname)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to remove existing link: %s", err)))
				return 1
			}
		} else {
			L.Push(lua.LString(fmt.Sprintf("link %s already exists", newname)))
			return 1
		}
	}

	if err := os.Link(oldname, newname); err == nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(err.Error()))
	}

	return 1
}

func luaFsModuleSymlink(L *lua.LState) int {
	oldname := L.CheckString(1)
	newname := L.CheckString(2)
	force := L.OptBool(3, false)

	if stat, err := os.Lstat(newname); err == nil {
		if stat.Mode()&os.ModeSymlink == 0 {
			L.Push(lua.LString(fmt.Sprintf("%s already exists and is not a symlink", newname)))
			return 1
		}

		if force {
			err := os.Remove(newname)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("failed to remove existing link: %s", err)))
				return 1
			}
		} else {
			L.Push(lua.LString(fmt.Sprintf("link %s already exists", newname)))
			return 1
		}
	}

	if err := os.Symlink(oldname, newname); err == nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(err.Error()))
	}

	return 1
}

func luaFsModuleCopyFile(L *lua.LState) int {
	src := L.CheckString(1)
	dst := L.CheckString(2)

	srcFile, err := os.Open(src)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	defer dstFile.Close()

	reader := bufio.NewReader(srcFile)
	writer := bufio.NewWriter(dstFile)

	_, err = io.Copy(writer, reader)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	err = writer.Flush()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaFsModuleJoin(L *lua.LState) int {
	cnt := L.GetTop()
	parts := []string{}

	for i := 1; i <= cnt; i++ {
		parts = append(parts, L.CheckString(i))
	}

	result := filepath.Join(parts...)
	L.Push(lua.LString(result))

	return 1
}

func luaFsModuleSplit(L *lua.LState) int {
	path := L.CheckString(1)
	dirname, basename := filepath.Split(path)
	L.Push(lua.LString(dirname))
	L.Push(lua.LString(basename))
	return 2
}

func luaFsModuleSplitExt(L *lua.LState) int {
	path := L.CheckString(1)
	ext := filepath.Ext(path)
	stem := path[:len(path)-len(ext)]
	L.Push(lua.LString(stem))
	L.Push(lua.LString(ext))
	return 2
}

func luaFsModuleDirname(L *lua.LState) int {
	path := L.CheckString(1)
	result := filepath.Dir(path)
	L.Push(lua.LString(result))
	return 1
}

func luaFsModuleBasename(L *lua.LState) int {
	path := L.CheckString(1)
	result := filepath.Base(path)
	L.Push(lua.LString(result))
	return 1
}

func luaFsModuleExt(L *lua.LState) int {
	path := L.CheckString(1)
	result := filepath.Ext(path)
	L.Push(lua.LString(result))
	return 1
}

func luaFsModuleStat(L *lua.LState) int {
	path := L.CheckString(1)

	stat, err := os.Stat(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	lAddFileInfoToState(L, stat)

	return 1
}

func luaFsModuleReadDir(L *lua.LState) int {
	path := L.CheckString(1)

	tbl := L.NewTable()

	entries, err := os.ReadDir(path)
	if err != nil {
		L.Push(tbl)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	for _, entry := range entries {
		tbl.Append(lWrapDirEntry(L, entry))
	}

	L.Push(tbl)

	return 1
}
