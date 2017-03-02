package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func main() {
	// read config from a file
	viper.SetConfigName("config") // (JSON | TOML | YAML | HCL | Java properties)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	// set global defaults
	viper.SetDefault("port", 8080)
	viper.SetDefault("timeout", 10) // timeout in seconds

	// validate we have all the config required to start
	missing_config := false
	if !viper.IsSet("mirror_list_url") {
		log.Println("Error: Missing the required 'mirror_list_url' config which points to the mirror list file.")
		missing_config = true
	}
	if missing_config {
		log.Fatal("Missing required configuration details, please update the config file.")
	}

	// handle all routes
	http.HandleFunc("/", handleIndex)

	// start the web server
	log.Printf("mirror302 started - http://localhost:%d\n", viper.GetInt("port"))
	http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("port")), nil)
}

// index page (dashboard)
func handleIndex(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	mirrors := make([]*url.URL, 0)

	if path == "/favicon.ico" { // ignore this path
		return
	}

	// get mirror list
	resp, err := http.Get(viper.GetString("mirror_list_url"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		url, err := url.Parse(scanner.Text())
		if err != nil {
			log.Printf("Error parsing url: %s\n", err.Error())
			continue
		}
		mirrors = append(mirrors, url)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning mirror file: %s\n", err.Error())
	}

	// make sure we have at least one mirror to work with
	if len(mirrors) == 0 {
		http.Error(w, "No mirrors found to handle request.", http.StatusInternalServerError)
	}

	// handle the request simply be grabbing the first mirror with a 200 for a HEAD of the path
	url, err := firstMatch(path, mirrors)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else {
		http.Redirect(w, r, url.String(), 302)
	}

	return
}

func firstMatch(path string, mirrors []*url.URL) (*url.URL, error) {
	// process each url in the mirror list
	process_url := func(p string, u *url.URL, c chan *url.URL) {
		u.Path = fmt.Sprintf("%s/%s",
			strings.TrimRight(strings.TrimSpace(u.Path), "/"),
			strings.TrimLeft(strings.TrimSpace(p), "/"))

		res, err := http.Head(u.String())
		if err == nil && res.StatusCode == http.StatusOK {
			c <- u // valid, so send this url back on the result channel
		}
	}

	// process each mirror in its own goroutine (concurrently)
	resc := make(chan *url.URL) // channel to receive the first valid result
	for _, m := range mirrors {
		go process_url(path, m, resc)
	}

	// return the first result or timeout
	timeout := time.After(time.Duration(viper.GetInt("timeout")) * time.Second)
	select {
	case url := <-resc:
		return url, nil
	case <-timeout:
		return nil, errors.New("Timed out waiting for a mirror which can serve this path.")
	}
}
