.SILENT: clean test 9ccgo
SRCS=$(wildcard *.go)

9ccgo: clean
	go build -gcflags '-N -l' -o 9ccgo $(SRCS)
	
test: 9ccgo
	go test -v $(SRCS)
	./9ccgo -test
	./test.sh

clean:
	rm -f 9ccgo *.o *~ tmp* a.out

