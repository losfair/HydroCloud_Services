#include <stdio.h>
#include <stdlib.h>
#include "detect.h"

int main(int argc, char *argv[]) {
	const id_type uid = 233;
	int i;

	for (i = 1; i < argc; i++) {
		chatMsgInput(argv[i],uid);
	}

	printf("%f\n",chatGetTotalWeight(uid));
}
