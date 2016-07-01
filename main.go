package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/andrew-d/go-termutil"
	"github.com/garyburd/redigo/redis"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type pattern struct {
	provider string
	re       *regexp.Regexp
	matches  []string
}

func appendUnique(slice []string, add string) []string {
	for _, el := range slice {
		if el == add {
			return slice
		}
	}
	return append(slice, add)
}

func httpGet(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return data
}

func main() {
	var testFlag bool
	flag.BoolVar(&testFlag, "test", false, "Test API keys to see if they are still active")
	flag.Parse()

	if termutil.Isatty(os.Stdin.Fd()) {
		fmt.Printf("Usage: git log -p | %s\n", os.Args[0])
		fmt.Println("Use -test to test found API keys.")
		return
	}

	scanners := []pattern{
		{
			"aws_access_key_id",
			regexp.MustCompile("AKIA[0-9A-Z]{16}"),
			make([]string, 0),
		},
		{
			"google_access_token",
			regexp.MustCompile("ya29.[0-9a-zA-Z_\\-]{68}"),
			make([]string, 0),
		},
		{
			"google_api",
			regexp.MustCompile("AIzaSy[0-9a-zA-Z_\\-]{33}"),
			make([]string, 0),
		},
		{ // xoxp are Slack API keys
			"slack_xoxp",
			regexp.MustCompile("xoxp-\\d+-\\d+-\\d+-[0-9a-f]+"),
			make([]string, 0),
		},
		{ // xoxb are Slack bot credentials
			"slack_xoxb",
			regexp.MustCompile("xoxb-\\d+-[0-9a-zA-Z]+"),
			make([]string, 0),
		},
		{
			"redis_url",
			regexp.MustCompile("redis://[0-9a-zA-Z:@.\\-]+"),
			make([]string, 0),
		},
		{
			"gemfury1",
			regexp.MustCompile("https?://[0-9a-zA-Z]+@[a-z]+\\.(gemfury.com|fury.io)(/[a-z]+)?"),
			make([]string, 0),
		},
		{
			"gemfury2",
			regexp.MustCompile("https?://[a-z]+\\.(gemfury.com|fury.io)/[0-9a-zA-Z]{20}"),
			make([]string, 0),
		},
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		for i := range scanners {
			matches := scanners[i].re.FindAllString(line, -1)
			for _, match := range matches {
				scanners[i].matches = appendUnique(scanners[i].matches, match)
			}
		}
	}

	for _, v := range scanners {
		if len(v.matches) == 0 {
			continue
		}
		if v.provider == "google_access_token" {
			fmt.Println("Found Google access tokens:")
			for _, m := range v.matches {
				fmt.Printf("- %s\n", m)
				if testFlag {
					url := "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=" + m
					fmt.Printf("%s\n%s\n", url, httpGet(url))
				}
			}
			fmt.Println()
		} else if v.provider == "google_api" {
			fmt.Println("Found Google API keys:")
			for _, m := range v.matches {
				fmt.Printf("- %s\n", m)
				if testFlag {
					// Just testing to get info about gangnam style, is there a better endpoint?
					url := "https://www.googleapis.com/youtube/v3/videos?part=id&id=9bZkp7q19f0&key=" + m
					fmt.Printf("%s\n%s\n", url, httpGet(url))
				}
			}
			fmt.Println()
		} else if strings.HasPrefix(v.provider, "slack_") {
			fmt.Println("Found Slack keys:")
			for _, m := range v.matches {
				fmt.Printf("- %s\n", m)
				if testFlag {
					url := "https://slack.com/api/auth.test?token=" + m
					data := httpGet(url)
					fmt.Printf("%s\n%s\n", url, data)
					if data[len(data)-1] != '\n' {
						fmt.Println()
					}
				}
			}
			fmt.Println()
		} else if v.provider == "redis_url" {
			fmt.Println("Found Redis URLs:")
			for _, m := range v.matches {
				fmt.Printf("- %s\n", m)
				if testFlag {
					c, err := redis.DialURL(m, redis.DialConnectTimeout(time.Second))
					if err == nil {
						fmt.Print("Connection successful!\n\n")
						defer c.Close()
					} else {
						fmt.Printf("Connection failed: %s\n\n", err)
					}
				}
			}
			fmt.Println()
		} else {
			fmt.Printf("%s -> %s\n\n", v.provider, v.matches)
		}
	}
}
