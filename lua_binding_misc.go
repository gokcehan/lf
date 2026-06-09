package main

import (
	"bufio"
	"io/fs"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type bufio.Writer

const LuaBufWriterTypeName = "bufio.Writer"

func LRegisterBufWriterType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaBufWriterTypeName)

	// L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"available":    luaBufWriterAvailable,
		"buffered":     luaBufWriterBuffered,
		"flush":        luaBufWriterFlush,
		"size":         luaBufWriterSize,
		"write_string": luaBufWriterWriteString,
	}))

	return mt
}

func LCheckBufWriter(L *lua.LState, index int) *bufio.Writer {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*bufio.Writer); ok {
		return v
	}

	L.ArgError(index, "value of type `BufWriter` expected")

	return nil
}

func LWrapBufWriter(L *lua.LState, data *bufio.Writer) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaBufWriterTypeName))

	return ud
}

func LAddBufWriterToState(L *lua.LState, data *bufio.Writer) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapBufWriter(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaBufWriterAvailable(L *lua.LState) int {
	writer := LCheckBufWriter(L, 1)
	L.Push(lua.LNumber(writer.Available()))
	return 1
}

func luaBufWriterBuffered(L *lua.LState) int {
	writer := LCheckBufWriter(L, 1)
	L.Push(lua.LNumber(writer.Buffered()))
	return 1
}

func luaBufWriterFlush(L *lua.LState) int {
	writer := LCheckBufWriter(L, 1)
	err := writer.Flush()

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaBufWriterSize(L *lua.LState) int {
	writer := LCheckBufWriter(L, 1)
	L.Push(lua.LNumber(writer.Size()))
	return 1
}

func luaBufWriterWriteString(L *lua.LState) int {
	writer := LCheckBufWriter(L, 1)

	nArgs := L.GetTop()
	sum := 0
	for i := 2; i <= nArgs; i++ {
		str := L.CheckString(i)
		n, err := writer.WriteString(str)
		sum += n

		if err != nil {
			L.Push(lua.LNumber(sum))
			L.Push(lua.LString(err.Error()))
			return 2
		}
	}

	L.Push(lua.LNumber(sum))

	return 1
}

// ----------------------------------------------------------------------------
// type bufio.Reader

const LuaBufReaderTypeName = "bufio.Reader"

func LRegisterBufReaderType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaBufReaderTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"buffered":    luaBufReaderBuffered,
		"discard":     luaBufReaderDiscard,
		"peek":        luaBufReaderPeek,
		"read":        luaBufReaderRead,
		"read_line":   luaBufReaderReadLine,
		"read_string": luaBufReaderReadString,
		"size":        luaBufReaderSize,
	}))

	return mt
}

func LCheckBufReader(L *lua.LState, index int) *bufio.Reader {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*bufio.Reader); ok {
		return v
	}

	L.ArgError(index, "value of type `BufReader` expected")

	return nil
}

func LWrapBufReader(L *lua.LState, data *bufio.Reader) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaBufReaderTypeName))

	return ud
}

func LAddBufReaderToState(L *lua.LState, data *bufio.Reader) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapBufReader(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaBufReaderBuffered(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)
	L.Push(lua.LNumber(reader.Buffered()))
	return 1
}

func luaBufReaderDiscard(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)
	n := L.CheckInt(2)

	discarded, err := reader.Discard(n)
	L.Push(lua.LNumber(discarded))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

func luaBufReaderPeek(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)
	n := L.CheckInt(2)

	buf, err := reader.Peek(n)
	L.Push(lua.LString(string(buf)))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

func luaBufReaderRead(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)
	n := L.CheckInt(2)

	buf := make([]byte, n)
	nRead, err := reader.Read(buf)

	L.Push(lua.LString(string(buf[:nRead])))
	L.Push(lua.LNumber(nRead))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 3
	}

	return 2
}

func luaBufReaderReadLine(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)

	line, isPrefix, err := reader.ReadLine()

	L.Push(lua.LString(string(line)))
	L.Push(lua.LBool(isPrefix))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 3
	}

	return 2
}

func luaBufReaderReadString(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)
	delim := L.CheckString(2)

	str, err := reader.ReadString(delim[0])
	L.Push(lua.LString(str))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

func luaBufReaderSize(L *lua.LState) int {
	reader := LCheckBufReader(L, 1)
	L.Push(lua.LNumber(reader.Size()))
	return 1
}

// ----------------------------------------------------------------------------
// Type fs.FileInfo

const LuaFileInfoTypeName = "fs.FileInfo"

func LRegisterFileInfoType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaFileInfoTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"name":     luaFileInfoName,
		"size":     luaFileInfoSize,
		"mode":     luaFileInfoMode,
		"mod_time": luaFileInfoModTime,
		"is_dir":   luaFileInfoIsDir,
	}))

	return mt
}

func LCheckFileInfo(L *lua.LState, index int) fs.FileInfo {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(fs.FileInfo); ok {
		return v
	}

	L.ArgError(index, "value of type `FileInfo` expected")

	return nil
}

func LWrapFileInfo(L *lua.LState, data fs.FileInfo) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaFileInfoTypeName))

	return ud
}

func LAddFileInfoToState(L *lua.LState, data fs.FileInfo) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapFileInfo(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaFileInfoName(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LString(info.Name()))
	return 1
}

func luaFileInfoSize(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LNumber(info.Size()))
	return 1
}

func luaFileInfoMode(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LNumber(info.Mode()))
	return 1
}

func luaFileInfoModTime(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	t := info.ModTime()
	return LAddTimeToState(L, &t)
}

func luaFileInfoIsDir(L *lua.LState) int {
	info := LCheckFileInfo(L, 1)
	L.Push(lua.LBool(info.IsDir()))
	return 1
}
