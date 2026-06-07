package main

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type time.Time

const LuaTimeTypeName = "time.Time"

func LRegisterTimeType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaTimeTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"now":            luaTimeNow,
		"new_unix":       luaTimeNewUnix,
		"new_unix_mili":  luaTimeNewUnixMili,
		"new_unix_micro": luaTimeNewUnixMicro,

		"since": luaTimeSince,
		"until": luaTimeUntil,

		"__eq": luaTimeMetaEq,
		"__lt": luaTimeMetaLt,
		"__le": luaTimeMetaLe,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"is_zero": luaTimeIsZero,
		"compare": luaTimeCompare,

		"date":     luaTimeDate,
		"year":     luaTimeYear,
		"month":    luaTimeMonth,
		"day":      luaTimeDay,
		"weekday":  luaTimeWeekday,
		"iso_week": luaTimeISOWeek,

		"clock":      luaTimeClock,
		"hour":       luaTimeHour,
		"minute":     luaTimeMinute,
		"second":     luaTimeSecond,
		"nanosecond": luaTimeNanosecond,

		"year_day": luaTimeYearDay,

		"add":      luaTimeAdd,
		"sub":      luaTimeSub,
		"add_date": luaTimeAddDate,

		"utc":              luaTimeUTC,
		"local":            luaLocal,
		"time_zone":        luaTimeZone,
		"time_zone_bounds": luaTimeZoneBounds,

		"unix":      luaTimeUnix,
		"unix_mili": luaTimeUnixMili,
		"unix_nano": luaTimeUnixNano,

		"format": luaTimeFormat,
	}))

	return mt
}

func LCheckTime(L *lua.LState, index int) *time.Time {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*time.Time); ok {
		return v
	}

	L.ArgError(index, "value of type `Time` expected")

	return nil
}

func LWrapTime(L *lua.LState, data *time.Time) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaTimeTypeName))

	return ud
}

