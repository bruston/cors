package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	input := flag.String("f", "", "path to list of urls, defaults to stdin if left empty")
	domain := flag.String("d", "", "domain name")
	concurrent := flag.Int("c", 10, "number of concurrent requests to make")
	timeout := flag.Int("t", 5, "timeout in seconds")
	cookies := flag.String("cookies", "", "cookies to send with the request")
	flag.Parse()

	var f io.ReadCloser
	if *input == "" {
		f = os.Stdin
	} else {
		file, err := os.Open(*input)
		if err != nil {
			log.Fatal(err)
		}
		f = file
	}
	defer f.Close()

	work := make(chan string)
	go func() {
		s := bufio.NewScanner(f)
		for s.Scan() {
			work <- s.Text()
		}
		if s.Err() != nil {
			log.Printf("error while scanning input: %v", s.Err())
		}
		close(work)
	}()

	wg := &sync.WaitGroup{}
	client := &http.Client{Timeout: time.Second * time.Duration(*timeout)}
	for i := 0; i < *concurrent; i++ {
		wg.Add(1)
		go check(client, *domain, *cookies, work, wg)
	}
	wg.Wait()
}

func check(client *http.Client, domain, cookies string, work chan string, wg *sync.WaitGroup) {
	for url := range work {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		if cookies != "" {
			req.Header.Set("Cookie", cookies)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
		origins := []string{"https://asdf.com", "https://asdf" + domain, "https://" + domain + "asdf.com", "null", "https://asdf." + domain + "asdf.com"}
		for _, v := range origins {
			req.Header.Set("Origin", v)
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			if resp.Header.Get("Access-Control-Allow-Origin") == v {
				fmt.Println(url, v)
				break
			}
		}
	}
	wg.Done()
}
