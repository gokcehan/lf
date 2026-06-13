package main

import (
	"strings"

	"github.com/gdamore/tcell/v3"
	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// type tcell.Style

const luaTcellStyleTypeName = "tcell.Style"

func lRegisterTcellStyleType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaTcellStyleTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new":          luaTcellStyleNew,
		"reset_string": luaTcellStyleRestString,
		"__tostring":   luaTcellStyleMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"tostring": luaTcellStyleTostring,
		"wrap":     luaTcellStyleWrap,

		"foreground":         luaTcellStyleForeground,
		"background":         luaTcellStyleBackground,
		"foreground_rgb":     luaTcellStyleForegroundRGB,
		"background_rgb":     luaTcellStyleBackgroundRGB,
		"foreground_name":    luaTcellStyleForegroundName,
		"background_name":    luaTcellStyleBackgroundName,
		"foreground_palette": luaTcellStyleForegroundPalette,
		"background_palette": luaTcellStyleBackgroundPalette,

		"normal":         luaTcellStyleNormal,
		"bold":           luaTcellStyleBold,
		"blink":          luaTcellStyleBlink,
		"dim":            luaTcellStyleDim,
		"italic":         luaTcellStyleItalic,
		"reverse":        luaTcellStyleReverse,
		"strike_through": luaTcellStyleStrikeThrough,
		"underline":      luaTcellStyleUnderline,

		"has_bold":           luaTcellStyleHasBold,
		"has_blink":          luaTcellStyleHasBlink,
		"has_reverse":        luaTcellStyleHasReverse,
		"has_italic":         luaTcellStyleHasItalic,
		"has_dim":            luaTcellStyleHasDim,
		"has_strike_through": luaTcellStyleHasStrikeThrough,
		"has_underline":      luaTcellStyleHasUnderline,
	}))

	return mt
}

func lCheckTcellStyle(L *lua.LState, index int) *tcell.Style {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*tcell.Style); ok {
		return v
	}

	L.ArgError(index, "value of type `TcellStyle` expected")

	return nil
}

func lWrapTcellStyle(L *lua.LState, data *tcell.Style) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaTcellStyleTypeName))

	return ud
}

func lAddTcellStyleToState(L *lua.LState, data *tcell.Style) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapTcellStyle(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTcellStyleNew(L *lua.LState) int {
	st := tcell.StyleDefault
	return lAddTcellStyleToState(L, &st)
}

// luaTcellStyleRestString returns reset CSI string.
func luaTcellStyleRestString(L *lua.LState) int {
	L.Push(lua.LString("\033[0m"))
	return 1
}

func luaTcellStyleMetaTostring(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LString(tcellStyleToString(*st)))
	return 1
}

// ----------------------------------------------------------------------------

// luaTcellStyleTostring converts current style to CSI string. Does the same thing
// as __tostring meta method.
func luaTcellStyleTostring(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LString(tcellStyleToString(*st)))
	return 1
}

// luaTcellStyleWrap takes a list of content strings, and wrap them with CSI string
// form of current style and reset CSI sequens. Result is returned as a single
// string.
func luaTcellStyleWrap(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)

	nArgs := L.GetTop()
	contents := make([]string, nArgs+1)

	contents[0] = tcellStyleToString(*st)
	for i := 2; i <= nArgs; i++ {
		contents[i-1] = L.CheckString(i)
	}
	contents[nArgs] = "\033[0m"

	L.Push(lua.LString(strings.Join(contents, "")))
	return 1
}

