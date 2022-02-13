# Harlock

Harlock is a small language with a focus on binary data manipulation and
data embedding in executables, mainly usable as a scripting language for 
binaries post-processing. It is based around the ideas discussed 
in the *Writing An Interpreter In Go* book by Thorsten Ball.

- [Harlock](#harlock)
  - [Download](#download)
  - [Build](#build)
  - [Usage](#usage)
    - [Run a script](#run-a-script)
    - [Start the REPL](#start-the-repl)
    - [Embed a script](#embed-a-script)
- [Language primitives & operations](#language-primitives--operations)
  - [License](#license)

## Download

```bash
go install github.com/Abathargh/harlock/cmd/harlock@latest
```
## Build 

Builds are executed using the linker flags to strip the debug symbols 
from the binaries, to achieve smaller sized executables.

```bash
make          # build in place
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
Harlock v0.0.0 - amd64 on windows
>>> var hello = "Hello World!"
>>> print(hello)
Hello World!
```

### Embed a script

You can embed a harlock script into an executable together with the harlock runtime:
```bash
harlock -embed script.hlk
```

# Language primitives & operations

```
# Values and variables
1                           # integers
"string test"               # strings
true                        # booleans
[1, 2, 10]                  # arrays
{key2: val2, key1: val1}    # maps
var ex = 12                 # variables
// this is a comment        # comments

# Integer values and operations
var normal = 12
var hex = 0xff
10 + 1
10 - 1
10 * 2
10 / 2
10 % 2
10 == 10
10 != 11
10 < 2
10 > 2
10 <= 2
10 >= 2
0xfe & 0x01
0xfe | 0x01
0xfe << 1
0xfe >> 1
~0xff
hex(12)

# Booleans
var t = true
var f = false
true && false
true || false
true == false
true != false

# Strings
var s = "test string"
var s = 'test string'
s == "test string"
s != "test strings"
len(s) == 11
hex("FF0102") == [255, 1, 2]

# Arrays
var arr = [1, 2, 3]
arr[0] == 1
arr.push(4)
arr.pop()
arr.slice(0, 1)
contains(arr, 1)

# Maps
var m = {"test": value, 1: 23}
len(m) == 2
m["test"] = value
m.set("new", 12)
m.pop("new")
contains(m, "test")

# Sets
var s = set(1, 3, 5, 4)
var s2 = set([1, 2, 4, 3, 4])
s.add(10)
s.remove(3)
contains(s, 1)

# Conditional expressions
if x < y { ret x } else { ret y }
if a * (2 - b) > c {
    ret "ok"
}

# Functions
var f = fun(x) { ret x }
var arr = [f, fun(){}]
var iter_func = fun(arr, func) {
    if len(arr) == 0 { ret }
    func(arr[0])
    if len(arr) == 1 { ret }
    iter_func(slice(arr, 1, len(arr)), func)
}

iter_func(arr, print)
print("test")

# Hex files
var h = open("test.hex", "hex")
print(h)
h.size()
h.record(0) 
h.write_at(0x1000, hex("DEADBEEF"))
h.read_at(0x1000, 4)
```


## License

The interpreter is licensed under the terms of the MIT License.