module [
    line!,
]

import Host

line! : Str => {}
line! = |str|
    Host.stdout_line!(str)
