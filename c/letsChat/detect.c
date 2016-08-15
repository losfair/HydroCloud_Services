#include <stdlib.h>
#include <string.h>
#include <time.h>

#include "badWords.h"

typedef unsigned long long id_type;

struct GroupWords {
	id_type id;
	float totalWeight;
	time_t lastUpdateTime;
	unsigned long lastMessageHash;
	unsigned int msgRepeatCount;
	struct GroupWords *next;
};

static const int GROUPWORDS_TIMEOUT_SECS = 300;
static const float WEIGHT_TO_OUTPUT = 1.0;
static const unsigned int REPEAT_COUNT_LEVELS[] = {5,10};

static struct GroupWords *listStart = NULL;

static int getCount (const char *str, const char *p) {
	int c = 0;
	size_t len = strlen(p);
	while( (str = strstr(str,p)) != NULL ) {
		str += len;
		c++;
	}
	return c;
}

// 32 or 64 bits for long. It just works.
static unsigned long badHash (const unsigned char *str) {
	int c;
	unsigned long hash = 5381;

	while (c = *str++) hash = ((hash << 5) + hash) + c;

	return hash;
}

static inline void initGroupWordsNode (struct GroupWords *n, id_type id) {
	n -> id = id;
	n -> totalWeight = 0.0;
	n -> lastUpdateTime = time(0);
	n -> lastMessageHash = 0;
	n -> msgRepeatCount = 0;
	n -> next = NULL;
}

static struct GroupWords *findGroupWordsById (id_type id) {
	struct GroupWords *current = listStart;

	for(;; current = current -> next) {
		if(current -> id == id) return current;
		if(current -> next == NULL) break;
	}

	current = (current -> next = malloc(sizeof(struct GroupWords)));
	initGroupWordsNode(current, id);

	return current;
}

static inline void updateTotalWeight (struct GroupWords *n, const char *msg) {
	int i;
	char *stripped;
	unsigned pos = 0;
	size_t msglen;

	if(time(0) - n->lastUpdateTime > GROUPWORDS_TIMEOUT_SECS) {
		n -> totalWeight = 0.0;
	}

	msglen = strlen(msg);

	stripped=malloc(msglen);

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

	for(i = 0; i < badWordSize; i++) n -> totalWeight += (wordWeights[i] * getCount(stripped, badWords[i]));

	free(stripped);
	n -> lastUpdateTime = time(0);
}

static void updateLastMessage (struct GroupWords *n, const char *msg) {
	size_t len = strlen(msg);
	unsigned long msgHash = badHash(msg);

	if(n -> lastMessageHash != 0 && n -> lastMessageHash == msgHash) {
		n -> msgRepeatCount++;
	} else {
		n -> msgRepeatCount = 0;
		n -> lastMessageHash = msgHash;
	}
}

static void __attribute__((constructor)) chatInit (void) {
	listStart = malloc(sizeof(struct GroupWords));
	initGroupWordsNode(listStart, 0xffffffff);
}

void chatMsgInput (const char *msg, id_type id) {
	int i;

	struct GroupWords *n = findGroupWordsById(id);

	updateLastMessage(n,msg);
	updateTotalWeight(n,msg);
}

float chatGetTotalWeight (id_type id) {
	struct GroupWords *n;

	n=findGroupWordsById(id);

	if(time(0) - n->lastUpdateTime > GROUPWORDS_TIMEOUT_SECS) {
		return 0.0;
	} else {
		return findGroupWordsById(id) -> totalWeight;
	}
}

void chatClearTotalWeight (id_type id) {
	findGroupWordsById(id) -> totalWeight = 0.0;
}

const char * chatGetOutputText (id_type id) {
	struct GroupWords *n = findGroupWordsById(id);

	if(n -> totalWeight > WEIGHT_TO_OUTPUT) {
		n -> totalWeight = 0.0;
		return "这群风气不太对啊。";
	} else if(n -> msgRepeatCount == REPEAT_COUNT_LEVELS[0]) {
		return "破队形。";
	} else if(n -> msgRepeatCount == REPEAT_COUNT_LEVELS[1]) {
		return "刷屏。不太好吧。";
	} else {
		return "OK";
	}
}

