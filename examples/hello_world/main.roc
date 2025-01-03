app [init_model, handle_events, handle_request!, Model] {
    webserver: platform "../../platform/main.roc",
}

import webserver.Http

Model : Str

init_model = "World"

handle_events : Model, List (List U8) -> Result Model _
handle_events = \model, event_list ->
    event_list
    |> List.walkTry model \_acc_model, event ->
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

            new_model = handle_events? model [event]
            Ok {
                body: new_model |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        _ ->
            Err MethodNotAllowed
