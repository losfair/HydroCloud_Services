CXXFLAGS := -fPIC -O2 -std=c++11

all:
	make letschat
	make test

letschat: letschat.o badWordChecker.o api.o
	$(CXX) -fPIC -shared -o libLetsChat.so letschat.o badWordChecker.o api.o
	ar rcs libLetsChat.a letschat.o badWordChecker.o api.o

test: test.o
	$(CC) -o test test.o -L. -lLetsChat

clean:
	find . -name "*.o" -exec rm '{}' ';'
	find . -name "*.so" -exec rm '{}' ';'
	find . -name "*.a" -exec rm '{}' ';'
	rm test

