# Autocompletion for nushell.
# 
# Documentation: https://www.nushell.sh/book/externs.html

# Either add this to your existing completions or copy this to
# your nu config file (Use 'config nu' in the nushell to open it).

module completions {
 export extern "lf" [
  --command                                     # command to execute on client initialization
  --config                                      # path to the config file (instead of the usual paths)
  --cpuprofile                                  # path to the file to write the CPU profile
  --doc                                         # show documentation
  --last-dir-path                               # path to the file to write the last dir on exit (to use for cd)
  --log                                         # path to the log file to write messages
  --memprofile                                  # path to the file to write the memory profile
  --remote                                      # send remote command to server
  --selection-path                              # path to the file to write selected files on open (to use as open file dialog)
  --server                                      # start server (automatic)
  --single                                      # start a client without server
  --version                                     # show version
  --help                                        # show help
 ]
}
use completions *