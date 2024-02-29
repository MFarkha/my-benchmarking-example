## This is an example of benchmarking program in Go

### Prerequisites

- `go install go.k6.io/k6@latest` - benchmarking tool

### How to start

- `GET /bench` - request for random webp image
- `GET /` - 'hits per image' statistics
- `k6 run bench.js`, `k6 run quickbench.js` - performance tests
- `go install github.com/codesenberg/bombardier@latest` - install another test tool 'bombardier'
- `bombardier -c 500 -d 5s http://localhost:8099/bench` - 500 users within 5s
- `bombardier -c 5000 -d 10s http://localhost:8099/bench` - 5000 users within 10s

### Kudos

- Based on the example is from [Golang course, Zero To Mastery](https://academy.zerotomastery.io/courses/enrolled/1600953)
