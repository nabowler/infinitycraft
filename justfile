run: build
  ./infinitecraft

run_cgo: build_cgo
  ./infinitecraft

build:
  CGO_ENABLED=0 go build .

build_cgo:
  CGO_ENABLED=1 go build .