app [init_model, update_model, handle_request!, Model] {
    webserver: platform "../../platform/main.roc",
    json: "https://github.com/lukewilliamboswell/roc-json/releases/download/0.11.0/z45Wzc-J39TLNweQUoLw3IGZtkQiEN3lTBv3BXErRjQ.tar.br",
}

import webserver.Host
import json.Json

Model : {}

init_model = {}

update_model = \model, _event_list ->
    Ok model

handle_request! = \_request, _model ->
    when Host.get!("https://api.github.com/repos/ostcar/kingfisher/stargazers") is
        Ok response_body ->
            list : List {}
            list =
                response_body
                |> Decode.fromBytes(Json.utf8)
                |> Result.mapErr?(\_ -> ErrDecodingGitHubResponse)

            Ok {
                body: list |> List.len() |> \len -> "$(len |> Num.toStr)\n" |> Str.toUtf8(),
                headers: [],
                status: 200,
            }

        _ ->
            Ok {
                body: Str.toUtf8("Failed to fetch stars\n"),
                headers: [],
                status: 500,
            }
