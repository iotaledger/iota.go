gIOTA CLI tool
=====

CLI client tool for the IOTA reference implementation (IRI) using gIOTA lib.

Install
====
```
    $ go get -u github.com/iotaledger/giota/giotan
```

Features
====

1. Sending iota token using [public nodes](http://iotasupport.com/lightwallet.shtml) with local PoW.
2. List used and unused Addresses which can be generated from seed.

This CLI mainly focuses on functions using seeds.

If you want to add some functions, please make an issue.

Examples
====

```
    $ giotan new
    $ giotan addresses 
    $ giotan send --recipient=SOMERECIPIENT --amount=1234
    $ giotan send --recipient=SOMERECIPIENT --amount=1234 --sender=SOMEADDRESS1,SOMEADDRESS2,SOMEADDRESS3
```

When you use `addresses` and `send`, you will be prompted to input your seed.

When you use --sender, you must specify the addresses which can be generated from `seed`.

Note that `send` takes a long time to calculate Proof of Work.

Development Status: Alpha+
=========================

Tread lightly around here. This tool is still very much
in flux and there are going to be breaking changes.


TODO
=========================

* More functions(?)
* More tests :(

<hr>

Released under the [MIT License](LICENSE).
