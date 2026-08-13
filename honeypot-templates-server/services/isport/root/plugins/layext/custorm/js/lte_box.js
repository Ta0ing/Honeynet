(function ($){

	//全局系统对象
    window['LG'] = {};

    //显示成功提示窗口
    LG.showSuccess = function (message, successCallback){
    	layer.msg(message, {icon:1,time:0,btn:['确定']}, function(index, layero){
    		if(successCallback)
    			successCallback();
    	});
    };

    //显示成功提示窗口
    LG.confirm = function (message, yesCallback,noCallback){
    	layer.confirm(message, {
    		btn: ['是','否'] 
    	}, function(index, layero){
    		yesCallback();
    		layer.close(index);
    	}, function(){
    		if(noCallback)
    			noCallback();
    	});
    };

    //显示错误提示窗口
    LG.showError = function (message, errorCallback){
    	layer.msg(message, {icon:2,time:0,btn:['确定']},function(index, layero){
    		if(errorCallback)
    			errorCallback();
    	});
    };



    //显示错误提示窗口
    LG.showErrorConfirm = function (message, errorCallback,errorCallback1){
    	layer.msg(message, {icon:2,time:0,btn:['重试','取消']},function(index, layero){
    		errorCallback();
    	},function(){
    		errorCallback1();
    	});
    };

    //等待提示框
    LG.waiting=function(message){
    	var index=layer.msg(message, {icon:4,time:0,shade:0.3});
    	return {
    		"close":function(){
    			layer.close(index);
    		}
    	};
    };

    //打开遮罩
    LG.showMask=function(message){
    	layer.msg(message, {time:0,shade:0.3});
    };

    //隐藏遮罩
    LG.hiddenMask=function(message){
    	layer.closeAll('loading');
    };

    $.locationHelper={
    	"jumpTo":function(url){
    		url=$.locationHelper.wrapperURL(url);
    		window.location.href=url;
    	},
    	"wrapperURL":function(url){
    		return url;
    	},
    	"putVersionToJSON":function(json){
    		
    	}
    };

})(jQuery);