.SILENT: clean 9ccgo
.PHONY: test clean
SRCS=$(wildcard *.go)

9ccgo: clean
	go build -gcflags '-N -l' -o 9ccgo $(SRCS)
	
test: 9ccgo test/test.c
	go test -v $(SRCS)
	./9ccgo -test

	@gcc -E -P test/test.c > tmp-test.tmp
	@./9ccgo tmp-test.tmp > tmp-test.s
	@echo 'int global_arr[1] = {5};' | gcc -xc -c -o tmp-test2.o -
	@gcc -static -o tmp-test tmp-test.s tmp-test2.o
	./tmp-test

clean:
	rm -f 9ccgo *.o *~ tmp* a.out test/*~

