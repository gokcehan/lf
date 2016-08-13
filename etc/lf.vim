" use lf to select and open a file in vim
" adapted from the similar script for ranger
"
" you need to add something like the following to your ~/.vimrc:
"
" let lfvim = $GOPATH . "/src/github.com/gokcehan/lf/etc/lf.vim"
" if filereadable(lfvim)
"     exec "source " . lfvim
"     nnoremap <leader>l :<c-u>StartLF<cr>
" endif
"

function! StartLF()
    let temp = tempname()
    exec 'silent !lf -selection-path=' . shellescape(temp)
    if !filereadable(temp)
        redraw!
        return
    endif
    let names = readfile(temp)
    if empty(names)
        redraw!
        return
    endif
    exec 'edit ' . fnameescape(names[0])
    for name in names[1:]
        exec 'argadd ' . fnameescape(name)
    endfor
    redraw!
endfunction
command! -bar StartLF call StartLF()
