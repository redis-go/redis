<p align="center">
<img
    src="https://redislabs.com/wp-content/uploads/2018/03/golang-redis.jpg"
    width="436" height="235" border="0">
<br>
</p>

<p align="center">Work-In-Progress Redis implementation in Go</p>

This project intentionally started to see how easy it is in Go to implement a full blown Redis clone.
As one of the side effects, imagine you could write redis modules in Go, that would be awesome!

# WIP
A *work-in-progress* implementation of redis for Go.
We are searching contributors!

### Roadmap
- [ ] Implementing data structures
  - [ ] Simple Key
  - [ ] List
  - [ ] Set
  - [ ] Sorted Set
  - [ ] Hash
  - [ ] ...
- [ ] Examples
- [ ] Alpha Release

### TODO after Roadmap is done
- [ ] Persistence
- [ ] Redis config
  - [ ] Default redis config format
  - [ ] YAML support
  - [ ] Json support
- [ ] Pub/Sub
- [ ] Redis modules
- [ ] Tests


### Documentation

godoc: https://godoc.org/github.com/redis-go/redis

### Getting Started

To install, run:
```bash
go get github.com/redis-go/redis
```