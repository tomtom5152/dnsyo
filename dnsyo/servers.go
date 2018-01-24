package dnsyo

import (
	"os"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/miekg/dns"
	"errors"
	"net"
	"time"
	"fmt"
	"math/rand"
	"sync"
	"strings"
	"syscall"
)

const (
	answerResult = 4
)

type Server struct {
	Ip       string
	Country  string
	Provider string
	Reverse  string
}

type ServerList []Server

type Results struct {
	SuccessCount, ErrorCount int
	Success                  map[string]int
	Errors                   map[string]int
	mutex sync.Mutex
}

func ServersFromFile(filename string) (sl ServerList, err error) {
	if filename[0] != '/' {
		pwd, _ := os.Getwd()
		filename = pwd + "/" + filename
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &sl)
	if err != nil {
		return nil, err
	}

	return
}

func (sl *ServerList) DumpToFile(filename string) (err error) {
	if filename[0] != '/' {
		pwd, _ := os.Getwd()
		filename = pwd + "/" + filename
	}

	yml, err := yaml.Marshal(sl)
	if err != nil {
		return
	}

	data := append([]byte("#### GENERATED BY dnsyo update ####\n\n"), yml...)

	err = ioutil.WriteFile(filename, data, 0744)
	return
}

func (sl *ServerList) FilterCountry(country string) (fl ServerList, err error){
	for _, s := range *sl {
		if s.Country == country {
			fl = append(fl, s)
		}
	}

	if len(fl) == 0 {
		err = errors.New(fmt.Sprintf("no servers matching country %s were found", country))
	}

	return
}

func (sl *ServerList) NRandom(n int) (rl ServerList, err error) {
	ql := *sl

	if len(ql) < n {
		return nil, errors.New("insufficient servers to populate list")
	}

	rl = make(ServerList, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(n)
	for i, randIndex := range perm[:n] {
		rl[i] = ql[randIndex]
	}

	return
}

func (sl *ServerList) Query(name string, recordType uint16, rate time.Duration) (r *Results) {
	var wg sync.WaitGroup
	r = new(Results)
	r.SuccessCount = 0
	r.ErrorCount = 0
	r.Success = make(map[string]int, 0)
	r.Errors = make(map[string]int, 0)
	limiter := time.Tick(rate)
	for _, s := range *sl {
		wg.Add(1)
		go func(s Server) {
			<-limiter
			defer wg.Done()
			result, err := s.Lookup(name, recordType)
			key := strings.Join(result, "\n")

			r.mutex.Lock()
			if err != nil {
				r.ErrorCount++
				r.Errors[err.Error()]++
			} else {
				r.SuccessCount++
				r.Success[key]++
			}
			r.mutex.Unlock()
		}(s)
	}

	wg.Wait()
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
			err =  errors.New("server did not return a result")
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
