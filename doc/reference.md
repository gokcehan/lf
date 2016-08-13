# Reference

## Keys

    up                (default "k" and "<up>")
    down              (default "j" and "<down>")
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
    redraw            (default "<c-l>")

## Options

    preview    bool    (default on)
    hidden     bool    (default off)
    tabstop    int     (default 8)
    scrolloff  int     (default 0)
    sortby     string  (default name)
    showinfo   string  (default none)
    opener     string  (default xdg-open)
    ratios     string  (default 1:2:3)

## Variables

    $f   current file
    $fs  marked file(s) (seperated with ':')
    $fx  current file or marked file(s) if any
