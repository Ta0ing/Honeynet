/*
 让页面鼠标右键失效
*/
/* function clickIE4(){
         if (event.button==2){
                 return false;
         }
 }
  
 function clickNS4(e){
         if (document.layers||document.getElementById&&!document.all){
                 if (e.which==2||e.which==3){
                         return false;
                 }
         }
 }
  
 function OnDeny(){
         if(event.ctrlKey || event.keyCode==78 && event.ctrlKey || event.altKey || event.altKey && event.keyCode==115){
                 return false;
         }
 }
  
 if (document.layers){
         document.captureEvents(Event.MOUSEDOWN);
         document.onmousedown=clickNS4;
         document.onkeydown=OnDeny();
 }else if (document.all&&!document.getElementById){
         document.onmousedown=clickIE4;
         document.onkeydown=OnDeny();
 }
  
 document.oncontextmenu=new Function("return false");

*/

/*------------------prototype extend---------------------*/

Element.addMethods({
	hideAndDisable: function(element){ //隐藏并禁用
		element = $(element);
		element.hide();
		element.disabled = true;
    	return element;
	},
	showAndEnable: function(element){
		element = $(element);
		element.show();
		element.disabled = false;
		return element;
	},
	getPos: function(element){
		element = $(element);
		element.top = element.getBoundingClientRect().top + document.documentElement.scrollTop;
		element.left = element.getBoundingClientRect().left + document.documentElement.scrollLeft;
		element.right = element.getBoundingClientRect().right + document.documentElement.scrollLeft;
		element.bottom = element.getBoundingClientRect().bottom + document.documentElement.scrollTop;
		return element;
	}
});

/*------------------common function----------------------*/
/*
表格行滑过时效果类
参数table_id为使用该效果的表格
表格必须使用tbody
忽略该效果的行在该行上加class="noLight"
e.g:   new MoveOverLightTable("a_Table");
*/

function MoveOverLightTable(table_id){
	this.table_id = table_id;
	var self = this;
	
	this.regTable = function(){
		var trs = $$("#"+table_id+" tbody tr");
		
		if (trs.length == 0){
			return;
		}
		var bgcolor = trs[0].style.backgroundColor;
		for(var i = trs.length - 1; i >= 0; i--){
			if(trs[i].className.indexOf("noLight") == -1){
				trs[i].onmouseover = function(){
					this.style.backgroundColor = "#F6EEFC";
				}
				trs[i].onmouseout = function(){
					this.style.backgroundColor = bgcolor;
				}
			}
		}
	}
	Event.observe(window,'load',this.regTable);
}

/* 
确认删除类
new ConfirmDelForLink();
在需要确认的删除按钮或链接上添加class:delLink	
*/

function ConfirmDelForLink(){
	
	this.regConfirm = function(){
		var oArr = $$(".delLink");
		for(var i = oArr.length-1;i >= 0;i--){
			oArr[i].onclick = function(){
				if(confirm(this.title+"?")){
					return true;
				}else{
					return false;
				}
			}
		}
	}
	Event.observe(window,'load',this.regConfirm);
}

/* 
全选类
参数1：全选按钮的id
参数2：将被全部选中的className
页面加载完成时调用
e.g: new AllCheck('chk_btn','checks');
*/

var AllCheck = Class.create();

AllCheck.prototype = {
	initialize: function(btn_id,chk_class){
		$(btn_id).onclick = function(){
			var chks = $$("input."+chk_class);
			if($(btn_id).checked){
				for(var i = chks.length - 1; i >= 0; i--){
					chks[i].checked = true;
				}
			}else{
				for(var i = chks.length - 1; i >= 0; i--){
					chks[i].checked = false;
				}
			}
		}
	}
};

/*
表单检查类
new ValidateForm('loginForm',[checkEmail,checkEmpty]);
在页面加载完成时创建验证对象
参数1：要验证的Form ID
参数2：要调用的验证函数
参数3：
还需在需要验证的表单元素中添加相应的class(请查看验证函数)
如：<input class="isEmail" type="text">
如需要改变抛出信息方法，请修改throwErr函数。
*/

var ValidateForm = Class.create();

