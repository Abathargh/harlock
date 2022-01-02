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

Builds are executed using the linker flags to strip the debug symbols 
from the binaries, to achieve smaller sized executables.

```bash
make          # build in place
make install  # build and install in $GOPATH/bin
```

# Usage

```
# Values and variables
1                           # integers
"string test"               # strings
true                        # booleans
[1, 2, 10]                  # arrays
{key2: val2, key1: val1}    # maps
var ex = 12                 # variables

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

# Arrays
var arr = [1, 2, 3]
arr[0] == 1
push(arr, 4)
pop(arr)

# Maps
var m = {"test": value, 1: 23}
m["test"] = value
map_set(m, "new", 12)
pop(m, "new")
len(m) == 2

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
```


# License

The interpreter is licensed under the terms of the MIT License.