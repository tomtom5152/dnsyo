package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/miekg/dns"
)

const (
	testYaml = "config/test-resolver-list.yml"
)

func TestServerListFromCSVUrl(t *testing.T) {
	Convey("server list is loaded from the file", t, func() {
		sl, err := ServersFromFile(testYaml)
		So(err, ShouldBeNil)
		So(sl, ShouldHaveLength, 8)

		Convey("check a few common results are in there", func() {
			googleA := Server{
				Ip:       "8.8.8.8",
				Country:  "US",
				Provider: "GOOGLE - Google Inc.",
				Reverse:  "google-public-dns-a.google.com",
			}

			So(sl, ShouldContain, googleA)
		})

		Convey("all items are in the yaml, there aren't any fakes", func() {
			l3 := Server{
				Country:  "GB",
				Ip:       "193.240.163.34",
				Provider: "LVLT-3549 - Level 3 Communications, Inc.,US",
				Reverse:  "193.240.163.34",
			}

			So(sl, ShouldNotContain, l3)
		})
	})
}

func TestServer_Test(t *testing.T) {
	Convey("valid server is ok", t, func() {
		Convey("Google A", func() {
			s := Server{
				Ip:       "8.8.8.8",
				Country:  "US",
				Provider: "GOOGLE - Google Inc.",
				Reverse:  "google-public-dns-a.google.com",
			}

			ok, err := s.Test()
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})

		Convey("DNS Watch A", func() {
			s := Server{
				Ip:       "84.200.69.80",
				Country:  "DE",
				Provider: "ACCELERATED - Google Inc.",
				Reverse:  "resolver1.dns.watch",
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
				Provider: "!UNI-NOTTINGHAM - TEC PA & Lighting",
				Reverse:  "s.nottingham.ac.uk",
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
			Provider: "GOOGLE - Google Inc.",
			Reverse:  "google-public-dns-a.google.com",
		}

		Convey("google.com NS as these are unlikely to change", func() {
			results, err := s.Lookup("google.com", dns.TypeNS)
			So(err, ShouldBeNil)
			So(results, ShouldHaveLength, 4)
			So(results, ShouldContain, "google.com.	21599	IN	NS	ns1.google.com.")
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
			Provider: "!UNI-NOTTINGHAM - TEC PA & Lighting",
			Reverse:  "s.nottingham.ac.uk",
		}

		Convey("itsg.host NS as these are unlikely to change", func() {
			results, err := s.Lookup("itsg.host", dns.TypeNS)
			So(results, ShouldBeNil)
			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "TIMEOUT")
		})
	})
}
