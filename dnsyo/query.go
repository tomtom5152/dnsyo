package dnsyo

import (
	"fmt"
	"github.com/miekg/dns"
	"errors"
	"strings"
)

type resultSummary struct {
	SuccessCount, ErrorCount int
	Answers                  map[string]int
	Errors                   map[string]int
}

type Query struct {
	Results QueryResults
	Domain  string
	Type    uint16
}

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

func (q *Query) SetType(recordType string) error {
	recordType = strings.ToUpper(recordType)
	t, ok := dns.StringToType[recordType]
	if !ok {
		return errors.New(fmt.Sprintf("unable to use record type %s", recordType))
	}
	q.Type = t
	return nil
}

func (q *Query) GetType() string {
	if q.Type != 0 {
		return dns.TypeToString[q.Type]
	}
	return ""
}
