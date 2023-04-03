package env_test

import (
	"os"
	"testing"

	"github.com/Andrew-M-C/go.util/runtime/env"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestEnv(t *testing.T) {
	cv("测试环境变量工具", t, func() { testEnv(t) })
}

func testEnv(t *testing.T) {
	cv("string", func() {
		const key = "STRING"
		s := env.GetString(key, "empty")
		so(s, eq, "empty")

		os.Setenv(key, "string")
		s = env.GetString(key, "empty")
		so(s, eq, "string")
	})

	cv("int", func() {
		const key = "INT"
		i := env.GetInt[int8](key, 20)
		so(i, eq, 20)

		os.Setenv(key, "-20")
		i = env.GetInt[int8](key, 20)
		so(i, eq, -20)
	})

	cv("bool", func() {
		const key = "BOOL"
		so(env.GetBool(key), eq, false)

		os.Setenv(key, "TRUE")
		so(env.GetBool(key), eq, true)

		os.Setenv(key, "true")
		so(env.GetBool(key), eq, true)

		os.Setenv(key, "1")
		so(env.GetBool(key), eq, true)

		os.Setenv(key, "2")
		so(env.GetBool(key), eq, true)

		os.Setenv(key, "0")
		so(env.GetBool(key), eq, false)

		os.Setenv(key, "-1")
		so(env.GetBool(key), eq, false)

		os.Unsetenv(key)
		so(env.GetBool(key), eq, false)
	})
}
