package dnsyo

import (
	"os"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/miekg/dns"
	"errors"
	"net"
)

type Server struct {
	Ip       string
	Country  string
	Provider string
	Reverse  string
}

type ServerList []Server

func ServersFromFile(filename string) (sl ServerList, err error) {
	pwd, _ := os.Getwd()
	data, err := ioutil.ReadFile(pwd + "/" + filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &sl)
	if err != nil {
		return nil, err
	}

	return
}

func (s *Server) Test() (ok bool, err error) {
	tests := []dns.Question{
		{dns.Fqdn("google.com"), dns.TypeA, dns.ClassINET},
		{dns.Fqdn("facebook.com"), dns.TypeA, dns.ClassINET},
		{dns.Fqdn("amazon.com"), dns.TypeA, dns.ClassINET},
	}

	addr := s.Ip + ":53"
	c := new(dns.Client)

	for _, q := range tests {
		msg := new(dns.Msg)
		msg.Id = dns.Id()
		msg.RecursionDesired = true
		msg.Question = make([]dns.Question, 1)
		msg.Question[0] = q

		resp, _, err := c.Exchange(msg, addr)
		if err != nil {
			return false, err
		}

		if resp == nil {
			return false, errors.New("server did not return a result")
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
		Name: dns.Fqdn(name),
		Qtype: recordType,
		Qclass: dns.ClassINET,
	}

	resp, _, err := c.Exchange(msg, addr)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, errors.New("TIMEOUT")
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
		results = append(results, rr.String())
	}

	return
}
