package dnsyo

import (
	"errors"
	"net"
	"syscall"
	"strings"
	"github.com/miekg/dns"
)

type Server struct {
	Ip      string
	Country string
	Name    string
}

func (s *Server) Test() (ok bool, err error) {
	tests := []dns.Question{
		{dns.Fqdn("google.com"), dns.TypeA, dns.ClassINET},
		{dns.Fqdn("facebook.com"), dns.TypeA, dns.ClassINET},
		{dns.Fqdn("amazon.com"), dns.TypeA, dns.ClassINET},
	}

	addr := s.Ip + ":53"
	c := new(dns.Client)
	var lastErr error

	for _, q := range tests {
		msg := new(dns.Msg)
		msg.Id = dns.Id()
		msg.RecursionDesired = true
		msg.Question = make([]dns.Question, 1)
		msg.Question[0] = q

		resp, _, err := c.Exchange(msg, addr)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				// instant fail
				return false, errors.New("TIMEOUT")
			}
			if lastErr != nil && err.Error() == lastErr.Error() {
				return false, err
			}
			lastErr = err
			continue
		}

		if resp == nil {
			err = errors.New("server did not return a result")
			if lastErr != nil && err.Error() == lastErr.Error() {
				return false, err
			}
			lastErr = err
			continue
		}
	}

	return true, nil
}

func (s *Server) Lookup(name string, recordType uint16) (results []string, err error) {
	addr := s.Ip + ":53"
	c := new(dns.Client)

	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{
		Name:   dns.Fqdn(name),
		Qtype:  recordType,
		Qclass: dns.ClassINET,
	}

	resp, _, err := c.Exchange(msg, addr)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, errors.New("TIMEOUT")
		}

		switch t := err.(type) {
		case *net.OpError:
			if t.Op == "read" {
				err = errors.New("CONNECTION REFUSED")
			}

		case syscall.Errno:
			switch t {
			case syscall.ECONNREFUSED:
				err = errors.New("CONNECTION REFUSED")
			}
		}

		return
	}

	if resp.Rcode != dns.RcodeSuccess {
		return nil, errors.New(dns.RcodeToString[resp.Rcode])
	}

	if len(resp.Answer) == 0 {
		return nil, errors.New("NOANSWER")
	}

	for _, rr := range resp.Answer {
		results = append(results, strings.Split(rr.String(), "\t")[answerResult])
	}

	return
}

func (s *Server) String() (string) {
	if s.Name != "" {
		return s.Name
	}
	return s.Ip
}