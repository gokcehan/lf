package main

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// Type time.Time

const luaTimeTypeName = "time.Time"

func lRegisterTimeType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaTimeTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"now":            luaTimeNow,
		"new_unix":       luaTimeNewUnix,
		"new_unix_mili":  luaTimeNewUnixMili,
		"new_unix_micro": luaTimeNewUnixMicro,

		"since_time": luaTimeSince,
		"until_time": luaTimeUntil,

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
		"local_time":       luaLocal,
		"time_zone":        luaTimeZone,
		"time_zone_bounds": luaTimeZoneBounds,

		"to_unix":      luaTimeUnix,
		"to_unix_mili": luaTimeUnixMili,
		"to_unix_nano": luaTimeUnixNano,

		"format": luaTimeFormat,
	}))

	return mt
}

func lCheckTime(L *lua.LState, index int) *time.Time {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*time.Time); ok {
		return v
	}

	L.ArgError(index, "value of type `Time` expected")

	return nil
}

func lWrapTime(L *lua.LState, data *time.Time) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaTimeTypeName))

	return ud
}

func lAddTimeToState(L *lua.LState, data *time.Time) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapTime(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTimeNow(L *lua.LState) int {
	t := time.Now()
	return lAddTimeToState(L, &t)
}

func luaTimeNewUnix(L *lua.LState) int {
	sec := L.CheckInt64(1)
	nsec := L.CheckInt64(2)
	t := time.Unix(sec, nsec)
	return lAddTimeToState(L, &t)
}

func luaTimeNewUnixMili(L *lua.LState) int {
	millis := L.CheckInt64(1)
	t := time.UnixMilli(millis)
	return lAddTimeToState(L, &t)
}

func luaTimeNewUnixMicro(L *lua.LState) int {
	micro := L.CheckInt64(1)
	t := time.UnixMicro(micro)
	return lAddTimeToState(L, &t)
}

func luaTimeSince(L *lua.LState) int {
	t := lCheckTime(L, 1)
	return lAddDurationToState(L, time.Since(*t))
}

func luaTimeUntil(L *lua.LState) int {
	t := lCheckTime(L, 1)
	return lAddDurationToState(L, time.Until(*t))
}

func luaTimeMetaEq(L *lua.LState) int {
	self := lCheckTime(L, 1)
	other := lCheckTime(L, 2)
	L.Push(lua.LBool(self.Equal(*other)))
	return 1
}

func luaTimeMetaLt(L *lua.LState) int {
	self := lCheckTime(L, 1)
	other := lCheckTime(L, 2)
	L.Push(lua.LBool(self.Before(*other)))
	return 1
}

func luaTimeMetaLe(L *lua.LState) int {
	self := lCheckTime(L, 1)
	other := lCheckTime(L, 2)
	after := self.After(*other)
	L.Push(lua.LBool(!after))
	return 1
}

// ----------------------------------------------------------------------------

func luaTimeIsZero(L *lua.LState) int {
	t := lCheckTime(L, 1)
	L.Push(lua.LBool(t.IsZero()))
	return 1
}

func luaTimeCompare(L *lua.LState) int {
	self := lCheckTime(L, 1)
	other := lCheckTime(L, 2)
	L.Push(lua.LNumber(self.Compare(*other)))
	return 1
}

func luaTimeDate(L *lua.LState) int {
	t := lCheckTime(L, 1)
	year, month, day := t.Date()
	L.Push(lua.LNumber(year))
	lAddMonthToState(L, month)
	L.Push(lua.LNumber(day))
	return 3
}

func luaTimeYear(L *lua.LState) int {
	t := lCheckTime(L, 1)
	year := t.Year()
	L.Push(lua.LNumber(year))
	return 1
}

func luaTimeMonth(L *lua.LState) int {
	t := lCheckTime(L, 1)
	month := t.Month()
	return lAddMonthToState(L, month)
}

func luaTimeDay(L *lua.LState) int {
	t := lCheckTime(L, 1)
	day := t.Day()
	L.Push(lua.LNumber(day))
	return 1
}

func luaTimeWeekday(L *lua.LState) int {
	t := lCheckTime(L, 1)
	weekday := t.Weekday()
	return lAddWeekdayToState(L, weekday)
}

func luaTimeISOWeek(L *lua.LState) int {
	t := lCheckTime(L, 1)
	year, week := t.ISOWeek()
	L.Push(lua.LNumber(year))
	L.Push(lua.LNumber(week))
	return 2
}

func luaTimeClock(L *lua.LState) int {
	t := lCheckTime(L, 1)
	hour, min, sec := t.Clock()
	L.Push(lua.LNumber(hour))
	L.Push(lua.LNumber(min))
	L.Push(lua.LNumber(sec))
	return 3
}

func luaTimeHour(L *lua.LState) int {
	t := lCheckTime(L, 1)
	hour := t.Hour()
	L.Push(lua.LNumber(hour))
	return 1
}

func luaTimeMinute(L *lua.LState) int {
	t := lCheckTime(L, 1)
	minute := t.Minute()
	L.Push(lua.LNumber(minute))
	return 1
}

func luaTimeSecond(L *lua.LState) int {
	t := lCheckTime(L, 1)
	second := t.Second()
	L.Push(lua.LNumber(second))
	return 1
}

func luaTimeNanosecond(L *lua.LState) int {
	t := lCheckTime(L, 1)
	nanosecond := t.Nanosecond()
	L.Push(lua.LNumber(nanosecond))
	return 1
}

func luaTimeYearDay(L *lua.LState) int {
	t := lCheckTime(L, 1)
	yearDay := t.YearDay()
	L.Push(lua.LNumber(yearDay))
	return 1
}

func luaTimeAdd(L *lua.LState) int {
	t := lCheckTime(L, 1)
	dur := lCheckDuration(L, 2)
	result := t.Add(dur)
	return lAddTimeToState(L, &result)
}

func luaTimeSub(L *lua.LState) int {
	t := lCheckTime(L, 1)
	other := lCheckTime(L, 2)
	result := t.Sub(*other)
	return lAddDurationToState(L, result)
}

func luaTimeAddDate(L *lua.LState) int {
	t := lCheckTime(L, 1)
	years := L.CheckInt(2)
	months := L.CheckInt(3)
	days := L.CheckInt(4)

	result := t.AddDate(years, months, days)
	return lAddTimeToState(L, &result)
}

func luaTimeUTC(L *lua.LState) int {
	t := lCheckTime(L, 1)
	result := t.UTC()
	return lAddTimeToState(L, &result)
}

func luaLocal(L *lua.LState) int {
	t := lCheckTime(L, 1)
	result := t.Local()
	return lAddTimeToState(L, &result)
}

func luaTimeZone(L *lua.LState) int {
	t := lCheckTime(L, 1)
	name, offset := t.Zone()
	L.Push(lua.LString(name))
	L.Push(lua.LNumber(offset))
	return 2
}

func luaTimeZoneBounds(L *lua.LState) int {
	t := lCheckTime(L, 1)
	minOffset, maxOffset := t.ZoneBounds()
	lAddTimeToState(L, &minOffset)
	lAddTimeToState(L, &maxOffset)
	return 2
}

func luaTimeUnix(L *lua.LState) int {
	t := lCheckTime(L, 1)
	sec := t.Unix()
	L.Push(lua.LNumber(sec))
	return 1
}

func luaTimeUnixMili(L *lua.LState) int {
	t := lCheckTime(L, 1)
	millis := t.UnixMilli()
	L.Push(lua.LNumber(millis))
	return 1
}

func luaTimeUnixNano(L *lua.LState) int {
	t := lCheckTime(L, 1)
	nanos := t.UnixNano()
	L.Push(lua.LNumber(nanos))
	return 1
}

func luaTimeFormat(L *lua.LState) int {
	t := lCheckTime(L, 1)
	fmtStr := L.CheckString(2)
	L.Push(lua.LString(t.Format(fmtStr)))
	return 1
}

// ----------------------------------------------------------------------------
// type time.Month

const luaMonthTypeName = "time.Month"

func lRegisterMonthType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaMonthTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"__tostring": luaMonthMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func lCheckMonth(L *lua.LState, index int) time.Month {
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

func lWrapMonth(L *lua.LState, data time.Month) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaMonthTypeName))

	return ud
}

