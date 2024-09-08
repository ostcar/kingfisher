app [server, Model] {
    webserver: platform "../../platform/main.roc",
}

Model : Str

server = {
    updateModel,
    respond,
}

updateModel = \eventList, _initOrModel ->
    List.walk
        eventList
        (Ok "World")
        \_, event ->
            event
            |> Str.fromUtf8
            |> Result.mapErr \_ -> "invalid event"

respond = \request, model ->
    when request.method is
        Get ->
            Task.ok! {
                body: "Hello $(model)\n" |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        Post saveEvent ->
            newModel =
                if List.isEmpty request.body then
                    "World"
                else
                    request.body
                    |> Str.fromUtf8
                    |> Result.withDefault "invalid body"
            saveEvent (newModel |> Str.toUtf8)
                |> Task.mapErr! \_ -> ServerErr "Can not save event"
            Task.ok! {
                body: newModel |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        _ ->
            Task.err! (ServerErr "Unknown request method")

