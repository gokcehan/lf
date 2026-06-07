package main

import (
	"bufio"

	lua "github.com/yuin/gopher-lua"
)

const LuaBufWriterTypeName = "lf.buf_writer"

func LRegisterBufWriterType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaBufWriterTypeName)

	L.SetFuncs(mt, luaBufWriterStaticMethod)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaBufWriterMethods))

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

var luaBufWriterStaticMethod = map[string]lua.LGFunction{}

// ----------------------------------------------------------------------------

var luaBufWriterMethods = map[string]lua.LGFunction{
	"available":    luaBufWriterAvailable,
	"buffered":     luaBufWriterBuffered,
	"flush":        luaBufWriterFlush,
	"size":         luaBufWriterSize,
	"write_string": luaBufWriterWriteString,
}

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
	str := L.CheckString(2)
	n, err := writer.WriteString(str)

	if err != nil {
		L.Push(lua.LNumber(n))
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LNumber(n))

	return 1
}
