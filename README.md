[![Build Status](https://travis-ci.org/iotaledger/iota.lib.go.svg?branch=master)](https://travis-ci.org/iotaledger/iota.lib.go)
[![GoDoc](https://godoc.org/github.com/iotaledger/iota.lib.go?status.svg)](https://godoc.org/github.com/iotaledger/iota.lib.go)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/iotaledger/iota.lib.go/master/LICENSE)


gIOTA
=====

Client library for the IOTA reference implementation (IRI).

This library is still in flux and there maybe breaking changes.

Consider to use a dependency tool to use vendoring,
e.g. [godep](https://github.com/tools/godep), [glide](https://github.com/Masterminds/glide) or [govendor](https://github.com/kardianos/govendor).


Refer to [godoc](https://godoc.org/github.com/iotaledger/iota.lib.go) for details.

Install
====

You will need C compiler for linux to compile PoW routine in C.

```
    $ go get -u github.com/iotaledger/giota
```

You will need C compiler and OpenCL environemnt(hardware and software)  to compile PoW routine for GPU 
and need to add `opencl` tag when you build.

```
	$ go build -tags=gpu
```

Examples
====

```go

import "github.com/iotaledger/giota"

//Trits
tritsFrom:=[]int8{1,-1,1,0,1,1,0,-1,0}
trits,err:=giota.ToTrits(tritsFrom)

//Trytes
trytes:=trits.Trytes()
trytesFrom:="ABCDEAAC9ACB9PO..."
trytes2,err:=giota.ToTrytes(trytesFrom)

//Hash
hash:=trytes.Hash()

//API
api := giota.NewAPI("http://localhost:14265", nil)
resp, err := api.FindTransactions([]Trytes{"DEXRPL...SJRU"})

///Address
index:=0
security:=2
adr,err:=giota.NewAddress(trytes,index,security) //without checksum.
adrWithChecksum := adr.WithChecksum() //adrWithChecksum is trytes type.

//transaction
tx,err:=giota.NewTransaction(trytes)
mwm := 14
if tx.HasValidNonce(mwm){...}
trytes2:=tx.trytes()

//create signature
key := giota.NewKey(seed, index, security)
norm := bundleHash.Normalize()
sign := giota.Sign(norm[:27], key[:6561/3])

//validate signature
if giota.ValidateSig(adr, []giota.Trytes{sign}, bundleHash) {...}

//send
trs := []giota.Transfer{
	giota.Transfer{
		Address: "KTXF...QTIWOWTY",
		Value:   20,
		Tag: "MOUDAMEPO",
	},
}
_, pow := giota.GetBestPoW()
bdl, err = giota.Send(api, seed, security, trs, mwm, pow)
```

PoW(Proof of Work) Benchmarking
====

You can benchmark PoWs(by C,Go,SSE) by

```
    $ go test -v -run Pow
```

or if you want to add OpenCL PoW,

```
    $ go test -tags=gpu -v -run Pow
```

then it outputs like:

```
	$ go test -tags=gpu -v -run Pow
=== RUN   TestPowC
--- PASS: TestPowC (15.93s)
	pow_c_test.go:50: 1550 kH/sec on C PoW
=== RUN   TestPowCL
--- PASS: TestPowCL (17.45s)
	pow_cl_test.go:49: 332 kH/sec on GPU PoW
=== RUN   TestPowGo
--- PASS: TestPowGo (21.21s)
	pow_go_test.go:50: 1164 kH/sec on Go PoW
=== RUN   TestPowSSE
--- PASS: TestPowSSE (13.41s)
	pow_sse_test.go:52: 2292 kH/sec on SSE PoW
```

Note that in [travis CI](https://travis-ci.org/iotaledger/iota.lib.go/jobs/227452499)
the result is:

```
=== RUN   TestPowSSE
--- PASS: TestPowSSE (2.73s)
	pow_sse_test.go:52: 12902 kH/sec on SSE PoW
=== RUN   TestPowSSE1
--- PASS: TestPowSSE1 (16.19s)
	pow_sse_test.go:59: 1900 kH/sec on SSE PoW
=== RUN   TestPowSSE32
--- PASS: TestPowSSE32 (1.36s)
	pow_sse_test.go:67: 16117 kH/sec on SSE PoW
=== RUN   TestPowSSE64
--- PASS: TestPowSSE64 (0.73s)
	pow_sse_test.go:75: 20226 kH/sec on SSE PoW
```

It gets over `20MH/s` for 64 threads using SSE2.

Now IOTA uses Min Weight Magnitude = 15, which means 
3^15≒14M Hashes are needed to finish PoW in average.
So it takes just 14/20 < 0.7sec for 1 tx to do PoW.


TODO
=========================

* [ ] Multisig
* [ ] More tests :(

<hr>

Released under the [MIT License](LICENSE).
