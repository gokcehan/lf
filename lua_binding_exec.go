package main

import (
	"bufio"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// type exec.Cmd

const LuaCmdTypeName = "exec.Cmd"

func LRegisterCmdType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(LuaCmdTypeName)

	L.SetFuncs(mt, map[string]lua.LGFunction{
		"new":        luaCmdNew,
		"__tostring": luaCmdMetaTostring,
	})
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"environ":     luaCmdEnviron,
		"add_environ": luaCmdAddEnviron,

		"combined_output": luaCmdCombinedOutput,
		"output":          luaCmdOutput,
		"run":             luaCmdRun,

		"start": luaCmdStart,
		"wait":  luaCmdWait,

		"stderr_pipe": luaCmdStrerrPipe,
		"stdout_pipe": luaCmdStdoutPipe,
		"stdin_pipe":  luaCmdStdinPipe,

		"exit_code": luaCmdExitCode,
	}))

	return mt
}

func LCheckCmd(L *lua.LState, index int) *exec.Cmd {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*exec.Cmd); ok {
		return v
	}

	L.ArgError(index, "value of type `Cmd` expected")

	return nil
}

func LWrapCmd(L *lua.LState, data *exec.Cmd) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(LuaCmdTypeName))

	return ud
}

func LAddCmdToState(L *lua.LState, data *exec.Cmd) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := LWrapCmd(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

func luaCmdNew(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.CheckAny(i).String()
	}

	cmd := exec.Command(cmdStr, args...)

	return LAddCmdToState(L, cmd)
}

func luaCmdMetaTostring(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)
	L.Push(lua.LString(cmd.String()))
	return 1
}

// ----------------------------------------------------------------------------

func luaCmdEnviron(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	env := cmd.Environ()

	envTable := L.NewTable()
	for _, kv := range env {
		envTable.Append(lua.LString(kv))
	}

	L.Push(envTable)

	return 1
}

func luaCmdAddEnviron(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)
	kv := L.CheckString(2)

	cmd.Env = append(cmd.Env, kv)

	return 0
}

func luaCmdCombinedOutput(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	output, err := cmd.CombinedOutput()
	L.Push(lua.LString(string(output)))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

func luaCmdOutput(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	output, err := cmd.Output()
	L.Push(lua.LString(string(output)))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

func luaCmdRun(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	err := cmd.Run()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaCmdStart(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	err := cmd.Start()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaCmdWait(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	err := cmd.Wait()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func luaCmdStrerrPipe(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	reader := bufio.NewReader(stderrPipe)

	return LAddBufReaderToState(L, reader)
}

func luaCmdStdoutPipe(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	reader := bufio.NewReader(stdoutPipe)

	return LAddBufReaderToState(L, reader)
}

func luaCmdStdinPipe(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	writer := bufio.NewWriter(stdinPipe)

	return LAddBufWriterToState(L, writer)
}

func luaCmdExitCode(L *lua.LState) int {
	cmd := LCheckCmd(L, 1)

	if cmd.ProcessState == nil {
		return 0
	}

	exitCode := cmd.ProcessState.ExitCode()

	L.Push(lua.LNumber(exitCode))

	return 1
}
