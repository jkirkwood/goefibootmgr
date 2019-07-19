# goefibootmgr

Simple wrapper for [efibootmgr](https://github.com/rhboot/efibootmgr) command.

## Installation

Run:

```
go get github.com/jkirkwood/goefibootmgr
```

## Docs

[![GoDoc](https://godoc.org/github.com/jkirkwood/goefibootmgr?status.svg)](https://godoc.org/github.com/jkirkwood/goefibootmgr)

## Usage

```go
package main

import "github.com/jkirkwood/goefibootmgr"

func main() {
  info, err := goefibootmgr.BootInfo()

  if err != nil {
    panic(err)
  }

  err = goefibootmgr.SetBootOrder(1, 3, 2)

  if err != nil {
    panic(err)
  }
}

```
