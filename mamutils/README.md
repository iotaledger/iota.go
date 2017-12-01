# mamgiota

Small project to implement Masked Authenticated Messaging on the IOTA tangle with Golang.

This project is still under construction (see TODO) with the aim to get the ruuVi tag and other IoT sensors and devices to send MAMs.
To test this functionality you will need a message board URL with its IOTA address so it can receive these massages.

## Install

It is assumed that you have Golang installed. You also need to install the Go library API for IOTA which you can download at:

```javascript
go get -u github.com/iotaledger/giota
```

After that you can download `send-message` to test from

```javascript
go get -u github.com/habpygo/mamgoiota
```

Have fun!

## TODO

Get ruuVi tag or other sensors to send to the tangle.
