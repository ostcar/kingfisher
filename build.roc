app [main!] {
    cli: platform "https://github.com/roc-lang/basic-cli/releases/download/0.19.0/Hj-J_zxz7V9YurCSTFcFdu6cQJie4guzsPMUi5kBYUk.tar.br",
}

import cli.Cmd
import cli.Stdout

main! = |_args|
    build_for_surgical_linker!({})?
    build_for_legacy_linker!({})

build_for_surgical_linker! = |_|
    build_libapp_so!({})?
    build_dynhost!({})?
    preprocess!({})

build_libapp_so! = |_|
    Cmd.exec!("roc", ("build --lib examples/hello_world/main.roc --output host/libapp.so" |> Str.split_on(" ")))

build_dynhost! = |_|
    Cmd.new("go")
    |> Cmd.args(("build -C host -buildmode pie -o ../platform/dynhost" |> Str.split_on(" ")))
    |> Cmd.envs([("GOOS", "linux"), ("GOARCH", "amd64"), ("CC", "zig cc")])
    |> Cmd.status!
    |> Result.map_ok(|_| {})

preprocess! = |_|
    Cmd.exec!("roc", ("preprocess-host platform/dynhost platform/main.roc host/libapp.so" |> Str.split_on(" ")))

build_for_legacy_linker! = |_|
    [MacosArm64, MacosX64, LinuxArm64, LinuxX64]
    |> List.for_each_try!(|target| build_dot_a!(target))

build_dot_a! = |target|
    (goos, goarch, zig_target, prebuilt_binary) =
        when target is
            MacosArm64 -> ("darwin", "arm64", "aarch64-macos", "macos-arm64.a")
            MacosX64 -> ("darwin", "amd64", "x86_64-macos", "macos-x64.a")
            LinuxArm64 -> ("linux", "arm64", "aarch64-linux", "linux-arm64.a")
            LinuxX64 -> ("linux", "amd64", "x86_64-linux", "linux-x64.a")

    Stdout.line!("build host for ${Inspect.to_str(target)}")?
    Cmd.new("go")
    |> Cmd.envs([("GOOS", goos), ("GOARCH", goarch), ("CC", "zig cc -target ${zig_target}"), ("CGO_ENABLED", "1")])
    |> Cmd.args(("build -C host -buildmode c-archive -o ../platform/${prebuilt_binary} -tags legacy,netgo" |> Str.split_on(" ")))
    |> Cmd.status!
    |> Result.map_ok(|_| {})
