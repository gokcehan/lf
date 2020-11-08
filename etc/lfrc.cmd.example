# interpreter for shell commands
set shell cmd

# Shell commands with multiline definitions and/or positional arguments and/or
# quotes do not work in Windows. For anything but the simplest shell commands,
# it is recommended to create separate script files and simply call them here
# in commands or mappings.

# change the editor used in default editor keybinding
# There is no builtin terminal editor installed in Windows. The default editor
# mapping uses 'notepad' which launches in a separate GUI window. You may
# instead install a terminal editor of your choice and replace the default
# editor keybinding accordingly.
map e $vim %f%

# change the pager used in default pager keybinding
# The standard pager used in Windows is 'more' which is not a very capable
# pager. You may instead install a pager of your choice and replace the default
# pager keybinding accordingly.
map i $less %f%

# change the shell used in default shell keybinding
map w $powershell

# change 'doc' command to use a different pager
cmd doc $lf -doc | less

# leave some space at the top and the bottom of the screen
set scrolloff 10

# use enter for shell commands
map <enter> shell
