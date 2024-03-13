module github.com/gaukas/watermob

go 1.21

require (
	github.com/refraction-networking/water v0.6.3
	golang.org/x/mobile v0.0.0-20240213143359-d1f7d3436075
)

replace github.com/tetratelabs/wazero v1.6.0 => github.com/refraction-networking/wazero v1.6.6-w

require (
	github.com/tetratelabs/wazero v1.6.0 // indirect
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a // indirect
	golang.org/x/mod v0.15.0 // indirect
	golang.org/x/tools v0.18.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
