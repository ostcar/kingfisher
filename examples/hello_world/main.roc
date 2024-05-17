app [main, Model] {
    webserver: platform "../../platform/main.roc",
}

import webserver.Webserver exposing [Request, Response]

Program : {
    decodeModel : [Init, Existing (List U8)] -> Result Model Str,
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

decodeModel : [Init, Existing (List U8)] -> Result Model Str
decodeModel = \fromPlatform ->
    when fromPlatform is
        Init ->
            Ok "World"

        Existing encoded ->
            encoded
            |> Str.fromUtf8
            |> Result.mapErr \_ -> "Error: Can not decode database."

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
            [] -> "World"
            _ -> request.body |> Str.fromUtf8 |> Result.withDefault "invalid body"
    (
        {
            body: model |> Str.toUtf8,
            headers: [],
            status: 200,
        },
        model,
    )
