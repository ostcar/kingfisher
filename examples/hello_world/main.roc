app [server, Model] {
    webserver: platform "../../platform/main.roc",
    json: "https://github.com/lukewilliamboswell/roc-json/releases/download/0.10.2/FH4N0Sw-JSFXJfG3j54VEDPtXOoN-6I9v_IA8S18IGk.tar.br",
}

import json.Json

Model : Str

server = {
    updateModel,
    respond,
}

updateModel = \eventList, _initOrModel ->
    initModel = "World"

    List.walk
        eventList
        (Ok initModel)
        \_, (eventType, eventPayload) ->
            when eventType is
                "update-name" ->
                    nameEvent : Result Str _
                    nameEvent = Decode.fromBytes eventPayload Json.utf8

                    userEvent
                    |> Result.map \payload ->
                        payload
                    |> Result.mapErr \_ -> "Can not encode update-name payload"

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

            Encode.toBytes newModel
                |> saveEvent "update-name"
                |> Task.mapErr! \_ -> ServerErr "Can not save event"
            Task.ok! {
                body: newModel |> Str.toUtf8,
                headers: [],
                status: 200,
            }

        _ ->
            Task.err! (ServerErr "Unknown request method")