ValidateForm.prototype = {
	initialize:function(formId, valiRules, preFun) {
		this.idForm = formId;
		this.illegalArr = [];
		this.ruleList = { //过滤规则表
			noEmpty : "_o.value == ''",
			isEmail: "_o.value != '' && !/^([a-zA-Z0-9_.-])+@(([a-zA-Z0-9-])+.)+([a-zA-Z0-9]{2,4})+$/.test(_o.value)",
			isIp: "_o.value != '' && !/^(((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))\\.){3}((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))$/.test(_o.value)",
			isMac: "_o.value != '' && !/^([0-9a-fA-F]{2})(([/\\s:-][0-9a-fA-F]{2}){5})$/.test(_o.value)",
			isPort: "_o.value !='' && (!/^[0-9]*[1-9][0-9]*$/.test(_o.value) || _o.value < 0 || _o.value > 65535)",
			isIpRange: "_o.value !='' && !(/^(((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))\\.){3}((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))$/.test(_o.value) || /^(((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))\\.){3}((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))\\/([0-9]|1[0-9]|2[0-9]|3[0-2])$/.test(_o.value) || /^(((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))\\.){3}((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))-(((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))\\.){3}((\\d{1})|(\\d{2})|(1\\d{2})|(2[0-4]\\d{1})|(25[0-5]))$/.test(_o.value))",
			isInt: "_o.value !='' && !/^[0-9]*[1-9][0-9]*$/.test(_o.value)"
		};
		this.valiRules = valiRules;
		this.preFun = preFun ? preFun : function(){};

		//this.options = options? options : [];
		$(formId).onsubmit = this.validate.bind(this);
	},
	/* 规则处理 */
	collect: function(str){
		if(str != ''){
			var arr = [];
			var o = $$('#'+this.idForm+' .'+str)
			for(var i = o.length - 1; i >= 0; i--){
				var rules = this.ruleList[str].replace(/_o/g,'o[i]');
				if(eval('('+rules+')')){
					arr.push(o[i]);
				}
			}
		}
		return arr;
	},

	/* 收集未通过验证的表单元素和错误消息 */
	validate: function() {
		   
		if(!this.valiRules || this.valiRules.length < 1) return;
	   
		$$('#'+this.idForm+' input').each(function(node){
			node.value = node.value.replace(/(^\s*)|(\s*$)/g,"");   //清除前后空格
		});
	   
		$$('#'+this.idForm+' textarea').each(function(node){
			node.value = node.value.replace(/(^\s*)|(\s*$)/g,"");                                                   
		});
	   
		for(var i = 0; i < this.valiRules.length; i++){
			if(this.valiRules[i]=='checkIp2')
			{	
				
				var result = new checkIp2(this.idForm);	
			}else
			{
				var result = this.collect(this.valiRules[i]);
			}
			
			if(result){
				this.illegalArr = this.illegalArr.concat(result);
			}
		}

		for(var i = this.illegalArr.length - 1; i >= 0; i--){
			if(this.illegalArr[i].disabled || this.illegalArr[i].display == 'none'){
				this.illegalArr.splice(i,1);
			}	
		}
		return this.throwErr();
	},

	/* 抛出异常表单元素及消息 */
	throwErr: function() {
		var len = this.illegalArr.length;
		this.preFun();
		if(len > 0){
			for(var i = 0; i < len; i++){
				//this.illegalArr[i].style.backgroundColor = "#e42495";
				if(this.illegalArr[i].className.indexOf('error') < 0) this.illegalArr[i].addClassName('error');
				this.illegalArr[i].onfocus = clear;
			}

			this.illegalArr = [];
			return false;
		}
		return true;
	}
};


/* 清除错误提示 */
function clear(){
	this.className = this.className.replace("error","");
}


/* 验证函数 */

function checkWidth(idForm){
	var isWidth = $$('#'+idForm+' .isWidth');
	var arr = [];
	var rg = /^[0-9]*[1-9][0-9]*$/;
	if(isWidth.length == 0) return;
	
	for(var i = 0;i < isWidth.length;i++){
		if(
		   	!(  
				parseInt(isWidth[i].value) > 8 
				&& parseInt(isWidth[i].value) < 1000000 
				&& rg.test(isWidth[i].value)
			)
			&& ( isWidth[i].value != '未限制' 
				 && isWidth[i].value != '未预留' 
				 && isWidth[i].value != '未保障'
				 && isWidth[i].value != '' 
				 && isWidth[i].style.display != 'none' 
				 && isWidth[i].disabled == false 
				 )
		){
			arr.push(isWidth[i]);
		}
	}
	return arr;
}

