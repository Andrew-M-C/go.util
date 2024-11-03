package gotest

import "github.com/smartystreets/goconvey/convey"

// 本文件提供 convey 的封装

var (
	Convey = convey.Convey

	So = convey.So

	EQ = convey.ShouldEqual
	NE = convey.ShouldNotEqual
	LT = convey.ShouldBeLessThan
	GT = convey.ShouldBeGreaterThan
	LE = convey.ShouldBeLessThanOrEqualTo
	GE = convey.ShouldBeGreaterThanOrEqualTo

	NotNil = convey.ShouldNotBeNil
	IsNil  = convey.ShouldBeNil
)
