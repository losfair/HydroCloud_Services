const red = "#d62d20";
const blue = "#0057e7";
const green = "#008744";
const yellow = "#ffa700";

var savedLastId = 0;
var pageTitle;

function showLoading(callback) {
	if(!callback) callback=null;
	$("#container").fadeOut(750);
	$("#loading-container").fadeIn(750,callback);
}
function hideLoading(callback) {
	if(!callback) callback=null;
	$("#loading-container").fadeOut(750);
	$("#container").fadeIn(750,callback);
}

function clientLog(data) {
	$.get("/timeline/ajax/clientLog/"+encodeURIComponent(data), function(result){});
}

function getUpdates(lastId,callback) {
	var xhr = new XMLHttpRequest;
	xhr.open("GET","/timeline/ajax/getUpdates/"+lastId,true);
	xhr.onreadystatechange=function() {
		if(xhr.readyState == 4 && xhr.status == 200) {
			callback(xhr.responseText);
		}
	};
	xhr.send(null);
}

function getImageToken(url,callback) {
	$.post("/timeline/ajax/requestImageToken/set",url,callback);
}

function plusoneGet(id,callback) {
	$.post("/timeline/ajax/plusoneGet",""+id,callback);
}

function plusoneUpdate(id,callback) {
	$.post("/timeline/ajax/plusoneUpdate",""+id,callback);
}

var ref = null;

function createContentBlock(props) {
	var newBlock = document.createElement("div");
	newBlock.className = "content-block";

	var contentArea = document.createElement("div");
	contentArea.className = "content-area";

	$(contentArea).css("background-image","url("+props["imgURL"]+")");
	$(contentArea).css("background-size","cover");

	$(contentArea).click(function() {
		$("#image-container").css("background-image","url("+props["imgURL"]+")");
		$("#image-container").css("background-size","contain");
		$("#image-container").css("background-position","center");
		$("#image-container").css("background-repeat","no-repeat");
		$("#image-container").click(function() {
			$("#image-container").fadeOut();
			$("#cover").fadeOut();
			$("#confirm-button").fadeOut();
		});
		$("#confirm-button").click(function() {
			var timeoutId = setTimeout(function() {
				showLoading();
				$("#cover").fadeOut();
				$("#image-container").fadeOut();
				$("#topbar").css("background-color",red);
				$("#topbar-title").text("加载中...");
				$("#reload-button").fadeOut();
				setTimeout(function() {
					getImageToken(props["imgURL"],function(imgToken) {
						window.location = "imgloader.html?img="+imgToken;
					});
				},2000);
			},800);
		});
		$("#confirm-button").fadeIn();
		$("#cover").fadeIn();
		$("#image-container").fadeIn();
	});

	contentArea.itemProperties = props;

	var commentArea = document.createElement("div");
	commentArea.className = "comment-area";

	var plusOne = document.createElement("div");
	plusOne.className = "comment-plusone";
	plusOne.innerHTML = "+1";
	plusOne.itemId = props['id'];
	plusoneGet(plusOne.itemId,function(newValue) {
		plusOne.innerHTML = "+" + newValue;
	});

	var plusOneDone = false;
	$(plusOne).click(function() {
		if(plusOneDone) return;
		plusOneDone = true;
		plusoneUpdate(props['id'],function(newValue) {
			plusOne.innerHTML = "Get!";
			setTimeout(function() {
				plusOne.innerHTML = newValue;
			},1000);
		});
	});

	newBlock.appendChild(contentArea);
	newBlock.appendChild(commentArea);
	commentArea.appendChild(plusOne);

//	$(newBlock).hide();

	document.getElementById("container").insertBefore(newBlock,ref);

//	$(newBlock).show(200);

	ref = newBlock;
}

var isInPost = false;

var postBox;
var urlBox;

function sendPost() {
	$.post("/timeline/ajax/createUpdate",urlBox.value,hidePost);
}

function hidePost() {
	isInPost = false;
	$("#post-button").fadeOut();
	document.getElementById("container").removeChild(postBox);
}

