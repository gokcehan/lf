# Autocompletion for nushell.
#
# Documentation: https://www.nushell.sh/book/externs.html

# To enable autocompletion you may put this file into a directory:
#
#     mkdir -p ~/.config/nushell/completions
#     ln -s "/path/to/lf.nu" ~/.config/nushell/completions
#
# Then you need to source this file in your nu config (Open the config with the
# command 'config nu' inside the nushell) by adding:
#
#     source ~/.config/nushell/completions/lf.nu

export extern "lf" [
  --command                   # command to execute on client initialization
  --config: string            # path to the config file (instead of the usual paths)
  --cpuprofile: string        # path to the file to write the CPU profile
  --doc                       # show documentation
  --last-dir-path: string     # path to the file to write the last dir on exit (to use for cd)
  --log: string               # path to the log file to write messages
  --memprofile: string        # path to the file to write the memory profile
  --print-last-dir            # print the last dir to stdout on exit (to use for cd)
  --print-selection           # print the selected files to stdout on open (to use as open file dialog)
  --remote: string            # send remote command to server
  --selection-path: string    # path to the file to write selected files on open (to use as open file dialog)
  --server                    # start server (automatic)
  --single                    # start a client without server
  --version                   # show version
  --help                      # show help
]
