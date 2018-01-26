package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testYaml = "../config/test-resolver-list.yml"
)

func TestServerListFromFile(t *testing.T) {
	Convey("server list is loaded from the file", t, func() {
		sl, err := ServersFromFile(testYaml)
		So(err, ShouldBeNil)
		So(sl, ShouldHaveLength, 9)

		Convey("check a few common results are in there", func() {
			googleA := Server{
				Ip:       "8.8.8.8",
				Country:  "US",
				Name:  "google-public-dns-a.google.com",
			}

			So(sl, ShouldContain, googleA)
		})

		Convey("all items are in the yaml, there aren't any fakes", func() {
			l3 := Server{
				Country:  "GB",
				Ip:       "193.240.163.34",
				Name:  "193.240.163.34",
			}

			So(sl, ShouldNotContain, l3)
		})
	})
}

func TestServerList_FilterCountry(t *testing.T) {
	sl, _ := ServersFromFile(testYaml)
	if len(sl) != 9 {
		t.Error("incorred number of servers, double check test list")
	}

	Convey("filtering to GB results in only one server", t, func() {
		gb, err := sl.FilterCountry("GB")
		So(err, ShouldBeNil)
		So(gb, ShouldHaveLength, 1)

		Convey("that should be the postec entry", func() {
			s := Server{
				Ip:       "128.243.103.175",
				Country:  "GB",
				Name:  "!postec.nottingham.ac.uk",
			}

			So(gb[0], ShouldResemble, s)
		})
	})

	Convey("filtering to a country not in the list yields an error", t, func() {
		_, err := sl.FilterCountry("NOTME")
		So(err, ShouldBeError)
	})
}

func TestServerList_NRandom(t *testing.T) {
	sl, _ := ServersFromFile(testYaml)
	if len(sl) != 9 {
		t.Error("incorred number of servers, double check test list")
	}

	Convey("selecting a random number of servers", t, func() {
		rl, err := sl.NRandom(6)
		So(err, ShouldBeNil)

		So(rl, ShouldHaveLength, 6)
	})

	Convey("selecting too many servers produces an error", t, func() {
		_, err := sl.NRandom(10)
		So(err, ShouldBeError)
	})
}

func TestServerList_Query(t *testing.T) {
	sl, _ := ServersFromFile(testYaml)
	if len(sl) != 9 {
		t.Error("incorred number of servers, double check test list")
	}

	Convey("perform a query that will work and we can compare the results", t, func() {
		q := &Query{
			Domain: "example.com",
			Type: "A",
		}
		result := sl.ExecuteQuery(q, 10)
		So(result, ShouldNotBeNil)

		// every server should be polled
		So(len(result), ShouldEqual, len(sl))

		// check the result we have is correct
		So(result[sl[8].String()], ShouldResemble, &Result{Error:"TIMEOUT"})
	})
}
