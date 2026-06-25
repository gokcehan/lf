package main

import (
	"bufio"
	"io"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

// ----------------------------------------------------------------------------
// type exec.Cmd

const luaCmdTypeName = "exec.Cmd"

func lRegisterCmdType(L *lua.LState) *lua.LTable {
	mt := L.NewTypeMetatable(luaCmdTypeName)

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

		"set_stdout_writer":      luaCmdSetStdoutWriter,
		"set_stderr_writer":      luaCmdSetStderrWriter,
		"set_stdout_writer_func": luaCmdSetStdoutWriterFunc,
		"set_stderr_writer_func": luaCmdSetStderrWriterFunc,

		"kill": luaCmdKill,
	}))

	return mt
}

func lCheckCmd(L *lua.LState, index int) *exec.Cmd {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*exec.Cmd); ok {
		return v
	}

	L.ArgError(index, "value of type `Cmd` expected")

	return nil
}

func lWrapCmd(L *lua.LState, data *exec.Cmd) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = data

	L.SetMetatable(ud, L.GetTypeMetatable(luaCmdTypeName))

	return ud
}

func LAddCmdToState(L *lua.LState, data *exec.Cmd) int {
	if data == nil {
		L.Push(lua.LNil)
		return 1
	}

	ud := lWrapCmd(L, data)
	L.Push(ud)

	return 1
}

// ----------------------------------------------------------------------------

// luaCmdNew creates a new command object.
func luaCmdNew(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	st := 2
	nArgs := L.GetTop()
	args := make([]string, nArgs-st+1)
	for i := st; i <= nArgs; i++ {
		args[i-st] = L.Get(i).String()
	}

	cmd := exec.Command(cmdStr, args...)

	return LAddCmdToState(L, cmd)
}

func luaCmdMetaTostring(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	L.Push(lua.LString(cmd.String()))
	return 1
}

// ----------------------------------------------------------------------------

// luaCmdEnviron is a getter & setter for environment variable list of Cmd.
// When used as a getter, it returns a copy of command's environment variable
// list as table.
// Every environment variable is set in form of a `<key>=<value>` string.
func luaCmdEnviron(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	if L.GetTop() >= 2 {
		kvList := L.CheckTable(2)
		nElem := kvList.Len()
		env := make([]string, nElem)

		for i := 0; i < nElem; i++ {
			env[i] = kvList.RawGetInt(i + 1).String()
		}
		cmd.Env = env

		L.Push(kvList)

		return 1
	}

	env := cmd.Environ()

	envTable := L.NewTable()
	for _, kv := range env {
		envTable.Append(lua.LString(kv))
	}

	L.Push(envTable)

	return 1
}

// luaCmdAddEnviron appends new key-value string to command's environment variable
// list.
func luaCmdAddEnviron(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	kv := L.CheckString(2)

	cmd.Env = append(cmd.Env, kv)

	return 0
}

// luaCmdCombinedOutput runs command and returns string result containing both
// stdout and stderr output.
func luaCmdCombinedOutput(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	output, err := cmd.CombinedOutput()
	L.Push(lua.LString(string(output)))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// luaCmdOutput runs command and returns string result containing only stdout
// output.
func luaCmdOutput(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	output, err := cmd.Output()
	L.Push(lua.LString(string(output)))

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

// luaCmdRun runs current command.
func luaCmdRun(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	err := cmd.Run()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// luaCmdStart stars execution command. Caller should then calls `wait` method
// to wait for execution ends.
func luaCmdStart(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	err := cmd.Start()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// luaCmdWait blocks execution until execution of command ends.
func luaCmdWait(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	err := cmd.Wait()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// luaCmdStrerrPipe returns a reader handle to command's stderr output. This should
// be called before command starts execution.
func luaCmdStrerrPipe(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	reader := bufio.NewReader(stderrPipe)

	return lAddBufReaderToState(L, reader)
}

// luaCmdStdoutPipe returns a reader handle to command's stdout output. This should
// be called before command starts execution.
func luaCmdStdoutPipe(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	reader := bufio.NewReader(stdoutPipe)

	return lAddBufReaderToState(L, reader)
}

// luaCmdStdinPipe returns a writer handle to command's stdin input. This should
// be called before command starts execution.
func luaCmdStdinPipe(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	writer := bufio.NewWriter(stdinPipe)

	return lAddBufWriterToState(L, writer)
}

// luaCmdExitCode returns exit code of finished command. When exit code is not
// available, this method returns `nil`.
func luaCmdExitCode(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)

	if cmd.ProcessState == nil {
		return 0
	}

	exitCode := cmd.ProcessState.ExitCode()

	L.Push(lua.LNumber(exitCode))

	return 1
}

// luaCmdSetStdoutWriter sets a writer value for command stdout.
func luaCmdSetStdoutWriter(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	ud := L.CheckUserData(2)

	writer, ok := ud.Value.(io.Writer)
	if !ok {
		L.ArgError(2, "is not a writer")
	}

	cmd.Stdout = writer

	return 0
}

// luaCmdSetStderrWriter sets a writer value for command stdout.
func luaCmdSetStderrWriter(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	ud := L.CheckUserData(2)

	writer, ok := ud.Value.(io.Writer)
	if !ok {
		L.ArgError(2, "is not a writer")
	}

	cmd.Stderr = writer

	return 0
}

// luaCmdSetStdoutWriterFunc sets a Lua function as writer used for command stdout.
func luaCmdSetStdoutWriterFunc(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	fn := L.CheckFunction(2)

	writer := &luaFuncWriter{
		luaState: L,
		fn:       fn,
	}

	cmd.Stdout = writer

	return 0
}

// luaCmdSetStderrWriterFunc sets a Lua function as writer used for command stderr.
func luaCmdSetStderrWriterFunc(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	fn := L.CheckFunction(2)

	writer := &luaFuncWriter{
		luaState: L,
		fn:       fn,
	}

	cmd.Stderr = writer

	return 0
}

func luaCmdKill(L *lua.LState) int {
	cmd := lCheckCmd(L, 1)
	err := cmd.Process.Kill()

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}
