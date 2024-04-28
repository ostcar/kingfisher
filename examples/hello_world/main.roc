app "hello_world"
    packages {
        webserver: "../../platform/main.roc",
    }
    imports [webserver.Webserver.{ Request, Response }]
    provides [main, Model] to webserver

Program : {
    decodeModel : [Init, Existing (List U8)] -> Model,
    encodeModel : Model -> List U8,
    handleReadRequest : Request, Model -> Response,
    handleWriteRequest : Request, Model -> (Response, Model),
}

Model : Str

main : Program
main = {
    decodeModel,
    encodeModel,
    handleReadRequest,
    handleWriteRequest,
}

decodeModel : [Init, Existing (List U8)] -> Model
decodeModel = \fromPlatform ->
    when fromPlatform is
        Init ->
            "World"

        Existing encoded ->
            when encoded |> Str.fromUtf8 is
                Ok model -> model
                Err _ -> crash "Error: Can not decode database."

encodeModel : Model -> List U8
encodeModel = \model ->
    model |> Str.toUtf8

handleReadRequest : Request, Model -> Response
handleReadRequest = \_request, model -> {
    body: "Hello $(model)\n" |> Str.toUtf8,
    headers: [],
    status: 200,
}

handleWriteRequest : Request, Model -> (Response, Model)
handleWriteRequest = \request, _model ->
    model =
        when request.body is
            EmptyBody -> "World"
            Body b -> b.body |> Str.fromUtf8 |> Result.withDefault "invalid body"
    (
        {
            body: model |> Str.toUtf8,
            headers: [],
            status: 200,
        },
        model,
    )
