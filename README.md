# Kingfisher

Kingfisher is a webserver platform for the [Roc
language](https://www.roc-lang.org/).

It lets you build websites by defining your own Model. The model is held in
memory and saved on disk.

Kingfisher lets you save your data in memory. No need for SQL. The data is saved in an event store.


## How to use it

Use the platform with the following roc-application-header:

```roc
app [init_model, update_model, handle_request!, Model] {
    webserver: platform "TODO update me after a release",
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

The platform needs three functions:

```roc
init_model: Model
update_model : Model, List (List U8) -> Result Model _
handle_request! : Http.Request, Model => Result Http.Response _
```

* **init_model**: `init_model` has to return the first version of the model. Be
careful, when changing this. Your events need to be compatible.

* **update_model**: `update_model` runs the events. Be careful, when changing
the implementation. Old events can not be migrated.

* **handle_request!**: `handle_request!` is called for each HTTP requests.
The function is called with the `Request` and the `Model` and has to return
a `Result Response _`.

* **save_event!**: To change the model, you have to pattern match on the request
method. Write requests (`POST`, `PUT`, `PATCH` and `DELETE`) have an function as
payload this function can be used to save an event.

```roc
when request.method is
    Get ->
        ...

    Post save_event! ->
        save_event! my_event
```

An event is a `List U8`. Make sure to implement `update_model` to handle all of
your events.

The platform makes the distinction between a read and a write request on the
HTTP method. `POST`, `PUT`, `PATCH` and `DELETE` requests are write requests.
All other methods are read requests. This means, that a `GET` request can not
alter the Model.

The platform can handle many read requests at the same time. But there can only
be one concurent write request. When a write request is processed, all other
write request and all read requests have to wait.

All events are only persisted to disk, if `handle_request!` returns with Ok. On
error, all events gets discarded.


## Build the platform

To build the platform from source, you need to install
[roc](https://www.roc-lang.org/install), [go](https://go.dev/dl/) and
[zig](https://ziglang.org/learn/getting-started/#installing-zig). Zig is used to
crosscompile the go code. At the moment, it only works with zig `0.11.0`.

Run:

    roc build.roc

to build the platform for linux and mac.

Afterwards, the example can be run with:

    roc run examples/hello_world/main.roc


## License

MIT
