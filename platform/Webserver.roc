interface Webserver
    exposes [
        Request,
        Response,
        Header,
        RequestBody,
    ]
    imports []

# Request is the same as: https://github.com/roc-lang/basic-webserver/blob/main/platform/InternalHttp.roc
Request : {
    method : [Options, Get, Post, Put, Delete, Head, Trace, Connect, Patch],
    headers : List Header,
    url : Str,
    body : RequestBody,
    timeout : [TimeoutMilliseconds U64, NoTimeout],
}

Response : {
    status : U16,
    headers : List Header,
    body : List U8,
}

Header : { name : Str, value : Str }

RequestBody : [
    Body { mimeType : Str, body : List U8 },
    EmptyBody,
]
