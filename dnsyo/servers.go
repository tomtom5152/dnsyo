package dnsyo

import (
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
	"net/http"
	"github.com/gocarina/gocsv"
)

const (
	answerResult         = 4
	reliabilityThreshold = 0.97
)

type Server struct {
	Ip      string
	Country string
	Name    string
}

type csvNameserver struct {
	// IPAddress is the ipv4 address of the server
	IPAddress string `csv:"ip"`

	// Name is the hostname of the server if the server has a hostname
	Name string `csv:"name"`

	// Country is the two-letter ISO 3166-1 alpha-2 code of the country
	Country string `csv:"country_id"`

	// City specifies the city that the server is hosted on
	City string `csv:"city"`

	// Version is the software version of the dns daemon that the server is using
	Version string `csv:"version"`

	// Error is the error that the server returned. Probably will be empty if you use the valid nameserver dataset
	Error string `csv:"error"`

	// DNSSec is a boolean to indicate if the server supports DNSSec or not
	DNSSec bool `csv:"dnssec"`

	// Realiability is a normalized value - from 0.0 - 1.0 - to indicate how stable the server is
	Reliability float64 `csv:"reliability"`

	// CheckedAt is a timestamp to indicate the date that the server was last checked
	CheckedAt time.Time `csv:"checked_at"`

	// CreatedAt is a timestamp to indicate when the server was inserted in the database
	CreatedAt time.Time `csv:"created_at"`
}

type ServerList []Server

type Results struct {
	SuccessCount, ErrorCount int
	Success                  map[string]int
	Errors                   map[string]int
}

func ServersFromFile(filename string) (sl ServerList, err error) {
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

func ServersFromCSVURL(url string) (sl ServerList, err error) {
	csvFile, err := http.Get(url)
	if err != nil {
		return
	}

	// sometimes the file can be truncated and have an incomplete final line.
	// run a basic check to ensure it isn't going to error later.
	bytes, err := ioutil.ReadAll(csvFile.Body)
	if err != nil {
		return
	}

	rows := strings.Split(string(bytes), "\n")
	nCols := strings.Count(rows[0], ",")
	if strings.Count(rows[len(rows)-1], ",") < nCols {
		rows = rows[:len(rows)-2]
	}
	data := strings.Join(rows, "\n")

	var servers []csvNameserver
	err = gocsv.Unmarshal(strings.NewReader(data), &servers)
	if err != nil {
		return
	}

	for _, ns := range servers {
		if ip := net.ParseIP(ns.IPAddress); ip.To4() == nil {
			continue // we can't process IPv6 yet
		}
		if ns.Reliability >= reliabilityThreshold {
			s := Server{
				Ip:      ns.IPAddress,
				Country: strings.ToUpper(ns.Country),
				Name:    ns.Name,
			}
			sl = append(sl, s)
		}
	}

	return
}

func (sl *ServerList) DumpToFile(filename string) (err error) {
	yml, err := yaml.Marshal(sl)
	if err != nil {
		return
	}

	data := append([]byte("#### GENERATED BY dnsyo update ####\n\n"), yml...)

	err = ioutil.WriteFile(filename, data, 0744)
	return
}

func (sl *ServerList) FilterCountry(country string) (fl ServerList, err error) {
	for _, s := range *sl {
		if s.Country == strings.ToUpper(country) {
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
		return nil, errors.New(fmt.Sprintf("insufficient servers to populate list: %d of %d", len(ql), n))
	}

	rl = make(ServerList, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(n)
	for i, randIndex := range perm[:n] {
		rl[i] = ql[randIndex]
	}

	return
}

func (sl *ServerList) Query(name string, recordType uint16, threads int) (r *Results) {
	var wg sync.WaitGroup
	r = new(Results)
	r.SuccessCount = 0
	r.ErrorCount = 0
	r.Success = make(map[string]int, 0)
	r.Errors = make(map[string]int, 0)
	var mtx sync.Mutex

	queue := make(chan Server, len(*sl))

	// start workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for s := range queue {
				res, err := s.Lookup(name, recordType)
				key := strings.Join(res, "\n")

				mtx.Lock()
				if err != nil {
					r.ErrorCount++
					r.Errors[err.Error()]++
				} else {
					r.SuccessCount++
					r.Success[key]++
				}
				mtx.Unlock()
			}

		}(i)
	}

	for _, s := range *sl {
		queue <- s
	}
	close(queue)

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
