package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/miekg/dns"
)

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
