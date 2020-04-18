// +build !windows

package proxyplease

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/git-lfs/go-ntlm/ntlm"
)

const (
	negotiateUnicode    = 0x0001 // Text strings are in unicode
	negotiateOEM        = 0x0002 // Text strings are in OEM
	requestTarget       = 0x0004 // Server return its auth realm
	negotiateSign       = 0x0010 // Request signature capability
	negotiateSeal       = 0x0020 // Request confidentiality
	negotiateLMKey      = 0x0080 // Generate session key
	negotiateNTLM       = 0x0200 // NTLM authentication
	negotiateLocalCall  = 0x4000 // client/server on same machine
	negotiateAlwaysSign = 0x8000 // Sign for all security levels
)

var (
	put32     = binary.LittleEndian.PutUint32
	put16     = binary.LittleEndian.PutUint16
	encBase64 = base64.StdEncoding.EncodeToString
	decBase64 = base64.StdEncoding.DecodeString
)

// generates NTLM Negotiate type-1 message
// for details see http://www.innovation.ch/personal/ronald/ntlm.html
func negotiateNTLMv1Message() []byte {
	ret := make([]byte, 44)
	flags := negotiateAlwaysSign | negotiateNTLM | requestTarget | negotiateOEM | negotiateUnicode

	copy(ret, []byte("NTLMSSP\x00")) // protocol
	put32(ret[8:], 1)                // type
	put32(ret[12:], uint32(flags))   // flags
	put16(ret[16:], 0)               // NT domain name length
	put16(ret[18:], 0)               // NT domain name max length
	put32(ret[20:], 0)               // NT domain name offset
	put16(ret[24:], 0)               // local workstation name length
	put16(ret[26:], 0)               // local workstation name max length
	put32(ret[28:], 0)               // local workstation name offset
	put16(ret[32:], 0)               // unknown name length
	put16(ret[34:], 0)               // ...
	put16(ret[36:], 0x30)            // unknown offset
	put16(ret[38:], 0)               // unknown name length
	put16(ret[40:], 0)               // ...
	put16(ret[42:], 0x30)            // unknown offset

	return ret
}

func dialNTLM(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	debugf("ntlm> Attempting to authenticate")

	conn, err := baseDial()
	if err != nil {
		debugf("ntlm> Could not call dial context with proxy: %s", err)
		return conn, err
	}

	h := p.Headers.Clone()
	h.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiateNTLMv1Message())))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: h,
	}
	if err := connect.Write(conn); err != nil {
		debugf("ntlm> Could not write negotiate message to proxy: %s", err)
		return conn, err
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connect)
	if err != nil {
		debugf("ntlm> Could not read negotiate response from proxy: %s", err)
		return conn, err
	}

	if resp.StatusCode != http.StatusProxyAuthRequired {
		debugf("ntlm> Expected %d as return status, got: %d", http.StatusProxyAuthRequired, resp.StatusCode)
		return conn, errors.New("Unexpected HTTP status code")
	}

	session, err := ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionlessMode)
	if err != nil {
		debugf("ntlm> Error creating client session")
		return conn, err
	}

	session.SetUserInfo(p.Username, p.Password, p.Domain)

	challengeBytes, err := base64.StdEncoding.DecodeString(resp.Header["Proxy-Authenticate"][0][5:])
	if err != nil {
		debugf("ntlm> Could not read challenge response")
		return conn, err
	}

	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	if err != nil {
		debugf("ntlm> Error parsing challenge message")
		return conn, err
	}

	err = session.ProcessChallengeMessage(challenge)
	if err != nil {
		debugf("ntlm> Error processing challenge messgage")
		return conn, err
	}

	authenticate, err := session.GenerateAuthenticateMessage()
	if err != nil {
		debugf("ntlm> Error generating  authenticate message")
		return conn, err
	}

	resp.Body.Close()
	h = p.Headers.Clone()
	h.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticate.Bytes())))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect = &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: h,
	}
	if err := connect.Write(conn); err != nil {
		debugf("ntlm> Could not write authenticate message to proxy: %s", err)
		return conn, err
	}
	br = bufio.NewReader(conn)
	resp, err = http.ReadResponse(br, connect)
	if err != nil {
		debugf("ntlm> Could not read authenticate response from proxy: %s", err)
		return conn, err
	}

	if resp.StatusCode == http.StatusOK {
		debugf("ntlm> Successfully injected NTLM to connection")
		return conn, nil
	}

	debugf("ntlm> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
	return conn, errors.New(http.StatusText(resp.StatusCode))

}
