app [main!] {
    cli: platform "https://github.com/roc-lang/basic-cli/releases/download/0.18.0/0APbwVN1_p1mJ96tXjaoiUCr8NBGamr8G8Ac_DrXR-o.tar.br",
}

import cli.Cmd
import cli.Stdout

main! = \_args ->
    build_for_surgical_linker!? {}
    build_for_legacy_linker! {}

build_for_surgical_linker! = \_ ->
    build_libapp_so!? {}
    build_dynhost!? {}
    preprocess! {}

build_libapp_so! = \_ ->
    Cmd.exec! "roc" ("build --lib examples/hello_world/main.roc --output host/libapp.so" |> Str.splitOn " ")

build_dynhost! = \_ ->
    Cmd.new "go"
    |> Cmd.args ("build -C host -buildmode pie -o ../platform/dynhost" |> Str.splitOn " ")
    |> Cmd.envs [("GOOS", "linux"), ("GOARCH", "amd64"), ("CC", "zig cc")]
    |> Cmd.status!
    |> Result.map \_ -> {}

preprocess! = \_ ->
    Cmd.exec! "roc" ("preprocess-host platform/dynhost platform/main.roc host/libapp.so" |> Str.splitOn " ")

build_for_legacy_linker! = \_ ->
    [MacosArm64, MacosX64, LinuxArm64, LinuxX64]
    |> List.forEachTry! \target -> build_dot_a! target

build_dot_a! = \target ->
    (goos, goarch, zigTarget, prebuiltBinary) =
        when target is
            MacosArm64 -> ("darwin", "arm64", "aarch64-macos", "macos-arm64.a")
            MacosX64 -> ("darwin", "amd64", "x86_64-macos", "macos-x64.a")
            LinuxArm64 -> ("linux", "arm64", "aarch64-linux", "linux-arm64.a")
            LinuxX64 -> ("linux", "amd64", " x86_64-linux", "linux-x64.a")
            WindowsArm64 -> ("windows", "arm64", "aarch64-windows", "windows-arm64.obj")
            WindowsX64 -> ("windows", "amd64", "x86_64-windows", "windows-x64.obj")
    Stdout.line!? "build host for $(Inspect.toStr target)"
    Cmd.new "go"
    |> Cmd.envs [("GOOS", goos), ("GOARCH", goarch), ("CC", "zig cc -target $(zigTarget)"), ("CGO_ENABLED", "1")]
    |> Cmd.args ("build -C host -buildmode c-archive -o ../platform/$(prebuiltBinary) -tags legacy,netgo" |> Str.splitOn " ")
    |> Cmd.status!
    |> Result.map \_ -> {}
