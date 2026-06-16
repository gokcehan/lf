package main

import (
	"bufio"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type bufio.Writer

const luaBufWriterTypeName = "bufio.Writer"

func lRegisterBufWriterType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaBufWriterTypeName)

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

func lCheckBufWriter(L *lua.LState, index int) *bufio.Writer {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*bufio.Writer); ok {
		return v
	}

	L.ArgError(index, "value of type `BufWriter` expected")

	return nil
}

func lWrapBufWriter(L *lua.LState, data *bufio.Writer) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaBufWriterTypeName))

	return ud
}

func lAddBufWriterToState(L *lua.LState, data *bufio.Writer) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapBufWriter(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaBufWriterAvailable returns available byte count in buffer.
func luaBufWriterAvailable(L *lua.LState) int {
	writer := lCheckBufWriter(L, 1)
	L.Push(lua.LNumber(writer.Available()))
	return 1
}

// luaBufWriterBuffered returns number of bytes that has been written to buffer.
func luaBufWriterBuffered(L *lua.LState) int {
	writer := lCheckBufWriter(L, 1)
	L.Push(lua.LNumber(writer.Buffered()))
	return 1
}

// luaBufWriterFlush writes buffered data to underlying output writer.
func luaBufWriterFlush(L *lua.LState) int {
	writer := lCheckBufWriter(L, 1)
	err := writer.Flush()

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// luaBufWriterSize returns byte size of underlying buffer.
func luaBufWriterSize(L *lua.LState) int {
	writer := lCheckBufWriter(L, 1)
	L.Push(lua.LNumber(writer.Size()))
	return 1
}

// luaBufWriterWriteString writes string value to buffer.
func luaBufWriterWriteString(L *lua.LState) int {
	writer := lCheckBufWriter(L, 1)

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

const luaBufReaderTypeName = "bufio.Reader"

func lRegisterBufReaderType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaBufReaderTypeName)

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

func lCheckBufReader(L *lua.LState, index int) *bufio.Reader {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*bufio.Reader); ok {
		return v
	}

	L.ArgError(index, "value of type `BufReader` expected")

	return nil
}

func lWrapBufReader(L *lua.LState, data *bufio.Reader) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaBufReaderTypeName))

	return ud
}

func lAddBufReaderToState(L *lua.LState, data *bufio.Reader) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapBufReader(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaBufReaderBuffered returns number of bytes buffered.
func luaBufReaderBuffered(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)
	L.Push(lua.LNumber(reader.Buffered()))
	return 1
}

// luaBufReaderDiscard skips following n bytes, returns number of bytes skipped.
func luaBufReaderDiscard(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)
	n := L.CheckInt(2)

	discarded, err := reader.Discard(n)
	L.Push(lua.LNumber(discarded))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// luaBufReaderPeek returns next n bytes without advancing reader.
func luaBufReaderPeek(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)
	n := L.CheckInt(2)

	buf, err := reader.Peek(n)
	L.Push(lua.LString(string(buf)))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// luaBufReaderRead reads n bytes from reader, and returns datga as a string.
func luaBufReaderRead(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)
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

// luaBufReaderReadLine reads one line of data. Returned data will not contain
// trailing `\r\n` or `\n`
func luaBufReaderReadLine(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)

	line, isPrefix, err := reader.ReadLine()

	L.Push(lua.LString(string(line)))
	L.Push(lua.LBool(isPrefix))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 3
	}

	return 2
}

// luaBufReaderReadString takes a delimiter string, and reads until that string
// occurs.
func luaBufReaderReadString(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)
	delim := L.CheckString(2)

	str, err := reader.ReadString(delim[0])
	L.Push(lua.LString(str))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// luaBufReaderSize returns byte size of underlying buffer.
func luaBufReaderSize(L *lua.LState) int {
	reader := lCheckBufReader(L, 1)
	L.Push(lua.LNumber(reader.Size()))
	return 1
}
