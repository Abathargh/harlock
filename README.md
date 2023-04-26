# Harlock

Harlock is a small language with a focus on binary data manipulation and
data embedding in executables, mainly usable as a scripting language for 
binaries post-processing. It is based around the ideas discussed 
in the *Writing An Interpreter In Go* book by Thorsten Ball.

- [Harlock](#harlock)
	- [Why a language?](#why-a-language)
	- [Download](#download)
	- [Build](#build)
	- [Usage](#usage)
		- [Run a script](#run-a-script)
		- [Start the REPL](#start-the-repl)
		- [Embed a script](#embed-a-script)
	- [Use cases](#use-cases)
		- [File APIs](#file-apis)
		- [Hex file manipulation](#hex-file-manipulation)
		- [Elf file manipulation](#elf-file-manipulation)
		- [Generic file manipulation](#generic-file-manipulation)
	- [Language primitives \& operations](#language-primitives--operations)
	- [License](#license)


## Why a language?

This is an experiment based on a hobby project combined with a real use case that I had.

The ihex parser is pretty complete and I have an experimental project where I have partially ported it to C as a 
standalone library. This is something that I'm thinking of doing too for the core go sub-package.

One big advantage of having a small simple DSL is that it's quick and easy to just write a 4-5 line script, that you
can then insert into your build pipeline.

## Download

```bash
go install github.com/Abathargh/harlock/cmd/harlock@latest
```

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

## Use cases

Embedding metadata is supported for hex an elf formats: this language implements a thin wrapper on the ```debug/elf``` 
go package, and an ihex parser to make it possible to write and read from and to a file with an API that gives random 
access capabilities to the user.

### File APIs

Files can be accessed in order to extract data from them and manipulating their contents. 

When you access a file, its contents get loaded in memory and the os handle gets closed. You can manipulate the data and 
then choose to save the file onto the original one.
You can interact with files with the following builtin functions:

- ```open(filename: str, type: str) -> file_handle```: open the file with the passed name, type can be ```hex```, 
```elf```, or ```bytes```.
- ```save(file_handle)```: saves the changes to the file on the fs.
- ```as_bytes(file_handle) -> []int```: returns a copy of the file contents as an array of ints.

### Hex file manipulation

Consider the following simple c file:

```c
int main(void)
{
  volatile int c = 0;
  c++;
  c *= 2;
  while(1);
}
```

And let's compile it for a generic avr mcu:
```bash
avr-gcc -O0 -o test.elf main.c
avr-objcopy -Oihex test.elf test.hex
```

We can then manipulate the hex file as follows:

```
// test.hex:
// :10000000CF93DF9300D0CDB7DEB71A8219828981F2
// :100010009A8101969A83898389819A81880F991F91
// :060020009A838983FFCFE3
// :00000001FF 

var h = try open("test.hex", "hex")
print(h) // prints the contents as a string

// Read n bytes (qst param) at position m (0-th param)
print(h.read_at(0, 16))      // prints [207, 147, 223, 147, 0, 208, 205, 183, 222, 183, 26, 130, 25, 130, 137, 129]

// Write the array at the specified position
try h.write_at(0, [1, 2, 3]) 
print(h.read_at(0, 16))      // prints [1, 2, 3, 147, 0, 208, 205, 183, 222, 183, 26, 130, 25, 130, 137, 129]


try h.write_at(0, hex("deadbeef")) 
print(h.read_at(0, 16))      // prints [222, 173, 190, 239, 0, 208, 205, 183, 222, 183, 26, 130, 25, 130, 137, 129]

print(as_bytes(h))           // prints [58, 49, 48, 48, 48, 48, ... 70, 70, 13, 10]

// Read the i-th record as a string
print(h.record(0))           // prints ":10000000DEADBEEF00D0CDB7DEB71A82198289818E"
print(h.size())              // prints 4

// Dump the written contents back to the file
save(h)
```

### Elf file manipulation

Consider the following simple c file:

```c
static const unsigned char
data[256] __attribute__((section(".embmetadata")));

int main(void)
{
  while(1);
}
```

And let's compile it for a generic avr mcu:
```bash
avr-gcc -O0 -o test.elf main.c
```

We can then manipulate the hex file as follows:

```
var e = try open("test.elf", "elf")
print(e) // prints the elf object info

// Check if the section with the given name exists
if e.has_section(".embmetadata") {
  print("section .embmetadata exists")  
}

// This print will not get executed 
if e.has_section(".test") {
  print("section .test exists")  
}

print(e.sections()) // prints [.text, .embmetadata, .data, .comment, .symtab, .strtab, .shstrtab]

// Read the contents of the given section and returns a copy as an array of ints
print(e.read_section(".embmetadata")) // prints [0, 0, 0, 0, 0, 0, ... , 0, 0]

// Write some data (an array of ints here) into the given section at offset 0
try e.write_section(".embmetadata", [58, 59], 0)
print(e.read_section(".embmetadata")) // prints [58, 59, 0, 0, 0, 0, ... , 0, 0]


print(as_bytes(h)) // prints [127, 69, 76, 70, 1, 1, 1, 0, 0, 0, 0, 0, 0, ... ]

// Dump the written contents back to the file
save(e)
```

### Generic file manipulation

A generic file can be manipulated as a stream of bytes:

```
// test is a text file containing the following contents:
// "hello world!"

var b = try open("test", "bytes")
print(b) // prints the bytes object 

// Read 5 bytes from the file at position 0
print(b.read_at(0, 5)) // prints [32, 119, 111, 114, 108]

// Write some data (an array of ints here) at position 0
try b.write_at(0, [1, 2, 3])
print(b.read_at(0, 5)) // prints [1, 2, 3, 114, 108]

// Dump the written contents back to the file
save(e)
```

## Language primitives & operations

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
```


## License

The interpreter is licensed under the terms of the MIT License.