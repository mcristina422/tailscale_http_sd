package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"inet.af/netaddr"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
)

var EnabledHosts arrayFlags

func main() {
	flag.Var(&EnabledHosts, "host", "The hostnames of the peers added to the target.")
	flag.Parse()

	fmt.Printf("Looking for hosts: %s\n", EnabledHosts)
	fmt.Println("Serving Tailscale status at :8773")

	http.HandleFunc("/prometheus", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		g := []TargetGroup{}
		status, err := tailscale.Status(context.Background())
		if err != nil {
			fmt.Println(fmt.Errorf("error getting status: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, p := range status.Peer {
			if p == nil {
				continue
			}
			if !peerEnabled(p) {
				continue
			}
			target := TargetGroup{
				Targets: []string{fmt.Sprintf("%s:%s", firstIPString(p.TailscaleIPs), "9100")}, //TODO get all addrs?
				Labels:  map[string]string{"hostname": p.HostName},                             //TODO more?
			}
			g = append(g, target)
		}

		j, err := json.MarshalIndent(g, "", "  ")
		if err != nil {
			fmt.Println(fmt.Errorf("error marshalling json: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s\n", j)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status, err := tailscale.Status(context.Background())
		if err != nil {
			fmt.Println(fmt.Errorf("error getting status: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		j, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			fmt.Println(fmt.Errorf("error marshalling json: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s\n", j)
	})

	http.ListenAndServe(":8773", nil)
}

func firstIPString(v []netaddr.IP) string {
	if len(v) == 0 {
		return ""
	}
	return v[0].String()
}

func peerEnabled(p *ipnstate.PeerStatus) bool {
	if p == nil {
		return false
	}
	for _, h := range EnabledHosts {
		if p.HostName == h {
			return true
		}
	}

	return false
}

//
// [
//   {
//     "targets": [ "<host>", ... ],
//     "labels": {
//       "<labelname>": "<labelvalue>", ...
//     }
//   },
//   ...
// ]
type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
