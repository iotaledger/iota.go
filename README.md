[![Build Status](https://travis-ci.org/iotaledger/iota.lib.go.svg?branch=master)](https://travis-ci.org/iotaledger/iota.lib.go)
[![GoDoc](https://godoc.org/github.com/iotaledger/iota.lib.go?status.svg)](https://godoc.org/github.com/iotaledger/iota.lib.go)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/iotaledger/iota.lib.go/master/LICENSE)
[![Coverage Status](https://coveralls.io/repos/iotaledger/iota.lib.go/badge.svg?branch=master)](https://coveralls.io/r/iotaledger/iota.lib.go?branch=master)


gIOTA
=====

Client library for the IOTA reference implementation (IRI).

Refer to [godoc](https://godoc.org/github.com/iotaledger/iota.lib.go) for details.

Install
====
```
    $ go get -u github.com/iotaledger/giota
```

Examples
====

```go

import github.com/iotaledger/giota

//Trits
tritsFrom:=[]int8{1,-1,1,0,1,1,0,-1,0}
trits,err:=giota.ToTrits(tritsFrom)

//Trytes
trytes:=trits.Trytes()
trytesFrom:="ABCDEAAC9ACB9PO..."
trytes2,err:=giota.ToTrytes(trytesFrom)

//Hash
hash:=trits.Hash()

//API
api := giota.NewAPI("http://localhost:14265", nil)
ftr := &giota.FindTransactionsRequest{Bundles: []Trytes{"DEXRPL...SJRU"}}
resp, err := api.FindTransactions(ftr)

///Address
index:=0
security:=2
adr,err:=giota.NewAddress(trytes,index,seciruty) //without checksum.
adrWithChecksum := adr.WithChecksum() //adrWithChecksum is trytes type.

//transaction
tx,err:=giota.NewTransaction(trits)
if tx.HasValidNonce(){...}
trits2:=tx.Trits()

//create signature
key := giota.NewKey(seed.Trits(), index, security)
norm := bundleHash.Normalize()
sign := giota.Sign(norm[:27], key[:6561])

//validate signature
if giota.ValidateSig(adr, []Trits{sign}, bundleHash) {...}

//send
	trs := []Transfer{
		Transfer{
			Balance: Balance{
				Address: "KTXF...QTIWOWTY",
				Value:   20,
			},
			Tag: "MOUDAMEPO",
		},
	}
	bdl, err = Send(api, seed, 2, trs, PowGo)
```



Development Status: Alpha+
=========================

Tread lightly around here. This library is still very much
in flux and there are going to be breaking changes.


TODO
=========================

* Multisig
* More tests :(

<hr>
Released under the [MIT License](https://raw.githubusercontent.com/iotaledger/iota.lib.go/master/LICENSE).
