# Lua Plugin Support

Lua plugin support implementation is now written in following files:

- `lua_plugin.go`, functions for handling Lua state management, calling Lua functions.
- `lua_module_*.go` provides functionalities exposed to Lua as module.
- `lua_binding_*.go` defines data types exposed to Lua as user data and meta table

Following variables are provided as Lua global variable:

- `app`, user data pointing to app object.
- `lf_type`, a table containing all metatable of exported types.

## Execution Structure

Lua state gets initialized right before user config files are sourced, so that
sorting methods, commands, etc. registered in Lua plugins can be used in user
config file.

When initializing, program will look for `plugins` ditectory under parent directory
of config file (if no plugin directory path or config file path is specified from
command line), all subdirectory under `plugins` directories that contains `init.lua`
in them will be considered a plugin. And those `init.lua` will be used as plugin
entrance.

Lua states, are single threaded Lua interpreter state machines. To execute Lua
script, two different groups of Lua state objects are used:

- A single Lua state object that runs all scripts requiring synchronous execution.
- A Lua state pool, makes more Lua states when asked by different goroutine,
makes running Lua code concurrently possible.

Different Lua states share no data, so if a function depends on a variable data
stored on Lua state, then one must run this funtion in synchronous mode to make
sure this function gets executed on the one and only synchronous Lua state.

Lua states are encapsulate in a global variable `gLuaPool` defined in `lua_plugin.go`.
Lua state used for code execution can be acquired from this object.

Whenever a new Lua state is instanciated, plugin script evaluation will be run
on this Lua state to make sure it has the same initial state as others. And this
process is also required, since all data yielded during execution are local to
this Lua state.

## Plugin Function Registration

Every plugin entrance script must return a table for registering function to lf.
And each value in this table must also be a table. Let's call the value a registry
table, and key-value paris of registry tables are registry entries.

Not all, but many of supported registry tables provide messages, keys in those
registry tables would be used as message names.

The structure looks like this:

```lua
-- plugins/foo/init.lua
return {
    registry_key = {
        message_name = message_entry,
    },
}
```

Basicaly, message entry takes one of three forms:

- plain value, which can be used directly by lf.
- function, gets called by lf, to provide extension to the program, let's call
  this a message action function.
- table, when meta data is required, or there is more than one action associated
  with one message name, a table is used to represent message entry.

Those data tables returned by plugin scripts will be stored in a global struct
called `gLuaRegistry`. Data tables are associated with the Lua state they belongs
to.

```go
var gLuaRegistry struct {
    // first level indexing: pointer to data tables's owner state
    // second level indexing: path to script that returns the table
	stateDataMap map[*lua.LState]map[string]*lua.LTable

    // ...
}
```

### Calling Message Action

The registry table structure makes distributing tasks amoung Lua states possible.

When calling a message, lf first acquires a Lua state for execution, and retrives
data table map of this state. Then, tries to locate a message action with 4 components:

- Source name, this is the path to plugin entrance script that provides this message.
- Registry key, used for fetching registry table from data table returned by
  that plugin scripts.
- Message name.
- Variant name, when a message entry uses table as value, and contains multiple
  action in it. lf will fetch action function from that table with variant name
  as key.

  Supported variant name varies depending on registry type.

There are serval functions provided for calling Lua message in `lua_plugin.go`,
such as `callLuaMsg`.

Lua state used for execution is not determined when calling `callLuaMsg`, but
creating Lua value arguments needed for message action call often requires access
to Lua state object. Hence a function with signature `func(L *lua.LState) []lua.LValue`
is passed in, so that message action arguments can be generated after Lua state
is acquired.

### Variant Handling

Some message entry support defining multiple actions. For example, a command entry
can have both an main action and a completion action.

A message action extractor is a function that takes a message entry and returns
a Lua function pointer as message action.

Each message variant has a extractor defined for it. Those extractors returns
action function found in message entry when it's defined as a table. And returns
default action function when message entry is defined as a non-table value or
such variant cannot be found.

### Asynchronous Message Action

For now, all message actions are synchronous by default, when message action
needs to be marked asynchronous, a message entry table with `is_async` set to true
is used.

For example:

```lua
return {
    command = {
        foo = {
            action = function()
                app:ui():echo("bar")
            end,
            is_async = true,
        },
    },
}
```

There are some Lua APIs require synchronous state to call, when those APIs being
called under asynchronous mode, a Lua side error will be raised.

## Supported Registry Keys

> All entry types, returns types in this section are Lua types, some of them are
  exposed from Go to Lua via binding.

Currently, following keys are supported:

- `command`, is a message table, adds new command to lf.

  Entry type: `string`, `function`, `table`

  Message variant:

  - `action`, a function to run when this command gets called, can be a string or a function
  - `completion`, a function that returns a list of `CompMatch` and matched string
