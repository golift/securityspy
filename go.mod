module golift.io/securityspy

go 1.25.7

toolchain go1.26.0

// This was never supposed to be released passed v0. Not sure why this happened!
retract [v1.0.0, v2.0.2+incompatible]

require (
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.6.0
	golift.io/ffmpeg v1.1.3-0.20260305034210-b29df3be46aa
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
