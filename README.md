[![Build Status](https://travis-ci.org/Confbase/cfgd.svg?branch=master)](https://travis-ci.org/Confbase/cfgd) [![Go Report Card](https://goreportcard.com/badge/github.com/Confbase/cfgd)](https://goreportcard.com/report/github.com/Confbase/cfgd)

# Overview

**cfgd** is a server which hosts [cfg](https://github.com/Confbase/cfg) bases
and snapshots.

Snapshots are stored on the file system, but different backends can act as a
cache between end-users and the file system on which cfgd is running. For
example, [redis](https://redis.io) is an officially supported backend. Using
redis as a backend is as easy as launching cfgd with the `--backend` flag set to
"redis":

```
$ cfgd --backend=redis
```

**cfgd** also provides an API so that custom backends can be written without
modifying the source code of cfgd or having to re-install it. See the
[Custom Backend API](#custom-backend-api) section below.

## Contents

This repository contains four things:

1. cfgd---the cfg server daemon
2. a binary for creating cfg snapshot messages called cfg-build-snap
3. a binary for uploading snapshots to cfgd called cfg-send-snap
4. a post-receive git hook which leverages cfg-build-snap and cfg-send-snap to
upload snapshots to cfgd

## Installation

To build from source, run `go get -u github.com/Confbase/cfgd`.

Place the post-receive hook, cfg-build-snap, and cfg-send-snap in the
appropriate directories (TODO: complete this documentation).

## Testing

Run `./test.sh`. Tests require bash and curl.

## Custom Backend API

Custom backends are executable binaries which parse commands from standard
input. To use a custom backend, specifiy it with the `--custom-backend` flag
when launching cfgd:

```
$ cfgd --custom-backend=etcd
```

The string value of the `--custom-backend` flag is interpreted as the path to an
executable binary.

Custom backends must implement the GET command and the PUT command.

### The GET Command

When cfgd needs to retrieve a file from the custom backend, it executes the
binary and pipes to it the string

```
GET <file-key>
```

where `<file-key>` is in the format

```
<base-name>/<snapshot-name>/<path-to-file>
```

**cfgd** expects the binary to write exactly the string

```
OK<file-contents>
```

to its standard output, if the requested file exists. Otherwise **cfgd** expects
the binary to write exactly the string (just these two bytes)

```
NO
```

to its standard output.

If the binary exits with a non-zero exit status, then it is assumed to have
failed and cfgd will respond to the user who requested the file with
"500 Internal Server Error".

### The PUT Command

When cfgd needs to store data in the custom backend, it executes the binary and
pipes to it the string (called the *PUT header*)

```
PUT <snap-key>\n
```

where `<snap-key>` is formatted like this

```
<base-name>/<snapshot-name>
```

followed by any number of *binary PUT messages*, each containing a file path and
file contents, formatted like this

| File Path String Length                                | File Content Length                            | File Path String       | File Content    |
|----------------------------------------------------|---------------------------------------------------|-----------------|-----------------|
| First 2 bytes (unsigned big endian 16-bit integer) | Next 8 bytes (unsigned big endian 64-bit integer) | Variable length | Variable length |

**Note:** if you intend to write a custom backend in Go, you can simply
import the code which cfgd uses to parse PUT messages. See this excerpt from
cfgd's source code:

```
import (
	...
	"github.com/Confbase/cfgd/snapshot"
	...
)

...

snapReader := snapshot.NewReader(br)
for {
	sf, done, err := snapReader.Next()
	if err != nil {
		return false, fmt.Errorf("snapReader failed: %v", err)
	}
	if done {
		break
	}
	// use sf.FilePath and sf.Body here
}
```

The binary cfg-build-snap can serve as a reference for this format. It takes one
argument---the path to a directory---and recursively traverses the directory,
writing a *binary PUT message* to standard output for each file it finds. Note
that cfg-build-snap does not include the *PUT header*. Example:

```
$ cd $GOPATH/src/github.com/Confbase/cfgd/cfg-build-snap
$ ls test_snapshot
hello.toml  hello_schema.json
$ cfg-build-snap test_snapshot > reference_snap
```

The file `reference_snap` will contain two *binary PUT messages*, one for
`hello.toml` and one for `hello_schema.json`.

## LICENSE

Apache V2

## Contributing

Pull requests and new issues are welcome.
