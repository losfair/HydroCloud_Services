#include <stdio.h>

#include <map>
#include <string>

namespace LetsChat {

std::map<std::string,unsigned> learnedWords;

static const unsigned chCount = 2;
static const unsigned targetWordSize = 4 * chCount + 1;
static unsigned char targetWord[targetWordSize];
static unsigned targetWordPos = 0;
static unsigned currentChCount = 0;

static void chInputEndString() {
	targetWord[targetWordPos] = '\0';

	targetWordPos = 0;
	currentChCount = 0;

	auto targetString = (std::string)targetWord;
	auto v = learnedWords.find(targetString)

	if(v==learnedWords.end()) learnedWords[targetString] = 1;
	else learnedWords[targetString] = v->second + 1;
}

static void chInputEndCh() {
	currentChCount++;
	if(currentChCount == chCount) chInputEndString();
}

static void chInput(char ch) {
	if(targetWordPos == targetWordSize - 1) return;
	targetWord[targetWordPos] = ch;
	targetWordPos++;
}


void learnSentence(const char *_content) {
	int i;
	const unsigned char *content = (const unsigned char *)_content;
	unsigned unicodeBits;

	while(*content) {
		if((*content)>>7) {
			for(unicodeBits=0;unicodeBits<4;unicodeBits++) {
				if(!( (*content) & (1<<(7-unicodeBits)) )) break;
			}
			for(i=0;i<unicodeBits;i++) {
				chInput(*content);
				content++;
				if(!(*content)) break;
			}
		} else {
			chInput(*content);
			content++;
		}
		chInputEndCh();
	}
}

}