func lAddMonthToState(L *lua.LState, data time.Month) int {
	ud := lWrapMonth(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaMonthMetaTostring(L *lua.LState) int {
	month := lCheckMonth(L, 1)
	L.Push(lua.LString(month.String()))
	return 1
}

// ----------------------------------------------------------------------------

const luaWeekdayTypeName = "time.Weekday"

func lRegisterWeekdayType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaWeekdayTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"__tostring": luaWeekdayMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{}))

	return mt
}

func lCheckWeekday(L *lua.LState, index int) time.Weekday {
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

func lWrapWeekday(L *lua.LState, data time.Weekday) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaWeekdayTypeName))

	return ud
}

func lAddWeekdayToState(L *lua.LState, data time.Weekday) int {
	ud := lWrapWeekday(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaWeekdayMetaTostring(L *lua.LState) int {
	weekday := lCheckWeekday(L, 1)
	L.Push(lua.LString(weekday.String()))
	return 1
}

// ----------------------------------------------------------------------------
// type time.Duration

const luaDurationTypeName = "time.Duration"

func lRegisterDurationType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaDurationTypeName)

	addDurationConstantToMt(L, mt)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new": luaDurationNew,

		"__add": luaDurationMetaAdd,
		"__sub": luaDurationMetaSub,
		"__mul": luaDurationMetaMul,
		"__div": luaDurationMetaDiv,

		"__eq": luaDurationMetaEq,
		"__lt": luaDurationMetaLt,
		"__le": luaDurationMetaLe,

		"__tostring": luaDurationMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"add": luaDurationMetaAdd,
		"sub": luaDurationMetaSub,
		"mul": luaDurationMetaMul,
		"div": luaDurationMetaDiv,
		"eq":  luaDurationMetaEq,
		"lt":  luaDurationMetaLt,
		"le":  luaDurationMetaLe,

		"nanoseconds":  luaDurationNanoseconds,
		"microseconds": luadurationMicroseconds,
		"milliseconds": luaDurationMilliseconds,
		"seconds":      luaDurationSeconds,
		"minutes":      luaDurationMinutes,
		"hours":        luaDurationHours,
		"truncate":     luaDurationTruncate,
		"round":        luaDurationRound,
		"abs":          luaDurationAbs,

		"to_number": luaDurationToNumber,
	}))

	return mt
}

