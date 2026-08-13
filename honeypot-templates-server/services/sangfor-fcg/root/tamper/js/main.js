// JavaScript Document
$(document).ready(function() {
    //表单样式
    var input_button = $(".input_button dd span, .input_button dd input, .input_button_max dd input");
    var input_text = $(".input_text dd span, .input_text dd span input");
    $(input_text).hover(function() {
        $(this).addClass("input_text_h");
    }, function() {
        $(this).removeClass("input_text_h");
    });
    $(input_text).focus(function() {
        $(this).addClass("input_text_b");
    });
    $(input_text).blur(function() {
        $(this).removeClass("input_text_b");
        /*if (this.value === '') {
        	if (this.name === 'user') {
        		$('.tip_error')[0].innerHTML = '请输入你的用户名';
        		$('#tip_err').show();
        	} else if (this.name == 'pwd') {
        		$('.tip_error')[0].innerHTML = '请输入你的密码';
        		$('#tip_err').show();
        	}
        } else {
        	if (this.name === 'user') {
        		$('#tip_err').hide();
        	} else if (this.name === 'pwd') {
        		$('#tip_err').hide();
        	}
        }*/
    });
    $(input_text).keypress(function() {
        if (this.name === 'user') {
            $('#tip_err').hide();
        } else if (this.name === 'pwd') {
            $('#tip_err').hide();
        }
    });
    $(input_button).hover(function() {
        $(this).addClass("buttons_h");
    }, function() {
        $(this).removeClass("buttons_h");
    });
    $(input_button).focus(function() {
        $(this).addClass("buttons_h");
    });
    $(input_button).blur(function() {
        $(this).removeClass("buttons_h");
    });

    //window
    $(window).resize(function() {
        FrameworkHeight();
        table();
    });
    FrameworkHeight();
    table();
    //左树...
});

function table() {
    // 悬停
    $(".table dl").hover(function() {
        $(this).addClass("dl_h");
    }, function() {
        $(this).removeClass("dl_h");
    });
    //设置表格最后一行样式
    $(".table dl:last-child").css({
        "border-bottom": "1px solid #DDE6EC !important"
    });
};

function FrameworkHeight() {
    //获取窗高度口宽度用到BODY
    if (window.innerWidth)
        winWidth = window.innerWidth;
    else if ((document.body) && (document.body.clientWidth))
        winWidth = document.body.clientWidth;
    //获取窗口高度
    if (window.innerHeight)
        winHeight = window.innerHeight;
    else if ((document.body) && (document.body.clientHeight))
        winHeight = document.body.clientHeight;
    //通过深入Document内部对body进行检测，获取窗口大小
    if (document.documentElement && document.documentElement.clientHeight && document.documentElement.clientWidth) {
        winHeight = document.documentElement.clientHeight;
        winWidth = document.documentElement.clientWidth;
    }
    //获取Portal框架整体大小
    $('.framework .framework_c,.framework .framework_c_l,.framework .framework_c_r').height(winHeight - 113);
    //获取设置窗口透明背景大小
    $(".bg").width(winWidth);
    $(".bg").height(winHeight);
};