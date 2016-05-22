package main

import (
  "os"
  "flag"
  "io"
  "io/ioutil"
  "bufio"
  "fmt"
  "regexp"
  "net/http"
  "github.com/andrew-d/go-termutil"
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

func get_url(url string) []byte {
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
          fmt.Printf("%s\n%s\n", url, get_url(url))
        }
      }
    } else {
      fmt.Printf("%s -> %s\n", v.provider, v.matches)
    }
  }
}
