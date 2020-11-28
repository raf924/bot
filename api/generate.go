package api

//go:generate protoc --proto_path=definitions -I=$GOPATH/src --go_out=messages --go_opt=paths=source_relative definitions/user.proto
//go:generate protoc --proto_path=definitions -I=$GOPATH/src --go_out=messages --go_opt=paths=source_relative definitions/packets.proto
//go:generate protoc --proto_path=definitions -I=$GOPATH/src --go_out=messages --go_opt=paths=source_relative definitions/command.proto
