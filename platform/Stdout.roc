module [
    line!,
]

import Host

## Write the given string to [standard output](https://en.wikipedia.org/wiki/Standard_streams#Standard_output_(stdout)),
## followed by a newline.
line! : Str => {}
line! = \str ->
    Host.stdout_line! str
