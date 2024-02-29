## This is an example of benchmarking program in Go

### Prerequisites

- `go install go.k6.io/k6@latest` - benchmarking tool

### How to start

- `GET /bench` - request for random webp image
- `GET /` - 'hits per image' statistics
- `k6 run bench.js`, `k6 run quickbench.js` - performance tests

### Kudos

- Based on the example is from [Golang course, Zero To Mastery](https://academy.zerotomastery.io/courses/enrolled/1600953)
