//go:generate gen/docstring.sh

/*
lf is a terminal file manager.

Keys

Following commands along with the default keybindings are provided by lf.

    up                (default "k" and "<up>")
    half-up           (default "<c-u>")
    page-up           (default "<c-b>")
    down              (default "j" and "<down>")
    half-down         (default "<c-d>")
    page-down         (default "<c-f>")
    updir             (default "h" and "<left>")
    open              (default "l" and "<right>")
    quit              (default "q")
    bot               (default "G")
    top               (default "gg")
    read              (default ":")
    read-shell        (default "$")
    read-shell-wait   (default "!")
    read-shell-async  (default "&")
    search            (default "/")
    search-back       (default "?")
    toggle            (default "<space>")
    yank              (default "y")
    delete            (default "d")
    paste             (default "p")
    renew             (default "<c-l>")

Options

Following options can be used to customize the behavior of lf.

    hidden     bool    (default off)
    preview    bool    (default on)
    scrolloff  int     (default 0)
    tabstop    int     (default 8)
    ifs        string  (default not set)
    shell      string  (default $SHELL)
    showinfo   string  (default none)
    sortby     string  (default name)
    ratios     string  (default 1:2:3)

Variables

Following variables are exported for shell commands.

    $f   current file
    $fs  marked file(s) (separated with ':')
    $fx  current file or marked file(s) if any
*/
package main
