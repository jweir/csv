test:
	go test

covh: 
	go test . -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out 

cov: 
	go test . -covermode=count -coverprofile=coverage.out
	go tool cover -func=coverage.out 
