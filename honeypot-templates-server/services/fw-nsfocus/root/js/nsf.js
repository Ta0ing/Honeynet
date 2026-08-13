//给jQuery增加方法：$.getUrlParam()
//用于获取url中的参数
(function($) {
    $.getUrlParam = function(name) {
        var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
        var r = window.location.search.substr(1).match(reg);
        if (r!=null) return unescape(r[2]); return null;
    }
})(jQuery);

(function($) {
    $.getDateDiff = function(datetime_str) {
        var now = new Date().getTime();
        var diff_value = now - Date.parse((datetime_str).replace(/-/gi,"/"));
        var result = "";
        var suffix = "前";
        if(diff_value < 0) {
            diff_value = 0-diff_value;
            suffix = "以后";
        }
        var yearC = diff_value/(1000 * 60 * 60 * 24 * 365);
        var monthC = diff_value/(1000 * 60 * 60 * 24 * 30);
        var weekC = diff_value/(1000 * 60 * 60 * 24 * 7);
        var dayC = diff_value/(1000 * 60 * 60 * 24);
        var hourC = diff_value/(1000 * 60 * 60);
        var minC = diff_value/(1000 * 60);
        if(yearC>=1) {
            result = parseInt(yearC) + "年";
            var months = parseInt(monthC%12);
            if(months>0) {
                result += "，" + months + "个月";
            }
        } else if(monthC>=1){
            result = parseInt(monthC) + "个月";
            var weeks = parseInt(weekC%4);
            if(weeks>0) {
                result += "，" + weeks + "周";
            }
        } else if(weekC>=1){
            result = parseInt(weekC) + "周";
            var days = parseInt(days%7);
            if(days>0) {
                result += "，" + days + "天";
            }
        } else if(dayC>=1) {
            result =  parseInt(dayC) + "天";
            var hours = parseInt(hourC%24);
            if(hours>0) {
                result += "，" + hours + "小时";
            }
        } else if(hourC>1) {
            result += parseInt(hourC%24) + "小时";
            var minutes = parseInt(minC%60);
            if(minutes>0) {
                result += "，" + minutes + "分钟";
            }
        } else if(minC>1) {
            result += parseInt(minC%60) + "分钟";
        }
        if(result) {
            result = result + suffix;
        } else {
            result = "最近1分钟";
        }
        return result;
    }
})(jQuery);

(function($) {
    $.getBaseUrl = function(url) {
        if(typeof url == "undefined" || null == url) {
            url = window.location.href;
        }
        var regex = /^(https?\:\/\/[^\/]*).*/;
        var match = url.match(regex);
        if(typeof match != "undefined" && null != match) {
            return match[1];
        }
        return '';
    }
})(jQuery);

//显示表单错误，如果传入了div_id，则会把对应的表单项标红
var show_input_error = function(errormsg, div_id) {
    $("#successMsg").addClass("hide");
    if(div_id != undefined) {
        $(div_id).addClass("error");
    }
    if(errormsg == undefined) {
        errormsg = "处理请求失败，请稍后再试";
    }
    $("#errorMsg span").html(errormsg);
    $("#errorMsg").removeClass("hide");
}

//显示表单执行成功
var show_success = function(msg) {
    $("#errorMsg").addClass("hide");
    $("#successMsg span").html(msg);
    $("#successMsg").removeClass("hide");
}

//获取表单中的csrftoken
var get_csrf_token = function() {
    return $.trim($("input[name='csrf_token']").first().val());
}

//获取响应数据
var get_response_data = function(response_str) {
    try {
        var data = eval("("+response_str+")");
        if(data["status"]==0) {
            return false;
        } else {
            return data["data"];
        }
    } catch(e) {
        return false;
    }
}

//显示验证码
var show_captcha = function() {
    $divCaptcha = $("#divCaptcha");
    if($divCaptcha.is(":hidden")) {
        $divCaptcha.removeClass("hide");
        //给图片和换一张的点击事件绑定刷新验证码响应。
        $("#divCaptcha img").unbind("click").bind("click", refresh_captcha);
        $("#divCaptcha a").unbind("click").bind("click", refresh_captcha);
        refresh_captcha();
    }
}

