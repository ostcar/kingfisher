# Hello World

It uses a s string as model. The initial string is "World"

On a GET request, the application returns "Hello World" where "World" is
repleaced by the current model.

On a POST request, the current model is replaced by the body. With an empy Body,
the model is reset to "World".

For example:

```sh
$ curl localhost:9000
Hello World

$ curl localhost:9000 -d "Kingfisher"
Kingfisher

$ curl localhost:9000
Hello Kingfisher

$ curl localhost:9000 -X POST
World

$ curl localhost:9000
Hello World
```
