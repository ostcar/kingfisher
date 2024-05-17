platform "webserver"
    requires { Model } { main : _ }
    exposes []
    packages {}
    imports [Webserver.{ Request, Response }]
    provides [mainForHost]

ProgramForHost : {
    decodeModel : [Init, Existing (List U8)] -> Result (Box Model) Str,
    encodeModel : Box Model -> List U8,
    handleReadRequest : Request, Box Model -> Response,
    handleWriteRequest : Request, Box Model -> (Response, Box Model),
}

mainForHost : ProgramForHost
mainForHost = {
    decodeModel,
    encodeModel,
    handleReadRequest,
    handleWriteRequest,
}

decodeModel : [Init, Existing (List U8)] -> Result (Box Model) Str
decodeModel = \fromHost ->
    main.decodeModel fromHost
    |> Result.map \model -> Box.box model

encodeModel : Box Model -> List U8
encodeModel = \boxedModel ->
    main.encodeModel (Box.unbox boxedModel)

handleReadRequest : Request, Box Model -> Response
handleReadRequest = \request, boxedModel ->
    main.handleReadRequest request (Box.unbox boxedModel)

handleWriteRequest : Request, Box Model -> (Response, Box Model)
handleWriteRequest = \request, boxedModel ->
    (resp, newModel) = main.handleWriteRequest request (Box.unbox boxedModel)
    (resp, newModel |> Box.box)