func luaTcellStyleForeground(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	color := lCheckTcellColor(L, 2)
	*st = st.Foreground(*color)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleBackground(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	color := lCheckTcellColor(L, 2)
	*st = st.Background(*color)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleForegroundRGB(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	r := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	*st = st.Foreground(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleBackgroundRGB(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	r := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	*st = st.Background(tcell.NewRGBColor(int32(r), int32(g), int32(b)))

	return lAddTcellStyleToState(L, st)
}

// luaTcellStyleForegroundName sets foreground color with color name or hex code
// starting with `#`.
func luaTcellStyleForegroundName(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	name := L.CheckString(2)

	*st = st.Foreground(tcell.GetColor(name))

	return lAddTcellStyleToState(L, st)
}

// luaTcellStyleBackgroundName sets background color with color name or hex code
// starting with `#`.
func luaTcellStyleBackgroundName(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	name := L.CheckString(2)

	*st = st.Background(tcell.GetColor(name))

	return lAddTcellStyleToState(L, st)
}

// luaTcellStyleForegroundPalette sets foreground color with palette index.
func luaTcellStyleForegroundPalette(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	index := L.CheckInt(2)

	*st = st.Foreground(tcell.PaletteColor(index))

	return lAddTcellStyleToState(L, st)
}

// luaTcellStyleBackgroundPalette sets background color with palette index.
func luaTcellStyleBackgroundPalette(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	index := L.CheckInt(2)

	*st = st.Background(tcell.PaletteColor(index))

	return lAddTcellStyleToState(L, st)
}

// luaTcellStyleNormal returns the style with all attributes disabled.
// Colors and hyperlinks are preserved
func luaTcellStyleNormal(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	*st = st.Normal()
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleBold(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Bold(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleBlink(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Blink(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleDim(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Dim(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleItalic(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Italic(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleReverse(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Reverse(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleStrikeThrough(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.StrikeThrough(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleUnderline(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	isActive := L.CheckBool(2)
	*st = st.Underline(isActive)
	return lAddTcellStyleToState(L, st)
}

func luaTcellStyleHasBold(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasBold()))
	return 1
}

func luaTcellStyleHasBlink(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasBlink()))
	return 1
}

func luaTcellStyleHasReverse(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasReverse()))
	return 1
}

func luaTcellStyleHasItalic(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasItalic()))
	return 1
}

func luaTcellStyleHasDim(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasDim()))
	return 1
}

func luaTcellStyleHasStrikeThrough(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasStrikeThrough()))
	return 1
}

func luaTcellStyleHasUnderline(L *lua.LState) int {
	st := lCheckTcellStyle(L, 1)
	L.Push(lua.LBool(st.HasUnderline()))
	return 1
}

// ----------------------------------------------------------------------------
// type tcell.Color

const luaTcellColorTypeName = "tcell.Color"

func lRegisterTcellColorType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaTcellColorTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new_rgb":     luaTcellColorNewRgb,
		"new_hex":     luaTcellColorNewHex,
		"new_name":    luaTcellColorNewName,
		"new_palette": luaTcellColorNewPalette,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func lCheckTcellColor(L *lua.LState, index int) *tcell.Color {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*tcell.Color); ok {
		return v
	}

	L.ArgError(index, "value of type `TcellColor` expected")

	return nil
}

func lWrapTcellColor(L *lua.LState, data *tcell.Color) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaTcellColorTypeName))

	return ud
}

func lAddTcellColorToState(L *lua.LState, data *tcell.Color) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapTcellColor(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTcellColorNewRgb(L *lua.LState) int {
	r := L.CheckInt(1)
	g := L.CheckInt(2)
	b := L.CheckInt(3)

	color := tcell.NewRGBColor(int32(r), int32(g), int32(b))

	return lAddTcellColorToState(L, &color)
}

func luaTcellColorNewHex(L *lua.LState) int {
	hex := L.CheckInt64(1)
	color := tcell.NewHexColor(int32(hex))
	return lAddTcellColorToState(L, &color)
}

// luaTcellColorNewName creates a color with color name or hex code starting with
// `#`.
func luaTcellColorNewName(L *lua.LState) int {
	name := L.CheckString(1)
	color := tcell.GetColor(name)
	return lAddTcellColorToState(L, &color)
}

func luaTcellColorNewPalette(L *lua.LState) int {
	index := L.CheckInt(1)
	color := tcell.PaletteColor(index)
	return lAddTcellColorToState(L, &color)
}

// ----------------------------------------------------------------------------
// type tcell.Screen

const LuaTcellScreenTypeName = "tcell.Screen"

func lRegisterTcellScreenType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaTcellScreenTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func lCheckTcellScreen(L *lua.LState, index int) tcell.Screen {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(tcell.Screen); ok {
		return v
	}

	L.ArgError(index, "value of type `TcellScreen` expected")

	return nil
}

func lWrapTcellScreen(L *lua.LState, data tcell.Screen) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaTcellScreenTypeName))

	return ud
}

func lAddTcellScreenToState(L *lua.LState, data tcell.Screen) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapTcellScreen(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------