//刷新验证码
var refresh_captcha = function() {
    $("#divCaptcha img").attr("src", "/auth/captcha?r="+Math.random());
    $("#inputCaptcha").val("");
    $("#divCaptcha a").blur();
    return false;
}

//认证表单检查，成功后调用do_auth
var check_login_form = function() {
    var domain_id = $.trim($("#inputDomainId").val());
    var params = {domain_id:domain_id};
    var username = $.trim($("#inputUsername").val());
    if(username == "") {
        show_input_error("用户名不能为空", "#divUsername");
        return false;
    } else {
        $("#divUsername").removeClass("error");
        params["username"] = username;
    }
    var password = $("#inputPassword").val();
    if(password == "") {
        show_input_error("密码不能为空", "#divPassword");
        return false;
    } else {
        $("#divPassword").removeClass("error");
        params["password"] = password;
    }
    if(!$("#divCaptcha").is(":hidden")) {
        var captcha = $.trim($("#inputCaptcha").val());
        if(captcha == "") {
            show_input_error("验证码不能为空", "#divCaptcha");
            return false;
        } else {
            $("#divCaptcha").removeClass("error");
            params["captcha"] = captcha;
        }
    }
    $("#errorMsg").addClass("hide");
    do_auth(params);
    return false;
}

//恢复提交按钮，延时1000秒防止重复提交（仅前端防止）
var reset_submit_button = function() {
    setTimeout(function(){$("#btnSubmit").button("reset")}, 1000);
}

//执行认证
var do_auth = function(params) {
    $("#btnSubmit").button("loading");
    $.ajax({
        type: "POST",
        url: "/login",
        data: params,
        success: function(res){
            alert("用户名或密码错误")
        }
    });
}

//修改密码表单检查，成功后调用do_modifypwd
var check_modifypwd_form = function() {
    var domain_id = $.trim($("#inputDomainId").val());
    var params = {domain_id:domain_id};
    var username = $.trim($("#inputUsername").val());
    if(username == "") {
        show_input_error("用户名不能为空", "#divUsername");
        return false;
    } else {
        params["username"] = username;
        $("#divUsername").removeClass("error");
    }
    var password = $("#inputPassword").val();
    if(password == "") {
        show_input_error("密码不能为空", "#divPassword");
        return false;
    } else {
        params["password_old"] = password;
        $("#divPassword").removeClass("error");
    }
    var password1 = $("#inputPassword1").val();
    if(password1 == "") {
        show_input_error("新密码不能为空", "#divPassword1");
        return false;
    } else if(password!="" && password1 == password){
        show_input_error("新密码不能与旧密码相同", "#divPassword1");
        return false;
    } else {
        params["password_new"] = password1;
        $("#divPassword1").removeClass("error");
    }
    var password2 = $("#inputPassword2").val();
    if(password2 == "") {
        show_input_error("确认密码不能为空", "#divPassword2");
        return false;
    } else if(password2 != password1) {
        show_input_error("两次密码输入不一致", "#divPassword2");
        return false;
    } else {
        $("#divPassword2").removeClass("error");
    }
    do_modifypwd(params);
    return false;
}

//执行密码修改
var do_modifypwd = function(params, show_auth_dialog) {
    $("#errorMsg").addClass("hide");
    $("#btnSubmit").button("loading");
    $.ajax({
        type: "POST",
        url: "/auth/modifypwd",
        data:params,
        success: function(res){
            var data = get_response_data(res);
            if(!data) {
                refresh_captcha();
                show_input_error();
                reset_submit_button();
            } else {
                if(data["result"]==true) {
                    show_success(data["msg"]);
                    setTimeout("$('#myModal').modal('hide')", 500);
                    //如果密码栏为空，表示是第一次登录修改密码。
                    if($("#divPassword").is(":hidden")) {
                        //修改密码后重新认证
                        setTimeout(function(){show_login(params["domain_id"], params["username"])}, 1000);
                    }
                } else {
                    show_input_error(data["msg"]);
                    reset_submit_button();
                }
            }
        }
    });
}

