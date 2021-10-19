# iota.go

Official Go library.

This library allows you to do the following:

- Create messages with indexation and transaction payloads
- Get messages and outputs
- Sign transactions
- Generate addresses
- Interact with an IOTA node
- Act as a foundation for Go based node software

If you need to have more sophisticated account management, have a look
at [wallet.rs](https://github.com/iotaledger/wallet.rs) for which we also provide bindings in Python and JavaScript.

## Requirements

> This library was mainly tested with Go version 1.16.x

To use the library, we recommend you update Go [to the latest stable version](https://golang.org/).

## Using the library

Using the library is easy, just `go get` it as any other dependency:

```bash
go get github.com/iotaledger/iota.go/v3
```

## API reference

You can read the API reference [here](https://pkg.go.dev/github.com/iotaledger/iota.go/v3).

## Joining the discussion

If you want to get involved in the community, need help with setting up, have any issues or just want to discuss IOTA
with other people, feel free to join our [Discord](https://discord.iota.org/) in the `#clients-dev`
and `#clients-discussion` channels.

## License

The MIT license can be found [here](LICENSE).