module [
    Request,
    Response,
    Header,
    Event,
]

Event : Str
SaveEvent : Event -> Task {} [SaveFailed]

RequestMethod : [
    Options,
    Get,
    Post SaveEvent,
    Put SaveEvent,
    Delete SaveEvent,
    Head,
    Trace,
    Connect,
    Patch SaveEvent,
]

Request : {
    method : RequestMethod,
    headers : List Header,
    url : Str,
    mimeType : Str,
    body : List U8,
    timeout : [TimeoutMilliseconds U64, NoTimeout],
}

Response : {
    status : U16,
    headers : List Header,
    body : List U8,
}

Header : { name : Str, value : List U8 }
