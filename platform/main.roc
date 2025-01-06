platform "webserver"
    requires { Model } {
        init_model : Model,
        update_model : Model, List (List U8) -> Result Model _,
        handle_request! : Request, Model => Result Response _,
    }
    exposes []
    packages {}
    imports [Http.{ Request, Response, RequestFromHost, request_from_host }]
    provides [
        init_model_for_host,
        update_model_for_host,
        handle_request_for_host!,
    ]

init_model_for_host : Box Model
init_model_for_host = Box.box(init_model)

update_model_for_host : Box Model, List (List U8) -> Result (Box Model) Str
update_model_for_host = \boxed_model, event_list ->
    Box.unbox(boxed_model)
    |> update_model(event_list)
    |> Result.map(\model -> Box.box(model))
    |> errToStr()

handle_request_for_host! : RequestFromHost, Box Model => Result Response Str
handle_request_for_host! = \host_request, boxed_model ->
    Box.unbox(boxed_model)
    |> \model -> handle_request!(request_from_host host_request, model)
    |> errToStr()

errToStr : Result a _ -> Result a Str
errToStr = \r ->
    when r is
        Ok v -> Ok v
        Err ThisLineIsNecessary -> Err(Inspect.toStr(ThisLineIsNecessary)) # https://github.com/roc-lang/roc/issues/7289
        Err err -> Err(Inspect.toStr(err))
