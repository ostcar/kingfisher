# Kingfisher

A webserver platform for the [Roc language](https://www.roc-lang.org/).


## How to build it

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


## Create the bundle

```
roc build --bundle .tar.br platform/main.roc
```
