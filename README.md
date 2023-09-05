# rc [![Go Reference](https://pkg.go.dev/badge/github.com/k1LoW/rc.svg)](https://pkg.go.dev/github.com/k1LoW/rc) [![build](https://github.com/k1LoW/rc/actions/workflows/ci.yml/badge.svg)](https://github.com/k1LoW/rc/actions/workflows/ci.yml) ![Coverage](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/rc/coverage.svg) ![Code to Test Ratio](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/rc/ratio.svg) ![Test Execution Time](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/rc/time.svg)

`rc` is a **r**esponse **c**ache middleware for multiple cache rules and multiple cache stores.

## Usage

Prepare an instance that implements [`rc.Casher`](https://pkg.go.dev/github.com/k1LoW/rc#Cacher) interface.

Then, generate the middleware ( `func(next http.Handler) http.Handler` ) with [`rc.New`](https://pkg.go.dev/github.com/k1LoW/rc#New)

```go
package main

import (
    "log"
    "net/http"

    "github.com/k1LoW/rc"
)

func main() {
    r := http.NewServeMux()
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
    })

    var c rc.Cacher = newMyCacher()
    m := rc.New(c)

    log.Fatal(http.ListenAndServe(":8080", m(r)))
}
```

## Utility functions

See https://github.com/k1LoW/rcutil
