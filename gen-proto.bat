

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc -I=./protocol -I./third_party/googleapis --go_out=. --go-grpc_out=.  ./protocol/api/*.proto
protoc -I=./protocol -I./third_party/googleapis --go_out=. --go-grpc_out=. ./protocol/core/*.proto
protoc -I=./protocol -I./third_party/googleapis --go_out=. --go-grpc_out=.  ./protocol/core/contract/*.proto