//显示对话框基础方法，option为键值对象
var _show_dialog = function(option) {
    //先关闭该对话框
    $("#myModal").removeClass("fade");
    $("#myModal").modal("hide");
    $("#myModal").addClass("fade");
    option.has_footer = option.has_footer!=undefined?option.has_footer:true;
    option.cancel = option.cancel!=undefined?option.cancel:true;
    option.ok = option.ok!=undefined?option.ok:true;
    option.title = option.title!=undefined?option.title:"提示";
    option.height = option.height!=undefined?option.height:"";
    option.width = option.width!=undefined?option.width:"auto",
    
    $("#myModalContent").html('<img src="/img/loading.gif"/> 载入中...');
    //$("#myModalContent").css("height", option.height);
    if(option.url != null) {
        $("#myModalContent").load(option.url,"",function(){
            if($.bootstrapIE6!=undefined){
                setTimeout(function(){$.bootstrapIE6($("#myModalContent"))}, 1);
            }
            if(typeof option.onload_callback=="function") {
                setTimeout(option.onload_callback, 50);
            }
        });
    } else {
        $("#myModalContent").html(option.msg);
    }
    if(option.has_footer===true) {
        $("#myModalFooter").removeClass("hide");
        if(option.cancel===true) {
            $("#myModalCancel").removeClass("hide");
        } else {
            $("#myModalCancel").addClass("hide");
        }
        if(option.ok===true) {
            $("#myModalOk").removeClass("hide");
        } else {
            $("#myModalOk").addClass("hide");
        }
    } else {
        $("#myModalFooter").addClass("hide");
    }
    $("#myModalOk").unbind("click");
    if(option.ok_callback) {
        $("#myModalOk").one("click", option.ok_callback);
    }
    $("#myModalTitle").html(option.title);
    $("#myModal").modal("toggle");
    if(option.width == "auto") {
        $("#myModal").modal().css({
            "width": "auto",
            "margin-left": function () {return -($(this).width()/2);}
        });
    } else {
        $("#myModal").modal().css({
            "width": option.width+"px",
            "margin-left": function () {return -(option.width/2);}
        });
    }
}

//从服务器获取csrftoken
var get_csrf_token_from_server = function() {
    var token='';
    $.ajax({
            type: "GET",
            url: "/csrftoken",
            async: false, //这里是同步请求
            success: function(data) {token = data}
            });
    return token;
}

//注销，需要csrftoken
var logout = function(domain_id) {
    _show_dialog({"msg":"确定要断开连接？", "width":400, "ok_callback":function(e){
        var params = {domain_id:domain_id, csrf_token:get_csrf_token_from_server()};
        $.ajax({
            type: "POST",
            url: "/auth/logout",
            data: params,
            success: function(res) {
                var data = get_response_data(res);
                if(!data) {
                   //do something.
                } else {
                    refresh_auth_status();
                }
            }});
    }});
}

//显示关于对话框
var show_about = function() {
    _show_dialog({"url":"/about", "title":"最终用户许可协议", "width":600, "cancel":false});
}

//显示登录对话框
var show_login = function(domain_id, username) {
    domain_id = domain_id?domain_id:0;
    _show_dialog({"url":"/login?domain_id="+domain_id, "title":"认证", "width":510, "has_footer":false, 
                "onload_callback":function(){
                    $.get("/auth/needcaptcha", function(res){
                        var data = get_response_data(res);
                        //如果传递了username参数，给表单设置上
                        if(username!=undefined) {
                            $("#inputUsername").val(username);
                        }
                        if(data != false) {
                            if(data["result"] == true) {
                                show_captcha();
                            }
                        }});
                    }
                });
}

//从服务器获取可认证的域
var get_auth_domains = function() {
    $("#select_domains").html('<li><img src="/img/loading.gif" /> 载入中...</li>');
    setTimeout(function(){
        $.ajax({
            type: "GET",
            url: "/auth/domain",
            success: function(res) {
                var data = get_response_data(res);
                if(!data) {
                   //do something.
                } else {
                    var list_html = '';
                    var domains = data.domains;
                    for(var i in domains) {
                        list_html += '<li><a domain_id="'+domains[i].id+'" onclick="change_domain(this)" href="#">'+domains[i].name+'</a></li>';
                    }
                    $("#select_domains").html(list_html);
                    
                }
            }});
        }, 200);
}

