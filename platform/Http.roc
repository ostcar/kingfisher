module [
    Request,
    Response,
    Header,
    RequestFromHost,
    request_from_host,
]

import Host

SaveEvent : List U8 => {}

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

Header : { name : Str, value : Str }

Request : {
    method : RequestMethod,
    headers : List Header,
    url : Str,
    body : List U8,
    timeout : [TimeoutMilliseconds U64, NoTimeout],
}

Response : {
    status : U16,
    headers : List Header,
    body : List U8,
}

MethodFromHost : [
    Options,
    Get,
    Post,
    Put,
    Delete,
    Head,
    Trace,
    Connect,
    Patch,
]

RequestFromHost : {
    method : MethodFromHost,
    headers : List Header,
    url : Str,
    body : List U8,
    timeout : [TimeoutMilliseconds U64, NoTimeout],
}

request_from_host : RequestFromHost -> Request
request_from_host = \from_host ->
    method =
        when from_host.method is
            Post -> Post Host.save_event!
            Put -> Put Host.save_event!
            Delete -> Delete Host.save_event!
            Patch -> Patch Host.save_event!
            Options -> Options
            Get -> Get
            Head -> Head
            Trace -> Trace
            Connect -> Connect
    {
        headers: from_host.headers,
        url: from_host.url,
        body: from_host.body,
        timeout: from_host.timeout,
        method: method,
    }
