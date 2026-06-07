package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

func LfFsModuleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), LfFsModuleExports)

	L.Push(mod)

	return 1
}

var LfFsModuleExports = map[string]lua.LGFunction{
	"mkdir":     luaFsMkdir,
	"mkdir_all": luaFsMkdirAll,
	"link":      luaFsLink,
	"symlink":   luaFsSymlink,
	"copy":      luaFsCopyFile,

	"join":      luaFsJoin,
	"split":     luaFsSplit,
	"split_ext": luaFsSplitExt,
	"dirname":   luaFsDirname,
	"basename":  luaFsBasename,
	"ext":       luaFsExt,

	"stat":    luaFsStat,
	"readdir": luaFsReadDir,
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

func luaFsMkdirAll(L *lua.LState) int {
	path := L.CheckString(1)

	if err := os.MkdirAll(path, 0o777); err == nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(err.Error()))
	}

	return 1
}

func luaFsLink(L *lua.LState) int {
	oldname := L.CheckString(1)
	newname := L.CheckString(2)
	force := L.OptBool(3, false)

	if stat, err := os.Lstat(newname); err == nil {
		if stat.Mode()&os.ModeSymlink == 0 {
			L.Push(lua.LString(fmt.Sprintf("%s already exists and is not a symlink", newname)))
			return 1
		}

		if force {
			os.Remove(newname)
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

func luaFsSymlink(L *lua.LState) int {
	oldname := L.CheckString(1)
	newname := L.CheckString(2)
	force := L.OptBool(3, false)

	if stat, err := os.Lstat(newname); err == nil {
		if stat.Mode()&os.ModeSymlink == 0 {
			L.Push(lua.LString(fmt.Sprintf("%s already exists and is not a symlink", newname)))
			return 1
		}

		if force {
			os.Remove(newname)
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

func luaFsCopyFile(L *lua.LState) int {
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
	defer writer.Flush()

	_, err = io.Copy(writer, reader)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)

	return 1
}

func luaFsJoin(L *lua.LState) int {
	cnt := L.GetTop()
	parts := []string{}

	for i := 1; i <= cnt; i++ {
		parts = append(parts, L.CheckString(i))
	}

	result := filepath.Join(parts...)
	L.Push(lua.LString(result))

	return 1
}

func luaFsSplit(L *lua.LState) int {
	path := L.CheckString(1)
	dirname, basename := filepath.Split(path)
	L.Push(lua.LString(dirname))
	L.Push(lua.LString(basename))
	return 2
}

func luaFsSplitExt(L *lua.LState) int {
	path := L.CheckString(1)
	ext := filepath.Ext(path)
	stem := path[:len(path)-len(ext)]
	L.Push(lua.LString(stem))
	L.Push(lua.LString(ext))
	return 2
}

func luaFsDirname(L *lua.LState) int {
	path := L.CheckString(1)
	result := filepath.Dir(path)
	L.Push(lua.LString(result))
	return 1
}

func luaFsBasename(L *lua.LState) int {
	path := L.CheckString(1)
	result := filepath.Base(path)
	L.Push(lua.LString(result))
	return 1
}

func luaFsExt(L *lua.LState) int {
	path := L.CheckString(1)
	result := filepath.Ext(path)
	L.Push(lua.LString(result))
	return 1
}

func luaFsStat(L *lua.LState) int {
	path := L.CheckString(1)

	stat, err := os.Stat(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	LAddFileInfoToState(L, stat)

	return 1
}

func luaFsReadDir(L *lua.LState) int {
	path := L.CheckString(1)

	files, err := readdir(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	tbl := L.NewTable()
	for _, f := range files {
		tbl.Append(LWrapFile(L, f))
	}

	L.Push(tbl)

	return 1
}
