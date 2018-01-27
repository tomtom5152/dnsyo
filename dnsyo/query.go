package dnsyo

import (
	"fmt"
	"github.com/miekg/dns"
	"strings"
)

type resultSummary struct {
	SuccessCount, ErrorCount int
	Answers                  map[string]int
	Errors                   map[string]int
}

// Query represents a lookup of Type for a given Domain and stores the Results for later processing
type Query struct {
	Results QueryResults
	Domain  string
	Type    uint16
}

// ToTextSummary prints a human readable output of the current query's results for use in the CLI.
func (q *Query) ToTextSummary() (text string) {
	var rs resultSummary
	rs.Answers = make(map[string]int)
	rs.Errors = make(map[string]int)

	for _, r := range q.Results {
		if r.Error == "" && r.Answer != "" {
			rs.SuccessCount++
			rs.Answers[r.Answer]++
		} else {
			rs.ErrorCount++
			rs.Errors[r.Error]++
		}
	}

	text = fmt.Sprintf(`
 - RESULTS
I asked %d servers for %s records related to %s,
%d responded with records and %d gave errors
Here are the results;`, len(q.Results), q.GetType(), q.Domain, rs.SuccessCount, rs.ErrorCount)
	text += "\n\n\n"

	if rs.SuccessCount > 0 {
		for result, count := range rs.Answers {
			text += fmt.Sprintf("%d servers responded with;\n%s\n\n", count, result)
		}
	}

	if rs.ErrorCount > 0 {
		text += fmt.Sprint("\nAnd here are the errors;\n\n")

		for err, count := range rs.Errors {
			text += fmt.Sprintf("%d servers responded with;\n%s\n\n", count, err)
		}
	}

	return text
}

// SetType converts a string representation of a query type to the internal uint16. This is then set on the current Query.
// An error is returned if the type cannot be found in the miekg/dns library.
func (q *Query) SetType(recordType string) error {
	recordType = strings.ToUpper(recordType)
	t, ok := dns.StringToType[recordType]
	if !ok {
		return fmt.Errorf("unable to use record type %s", recordType)
	}
	q.Type = t
	return nil
}

// GetType looks up the current Query's uint16 Type and returns the string representation of it from the miekg/dns library.
func (q *Query) GetType() string {
	if q.Type != 0 {
		return dns.TypeToString[q.Type]
	}
	return ""
}
