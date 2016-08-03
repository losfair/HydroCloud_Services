#!/usr/bin/env python
# coding: utf-8

import time
import json
import os
from wxbot import *

def do_push_data(data_json,custom_id):
	print "Pushing data"
	headers = {"Host":"wxmsg.internal.ixservices.net","Content-type": "text/json", "Accept": "text/html"}
	httpClient = httplib.HTTPConnection("127.0.0.1",8081)
	httpClient.request("POST","/accountInfoCollector.php?api_key=f5cebf0a41f47977d706a49ad46c63b35886e83a0182&custom_id="+custom_id,data_json,headers)
	resp=httpClient.getresponse()
	print "Result: ",resp.read()

class MyWXBot(WXBot):
	def schedule(self):
		data={}
		data['contact_list']=self.contact_list
		data['group_list']=self.group_list
		data['public_list']=self.public_list
		data['special_list']=self.special_list
		data['my_account']=self.my_account
		do_push_data(json.dumps(data),self.custom_id)
		os._exit(0)
#        print str(msg['msg_type_id'])+'/'+str(msg['content']['type'])+': '+str(msg['content']['data'])

def main():
	bot = MyWXBot()
	bot.DEBUG = True
	bot.conf['qr'] = 'png'
	if len(sys.argv)==2:
		print "Using custom data ID:",sys.argv[1]
		custom_id=sys.argv[1]
	else:
		custom_id=""
	bot.run(custom_id)

if __name__ == '__main__':
	main()
