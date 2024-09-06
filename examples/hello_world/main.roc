app [server, Model] {
    webserver: platform "../../platform/main.roc",
}

Model : Str

server = {
    initModel,
    updateModel,
    respond,
}

# TODO: I don't like initModel. Would it be possible to make the Model in updateModel optional and start with an empty event?
initModel = "World"

updateModel = \eventList, model ->
    List.walk
        eventList
        model
        \event, _ ->
            event

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
            saveEvent newModel
                |> Task.mapErr! \_ -> ServerErr "Can not save event"
            Task.ok! {
                body: newModel |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        _ ->
            Task.err! (ServerErr "Unknown request method")

