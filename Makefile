default:
	go build ./...

world-export:
	go run ./cmd/world-export -dat ../server/dat -all
