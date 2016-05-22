package main

import "os"
import "io"
import "bufio"
import "fmt"
import "regexp"

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

func main() {
  scanners := []Pattern {
    {
      "google_access_token",
      regexp.MustCompile("ya29.[0-9a-zA-Z]{68}"),
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
    if v.provider == "google_access_token" {
      fmt.Println("Found Google access tokens:")
      for _, m := range v.matches {
        url := "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token="+m
        fmt.Printf("- %s => %s\n", m, url)
      }
    } else {
      fmt.Printf("%s -> %s\n", v.provider, v.matches)
    }
  }

  fmt.Println("Goodbye!")
}