func LAddTimeToState(L *lua.LState, data *time.Time) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapTime(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTimeNow(L *lua.LState) int {
	t := time.Now()
	return LAddTimeToState(L, &t)
}

func luaTimeNewUnix(L *lua.LState) int {
	sec := L.CheckInt64(1)
	nsec := L.CheckInt64(2)
	t := time.Unix(sec, nsec)
	return LAddTimeToState(L, &t)
}

func luaTimeNewUnixMili(L *lua.LState) int {
	millis := L.CheckInt64(1)
	t := time.UnixMilli(millis)
	return LAddTimeToState(L, &t)
}

func luaTimeNewUnixMicro(L *lua.LState) int {
	micro := L.CheckInt64(1)
	t := time.UnixMicro(micro)
	return LAddTimeToState(L, &t)
}

func luaTimeSince(L *lua.LState) int {
	t := LCheckTime(L, 1)
	return LAddDurationToState(L, time.Since(*t))
}

func luaTimeUntil(L *lua.LState) int {
	t := LCheckTime(L, 1)
	return LAddDurationToState(L, time.Until(*t))
}

func luaTimeMetaEq(L *lua.LState) int {
	self := LCheckTime(L, 1)
	other := LCheckTime(L, 2)
	L.Push(lua.LBool(self.Equal(*other)))
	return 1
}

func luaTimeMetaLt(L *lua.LState) int {
	self := LCheckTime(L, 1)
	other := LCheckTime(L, 2)
	L.Push(lua.LBool(self.Before(*other)))
	return 1
}

func luaTimeMetaLe(L *lua.LState) int {
	self := LCheckTime(L, 1)
	other := LCheckTime(L, 2)
	after := self.After(*other)
	L.Push(lua.LBool(!after))
	return 1
}

// ----------------------------------------------------------------------------

func luaTimeIsZero(L *lua.LState) int {
	t := LCheckTime(L, 1)
	L.Push(lua.LBool(t.IsZero()))
	return 1
}

func luaTimeCompare(L *lua.LState) int {
	self := LCheckTime(L, 1)
	other := LCheckTime(L, 2)
	L.Push(lua.LNumber(self.Compare(*other)))
	return 1
}

func luaTimeDate(L *lua.LState) int {
	t := LCheckTime(L, 1)
	year, month, day := t.Date()
	L.Push(lua.LNumber(year))
	LAddMonthToState(L, month)
	L.Push(lua.LNumber(day))
	return 3
}

func luaTimeYear(L *lua.LState) int {
	t := LCheckTime(L, 1)
	year := t.Year()
	L.Push(lua.LNumber(year))
	return 1
}

func luaTimeMonth(L *lua.LState) int {
	t := LCheckTime(L, 1)
	month := t.Month()
	return LAddMonthToState(L, month)
}

func luaTimeDay(L *lua.LState) int {
	t := LCheckTime(L, 1)
	day := t.Day()
	L.Push(lua.LNumber(day))
	return 1
}

func luaTimeWeekday(L *lua.LState) int {
	t := LCheckTime(L, 1)
	weekday := t.Weekday()
	return LAddWeekdayToState(L, weekday)
}

func luaTimeISOWeek(L *lua.LState) int {
	t := LCheckTime(L, 1)
	year, week := t.ISOWeek()
	L.Push(lua.LNumber(year))
	L.Push(lua.LNumber(week))
	return 2
}

func luaTimeClock(L *lua.LState) int {
	t := LCheckTime(L, 1)
	hour, min, sec := t.Clock()
	L.Push(lua.LNumber(hour))
	L.Push(lua.LNumber(min))
	L.Push(lua.LNumber(sec))
	return 3
}

func luaTimeHour(L *lua.LState) int {
	t := LCheckTime(L, 1)
	hour := t.Hour()
	L.Push(lua.LNumber(hour))
	return 1
}

func luaTimeMinute(L *lua.LState) int {
	t := LCheckTime(L, 1)
	minute := t.Minute()
	L.Push(lua.LNumber(minute))
	return 1
}

func luaTimeSecond(L *lua.LState) int {
	t := LCheckTime(L, 1)
	second := t.Second()
	L.Push(lua.LNumber(second))
	return 1
}

func luaTimeNanosecond(L *lua.LState) int {
	t := LCheckTime(L, 1)
	nanosecond := t.Nanosecond()
	L.Push(lua.LNumber(nanosecond))
	return 1
}

func luaTimeYearDay(L *lua.LState) int {
	t := LCheckTime(L, 1)
	yearDay := t.YearDay()
	L.Push(lua.LNumber(yearDay))
	return 1
}

func luaTimeAdd(L *lua.LState) int {
	t := LCheckTime(L, 1)
	dur := LCheckDuration(L, 2)
	result := t.Add(dur)
	return LAddTimeToState(L, &result)
}

func luaTimeSub(L *lua.LState) int {
	t := LCheckTime(L, 1)
	other := LCheckTime(L, 2)
	result := t.Sub(*other)
	return LAddDurationToState(L, result)
}

func luaTimeAddDate(L *lua.LState) int {
	t := LCheckTime(L, 1)
	years := L.CheckInt(2)
	months := L.CheckInt(3)
	days := L.CheckInt(4)

	result := t.AddDate(years, months, days)
	return LAddTimeToState(L, &result)
}

func luaTimeUTC(L *lua.LState) int {
	t := LCheckTime(L, 1)
	result := t.UTC()
	return LAddTimeToState(L, &result)
}

func luaLocal(L *lua.LState) int {
	t := LCheckTime(L, 1)
	result := t.Local()
	return LAddTimeToState(L, &result)
}

func luaTimeZone(L *lua.LState) int {
	t := LCheckTime(L, 1)
	name, offset := t.Zone()
	L.Push(lua.LString(name))
	L.Push(lua.LNumber(offset))
	return 2
}

func luaTimeZoneBounds(L *lua.LState) int {
	t := LCheckTime(L, 1)
	minOffset, maxOffset := t.ZoneBounds()
	LAddTimeToState(L, &minOffset)
	LAddTimeToState(L, &maxOffset)
	return 2
}

func luaTimeUnix(L *lua.LState) int {
	t := LCheckTime(L, 1)
	sec := t.Unix()
	L.Push(lua.LNumber(sec))
	return 1
}

func luaTimeUnixMili(L *lua.LState) int {
	t := LCheckTime(L, 1)
	millis := t.UnixMilli()
	L.Push(lua.LNumber(millis))
	return 1
}

func luaTimeUnixNano(L *lua.LState) int {
	t := LCheckTime(L, 1)
	nanos := t.UnixNano()
	L.Push(lua.LNumber(nanos))
	return 1
}

func luaTimeFormat(L *lua.LState) int {
	t := LCheckTime(L, 1)
	fmtStr := L.CheckString(2)
	L.Push(lua.LString(t.Format(fmtStr)))
	return 1
}

// ----------------------------------------------------------------------------
// type time.Month

const LuaMonthTypeName = "time.Month"

func LRegisterMonthType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaMonthTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"__tostring": luaMonthMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func LCheckMonth(L *lua.LState, index int) time.Month {
	value := L.Get(index)
	switch value.Type() {
	case lua.LTNumber:
		mo := time.Month(value.(lua.LNumber))
		return mo
	case lua.LTUserData:
		if v, ok := value.(*lua.LUserData).Value.(time.Month); ok {
			return v
		}
	}

	L.ArgError(index, "value of type `Month` expected")

	return 0
}

func LWrapMonth(L *lua.LState, data time.Month) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaMonthTypeName))

	return ud
}

