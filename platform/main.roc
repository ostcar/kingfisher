platform "webserver"
    requires { Model } {
        handle_events : [Init, Existing Model], List (List U8) -> Result Model _,
        handle_request! : Request, Model => Result Response _,
    }
    exposes []
    packages {}
    imports [Http.{ Request, Response, RequestFromHost, request_from_host }]
    provides [handle_events_for_host, handle_request_for_host!]

handle_events_for_host : [Init, Existing (Box Model)], List (List U8) -> Result (Box Model) Str
handle_events_for_host = \may_boxed_model, event_list ->
    may_model =
        when may_boxed_model is
            Init -> Init
            Existing boxed_model -> boxed_model |> Box.unbox |> Existing

    handle_events may_model event_list
    |> Result.map \model -> model |> Box.box
    |> Result.mapErr Inspect.toStr

handle_request_for_host! : RequestFromHost, Box Model => Result Response Str
handle_request_for_host! = \host_request, boxed_model ->
    boxed_model
    |> Box.unbox
    |> \model -> handle_request! (request_from_host host_request) model
    |> Result.mapErr Inspect.toStr
