" Use lf to select and open file(s) in vim (adapted from ranger).
"
" You need to either copy the content of this file to your ~/.vimrc or source
" this file directly:
"
"     let lfvim = "/path/to/lf.vim"
"     if filereadable(lfvim)
"         exec "source " . lfvim
"     endif
"
" You may also like to assign a key to this command:
"
"     nnoremap <leader>l :LF<cr>
"

function! LF()
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
command! -bar LF call LF()
