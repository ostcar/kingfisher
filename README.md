# Kingfisher

Kingfisher is a webserver platform for the [Roc
language](https://www.roc-lang.org/).

It lets you build websites, by defining your own Model. The model is hold in
memory and saved on disk.


## How to use it

Use the platform, with the following roc-application-header:

```roc
app "hello_world"
    packages {
        webserver: "https://github.com/ostcar/kingfisher/releases/download/v0.0.1/DyE2hmSORHg6McbXh2T92yUjL7edUi2m6ZjiC2ypqfQ.tar.br",
    }
    imports [webserver.Webserver.{ Request, Response }]
    provides [main, Model] to webserver

```

The platform requires a `Model`. The `Model`, can be any valid roc type. For example

```
Model : {
    modelVersion: U64,
    users: List User,
    admin: [NoAdmin, Admin Str],
}
```

The platform needs four functions:

```
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
  encodedModel` where `encodedModel` is an representation of the `Model`. The
  function has to create the initial model or decode the model from the data.
* **encodeModel**: `encodeModel` is the counterpart of `decodeModel`. It has to
  create a byte representation of the `Model`.

`decodeModel` and `encodeModel` are used to create a snapshot of the `Model`,
and persist it on disk. The functions can also be use to migrate an older
version of a model.

* **handleReadRequest**: `handleReadRequest` is called for HTTP requests, that
  can not update the `Model`. The function is called with the `Request` and the
  `Model` and has to return a `Response`.
* **handleWriteRequest**: `handleWriteRequest` is called for HTTP request, that
  can update the `Model`. The signature is simular then `handleReadRequest`, but
  it also returns a new `Model`.

The platform makes the distinction between a read and a write request on the
HTTP method. `Post`, `PUT`, `PATCH` and `DELETE` requests are write requests.
All other are read requests. This means, that a `GET` request can not alter the
Model.

The platform can handle many read requests at the same time. But there can only
be one concurent write request. When a write request is processed, all other
write request and all read requests have to wait.

Each write request gets persisted on disk. On server failer, the logged write
requests are used, to recreate the `Model`.


## Build the platform

The easiest way to build the platform is with [Task](https://taskfile.dev/).

Run:
```
task preprocess
```

to preprocess the platform.

Afterwards, the example can be run with:

```
roc run --prebuilt-platform examples/hello_world/main.roc
```
