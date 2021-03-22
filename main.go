package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr   = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	ecosol = flag.String("ecosol.address", "http://admin:admin@192.168.2.9/econet/regParams", "The address to econet's metrics endpoint including credentials")
	loop = flag.Duration("ecosol.probe-period", time.Minute, "how often to probe the endpoint")
)

var (
	curr = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ecosol_curr",
		Help: "various metrics i haven't deciphered yet",
	}, []string{"metric"})
)

func init() {
	prometheus.MustRegister(curr)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

type RegParams struct {
	Curr map[string]*float64 `json:"curr"`
}

func main() {
	flag.Parse()
	go func() {
		client := &http.Client{}
		for {
			resp, err := client.Get(*ecosol)
			if err == nil {
				var params RegParams
				defer resp.Body.Close()
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&params)
				if err == nil {
					for k, v := range params.Curr {
						if v != nil {
							curr.WithLabelValues(k).Set(*v)
						}
					}
				} else {
					log.Print(err)
				}
			} else {
				log.Print(err)
			}
			time.Sleep(*loop)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
