// JavaScript Document
(function($){
$.su = $.su || {};
$.su.dutMode = 1 ;
$.su.dutModeName = "" ;
$.su.dutManageMode = "" ;
$.su.dutWorkMode = "" ;

$.su.regionSupport5G = true;

$.su.func = {};
$.su.func.ipToInt = function(ip){
	var patternIp = /^\s*[0-9]{1,3}\.{1}[0-9]{1,3}\.{1}[0-9]{1,3}\.{1}[0-9]{1,3}\s*$/;
    var ipArray = ip.split(".");
	
    if(ipArray.length != 4){
        return -1;      
    }
	
    if(!patternIp.test(ip)){
        return -1;
    }
	
    return (Number(ipArray[0])*(1<<24)+(Number(ipArray[1])<<16 |Number(ipArray[2])<<8 |Number(ipArray[3])));
};

$.su.func.isSameNet = function(ip1, ip2, mask){
	if(ip1 == ""){
        return false;
    };

    var ipToInt = $.su.func.ipToInt;

    var intIp1 = ipToInt(ip1);
    var intIp2 = ipToInt(ip2);
    var intMask = ipToInt(mask);
    if(intMask == 0)
    {
    	return false;
    }
	

    if(((intIp1&intMask)!=(intIp2&intMask)) || ((intIp1&0xffffffff)==(intIp2&intMask))||((intIp1&(~intMask))==(~intMask))){
        return false;
    };
    return true;
};

$.su.func.isNetIp = function(ip, mask){
	var ipToInt = $.su.func.ipToInt;

    if(!ip){
		return false;
	};
	var ipint = ipToInt(ip);
	var maskint = ipToInt(mask);

	if(0x00000000 == (ipint & (~maskint))){
		return true;
	};

	return false;
};

$.su.func.isNetIpLegal = function(ip, mask){
	var ipToInt = $.su.func.ipToInt;

	if (!ip){
		return false;
	};

	var ipint = ipToInt(ip);
	var maskint = ipToInt(mask);
	var res = ipint & maskint;

	if (res == 0x00000000 || res == maskint){
		return false;
	};

	return true;
}

$.su.func.isBroadCastIp = function(ip, mask){
	var ipToInt = $.su.func.ipToInt;

	if(!ip){
		return false;
	};

	var ipint = ipToInt(ip);
	var maskint = ipToInt(mask);

	if(((ipint & (~maskint)) == (~maskint))){
		return true;
	};

	return false;
};

$.su.func.escapeHtml = function(string){

	var r = string.toString();
		r = r.replace(/\&/g, "&amp;");
		r = r.replace(/\</g, "&lt;");
		r = r.replace(/\>/g, "&gt;");
		r = r.replace(/\"/g, "&quot;");
		r = r.replace(/\s/g ,"&nbsp;");
		
	return r;
};

$.su.vtype = function(options){
	var defaults = {
		type: "sample",
		regex: /^[a-zA-Z0-9]&/,
		vtypeText: "sample run!",
		validator: null
	};
	//$.extend(types, $.su.vtype);
	var name = "",
		opt = {};

	if ($.type(options) === "string"){
		name = options;
	}else if ($.type(options) === "object" && options.vtype){
		name = options.vtype;
		opt = options;
	};

	if (!$.su.vtype.types[name]){
		//console.error("wrong vtype name!");
		return null;
	}else{
		defaults = $.su.vtype.types[name];
	};

	$.extend(this, defaults, opt, {isVtype: true});
	//console.log(this.validate())
};
$.su.vtype.types = {
	email: {
		regex: /^[a-zA-Z0-9_+.-]+\@([a-zA-Z0-9-]+\.)+[a-zA-Z0-9]{2,4}$/,
		vtypeText: $.su.CHAR.VTYPETEXT.EMAIL
	},
	number: {
		regex: /^-?[0-9]\d*$/,
		vtypeText: $.su.CHAR.VTYPETEXT.NUMBER,
		validator: function(value){
			value = parseInt(value, 10);
			if (this.max !== null || this.max !== undefined){
				if (value > this.max){
					if (this.min !== null || this.min !== undefined)
					{
					      var str = $.su.CHAR.VTYPETEXT.NUMBER_MIN_MAX.replace("%min", this.min.toString()).replace("%max", this.max.toString());
						return str;
					}
					else
					{
						var str = $.su.CHAR.VTYPETEXT.NUMBER_MAX.replace("%max", this.max.toString());
						return 	str;
					}
				};
			};

			if (this.min !== null || this.min !== undefined){
				if (value < this.min){
					if (this.max !== null || this.max !== undefined)
					{
						var str = $.su.CHAR.VTYPETEXT.NUMBER_MIN_MAX.replace("%min", this.min.toString()).replace("%max", this.max.toString());
						return str;
					}
					else
					{
						var str = $.su.CHAR.VTYPETEXT.NUMBER_MIN.replace("%min", this.min.toString());
						return 	str;
						//return 	$.su.CHAR.VTYPETEXT.NUMBER_MIN.replace("%min", this.min.toString());
					}
				};
			};

			return true;
		},
		keybordHandler: function(e){
			e.stopPropagation();
			var keyCode = e.keyCode,
				shiftKey = e.shiftKey,
				ctrlKey = e.ctrlKey;

			if (shiftKey){
				return false;
			};

			if (keyCode == 37 || keyCode == 39){
				return true;
			};

			if (keyCode == 38 || keyCode == 40){
				var input = $(this),
					_value = input.val();

				if ($.su.vtype.types.number.regex.test(_value)){
					if (keyCode == 38){
						var _maxValue = input.hasClass("hour-text") ? 23: 59;
						if (_value < _maxValue){
							input.val(parseInt(_value, 10)+1);
						}else{
							return false;
						};
					}else{
						if (_value == 0){
							return false;
						}else{
							input.val(parseInt(_value, 10)-1);
						}
					};
				}else{
					input.val(0);
				};
			};

			if (!ctrlKey){
				if ((keyCode < 48 || keyCode > 57) && keyCode > 32){
					return false;
				};
			};
		}
	},
	string: {
		regex: /^[^\o\t\n\f\r\v]+$/	/*,
		maxLength: null,
		minLength: null,
		validator: function(value){
			//var value = 
		}*/
	},
	date:{
		//regex: /^[(0[0-9]{1})|(1[0,2]{1})]]$/,
		regex: /^(0[1-9]|1[0-2])\/(0[1-9]|1[0-9]|2[0-9]|3[0-1])\/([0-9]{4})$/,
		vtypeText: $.su.CHAR.VTYPETEXT.DATE,
		validator:function(val)
		{
			//format:  mm/dd/yy
			
			var arr = val.split("/"); 
			if(parseInt(arr[2],10) < 1980)
			{
				return false;
			}
			if(parseInt(arr[2],10) > 2030)
			{
				return false;
			}
			var formatDate = arr[2] + "/" + arr[0] + "/" + arr[1];
			return (new Date(formatDate).getDate()==formatDate.substring(formatDate.length-2));
		}
	},
	ip_domain:{
		regex: /^((\d{1,2}|1\d\d|2[0-4]\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){3}|([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6})$/,
		vtypeText: $.su.CHAR.VTYPETEXT.IP_DOMAIN
	},
	ipv6:{
		isPrefixFlag:false,
		regex: /^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$/,
		vtypeText: $.su.CHAR.VTYPETEXT.IPV6,
		validator:function(val)
		{
		       var reg1 = new RegExp("^[2-3][0-9A-Fa-f]{1,3}:");
			if(!reg1.test(val))
			{
				return $.su.CHAR.VTYPETEXT.IPV6_NOT_GLOBAL;
			}
			var reg2 = new RegExp("::$");
			if(this.isPrefixFlag)
			{
				if(!reg2.test(val))
				{
					return $.su.CHAR.VTYPETEXT.IPV6_NOT_PREFIX;
				}
			}
			else
			{
				if(reg2.test(val))
				{
					return $.su.CHAR.VTYPETEXT.IPV6_NOT_GLOBAL;
				}
			}
			return true;
		}
	},
	ip: {
		//regex: /^[0-9]{1,3}\.{1}[0-9]{1,3}\.{1}[0-9]{1,3}\.{1}[0-9]{1,3}$/,
		//regex: /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
		
		allowAllZeroFlag:false,
		disallowAllZeroText: $.su.CHAR.VTYPETEXT.IP_NO_ALL_ZERO,
		allowLoopFlag:false,
		disallowLoopText: $.su.CHAR.VTYPETEXT.IP_NO_LOOP,
		allowDTypeFlag:false,
		disallowDTypeText: $.su.CHAR.VTYPETEXT.IP_NO_D_TYPE,
		allowETypeFlag:false,
		disallowETypeText: $.su.CHAR.VTYPETEXT.IP_NO_E_TYPE,
		allowAllOneFlag:false,
		disallowAllOneText: $.su.CHAR.VTYPETEXT.IP_NO_ALL_ONE,
		disallowFirstAllOneText: $.su.CHAR.VTYPETEXT.IP_NO_FIRST_ALL_ONE,
		disallowBoardCastText: $.su.CHAR.VTYPETEXT.IP_NO_BOARDCAST_IP,
		disallowNetIpText: $.su.CHAR.VTYPETEXT.IP_NO_NET_IP,
		disallowStartWithZeroText: $.su.CHAR.VTYPETEXT.IP_NO_FIRST_ZERO,
		regex: /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){3}$/,
		validator: function(value){
			ipStr = value;
			var szarray = [0,0,0,0];
			var remain;
			var i;
			for(i = 0; i < 3; i++)
			{
				var n = ipStr.indexOf('.');
				if (n == -1){
					szarray[i] = ipStr;
				}else{
					szarray[i] = ipStr.substring(0,n);
					remain = ipStr.substring(n+1);
					ipStr = remain;
				}
			}
			if (n != -1){
				szarray[3] = remain;
			}
			
			for(i = 0; i < 4; i++)
			{
				if (szarray[i] < 0 || szarray[i] > 255)
				{
					return false;
				}
			}
			if (szarray[0] == 0)
			{
				return this.disallowStartWithZeroText;
			}
			/*å°†çŽ¯å›žåœ°å€çš„æ£€æµ?ç”?27.0.0.xxxæ”¾å®½åˆ?27.xxx.xxx.xxx, - 2010-04-15 whf */
			/*if(szarray[0]==127 && szarray[1]==0 && szarray[2]==0)*/ //æ£€æŸ¥çŽ¯å›žåœ°å€
/*			if(!this.allowLoopFlag)
			{
				if(szarray[0]==127)
				{
					return this.disallowLoopText;
				}
			}
			if(!this.allowDTypeFlag)
			{
				if(szarray[0] >= 224 && szarray[0] <=239) //æ£€æŸ¥å¤šæ’?
				{
					return this.disallowDTypeText;
				}
			}
			if(!this.allowETypeFlag)
			{
				if(szarray[0] >= 240 && szarray[0] <=254) //æ£€æŸ¥å¤šæ’?
				{
					return this.disallowETypeText;
				}
			}
			if(!this.allowAllOneFlag)
			{
				if(szarray[0]==255) //æ£€æŸ¥æ˜¯å¦ä¸º255.XXX.XXX.XXX
				{
					return this.disallowAllOneText;
				}
				else if(szarray[3]==255)
				{
					return this.disallowBoardCastText;
				}
			}
			else
			{
				if(szarray[0]==255 && szarray[1] ==255 && szarray[2] ==255 && szarray[3] ==255) 
				{
					
				}
				else
				{
					if(szarray[0]==255) //æ£€æŸ¥æ˜¯å¦ä¸º255.XXX.XXX.XXX
					{
						return this.disallowFirstAllOneText;
					}
				}
			}
			if(!this.allowAllZeroFlag)
			{
				if(szarray[0]==0 && szarray[1] ==0 && szarray[2] ==0 && szarray[3] ==0) 
				{
					return this.disallowAllZeroText;
				}
				else if(szarray[3] ==0)
				{
					return this.disallowNetIpText;
				}
			}*/

			//255.XXX.XXX.XXX
			if (szarray[0] == 255)
			{
				return this.disallowFirstAllOneText;
			}
			//E class
		    	if (szarray[0] >= 240 && szarray[0] <= 254)
			{
				return this.disallowETypeText;
			}
			//Muti bcast
			if(szarray[0] >= 224 && szarray[0] <= 239)
			{
				return this.disallowDTypeText;
			}
			//127.xxx.xxx.xxx
			if(szarray[0] == 127)
			{
				return this.disallowLoopText;
			}
			return true;
		},
		vtypeText: $.su.CHAR.VTYPETEXT.IP
	},
	ip_no_zero: {

	},
	mac: {
		regex: /^[a-fA-F\d]{2}\-[a-fA-F\d]{2}\-[a-fA-F\d]{2}\-[a-fA-F\d]{2}\-[a-fA-F\d]{2}\-[a-fA-F\d]{2}$/,
		validator: function(value){
			var patternMulti = /^\s*[0-9A-Fa-f]{1}[13579bdfBDF]{1}(\-[A-Fa-f0-9]{2}){5}\s*$/;
			var flag = patternMulti.test(value);
			if(flag)
			{	
				return $.su.CHAR.VTYPETEXT.CAST_MAC;
			}
			return true;
		},
		vtypeText: $.su.CHAR.VTYPETEXT.MAC
	},
	netmask: {
		//regex: /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){3}$/,
		allowAllOneFlag:false,
		allowAllZeroFlag:false,
		allowEndWith254Flag:false,
		disallowAllOneText: $.su.CHAR.VTYPETEXT.MASK_NO_ALL_ONE,
		disallowAllZeroText: $.su.CHAR.VTYPETEXT.MASK_NO_ALL_ZERO,
		disallowEndWith254Text: $.su.CHAR.VTYPETEXT.MASK_NO_END_WITH_254,
		regex: /^(254|252|248|240|224|192|128|0)\.0\.0\.0$|^(255\.(254|252|248|240|224|192|128|0)\.0\.0)$|^(255\.255\.(254|252|248|240|224|192|128|0)\.0)$|^(255\.255\.255\.(254|252|248|240|224|192|128|0))$|^255.255.255.255$/ ,
		vtypeText: $.su.CHAR.VTYPETEXT.MASK,
		validator: function(value){
			//console.log(value);
			if(value == "255.255.255.255")
			{
				//console.log(this.allowAllOneFlag);
				if(this.allowAllOneFlag)
				{
					return true;
				}
				else
				{
					return this.disallowAllOneText;
				}
			}
			else if(value == "0.0.0.0")
			{
				//console.log(this.allowAllOneFlag);
				if(this.allowAllZeroFlag)
				{
					return true;
				}
				else
				{
					return this.disallowAllZeroText;
				}
			}
			else if(value == "255.255.255.254")
			{
				//console.log(this.allowAllOneFlag);
				if(this.allowEndWith254Flag)
				{
					return true;
				}
				else
				{
					return this.disallowEndWith254Text;
				}
			}
			return true;

		}
	},
	domain: {
		regex: /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$/,
		vtypeText: $.su.CHAR.VTYPETEXT.DOMAIN
	},
	ascii_visible:{
		regex: /^[\x21-\x7e]+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.ASCII_VISIBLE
	},
	string_visible:{
		regex: /^\S+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.STRING_VISIBLE
	},
	string_strict:{
		//regex: /^[A-Za-z0-9\`\~\!\@\$\*\(\)\-\_\[\]\{\}\;\:\'\"\\\/\?\.\,\%\/\ ]+$/,
		regex: /^[^\+\#\&\=\|\<\>\^\\]+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.STRING_STRICT
	},
	string_strict_strict:{
		regex: /^[^\+\#\&\=\|\<\>\^\\\~\`\!\@\$\%\*\(\)\[\]\{\}\;\:\'\"\,\.\/\?]+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.STRING_STRICT_STRICT
	},
	
	password:{
		regex: /^[A-Za-z0-9\`\~\!\@\#\$\&\*\(\)\-\=\_\+\[\]\{\}\;\:\'\"\\\|\/\?\.\,\<\>\%\^\/\ ]+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.PWD
	},
	string_visible_allow_blank:{
		regex: /^(\S|\x20)+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.STRING_VISIBLE_ALLOW_BLANK
	},
	name:{
		regex: /^[A-Za-z0-9\_]+$/,
		vtypeText: $.su.CHAR.VTYPETEXT.NAME
	},
	note: {
		regex: /^[A-Za-z0-9\`\~\!\@\#\$\%\^\&\*\(\)\-\=\_\+\[\]\{\}\;\:\'\"\\\|\/\?\.\,\<\>\ ]*$/,
		vtypeText: $.su.CHAR.VTYPETEXT.NOTE
	}
};
$.su.vtype.prototype.validate = function(value){
	var resultVal = true;
	var resultReg = true;

	if (this.validator){
		resultVal = this.validator(value);
		if (resultVal !== true){
			return resultVal;
		};
	};

	if (this.regex){
		resultReg = this.regex.test(value);
		if (resultReg !== true){
			return this.vtypeText;
		};
	};

	return true;
};

//$.su.create = function(className, options){
//	var obj = null;
//	if ($.su[className]){
		//obj = new $.su[className](options);
//		return obj;
//	}/*else if ($.su.grid[className]){
//		obj = new $.su.grid[className](options);
//		return obj;
//	}*/else{
//		//console.error("Debug: className error!");
//		return null;
//	};
//};
$.su.clearAttrAll = function(obj){
	var obj = $(obj);

	obj.removeAttr("id");
	obj.removeAttr("name");
	obj.removeAttr("class");
};
$.su.clone = function(obj){	//éœ€è¦éªŒè¯ä¸€ä¸‹è¿™ä¸ªå‡½æ•°æœ‰æ²¡æœ‰é—®é¢˜ï¿?
	var type = Object.prototype.toString.call(obj);
	switch (type){
		case "[object Array]":
			var result = [];
			for (var index = 0; index < obj.length; index++){
				result[index] = $.su.clone(obj[index]);
			};
			return result;
			break;
		case "[object Object]":
			var result = {};
			for (var index in obj){
				result[index] = $.su.clone(obj[index]);
			};
			return result;
			break;
		case "[object Function]":
			return obj;
			break;
		case "[object String]":
			var result = obj;
			return result;
			break;
		case "[object Number]":
			var result = obj;
			return result;
			break;
		case "[object Boolean]":
			if (!obj){
				return false;
			}else{
				return true;
			}
			break;
		default:
			return undefined;
			break;
	};
};

$.su.format = {
	"string" : function(str){
		return str.toString();
	},
	"int": function(int){
		return parseInt(int, 10);
	},
	"boolean": function(bool){
		if (!bool){
			return false;
		}else{
			return true;
		};
	}
};

$.su.randomId = function(type){
	return type+"-"+parseInt(Math.random()*1000*1000*1000*10, 10).toString();
};

/*Manageréƒ¨åˆ†*/
$.su.Manager = function(options){
	var config = {
		type: "Store",
		id: $.su.randomId("manager")
	};
	this.local = {};
	this.global = {};

	$.extend(this, config, options);

	$.su.Manager.managers = $.su.Manager.managers || [];
	$.su.Manager.managers.push(this);
};
$.su.Manager.removeLocal = function(){
	var managers = $.su.Manager.managers;
	//console.log(managers);
	if (managers){
		for (var index = 0, len = managers.length; index < len; index++){
			if (managers[index].removeLocal){
				managers[index].removeLocal();
			};
		};
	};
};
$.su.Manager.removeGlobal = function(){
	var managers = $.su.Manager.managers;
	if (managers){
		for (var index = 0, len = managers.length; index < len; index++){
			if (managers[index].removeGlobal){
				managers[index].removeGlobal();
			};
		};
	};
};

$.su.Manager.prototype.get = function(id){
	var store = this.local[id] || this.global[id];
	if (!store){
		//console.error("Debug: Invalid id!");
		return null;
	};
	return store;
};
$.su.Manager.prototype.add = function(obj){
	var id = obj.id;
	if (!id){
		//console.log("Debug: Store Id!");
		return false;
	};
	var me = this;
	if (obj.global){
		me.global[id] = obj;
	}else{
		me.local[id] = obj;
	};
	if (me.addCallback){
		me.addCallback(obj);
	};
};
$.su.Manager.prototype.remove = function(id){
	var obj = this.local[id] || this.global[id];
	if (!obj){
		//console.error("Debug: Invalid id!");
	};
	if (obj.nodeType){
		$(obj).remove();
	}else{
		delete obj;
	}
	delete this.local[id];
	delete this.global[id];
	obj = null;

	return this;
};
$.su.Manager.prototype.removeLocal = function(){
	for (var index in this.local){
		var obj = this.local[index];
		if (obj.nodeType){
			$(obj).remove();
		}else{
			delete obj;
		};
	};

	this.local = null;
	delete this.local;
	this.local = {};
	return this;
};
$.su.Manager.prototype.removeGlobal = function(){
	for (var index in this.global){
		var obj = this.global[index];
		if (obj.nodeType){
			$(obj).remove();
		}else{
			delete obj;
		};
	};
	this.global = null;
	delete this.global;
	this.global = {};
	return this;
};

$.su.storeManager = new $.su.Manager({
	type: "Stroe",
	id: "storeManager"
});

$.su.componentManager = new $.su.Manager({
	type: "Component",
	id: "componentManager"
});

$.su.tipManager = new $.su.Manager({
	type: "Tip",
	id: "tipManager"
});

$.su.msgManager = new $.su.Manager({
	type: "Msg",
	id: "msgManager"
});

$.su.tipManager.add = function(obj) {
	$.su.Manager.prototype.add.call(this, obj);

	var container = $("div.tips-container#tips-container").length === 0 ? $("<div class=\"tips-container\" id=\"tips-container\"></div>").appendTo("body") : $("div.tips-container#tips-container");
	container.append(obj);
};

/*$.su.tipManager.addCallback = function(obj){
	var id = obj.id;
	if (!id){
		//console.log("Debug: Tips Id!");
		return false;
	};
};*/

/*æå–èœå•æž„é€ å‡½æ•°ï¼Œé€šç”¨åªbasic*/
$.su.Menu = function(options){
	var defaults = {
		container: "",	//å®¹å™¨çš„id
		data: "",		//èœå•ç»“æž„æ•°æ®çš„è·¯å¾?
		urlPath: "",	//åŠŸèƒ½é¡µé¢çš„ä¸»è·¯å¾„
		cls: "",
		idTag: "",
		handlers: null	//init, click
	};

	var settings = $.extend({}, defaults, options);
	var menu = this;

	this.settings = settings;
	$.ajax({
		url: settings.data,
		type: "GET",
		cache:false,
		dataType: "json",
		success: function(data, status, xhr){
			menu.init(data, settings);
		},
		error:function(){
			//console.dir(arguments);
		}
	});
	//console.log(settings.handlers.goTo)
	//è‡ªèº«äº‹ä»¶ç›‘å¬
	if (this.settings.handlers){
		if (settings.handlers.init){
			$(this).on("ev_init", this.settings.handlers.init);
		};
		if (settings.handlers.goTo){
			$(this).on("ev_goto", this.settings.handlers.goTo);
		};
	};
	
};
$.su.Menu.prototype.init = function(data, settings){
	var container = $(settings.container).addClass("menu-container").html("&nbsp;"),
		menu = this;

	if (container.length < 0){
		return;
	};

	//var inHTML = 	"<div>";
	var inHTML = "<ul class=\"menu-tree "+settings.cls+"\">";

	for (var index = 0; index < data.length; index++){
		var fstMenu = data[index];
		
		fstMenu.auth = fstMenu.auth || 1;
		if (fstMenu.auth === 0){
			continue;
		};
		
		if (fstMenu.children){
			var secMenus = fstMenu.children;
			//æœ‰å­èœå•çš„æƒ…å†?
			inHTML += 	"<li class=\"fst\" id=\""+"menu-"+settings.idTag+"-"+fstMenu.name+"-li"+"\">";
			inHTML += 		"<a class=\"fst\" data-name=\""+fstMenu.name+"\" name=\""+fstMenu.name+"\" href=\""+settings.urlPath+secMenus[0].url+"\">";
			inHTML += 			"<span class=\"icon\"></span>";
			inHTML += 			"<span class=\"text\">"+$.su.CHAR.MENU[fstMenu.text]+"</span>";
			inHTML += 		"</a>";

			inHTML += 		"<div class=\"menu-sec-wrap\">";
			inHTML += 			"<ul class=\"sec\">";
			for (var jndex = 0, len = secMenus.length; jndex < len; jndex++){
				var secMenu = secMenus[jndex],
					pos = (jndex == 0) ? "sec-fst" : ((jndex == len - 1) ? "sec-lst" : "");
				
				secMenu.auth = secMenu.auth || 1;
				if (secMenu.auth === 0){
					continue;
				};

				inHTML += 			"<li class=\"sec "+pos+"\">";
				inHTML +=				"<a class=\"sec\" data-name=\""+secMenu.name+"\" name=\""+secMenu.name+"\" href=\""+settings.urlPath+secMenu.url+"\">";
				inHTML +=					"<span class=\"icon\"></span>";
				inHTML +=					"<span class=\"text\">"+$.su.CHAR.MENU[secMenu.text]+"</span>";
				inHTML +=				"</a>";
				inHTML +=			"</li>";
			};

			inHTML += 			"</ul>";
			inHTML +=		"</div>";
			inHTML +=	"</li>";

		}else{
			//æ²¡æœ‰å­èœå•çš„æƒ…å†µ
			inHTML +=	"<li class=\"fst\" id=\""+"menu-"+settings.idTag+"-"+fstMenu.name+"-li"+"\">";
			inHTML +=		"<a class=\"fst none-sec\" data-name=\""+fstMenu.name+"\" name=\""+fstMenu.name+"\" href=\""+settings.urlPath+fstMenu.url+"\">";
			inHTML +=			"<span class=\"icon\"></span>";
			inHTML +=			"<span class=\"text\">"+$.su.CHAR.MENU[fstMenu.text]+"</span>"
			inHTML +=		"</a>";
			inHTML +=	"</li>";
		}
	};
	
	inHTML += "</ul>";
	//inHTML +=		"</div>";
/*	alert(container.get(0).innerHTML);
	alert(inHTML);*/
	container.append($(inHTML));

	//äº‹ä»¶ç›‘å¬éƒ¨åˆ†
	container.delegate("ul.menu-tree li.fst a.fst", "click", function(e){
		e.preventDefault();
		e.stopPropagation();

		var me = $(this)
			name = me.attr("data-name");
		
		if (me.hasClass("deployed")){
			//èœå•å·²ç»å±•å¼€
			menu.collapse(name);
		}else{
			//èœå•æœªå±•å¼€
			menu.goTo(name);
		};

	}).delegate("ul.menu-tree li.sec a.sec", "click", function(e){
		e.preventDefault();
		e.stopPropagation();

		var name = $(this).attr("data-name");
		menu.goTo(name);
	});

	$(this).trigger("ev_init");
};
$.su.Menu.prototype.goTo = function(name){
	var	menuItem = null,
		container = $(this.settings.container),
		menuItem = container.find("ul.menu-tree li a[name="+name+"]"),
		href = "";

	if (!menuItem.length){
		return;
	};
	//console.log("111")
	if (menuItem.hasClass("fst")){
		//ä¸€çº§èœå?
		container.find("ul.menu-tree li, ul.menu-tree li a").removeClass("selected deployed");
		menuItem.addClass("selected");
		menuItem.closest("li.fst").addClass("selected");

		var secMenuList = menuItem.closest("li.fst").find("ul.sec");
		if (secMenuList.length == 0){
			//æ— äºŒçº§èœå?
			container.find("ul.menu-tree li ul.sec").slideUp(150);
			href = menuItem.attr("href");
		}else{
			//æœ‰äºŒçº§èœå?
			if (secMenuList.is(":hidden")){
				//äºŒçº§èœå•å·²ç»éšè—
				container.find("ul.menu-tree li ul.sec").slideUp(150);
				//console.log(menuItem)
				menuItem.addClass("deployed");
				secMenuList.slideDown(120, function(){
					//$(this).trigger("ev_secopen", [name, href]);
					$.su.layout.doLayout();
				});
			};
			//liæ ‡ç­¾çš„æ“ä½?
			secMenuList.find("li.sec, a.sec").removeClass("selected");
			secMenuList.find("li.sec:first, li.sec a.sec:first").addClass("selected");
			href = secMenuList.find("li.sec a.sec:first").attr("href");
		};
	}else{
		//äºŒçº§èœå•
		var	secMenuList = menuItem.closest("ul.sec"),
			href = menuItem.attr("href");

		if (secMenuList.is(":hidden")){
			container.find("ul.menu-tree li, ul.menu-tree li a").removeClass("selected deployed");
			container.find("ul.menu-tree li ul.sec").slideUp(150);
			
			secMenuList.slideDown(120, function(){
				$.su.layout.doLayout();
			});

			secMenuList.closest("li.fst").addClass("selected").find("a.fst").addClass("selected deployed");
		};

		secMenuList.find("li.sec, a.sec").removeClass("selected");
		menuItem.addClass("selected").closest("li.sec").addClass("selected");
	};

	$(this).trigger("ev_goto", [name, href]);
};
$.su.Menu.prototype.collapse = function(name){
	//console.log("collapse")
	var	menuItem = null,
		container = $(this.settings.container),
		menuItem = container.find("ul.menu-tree li a[name="+name+"]");
	
	if (!menuItem.length || !menuItem.hasClass("fst")){
		return;
	}else{
		menuItem.next("div.menu-sec-wrap").find("ul.sec").slideUp(150);
		menuItem.removeClass("deployed");
	};
};


/*$.su.Controller = function(options){
	handlers: {
		navChange: [
			$.su.Manager.removeLocal
		],
		menuChange: [
			$.su.Manager.removeLocal
		]
	},
	runHandlers: function(handlerId){
		var handlers = this.handlers[handlerId];
		for (var index = 0, len = handlers.length; index < len; index++){
			if (typeof handlers[index] == "function"){
				handlers[index]();
			}
		};
	}
};*/
/*$.su.Controller = function(options){
	var defaults = {

	};
	$.extend(this, defaults, options);
	this.init();
};
$.su.Controller.prototype.init = function(){};
$.su.Controller.prototype.control = function(options){

	for (var index in options){
		var obj = options[index];
		if (obj){
			//æ˜¯æ¡†æž¶ä¸Šçš„å¯¹è±?
			for (var ev in obj){
				var fn = obj[ev];
				if ($.isFunction(fn)){
				};
			}
		}else{
			//
		};
	};
};

$.su.appController = new $.su.Controller({
	init: function(){
		this.control({
			"nav": {
				ev_goto: function(){
					alert("rrrr")
					$.su.Manager.removeLocal();
				}
			}
		})
	}
});
*/
/*$.su.moduleController = new $.su.Controller({
	handlers
});*/
$.su.app = null;

$.su.mask = {
	dom: $("div.mask#mask").length == 0 ? $("<div id=\"mask\" class=\"mask\"></div>").appendTo($("body")) : $("div.mask#mask"),
	displayed: false,
	height: 0,
	width: 0,
	show: function(){
		var _this = this;
		_this.dom.fadeIn(200, function(){
			$(this).css("display", "block");
			_this.displayed = true;
		});
	},
	hide: function(){
		var _this = this;
		_this.dom.fadeOut(200, function(){
			$(this).css("display", "none");
			_this.displayed = false;
		})
	},
	init: function(){
		//åˆå§‹åŒ–é•¿ï¿?
		var _this = this;
		_this.height = $(window).height() > $(document).height() ? $(window).height() : $(document).height();
		_this.width = $(window).width() > $(document).width() ? $(window).width() : $(document).width();

		var changeSize = function(){
			_this.height = $(window).height() > $(document).height() ? $(window).height() : $(document).height();
			_this.width = $(window).width() > $(document).width() ? $(window).width() : $(document).width();
			
			_this.dom.width(_this.width).height(_this.height);
			
			//console.log("run changeSize");
		};

		$(window).on("resize", changeSize);
		//$("body").resize(changeSize);
	}
};

$.su.mask.init();


/* é”®ç›˜å“åº”çš„æ‰©ï¿? */
$.su.addEventListener = function(obj, name, func, capture){		//è¿™ä¸ªäº‹ä»¶ç»‘å®šçš„æ²¡åŠžæ³•removeæŽ‰ï¼ç‰¹æ®Šæƒ…å†µæ‰ç”¨è¿™ä¸ªå‡½æ•°
	if (!obj){
		//console.error("The Obj is null!");
		return;
	};

	if (obj.addEventListener){
		obj.addEventListener(name, function(ev){
			func.call(obj, ev || window.event);
		}, capture);
	}else{
		obj.attachEvent("on"+name, function(ev){
			func.call(obj, ev || window.event);
		});
	};

	return func;
};

$.su.mouse = {
	/*longpress: function(fn, callback){
		var me = $(this),
			t = 0;
		me.mousedown(function(e){
			t = setInterval(function(){
				if ($.type(fn) === "function"){
					fn.call(me, e);
				}
			}, 100);
		}).bind("mouseup mouseleave", function(e){
			clearInterval(t);
			if ($.type(callback) === "function"){
				callback.call(me, e);
			}
		});
	}*/
	mousescroll: function(handler){
		/*var type = $.type(obj);
		if (type === "string"){
			obj = $(obj).get(0);
		}else if(type === "object"){
			if (obj.jquery){
				obj = obj.get(0);
			};
		};*/
		var obj = this.get(0);
		if (!obj || !handler){
			//console.error("The Obj or the handler is null!");
			return;
		};

		var wheelType = "mousewheel";
		try{
			document.createEvent("MouseScrollEvents");
			wheelType = "DOMMouseScroll";
		}catch(e){};

		$.su.addEventListener(obj, wheelType, function(e){
			if ("detail" in e){
				e.wheelDelta = e.detail*(-40);
			};
			//console.log(e.wheelDelta)
			handler.call(obj, e);
		});
	},
	draggable: function(options){
		var me = this;
		var defaults = {
			zIndex: 99,
			opacity: 0.7,
			handle: null,
			containment: "parent",
			axis: false,
			distance: 1,
			//callback
			drag: null,
			start: null,
			stop: null
		};

		var dd = {
			_init: function(){
	
				
				this._start();
			},
			_setContainment: function(){
				//containment
				switch(this.containment){
					case "window":
						this.containment = [
							$(window).scrollLeft() - this.position.relative.left - this.offset.parent.left,
							$(window).scrollTop() - this.position.relative.top - this.offset.top,
							$(window).scrollLeft() + $(window).width() - this.margins.left,
							$(window).scrollTop() + ($(window).height() || document.body.parentNode.scrollHeight) - this.margins.top 
						]
						break;
					case "document":
						this.containment = [
							0,
							0,
							$(document).width() - this.margins.left,
							($(document).height() || document.body.parentNode.scrollHeight) - this.margins.top
						]
						break;
					case "parent":
						break;
				};
			},
			_cacheMargins: function(){

			},
			_destroy: function(){

			},
			_capture: function(){

			},
			_start: function(){
				//è®¾ç½®margin
				this._cacheMargins();
			},
			_drag: function(){

			},
			_stop: function(){

			},
			_cancel: function(){

			}
		};

		me.each(function(i, obj){
			$.extend(dd, defaults, options);
			dd.element = obj;
			obj.dragdrop = dd;

			obj.dragdrop._init();
		});




		/*var settings = $.extend({}, defaults, options);
		var dd = {
			drag: function(e){
				var dragData = e.data.dragData;
				dragData.target.css({
					left: dragData.left + e.pageX - dragData.offLeft,
					top: dragData.top + e.pageY - dragData.offTop
				});
				// dragData.handler.css({
				// 	cursor: "move"
				// });
				dragData.onMove(e);
				//console.log("dd.drag")
			},
			drop: function(e){
				var dragData = e.data.dragData;
				dragData.onDrop(e);
				$().unbind("mousemove", dd.drag).unbind("mouseup", dd.drop);
				//console.log("dd.drop")
			}
		};

		me.each(function(i, obj){
			var handler = null;
			if (!settings.handler){
				handler = $(obj);
			}else{
				handler = $(settings.handler);
			};

			handler.bind("mousedown", {e: me}, function(s){
				var target = $(s.data.e).css("position", "absolute");
				var oldInfo = {
					cursor : target.css("cursor") || "default",
					opacity: target.css("opacity") || 1
				};

				var dragData = {
					left: target.offset().left || 0,
					top: target.offset().top || 0,
					offLeft: s.pageX,
					offTop: s.pageY,
					onMove: settings.onMove,
					onDrop: settings.onDrop,
					handler: handler,
					target: target
				};

				$().unbind("mousemove", {dragData: dragData}, dd.drag).unbind("mouseup", {dragData: dragData}, dd.drop);
			});
		});*/

		return me;
	}	
};

$.su.keybord = {

};


$.extend($.fn, $.su.mouse, $.su.keybord);

/*
$.ajaxSetup({
	cache: false,
	dataType: "json",
	error: function(){

	}
});
*/
$.su.Help = function(options){
	var defaults = {
		container: "",
		content: "",
		create: function(){
			//å»ºç«‹domå’Œäº‹ä»¶ç›‘å?
			var inHTML = 	"";
				inHTML +=	"<div class=\"help-container-wrap\">";

				inHTML +=		"<div class=\"help-btn-container\">";
				inHTML +=			"<a class=\"btn-help\" href=\"javascript:void(0);\">";
				inHTML +=			"</a>";
				inHTML +=		"</div>";
				
				
			// for ie shadow			
				
				inHTML +=		"<div class=\"help-content-container\"><span class=\"help-content-delta\"></span>";
				
				inHTML +=			'<div class="position-top-left"></div><div class="position-top-center"></div><div class="position-top-right"></div>';
				
				inHTML +=			'<div class="position-center-left"><div class="position-center-right">';

				inHTML +=				"<div class=\"help-content-wrap\">";
				inHTML +=					"<div class=\"help-content-container-wrap\">";
				inHTML +=					"<div class=\"help-content\"></div>";
				inHTML +=					"</div>";
				inHTML +=				"</div>";
				
				inHTML +=			'</div></div>';
				
				
				inHTML +=			'<div class="position-bottom-left"></div><div class="position-bottom-center"></div><div class="position-bottom-right"></div>';
				
				inHTML +=		"</div>";
				
			//////////////////////////	

				

				inHTML +=	"</div>";

			//console.log(this);
			var container = $(this.container);
			//console.log(container)
			container.addClass("help-container").append($(inHTML));

			return inHTML;

		},
		layout: function(){
			var wh = $(window).height();

			var help = this,
				wrap = help.wrap,
				contentContainer = help.contentContainer;

			wrap.css({
				"height": "inherit"
			});

			var h = contentContainer.height();

			if (h > wh){
				wrap.css({
					"height": wh-260
				});
			};
		}
	};

	var settings = $.extend({}, defaults, options);

	this.settings = settings;
	this.init();
	//return this.init();
};

$.su.Help.prototype.init = function(){
	var help = this;
	if (!help.settings){
		return null;
	};

	help.settings.create();

	var container = $(help.settings.container),
		contentContainer = container.find("div.help-content-container"),
		wrap = container.find("div.help-content-container-wrap"),
		content = container.find("div.help-content"),
		button = container.find("a.btn-help");

	help.container = container;
	help.wrap = wrap;
	help.content = content;
	help.contentContainer = contentContainer;
	help.button = button;

	button.on("click", function(e){
		e.preventDefault();
		e.stopPropagation();

		var btn = $(this);
		if (help.show){
			help.close();
		}else{
			help.load();
		};
	});

	//return this;
};

$.su.Help.prototype.load = function(name){
	var help = this;
	if (!help.loaded){
		//åŠ è½½å†…å®¹
		var render = function(contentObj){
		       var htmlStr = "";
			if(contentObj.TITLE)
			{
				htmlStr+="<h4 class=\"title\">"+contentObj.TITLE+"</h4>";
			}
			if(contentObj.CONTENT)
			{
				$(contentObj.CONTENT).each(function(index,element){
					htmlStr+=contentRender(element);
				})
			}
			return htmlStr;
		};
		var contentRender = function(obj){
			var htmlStr = "";
			var type = obj.type;
			var connector = obj.connector || " - ";
			switch(type){
				case "title":
					htmlStr +=	"<div class=\"name-container\">";
					htmlStr +=		"<h5 class=\"title title-title\">"+obj.content+"</h5>";
					htmlStr +=	"</div>";
					break;
				case "name":
					htmlStr +=	"<div class=\"name-container\">";
					htmlStr +=		"<h5 class=\"title name-title\">"+obj.title+"</h5>";
					htmlStr += 		"<span class=\"connector\">"+connector+"</span>";
					htmlStr +=		"<span>"+obj.content+"</span>";
					htmlStr +=	"</div>";
					break;
				case "paragraph":
					htmlStr +=	"<div class=\"paragraph-container\">";

					if (obj.title){
						htmlStr +=		"<h5 class=\"title paragraph-title\">"+obj.title+"</h5>";
					};
					
					htmlStr +=		"<p class=\"paragraph\">"+obj.content+"</p>";
					htmlStr +=	"</div>";
					break;
				case "note":
					htmlStr +=	"<div class=\"note-container\">";
					htmlStr +=		"<h5 class=\"title note-title\">"+obj.title+"</h5>";
					htmlStr +=		"<ol class=\"note\">";
					$(obj.content).each(function(index,note){
						htmlStr +=		"<li>"+note+"</li>";
					});
					htmlStr +=		"</ol>";
					htmlStr +=	"</div>";
					break;
			};

			if(obj.children){
				htmlStr +=	"<div class=\"container sub-container  sub-container-"+index+"\">";
					$(obj.children).each(function(index,child){
						htmlStr += 		contentRender(child);
					});
				htmlStr +=	"</div>";
			};
			/*
			var resultStr = "";
			resultStr+="<label class='title'>"+obj.title+" </label>";
			resultStr+= " - ";
			resultStr+="<span>"+obj.content+"</span>";
			//console.log(obj.title);
			//console.dir(obj);
			//console.log(obj.children);
			if(obj.children)
			{
				$(obj.children).each(function(index,child){
					resultStr+="<div class='container' style='padding-left:10px;'>";
					resultStr+= render(child);
					resultStr+="</div>";
				});
			}*/
			return htmlStr;
		}



		var name = name || this.settings.content;
		if (!$.isArray(name)){
			name = [name];
		}
		//console.log(name);
		for (var index = 0, len = name.length; index < len; index++ ){
			var contentObj = $.su.CHAR.HELP[name[index]];
			var htmlStr = "";
			htmlStr = render(contentObj);
			this.content.append($(htmlStr));
	};
	
		help.loaded = true;		
	};
	
	help.settings.layout.call(this);

	help.contentContainer.slideDown(150, function(){
		$(this).css("display", "block");
		help.wrap.scrollTo(0);
		help.show = true;
	});

	help.contentContainer.on("click", function(e){
		e.stopPropagation();
		e.preventDefault();
	});

	var btn = help.button;
	btn.addClass("clicked");

	$("html").one("click", function(){
		help.close();
	});
};

$.su.Help.prototype.close = function(){
	var help = this;
	help.contentContainer.fadeOut(150, function(){
		$(this).css("display", "none");
		help.show = false;
	});
	var btn = help.button;
	btn.removeClass("clicked");
};
/*$.su.help = {
	width: 360,
	setPostion: function(){
		var windowObj = $(window),
			helpContainer = $("div.btn-help-container"),
			contentContainer = $("div.content-container"),
			btn = $("a.btn-help");

		var showFlag = btn.hasClass("closed");

		if (windowObj.scrollTop() < 85){	//è¿™ä¸ªé«˜åº¦æœ‰ç¡¬ç¼–ç ï¼Œè°ƒæ•´æ—¶èŽ·å–åˆå§‹ï¿?
			helpContainer.css({
				"position": "absolute",
				"right": (showFlag ? 0 : $.su.help.width - btn.width())
			});
		}else{
			var posR = windowObj.width() + windowObj.scrollLeft() - contentContainer.offset().left - contentContainer.width();
			helpContainer.css({
				"position": "fixed",
				"right": (showFlag ? posR : $.su.help.width + posR - btn.width())
			});
		};

		//$.su.help.setCurrent();
	},*/
	/*setCurrent: function(){
		var helpContainer = $("div.btn-help-container"),
			contents = $("div.help-content-container h3.title, div.help-content-container h4.title, div.help-content-container p");
		
		if (helpContainer.length == 0){
			return false;
		};

		var btnTop = helpContainer.offset().top,
			btnHeight = helpContainer.height();
		contents.each(function(i, obj){
			var objP = $(obj),
				objTop = objP.offset().top,
				objHeight = objP.outerHeight(),
				marginTop = parseInt(objP.css("margin-top"), 10),
				marginBot = parseInt(objP.css("margin-bottom"), 10);

			//console.log(objHeight, marginTop, marginBot)

			if (((objTop - marginTop) < btnTop) && ((btnHeight + btnTop) < (objHeight + objTop + marginBot))){
				objP.addClass("current");
				//console.log(objP.next())
				//objP.next().addClass("current");
			}else{
				objP.removeClass("current");
				//obj.prev().removeClass("current");
			};
		});
	},*/
	/*load: function(id){
		var content = $.su.CHAR.HELP[id];
		if (content){
			//console.log(content);
			
			var inHTML = 	"";
			if (content.TITLE){
				inHTML +=	"<h3 class=\"title\">"+content.TITLE+"</h3>";
			};

			if (content.PARAGRAPH){
				for (var index = 0; index < content.PARAGRAPH.length; index++){
					var p = content.PARAGRAPH[index];
						inHTML +=	"<p>";
					if (p.TITLE){
						inHTML +=		"<span class=\"title\">"+p.TITLE+"</span>";
						inHTML +=		"<span> - </span>"						
					};
					if (p.CONTENT){
						inHTML +=		"<span class=\"content\">"+p.CONTENT+"</span>";
					};
						inHTML +=	"</p>";
				};
			};

			if (content.NOTE){
				inHTML +=	"<h4 class=\"title\">"+$.su.CHAR.HELP.NOTE+"</h4>"
				for (var index = 0; index < content.NOTE.length; index++){
					var note = content.NOTE[index];
					inHTML +=	"<p>";
					inHTML +=		"<span class=\"content\">"+note.CONTENT+"</span>";
					inHTML +=	"</p>";
				};
			};
			$("div.help-content-container").empty().append($(inHTML));
		}else{
			//console.error("Can't load the help files~!");
			return false;
		}
		//$.su.help.setCurrent();
	},
	init: function(){
		/*$(window).bind("scroll resize", function(){
			$.su.help.setPostion();
			$.su.help.setCurrent();
		});*/
		//$(window).bind("scroll resize", $.su.help.setPostion);
	/*}
};
$.su.help.init();*/


$.su.layout = {
	doLayout: function(){

	}
};

//unicode
$.su.unicode = {
	encode: function(str) {
        var a = [];
        for (var i = 0; i < str.length; i++){
			var code = str.charCodeAt(i);
			code = ((code.toString(16).length) % 2) == 0 ? code.toString(16) : "0" + code.toString(16);
			a[i] = code;
		};
		//console.log(a);
        return "&#" + a.join(";&#") + ";"
    },
    decode: function(str) {
        return str.replace(/&#(x)?([^&]{1,6});?/g,
        function(a, b, c) {
            return parseInt("0x"+String, 16).fromCharCode(parseInt(c, b ? 16 : 10))
        })
    },
    encodeNoPrefix: function(str){
    	var unicode = this.encode(str);
		return unicode.replace(/[&#;]/g, "").toString();
    },
    getUnicodeLength: function(str){
    	var unicode = this.encode(str);
		unicode = unicode.replace(/[&#;]/g, "");
    	return Math.ceil( unicode.length / 2)
    },
	getUtf8Length: function(str){
		var unicode = this.encode(str);
		var len = 0;
		unicode.replace(/&#(x)?([^&]{1,6});?/g, 
		
		function(code){
			var utf8Code = code.replace(/\&\#/g, "0x").replace(/\;/g, "").toString();
			
			if (0x0000 <= utf8Code && utf8Code <= 0x007F){
				len = len + 1;
			}else if (0x0080 <= utf8Code && utf8Code <= 0x07FF){
				len = len + 2;
			}else if (0x0800 <= utf8Code && utf8Code <= 0xFFFF){
				len = len + 3;
			}else if (0x10000 <= utf8Code && utf8Code <= 0x1FFFFF){
				len = len + 4;
			}else if (0x200000 <= utf8Code && utf8Code <= 0x3FFFFFF){
				len = len + 5;
			}else if (0x4000000 <= utf8Code && utf8Code <= 0x7FFFFFFF){
				len = len + 6;
			}
		});

		return len;
	},
	checkUtf8Length: function(str, minLength, maxLength){
		var minLength = minLength || 0,
			maxLength = maxLength || "";
			
		var len = $.su.unicode.getUtf8Length(str);
		
		if (len < minLength){
			return false;
		};
		
		if (maxLength !== ""){
			if (maxLength < len){
				return false;
			};
		};
		
		return true;
		
	},
    checkUnicodeLength: function(str, minLength, maxLength){	//ä¼ å…¥çš„minLengthå’ŒmaxLengthç”¨å­—èŠ‚è®¡ç®?
    	var minLength = minLength*2 || 0,
			maxLength = maxLength*2 || "";

		var pattern = "^.{%minLength%,%maxLength%}$";
		pattern = pattern.replace("%minLength%", minLength).replace("%maxLength%", maxLength);
		pattern = new RegExp(pattern);
		//console.log(pattern, str)
		return pattern.test(this.encodeNoPrefix(str));
    }
}

$.su.loadComboData = {
	parseData:function(source, arr){
		var retArr = [];
		if(!$.isArray(arr)){
			return retArr;
		}
		$(source).each(
			function(index, element){
				var tempObj = {};
			    var len = arr.length;
				for(var i=0; i < len; i++){
					//console.log(i, arr, arr[i], len)
					if (arr[i]){
						if(arr[i].func){
							tempObj[arr[i].dstKey] = arr[i].func(element[arr[i].srcKey], arr[i].para);
						}else{
							tempObj[arr[i].dstKey] = element[arr[i].srcKey];
						};
					};
				}
				retArr.push(tempObj);
			}
		);
		return retArr;
	},
	load:function(url, para, arr){
		var me = this;
		var regionItem = [];
		$.ajax({url:url, async:false, dataType:"json", success:function(response,status,xhr){
			regionItem = me.parseData(response.data, arr);
		}
		});
		return regionItem;
		
	}
}

/****************************/

/*
json.js
2013-05-26

Public Domain

No warranty expressed or implied. Use at your own risk.

This file has been superceded by http://www.JSON.org/json2.js

See http://www.JSON.org/js.html

This code should be minified before deployment.
See http://javascript.crockford.com/jsmin.html

USE YOUR OWN COPY. IT IS EXTREMELY UNWISE TO LOAD CODE FROM SERVERS YOU DO
NOT CONTROL.

This file adds these methods to JavaScript:

object.toJSONString(whitelist)
This method produce a JSON text from a JavaScript value.
It must not contain any cyclical references. Illegal values
will be excluded.

The default conversion for dates is to an ISO string. You can
add a toJSONString method to any date object to get a different
representation.

The object and array methods can take an optional whitelist
argument. A whitelist is an array of strings. If it is provided,
keys in objects not found in the whitelist are excluded.

string.parseJSON(filter)
This method parses a JSON text to produce an object or
array. It can throw a SyntaxError exception.

The optional filter parameter is a function which can filter and
transform the results. It receives each of the keys and values, and
its return value is used instead of the original value. If it
returns what it received, then structure is not modified. If it
returns undefined then the member is deleted.

Example:

// Parse the text. If a key contains the string 'date' then
// convert the value to a date.

myData = text.parseJSON(function (key, value) {
return key.indexOf('date') >= 0 ? new Date(value) : value;
});

This file will break programs with improper for..in loops. See
http://yuiblog.com/blog/2006/09/26/for-in-intrigue/

This file creates a global JSON object containing two methods: stringify
and parse.

JSON.stringify(value, replacer, space)
value any JavaScript value, usually an object or array.

replacer an optional parameter that determines how object
values are stringified for objects. It can be a
function or an array of strings.

space an optional parameter that specifies the indentation
of nested structures. If it is omitted, the text will
be packed without extra whitespace. If it is a number,
it will specify the number of spaces to indent at each
level. If it is a string (such as '\t' or '&nbsp;'),
it contains the characters used to indent at each level.

This method produces a JSON text from a JavaScript value.

When an object value is found, if the object contains a toJSON
method, its toJSON method will be called and the result will be
stringified. A toJSON method does not serialize: it returns the
value represented by the name/value pair that should be serialized,
or undefined if nothing should be serialized. The toJSON method
will be passed the key associated with the value, and this will be
bound to the object holding the key.

For example, this would serialize Dates as ISO strings.

Date.prototype.toJSON = function (key) {
function f(n) {
// Format integers to have at least two digits.
return n < 10 ? '0' + n : n;
}

return this.getUTCFullYear() + '-' +
f(this.getUTCMonth() + 1) + '-' +
f(this.getUTCDate()) + 'T' +
f(this.getUTCHours()) + ':' +
f(this.getUTCMinutes()) + ':' +
f(this.getUTCSeconds()) + 'Z';
};

You can provide an optional replacer method. It will be passed the
key and value of each member, with this bound to the containing
object. The value that is returned from your method will be
serialized. If your method returns undefined, then the member will
be excluded from the serialization.

If the replacer parameter is an array of strings, then it will be
used to select the members to be serialized. It filters the results
such that only members with keys listed in the replacer array are
stringified.

Values that do not have JSON representations, such as undefined or
functions, will not be serialized. Such values in objects will be
dropped; in arrays they will be replaced with null. You can use
a replacer function to replace those with JSON values.
JSON.stringify(undefined) returns undefined.

The optional space parameter produces a stringification of the
value that is filled with line breaks and indentation to make it
easier to read.

If the space parameter is a non-empty string, then that string will
be used for indentation. If the space parameter is a number, then
the indentation will be that many spaces.

Example:

text = JSON.stringify(['e', {pluribus: 'unum'}]);
// text is '["e",{"pluribus":"unum"}]'


text = JSON.stringify(['e', {pluribus: 'unum'}], null, '\t');
// text is '[\n\t"e",\n\t{\n\t\t"pluribus": "unum"\n\t}\n]'

text = JSON.stringify([new Date()], function (key, value) {
return this[key] instanceof Date ?
'Date(' + this[key] + ')' : value;
});
// text is '["Date(---current time---)"]'


JSON.parse(text, reviver)
This method parses a JSON text to produce an object or array.
It can throw a SyntaxError exception.

The optional reviver parameter is a function that can filter and
transform the results. It receives each of the keys and values,
and its return value is used instead of the original value.
If it returns what it received, then the structure is not modified.
If it returns undefined then the member is deleted.

Example:

// Parse the text. Values that look like ISO date strings will
// be converted to Date objects.

myData = JSON.parse(text, function (key, value) {
var a;
if (typeof value === 'string') {
a =
/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2}(?:\.\d*)?)Z$/.exec(value);
if (a) {
return new Date(Date.UTC(+a[1], +a[2] - 1, +a[3], +a[4],
+a[5], +a[6]));
}
}
return value;
});

myData = JSON.parse('["Date(09/09/2001)"]', function (key, value) {
var d;
if (typeof value === 'string' &&
value.slice(0, 5) === 'Date(' &&
value.slice(-1) === ')') {
d = new Date(value.slice(5, -1));
if (d) {
return d;
}
}
return value;
});


This is a reference implementation. You are free to copy, modify, or
redistribute.
*/

/*jslint evil: true, regexp: true, unparam: true */

/*members "", "\b", "\t", "\n", "\f", "\r", "\"", JSON, "\\", apply,
call, charCodeAt, getUTCDate, getUTCFullYear, getUTCHours,
getUTCMinutes, getUTCMonth, getUTCSeconds, hasOwnProperty, join,
lastIndex, length, parse, parseJSON, prototype, push, replace, slice,
stringify, test, toJSON, toJSONString, toString, valueOf
*/


// Create a JSON object only if one does not already exist. We create the
// methods in a closure to avoid creating global variables.

if (typeof JSON !== 'object') {
    JSON = {};
}

(function () {
    'use strict';

    function f(n) {
        // Format integers to have at least two digits.
        return n < 10 ? '0' + n : n;
    }

    if (typeof Date.prototype.toJSON !== 'function') {

        Date.prototype.toJSON = function (key) {

            return isFinite(this.valueOf())
                ? this.getUTCFullYear() + '-' +
                    f(this.getUTCMonth() + 1) + '-' +
                    f(this.getUTCDate()) + 'T' +
                    f(this.getUTCHours()) + ':' +
                    f(this.getUTCMinutes()) + ':' +
                    f(this.getUTCSeconds()) + 'Z'
                : null;
        };

        String.prototype.toJSON =
            Number.prototype.toJSON =
            Boolean.prototype.toJSON = function (key) {
                return this.valueOf();
            };
    }

    var cx = /[\u0000\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g,
        escapable = /[\\\"\x00-\x1f\x7f-\x9f\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g,
        gap,
        indent,
        meta = { // table of character substitutions
            '\b': '\\b',
            '\t': '\\t',
            '\n': '\\n',
            '\f': '\\f',
            '\r': '\\r',
            '"' : '\\"',
            '\\': '\\\\'
        },
        rep;


    function quote(string) {

// If the string contains no control characters, no quote characters, and no
// backslash characters, then we can safely slap some quotes around it.
// Otherwise we must also replace the offending characters with safe escape
// sequences.

        escapable.lastIndex = 0;
        return escapable.test(string) ? '"' + string.replace(escapable, function (a) {
            var c = meta[a];
            return typeof c === 'string'
                ? c
                : '\\u' + ('0000' + a.charCodeAt(0).toString(16)).slice(-4);
        }) + '"' : '"' + string + '"';
    }


    function str(key, holder) {

// Produce a string from holder[key].

        var i, // The loop counter.
            k, // The member key.
            v, // The member value.
            length,
            mind = gap,
            partial,
            value = holder[key];

// If the value has a toJSON method, call it to obtain a replacement value.

        if (value && typeof value === 'object' &&
                typeof value.toJSON === 'function') {
            value = value.toJSON(key);
        }

// If we were called with a replacer function, then call the replacer to
// obtain a replacement value.

        if (typeof rep === 'function') {
            value = rep.call(holder, key, value);
        }

// What happens next depends on the value's type.

        switch (typeof value) {
        case 'string':
            return quote(value);

        case 'number':

// JSON numbers must be finite. Encode non-finite numbers as null.

            return isFinite(value) ? String(value) : 'null';

        case 'boolean':
        case 'null':

// If the value is a boolean or null, convert it to a string. Note:
// typeof null does not produce 'null'. The case is included here in
// the remote chance that this gets fixed someday.

            return String(value);

// If the type is 'object', we might be dealing with an object or an array or
// null.

        case 'object':

// Due to a specification blunder in ECMAScript, typeof null is 'object',
// so watch out for that case.

            if (!value) {
                return 'null';
            }

// Make an array to hold the partial results of stringifying this object value.

            gap += indent;
            partial = [];

// Is the value an array?

            if (Object.prototype.toString.apply(value) === '[object Array]') {

// The value is an array. Stringify every element. Use null as a placeholder
// for non-JSON values.

                length = value.length;
                for (i = 0; i < length; i += 1) {
                    partial[i] = str(i, value) || 'null';
                }

// Join all of the elements together, separated with commas, and wrap them in
// brackets.

                v = partial.length === 0
                    ? '[]'
                    : gap
                    ? '[\n' + gap + partial.join(',\n' + gap) + '\n' + mind + ']'
                    : '[' + partial.join(',') + ']';
                gap = mind;
                return v;
            }

// If the replacer is an array, use it to select the members to be stringified.

            if (rep && typeof rep === 'object') {
                length = rep.length;
                for (i = 0; i < length; i += 1) {
                    k = rep[i];
                    if (typeof k === 'string') {
                        v = str(k, value);
                        if (v) {
                            partial.push(quote(k) + (gap ? ': ' : ':') + v);
                        }
                    }
                }
            } else {

// Otherwise, iterate through all of the keys in the object.

                for (k in value) {
                    if (Object.prototype.hasOwnProperty.call(value, k)) {
                        v = str(k, value);
                        if (v) {
                            partial.push(quote(k) + (gap ? ': ' : ':') + v);
                        }
                    }
                }
            }

// Join all of the member texts together, separated with commas,
// and wrap them in braces.

            v = partial.length === 0 ? '{}'
                : gap
                ? '{\n' + gap + partial.join(',\n' + gap) + '\n' + mind + '}'
                : '{' + partial.join(',') + '}';
            gap = mind;
            return v;
        }
    }

// If the JSON object does not yet have a stringify method, give it one.

    if (typeof JSON.stringify !== 'function') {
        JSON.stringify = function (value, replacer, space) {

// The stringify method takes a value and an optional replacer, and an optional
// space parameter, and returns a JSON text. The replacer can be a function
// that can replace values, or an array of strings that will select the keys.
// A default replacer method can be provided. Use of the space parameter can
// produce text that is more easily readable.

            var i;
            gap = '';
            indent = '';

// If the space parameter is a number, make an indent string containing that
// many spaces.

            if (typeof space === 'number') {
                for (i = 0; i < space; i += 1) {
                    indent += ' ';
                }

// If the space parameter is a string, it will be used as the indent string.

            } else if (typeof space === 'string') {
                indent = space;
            }

// If there is a replacer, it must be a function or an array.
// Otherwise, throw an error.

            rep = replacer;
            if (replacer && typeof replacer !== 'function' &&
                    (typeof replacer !== 'object' ||
                    typeof replacer.length !== 'number')) {
                throw new Error('JSON.stringify');
            }

// Make a fake root object containing our value under the key of ''.
// Return the result of stringifying the value.

            return str('', {'': value});
        };
    }


// If the JSON object does not yet have a parse method, give it one.

    if (typeof JSON.parse !== 'function') {
        JSON.parse = function (text, reviver) {

// The parse method takes a text and an optional reviver function, and returns
// a JavaScript value if the text is a valid JSON text.

            var j;

            function walk(holder, key) {

// The walk method is used to recursively walk the resulting structure so
// that modifications can be made.

                var k, v, value = holder[key];
                if (value && typeof value === 'object') {
                    for (k in value) {
                        if (Object.prototype.hasOwnProperty.call(value, k)) {
                            v = walk(value, k);
                            if (v !== undefined) {
                                value[k] = v;
                            } else {
                                delete value[k];
                            }
                        }
                    }
                }
                return reviver.call(holder, key, value);
            }


// Parsing happens in four stages. In the first stage, we replace certain
// Unicode characters with escape sequences. JavaScript handles many characters
// incorrectly, either silently deleting them, or treating them as line endings.

            text = String(text);
            cx.lastIndex = 0;
            if (cx.test(text)) {
                text = text.replace(cx, function (a) {
                    return '\\u' +
                        ('0000' + a.charCodeAt(0).toString(16)).slice(-4);
                });
            }

// In the second stage, we run the text against regular expressions that look
// for non-JSON patterns. We are especially concerned with '()' and 'new'
// because they can cause invocation, and '=' because it can cause mutation.
// But just to be safe, we want to reject all unexpected forms.

// We split the second stage into 4 regexp operations in order to work around
// crippling inefficiencies in IE's and Safari's regexp engines. First we
// replace the JSON backslash pairs with '@' (a non-JSON character). Second, we
// replace all simple value tokens with ']' characters. Third, we delete all
// open brackets that follow a colon or comma or that begin the text. Finally,
// we look to see that the remaining characters are only whitespace or ']' or
// ',' or ':' or '{' or '}'. If that is so, then the text is safe for eval.

            if (/^[\],:{}\s]*$/
                    .test(text.replace(/\\(?:["\\\/bfnrt]|u[0-9a-fA-F]{4})/g, '@')
                        .replace(/"[^"\\\n\r]*"|true|false|null|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g, ']')
                        .replace(/(?:^|:|,)(?:\s*\[)+/g, ''))) {

// In the third stage we use the eval function to compile the text into a
// JavaScript structure. The '{' operator is subject to a syntactic ambiguity
// in JavaScript: it can begin a block or an object literal. We wrap the text
// in parens to eliminate the ambiguity.

                j = eval('(' + text + ')');

// In the optional fourth stage, we recursively walk the new structure, passing
// each name/value pair to a reviver function for possible transformation.

                return typeof reviver === 'function'
                    ? walk({'': j}, '')
                    : j;
            }

// If the text is not JSON parseable, then a SyntaxError is thrown.

            throw new SyntaxError('JSON.parse');
        };
    }

// Augment the basic prototypes if they have not already been augmented.
// These forms are obsolete. It is recommended that JSON.stringify and
// JSON.parse be used instead.

    /*if (!Object.prototype.toJSONString) {
        Object.prototype.toJSONString = function (filter) {
            return JSON.stringify(this, filter);
        };
        Object.prototype.parseJSON = function (filter) {
            return JSON.parse(this, filter);
        };
    }*/
	$.su.json = {};
	$.su.json.toJSONString = function(str, filter){
		return JSON.stringify(str, filter);
	};
	
	$.su.json.parseJSON = function(str, filter){
		return JSON.parse(str, filter);
	};

}());


})(jQuery);