func addDurationConstantToMt(L *lua.LState, tbl *lua.LTable) {
	tbl.RawSetString("Nanosecond", lWrapDuration(L, time.Nanosecond))
	tbl.RawSetString("Microsecond", lWrapDuration(L, time.Microsecond))
	tbl.RawSetString("Millisecond", lWrapDuration(L, time.Millisecond))
	tbl.RawSetString("Second", lWrapDuration(L, time.Second))
	tbl.RawSetString("Minute", lWrapDuration(L, time.Minute))
	tbl.RawSetString("Hour", lWrapDuration(L, time.Hour))
}

func lCheckDuration(L *lua.LState, index int) time.Duration {
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

func lWrapDuration(L *lua.LState, data time.Duration) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaDurationTypeName))

	return ud
}

func lAddDurationToState(L *lua.LState, data time.Duration) int {
	ud := lWrapDuration(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaDurationNew(L *lua.LState) int {
	value := L.CheckNumber(1)
	dur := time.Duration(value)
	return lAddDurationToState(L, dur)
}

func luaDurationMetaAdd(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	result := self + other
	return lAddDurationToState(L, result)
}

func luaDurationMetaSub(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	result := self - other
	return lAddDurationToState(L, result)
}

func luaDurationMetaMul(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	result := self * other
	return lAddDurationToState(L, result)
}

func luaDurationMetaDiv(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	L.Push(lua.LNumber(float64(self) / float64(other)))
	return 1
}

func luaDurationMetaEq(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	L.Push(lua.LBool(self == other))
	return 1
}

func luaDurationMetaLt(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	L.Push(lua.LBool(self < other))
	return 1
}

func luaDurationMetaLe(L *lua.LState) int {
	self := lCheckDuration(L, 1)
	other := lCheckDuration(L, 2)
	L.Push(lua.LBool(self <= other))
	return 1
}

func luaDurationMetaTostring(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LString(dur.String()))
	return 1
}

// ----------------------------------------------------------------------------

func luaDurationNanoseconds(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Nanoseconds()))
	return 1
}

func luadurationMicroseconds(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Microseconds()))
	return 1
}

func luaDurationMilliseconds(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Milliseconds()))
	return 1
}

func luaDurationSeconds(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Seconds()))
	return 1
}

func luaDurationMinutes(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Minutes()))
	return 1
}

func luaDurationHours(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	L.Push(lua.LNumber(dur.Hours()))
	return 1
}

func luaDurationTruncate(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	m := lCheckDuration(L, 2)
	result := dur.Truncate(m)
	return lAddDurationToState(L, result)
}

func luaDurationRound(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	m := lCheckDuration(L, 2)
	result := dur.Round(m)
	return lAddDurationToState(L, result)
}

func luaDurationAbs(L *lua.LState) int {
	dur := lCheckDuration(L, 1)
	return lAddDurationToState(L, dur.Abs())
}

// luaDurationToNumber converts duration userdata to number value.
func luaDurationToNumber(L *lua.LState) int {
	duration := lCheckDuration(L, 1)
	L.Push(lua.LNumber(duration))
	return 1
}

// ----------------------------------------------------------------------------
// type time.Timer

const LuaTimerTypeName = "time.Timer"

func lRegisterTimerType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaTimerTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"reset": luaTimerReset,
		"stop":  luaTimerStop,
	}))

	return mt
}

func lCheckTimer(L *lua.LState, index int) *time.Timer {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*time.Timer); ok {
		return v
	}

	L.ArgError(index, "value of type `Timer` expected")

	return nil
}

func lWrapTimer(L *lua.LState, data *time.Timer) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaTimerTypeName))

	return ud
}

func lAddTimerToState(L *lua.LState, data *time.Timer) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapTimer(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaTimerReset(L *lua.LState) int {
	timer := lCheckTimer(L, 1)
	duration := lCheckDuration(L, 2)
	L.Push(lua.LBool(timer.Reset(duration)))
	return 1
}

func luaTimerStop(L *lua.LState) int {
	timer := lCheckTimer(L, 1)
	L.Push(lua.LBool(timer.Stop()))
	return 1
}
