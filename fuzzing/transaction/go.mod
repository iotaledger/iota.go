module transaction-fuzzing

go 1.16

replace github.com/iotaledger/iota.go/v2 => ./../..

require (
	github.com/dvyukov/go-fuzz v0.0.0-20210103155950-6a8e9d1f2415 // indirect
	github.com/iotaledger/iota.go/v2 v2.0.0-20210303080450-b3ab28991fc8
)
