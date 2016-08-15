CFLAGS := -fPIC

all:
	make detect
	make test

detect: detect.o
	$(CC) -fPIC -shared -o libLetsChat.so detect.o
	ar rcs libLetsChat.a detect.o

test: test.o
	$(CC) -o test test.o -L. -lLetsChat

clean:
	find . -name "*.o" -exec rm '{}' ';'
	find . -name "*.so" -exec rm '{}' ';'
	find . -name "*.a" -exec rm '{}' ';'
	rm test

