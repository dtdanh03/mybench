package main

import (
  "fmt"
  "flag"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "time"
)

type responseInfo struct {
  status int
  bytes int64
  duration time.Duration
}

type summaryInfo struct {
  requested int64
  responded int64
}

func main() {
  fmt.Println("Hello from my app")

  requests := flag.Int64("n", 1, "Number of requests to perform")
  concurrency := flag.Int64("c", 1, "Number of multiple requests to make at a time")
  fmt.Println(requests, concurrency)

  flag.Parse()
  if flag.NArg() == 0 || *requests == 0 || *requests < *concurrency {
    flag.PrintDefaults()
    os.Exit(-1)
  }

  link := flag.Arg(0)
  channel := make(chan responseInfo)
  summary := summaryInfo {}
  for i := int64(0); i < *concurrency; i++ {
    summary.requested++
    go checkLink(link, channel)
  }

  for response := range channel {
    if summary.requested < *requests {
      summary.requested++
      go checkLink(link, channel)
    }
    summary.responded++
    // fmt.Println(response)
    fmt.Print("Http status code: ", response.status)
    fmt.Print(" | Bytes read: ", response.bytes)
    fmt.Println(" | Duration: ", response.duration)
    if summary.responded == summary.requested {
      break
    }
  }

}

func checkLink(link string, channel chan responseInfo) {
  startTime := time.Now()

  response, error := http.Get(link)
  if error != nil {
    panic(error)
  }
  bytesRead, _ := io.Copy(ioutil.Discard, response.Body)

  channel <- responseInfo {
    status: response.StatusCode,
    bytes: bytesRead,
    duration: time.Now().Sub(startTime),
  }
}
