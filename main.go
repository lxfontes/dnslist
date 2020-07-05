package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/coredns/coredns/pb"
	"github.com/miekg/dns"
	"google.golang.org/grpc"
)

const (
	bindAddr = ":3030"
)

type server struct {
}

func (s *server) Query(ctx context.Context, msg *pb.DnsPacket) (*pb.DnsPacket, error) {
	m := new(dns.Msg)
	if err := m.Unpack(msg.Msg); err != nil {
		return nil, fmt.Errorf("failed to unpack msg: %v", err)
	}
	r := new(dns.Msg)
	r.SetReply(m)
	r.Authoritative = true

	// TODO: query a database and provide real answers here!
	for _, q := range r.Question {
		hdr := dns.RR_Header{Name: q.Name, Rrtype: q.Qtype, Class: q.Qclass}
		switch q.Qtype {
		case dns.TypeA:
			r.Answer = append(r.Answer, &dns.A{Hdr: hdr, A: net.IPv4(127, 0, 0, 1)})
		case dns.TypeAAAA:
			r.Answer = append(r.Answer, &dns.AAAA{Hdr: hdr, AAAA: net.IPv6loopback})
		default:
			return nil, fmt.Errorf("only A/AAAA supported, got qtype=%d", q.Qtype)
		}
	}

	if len(r.Answer) == 0 {
		r.Rcode = dns.RcodeNameError
	}

	out, err := r.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack msg: %v", err)
	}
	return &pb.DnsPacket{Msg: out}, nil
}

func main() {
	lis, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDnsServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
