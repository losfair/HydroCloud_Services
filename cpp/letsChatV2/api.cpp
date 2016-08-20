#include <map>
#include <vector>
#include <string>

#include "letschat.h"

extern "C" void chatMsgInput(LetsChat::id_type ChatId, LetsChat::id_type SenderId, const char *Content) {
	auto targetChat = LetsChat::Chat::GetById(ChatId);
	targetChat -> NewMessage(SenderId, Content);
}

extern "C" float chatGetBadWordWeight(LetsChat::id_type ChatId) {
	auto targetChat = LetsChat::Chat::GetById(ChatId);
	return targetChat -> GetBadWordWeight();
}
