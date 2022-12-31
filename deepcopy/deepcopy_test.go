// Package deepcopy 用于分析并输出一个 struct 的深复制代码
package deepcopy

import (
	"testing"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	isNil = convey.ShouldBeNil
)

func TestDeepCopy(t *testing.T) {
	cv("test GenerateDeepCopyCode", t, func() { testGenerateDeepCopyCode(t) })
}

func testGenerateDeepCopyCode(t *testing.T) {
	cv("debug", func() { testGenerateDeepCopyCode_Debug(t) })
	// TODO:
}

func testGenerateDeepCopyCode_Debug(t *testing.T) {
	code, err := GenerateDeepCopyCode(jsonvalue.NewObject(), WithDebug(true), WithLogFunc(t.Logf))
	so(err, isNil)
	t.Logf("got code:\n---\n%s\n----", code)

	code, err = GenerateDeepCopyCode(myStruct{}, WithDebug(true), WithLogFunc(t.Logf))
	so(err, isNil)
	t.Logf("got code:\n---\n%s\n----", code)

	code, err = GenerateDeepCopyCode(&myStruct{}, WithDebug(false), WithLogFunc(t.Logf), WithPackageName("deepcopy_test"))
	so(err, isNil)
	t.Logf("got code:\n---\n%s\n----", code)
}

type myStruct struct {
	Name *Name
	Tags []*Tag

	Login struct {
		Username string
		Password string
	}

	Property   *jsonvalue.V
	Properties []*jsonvalue.V

	Chan chan bool

	internalName Name
}

type Name struct {
	First  string
	Middle string
	Last   string
}

type Tag struct {
	Name string
}
