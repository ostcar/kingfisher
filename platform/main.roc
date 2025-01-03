platform "webserver"
    requires { Model } { server : {
        updateModel : List Event, [Init, Existing Model] -> Result Model Str,
        respond : Request, Model -> Task Response _,
    } }
    exposes []
    packages {
        json: "https://github.com/lukewilliamboswell/roc-json/releases/download/0.10.2/FH4N0Sw-JSFXJfG3j54VEDPtXOoN-6I9v_IA8S18IGk.tar.br",
    }
    imports [Webserver.{ Request, Response }, Stderr, json.Json, Utc]
    provides [mainForHost]

ServerForHost : {
    updateModel : List FromHostEvent, [Init, Existing (Box Model)] -> Result (Box Model) Str,
    respond : Webserver.HostRequest, Box Model -> Task Response [],
}

FromHostEvent : {
    type : Str,
    timestamp : I64,
    payload : List U8,
}

Event : (Str, val) where val implements Decoding

mainForHost : ServerForHost
mainForHost = {
    updateModel,
    respond,
}

EventTypes : [Str]

updateModel : List FromHostEvent, [Init, Existing (Box Model)] -> Result (Box Model) Str
updateModel = \hostEventList, maybeBoxedModel ->
    maybeModel =
        when maybeBoxedModel is
            Init -> Init
            Existing boxedModel -> Existing (Box.unbox boxedModel)

    hostEventList
    |> List.map \hostEvent ->
        utc = hostEvent.timestamp |> Num.toI128 |> Utf.fromMillisSinceEpoch
        {
            type: outer.type,
            timestamp: utc,
            payload: outer.payload,
        }
    |> Result.try \eventList -> server.updateModel eventList maybeModel
    |> Result.map Box.box

respond : Webserver.HostRequest, Box Model -> Task Response []
respond = \hostRequest, boxedModel ->
    request = Webserver.requestFromHost hostRequest

    when server.respond request (Box.unbox boxedModel) |> Task.result! is
        Ok response -> Task.ok response
        Err (ServerErr msg) ->
            Stderr.line msg
                |> Task.onErr! \_ -> crash "unable to write to stderr"

            # returns a http server error response
            Task.ok {
                status: 500,
                headers: [],
                body: [],
            }

        Err err ->
            """
            Server error:
                $(Inspect.toStr err)

            Tip: If you do not want to see this error, use `Task.mapErr` to handle the error.
            Docs for `Task.mapErr`: <https://www.roc-lang.org/builtins/Task#mapErr>
            """
                |> Stderr.line
                |> Task.onErr! \_ -> crash "unable to write to stderr"

            Task.ok {
                status: 500,
                headers: [],
                body: [],
            }
