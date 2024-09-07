module [
    Request,
    Response,
    Header,
    Event,
    requestFromHost,
    HostRequest,
    RequestMethod,
]

import PlatformTasks

# TODO: Event has to be (List U8). Also make sure on the host, that newline is allowed
Event : Str
SaveEvent : Event -> Task {} Str

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

Header : { name : Str, value : Str }

HostRequestMethod : [
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

HostRequest : {
    method : HostRequestMethod,
    headers : List Header,
    url : Str,
    mimeType : Str,
    body : List U8,
    timeout : [TimeoutMilliseconds U64, NoTimeout],
}

requestFromHost : HostRequest -> Request
requestFromHost = \fromHost ->
    method =
        when fromHost.method is
            Post -> Post PlatformTasks.saveEvent
            Put -> Put PlatformTasks.saveEvent
            Delete -> Delete PlatformTasks.saveEvent
            Patch -> Patch PlatformTasks.saveEvent
            Options -> Options
            Get -> Get
            Head -> Head
            Trace -> Trace
            Connect -> Connect
    {
        headers: fromHost.headers,
        url: fromHost.url,
        mimeType: fromHost.mimeType,
        body: fromHost.body,
        timeout: fromHost.timeout,
        method: method,
    }
