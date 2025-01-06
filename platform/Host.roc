hosted Host
    exposes [
        save_event!,
        stdout_line!,
        posix_time!,
    ]
    imports []

save_event! : List U8 => {}
stdout_line! : Str => {}
posix_time! : {} => U128 # TODO why is this a U128 but then getting converted to a I128 in Utc.roc?
