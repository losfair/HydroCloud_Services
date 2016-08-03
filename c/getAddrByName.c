#include <stdio.h>
#include <stdlib.h>
#include <netdb.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <string.h>

const char* GetAddrByName(char *name) {
	struct addrinfo *answer, hint, *curr;
	char *ipstr=malloc(256);
	bzero(&hint, sizeof(hint));
	hint.ai_family = AF_INET;
	hint.ai_socktype = SOCK_STREAM;

	int ret = getaddrinfo(name, NULL, &hint, &answer);
	if (ret != 0) {
		printf("GetAddrByName: Error #1, %s\n",strerror(ret));
		return NULL;
	}

	inet_ntop(AF_INET,&(((struct sockaddr_in *)(answer->ai_addr))->sin_addr),ipstr, 256);

	freeaddrinfo(answer);

	return ipstr;
}
