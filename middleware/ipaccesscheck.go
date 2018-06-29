package middleware

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

//TODO this will be replaced as per issue ecad/auth#29

// IPAcessChecker is a middleware that provides checks on the Requesting IP
// againse a list of approved IP CIDR ranges.
type IPAcessChecker struct {
	approvedCIDRs []*net.IPNet
}

func NewIPAccessChecker(approvedIPStrings []string) (*IPAcessChecker, error) {
	if len(approvedIPStrings) == 0 {
		return nil, errors.New("no approved addresses provided")
	}

	checker := IPAcessChecker{}
	for _, ipString := range approvedIPStrings {
		_, ip, err := net.ParseCIDR(ipString)
		if err != nil {
			return nil, fmt.Errorf("Error parsing approved CIDR %s", ipString)
		}
		checker.approvedCIDRs = append(checker.approvedCIDRs, ip)

	}
	return &checker, nil
}

func (c *IPAcessChecker) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := extractIP(r)
		if err != nil {
			log.Printf("Error getting requestos IP address: %e", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !c.Contains(ip) {
			log.Printf("Deny IP: %v, not it access list", ip)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
func extractIP(r *http.Request) (net.IP, error) {

	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip := net.ParseIP(forwardedFor)
		if ip == nil {
			return nil, fmt.Errorf("can't parse IP from address %s", ip)
		}
		return ip, nil

	}

	ipAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return nil, fmt.Errorf("can't parse IP from address %s", ip)
	}
	return ip, nil
}

func (c *IPAcessChecker) Contains(ip net.IP) bool {
	for _, i := range c.approvedCIDRs {
		if i.Contains(ip) {
			return true
		}
	}
	return false
}
