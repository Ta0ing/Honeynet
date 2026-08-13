//判断浏览器是否为ie6或7
var browser = navigator.appName;
if (browser == "Microsoft Internet Explorer") {
    var b_version = navigator.appVersion;
    var version = b_version.split(";");
    var trim_Version = version[1].replace(/[ ]/g, "");
    if (browser == "Microsoft Internet Explorer" && trim_Version == "MSIE7.0") {
        $(".loginBlock").remove();
        $("#lowie").show();
    } else if (browser == "Microsoft Internet Explorer" && trim_Version == "MSIE6.0") {
        $(".loginBlock").remove();
        $("#lowie").show();
        $(".lineDown").hide();
        $(".lineUp").hide();
        $(".top .logo").html("<img src='images/logoie6.png'><span> 火绒终端安全管理系统</span>");
        $(".downloadBlock .floatL").eq(0).html("<img src='images/login1ie6.png'>");
        $("#lowie").html("<img src='images/dfie6.png' class='verticalTop'>&nbsp;&nbsp;&nbsp;<span class='verticalMiddle'>您正在使用的浏览器版本过低，将不能正常浏览和使用控制中心</span>")

    } else if (browser == "Microsoft Internet Explorer" && trim_Version == "MSIE5.0") {
        $(".loginBlock").hide();
        $("#lowie").show();
    }
}
var portal = {
    init: function() {
        this.bundle();
    },
    bundle: function() {
        // 禁止输入空格
        _self.keyDownEvent({
            "obj": '.noSpaceInp'
        });
        // 绑定按钮事件
        $(document).on('click', '.eventBtn', function() {
            var event = $(this).attr('data-event');
            event && portal[event].call(this);
        });
        // 绑定重置密码校验
        $('.resetPwdPop').on('keyup', 'input[name="newPwd"]', function(event) {
            var e = event || window.event || arguments.callee.caller.arguments[0],
                reg = /^(?=.*[0-9])(?=.*[a-zA-Z])(?=.*[^a-zA-Z0-9]).{8,32}$/,
                $this = $(this),
                $pwdV = $('input[name="pwdVerify"]'),
                val = $this.val(),
                pwdV = $pwdV.val();
            if (e.keyCode == 9) {
                return;
            }
            if (e.keyCode == 13) {
                $('.resetPwdPop').find('a.eventBtn').click();
                return;
            }
            if (!!pwdV) {
                if (pwdV == val) {
                    $pwdV.siblings('font').hide();
                } else {
                    $pwdV.siblings('font').show();
                }
            }
            if (!reg.test(val)) {
                $this.siblings('font').show().html('密码必须由8-32位大小写字母、数字、特殊字符组成');
                return false;
            }
            $this.siblings('font').hide();
        }).on('blur', 'input[name="newPwd"]', function(event) {
            var val = $(this).val();
            if (!val) {
                $(this).siblings('font').show().html('请输入新密码');
            }
        }).on('keyup', 'input[name="pwdVerify"]', function(event) {
            var e = event || window.event || arguments.callee.caller.arguments[0],
                $this = $(this),
                $val = $('input[name="newPwd"]'),
                val = $val.val(),
                pwdV = $this.val();
            if (e.keyCode == 9) {
                return;
            }
            if (e.keyCode == 13) {
                $('.resetPwdPop').find('a.eventBtn').click();
                return;
            }
            if (pwdV == val) {
                $this.siblings('font').hide();
            } else {
                $this.siblings('font').show();
            }
        }).on('blur', 'input[name="pwdVerify"]', function(event) {
            var val = $(this).val();
            if (!val) {
                $(this).siblings('font').show();
            }
        });
    },
    // 确定重置初始密码
    resetPwd: function() {
        var $resetPwdPop = $('.resetPwdPop'),
            $newP = $resetPwdPop.find("input[name=newPwd]"),
            $VerifyP = $resetPwdPop.find("input[name=pwdVerify]"),
            $tips = $resetPwdPop.find('.tips'),
            newP = $newP.val(),
            VerifyP = $VerifyP.val();
        //校验所有input是否输入正确
        for (var i = 0; i < $tips.length; i++) {
            if ($tips.eq(i).css('display') != 'none') {
                return;
            }
        }
        if (newP == '') {
            $newP.siblings('font').show();
            return;
        }
        if (newP == VerifyP) {
            $VerifyP.siblings('font').hide();
        } else {
            $VerifyP.siblings('font').show();
            return;
        }
        var dataa = {
            "newpwd": hex_sha1(newP)
        };
        _self.request({
            url: '/auth/_reset_admin',
            data: dataa,
        }, function() {
            doLogin(newP);
        });
    }
};
// 初始化
portal.init();
//聚焦密码输入
$("input[name=username]").focus();
$("input[name=username]").next().hide();

