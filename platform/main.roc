platform "webserver"
    requires { Model } { server : {
        initModel : Model,
        updateModel : List Event, Model -> Model,
        respond : Request, Model -> Task Response [ServerErr Str]_,
    } }
    exposes []
    packages {}
    imports [Webserver.{ Request, Response, Event }, Stderr]
    provides [mainForHost]

ProgramForHost : {
    initModel : Box Model,
    updateModel : List Event, Box Model -> Box Model,
    respond : Request, Box Model -> Task Response [],
}

mainForHost : ProgramForHost
mainForHost = {
    initModel,
    updateModel,
    respond,
}

initModel : Box Model
initModel =
    server.initModel
    |> Box.box

updateModel : List Event, Box Model -> Box Model
updateModel = \eventList, boxedModel ->
    server.updateModel eventList (Box.unbox boxedModel)
    |> Box.box

respond : Request, Box Model -> Task Response []
respond = \request, boxedModel ->
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
