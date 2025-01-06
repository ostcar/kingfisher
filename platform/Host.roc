hosted Host
    exposes [
        save_event!,
        stdout_line!,
        posix_time!,
        get!,
    ]
    imports []

save_event! : List U8 => {}
stdout_line! : Str => {}
posix_time! : {} => U128
get! : Str => Result (List U8) Str # TODO: Improve this
