// ==========================================================
// VERSÃO FINAL DEFINITIVA: client/roundtripper.go
// Totalmente reescrito para compatibilidade com as bibliotecas de 2025.
// ==========================================================
package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/fhttp/http2"
	"golang.org/x/net/proxy"

	utls "github.com/refraction-networking/utls"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type roundTripper struct {
	sync.Mutex

	JA3       string
	UserAgent string

	Cookies           []Cookie
	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper

	dialer proxy.ContextDialer
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, properties := range rt.Cookies {
		req.AddCookie(&http.Cookie{
			Name:    properties.Name,
			Value:   properties.Value,
			Path:    properties.Path,
			Domain:  properties.Domain,
			Expires: properties.JSONExpires.Time,
		})
	}
	req.Header.Set("User-Agent", rt.UserAgent)
	addr := rt.getDialTLSAddr(req)

	rt.Lock()
	if _, ok := rt.cachedTransports[addr]; !ok {
		rt.Unlock()
		if err := rt.getTransport(req, addr); err != nil {
			return nil, err
		}
	} else {
		rt.Unlock()
	}

	return rt.cachedTransports[addr].RoundTrip(req)
}

func (rt *roundTripper) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports[addr] = &http.Transport{DialContext: rt.dialer.DialContext}
		return nil
	case "https":
	default:
		return fmt.Errorf("invalid URL scheme: [%v]", req.URL.Scheme)
	}

	_, err := rt.dialTLS(context.Background(), "tcp", addr)
	if err == errProtocolNegotiated {
		return nil
	}
	return err
}

func (rt *roundTripper) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	rt.Lock()
	defer rt.Unlock()

	if conn := rt.cachedConnections[addr]; conn != nil {
		delete(rt.cachedConnections, addr)
		return conn, nil
	}
	rawConn, err := rt.dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	var host string
	if host, _, err = net.SplitHostPort(addr); err != nil {
		host = addr
	}

	conn := utls.UClient(rawConn, &utls.Config{ServerName: host, InsecureSkipVerify: true}, utls.HelloChrome_120)
	if err = conn.Handshake(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("uTlsConn.Handshake() error: %w", err)
	}

	if rt.cachedTransports[addr] != nil {
		return conn, nil
	}

	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		// ESTA É A CORREÇÃO PRINCIPAL
		// A nova biblioteca http2 espera uma função DialTLS com assinatura utls.Config.
		// Nós criamos uma função anônima que chama a nossa função principal, garantindo a compatibilidade.
		t2 := http2.Transport{
			DialTLS: func(network, addr string, cfg *utls.Config) (net.Conn, error) {
				return rt.dialTLS(context.Background(), network, addr)
			},
			// E o TLSClientConfig também precisa ser do tipo utls.Config
			TLSClientConfig: &utls.Config{InsecureSkipVerify: true},
		}
		rt.cachedTransports[addr] = &t2
	default:
		// Para HTTP/1.1, usamos a função dialTLSContext para manter a compatibilidade
		rt.cachedTransports[addr] = &http.Transport{DialTLSContext: rt.dialTLS}
	}

	rt.cachedConnections[addr] = conn
	return nil, errProtocolNegotiated
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443")
}

func newRoundTripper(browser Browser, dialer ...proxy.ContextDialer) http.RoundTripper {
	if len(dialer) > 0 {
		return &roundTripper{
			dialer:            dialer[0],
			JA3:               browser.JA3,
			UserAgent:         browser.UserAgent,
			Cookies:           browser.Cookies,
			cachedTransports:  make(map[string]http.RoundTripper),
			cachedConnections: make(map[string]net.Conn),
		}
	}

	return &roundTripper{
		dialer:            proxy.Direct,
		JA3:               browser.JA3,
		UserAgent:         browser.UserAgent,
		Cookies:           browser.Cookies,
		cachedTransports:  make(map[string]http.RoundTripper),
		cachedConnections: make(map[string]net.Conn),
	}
}
