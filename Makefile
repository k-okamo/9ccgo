9ccgo:
	go build -o 9ccgo 9cc.go

test: 9ccgo
	./test.sh

clean:
	rm -f 9ccgo *.o *~ tmp*

