# Contributing

Code contributions are always welcomed in lf.

If you are going to introduce a new feature, it is best to open an issue first for discussion. If your feature can be implemented as a configuration option **please add it to the [wiki](https://github.com/gokcehan/lf/wiki)**.

For bug fixes, you can simply send a pull request.

## Code conventions

In addition to `gofmt` and friends (e.g. [go vet](https://pkg.go.dev/cmd/vet), [staticcheck](https://staticcheck.dev/), [golangci-lint](https://golangci-lint.run/)), we have a few conventions:

- Global variables are best avoided except when they are not.
Global variable names are prefixed with `g` as in `gFooBar`.
Exceptions are variables holding values of environmental variables which are prefixed with `env` as in `envFooBar` and regular expressions which are prefixed with `re` as in `reFooBar` when they are global.
- Type and function names are small case as in `fooBar` since we don't use exporting.
- For file name variables, `name`, `fname`, or `filename` should refer to the base name of the file as in `baz.txt`, and `path`, `fpath`, or `filepath` should refer to the full path of the file as in `/foo/bar/baz.txt`.
- Run `go fmt && go generate` using Go 1.19+. On earlier versions of Go, `go generate` creates needless whitespace differences.

Use the surrounding code as reference when in doubt as usual.

## Adding a new option

Adding a new option usually requires the following steps:

- Add option name/type to `gOpts` struct in `opts.go`
- Add default option value to `init` function in `opts.go`
- Add option evaluation logic to `setExpr.eval` in `eval.go`
- Implement the option somewhere in the code
- Add option name to `gOptWords` in `complete.go` for tab completion
- Add option name and its default value to `Quick Reference` and `Options` sections in `doc.go`
- Run `go generate` to update the documentation
- Commit your changes and send a pull request

## Adding a new builtin command

Adding a new command usually requires the following steps:

- Add default key if any to `init` function in `opts.go`
- Add command evaluation logic to `callExpr.eval` in `eval.go`
- Implement the command somewhere in the code
- Add command name to `gCmdWords` in `complete.go` for tab completion
- Add command name to `Quick Reference` and `Commands` sections in `doc.go`
- Run `go generate` to update the documentation
- Commit your changes and send a pull request

## Platform specific code

There are two files named `os.go` and `os_windows.go` for unix and windows specific code respectively.
If you add something to either of these files but not the other, you probably break the build for the other platform.
If your addition works the same in both platforms, your addition probably belongs to `main.go` instead.

## Make changes to the documentation

There are three files that contain the documentation in various formats:

- `doc.go` is the source of the [online documentation](https://pkg.go.dev/github.com/gokcehan/lf)
- `docstring.go` is used when running `lf -doc`
- `lf.1` is a man page which is commonly available on Unix-like operating systems (`man lf`)

You should only make changes to `doc.go`, the other files are automatically generated **and should not be edited manually**.
Run `go fmt && go generate` to ensure that the code is formatted and files are generated correctly.
