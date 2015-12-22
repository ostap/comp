comp: y.go
	go build -o comp .

test: y.go
	go test .

y.go:
	go tool yacc -o y.go -p "comp_" grammar.y

clean:
	rm -f comp y.go y.output
