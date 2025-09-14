package resolver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func fallbackResolve(q dns.Question, addresses []string, logger *slog.Logger) []dns.RR {
	cli := new(dns.Client)
	for _, address := range addresses {
		m := new(dns.Msg)
		m.SetQuestion(q.Name, q.Qtype)
		logger.Debug("Fallback resolving", "name", q.Name, "qtype", q.Qtype, "address", address)
		answer, _, answerErr := cli.Exchange(m, address)
		if answerErr != nil {
			logger.Debug("Fallback DNS query failed", "address", address, "err", answerErr)
			continue
		}
		if len(answer.Answer) == 0 {
			logger.Debug("No answer from fallback DNS", "address", address, "name", q.Name)
			continue
		}
		logger.Debug("Fallback DNS returned answers", "address", address, "count", len(answer.Answer), "name", q.Name)
		return answer.Answer
	}
	return nil
}

func resolveContainers(mapping HostnameIPMapping, r *dns.Msg, fallbackDns []string, logger *slog.Logger) []dns.RR {
	var answers []dns.RR
	for _, q := range r.Question {
		if ip, ok := mapping[q.Name[:len(q.Name)-1]]; ok {
		logger.Debug("Resolving question", "name", q.Name, "qtype", q.Qtype)
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
				logger.Info("Resolved from container mapping", "name", q.Name, "ip", ip)
				answers = append(answers, rr)
			default:
				logger.Debug("Unsupported query type", "qtype", q.Qtype, "name", q.Name)
			}
		} else {
			logger.Debug("No mapping, using fallback DNS", "name", q.Name)
			defAnsw := fallbackResolve(q, fallbackDns, logger)
			if len(defAnsw) != 0 {
				answers = append(answers, defAnsw...)
			} else {
				logger.Info("Fallback DNS failed to resolve", "name", q.Name)
			}
		}
	}
	return answers
}

type DNSHandler struct {
	inspector   Inspector
	filter      Filter
	fallbackDns []string
	logger      *slog.Logger
}

func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	h.logger.Info("Received DNS query", "questions", len(r.Question))
	mapping, mappingErr := h.inspector.GetContainerMapping(context.Background(), h.filter)
	if mappingErr != nil {
		h.logger.Error("Failed to get container mapping", "err", mappingErr)
		mapping = make(HostnameIPMapping)
	}

	answers := resolveContainers(mapping, r, h.fallbackDns, h.logger)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Answer = append(m.Answer, answers...)
	if err := w.WriteMsg(m); err != nil {
		h.logger.Error("Failed to write DNS response", "err", err)
	}
}

func NewDNSHandler(filter Filter, fallbackDns []string, inspector Inspector, logger *slog.Logger) (*DNSHandler, error) {
	return &DNSHandler{
		inspector:   inspector,
		filter:      filter,
		fallbackDns: fallbackDns,
		logger:      logger,
	}, nil
}

func Serve(port uint16, filter Filter, fallbackDns []string, inspector Inspector, logger *slog.Logger) error {
	handler, handlerErr := NewDNSHandler(filter, fallbackDns, inspector, logger)
	if handlerErr != nil {
		return handlerErr
	}
	server := &dns.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Net:     "udp",
		Handler: handler,
	}
	logger.Info("Starting DNS server", "port", port)
	return server.ListenAndServe()
}
