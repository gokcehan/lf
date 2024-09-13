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

let s:temp = tempname()
if executable('lf')
  command! -nargs=? -bar -complete=dir FilePicker call FilePicker('lf', '-selection-path', s:temp, <q-args>)
elseif executable('ranger')
  " The option --choosefiles was added in ranger 1.5.1.
  " Use --choosefile with ranger 1.4.2 through 1.5.0 instead.
  command! -nargs=? -bar -complete=dir FilePicker call FilePicker('ranger', '--choosefiles='..s:temp, '--selectfile', <q-args>)
elseif executable('nnn')
  command! -nargs=? -bar -complete=dir FilePicker call FilePicker('nnn', '-p', s:temp, <q-args>)
endif

if exists(':FilePicker') == 2
  function! FilePicker(...)
    let path = a:000[-1]
    let cmd = a:000[:-2] + (empty(path) ?
          \ [filereadable(expand('%')) ? expand('%:p') : '.'] : [path])
    if has('nvim')
      enew
      call termopen(cmd, { 'on_exit': function('s:open') })
    else
      if has('gui_running')
        if has('terminal')
          call term_start(cmd, {'exit_cb': function('s:term_close'), 'curwin': 1})
        else
          echomsg 'GUI is running but terminal is not supported.'
        endif
      else
        exec 'silent !'..join(cmd) | call s:open()
      endif
    endif
  endfunction

  if has('gui_running') && has('terminal')
    function! s:term_close(job_id, event)
      if a:event == 'exit'
        bwipeout!
        call s:open()
      endif
    endfunction
  endif

  function! s:open(...)
    if !filereadable(s:temp)
      " if &buftype ==# 'terminal'
      "   bwipeout!
      " endif
      redraw!
      " Nothing to read.
      return
    endif
    let names = readfile(s:temp)
    if empty(names)
      redraw!
      " Nothing to open.
      return
    endif
    " Edit the first item.
    exec 'edit' fnameescape(names[0])
    " Add any remaning items to the arg list/buffer list.
    for name in names[1:]
      exec 'argadd' fnameescape(name)
    endfor
    redraw!
  endfunction
endif
