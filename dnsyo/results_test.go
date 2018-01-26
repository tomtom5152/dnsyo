package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQueryResults_ToJson(t *testing.T) {
	qr := &QueryResults{
		"localhost": &Result{Answer: "127.0.0.1"},
		"error":     &Result{Error: "TESTERR"},
	}

	Convey("json is correct", t, func() {
		json, err := qr.ToJson()
		So(err, ShouldBeNil)
		So(json, ShouldEqual, `{"error":{"Answer":"","Error":"TESTERR"},"localhost":{"Answer":"127.0.0.1"}}`)
	})
}
