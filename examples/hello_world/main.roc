app [init_model, update_model, handle_request!, Model] {
    webserver: platform "../../platform/main.roc",
}

import webserver.Http

Model : Str

init_model = "World"

update_model : Model, List (List U8) -> Result Model _
update_model = |model, event_list|
    event_list
    |> List.walk_try(
        model,
        |_acc_model, event|
            Str.from_utf8(event)
            |> Result.map_err(|_| InvalidEvent),
    )

handle_request! : Http.Request, Model => Result Http.Response _
handle_request! = |request, model|
    when request.method is
        Get ->
            Ok(
                {
                    body: Str.to_utf8("Hello ${model}\n"),
                    headers: [],
                    status: 200,
                },
            )

        Post(save_event!) ->
            event =
                if List.is_empty(request.body) then
                    Str.to_utf8("World")
                else
                    when Str.from_utf8(request.body) is
                        Ok(_) -> request.body
                        Err(_) ->
                            return Err(InvalidBody)

            save_event!(event)

            new_model = update_model(model, [event])?
            Ok(
                {
                    body: Str.to_utf8(new_model),
                    headers: [],
                    status: 200,
                },
            )

        _ ->
            Err(MethodNotAllowed(Http.method_to_str(request.method)))
