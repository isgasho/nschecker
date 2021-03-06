package dnsmsg

import (
	"errors"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"strings"
)

// DNS gTLD servers Array.
func getTldServers() map[string]string {
	servers := make(map[string]string, 3)
	servers["com"] = "m.gtld-servers.net"
	servers["net"] = "m.gtld-servers.net"
	servers["jp"] = "a.dns.jp"
	return servers
}

func Lookup(domainName string) ([]string, error) {
	server, err := getAuthorityServerName(domainName)
	if err != nil {
		fmt.Println(" ..lookup from local DNS cache server.\n")
		return lookupFromDnsCacheServer(domainName)
	}
	fmt.Printf(" ..lookup from DNS root server: %s \n\n", server)
	return lookupFromDnsRoot(domainName, server)
}

// get DNS server for looking up name from DNS root servers.
func getAuthorityServerName(domainName string) (string, error) {
	labels := strings.Split(domainName, ".")

	// supports 2nd/3rd level domain only
	if len(labels) == 1 || len(labels) >= 4 {
		return "", errors.New("not support domain level.")
	}
	gTldLabel := labels[len(labels)-1] //get last value in array
	servers := getTldServers()
	server, ok := servers[gTldLabel]
	if !ok {
		return "", errors.New("not support TLD name.")
	}
	return server, nil
}

func lookupFromDnsCacheServer(domainName string) ([]string, error) {
	var ret []string

	nss, err := net.LookupNS(domainName)
	if err != nil {
		return nil, errors.New("NS Lookup Error.\n")
	}
	for _, ns := range nss {
		ret = append(ret, ns.Host)
	}
	return ret, nil
}

func lookupFromDnsRoot(domainName string, dnsServer string) ([]string, error) {
	var ret []string

	id := NewId()
	h := NewHeader(id)
	header := NewHeaderPayload(h)
	q := NewQuestion(domainName, TypeNS)
	question := NewQuestionPayload(q)
	req := append(header, question...)

	// sending UDP
	buf, err := Send(dnsServer+":53", req)
	if err != nil {
		return nil, err
	}

	var m dnsmessage.Message
	err = m.Unpack(buf)
	if err != nil {
		return nil, err
	}

	if len(m.Authorities) == 0 {
		return nil, errors.New("NS Lookup Error. No NS servers from DNS root.\n")
	}

	for _, authrotiy := range m.Authorities {
		rr := authrotiy.Body.(*dnsmessage.NSResource)
		ret = append(ret, rr.NS.String())
	}
	return ret, nil
}