function showPost() {
	isInPost = true;
	$("#post-button").fadeIn();
	postBox = document.createElement("div");
	$(postBox).css("margin","15px 15px");
	$(postBox).css("line-height","150px");
	
	urlBox = document.createElement("input");
	$(urlBox).attr("type","text");
	$(urlBox).css("width","100%");

	postBox.appendChild(urlBox);

	document.getElementById("container").insertBefore(postBox,ref);
}

function timedUpdatePlusone() {
	$(".comment-plusone").each(function() {
		var targetPlusone = this;
		var targetItemId = this.itemId;
		document.getElementById("debug-container").innerHTML += "targetPlusone: "+targetPlusone.itemId+"\n";
		plusoneGet(targetItemId,function(newValue) {
			var appendedFront="";
			if(targetPlusone.innerHTML[0] == "+") appendedFront = "+";
			targetPlusone.innerHTML = appendedFront + newValue;
		});
	});
}

var deltaScroll = 1;

function doLoadContent() {
	setTimeout(function() {
	if(document.body.scrollTop - deltaScroll < 0) document.body.scrollTop = 0;
	else document.body.scrollTop -= deltaScroll;
	deltaScroll *= 1.2;

	if(document.body.scrollTop > 0) doLoadContent();
	else showLoading(function() {
	document.getElementById("debug-container").innerHTML += "Starting getUpdates\n";
	getUpdates(savedLastId,function(text) {
		var data = eval('('+text+')');

		for(var i=0;i<data.length;i++) {
			savedLastId = savedLastId > data[i]["id"] ? savedLastId : data[i]["id"];
			createContentBlock(data[i]);
		}

		$("#topbar").css("background-color",green);
		$("#topbar-title").text(data.length+" 条新动态");
		setTimeout(function() {
			$("#topbar").css("background-color",blue);
			$("#topbar-title").text(pageTitle);
		},1000);

		$("#reload-button").fadeIn();
		hideLoading();
	});
	});
	},10);
}

function loadContent() {
	$("#topbar").css("background-color",red);
	$("#topbar-title").text("加载中...");
	$("#reload-button").fadeOut();
	doLoadContent();
}

function loadContentInitial() {
	document.getElementById("container").innerHTML = "<div id=\"content-before-here\"></div><br><center><code>&copy; 2016 hydrocloud.net.</code></center>";

	ref = document.getElementById("content-before-here");

	document.getElementById("debug-container").innerHTML += "Starting loadContent\n";

	loadContent();

	setInterval(timedUpdatePlusone,10000);
}

var previousStatusContainer = "";
var previousStatusLoadingContainer = "";

function toggleDebug() {
	if(document.getElementById("debug-container").style.display=="none") {
		$("#debug-container").fadeIn();
		previousStatusContainer = document.getElementById("container").style.display;
		previousStatusLoadingContainer = document.getElementById("loading-container").style.display;
		document.getElementById("container").style.display = "none";
		document.getElementById("loading-container").style.display = "none";
		$("#reload-button").fadeOut();
		$("#topbar").css("background-color",red);
		$("#topbar-title").text("Debug Info");
	} else {
		$("#debug-container").fadeOut();
		document.getElementById("container").style.display = previousStatusContainer;
		document.getElementById("loading-container").style.display = previousStatusLoadingContainer;
		$("#reload-button").fadeIn();
		$("#topbar").css("background-color",blue);
		$("#topbar-title").text(pageTitle);
	}
}

function saveValues() {
	pageTitle = $("#topbar-title").text();
}

var topbarStatus = "show";

function timedCheckScrolls() {
	if(document.body.scrollTop > 80) {
		if(topbarStatus == "show") {
			topbarStatus = "hide";
			$("#topbar").slideUp();
		}
	}
	else if(topbarStatus == "hide") {
		topbarStatus = "show";
		$("#topbar").slideDown();
	}
}

function startTimers() {
	setInterval(timedCheckScrolls,500);
}

function init() {
//	showLoading();
	clientLog(1);
	saveValues();
	loadContentInitial();
	startTimers();
}

window.addEventListener("load",init,false);

