#include <stdio.h>
#include <stdlib.h>
#include "api.h"

int main(int argc, char *argv[]) {
	const id_type uid = 233;
	int i;

	for (i = 1; i < argc; i++) {
		chatMsgInput(100,uid,argv[i]);
	}

	printf("%f\n",chatGetBadWordWeight(100));
}
