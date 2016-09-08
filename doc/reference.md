# Reference

## Keys

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

## Options

    hidden     bool    (default off)
    preview    bool    (default on)
    scrolloff  int     (default 0)
    tabstop    int     (default 8)
    ifs        string  (default not set)
    shell      string  (default $SHELL)
    showinfo   string  (default none)
    sortby     string  (default name)
    ratios     string  (default 1:2:3)

## Variables

    $f   current file
    $fs  marked file(s) (seperated with ':')
    $fx  current file or marked file(s) if any
