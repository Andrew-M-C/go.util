module github.com/Andrew-M-C/go.util/runtime

go 1.19

replace (
	github.com/Andrew-M-C/go.util/slice => ../slice
	github.com/Andrew-M-C/go.util/time => ../time
)

require (
	github.com/Andrew-M-C/go.jsonvalue v1.3.6
	github.com/Andrew-M-C/go.util/time v0.0.0-00010101000000-000000000000
	github.com/smartystreets/goconvey v1.8.1
	go.uber.org/automaxprocs v1.5.3
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a
)

require (
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
)
