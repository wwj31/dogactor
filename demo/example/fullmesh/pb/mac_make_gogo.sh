export PATH=$PATH:./

protoc -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.2/protobuf\
  -I=../pb \
  --gogofaster_out=../ *.proto

go generate ../msg/type.go