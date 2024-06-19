module github.com/gaukas/watermob

go 1.21

require (
	github.com/gaukas/benchmarkconn v0.0.1
	github.com/refraction-networking/water v0.7.0-alpha
	github.com/tetratelabs/wazero v1.7.3
	golang.org/x/mobile v0.0.0-20240604190613-2782386b8afd
)

require (
	github.com/blang/vfs v1.0.0 // indirect
	github.com/gaukas/wazerofs v0.1.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/tetratelabs/wazero => github.com/refraction-networking/wazero v1.7.3-w
