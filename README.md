[![Build Status](https://travis-ci.org/iotaledger/iota.lib.go.svg?branch=master)](https://travis-ci.org/iotaledger/iota.lib.go)
[![GoDoc](https://godoc.org/github.com/iotaledger/iota.lib.go?status.svg)](https://godoc.org/github.com/iotaledger/iota.lib.go)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/iotaledger/iota.lib.go/master/LICENSE)


gIOTA
=====

Client library for the IOTA reference implementation (IRI).

Refer to [godoc](https://godoc.org/github.com/iotaledger/iota.lib.go) for details.

Install
====

You will need `gcc` for linux to compile PoW routine in C.

```
    $ go get -u github.com/iotaledger/giota
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
adr,err:=giota.NewAddress(trytes,index,seciruty) //without checksum.
adrWithChecksum := adr.WithChecksum() //adrWithChecksum is trytes type.

//transaction
tx,err:=giota.NewTransaction(trytes)
if tx.HasValidNonce(){...}
trytes2:=tx.trytes()

//create signature
key := giota.NewKey(seed, index, security)
norm := bundleHash.Normalize()
sign := giota.Sign(norm[:27], key[:6561/3])

//validate signature
if giota.ValidateSig(adr, []Trytes{sign}, bundleHash) {...}

//send
trs := []giota.Transfer{
	giota.Transfer{
		Address: "KTXF...QTIWOWTY",
		Value:   20,
		Tag: "MOUDAMEPO",
	},
}
_, pow := giota.GetBestPow()
bdl, err = giota.Send(api, seed, security, trs, pow)
```



Development Status: Alpha+
=========================

Tread lightly around here. This library is still very much
in flux and there are going to be breaking changes.

Consider to use a dependency tool to use vendoring,
e.g. [godep](https://github.com/tools/godep), [glide](https://github.com/Masterminds/glide) or [govendor](https://github.com/kardianos/govendor).


TODO
=========================

* Multisig
* More tests :(

<hr>

Released under the [MIT License](LICENSE).
