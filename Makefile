.SILENT: clean test 9ccgo
SRCS=$(wildcard *.go)

9ccgo: clean
	go build -o 9ccgo $(SRCS)
	
test: 9ccgo
	./test.sh

clean:
	rm -f 9ccgo *.o *~ tmp*

