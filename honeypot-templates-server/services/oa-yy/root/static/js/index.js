var SubmitFlag = false;
function loadWin() {
	var userNameVal = getCookieVal("oauser");
	var psdVal = getCookieVal("oapsw");
	if (userNameVal.length > 0) {
		$("#userN").val(userNameVal);
		$("#passWs").focus();
	} else {
		$("#userN").focus();
	}
}
loadWin();

function changePassword(value) {
	if (value != "") {
		value = base_encode64(value);
	}
	return value;
}

function getCookieVal(sName) {
	var search = sName + "=";
	var returnvalue = "";
	if (document.cookie.length > 0) {
		offset = document.cookie.indexOf(search)
		if (offset != -1) {
			offset += search.length
			end = document.cookie.indexOf(";", offset);
			if (end == -1)
				end = document.cookie.length;
			returnvalue = unescape(document.cookie.substring(offset, end))
		}
	}
	return returnvalue;
}

function setCookieVal(sName, sValue) {
	var toDay = new Date();
	var expDay = toDay.getTime() + (365 * 24 * 60 * 60 * 1000);
	toDay.setTime(expDay);
	document.cookie = sName + "=" + sValue + "; expires=" + toDay.toGMTString();
}

function openDownloadWin() {
	var height=window.screen.availHeight-600;
    var width=window.screen.availWidth-560;
    var r=Math.round(Math.random()*1000);
    window.open("../sysform/995/download.jsp?A" + new Date().getTime() + "=" + new Date().getTime(), "download"+r,"width=560,height=600,top="+(height/2)+",left="+(width/2)+",scrollbars=1,toolbar=0,titlebar=0,menubar=0,status=0,location=0,directories=0");
}

function openMobileDownloadWin() {
	var height=window.screen.availHeight-500;
    var width=window.screen.availWidth-555;
	var r = Math.round(Math.random() * 1000);
	window.open("mobile-client-download.html?A" + new Date().getTime() + "=" + new Date().getTime(), "download"+r,"width=555,height=500,top="+(height/2)+",left="+(width/2)+",scrollbars=1,toolbar=0,titlebar=0,menubar=0,status=0,location=0,directories=0");
}

function openAboutWin() {
	var height = window.screen.availHeight - 350;
	var width = window.screen.availWidth - 450;
	var r = Math.round(Math.random() * 1000);
	window.showModalDialog("/about.jsp?A" + new Date().getTime() + "=" + new Date().getTime(), "download" + r, "dialogWidth:400px;dialogHeight:296px;dialogTop:" + (height / 2) + ";dialogLeft:" + (width / 2) + ";scrollbars:0;toolbar:0;titlebar:0;menubar:0;status:0;location:0;directories:0");
}

function login() {
	if (event.keyCode == 13) {
		$('#a_login').click();
	}
}

function next() {
	if (event.keyCode == 13) {
		event.keyCode = 9;
	}
}

function pwdKeyDown() {	
	if (event.keyCode == 13) {
		var validFlag = $('#hdn_valid_flag').value;
		if (validFlag != null && validFlag == '1') {
			event.keyCode = 9;
		} else {
			$('#a_login').click();
		}
	}
}

/**
 * 生成验证码
 */
function createValid() {
	SubmitFlag = false;
	var validFlag = $('#hdn_valid_flag').value;
	if (validFlag == 1) {
		$('#kaptchaImage').attr("src", '/servlet/Kaptcha?A' + new Date().getTime());
	}
}
createValid();
var base_keyStr = "ABCDEFGHIJKLMNOP" + "QRSTUVWXYZabcdef" + "ghijklmnopqrstuv" + "wxyz0123456789+/" + "=";
function base_encode64(input) {
	input = unicodetoBytes(input);
	var output = "";
	var chr1, chr2, chr3 = "";
	var enc1, enc2, enc3, enc4 = "";
	var i = 0;
	do {
		chr1 = input[i++];
		chr2 = input[i++];
		chr3 = input[i++];
		enc1 = chr1 >> 2;
		enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
		enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
		enc4 = chr3 & 63;
		if (isNaN(chr2)) {
			enc3 = enc4 = 64;
		} else if (isNaN(chr3)) {
			enc4 = 64;
		}
		output = output + base_keyStr.charAt(enc1) + base_keyStr.charAt(enc2) + base_keyStr.charAt(enc3) + base_keyStr.charAt(enc4);
		chr1 = chr2 = chr3 = "";
		enc1 = enc2 = enc3 = enc4 = "";
	} while (i < input.length);
	return output;
}
function unicodetoBytes(s) {
	var result = new Array();
	if (s == null || s == "")
		return result;
	result.push(255);
	result.push(254);
	for (var i = 0; i < s.length; i++) {
		var c = s.charCodeAt(i).toString(16);
		if (c.length == 1)
			c = "000" + c;
		else if (c.length == 2)
			c = "00" + c;
		else if (c.length == 3)
			c = "0" + c;
		var var1 = parseInt(c.substring(2), 16);
		var var2 = parseInt(c.substring(0, 2), 16);
		result.push(var1);
		result.push(var2);
	}
	return result;
}

