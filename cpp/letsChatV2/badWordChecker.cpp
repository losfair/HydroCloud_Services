#include <string.h>

#include "badWords.h"

namespace LetsChat {

static int getKeywordCount (const char *str, const char *p) {
	int c = 0;
	size_t len = strlen(p);
	while( (str = strstr(str,p)) != NULL ) {
		str += len;
		c++;
	}
	return c;
}

float getBadWordWeight (const char *msg) {
	int i;
	char *stripped;
	unsigned pos = 0;
	size_t msglen;

	float totalWeight = 0.0;

	msglen = strlen(msg);

	stripped = new char [msglen + 1];

	for(i = 0; i < msglen; i++) {
		if( msg[i] < 0 
		|| (msg[i] >= '0' && msg[i] <= '9')
		|| (msg[i] >= 'A' && msg[i] <= 'Z')
		|| (msg[i] >= 'a' && msg[i] <= 'z') ) {
			stripped[pos] = msg[i];
			pos++;
		}
	}
	stripped[pos] = '\0';

	for(i = 0; i < badWordSize; i++) totalWeight += (wordWeights[i] * getKeywordCount(stripped, badWords[i]));

	delete[] stripped;
	return totalWeight;
}

}

