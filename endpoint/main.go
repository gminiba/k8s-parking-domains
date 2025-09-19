package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var ourNS map[string]struct{}

func init() {
	// Read comma-separated NS list from environment variable
	nsEnv := os.Getenv("OUR_NS")
	if nsEnv == "" {
		log.Fatal("OUR_NS environment variable not set")
	}

	ourNS = make(map[string]struct{})
	for _, ns := range strings.Split(nsEnv, ",") {
		// Normalize: lowercase and ensure trailing dot
		ns = strings.ToLower(strings.TrimSpace(ns))
		if !strings.HasSuffix(ns, ".") {
			ns += "."
		}
		ourNS[ns] = struct{}{}
	}

	log.Printf("Allowed nameservers: %v", ourNS)
}

var resolvers []string

func main() {
	// Load nameservers from env
	nsEnv := os.Getenv("NAMESERVERS")
	if nsEnv == "" {
		// Fallback if not set
		nsEnv = "1.1.1.1,8.8.8.8"
	}
	resolvers = splitAndTrim(nsEnv)
	log.Printf("Using nameservers: %v\n", resolvers)

	// Load port from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	http.HandleFunc("/check-host", checkHostHandler)

	log.Printf("Ask endpoint running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// splitAndTrim splits comma-separated values and trims spaces
func splitAndTrim(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func checkHostHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "missing domain", http.StatusBadRequest)
		return
	}

	// Ensure trailing dot for DNS query
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeNS)

	in, err := dns.Exchange(m, "8.8.8.8:53") // use public resolver (or configurable)
	if err != nil {
		http.Error(w, "dns lookup failed", http.StatusForbidden)
		return
	}

	var found bool
	for _, ans := range in.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nsName := strings.ToLower(ns.Ns)
			if _, ok := ourNS[nsName]; ok {
				found = true
				break
			}
		}
	}

	if found {
		fmt.Fprintln(w, "ok")
	} else {
		http.Error(w, "forbidden", http.StatusForbidden)
	}
}