function doLogin(pwd) {
    var username = $("input[name=username]").val();
    var password = $('input[name=password]').val();
    var remeber = $("input[name=remeber]").is(":checked");
    if (username == '' && password == '') {
        $(".errorTips").show();
        $('.errorTips font').html(' 账户名和密码不能为空');
        return false;
    } else if (username == '') {
        $(".errorTips").show();
        $('.errorTips font').html(' 账户名不能为空');
        return false;
    } else if (password == '') {
        $(".errorTips").show();
        $('.errorTips font').html(' 密码不能为空');
        return false;
    }
    $.ajax({
        url: '/login',
        data: {
            "username": username,
            // "password": hex_sha1(pwd || password),
            "password": pwd || password,
        },
        type: 'POST',
        dataType: 'json',
        headers: {
            "HTTP-CSRF-TOKEN": getCookie('HRESSCSRF')
        },
        success: function(data) {
            alert("用户名或密码错误");
        //     if (data.errno == 0) {
        //         $(".errorTips").hide();
        //         // 判断是否需要动态验证
        //         if (data.data.totp) {
        //             _self.openTotpPop();
        //         } else {
        //             window.location.href = "/";
        //         }
        //
        //     } else if (data.errno == -1) {
        //         if (data.errmsg == 'Account Is Disabled.') {
        //             $(".errorTips").show();
        //             $('.errorTips font').html(' 此管理员账户已被停用');
        //         } else {
        //             $(".errorTips").show();
        //             $('.errorTips font').html(' 账户名或密码错误');
        //         }
        //
        //     } else if (data.errno == -2) {
        //         $(".aLockHintPop").show();
        //         $(".shade").show();
        //     } else if (data.errno == -4) {
        //         // 管理员默认密码需要重置(非本机登录)
        //         _self.showPop($(".notResetPwdPop"));
        //     } else if (data.errno == -3) {
        //         // 管理员默认密码需要重置(本机登录)
        //         _self.showPop($(".resetPwdPop"));
        //         $('.resetPwdPop').find('input').val('').end().find('.tips').hide().end().find('input[name=newPwd]').focus();
        //     }
        }
    });
}
//按回车登陆
document.onkeydown = function(e) {
    if (!e) e = window.event;

    if ((e.keyCode || e.which) == 13) {
        if ($('.totpPop').css('display') != 'block' && $('.resetPwdPop').css('display') != 'block') {
            doLogin();
        } else {
            var $popElem = $('.totpPop'),
                code = $popElem.find('input[name=code]').val(),
                pageName = $popElem.attr('data-page');
            if (code.length != 6) {
                $popElem.find('.errorTips').show();
                $popElem.find('.errorTips').html(' 动态口令为6位数');
            } else {
                _self.authTotp(code, pageName);
            }
        }
    }
}

$(".loginButton").click(function() {
    doLogin();
});

$(".loginButton").keydown(function(e) {
    if (!e) e = window.event;

    if ((e.keyCode || e.which) == 13) {
        if ($('.totpPop').css('display') != 'block') {
            doLogin();
        } else {

        }
    }
});



function forgetPWPop() {
    $(".forgetPWPop").show();
    $(".shade").show();
}

$(".closeW,.hideButton").click(function() {
    $(this).parents(".pop").hide();
    $(".shade").hide();
})

function sureFPWButton(a) {
    $(a).parents(".pop").hide();
    $(".shade").hide();
}
// 密码框获得焦点改变其属性
// $("input[name=password]").focus(function(){
//     $(this).attr("type","password");
//     $(this).next().hide();
// })
// 用户名和密码输入框聚焦时隐藏后面问号图标
$("input[name=username]").focus(function() {
    $(this).next().hide();
})
$("input[name=username]").blur(function() {
    $(this).next().show();
})
$("input[name=password]").blur(function() {
    $(this).next().show();
})
//placeholder属性兼容
$(".placeholderInput").focus(function() {
    $(this).siblings(".placeholder").hide();

});
$(".placeholder").click(function() {
    $(this).hide();
    $(this).siblings(".placeholderInput").focus();
});
$(".placeholderInput").blur(function() {
    var value = $(this).val();
    if (value == "") {
        $(this).siblings(".placeholder").show();
    }
});

//
$(".anwserIcon").mouseenter(function() {
    $(this).find("span").show();
})
$(".anwserIcon").mouseleave(function() {
    $(this).find("span").hide();
})

//文本框获得焦点时边框颜色

$("body").on("focus", "input[type=text],input[type=password]", function() {
    $(this).css({
        border: "1px solid #b3b3b3",
        transition: 0.5
    });
})
$("body").on("blur", "input[type=text],input[type=password]", function() {
    $(this).css({
        border: "1px solid #e5e6e7",
        transition: 0.5
    })
})

function getCookie(cookie_name) {
    var allcookies = document.cookie;
    var cookie_pos = allcookies.indexOf(cookie_name); //索引的长度

    // 如果找到了索引，就代表cookie存在，
    // 反之，就说明不存在。
    if (cookie_pos != -1) {
        // 把cookie_pos放在值的开始，只要给值加1即可。
        cookie_pos += cookie_name.length + 1; //这里容易出问题，所以请大家参考的时候自己好好研究一下
        var cookie_end = allcookies.indexOf(";", cookie_pos);

        if (cookie_end == -1) {
            cookie_end = allcookies.length;
        }

        var value = unescape(allcookies.substring(cookie_pos, cookie_end)); //这里就可以得到你想要的cookie的值了。。。
    }
    return value;
}