//显示修改密码对话框
var show_modify_password = function(domain_id, username) {
    domain_id = domain_id?domain_id:0;
    _show_dialog({"url":"/modifypwd?domain_id="+domain_id+"&username="+encodeURIComponent(username), "title":"修改密码", "width":510, "has_footer":false});
}

//显示第一次登录时修改密码对话框，载入后隐藏了原密码。
var show_modify_password_first_login = function(domain_id, username, password_old) {
    domain_id = domain_id?domain_id:0;
    _show_dialog({"url":"/modifypwd?domain_id="+domain_id+"&username="+encodeURIComponent(username), "title":"修改默认密码", "width":510, "has_footer":false,
                "onload_callback":function() {
                    $("#divPassword").addClass("hide");
                    $("#inputPassword").val(password_old);
                    }
                });
}

//选择其他域进行认证
var change_domain = function(dom) {
    var domain_id = $(dom).attr("domain_id");
    $("#auth").attr("domain_id", $(dom).attr("domain_id"));
    $("#auth > span").html($(dom).html());
    show_login(domain_id);
}

//显示重定向提示框，10秒倒计时
var show_redirect = function() {
    if($.cookie('auth_ok') && !redirect_canceled && redirect_url!="" && $("#redirect_alert").length>0) {
        $("#redirect_alert").removeClass("hide");
        var inverse = 10;
        $("#redirect_alert a[class='alert-link']").html(redirect_url);
        $("#redirect_alert span").html(inverse);
        redirect_interval = setInterval(function(){
            $("#redirect_alert span").html(--inverse);
            if(inverse <= 0) {
                do_redirect();
            }
        }, 1000);
        //跳转页面显示完毕后，尝试清除cookie
        $.cookie('redirect_url', '', {expires: -1});
        $.cookie('auth_ok', '', {expires: -1});
    }
}

//执行重定向，如果有心跳检查，在新窗口打开，否则在当前顶层窗口
var do_redirect = function() {
    cancel_redirect();
    if(heartbeat) {
        window.open(redirect_url, "_blank");
    } else {
        window.open(redirect_url, "_top");
    }
}


//取消重定向跳转
var cancel_redirect = function(button) {
    clearInterval(redirect_interval);
    redirect_canceled = true;
    if(button) {
        $(button).addClass("disabled");
        $(button).html("已取消");
    }
}


//开启心跳检查，在获取状态后调用
var show_heartbeat = function() {
    heartbeat = true;
    $("#heartbeat_alert").removeClass("hide");
    onbeforeunload = function(){
        return "如果关闭了该页面，可能会断开连接。";
    };
}


//取消心跳检查，在获取状态后调用
var cancel_heartbeat = function() {
    heartbeat = false;
    $("#heartbeat_alert").addClass("hide");
    window.onbeforeunload = null;
}


