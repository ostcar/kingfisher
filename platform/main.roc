platform "webserver"
    requires { Model } {
        init_model : Model,
        handle_events : Model, List (List U8) -> Result Model _,
        handle_request! : Request, Model => Result Response _,
    }
    exposes []
    packages {}
    imports [Http.{ Request, Response, RequestFromHost, request_from_host }]
    provides [
        handle_events_for_host,
        handle_request_for_host!,
        init_model_for_host,
    ]

init_model_for_host : Box Model
init_model_for_host = init_model |> Box.box

handle_events_for_host : Box Model, List (List U8) -> Result (Box Model) Str
handle_events_for_host = \boxed_model, event_list ->
    boxed_model
    |> Box.unbox
    |> handle_events event_list
    |> Result.map \model -> model |> Box.box
    |> Result.mapErr Inspect.toStr

handle_request_for_host! : RequestFromHost, Box Model => Result Response Str
handle_request_for_host! = \host_request, boxed_model ->
    boxed_model
    |> Box.unbox
    |> \model -> handle_request! (request_from_host host_request) model
    |> Result.mapErr Inspect.toStr
