var layerObj = {
    number: 0,
    //公用弹框
    // ID 容器id width height shade offset 弹框宽，高，背景透明度，弹框位置 btn 按钮名称（多个用数组方式显示） yes 确定时回调函数 success 弹框成功回调函数
    layuiLayer: function (ID, width, height, shade, offset, btn, yes, success) {
        layerObj.number++;
        layui.use(['element', 'form', 'table', 'laypage', 'layer'], function () {
            layer.open({
                type: 1
                , title: false
                , offset: offset //具体配置参考：http://www.layui.com/doc/modules/layer.html#offset
                , id: 'layerDemo' + layerObj.number //防止重复弹出
                , content: $("#" + ID)
                , btn: btn
                , area: [width, height]
                , shade: shade //不显示遮罩
                , yes: function () {
                    if (yes) {
                        yes();
                    }
                }
                , success: function () {
                    if (success) {
                        success();
                    }
                }
            });
        });
    },
    // 列表
    // col 设置表头，elId 元素id，url 请求地址 tableOn监听事件元素，done 请求成功后执行方法，event 事件调用方法
    tableList: function (col, elId, url, tableOn, done, event,tableID,page) {
        layui.use(['element', 'form', 'table', 'laypage', 'layer'], function () {
            var element = layui.element; //导航的hover效果、二级菜单等功能，需要依赖element模块
            var form = layui.form;
            var table = layui.table;
            var laypage = layui.laypage
                , layer = layui.layer;
            if(tableID==""||tableID==null||tableID==undefined){
            	tableID = "thisTable"
            };
            if(page==""||page==null||page==undefined){
                page = true;
            }else{
                page = false;
            }
            table.render({
                elem: '#' + elId
                , url: url
                , cellMinWidth: 80 //全局定义常规单元格的最小宽度，layui 2.2.1 新增
                , page: page
                , loading: true
                , id:tableID
                , cols: col
                , height: 490
                , done: function (res, curr, count) {
                    if (done) {
                        done(res, curr, count)
                    }
                    //没有数据表头可滚动
                    count || this.elem.next('.layui-table-view').find('.layui-table-header').css('overflow', 'auto');
                }
            });
            table.on(tableOn, function (obj) {
                if (event) (
                    event(obj)
                )
            });
            form.render();
        });
    },
    /**
     * 确认提示框
     * tip 提示内容  btn 按钮数组 title 确认框标题  callback 回调函数
    */
    confirm:function(tip,btn,title,callback){
    	layui.use(['layer'], function () {
	        layer.confirm(tip,{btn: btn,title:title,icon: 3}, function(){
	            if(callback){
	                callback();
	            }
	        });
    	});
    },
    /**
     * 成功信息提示
     * msg 信息内容
    */
    successMsg:function(msg){
        layer.msg(msg, {time: 2000,icon: 1});
    },
    /**
     * 失败信息提示
     * msg 信息内容
     */
    failMsg:function(msg){
        layer.msg(msg, {time: 2000,icon: 2});
    },
    // 日期插件
    dateTimeInput: function (id, range, val, type, done) {
        layui.use("laydate", function () {
            var laydate = layui.laydate;
            //日期
            laydate.render({
                elem: "#" + id,
                range: range,
                value: val,
                type: type
                , done: function (value, date, endDate) {
                    if (done) {
                        done(value, date, endDate);
                        console.log(value); //得到日期生成的值，如：2017-08-18
                        console.log(date); //得到日期时间对象：{year: 2017, month: 8, date: 18, hours: 0, minutes: 0, seconds: 0}
                        console.log(endDate); //得结束的日期时间对象，开启范围选择（range: true）才会返回。对象成员同上。
                    }
                }
            });

        });
    },
    //渲染表单样式
    formRender: function () {
        layui.use(['form', 'layedit', 'laydate','element'], function () {
            var form = layui.form
                , layedit = layui.layedit
                , laydate = layui.laydate
                ,element = layui.element
                ,tree = layui.tree
                ,util = layui.util;
            //日期时间范围
            laydate.render({
                elem: '#test10'
                , type: 'datetime'
                , range: true
            });
            layerObj.dateTimeInput("validity", true, '2019-07-30 00:00:00 - 2019-09-19 00:00:00', 'datetime');
            form.render();
        });
    },
    getURL: function() {
        // http://localhost:8080/manager/view/data_strain.html
        var currWWWPath = window.document.location.href;
        // 获取主机地址之后的目录，如：/manager/view/data_strain.html
        var pathName = window.document.location.pathname;
        // 获取 ‘/’ 出现的位置
        var index = currWWWPath.indexOf(pathName);
        // 获取 http://localhost:8080
        var localhostPath = currWWWPath.substring(0, index);
        // 获取项目名
        var projectName = pathName.substring(0, pathName.substr(1).indexOf('/') + 1);
        return localhostPath + projectName;
    },
    // 加密
    encrypt:function(word){
    	if(!isNullOrBlank(word)){
    		var key = CryptoJS.enc.Utf8.parse("BJyqXT!2020@0101");
    		var srcs = CryptoJS.enc.Utf8.parse(word);
    		var encrypted = CryptoJS.AES.encrypt(srcs, key, {mode:CryptoJS.mode.ECB,padding: CryptoJS.pad.Pkcs7});
    		return encrypted.toString();
    	}else{
    		return "";
    	}
    },
    // 解密
    decrypt:function(word){
    	if(!isNullOrBlank(word)){
    		var key = CryptoJS.enc.Utf8.parse("BJyqXT!2020@0101");
    		var decrypt = CryptoJS.AES.decrypt(word, key, {mode:CryptoJS.mode.ECB,padding: CryptoJS.pad.Pkcs7});
    		return CryptoJS.enc.Utf8.stringify(decrypt).toString();
    	}else{
    		return "";
    	}
    },
    selectChanage:function(id,childId,callback){
        $.ajax({
            url: ctx + "/service/dataDictionary/findByParentId",
            type: "get",
            async: false,
            data:{
                "parentId": id,
            },
            success:function (data) {
                var str="";
                $.each(data,function(index,item){
                    str+='<option value="'+$(item).id+'">'+$(item).name+'</option>';
                });
                $.ajax({
                    url: ctx + "/service/dataDictionary/findByParentId",
                    type: "get",
                    async: false,
                    data:{
                        "parentId": childId,
                    },
                    success:function (data) {
                        var strChild="";
                        $.each(data,function(index,item){
                            strChild+='<option value="'+$(item).id+'">'+$(item).name+'</option>';
                        });
                        return {street:str,area:strChild};
                    },error:function () {
                        console.log("请求失败")
                    }
                })

            },error:function () {
                console.log("请求失败")
            }
        })

    },
    //时间戳返回日期
    returnDate:function(timestamp){
        if( timestamp ){
            var date = new Date(timestamp );//时间戳为10位需*1000，时间戳为13位的话不需乘1000
            var Y = date.getFullYear() + '-';
            var M = (date.getMonth()+1 < 10 ? '0'+(date.getMonth()+1) : date.getMonth()+1) + '-';
            var D = (date.getDate() < 10 ? '0'+date.getDate() : date.getDate());
            return Y+M+D;
        }else{
            return ''
        }
    }
};