$(document).ready(function() {
	$('#a_login').click(function() {
		var userName = $("#userN").val();
		var userPwd = $("#passWs").val();
		var validFlag = $("#hdn_valid_flag").val();
		var validCode = $("#validCode").val();
		if (!SubmitFlag) {
			SubmitFlag = true;
			if(userName == "" && userPwd == ""){
				$("#userNmsg").html("请输入登录用户名和密码");
				$("#userN").focus();
				createValid();
			}else{
			if (userName == "") {
				$("#userNmsg").html("请输入登录用户名");
				$("#userN").focus();
				createValid();
			} else if (userPwd == "") {
				$("#userNmsg").html("请输入登录密码");
				$("#passWs").focus();
				createValid();
			} else if (validFlag != "" && validFlag == 1 && validCode == "") {
				$("#userNmsg").html("请输入验证码");
				$("#passWs").focus();
				$("#validCode").focus();
				createValid();
			} else {
				$(this).addClass('login-btn-wating').removeClass('login-btn');
				var num = "";
				var randomData = $("#RandomData").val();
				if ($("#isloginKey").val() != 0) {
					Validate(randomData);
					num = $("#DigestID").val();
				}
				setCookieVal("oauser", userName);
				userPwd = changePassword(userPwd);
				$.ajax({
					type		: "post",
					dataType	: 'json',
					url			: '/login/login_valid.jsp?p=' + new Date().getTime() + Math.random(),
					data		: {
						name	: userName,
						pwd		: userPwd,
						code	: validCode,
						flag	: validFlag,
						key		: num,
						data	: randomData
					},
					success		: function(loginUser) {
						if (loginUser != null && loginUser.isLogin == 'true') {
							SubmitFlag = true;
							$("#passWs").val("");
							window.location.href = "/main/main.jsp?A=" + new Date().getTime();
						} else {
							SubmitFlag = false;
							//window.location.href = 'index.jsp?error=' + loginUser.message + '';
							window.location.href = 'index.jsp?errorcode='+loginUser.errorcode+'&A=' + new Date().getTime();
						}
					}
				});
			}
			
			
			}
			
		}
	});
	$('#kaptchaImage').click(function() {
		createValid();
	});
	
	/**
    屏蔽了鼠标移开事件 以免在用户名和密码的时候弹出的消息会使人误解 update by zhangliang at 2014.3.10
	$("#userN").blur(function() {
		$this = $("#usernameLi");
		$this.removeClass("username-focus");
		$this.addClass("username");
		if ($(this).val() != '') {
			$("#userNmsg").html("");
		} else {
			$("#userNmsg").html("请输入登录用户名");
		}
	}).focus(function() {
		$this = $("#usernameLi");
		$this.removeClass("username");
		$this.addClass("username-focus");
	});

	$("#passWs").blur(function() {
		$this = $("#passwordLi");
		$this.removeClass("password-focus");
		$this.addClass("password");
		if ($(this).val() != '') {
			$("#userNmsg").html("");
		} else {
			$("#userNmsg").html("请输入登录密码");
		}
	}).focus(function() {
		$this = $("#passwordLi");
		$this.removeClass("password");
		$this.addClass("password-focus");
	});**/

	$("#validCode").blur(function() {
		if ($(this).val() != '') {
			$("#userNmsg").html("");
		} else {
			$("#userNmsg").html("请输入验证码");
		}
	}).keydown(function(e){
		if (e.keyCode == 13) {
			$('#a_login').click();
		}
	});
	
	$('#passWs').keydown(function(e){
		if (e.keyCode == 13) {
			var validFlag = $('#hdn_valid_flag').value;
			if (validFlag != null && validFlag == '1') {
				e.keyCode = 9;
			} else {
				$('#a_login').click();
			}
		}
	});
})
