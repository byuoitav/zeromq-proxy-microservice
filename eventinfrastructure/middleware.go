package eventinfrastructure

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"
)

const ContextPublisher = "publisher"
const ContextSubscriber = "subscriber"
const ContextFilters = "filters"
const ContextPublisherAddress = "publisheraddress"

func BindPublisherAndSubscriber(p *Publisher, s *Subscriber) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(ContextPublisher, s)
			c.Set(ContextSubscriber, p)
			return next(c)
		}
	}
}

func BindPublisher(p *Publisher) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(ContextPublisher, p)
			return next(c)
		}
	}
}

func BindSubscriber(s *Subscriber) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(ContextSubscriber, s)
			return next(c)
		}
	}
}

func BindFiltersAndPublisherAddress(filters []string, publisherAddr string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(ContextFilters, filters)
			c.Set(ContextPublisherAddress, publisherAddr)
			return next(c)
		}
	}
}

func SendConnectionRequest(url string, req ConnectionRequest) {
	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("[error] %s", err.Error())
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	for err != nil || resp.StatusCode != 200 {
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Printf("[error] failed to post. Response: (%v) %s", resp.StatusCode, body)
		} else {
			log.Printf("[error] failed to post. Error: %s", err.Error())
		}

		log.Printf("Trying again in 5 seconds.")
		time.Sleep(5 * time.Second)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(body))
	}

	log.Printf("Successfully posted connection request to %s", url)
	resp.Body.Close()
}

func GetIP() string {
	var ip net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err.Error()
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && strings.Contains(address.String(), "/24") {
			ip, _, err = net.ParseCIDR(address.String())
			if err != nil {
				log.Fatalf("[error] %s", err.Error())
			}
		}
	}

	if ip == nil {
		log.Printf("[error] failed to find an non-loopback IP Address. Using PI_HOSTNAME/DEVELOPMENT_HOSTNAME as IP.")

		devhn := os.Getenv("DEVELOPMENT_HOSTNAME")
		if len(devhn) != 0 {
			log.Printf("Development machine. Using hostname %s", devhn)
			return devhn
		}

		pihn := os.Getenv("PI_HOSTNAME")
		if len(pihn) == 0 {
			log.Fatalf("[error] PI_HOSTNAME is not set.")
		}
		return pihn
	}

	log.Printf("My IP address is %s", ip)
	return string(ip)
}
