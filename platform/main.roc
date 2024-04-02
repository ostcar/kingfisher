platform "webserver"
    requires { Model } { main : _ }
    exposes []
    packages {}
    imports [Webserver.{ Event, Request, Response }]
    provides [mainForHost]

ProgramForHost : {
    init : Box Model,
    handleReadRequest : Request, Box Model -> Response,
    handleWriteRequest : Request, Box Model -> (Response, Box Model),
}

mainForHost : ProgramForHost
mainForHost = {
    init,
    handleReadRequest,
    handleWriteRequest,
}

init : Box Model
init =
    main.init
    |> Box.box

handleReadRequest : Request, Box Model -> Response
handleReadRequest = \request, boxedModel ->
    main.handleReadRequest request (Box.unbox boxedModel)

handleWriteRequest : Request, Box Model -> (Response, Box Model)
handleWriteRequest = \request, boxedModel ->
    (resp, newModel) = main.handleWriteRequest request (Box.unbox boxedModel)
    (resp, newModel |> Box.box)