//从服务器获取数据，刷新认证状态。
//同时会判断是否有心跳检查。
var refresh_auth_status = function() {
    $.ajax({
        type: "GET",
        url: "/status/auth",
        success: function(res) {
            var data = get_response_data(res);
            if(!data) {
               //do something.
            } else {
                var list_html = '';
                var login_data_new = [];
                if(data["need_heartbeat"] == true) {
                    show_heartbeat();
                } else {
                    cancel_heartbeat();
                }
                $("#data").html("");
                //遍历认证状态，构造html
                for(var i in data.users) {
                    var user = data.users[i];
                    var status = user.login;
                    var domain_type = user.domain_type;
                    //如果不能操作，显示-
                    var controls = '-';
                    if(status==0) {
                        controls = '<button class="btn btn-primary btn-mini" onclick="show_login('+user.domain_id+',\''+user.user+'\')">重新认证</button></td>';
                    } else if(status==1) {
                        controls = '<button class="btn btn-primary btn-mini" onclick="logout('+user.domain_id+')">断开连接</button>';
                        if(domain_type==0) {
                            //多一个空格防止某些浏览器下两个按钮间没有间隙
                            controls += ' <button class="btn btn-primary btn-mini" onclick="show_modify_password('+user.domain_id+',\''+user.user+'\')">修改密码</button>';
                        }
                        controls += '</td>';
                    }
                    var user_with_domain = user.user+"<em>@"+user.domain_name+"</em>";
                    login_data_new.push(user_with_domain);
                    list_html += "<tr>"
                    +"<td>"+login_status[status]+"</td>"
                    +"<td>"+user_with_domain+"</td>"
                    +"<td><span data-toggle=\"tooltip\" title=\""+user.time+"\">"+$.getDateDiff(user.time)+"</span></td>"
                    +"<td>"+controls+"</td></tr>";
                }
                //没有认证数据
                if(login_data_new.length == 0) {
                    $("#nodata").removeClass("hide");
                } else {
                    $("#nodata").addClass("hide");
                    //登录状态展示到页面上
                    $("#data").html(list_html);
                    if($.bootstrapIE6!=undefined){
                        setTimeout(function(){$.bootstrapIE6($("#myModalContent"))}, 1);
                    }
                }
                //仅在[当前认证状态比原来少]的时候判断用户在其他地方登录的情况
                if(login_data_new.length < login_data_old.length) {
                    //判断是否丢失登录状态，以此来判断是否有用户在其他地方登录
                    var user_lost = [];
                    for(var i in login_data_old) {
                        //这里用jQuery的inArray方法，直接用Array.indexOf在IE下会有兼容问题
                        if($.inArray(login_data_old[i], login_data_new) == -1) {
                            user_lost.push(login_data_old[i]);
                        }
                    }
                    if(user_lost.length>0) {
                        //显示有用户在其他地方登录的提示消息
                        _show_dialog({"msg":"用户 <strong>"+user_lost.join('、')+"</strong> 已在其他地方登录。", "width":400});
                    }
                }
                login_data_old = login_data_new;
                if($.bootstrapIE6==undefined) {
                    $('#data span').tooltip({placement:'right'});
                    $('#data img').tooltip({placement:'right'});
                }
            }
        }
    });
}
//三种认证状态对应的图标。
var login_status = {"0":'<img data-toggle="tooltip" src="/img/login_off.gif" title="未连接"/>',"1":'<img data-toggle="tooltip" src="/img/login_on.gif" title="已连接"/>',"2":'<img data-toggle="tooltip" src="/img/headoff.gif" title="禁止登录"/>'};
//保存上一次刷新后的认证信息，用于判断是否有认证信息丢失。
var login_data_old = [];
//参数中的url，用于认证后跳转。
var redirect_url = null;
//是否取消了URL跳转
var redirect_canceled = false;
//是否启用心跳检查，默认为false。
var heartbeat = false;
//重定向interval
var redirect_interval = null;
//刷新认证状态的interval
var refresh_interval = null;
$(document).ready(function() {
    //获取参数中的域ID
    var domain_id = $.trim($.getUrlParam("domain_id"));
    //如果有跳转URL，获取保存。
    redirect_url = $.trim($.getUrlParam("url"));
    if(redirect_url) {
        var url_reg = /^(http:\/\/).*|(https:\/\/).*$/;
        if(!url_reg.test(redirect_url)) {
            redirect_url = "http://" + redirect_url;
        }
        redirect_url = encodeURI(decodeURI(redirect_url));
        //如果URL中有参数，清除认证成功的COOKIE
        $.cookie('auth_ok', '', {expires: -1});
        //将参数中的url保存到cookie
        $.cookie('redirect_url', redirect_url);
    } else {
        //如果参数中没有url，尝试从cookie去获取
        url = $.cookie('redirect_url');
        if(url!=null && url!='') {
            redirect_url = url
        }
    }
    show_redirect();

    //如果URL中有domain_id，弹出认证对话框
    if(domain_id!="") {
        $("#choose_domains").attr('disabled',"true");
        $("#domain_span").removeAttr('title');
        show_login(domain_id*1);
    }
    $("#domain_span").tooltip({placement:'right'});
    //警告框关闭按钮
    $("#redirect_alert").bind("closed.bs.alert", cancel_redirect);
    $("#location").html($.getBaseUrl());
    // setTimeout(function(){
    //     //1秒后开始刷新认证状态
    //     refresh_auth_status();
    //     //之后每隔10秒刷新认证状态
    //     refresh_interval = setInterval(refresh_auth_status, 10000);
    // }, 1000);
});