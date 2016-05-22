package main

import (
  "os"
  "flag"
  "io"
  "io/ioutil"
  "bufio"
  "fmt"
  "strings"
  "regexp"
  "time"
  "net/http"
  "github.com/andrew-d/go-termutil"
  "github.com/garyburd/redigo/redis"
)

type Pattern struct {
  provider string
  re *regexp.Regexp
  matches []string
}

func append_unique(slice []string, add string) []string {
  for _, el := range slice {
    if el == add {
      return slice
    }
  }
  return append(slice, add)
}

func http_get(url string) []byte {
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
  test_flag := flag.Bool("test", false, "Test API keys to see if they are still active")
  flag.Parse()

  if termutil.Isatty(os.Stdin.Fd()) {
    fmt.Println("Usage: git log -p | ./secrets-scanner")
    fmt.Println("Use -test to test found API keys.")
    return
  }

  scanners := []Pattern {
    {
      "google_access_token",
      regexp.MustCompile("ya29.[0-9a-zA-Z_\\-]{68}"),
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
        scanners[i].matches = append_unique(scanners[i].matches, match)
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
        if *test_flag {
          url := "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token="+m
          fmt.Printf("%s\n%s\n", url, http_get(url))
        }
      }
    } else if strings.HasPrefix(v.provider, "slack_") {
      fmt.Println("Found Slack keys:")
      for _, m := range v.matches {
        fmt.Printf("- %s\n", m)
        if *test_flag {
          url := "https://slack.com/api/auth.test?token="+m
          data := http_get(url)
          fmt.Printf("%s\n%s\n", url, data)
          if data[len(data)-1] != '\n' {
            fmt.Println()
          }
        }
      }
    } else if v.provider == "redis_url" {
      fmt.Println("Found Redis URLs:")
      for _, m := range v.matches {
        fmt.Printf("- '%s'\n", m)
        if *test_flag {
          c, err := redis.DialURL(m, redis.DialConnectTimeout(time.Second))
          if err == nil {
            fmt.Println("Connection successful!\n")
            defer c.Close()
          } else {
            fmt.Printf("Connection failed: %s\n\n", err)
          }
        }
      }
    } else {
      fmt.Printf("%s -> %s\n", v.provider, v.matches)
    }
  }
}
