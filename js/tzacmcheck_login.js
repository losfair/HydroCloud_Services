var page = require('webpage').create();
var system = require('system');

if(system.args.length!=3) {
	console.log(JSON.stringify({"err":1,"msg":"Bad arguments"}));
	phantom.exit(0);
}

var loginInfo={
	"userName": system.args[1],
	"password": system.args[2]
};

page.open('http://noi.ntzx.cn:8080/acmhome/welcome.do?method=index',function() {
	var ret=page.evaluate(function(loginInfo) {
		var xhr=new XMLHttpRequest;
		xhr.open("POST",'/acmhome/login.do',false);
		xhr.setRequestHeader('Content-Type','application/x-www-form-urlencoded');
		xhr.send("userName="+encodeURIComponent(loginInfo['userName'])+'&password='+encodeURIComponent(loginInfo['password']));
		if(xhr.responseText.indexOf("欢迎你的到来")==-1) return {"err":1,"msg":"Login failed"};
		xhr=new XMLHttpRequest;
		xhr.open("GET","http://noi.ntzx.cn:8080/acmhome/userDetail.do?userName="+encodeURIComponent(loginInfo['userName']),false);
		xhr.send(null);
		document.write(xhr.responseText);

		var ret=document.getElementById("content").getElementsByTagName("table")[0].getElementsByTagName("tr")[0].getElementsByTagName("strong")[0].parentNode.parentNode.getElementsByTagName("a");

		var arr=[]
		for(var i=0;i<ret.length;i++) arr.push(ret[i].innerHTML);

		return {"err":0,"msg":arr};

	},loginInfo);
	console.log(JSON.stringify(ret));
	phantom.exit(0);
});
