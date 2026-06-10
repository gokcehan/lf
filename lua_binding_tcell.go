package main

import (
	"github.com/gdamore/tcell/v3"
	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// type tcell.Style

const LuaTcellStyleTypeName = "tcell.Style"

func LRegisterTcellStyleType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaTcellStyleTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaTcellStyleNew,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"foreground":     luaTcellStyleForeground,
		"background":     luaTcellStyleBackground,
		"foreground_rgb": luaTcellStyleForegroundRGB,
		"background_rgb": luaTcellStyleBackgroundRGB,

		"normal":         luaTcellStyleNormal,
		"bold":           luaTcellStyleBold,
		"blink":          luaTcellStyleBlink,
		"dim":            luaTcellStyleDim,
		"italic":         luaTcellStyleItalic,
		"reverse":        luaTcellStyleReverse,
		"strike_through": luaTcellStyleStrikeThrough,
		"underline":      luaTcellStyleUnderline,
	}))

	return mt
}

func LCheckTcellStyle(L *lua.LState, index int) *tcell.Style {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*tcell.Style); ok {
		return v
	}

	L.ArgError(index, "value of type `TcellStyle` expected")

	return nil
}

func LWrapTcellStyle(L *lua.LState, data *tcell.Style) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaTcellStyleTypeName))

	return ud
}

func LAddTcellStyleToState(L *lua.LState, data *tcell.Style) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapTcellStyle(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTcellStyleNew(L *lua.LState) int {
	st := tcell.StyleDefault
	return LAddTcellStyleToState(L, &st)
}

// ----------------------------------------------------------------------------

func luaTcellStyleForeground(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	color := LCheckTcellColor(L, 2)
	*st = st.Foreground(*color)
	return 0
}

func luaTcellStyleBackground(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	color := LCheckTcellColor(L, 2)
	*st = st.Background(*color)
	return 0
}

func luaTcellStyleForegroundRGB(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	r := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	*st = st.Foreground(tcell.NewRGBColor(int32(r), int32(g), int32(b)))

	return 0
}

func luaTcellStyleBackgroundRGB(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	r := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	*st = st.Background(tcell.NewRGBColor(int32(r), int32(g), int32(b)))

	return 0
}

func luaTcellStyleNormal(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	*st = st.Normal()
	return 0
}

func luaTcellStyleBold(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Bold(isActive)
	return 0
}

func luaTcellStyleBlink(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Blink(isActive)
	return 0
}

func luaTcellStyleDim(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Dim(isActive)
	return 0
}

func luaTcellStyleItalic(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Italic(isActive)
	return 0
}

func luaTcellStyleReverse(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Reverse(isActive)
	return 0
}

func luaTcellStyleStrikeThrough(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.StrikeThrough(isActive)
	return 0
}

func luaTcellStyleUnderline(L *lua.LState) int {
	st := LCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Underline(isActive)
	return 0
}

// ----------------------------------------------------------------------------
// type tcell.Color

const LuaTcellColorTypeName = "tcell.Color"

func LRegisterTcellColorType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaTcellColorTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new_rgb":     luaTcellColorNewRgb,
		"new_hex":     luaTcellColorNewHex,
		"new_name":    luaTcellColorNewName,
		"new_palette": luaTcellColorNewPalette,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func LCheckTcellColor(L *lua.LState, index int) *tcell.Color {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*tcell.Color); ok {
		return v
	}

	L.ArgError(index, "value of type `TcellColor` expected")

	return nil
}

func LWrapTcellColor(L *lua.LState, data *tcell.Color) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaTcellColorTypeName))

	return ud
}

func LAddTcellColorToState(L *lua.LState, data *tcell.Color) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapTcellColor(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTcellColorNewRgb(L *lua.LState) int {
	r := L.CheckInt(1)
	g := L.CheckInt(2)
	b := L.CheckInt(3)

	color := tcell.NewRGBColor(int32(r), int32(g), int32(b))

	return LAddTcellColorToState(L, &color)
}

func luaTcellColorNewHex(L *lua.LState) int {
	hex := L.CheckInt64(1)
	color := tcell.NewHexColor(int32(hex))
	return LAddTcellColorToState(L, &color)
}

func luaTcellColorNewName(L *lua.LState) int {
	name := L.CheckString(1)
	color := tcell.GetColor(name)
	return LAddTcellColorToState(L, &color)
}

func luaTcellColorNewPalette(L *lua.LState) int {
	index := L.CheckInt(1)
	color := tcell.PaletteColor(index)
	return LAddTcellColorToState(L, &color)
}
