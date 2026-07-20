package securityspy_test

import _ "embed"

//go:embed testdata/systemInfo-v6.xml
var testSystemInfoV6 string

// testSystemInfo remains the legacy v5 fixture for dual-read regression tests.
