# Harlock

Harlock is a small language with a focus on binary data manipulation and
data embedding in executables, mainly usable as a scripting language for 
binaries post-processing. It is based around the ideas discussed 
in the *Writing An Interpreter In Go* book by Thorsten Ball.

## Documentation

To get started on how to use harlock for your builds or as a standalone tools, please refer to the 
[harlock wiki](https://github.com/Abathargh/harlock/wiki)!

## Download

You can download the latest pre-built binaries for your architecture/os from 
the [harlock release page](https://github.com/Abathargh/harlock/releases).

If you have installed the go toolchain, you can directly install the interpeter using the following command:

```bash
go install github.com/Abathargh/harlock/cmd/harlock@latest
```

Changing the `latest` bit in the URL with the version you desire to download, if you do not want the latest one.

## Build

**Required: Go 1.18+**

Builds are executed using the linker flags to strip the debug symbols 
from the binaries, to achieve smaller sized executables.

```bash
make build    # build in place
make install  # build and install in $GOPATH/bin
```

## Usage

### Run a script

```bash
harlock script.hlk
```

### Start the REPL

```bash
harlock
Harlock v0.4.1 - amd64 on linux
>>> var hello = "Hello World!"
>>> print(hello)
Hello World!
```

### Embed a script

You can embed a harlock script into an executable together with the harlock runtime:
```bash
harlock -embed script.hlk
```

## License

Harlock is licensed under the terms of the MIT License.