# Harlock

Harlock is a small language with a focus on binary data manipulation and
data embedding in executables, mainly usable as a scripting language for 
binaries post-processing. It is based around the ideas discussed 
in the *Writing An Interpreter In Go* book by Thorsten Ball.

# Download

```bash
go install github.com/Abathargh/harlock/cmd/harlock@latest
```
# Build 

Builds are created by using the linker flags to strip the debug symbols, 
to achieve smaller binaries.

```bash
make          # build in place
make install  # build and install in $GOPATH/bin
```

# License

The interpreter is licensed under the terms of the MIT License.