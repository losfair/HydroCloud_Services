#include <time.h>

#include <map>
#include <string>
#include <list>

#include "letschat.h"
#include "badWordChecker.h"

namespace LetsChat {

std::map<id_type,Chat*> chats;

Message::Message() {
	Sender = 0;
	TimestampOnCreate = time(0);
}

Chat::Chat() {
	Id = 0;
	TimestampOnUpdate = TimestampOnCreate = time(0);
}

Chat* Chat::Create(id_type Id) {
	auto newChat = new Chat;
	newChat -> Id = Id;
	chats[Id] = newChat;
	return newChat;
}

Chat* Chat::GetById(id_type Id) {
	auto itr = chats.find(Id);
	Chat *targetChat;

	if(itr == chats.end()) targetChat = Create(Id);
	else targetChat = itr->second;

	return targetChat;
}

void Chat::Destroy() {
	chats.erase(chats.find(Id));
	delete this;
}

void Chat::NewMessage(id_type& Sender, std::string Content) {
	auto msg = new Message;

	msg -> Sender = Sender;
	msg -> Content = Content;

	Messages.push_back(msg);

	TimestampOnUpdate = time(0);
}

void Chat::ForEachMessage(void (*handler)(Message*)) {
	for(auto item:Messages) handler(item);
}

std::list<Message*>::iterator Chat::GetRecentMessages(unsigned n) {
	auto ptr = Messages.end();
	auto start = Messages.begin();
	for(int i=0;i<n;i++) {
		if(ptr == start) return ptr;
		ptr--;
	}
	return ptr;
}

float Chat::GetBadWordWeight() {
	float ret = 0.0;

	for(auto item = GetRecentMessages(10); item != Messages.end(); item++) {
		ret += getBadWordWeight(((*item)->Content).c_str());
	}

	return ret;
}

}