- `event_hook`, is a message table, adds callback function for lf event like `on-init`, `on-load`, ...

  Entry type: `function`, `table`

  Message variant:

  - `action`, a function to run when event happens.
- `key_map`, defines new key maps. It's registry table looks like this:

  ```lua
  return {
      key_map = {
          n = {
              ["<c-f>"] = {
                  action = function()
                      app:ui:echo("hello")
                  end,
                  is_async = true,
              },
          }
      }
  }
  ```

  Registry entry in `key_map` registry table uses key map mode as key (`n`, `v` and `c`),
  and its value defines action of different keys under this mode.

  Here, key map actions are defined just like message entries.

  Key map entry type: `function`, `table`

  Variant:

  - `action`, a function to run when key map is triggered.
- `local_option`, each key in this registry table is a directory paths,
  corresponding value is table containing options for this directory. All values
  in option table are string, and allowed keys are option names allowed by `setlocal`
  command.
- `misc`, is a message table, provide some extension to lf.

  Entry type: `function`, `table`

  Message variant provided by each message may vary depending on how lf use them,
  but all of them has a `action` variant as main message action.

  Currently supported messages are:

  - `shell`, takes shell command and argument list, and makes an `exec.Cmd` from
    them.
- `option`, is a table of string keys and string values. Keys are lf option names,
  and its value will be set to corresponding lf option.
- `previewer`, is a message table, adds new preview action.

  Entry type: `function`, `table`

  Message variant:

  - `action`, a function that takes a data writer and preview arguments, display
    preview content by writing data to the writer.
  - `clean`, cleaner function for this previewer.
  - `condition`, a function that takes the path of target file, and returns a boolean
    value for indicating whether this previewer is active for that file.
- `sorting_method`, is a message table, adds new sorting method to lf.

  Entry type: `function`, `table`

  Variant:

  - `action`, a function takes a list of `File`, and returns sorted list of files.
- `ui_formatter`, is a message table, but allowed message names are predefined.

  Provides formatter function for different UI element in place of `fmt` options
  that with formatting verbs in them.

  Entry type: `function`, `table`

  Variant:

  - `action`, a function, its argument and return type differs by UI element type.

  Currently supported keys are:

  - `cursoractive`, formatting file entry under cursor in active directory window.
  - `cursorparent`, formatting file entry under cursor in parent directory window.
  - `cursorpreview`, formatting file entry under cursor in preview directory window.
  - `dupfile`, generates name for duplicated files during copy/move operation.
  - `error`, formatting error message.
  - `numbercursor`, formatting directory line number under corsor.
  - `number`, formatting directory line number.
  - `ruler`, returns ruler content
  - `prompt`, returns prompt content
  - `tag`: formatting tags
- `ui_style`

  This registry table provides style values for `fmt` options that do not allow
  formatting verbs in their value. namely:

  - borderfmt
  - copyfmt
  - cutfmt
  - menufmt
  - menuheaderfmt
  - menuselectfmt
  - selectfmt
  - visual

  Registry entry keys in this registry table are names of those options but without
  that `fmt` suffix.

  Entry type: `Style`, `fun(): Style`

## Type Binding

Some types are exposed to Lua via binding. Bindings are written in `lua_binding_*.go`,
types are grouped by the module they belong to.

Some of them are listed below:

- `lua_binding_bufio.go`, provides reader and writer for exchanging data between
  lf, Lua and possible subprocess spawned in Lua script.
- `lua_binding_exec.go`, allows Lua to spawn subprocess.
- `lua_binding_main.go`, expose lf types like `app`, `nav`, `ui` ...
- `lua_binding_tcell.go`, expose `tcell.Style` type as a tool for writing
  `ui_formatter` and setting `ui_stylel`.

  One can build CSI styled string with builder style calls.

  ```lua
  {
      ui_formatter = {
          tag = function(tag)
              if tag == "-" then
                  return Style.new():foreground_name("yellow"):background_name("gray"):wrap(tag)
              end
              return Style:new():foreground_name("red"):wrap(tag)
          end,
      }
  }
  ```

## Modules

Modules are written in `lua_module_*.go`, they are exposed to Lua as preload modules.

Modules can be accessed in Lua via `require` call:

```lua
local lf = require "lf"
```

- `lua_module_fs.go`: exposed as module `lf.fs` file and filepath operation.
- `lua_module_main.go`: exposed as module `lf`, provides API for accessing lf
  functionalities, and miscellaneous helper functions.
- `lua_module_ui.go`: exposed as module `lf.ui` functions about drwaing UI.
- `lua_module_utf8.go`: exposed as module `lf.utf8`, helper function for dealing with UTF-8 strings.

  Lua strings are plain byte blobs with no predefined structure. This module is
  required if one wants to handle UTF-8 runes.
