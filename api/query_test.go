package api

import (
	"testing"
	"github.com/tomtom5152/dnsyo/dnsyo"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"net/http"
	"io/ioutil"
	"strings"
)

const (
	testYaml = "../config/test-resolver-list.yml"
)

func TestAPIServer_QueryHandler(t *testing.T) {
	sl, _ := dnsyo.ServersFromFile(testYaml)
	if len(sl) != 9 {
		t.Error("incorred number of servers, double check test list")
	}

	api := NewAPIServer(sl)

	server := httptest.NewServer(api.r)
	defer server.Close()

	testURL := server.URL + "/v1/query/example.com"

	Convey("make a test query to example.com and check the result", t, func() {
		resp, err := http.Get(testURL + "?q=9")
		So(err, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, http.StatusOK)

		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		json := string(data)

		So(json, ShouldNotBeBlank)
		So(json, ShouldStartWith, "{")
		So(json, ShouldEndWith, "}\n")

		Convey("check the postec fail is in there", func() {
			So(json, ShouldContainSubstring, `"!postec.nottingham.ac.uk":{"Answer":"","Error":"TIMEOUT"}`)
		})

		Convey("check the google result is sensible", func() {
			So(json, ShouldContainSubstring, `"google-public-dns-a.google.com":{"Answer":"93.184.216.34"}`)
		})
	})

	Convey("check query string params", t, func() {
		Convey("number of servers", func() {
			Convey("short form", func() {
				resp, err := http.Get(testURL + "?q=1")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				json := string(data)

				So(strings.Count(json, "Answer"), ShouldEqual, 1)
			})

			Convey("long form", func() {
				resp, err := http.Get(testURL + "?servers=3")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				json := string(data)

				So(strings.Count(json, "Answer"), ShouldEqual, 3)
			})
		})

		Convey("country", func() {
			Convey("short form", func() {
				resp, err := http.Get(testURL + "?c=GB")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				json := string(data)

				So(strings.Count(json, "Answer"), ShouldEqual, 1)
				So(json, ShouldContainSubstring, "!postec.nottingham.ac.uk")
			})

			Convey("long form", func() {
				resp, err := http.Get(testURL + "?country=GB")
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				json := string(data)

				So(strings.Count(json, "Answer"), ShouldEqual, 1)
				So(json, ShouldContainSubstring, "!postec.nottingham.ac.uk")
			})
		})

		Convey("type", func() {
			Convey("short form", func() {
				resp, err := http.Get(server.URL + "/v1/query/exmaple.com?t=MX") // deliberate typo so we get a result
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				json := string(data)

				So(json, ShouldContainSubstring, "10 numpty.absolutelyplastered.com.")
			})

			Convey("long form", func() {
				resp, err := http.Get(server.URL + "/v1/query/exmaple.com?type=MX") // deliberate typo so we get a result
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				json := string(data)

				So(json, ShouldContainSubstring, "10 numpty.absolutelyplastered.com.")
			})
		})
	})

	Convey("check request based errors", t, func() {
		Convey("bad type", func() {
			resp, err := http.Get(testURL + "?t=foo")
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("too many servers requested", func() {
			resp, err := http.Get(testURL + "?q=10")
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("more than the maximum number of servers", func() {
			resp, err := http.Get(testURL + "?q=1000")
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("invalid country", func() {
			resp, err := http.Get(testURL + "?c=FOO")
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}