func LAddMonthToState(L *lua.LState, data time.Month) int {
	ud := LWrapMonth(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaMonthMetaTostring(L *lua.LState) int {
	month := LCheckMonth(L, 1)
	L.Push(lua.LString(month.String()))
	return 1
}

// ----------------------------------------------------------------------------

const LuaWeekdayTypeName = "time.Weekday"

func LRegisterWeekdayType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaWeekdayTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"__tostring": luaWeekdayMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func LCheckWeekday(L *lua.LState, index int) time.Weekday {
	value := L.Get(index)
	switch value.Type() {
	case lua.LTNumber:
		wd := time.Weekday(value.(lua.LNumber))
		return wd
	case lua.LTUserData:
		if v, ok := value.(*lua.LUserData).Value.(time.Weekday); ok {
			return v
		}
	}

	L.ArgError(index, "value of type `Weekday` expected")

	return 0
}

func LWrapWeekday(L *lua.LState, data time.Weekday) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaWeekdayTypeName))

	return ud
}

func LAddWeekdayToState(L *lua.LState, data time.Weekday) int {
	ud := LWrapWeekday(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaWeekdayMetaTostring(L *lua.LState) int {
	weekday := LCheckWeekday(L, 1)
	L.Push(lua.LString(weekday.String()))
	return 1
}

// ----------------------------------------------------------------------------
// type time.Duration

const LuaDurationTypeName = "time.Duration"

func LRegisterDurationType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaDurationTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new":        luaDurationNew,
		"__mul":      luaDurationMetaMul,
		"__tostring": luaDurationMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"nanoseconds":  luaDurationNanoseconds,
		"microseconds": luadurationMicroseconds,
		"milliseconds": luaDurationMilliseconds,
		"seconds":      luaDurationSeconds,
		"minutes":      luaDurationMinutes,
		"hours":        luaDurationHours,
		"truncate":     luaDurationTruncate,
		"round":        luaDurationRound,
		"abs":          luaDurationAbs,
	}))

	return mt
}

func addDurationConstantToMt(L *lua.LState, tbl *lua.LTable) {
	tbl.RawSetString("Nanosecond", LWrapDuration(L, time.Nanosecond))
	tbl.RawSetString("Microsecond", LWrapDuration(L, time.Microsecond))
	tbl.RawSetString("Millisecond", LWrapDuration(L, time.Millisecond))
	tbl.RawSetString("Second", LWrapDuration(L, time.Second))
	tbl.RawSetString("Minute", LWrapDuration(L, time.Minute))
	tbl.RawSetString("Hour", LWrapDuration(L, time.Hour))
}

func LCheckDuration(L *lua.LState, index int) time.Duration {
	value := L.Get(index)
	switch value.Type() {
	case lua.LTNumber:
		dur := time.Duration(value.(lua.LNumber))
		return dur
	case lua.LTUserData:
		if v, ok := value.(*lua.LUserData).Value.(time.Duration); ok {
			return v
		}
	}

	L.ArgError(index, "value of type `Duration` expected")

	return 0
}

func LWrapDuration(L *lua.LState, data time.Duration) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaDurationTypeName))

	return ud
}

func LAddDurationToState(L *lua.LState, data time.Duration) int {
	ud := LWrapDuration(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaDurationNew(L *lua.LState) int {
	value := L.CheckNumber(1)
	dur := time.Duration(value)
	return LAddDurationToState(L, dur)
}

func luaDurationMetaMul(L *lua.LState) int {
	self := LCheckDuration(L, 1)
	other := LCheckDuration(L, 2)
	result := self * other
	return LAddDurationToState(L, result)
}

func luaDurationMetaTostring(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LString(dur.String()))
	return 1
}

// ----------------------------------------------------------------------------

func luaDurationNanoseconds(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Nanoseconds()))
	return 1
}

func luadurationMicroseconds(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Microseconds()))
	return 1
}

func luaDurationMilliseconds(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Milliseconds()))
	return 1
}

func luaDurationSeconds(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Seconds()))
	return 1
}

func luaDurationMinutes(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Minutes()))
	return 1
}

func luaDurationHours(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Hours()))
	return 1
}

func luaDurationTruncate(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	trunc := L.CheckNumber(2)
	result := dur.Truncate(time.Duration(trunc))
	return LAddDurationToState(L, result)
}

func luaDurationRound(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	round := L.CheckNumber(2)
	result := dur.Round(time.Duration(round))
	return LAddDurationToState(L, result)
}

func luaDurationAbs(L *lua.LState) int {
	dur := LCheckDuration(L, 1)
	return LAddDurationToState(L, dur.Abs())
}
