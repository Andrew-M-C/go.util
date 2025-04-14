package admindivision_test

import (
	"os"
	"testing"
	"time"

	ad "github.com/Andrew-M-C/go.util/china/admindivision"
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

func TestGeneral(t *testing.T) {
	cv("正常区位", t, func() {
		chain := ad.SearchDivisionByCode("610102")
		desc := ad.DescribeDivisionChain(chain, "/")
		so(desc, eq, "陕西省/西安市/新城区")
	})

	cv("东莞的市辖区", t, func() {
		chain := ad.SearchDivisionByCode("4419")
		desc4419 := ad.DescribeDivisionChain(chain, "/")
		chain = ad.SearchDivisionByCode("441900")
		desc441900 := ad.DescribeDivisionChain(chain, "/")

		so(desc4419, eq, "广东省/东莞市")
		so(desc4419, eq, desc441900)
		so(chain[0].Deprecated(), eq, false)
		so(chain[1].Deprecated(), eq, false)
		so(chain[2].Deprecated(), eq, false)
	})

	cv("上海市", t, func() {
		chain := ad.SearchDivisionByCode("31")
		desc31 := ad.DescribeDivisionChain(chain, "/")
		chain = ad.SearchDivisionByCode("3101")
		desc3101 := ad.DescribeDivisionChain(chain, "/")

		so(desc31, eq, "上海市")
		so(desc31, eq, desc3101)

		chain = ad.SearchDivisionByCode("310101")
		desc310101 := ad.DescribeDivisionChain(chain, "/")
		so(desc310101, eq, "上海市/黄浦区")
	})

	cv("神农架林区", t, func() {
		chain := ad.SearchDivisionByCode("429021")
		desc := ad.DescribeDivisionChain(chain, "/")
		so(desc, eq, "湖北省/神农架林区")
	})

	cv("已撤销的行政区划", t, func() {
		chain := ad.SearchDivisionByCode("352229")
		desc := ad.DescribeDivisionChain(chain, "/")
		so(desc, eq, "福建省/宁德地区/寿宁县")
		so(len(chain), eq, 3)
		so(chain[0].Deprecated(), eq, false)
		so(chain[1].Deprecated(), eq, true)
		so(chain[2].Deprecated(), eq, true)
	})
}

func TestDivisionByName(t *testing.T) {
	cv("精确匹配", t, func() {
		chain := ad.MatchDivisionByName("广东省", "东莞市")
		so(len(chain), eq, 2)
		so(ad.JoinDivisionCodes(chain), eq, "4419")
	})

	cv("模糊匹配", t, func() {
		start := time.Now()
		chain := ad.SearchDivisionByName("广东", "东莞")
		ela := time.Since(start)
		t.Logf("耗时: %v", ela)

		so(len(chain), eq, 2)
		so(ad.JoinDivisionCodes(chain), eq, "4419")
	})
}
