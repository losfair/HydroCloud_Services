#ifndef _LETSCHAT_H_
#define _LETSCHAT_H_

#include <map>
#include <vector>
#include <string>
#include <list>

namespace LetsChat {

typedef unsigned long long id_type;

class Message {
	public:
	id_type Sender;
	std::string Content;
	time_t TimestampOnCreate;

	Message();
};

class Chat {
	public:
	id_type Id;
	std::list<Message*> Messages;
	time_t TimestampOnCreate;
	time_t TimestampOnUpdate;

	Chat();

	static Chat* Create(id_type);
	static Chat* GetById(id_type);

	void Destroy();

	void NewMessage(id_type&,std::string);
	void ForEachMessage(void(*)(Message*));
	std::list<Message*>::iterator GetRecentMessages(unsigned);

	float GetBadWordWeight();
};

}
/*
void chatMsgInput (const char *, id_type);
float chatGetTotalWeight (id_type);
void chatClearTotalWeight (id_type);
const char * chatGetOutputText (id_type);
*/

#endif
