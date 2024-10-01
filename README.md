# Kingfisher

Kingfisher is a webserver platform for the [Roc
language](https://www.roc-lang.org/).

It lets you build websites by defining your own Model. The model is held in
memory and saved on disk.

Kingfisher follows a very minimalistic approach. You do not have to deal with
Tasks, SQL and I/O. Just write your views for read and write requests. Use a
model structure as you like and don't care about database tables and queries.


## How to use it

Use the platform with the following roc-application-header:

```roc
app [main, Model] {
    webserver: platform "https://github.com/ostcar/kingfisher/releases/download/v0.0.3/e8Mu5IplmOnXPU9VgpTCT6kyB463gX-SDC2nnMfAq7M.tar.br",
}
```

The platform requires a `Model`. The `Model` can be any valid roc type. For
example:

```roc
Model : {
    modelVersion: U64,
    users: List User,
    admin: [NoAdmin, Admin Str],
    books: List Str,
    specialOption: Bool,
}
```

The platform needs four functions:

```roc
Program : {
    decodeModel : [Init, Existing (List U8)] -> Model,
    encodeModel : Model -> List U8,
    handleReadRequest : Request, Model -> Response,
    handleWriteRequest : Request, Model -> (Response, Model),
}

main : Program
main = {
    decodeModel,
    encodeModel,
    handleReadRequest,
    handleWriteRequest,
}
```

* **decodeModel**: `decodeModel` is called when the server starts. On the first
  run, it is called with `Init` and on every other start with `Existing
  encodedModel` where `encodedModel` is a byte representation of the `Model`.
  The function has to create the initial model or decode the model from the
  given bytes.

* **encodeModel**: `encodeModel` is the counterpart of `decodeModel`. It has to
  create a byte representation of the `Model`.

`decodeModel` and `encodeModel` are used to create a snapshot of the `Model` and
persist it on disk. The functions can also be used to migrate an older version
of a model.

* **handleReadRequest**: `handleReadRequest` is called for HTTP requests, that
  are readonly and can not update the `Model`. The function is called with the
  `Request` and the `Model` and has to return a `Response`.

* **handleWriteRequest**: `handleWriteRequest` is called for HTTP requests, that
  can update the `Model`. The signature is simular then `handleReadRequest`, but
  it also returns a new `Model`.

The platform makes the distinction between a read and a write request on the
HTTP method. `POST`, `PUT`, `PATCH` and `DELETE` requests are write requests.
All other methods are read requests. This means, that a `GET` request can not
alter the Model.

The platform can handle many read requests at the same time. But there can only
be one concurent write request. When a write request is processed, all other
write request and all read requests have to wait.

When the server is shut down, it calls `encodeModel` and saves the snapshot to
disk. On restart, the snapshot is loaded to the server has the `Model` in memory
again.

Additionally each write request gets persisted on disk. On server failure, the
logged write requests are used to recreate the `Model` (the server just calls
the respective roc functions again).

After you have build your app, you can run your binary with the option `--help`
to see all available runtime options like listening address and path of the
snapshot file.


## Build the platform

To build the platform from source, you need to install
[roc](https://www.roc-lang.org/install), [go](https://go.dev/dl/) and
[zig](https://ziglang.org/learn/getting-started/#installing-zig). Zig is used to
crosscompile the go code. At the moment, it only works with zig `0.11.0`.

Run:

    roc run build.roc

to build the platform for linux, windows.

Afterwards, the example can be run with:

    roc run examples/hello_world/main.roc


## License

MIT
