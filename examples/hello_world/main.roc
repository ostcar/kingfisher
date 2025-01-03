app [handle_events, handle_request!, Model] {
    webserver: platform "../../platform/main.roc",
}

import webserver.Http

Model : Str

handle_events : [Init, Existing Model], List (List U8) -> Result Model _
handle_events = \may_model, event_list ->
    init_model =
        when may_model is
            Init -> "World"
            Existing existing_model -> existing_model

    event_list
    |> List.walkTry init_model \_, event ->
        event
        |> Str.fromUtf8
        |> Result.mapErr \_ -> InvalidEvent

handle_request! : Http.Request, Model => Result Http.Response _
handle_request! = \request, model ->
    when request.method is
        Get ->
            Ok {
                body: "Hello $(model)\n" |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        Post save_event! ->
            event =
                if List.isEmpty request.body then
                    "World" |> Str.toUtf8
                else
                    when request.body |> Str.fromUtf8 is
                        Ok _ -> request.body
                        Err _ ->
                            return Err InvalidBody

            save_event! event

            new_model = handle_events? (Existing model) [event]
            Ok {
                body: new_model |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        _ ->
            Err MethodNotAllowed
