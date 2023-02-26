package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"go.uber.org/zap"
)

func main() {
	var (
		dnsResolverIP        = "127.0.0.1:1053" // CoreDNS resolver.
		dnsResolverProto     = "udp"            // Protocol to use for the DNS resolver
		dnsResolverTimeoutMs = 500              // Timeout (ms) for the DNS resolver
		dialTimeoutMs        = 3000             // Timeout (ms) for the dial
	)

	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger := l.Sugar()

	dialer := &net.Dialer{
		Timeout: time.Duration(dialTimeoutMs) * time.Millisecond,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return dialer.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	tr := http.DefaultTransport
	tr.(*http.Transport).DialContext = dialContext

	httpClient := &http.Client{
		Transport: tr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// test HTTP client with the custom DNS resolver.
	req, err := http.NewRequestWithContext(ctx, "GET", "http://www.example.io:5001", nil)
	if err != nil {
		log.Fatalln(err)
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			logger.Debugf("[trace] Got Conn: %+v", connInfo)
		},
		ConnectDone: func(network, addr string, err error) {
			logger.Debugf("[trace] Conn done, addr, err : %+v, %v", addr, err)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			logger.Debugf("[trace] DNS Info: %+v", dnsInfo)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Debugw("get response body", "data", string(body))
}
