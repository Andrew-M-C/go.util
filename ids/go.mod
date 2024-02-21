module github.com/Andrew-M-C/go.util/ids

go 1.19

replace (
	github.com/Andrew-M-C/go.util/runtime => ../runtime
	github.com/Andrew-M-C/go.util/slice => ../slice
	github.com/Andrew-M-C/go.util/time => ../time
)

require (
	github.com/Andrew-M-C/go.util/runtime v0.0.0-00010101000000-000000000000
	github.com/Andrew-M-C/go.util/slice v0.0.0-00010101000000-000000000000
	github.com/bwmarrin/snowflake v0.3.0
	github.com/smartystreets/goconvey v1.8.1
)

require (
	github.com/Andrew-M-C/go.jsonvalue v1.3.6 // indirect
	github.com/Andrew-M-C/go.util/time v0.0.0-00010101000000-000000000000 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
)
