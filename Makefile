.SILENT: clean 9ccgo
.PHONY: test clean
SRCS=$(wildcard *.go)

9ccgo: clean
	go build -gcflags '-N -l' -o 9ccgo $(SRCS)
	
test: 9ccgo test/test.c
	go test -v $(SRCS)
	./9ccgo -test

	@gcc -E -P test/test.c | ./9ccgo - > tmp-test1.s
	@./9ccgo test/token.c > tmp-test2.s
	@gcc -c -o tmp-test2.o test/gcc.c
	@gcc -static -o tmp-test tmp-test1.s tmp-test2.s tmp-test2.o
	./tmp-test

clean:
	rm -f 9ccgo *.o *~ tmp* a.out test/*~

