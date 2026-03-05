module golift.io/securityspy

go 1.25.7

toolchain go1.26.0

retract [v2.0.0, v2.0.3]

retract v1.0.0

require (
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.6.0
	golift.io/ffmpeg 5b35b525534e
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
