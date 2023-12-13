package json_test

import (
	"encoding/json"
	"testing"

	jsutil "github.com/Andrew-M-C/go.util/beta/encoding/json"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	isNil = convey.ShouldBeNil
)

type adapterTestType01 struct {
	GameID string                  `json:"game_id"`
	Users  []adapterTestType01User `json:"users"`
}

type adapterTestType01User struct {
	UserID string            `json:"user_id"`
	Ext    map[string]string `json:"ext"`
}

type adapterTestType02 struct {
	GID     int32    `json:"gid"`
	OpenIDs []string `json:"openid"`
	Reflow  []int    `json:"reflow"`
}

func TestProtocolAdapter(t *testing.T) {
	cv("基础逻辑", t, func() {
		in := adapterTestType01{
			GameID: "1234",
			Users: []adapterTestType01User{
				{
					UserID: "123456",
					Ext: map[string]string{
						"reflow": "10",
					},
				}, {
					UserID: "ABCDEF",
					Ext:    nil,
				},
			},
		}

		const confRaw = `[
			{
				"from": "game_id",
				"to": "gid",
				"type": "int"
			}, {
				"from": "users.[n].user_id",
				"to": "openid.[n]"
			}, {
				"from": "users.[n].ext.reflow",
				"to": "reflow.[n]",
				"type": "int"
			}
		]`

		var conf []jsutil.ProtocolAdapterMapping
		err := json.Unmarshal([]byte(confRaw), &conf)
		so(err, isNil)

		adapter, err := jsutil.NewProtocolAdapterByFieldsConfig(conf)
		so(err, isNil)

		out := adapterTestType02{}
		err = adapter.ConvertJSON(in, &out)
		so(err, isNil)

		b, err := json.Marshal(out)
		so(err, isNil)

		so(out.GID, eq, 1234)
		so(len(out.OpenIDs), eq, 2)
		so(out.OpenIDs[0], eq, "123456")
		so(out.OpenIDs[1], eq, "ABCDEF")
		so(len(out.Reflow), eq, 2)
		so(out.Reflow[0], eq, 10)
		so(out.Reflow[1], eq, 0)

		t.Log(string(b))
	})
}
