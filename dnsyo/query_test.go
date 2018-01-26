package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/miekg/dns"
)

func TestQuery_ToTextSummary(t *testing.T) {
	s1 := Server{
		Ip: "127.0.0.1",
	}
	s2 := Server{
		Ip: "127.0.0.2",
	}
	q := &Query{
		Domain: "example.test",
		Type: dns.TypeA,
	}

	Convey("top block is formatted correctly", t, func() {
		text := q.ToTextSummary()
		So(text, ShouldStartWith, `
 - RESULTS
I asked 0 servers for A records related to example.test,
`)
	})

	Convey("successes", t, func() {
		Convey("two different", func() {
			q.Results = QueryResults{
				s1.String(): &Result{Answer: "1234"},
				s2.String(): &Result{Answer: "5678"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "2 responded with records and 0 gave errors")
			So(text, ShouldContainSubstring, "1 servers responded with;\n1234\n\n")
			So(text, ShouldContainSubstring, "1 servers responded with;\n5678\n\n")
		})

		Convey("two of the same", func() {
			q.Results = QueryResults{
				s1.String(): &Result{Answer: "1234"},
				s2.String(): &Result{Answer: "1234"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "2 responded with records and 0 gave errors")
			So(text, ShouldContainSubstring, "2 servers responded with;\n1234\n\n")
		})
	})

	Convey("errors", t, func() {
		Convey("two different", func() {
			q.Results = QueryResults{
				s1.String(): &Result{Error: "1234"},
				s2.String(): &Result{Error: "5678"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "0 responded with records and 2 gave errors")
			So(text, ShouldContainSubstring, "\nAnd here are the errors;\n\n")
			So(text, ShouldContainSubstring, "1 servers responded with;\n1234\n\n")
			So(text, ShouldContainSubstring, "1 servers responded with;\n5678\n\n")
		})

		Convey("two of the same", func() {
			q.Results = QueryResults{
				s1.String(): &Result{Error: "1234"},
				s2.String(): &Result{Error: "1234"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "0 responded with records and 2 gave errors")
			So(text, ShouldContainSubstring, "\nAnd here are the errors;\n\n")
			So(text, ShouldContainSubstring, "2 servers responded with;\n1234\n\n")
		})
	})
}

func TestQuery_SetType(t *testing.T) {
	q := new(Query)

	Convey("setting a valid type works as expected", t, func() {
		err := q.SetType("A")
		So(err, ShouldBeNil)
		So(q.Type, ShouldEqual, dns.TypeA)

		Convey("check even with lower case types", func() {
			err := q.SetType("aaaa")
			So(err, ShouldBeNil)
			So(q.Type, ShouldEqual, dns.TypeAAAA)
		})
	})

	Convey("setting an invalid type throws an error", t, func() {
		err := q.SetType("FOO")
		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "FOO")
	})
}

func TestQuery_GetType(t *testing.T) {
	q := new(Query)

	Convey("get empty string when no type is set", t, func() {
		q.Type = dns.TypeNone
		t := q.GetType()
		So(t, ShouldEqual, "")
	})

	Convey("get correct type when one is set", t, func() {
		q.Type = dns.TypeA
		t := q.GetType()
		So(t, ShouldEqual, "A")
	})
}
