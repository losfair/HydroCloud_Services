#ifndef _BXC_DETECT_H_
#define _BXC_DETECT_H_

typedef unsigned long long id_type;

void chatMsgInput (const char *, id_type);
float chatGetTotalWeight (id_type);
void chatClearTotalWeight (id_type);
const char * chatGetOutputText (id_type);

#endif
