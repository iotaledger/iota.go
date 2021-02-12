module github.com/iotaledger/iota.go/v2

go 1.15

require (
	filippo.io/edwards25519 v1.0.0-beta.2
	github.com/iotaledger/iota.go v1.0.0-beta.15.0.20210120184258-7eac7c1cc80b
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/h2non/gock.v1 v1.0.15
)

replace github.com/iotaledger/iota.go => github.com/wollac/iota.go v1.0.0-beta.9.0.20210211182213-ad429d349f5d
