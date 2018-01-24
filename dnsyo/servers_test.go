package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/miekg/dns"
	"time"
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
		result := sl.Query("example.com", dns.TypeA, time.Millisecond * 50)
		So(result, ShouldNotBeNil)

		// every server should be polled
		So(result.SuccessCount + result.ErrorCount, ShouldEqual, len(sl))

		// with out test list we should have 8 success and 1 failure
		So(result.SuccessCount, ShouldEqual, 8)
		So(result.ErrorCount, ShouldEqual, 1)

		// check the result we have is correct
		So(result.Success, ShouldResemble, map[string]int{"93.184.216.34": 8})
		So(result.Errors, ShouldResemble, map[string]int{"TIMEOUT": 1})
	})
}

func TestServer_Test(t *testing.T) {
	Convey("valid server is ok", t, func() {
		Convey("Google A", func() {
			s := Server{
				Ip:       "8.8.8.8",
				Country:  "US",
				Name:  "google-public-dns-a.google.com",
			}

			ok, err := s.Test()
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})

		Convey("DNS Watch A", func() {
			s := Server{
				Ip:       "84.200.69.80",
				Country:  "DE",
				Name:  "resolver1.dns.watch",
			}

			ok, err := s.Test()
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})
	})

	Convey("nonexistant server does not return ok", t, func() {
		Convey("postec.nottingham.ac.uk is not and will never be a DNS server", func() {
			s := Server{
				Ip:       "128.243.103.175",
				Country:  "GB",
				Name:  "!postec.nottingham.ac.uk",
			}

			ok, err := s.Test()
			So(err, ShouldBeError)
			So(ok, ShouldBeFalse)
		})

	})
}

func TestServer_Lookup(t *testing.T) {
	Convey("test against a valid server", t, func() {
		s := Server{
			Ip:       "8.8.8.8",
			Country:  "US",
			Name:  "google-public-dns-a.google.com",
		}

		Convey("google.com NS as these are unlikely to change", func() {
			results, err := s.Lookup("google.com", dns.TypeNS)
			So(err, ShouldBeNil)
			So(results, ShouldHaveLength, 4)
			So(results, ShouldContain, "ns1.google.com.")
		})

		Convey("dne.itsg.host A does not exist, check the failure", func() {
			results, err := s.Lookup("dne.itsg.host", dns.TypeA)
			So(results, ShouldBeNil)
			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "NOANSWER")
		})

		Convey("itsg.test A cannot exist, check the failure is NXDOMAIN", func() {
			results, err := s.Lookup("dne.itsg.test", dns.TypeA)
			So(results, ShouldBeNil)
			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "NXDOMAIN")
		})
	})

	Convey("test for a timeout with an invalid server", t, func() {
		s := Server{
			Ip:       "128.243.103.175",
			Country:  "GB",
			Name:  "!postec.nottingham.ac.uk",
		}

		Convey("itsg.host NS as these are unlikely to change", func() {
			results, err := s.Lookup("itsg.host", dns.TypeNS)
			So(results, ShouldBeNil)
			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "TIMEOUT")
		})
	})

	Convey("localhost should refuse the connection", t, func () {
		s := Server{
			Ip:       "127.0.0.1",
			Country:  "NA",
			Name:  "localhost",
		}

		Convey("google.com NS as these are unlikely to change", func() {
			results, err := s.Lookup("google.com", dns.TypeNS)
			So(results, ShouldBeNil)
			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "CONNECTION REFUSED")
		})
	})
}
