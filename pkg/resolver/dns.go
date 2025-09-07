package resolver

import (
	"context"
	"fmt"
	"net"

	"github.com/miekg/dns"
)

func fallbackResolve(q dns.Question, addresses []string) []dns.RR {
	cli := new(dns.Client)
	for _, address := range addresses {
		m := new(dns.Msg)
		m.SetQuestion(q.Name, q.Qtype)
		answer, _, answerErr := cli.Exchange(m, address)
		if answerErr != nil || len(answer.Answer) == 0 {
			continue
		}

		return answer.Answer
	}

	return nil
}

func resolveContainers(mapping HostnameIPMapping, r *dns.Msg, fallbackDns []string) []dns.RR {
	var answers []dns.RR
	for _, q := range r.Question {
		if ip, ok := mapping[q.Name[:len(q.Name)-1]]; ok {
			switch q.Qtype {
			case dns.TypeA:
				rr := &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
					},
					A: net.ParseIP(ip),
				}
				answers = append(answers, rr)
			}
		} else {
			defAnsw := fallbackResolve(q, fallbackDns)
			if len(defAnsw) != 0 {
				answers = append(answers, defAnsw...)
			}
		}
	}
	return answers
}

type DNSHandler struct {
	inspector   Inspector
	filter      Filter
	fallbackDns []string
}

func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	mapping, mappingErr := h.inspector.GetContainerMapping(context.Background(), h.filter)
	if mappingErr != nil {
		mapping = make(HostnameIPMapping)
	}

	answers := resolveContainers(mapping, r, h.fallbackDns)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Answer = append(m.Answer, answers...)
	w.WriteMsg(m)
}

func NewDNSHandler(filter Filter, fallbackDns []string, inspector Inspector) (*DNSHandler, error) {
	return &DNSHandler{
		inspector:   inspector,
		filter:      filter,
		fallbackDns: fallbackDns,
	}, nil
}

func Serve(port uint16, filter Filter, fallbackDns []string, inspector Inspector) error {
	handler, handlerErr := NewDNSHandler(filter, fallbackDns, inspector)
	if handlerErr != nil {
		return handlerErr
	}
	server := &dns.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Net:     "udp",
		Handler: handler,
	}
	return server.ListenAndServe()
}
