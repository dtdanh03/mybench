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

  var defaultRequests int64 = 1
  var defaultConcurrency int64 = 1
  var defaultTimeout int64 = 5
  var defaultTimeLimit int64 = 15

  requests := flag.Int64("n", defaultRequests, "Number of requests to perform")
  concurrency := flag.Int64("c", defaultConcurrency, "Number of multiple requests to make at a time")
  timeout := flag.Int64("timeout", defaultTimeout, "Maximum seconds to perform each request")
  timeLimit := flag.Int64("timeLimit", defaultTimeLimit, "Maximum seconds to spend for benchmarking")

  fmt.Println(requests, concurrency)

  flag.Parse()
  if flag.NArg() == 0 || *requests == 0 || *requests < *concurrency {
    flag.PrintDefaults()
    os.Exit(-1)
  }

  timeLimitChannel := time.After(time.Duration(*timeLimit) * time.Second)
  doneChannel := make(chan bool)

  go func() {
    startBenchmarking(requests, concurrency, *timeout)
    doneChannel <- true
  }()

  select {
    case <- timeLimitChannel:
      fmt.Println("Benchmarking process has exceeded the time limit")
    case <- doneChannel:
  }
}

func startBenchmarking(requests *int64, concurrency *int64, timeout int64) {
  link := flag.Arg(0)
  channel := make(chan responseInfo)
  summary := summaryInfo {}
  for i := int64(0); i < *concurrency; i++ {
    summary.requested++
    go checkLink(link, channel, timeout)
  }

  for response := range channel {
    if summary.requested < *requests {
      summary.requested++
      go checkLink(link, channel, timeout)
    }
    summary.responded++
    fmt.Print("Http status code: ", response.status)
    fmt.Print(" | Bytes read: ", response.bytes)
    fmt.Println(" | Duration: ", response.duration)
    if summary.responded == summary.requested {
      break
    }
  }
}

func checkLink(link string, channel chan responseInfo, timeout int64) {
  startTime := time.Now()
  client := makeHttpClientWithTimeout(timeout)
  response, error := client.Get(link)

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

func makeHttpClientWithTimeout(timeout int64) http.Client {
  return http.Client {
    Timeout: time.Duration(time.Duration(timeout) * time.Second),
  }
}
