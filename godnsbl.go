/*
Package godnsbl lets you perform RBL (Real-time Blackhole List - https://en.wikipedia.org/wiki/DNSBL)
lookups using Golang

JSON annotations on the types are provided as a convenience.
*/
package godnsbl

import (
	"fmt"
	"net"
	"strings"
        "regexp"
)

/*
Blacklists is the list of blackhole lists to check against
*/
var Blacklists = []string{
	"bl.spamcop.net",
	"cbl.abuseat.org",
	"recent.spam.dnsbl.sorbs.net",
        "dnsbl-1.uceprotect.net",
        "spamrbl.imp.ch",
	"bl.score.senderscore.com",
	"cidr.bl.mcafee.com",
        "bl.mailspike.net"}

/*
RBLResults holds the results of the lookup.
*/
type RBLResults struct {
	// List is the RBL that was searched
	List string `json:"list"`
	// Host is the host or IP that was passed (i.e. smtp.gmail.com)
	Host string `json:"host"`
	// Results is a slice of Results - one per IP address searched
	Results []Result `json:"results"`
}

/*
Result holds the individual IP lookup results for each RBL search
*/
type Result struct {
        //rbl domain
        Rbl string `json:"rbl"`
	// Address is the IP address that was searched
	Address string `json:"address"`
	// Listed indicates whether or not the IP was on the RBL
	Listed bool `json:"listed"`
	// RBL lists sometimes add extra information as a TXT record
	// if any info is present, it will be stored here.
	Text string `json:"text"`
	// Error represents any error that was encountered (DNS timeout, host not
	// found, etc.) if any
	Error bool `json:"error"`
	// ErrorType is the type of error encountered if any
	ErrorType error `json:"error_type"`
}

/*
Reverse the octets of a given IPv4 address
64.233.171.108 becomes 108.171.233.64
*/
func Reverse(ip net.IP) string {
	if ip.To4() == nil {
		return ""
	}

	splitAddress := strings.Split(ip.String(), ".")

	for i, j := 0, len(splitAddress)-1; i < len(splitAddress)/2; i, j = i+1, j-1 {
		splitAddress[i], splitAddress[j] = splitAddress[j], splitAddress[i]
	}

	return strings.Join(splitAddress, ".")
}

func query(rbl string, host string, r *Result) {
	r.Listed = false
        r.Rbl = rbl

	lookup := fmt.Sprintf("%s.%s", host, rbl)

	res, err := net.LookupHost(lookup)
	if len(res) > 0 {
		
                for _, ip := range res{
                    m, _ := regexp.MatchString("^127.0.0.*", ip)
                    
                    if m == true {
                        r.Listed = true
                    }
                }

		txt, _ := net.LookupTXT(lookup)
		if len(txt) > 0 {
			r.Text = txt[0]
		}
	}
	if err != nil {
		r.Error = true
		r.ErrorType = err
	}

	return
}

/*
Lookup performs the search and returns the RBLResults
*/
func Lookup(rblList string, targetHost string) (r RBLResults) {
	r.List = rblList
	r.Host = targetHost

	if ip, err := net.LookupIP(targetHost); err == nil {
		for _, addr := range ip {
			if addr.To4() != nil {
				res := Result{}
				res.Address = addr.String()

				addr := Reverse(addr)

				query(rblList, addr, &res)

				r.Results = append(r.Results, res)
			}
		}
	} else {
		r.Results = append(r.Results, Result{})
	}
	return
}