function checkEmail(idForm){
	//var isEmail = Element.select($(idForm), '.isEmail'); //prototype 1.6+
	var isEmail = $$('#'+idForm+' .isEmail');
	var arr = [];
	if(isEmail.length == 0) return;
	var rg = /^([a-zA-Z0-9_.-])+@(([a-zA-Z0-9-])+.)+([a-zA-Z0-9]{2,4})+$/i;
	for(var i = 0;i < isEmail.length;i++){
		if (!rg.test(isEmail[i].value)){
			arr.push(isEmail[i]);
		}
	}
	return arr;
}

//检查textarea里的ip
function checkIp2(idForm){
	//var isIp = Element.select($(idForm), '.isIp'); prototype 1.6+
	var isIp2 = $$('#'+idForm+' .isIp2');
	var ipArr=[];	//检查IP
	var ipArr2=[];
	var subArr=[];		//检查掩码
	
	var arr=[];
	var rg = /^(((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\.){3}((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))$/;
	
	var rg2 = /^(254|252|248|240|224|192|128|0)\.0\.0\.0|255\.(254|252|248|240|224|192|128|0)\.0\.0|255\.255\.(254|252|248|240|224|192|128|0)\.0|255\.255\.255\.(254|252|248|240|224|192|128|0)$/;
	var rg3 = /^([0-9]|1[0-9]|2[0-9]|3[0-2])$/;	

	var str1="All";
	var str2="全部";	
	var str3="全部";

	for(var i=0;i<isIp2.length;i++)	//获得所有调用该验证的个数
	{
		if(str1==isIp2[i].value || str2==isIp2[i].value || str3==isIp2[i].value)
		{			
				;	
			
		}else
		{
			
		ipArr=isIp2[i].value.split(/\n/g);   //查找每一个textarea里的ip总个数，以回车符分割;ipArr[i]为总行数组  
		
		
		
		for(var j=0;j<ipArr.length;j++)	//查找出的每一个IP判断是否带有掩码，若有则提取出IP
		{
			//----该if语句是用来判断是否为IE浏览器，是的话，则少取一个字符（IE取字符与FF不同）
			if((navigator.appName.indexOf("Microsoft")!= -1) && (j<ipArr.length-1)) 
			{
  			 	ipArr[j]=ipArr[j].substring(0,ipArr[j].length-1);
  			}
			
			if(ipArr[j].search(/\//g)!=-1)		//若不等于-1，说明该IP带有掩码,提取IP  ipArr[j]为每行ip数组（含掩码）
			{
				subArr.push(ipArr[j].substring(ipArr[j].search(/\//g)+1));	//提取掩码
				ipArr[j]=ipArr[j].substring(0,ipArr[j].search(/\//g));		//提取含掩码的ip
				
				//subArr[j]=ipArr[j].substring(ipArr[j].search(/\//g)+1));
			}else if(ipArr[j].search(/\-/g)!=-1)
			{
				ipArr2.push(ipArr[j].substring(ipArr[j].search(/\-/g)+1));
				ipArr[j]=ipArr[j].substring(0,ipArr[j].search(/\-/g));				
			}
		
			
			/*--------------------------判断IP是否合法-----------------*/
			if(isIp2.length == 0){return;}	

			if (isIp2[i].disabled == false && isIp2[i].value != "" && !rg.test(ipArr[j]))
			{
				arr.push(isIp2[i]);		
				
			}else if(isIp2[i].disabled == false && isIp2[i].value && ipArr2[j]!=null)
			{
				for(var k=0;k<ipArr2.length;k++)
				{
					if(!(rg.test(ipArr2[k])))	
					{
						arr.push(isIp2[i]);	
					}
				}
			}else if(isIp2[i].value != '' && isIp2[i].disabled == false && isIp2[i].style.display != 'none' )			
			{			
				
				for(var k=0;k<subArr.length;k++)
				{
					
					if(!(rg2.test(subArr[k]) || rg3.test(subArr[k])))
					{
					arr.push(isIp2[i]);	
					}
				}
			}
		}	
		}
	}	
	return arr;
}



function checkEmpty(idForm,idArr){
	
	var isNotEmpty = $$('#'+idForm+' .isNotEmpty');
	var arr = [];
	if(isNotEmpty.length == 0) return;
	
	
	for(var i=0;i<isNotEmpty.length;i++){
	
		if(idArr!=null){
				if(idArr.length!=0)
				{
					
					var dis=($(idArr[i]).style.display=='none')?true:false;		
					if(!dis&& isNotEmpty[i].disabled == false && isNotEmpty[i].value=='')
					{
						arr.push(isNotEmpty[i]);
					}
				}else if(isNotEmpty[i].style.display!='none' && isNotEmpty[i].disabled == false && isNotEmpty[i].value=='')
				{
					
					arr.push(isNotEmpty[i]);	
				}
		}else if(isNotEmpty[i].style.display!='none' && isNotEmpty[i].disabled == false && isNotEmpty[i].value=='')
		{
			
			arr.push(isNotEmpty[i]);	
		}
	}
	return arr;
}

// check phone number format
function checkNum(idForm){
	//var isNum = Element.select($(idForm), '.isNum'); //prototype 1.6+
	var isNum = $$('#'+idForm+' .isNum');
	var arr = [];
	var rg = /^[0-9]+$/;
	if(isNum.length == 0) return;
	for(var i = 0;i < isNum.length;i++){
		if (isNum[i].value != "" && isNum[i].disabled == false && isNum[i].style.display != 'none' && !rg.test(isNum[i].value)){
			arr.push(isNum[i]);
		}
	}
	return arr;
}


// check num (0-128)
function checkNum128(idForm)
{
	
	var isNum128=$$('#'+idForm+' .isNum128');	
	var arr=[];
	var rg=/^[0-9]\d{0,2}$/;
	if(isNum128.length==0) return ;
	for(var i=0;i<isNum128.length;i++)
	{
		
		if(isNum128[i].value!="" &&
			isNum128[i].disabled==false &&
			!(rg.test(isNum128[i].value)?((isNum128[i].value>=0 && isNum128[i].value<=128)?true:false):false))
		{
			
			arr.push(isNum128[i]);	
		}
	}
	
	return arr;
	
}

// check num (0-900)
function checkNum900(idForm)
{
	
	var isNum900=$$('#'+idForm+' .isNum900');	
	var arr=[];
	var rg=/^[0-9]\d{0,2}$/;
	if(isNum900.length==0) return ;
	for(var i=0;i<isNum900.length;i++)
	{
		
		if(isNum900[i].value!="" &&
			isNum900[i].disabled==false &&
			!(rg.test(isNum900[i].value)?((isNum900[i].value>=0 && isNum900[i].value<=900)?true:false):false))
		{
			
			arr.push(isNum900[i]);	
		}
	}
	
	return arr;
	
}


//-----check num (120-172800)---
function checkNum1728(idForm)
{
	var isNum1728=$$('#'+idForm+' .isNum1728');	
	var arr=[];
	var rg=/^[1-9]\d{2,5}$/;
	if(isNum1728.length==0) return ;
	for(var i=0;i<isNum1728.length;i++)
	{
		if(isNum1728[i].value!="" &&
			isNum1728[i].disabled==false &&
			!(rg.test(isNum1728[i].value)?((isNum1728[i].value>=120 && isNum1728[i].value<=172800)?true:false):false))
		{
			arr.push(isNum1728[i]);	
		}
	}
	return arr;
}

function checkIp(idForm){
	//var isIp = Element.select($(idForm), '.isIp'); prototype 1.6+
	var isIp = $$('#'+idForm+' .isIp');
	var arr = [];
	var rg = /^(((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\.){3}((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))$/;
	if(isIp.length == 0){return;}
	for(var i = 0;i < isIp.length;i++){
		if (isIp[i].disabled == false && isIp[i].value != "" && !rg.test(isIp[i].value)){
			arr.push(isIp[i]);
		}
	}
	return arr;
}



//该函数用来检测输入的ipv6是否合法
function checkIPv6(idForm)
{
	
	var isIPv6=$$('#'+idForm+' .isIPv6');	
	var arr=[];
	rg=/^([\da-fA-F]{1,4}:){6}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^::([\da-fA-F]{1,4}:){0,4}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:):([\da-fA-F]{1,4}:){0,3}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){2}:([\da-fA-F]{1,4}:){0,2}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){3}:([\da-fA-F]{1,4}:){0,1}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){4}:((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}$|^:((:[\da-fA-F]{1,4}){1,6}|:)$|^[\da-fA-F]{1,4}:((:[\da-fA-F]{1,4}){1,5}|:)$|^([\da-fA-F]{1,4}:){2}((:[\da-fA-F]{1,4}){1,4}|:)$|^([\da-fA-F]{1,4}:){3}((:[\da-fA-F]{1,4}){1,3}|:)$|^([\da-fA-F]{1,4}:){4}((:[\da-fA-F]{1,4}){1,2}|:)$|^([\da-fA-F]{1,4}:){5}:([\da-fA-F]{1,4})?$|^([\da-fA-F]{1,4}:){6}:$/;	
	
	if(isIPv6.length==0){return;}

	for(var i=0;i<isIPv6.length;i++)
	{		
		if(isIPv6[i].value!=""&&
		   isIPv6[i].disabled==false&&
		   isIPv6[i].style.display!='none')
		{
			if(!(rg.test(isIPv6[i].value)))
			{
				arr.push(isIPv6[i]);	
			}
		}
	}

	return arr;
}

//check protocol(1-254)
function checkProtocol(idForm){

	var isProtocol = $$('#'+idForm+' .isProtocol');
	var arr = [];
	var rg = /^[1-9]\d{0,2}$/;
	/*alert(((rg.test(isDigit[0].value))?((isDigit[0].value>=120&&isDigit[0].value<=172800)?true:false):false)+'--------test.value');
	alert('exp is right');*/
	if(isProtocol.length == 0){return;}
	
	for(var i = 0;i < isProtocol.length;i++){
		if (isProtocol[i].disabled == false && 
			isProtocol[i].value != "" && 
			!(((rg.test(isProtocol[i].value))?((isProtocol[i].value>=1&&isProtocol[i].value<=254)?true:false):false))){
			arr.push(isProtocol[i]);
		}
	}
	return arr;
}


function checkIpRange(idForm){
	//var isIp = Element.select($(idForm), '.isIp'); prototype 1.6+
	var isIpRange = $$('#'+idForm+' .isIpRange');
	var arr = [];
	var rg = /^(((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\.){3}((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))$/;
	var rg2 = /^(((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\.){3}((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\/([0-9]|1[0-9]|2[0-9]|3[0-2])$/;
	var rg3 = /^(((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\.){3}((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))-(((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))\.){3}((\d{1})|(\d{2})|(1\d{2})|(2[0-4]\d{1})|(25[0-5]))$/;
	
	if(isIpRange.length == 0){return;}
	
	for(var i = 0;i < isIpRange.length;i++){
		if (isIpRange[i].value != "" && 
			isIpRange[i].disabled == false && 
			isIpRange[i].style.display != "none" && 
			isIpRange[i].className.indexOf("hide")== -1 &&
			!(
			  	rg.test(isIpRange[i].value) || 
				rg2.test(isIpRange[i].value) || 
				rg3.test(isIpRange[i].value) || 
				isIpRange[i].value == "全部" || 
				isIpRange[i].value.toLowerCase() == 'all' 
			)){
			arr.push(isIpRange[i]);
		}
	}
	return arr;
}

function checkPort(idForm){
	var isPort = $$('#'+idForm+' .isPort');
	var arr = [];
	var rg = /^[0-9]*[1-9][0-9]*$/;
	if(isPort.length == 0){return;}
	
	for(var i = 0;i < isPort.length;i++){
		if ( isPort[i].value !="" && (!rg.test(isPort[i].value) || isPort[i].value < 0 || isPort[i].value > 65535)){
			arr.push(isPort[i]);
		}
	}
	return arr;
}

function checkMac(idForm){
	var isMac = $$('#'+idForm+' .isMac');
	var arr = [];
	var rg = /^([0-9a-fA-F]{2})(([/\s:][0-9a-fA-F]{2}){5})$/;
	if(isMac.length == 0){return;}
	
	for(var i = 0;i < isMac.length;i++){
		if (isMac[i].value != "" && isMac[i].disabled == false && !rg.test(isMac[i].value)){
			arr.push(isMac[i]);
		}
	}
	return arr;
}

function checkKey(idForm){
	
	var isKey = $$('#'+idForm+' .isKey');
	var arr = [];
	var rg = /[\u4e00-\u9fa5]/;
	if(isKey.length == 0){return;}
	
	for(var i = 0;i < isKey.length;i++){
		if (isKey[i].value != "" && rg.test(isKey[i].value) && isKey[i].value.replace(/[^\x00-\xff]/ig,'aa').length > 20){
			arr.push(isKey[i]);
		}
	}
	return arr;
}

function checkName(idForm){
	
	var isName = $$('#'+idForm+' .isName');
	var arr = [];
	if(isName.length == 0){return;}
	
	for(var i = 0;i < isName.length;i++){
		if (isName[i].value != "" && isName[i].value.replace(/[^\x00-\xff]/ig,'aa').length > 20){
			arr.push(isName[i]);
		}
	}
	return arr;
}



/* Tab 类 */

function TabClass(){}

TabClass.prototype = {
	init : function(menus,divs,func){
				
		if(menus.length != divs.length){
			alert("菜单层数量和内容层数量不一样!");
			return false;
		}
		
		for(var i = 0 ; i < menus.length ; i++){
			$(menus[i]).index = i;
			$(menus[i]).onclick = function(){
				for(var j = 0 ; j < menus.length ; j++) {
					$(menus[j]).className = "";
					if($(divs[j]).className.indexOf('hide') == -1){
						$(divs[j]).className += " hide";
					}
					
					$$("#"+divs[j]+" input").each(function(node){
						node.disabled = true;
					});
				}
				if(typeof func == 'function'){
					func(this.id);
				}
				$(menus[this.index]).className = "active";
				$(divs[this.index]).className = $(divs[this.index]).className.replace('hide','');
				
				$$("#"+divs[this.index]+" input").each(function(node){
					node.disabled = false;											
				});
				return false;
			}
		}
	}
};

function SaveTable(chkId,colClass){ //保存表格显示方式到cookie
	//alert(document.cookie);
	var path = location.pathname.replace(/\.php/g,'');
	path = path.replace(/\/systemdata\//g,'');
	var coo = getCookie(path+'hide');
	if(coo != null){
		var arr = coo.split('-').without('');
		arr.each(function(s){
			$$('table.list .'+s).each(function(o){
				if(o){ 
					o.hide(); 
				}
			});
		});
	}
	
	var coo = getCookie(path+'nochk');
	if(coo != null){
		var arr = coo.split('-').without('');
		arr.each(function(o){
			if($(o)){ $(o).checked = false; }
		});
	}
	
	chkId.each(function(o,index){
		o.observe('click',function(){
			var cls = o.id;
			var PHPSESSID = getCookie('PHPSESSID');
			//alert('---------');
			//alert(PHPSESSID);
			
			if(!this.checked){
				$$('table.list .'+cls).invoke('hide');
				
				if(getCookie(path+'hide') == null){
					setCookie(path+'hide','-'+cls);
					setCookie(path+'nochk','-'+this.id);
					setCookie('PHPSESSID',PHPSESSID);
				}else{
					var newCookie = getCookie(path+'hide')+'-'+cls;
					var newCookie1 = getCookie(path+'nochk')+'-'+this.id;
					setCookie(path+'hide',newCookie);
					setCookie(path+'nochk',newCookie1);
					setCookie('PHPSESSID',PHPSESSID);
				}
			}else{
				$$('table.list .'+cls).invoke('show');
				if(getCookie(path+'hide') != null){
					var newCookie = getCookie(path+'hide').replace('-'+cls,'');
					var newCookie1 = getCookie(path+'nochk').replace('-'+this.id,'');
					setCookie(path+'hide',newCookie);
					setCookie(path+'nochk',newCookie1);
					setCookie('PHPSESSID',PHPSESSID);
				}
			}
			//alert(document.cookie);
		});
	});
}

function showTable(){ //展开或收缩表格
	//e.stop();
	var showBt = $$('caption span.title');
	if(showBt.length == 0){return;}
	showBt.each(function(node){
		node.observe('click',function(){
			node.toggleClassName('open');
			//node.style.backgroundImage = node.style.backgroundImage == '../images/caption_open_icon.png' ? '../images/caption_cloase_icon.png' : '../images/caption_open_icon.png';
			var theads = node.parentNode.parentNode.getElementsByTagName('thead');
			var tbodys = node.parentNode.parentNode.getElementsByTagName('tbody');

			for(var i = theads.length -1; i >= 0; i--){
				theads[i].style.display = theads[i].style.display == 'none' ? '' : 'none';
			}
			
			for(var i = tbodys.length -1; i >= 0; i--){
				tbodys[i].style.display = tbodys[i].style.display == 'none' ? '' : 'none';
				// query is close or open
				var displayType = document.getElementById("displayType");
				if(displayType && displayType != 'undifined') {
					if(tbodys[i].style.display == 'none')
						displayType.value = "";
					else 
						displayType.value = "1";
				}
			}
			return false;
		});
	});
}

function linkage(id1,id2,data){ //联动select
	if(data==''){return;}
	var s1 = $(id1);
	var s2 = $(id2);
	var index = 0;
	//var data = {sType:[{sName:'n1',index:1},{sName:'n2',index:2},{selected:true}],sType2:[{sName:'n1111',index:3},{sName:'n22222',index:4}],sType3:[{sName:'n13333',index:3},{sName:'n24444',index:4}]};
	s1.update();
	for(var i in data){
		var selected = data[i]['selected'];
		var text = decodeURI(i);
		var value = data[i].value;
		s1.options.add(new Option(text,value));
		if(selected) {
			s1.selectedIndex = index;
		}
		index ++;
	}
	s2.update();
	
	for(var v = s1.options[s1.selectedIndex].text, len = data[v]['children'].length, i = 0; i < len; i++){
		var selected = data[v]['children'][i].selected;
		var text = decodeURI(data[v]['children'][i].name);
		var value = data[v]['children'][i].value;
		s2.options.add(new Option(text,value));
		if(selected) {
			s2.selectedIndex = i;
		}
	}
	s1.observe('change',function(){
		s2.update();
		
		for(var v = s1.options[s1.selectedIndex].text, len = data[v]['children'].length, i = 0; i < len; i++){
			var selected = data[v]['children'][i].selected;
			var text = decodeURI(data[v]['children'][i].name);
			var value = data[v]['children'][i].value;
			s2.options.add(new Option(text,value));
			if(selected) {
				s2.selectedIndex = i;
			}
		} 
	});
}

function setCookie(name,value){
	var Days = 30; //此 cookie 将被保存 30 天
	var exp  = new Date();    //new Date("December 31, 9998");
	exp.setTime(exp.getTime() + Days*24*60*60*1000);
	document.cookie = name + "="+ escape(value) +";expires="+ exp.toGMTString()+";Path=/";
}

function getCookie(name){
	var arr = document.cookie.match(new RegExp("(^| )"+name+"=([^;]*)(;|$)"));
	if(arr != null) return unescape(arr[2]); return null;
}

function delCookie(name){
	var exp = new Date();
	exp.setTime(exp.getTime() - 1);
	var cval = getCookie(name);
	if(cval != null) document.cookie=name +"="+cval+";expires="+exp.toGMTString();
}

function regFilterBt(){
	if(!$('filterBt') || !$('colFilter')) return;
	$('filterBt').observe('click',function(){
		$('colFilter').style.left = $('filterBt').getPos().left + 30 +'px';
		$('colFilter').style.top = $('filterBt').getPos().bottom+'px',
		//$('colFilter').style.width = $('filterBt').getPos().right - $('filterBt').getPos().left - 20 + 'px';
		$('colFilter').toggle();
	});
	
	Event.observe(document.body,'click',function(event){ 
		var target = Event.element(event);
		if(!target.ancestors().pluck('id').without('').include('colFilter') && target.id != 'filterBt'){
			$('colFilter').hide();
		}
	});
}

function regDateInput(){
	    var wdate= $$("input.Wdate");
		if(wdate.length == 0) return;
		wdate.each(function(node){
		if(node.id == 'minDate' && $('maxDate')){
			node.observe('focus',function(){ WdatePicker({dateFmt:'yyyy-MM-dd HH:mm:ss',isShowWeek:false,maxDate:'#F{$dp.$D(\'maxDate\')}',lang:'<?=$_SESSION["datelan"]?>'}); });
		}else if(node.id == 'maxDate' && $('minDate')){
			node.observe('focus',function(){ WdatePicker({dateFmt:'yyyy-MM-dd HH:mm:ss',isShowWeek:false,minDate:'#F{$dp.$D(\'minDate\')}',lang:'<?=$_SESSION["datelan"]?>'}); });
		}else{
			node.observe('focus',function(){ WdatePicker({dateFmt:'yyyy-MM-dd HH:mm:ss',isShowWeek:false,lang:'<?=$_SESSION["datelan"]?>'}); });				
		}
	});
}
document.observe('dom:loaded',showTable);
document.observe('dom:loaded',regFilterBt);