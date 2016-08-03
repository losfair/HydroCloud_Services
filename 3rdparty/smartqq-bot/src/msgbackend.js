(function() {
var http = require("http");
function post_json(tgthost,tgtport,tgtpath,reqArr) {
	var reqData=JSON.stringify(reqArr);
	console.log("PUSH DATA: "+reqData);
	var post_options = {
		host: tgthost,
		port: tgtport,
		path: tgtpath,
		method: 'POST',
		headers: {
			'Content-Type': 'text/json',
			'Content-Length': reqData.length
		}
	};
	 
	var post_req = http.request(post_options, function (response) {
		var responseText=[];
		var size = 0;
		response.on('data', function (data) {
			responseText.push(data);
			size+=data.length;
		});
		response.on('end', function () {
			responseText = Buffer.concat(responseText,size);
			console.log("PUSH RESPONSE: "+responseText);
//			callback(responseText);
		});
	});

	post_req.write(reqData);
	post_req.end();
}

function push_message(msg) {
	console.log("PUSH REQUEST");
	post_json("qqmsg.internal.ixservices.net","8081","/onMessage.php",msg);
}
})();
