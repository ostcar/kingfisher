platform "webserver"
    requires { Model } { server : {
        updateModel : List Event, [Init, Existing Model] -> Result Model Str,
        respond : Request, Model -> Task Response [ServerErr Str]_,
    } }
    exposes []
    packages {}
    imports [Webserver.{ Request, Response, Event }, Stderr]
    provides [mainForHost]

ServerForHost : {
    updateModel : List Event, [Init, Existing (Box Model)] -> Result (Box Model) Str,
    respond : Webserver.HostRequest, Box Model -> Task Response [],
}

mainForHost : ServerForHost
mainForHost = {
    updateModel,
    respond,
}

updateModel : List Event, [Init, Existing (Box Model)] -> Result (Box Model) Str
updateModel = \eventList, maybeBoxedModel ->
    maybeModel =
        when maybeBoxedModel is
            Init -> Init
            Existing boxedModel -> Existing (Box.unbox boxedModel)

    server.updateModel eventList maybeModel
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
