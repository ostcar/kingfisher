app [init_model, update_model, handle_request!, Model] {
    webserver: platform "../../platform/main.roc",
    json: "https://github.com/lukewilliamboswell/roc-json/releases/download/0.12.0/1trwx8sltQ-e9Y2rOB4LWUWLS_sFVyETK8Twl0i9qpw.tar.gz",
}

import webserver.Host
import json.Json

Model : {}

init_model = {}

update_model = |model, _event_list|
    Ok(model)

handle_request! = |_request, _model|
    when Host.get!("https://api.github.com/repos/ostcar/kingfisher/stargazers") is
        Ok(response_body) ->
            list : List {}
            list =
                response_body
                |> Decode.from_bytes(Json.utf8)
                |> Result.map_err(|_| ErrDecodingGitHubResponse)?

            Ok(
                {
                    body: list |> List.len() |> |len| "${len |> Num.to_str}\n" |> Str.to_utf8(),
                    headers: [],
                    status: 200,
                },
            )

        _ ->
            Ok(
                {
                    body: Str.to_utf8("Failed to fetch stars\n"),
                    headers: [],
                    status: 500,
                },
            )
