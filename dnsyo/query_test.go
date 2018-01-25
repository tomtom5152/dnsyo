package dnsyo

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
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
		Type: "FOO",
	}

	Convey("top block is formatted correctly", t, func() {
		text := q.ToTextSummary()
		So(text, ShouldStartWith, `
 - RESULTS
I asked 0 servers for FOO records related to example.test,
`)
	})

	Convey("successes", t, func() {
		Convey("two different", func() {
			q.Results = QueryResults{
				s1: &Result{Answer: "1234"},
				s2: &Result{Answer: "5678"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "2 responded with records and 0 gave errors")
			So(text, ShouldContainSubstring, "1 servers responded with;\n1234\n\n")
			So(text, ShouldContainSubstring, "1 servers responded with;\n5678\n\n")
		})

		Convey("two of the same", func() {
			q.Results = QueryResults{
				s1: &Result{Answer: "1234"},
				s2: &Result{Answer: "1234"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "2 responded with records and 0 gave errors")
			So(text, ShouldContainSubstring, "2 servers responded with;\n1234\n\n")
		})
	})

	Convey("errors", t, func() {
		Convey("two different", func() {
			q.Results = QueryResults{
				s1: &Result{Error: "1234"},
				s2: &Result{Error: "5678"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "0 responded with records and 2 gave errors")
			So(text, ShouldContainSubstring, "\nAnd here are the errors;\n\n")
			So(text, ShouldContainSubstring, "1 servers responded with;\n1234\n\n")
			So(text, ShouldContainSubstring, "1 servers responded with;\n5678\n\n")
		})

		Convey("two of the same", func() {
			q.Results = QueryResults{
				s1: &Result{Error: "1234"},
				s2: &Result{Error: "1234"},
			}

			text := q.ToTextSummary()
			So(text, ShouldContainSubstring, "0 responded with records and 2 gave errors")
			So(text, ShouldContainSubstring, "\nAnd here are the errors;\n\n")
			So(text, ShouldContainSubstring, "2 servers responded with;\n1234\n\n")
		})
	})
}
