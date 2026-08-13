$(document).ready(function() {

    //左导航展开收起
    $(".main .openStatus").click(function() {
        if ($(this).attr("src") == "images/open.png") {
            $(".mainLeft").animate({
                width: "0px"
            }, 500);

            $(this).attr("src", "images/close.png");
            $(".mainRight").animate({
                width: "1200px"
            }, 500);
        } else {
            $(".mainLeft").show();
            $(this).attr("src", "images/open.png");
            $(".mainLeft").animate({
                width: "250px"
            }, 500);

            $(".mainRight").animate({
                width: "930px"
            }, 500);
        }
    });
    //选项卡
    $(".mainRight .tab a").click(function() {
        $(".mainRight .tab a").removeClass("current");
        $(this).addClass("current");
    });
    $(".commonCon .tab a").click(function() {
        $(".commonCon .tab a").removeClass("current");
        $(this).addClass("current");
    });

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

});

//placeholder属性兼容
function fadeOut(a) {
    $(a).siblings(".placeholder").hide();
};

function spanFocus(a) {
    $(a).hide();
    $(a).siblings(".placeholderInput").focus();
};

function fadeIn(a) {
    var value = $(a).val();
    if (value == "") {
        $(a).siblings(".placeholder").show();
    }
};



//全选
function selectAll(checkbox) {
    $('.select').prop('checked', $(checkbox).prop('checked'));
}
//遮罩层显示
function shade() {
    $(".shade").show();
    parent.$(".topshade").show();

}
//二级遮罩显示
function shadee() {
    $(".shadee").show();

}
//时间戳转换成年月日24小时制时分秒

function getLocalTime(nS) {
    var date = new Date(nS * 1000);

    date.getFullYear();
    date.getMonth();
    date.getDate();
    date.getTime();
    date.getHours();
    date.getMinutes();
    date.getSeconds();

    Y = date.getFullYear() + '-';
    M = (date.getMonth() + 1 < 10 ? '0' + (date.getMonth() + 1) : date.getMonth() + 1) + '-';
    D = (date.getDate() < 10 ? '0' + (date.getDate()) : date.getDate()) + ' ';
    h = (date.getHours() < 10 ? '0' + (date.getHours()) : date.getHours()) + ':';
    m = (date.getMinutes() < 10 ? '0' + (date.getMinutes()) : date.getMinutes()) + ':';
    s = (date.getSeconds() < 10 ? '0' + (date.getSeconds()) : date.getSeconds());

    return (Y + M + D + h + m + s);

}
//时间戳换成年月日时分并是/连接
function getLocalTime1(nS) {
    var date = new Date(nS * 1000);

    date.getFullYear();
    date.getMonth();
    date.getDate();
    date.getTime();
    date.getHours();
    date.getMinutes();
    date.getSeconds();

    Y = date.getFullYear() + '/';
    M = (date.getMonth() + 1 < 10 ? '0' + (date.getMonth() + 1) : date.getMonth() + 1) + '/';
    D = (date.getDate() < 10 ? '0' + (date.getDate()) : date.getDate()) + ' ';
    h = (date.getHours() < 10 ? '0' + (date.getHours()) : date.getHours()) + ':';
    m = (date.getMinutes() < 10 ? '0' + (date.getMinutes()) : date.getMinutes()) + '';
    s = (date.getSeconds() < 10 ? '0' + (date.getSeconds()) : date.getSeconds());

    return (Y + M + D + h + m);

}
//时间戳换成年月日并是/连接
function getLocalTime2(nS) {
    var date = new Date(nS * 1000);

    date.getFullYear();
    date.getMonth();
    date.getDate();
    date.getTime();
    date.getHours();
    date.getMinutes();
    date.getSeconds();

    Y = date.getFullYear() + '/';
    M = (date.getMonth() + 1 < 10 ? '0' + (date.getMonth() + 1) : date.getMonth() + 1) + '/';
    D = (date.getDate() < 10 ? '0' + (date.getDate()) : date.getDate()) + ' ';
    h = (date.getHours() < 10 ? '0' + (date.getHours()) : date.getHours()) + ':';
    m = (date.getMinutes() < 10 ? '0' + (date.getMinutes()) : date.getMinutes()) + '';
    s = (date.getSeconds() < 10 ? '0' + (date.getSeconds()) : date.getSeconds());

    return (Y + M + D);

}
//时间戳换成年月日并是-连接
function getLocalTime3(nS) {
    var date = new Date(nS * 1000);

    date.getFullYear();
    date.getMonth();
    date.getDate();
    date.getTime();
    date.getHours();
    date.getMinutes();
    date.getSeconds();

    Y = date.getFullYear() + '-';
    M = (date.getMonth() + 1 < 10 ? '0' + (date.getMonth() + 1) : date.getMonth() + 1) + '-';
    D = (date.getDate() < 10 ? '0' + (date.getDate()) : date.getDate());
    h = (date.getHours() < 10 ? '0' + (date.getHours()) : date.getHours()) + ':';
    m = (date.getMinutes() < 10 ? '0' + (date.getMinutes()) : date.getMinutes()) + '';
    s = (date.getSeconds() < 10 ? '0' + (date.getSeconds()) : date.getSeconds());

    return (Y + M + D);

}
//时间戳换成时分
function getLocalTime4(nS) {
    var date = new Date(nS * 1000);

    date.getFullYear();
    date.getMonth();
    date.getDate();
    date.getTime();
    date.getHours();
    date.getMinutes();
    date.getSeconds();

    Y = date.getFullYear() + '-';
    M = (date.getMonth() + 1 < 10 ? '0' + (date.getMonth() + 1) : date.getMonth() + 1) + '-';
    D = (date.getDate() < 10 ? '0' + (date.getDate()) : date.getDate()) + ' ';
    h = (date.getHours() < 10 ? '0' + (date.getHours()) : date.getHours()) + ':';
    m = (date.getMinutes() < 10 ? '0' + (date.getMinutes()) : date.getMinutes()) + '';
    s = (date.getSeconds() < 10 ? '0' + (date.getSeconds()) : date.getSeconds());

    return (h + m);

}
//时间戳换成月日
function getLocalTime5(nS) {
    var date = new Date(nS * 1000);

    date.getFullYear();
    date.getMonth();
    date.getDate();
    date.getTime();
    date.getHours();
    date.getMinutes();
    date.getSeconds();

    Y = date.getFullYear() + '-';
    M = (date.getMonth() + 1 < 10 ? '0' + (date.getMonth() + 1) : date.getMonth() + 1) + '-';
    D = (date.getDate() < 10 ? '0' + (date.getDate()) : date.getDate()) + ' ';
    h = (date.getHours() < 10 ? '0' + (date.getHours()) : date.getHours()) + ':';
    m = (date.getMinutes() < 10 ? '0' + (date.getMinutes()) : date.getMinutes()) + '';
    s = (date.getSeconds() < 10 ? '0' + (date.getSeconds()) : date.getSeconds());

    return (M + D);

}
//几天前的日期
function GetDateStr(AddDayCount) {
    var dd = new Date();
    dd.setDate(dd.getDate() + AddDayCount); //获取AddDayCount天后的日期  
    var y = dd.getFullYear();
    var m = (dd.getMonth() + 1) < 10 ? "0" + (dd.getMonth() + 1) : (dd.getMonth() + 1); //获取当前月份的日期，不足10补0  
    var d = dd.getDate() < 10 ? "0" + dd.getDate() : dd.getDate(); //获取当前几号，不足10补0  
    return y + "-" + m + "-" + d;
}
//日期转换成时间戳
function getBeginTimes(YMD) {
    var str = YMD;
    str = str.replace(/-/g, '/');
    str = str + " 00:00:00";
    var date = new Date(str);
    var time = date.getTime();
    time = time / 1000;
    return time;

}

function getEndTimes(YMD) {
    var str = YMD;
    str = str.replace(/-/g, '/');
    str = str + " 23:59:59";
    var date = new Date(str);
    var time = date.getTime();
    time = time / 1000;
    return time;

}

//删除字符串左右两端的空格
function trim(str) {
    return str.replace(/(^\s*)|(\s*$)/g, "");
}
//截取字符串\最后的字符串并去掉空格（针对事件日志中的title）
function path(str) {
    var index = str.lastIndexOf("\\");
    var result = str.substring(index + 1, str.length);

    return result;
}
//去掉字符串中的所有空格
function pathtitle(str) {
    result = str.replace(/\s/g, "");
    return result;
}
//数字是否在数组中
function isInArray(arr, val) {
    var i, iLen;
    if (!(arr instanceof Array) || arr.length === 0) {
        return false;
    }
    if (typeof Array.prototype.indexOf === 'function') {
        return !!~arr.indexOf(val)
    }
    for (i = 0, iLen = arr.length; i < iLen; i++) {
        if (val === arr[i]) {
            return true;
        }
    }
    return false;
}
// ip地址正则验证
function isValidIP(ip) {
    var reg = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
    return reg.test(ip);
}

// 上传文件大小判断
function onUploadChange(fileInput) {
    var filesize;
    if ($.browser.msie) { //ie旧版浏览器
        var fileMgr = new ActiveXObject("Scripting.FileSystemObject");
        var filePath = $('#' + fileInput.id)[0].value;
        var fileObj = fileMgr.getFile(filePath);
        filesize = fileObj.size; //byte
        // filesize = filesize / 1024;//kb
        //size = size / 1024;//mb
    } else {
        var filePath = fileInput.value;
        if (fileInput.files && fileInput.files[0]) {
            filesize = fileInput.files[0].size;
        }
    }
    filesize = Math.floor(filesize / 1048576);
    if (filesize > 300) {
        resetFileInput($('#' + fileInput.id));

        $(".delayHide").show();
        $(".delayHide .p1").html("<img src='images/unusualw.png' class='verticalMiddle'> 不能超过300m");
        setTimeout(function() {
            $(".delayHide").hide()
        }, 2000);

        return;
    }


}
// 防御xss转换
function safeStr(str) {
    return str.replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}
//IP转成整型
function _ip2int(ip) {
    var num = 0;
    ip = ip.split(".");
    num = Number(ip[0]) * 256 * 256 * 256 + Number(ip[1]) * 256 * 256 + Number(ip[2]) * 256 + Number(ip[3]);
    num = num >>> 0;
    return num;
}
//整型解析为IP地址
function _int2ip(num) {
    var str;
    var tt = new Array();
    tt[0] = (num >>> 24) >>> 0;
    tt[1] = ((num << 8) >>> 24) >>> 0;
    tt[2] = (num << 16) >>> 24;
    tt[3] = (num << 24) >>> 24;
    str = String(tt[0]) + "." + String(tt[1]) + "." + String(tt[2]) + "." + String(tt[3]);
    return str;
}
//比较两个ip的大小,如果大于，返回1，等于返回0，小于返回-1 
function compareIP(ipBegin, ipEnd) {
    var temp1
    var temp2
    temp1 = ipBegin.split('.')
    temp2 = ipEnd.split('.')
    for (var i = 0; i < 4; i++) {
        if (parseInt(temp1[i]) > parseInt(temp2[i])) {
            return 1
        } else if (parseInt(temp1[i]) < parseInt(temp2[i])) {
            return -1
        }
    }
    return 0
}

//字符串转换成布尔值
function stringToBoolean(str) {
    switch (str.toLowerCase()) {
        case "true":
        case "yes":
        case "1":
            return true;
        case "false":
        case "no":
        case "0":
        case null:
            return false;
        default:
            return Boolean(str);
    }
}
//获取360浏览器信息
function _mime(option, value) {
    var mimeTypes = navigator.mimeTypes;
    for (var mt in mimeTypes) {
        if (mimeTypes[mt][option] == value) {
            return true;
        }
    }
    return false;
}
//判断ie版本
function _ieVersion(userAgent) {
    var ieVer = IE5 = IE55 = IE6 = IE7 = IE8 = IE9 = IE10 = '';
    var reIE = new RegExp("msie (\\d+\\.\\d+);");
    reIE.test(userAgent);
    var fIEVersion = parseFloat(RegExp["$1"]);
    console.log(fIEVersion);
    IE55 = fIEVersion == 5.5;
    IE6 = fIEVersion == 6.0;
    IE7 = fIEVersion == 7.0;
    IE8 = fIEVersion == 8.0;
    IE9 = fIEVersion == 9.0;
    IE10 = fIEVersion == 10.0;
    return ieVer;
}
//得到浏览器信息
function getBrowserInfo() {
    var userAgent = navigator.userAgent.toLowerCase(); //取得浏览器的userAgent字符串
    var isIE = userAgent.indexOf("compatible") > -1 && userAgent.indexOf("msie") > -1 && !isOpera; //判断是否IE浏览器
    var isIE11 = (userAgent.toLowerCase().indexOf("trident") > -1 && userAgent.indexOf("rv") > -1);
    var Sys = {};
    var pass = false;
    if (isIE11) {
        pass = false;
    } else if (isIE) {
        pass = false;
    } else {
        //区分360浏览器和谷歌
        var is360 = _mime("type", "application/vnd.chromium.remoting-viewer");
        var isOpera = userAgent.indexOf("opr") > -1; //判断是否Opera浏览器
        var isEdge = userAgent.indexOf("edge") > -1; //edge浏览器
        var isFF = userAgent.indexOf("firefox") > -1; //判断是否Firefox浏览器
        var isSafari = userAgent.indexOf("safari") > -1 && !is360; //判断是否Safari浏览器
        var isChrome = userAgent.indexOf("chrome") > -1 && !isEdge && !is360 && !isOpera; //判断是否chrome浏览器
        var re = /(firefox|safari|chrome|version).*?([\d.]+)/;
        var m = userAgent.match(re);
        Sys.browser = m[1].replace(/version/, "'safari");
        Sys.ver = parseFloat(m[2].split('.')[0]);

        if (isChrome) { //谷歌区分360浏览器\edge\opera
            if (Sys.ver >= 69) {
                pass = true;
            }
        }
        if (Sys.browser == 'firefox' && Sys.ver >= 60) {
            pass = true;
        }
        if (Sys.browser == 'safari' && Sys.ver >= 10) {
            pass = true;
        }
        if (isOpera) {
            re = /(opr|version).*?([\d.]+)/;
            m = userAgent.match(re);
            Sys.ver = parseFloat(m[2].split('.')[0]);
            if (Sys.ver >= 56) {
                pass = true;
            }
        }
        if (isEdge) {
            re = /(edge|version).*?([\d.]+)/;
            m = userAgent.match(re);
            Sys.ver = parseFloat(m[2].split('.')[0]);
            if (Sys.ver >= 17) {
                pass = true;
            }

        }
    }
    return pass;


}

//按回车搜索
$(document).on('keydown', '#numperpageinput', function(e) {
    if (!e) e = window.event;
    if ((e.keyCode || e.which) == 13) {
        $(this).blur();
    }
})

// 每页多少数据数字限制
$("body").on("keyup", "#numperpageinput", function() {
    var value = $(this).val().replace(/[^\d]/g, '');
    $(this).val(value);
})
var numperpage;
if (localStorage.getItem('numperpage')) {
    numperpage = parseInt(localStorage.getItem('numperpage'));
} else {
    numperpage = 15;
}



//排序----图片路径判断
function imgOrderyFun(thIndex, thCls) {
    if (thCls == 'th-ordery') {
        $('.main .tableth th').eq(thIndex).find('img').attr('src', 'images/th-ordery.png');
    } else if (thCls == 'th-ordery th-ordery-current th-ordery-up') {
        $('.main .tableth th').eq(thIndex).find('img').attr('src', 'images/th-ordery-up.png');
    } else if (thCls == 'th-ordery th-ordery-current th-ordery-up th-ordery-down') {
        $('.main .tableth th').eq(thIndex).find('img').attr('src', 'images/th-ordery-down.png');
    }

}

function imgOrderyFun1(thIndex, thCls) {
    if (thCls == 'th-ordery') {
        $('.pop .tableth th').eq(thIndex).find('img').attr('src', 'images/th-ordery.png');
    } else if (thCls == 'th-ordery th-ordery-current th-ordery-up') {
        $('.pop .tableth th').eq(thIndex).find('img').attr('src', 'images/th-ordery-up.png');
    } else if (thCls == 'th-ordery th-ordery-current th-ordery-up th-ordery-down') {
        $('.pop .tableth th').eq(thIndex).find('img').attr('src', 'images/th-ordery-down.png');
    }

}

// 原生ajax请求导出excel

function download_excel_using_blob(dataa, url, file_name) {
    var url = url; //【全部导出】请求url
    var xhr = new XMLHttpRequest();
    xhr.open('POST', url, true);
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
    xhr.responseType = "blob"; //返回类型blob
    xhr.onload = function() { //定义请求完成的处理函数
        //layer.closeAll('loading');
        if (this.status === 200) {
            var blob = new Blob([xhr.response], {
                type: 'application/vnd.ms-excel;'
            });
            // for IE
            if (window.navigator && window.navigator.msSaveOrOpenBlob) {
                window.navigator.msSaveOrOpenBlob(blob, file_name);
            } else {
                // for no-ie
                var objectURL = URL.createObjectURL(blob);
                var downloadElement = document.createElement('a'); //转换完成，创建一个a标签用于下载
                downloadElement.download = 'virusLog.xlsx';
                downloadElement.href = objectURL;
                document.body.appendChild(downloadElement);
                downloadElement.click();
                document.body.removeChild(downloadElement);
            }
        } else if (this.status === 401) {
            parent.window.location.href = '/';
        } else {

        }
    };
    xhr.send(JSON.stringify(dataa));
}
//ie8对于input类型为file文件不支持val（''）,下面方法兼容浏览器
function resetFileInput(file) {
    file.after(file.clone().val(""));
    file.remove();
}


// 文件导出
jQuery.download = function(url, method, param) {
    var content = '<form action="' + url + '" method="' + (method || 'post') + '" enctype="application/x-www-form-urlencoded">';
    content += "<input type='hidden' name='cond' value='" + JSON.stringify(param) + "'/>";
    content += '</form>';
    jQuery(content).appendTo('body').submit().remove();
}

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


var clientid = "";
var grouppolicyname = "";
var prevTName = "";

function detailPop(a, cb) {
    var top = parent.$(".detailPop");
    var spanIndex = $(a).attr('index');
    var clientid = $(a).parents("tr").attr("clientid");
    parent.$('.shade').show();
    top.show();
    top.attr('clientid', clientid);
    terminal_detail(clientid, function(data) {
        var data = data.data[0];
        var html = "";
        var terpolicyid = data.policy_id; //读取终端详情时的终端独立策略id
        grouppolicyname = data.policy_name;
        prevTName = data.hostname; //读取终端详情时的终端名称
        if (cb) {
            cb(prevTName);
            return;
        }
        if (data.status == 1) {
            top.find(".detailcom").html("<img src='images/unusualter.png'>");
        } else {
            if (data.online == true) {
                top.find(".detailcom").html("<img src='images/onlineter.png'>");
            } else {
                top.find(".detailcom").html("<img src='images/offlineter.png'>");
            }
        }

        // $(".detailPop .versionnum").html("<span style='display:inline-block;width:40px;text-align:justify; text-justify:inter-ideograph;'><font style='margin-right:12px;;'>版</font>本</span> : "+data.prod.ver);
        // $(".detailPop .dbversion").html("<span style='display:inline-block;width:40px;text-align:justify; text-justify:inter-ideograph;'>病毒库</span> : "+getLocalTime1(data.prod.dbver));

        html += "<div class='informationB informationBL'>";
        html += "<div class='heading'><img src='images/information2.png'><span>网络信息</span></div>";
        html += "<p>IP地址 : " + safeStr(data.ip) + "</p>";
        html += "<p>MAC地址 : " + safeStr(data.mac) + "</p>";
        html += "</div>";
        html += "<div class='clear'></div>";
        html += "<div class='informationB informationBD'>";
        html += "<div class='heading'><img src='images/information3.png'><span>环境信息</span></div>";
        html += "<p>操作系统：" + safeStr(data.osinfo.name) + "    " + safeStr(data.osinfo.arch) + "   " + safeStr(data.osinfo.ver) + "</p>";
        if ("cpus" in data.hwinfo) {
            if (data.hwinfo.cpus) {
                html += "<p><span style='width:11%;display:inline-block;vertical-align: top;'>处理器：</span><span style='display:inline-block;width:89%;'>"
                for (var i = 0; i < data.hwinfo.cpus.length; i++) {
                    html += "<a style='display:inline-block;width:100%'>" + data.hwinfo.cpus[i] + "</a>";
                }
                html += "</span></p>"
            } else {
                html += "<p>处理器：</p>";
            }

        } else {
            html += "<p>处理器：" + safeStr(data.hwinfo.cpu) + "</p>";
        }

        html += "<p>主板：" + safeStr(data.hwinfo.board) + "</p>";
        if ("physicalmemory" in data.hwinfo) {
            if (data.hwinfo.physicalmemory) {
                html += "<p><span style='width:8%;display:inline-block;vertical-align: top;'>内存：</span><span style='display:inline-block;width:90%;'>"
                for (var i = 0; i < data.hwinfo.physicalmemory.length; i++) {
                    html += "<a style='display:inline-block;width:100%'>" + data.hwinfo.physicalmemory[i] + "</a>";
                }
                html += "</span></p>"
            } else {
                html += "<p>内存：</p>";
            }

        } else {
            html += "<p>内存：" + data.hwinfo.memory + "</p>";
        }

        if ("hdds" in data.hwinfo) {
            if (data.hwinfo.hdds) {
                html += "<p><span style='width:8%;display:inline-block;vertical-align: top;'>硬盘：</span><span style='display:inline-block;width:90%;'>"
                for (var i = 0; i < data.hwinfo.hdds.length; i++) {
                    html += "<a style='display:inline-block;width:100%'>" + data.hwinfo.hdds[i] + "</a>";
                }
                html += "</span></p>"
            } else {
                html += "<p>硬盘：</p>";
            }

        } else {
            html += "<p>硬盘：" + data.hwinfo.hdd + "</p>";
        }

        if ("videos" in data.hwinfo) {
            if (data.hwinfo.videos) {
                html += "<p><span style='width:8%;display:inline-block;vertical-align: top;'>显卡：</span><span style='display:inline-block;width:90%;'>"
                for (var i = 0; i < data.hwinfo.videos.length; i++) {
                    html += "<a style='display:inline-block;width:100%'>" + data.hwinfo.videos[i] + "</a>";
                }
                html += "</span></p>"
            } else {
                html += "<p>显卡：</p>";
            }
        } else {
            html += "<p>显卡：" + data.hwinfo.video + "</p>";
        }

        html += "</div>";
        html += "<div class='informationB informationBD'>";
        html += "<div class='heading'><img src='images/information4.png'><span>其它信息</span></div>";
        html += "<p>终端版本 : " + safeStr(data.prod.ver) + "</p>";
        html += "<p>病毒库时间 : " + safeStr(getLocalTime1(data.prod.dbver)) + "</p>";
        html += "<p>上次登录时间 : " + safeStr(getLocalTime1(data.lasttime)) + "</p>";
        html += "</div>";
        top.find(".terOverview").html(html);
        if (data.status == 1) {
            top.find(".left .terminalstatus").html("<font style='color:#b28041;'>服务异常</font>");

        } else if (data.online == true && data.status == 0) {
            top.find(".left .terminalstatus").html("<font style='color:#42aa76;'>在线</font>");
        } else if (data.online == false && data.status == 0) {
            top.find(".left .terminalstatus").html("<font style='color:#6c6c6c;'>离线</font>");
        }
        top.find(".left .terminalname").html("<span class='verticalMiddle' style='padding-left:18px;display: inline-block;max-width: 160px;text-overflow:ellipsis;overflow: hidden;white-space: nowrap;vertical-align: middle;' title='" + safeStr(pathtitle(data.hostname)) + "'>" + safeStr(data.hostname) + "</span><span class='editTNIcon' index='" + parseInt(spanIndex) + "' onclick='editTN(this)'></span>");
        top.find(".left .isgroup").html(safeStr(data.groupname));
    })
};

function terminal_detail(clientid, cb) {
    $.ajax({
        url: '/mgr/clnt/_info?id=' + clientid + '',
        contentType: 'text/plain',
        headers: {
            "HTTP-CSRF-TOKEN": getCookie('HRESSCSRF')
        },
        error: function(xhr, textStatus, errorThrown) {
            if (xhr.status == 401) {
                parent.window.location.href = '/';
            }

        },
        success: function(data) {
            cb && cb(data);
        }
    });
}
//编辑终端名称按钮
function editTN(a) {
    var spanIndex = $(a).attr('index');
    var tern = $(a).prev().text();
    $(a).prev().remove();
    $(a).before("<input type='text' style='height:20px;line-height:20px;width:140px;color:#676a6c;border:1px solid rgb(231,234,236);text-align:center;font-size:12px;'  class='terminalNInput' index='" + parseInt(spanIndex) + "' onblur='sureETN(this)'>");
    $(a).prev().focus();
    $(a).prev().val(tern);
    $(a).prev().attr('oldVal', tern);
    $(a).hide();
}
//编辑终端名称文本框失去焦点
function sureETN(a) {
    var terminalname = trim($(a).val());
    var oldVal = $(a).attr('oldVal');
    var spanIndex = $(a).attr('index');
    var clientid = $(a).parents('.detailPop').attr('clientid');
    $(a).next().show();

    if (oldVal == terminalname) {
        $(".detailPop .left .terminalname").html("<span class='verticalMiddle' style='padding-left:18px;display: inline-block;max-width: 160px;text-overflow:ellipsis;overflow: hidden;white-space: nowrap;vertical-align: middle;'>" + safeStr(terminalname) + "</span><span class='editTNIcon' index='" + parseInt(spanIndex) + "' onclick='editTN(this)'></span>");
    } else {
        var dataa = {
            "client_id": parseInt(clientid),
            "aliasname": terminalname
        };
        $.ajax({
            url: '/mgr/clnt/_update',
            data: JSON.stringify(dataa),
            type: 'POST',
            contentType: 'text/plain',
            headers: {
                "HTTP-CSRF-TOKEN": getCookie('HRESSCSRF')
            },
            error: function(xhr, textStatus, errorThrown) {
                if (xhr.status == 401) {
                    parent.window.location.href = '/';
                }
            },
            success: function(data) {
                if (terminalname == '') {
                    terminal_detail(clientid, function(data) {
                        var hostname = data.data[0].hostname;
                        $(".detailPop .left .terminalname").html("<span class='verticalMiddle' style='padding-left:18px;display: inline-block;max-width: 160px;text-overflow:ellipsis;overflow: hidden;white-space: nowrap;vertical-align: middle;'>" + safeStr(hostname) + "</span><span class='editTNIcon' index='" + parseInt(spanIndex) + "' onclick='editTN(this)'></span>");
                    })
                } else {
                    $(".detailPop .left .terminalname").html("<span class='verticalMiddle' style='padding-left:18px;display: inline-block;max-width: 160px;text-overflow:ellipsis;overflow: hidden;white-space: nowrap;vertical-align: middle;'>" + safeStr(terminalname) + "</span><span class='editTNIcon' index='" + parseInt(spanIndex) + "' onclick='editTN(this)'></span>");
                }
                document.getElementById('mainFrame').contentWindow.refreshList();

            }
        });
    }
}

var _self = {
    data: {
        url_totp1: '/auth/_totp',
        url_totp2: '/auth/_totp2'
    },
    init: function(param) {
        _self.build();
    },
    request: function(options, cb) {
        $.ajax({
            url: options.url || '',
            data: JSON.stringify(options.data) || [],
            type: options.type || 'POST',
            contentType: options.contentType || 'text/plain',
            headers: {
                "HTTP-CSRF-TOKEN": getCookie('HRESSCSRF')
            },
            error: function(xhr, textStatus, errorThrown) {
                if (xhr.status == 401) {
                    parent.window.location.href = '/';
                }
            },
            success: function(data) {
                if ($.isFunction(cb)) cb(data)
            }
        });
    },
    build: function(param) {
        //更新认证
        $(document).on('click', '.totpPop .srueBtn', function() {
            //发送动态验证请求 
            var $popElem = $(this).parents('.totpPop'),
                code = $popElem.find('input[name=code]').val(),
                pageName = $popElem.attr('data-page');

            if (code.length != 6) {
                $popElem.find('.errorTips').show();
                $popElem.find('.errorTips').html(' 动态口令为6位数');
            } else {
                _self.authTotp(code, pageName);
            }

        }).on('blur', '.totpPop input[name=code]', function() {
            var $popElem = $(this).parents('.totpPop'),
                val = $(this).val();
            if (val.length != 6) {
                $popElem.find('.errorTips').show();
            } else {
                $popElem.find('.errorTips').hide();
            }
        }).on('keyup', '.totpPop input[name=code]', function() {
            var oldVal = $(this).val(),
                regVal = oldVal.replace(/[\D]/g, ""),
                finalVal;

            regVal.length > 6 ? finalVal = regVal.substring(0, 6) : finalVal = regVal;

            $(this).val(finalVal);
        })
        //按回车登陆
        document.onkeydown = function(e) {
            if (!e) e = window.event;

            if ((e.keyCode || e.which) == 13) {
                if ($('.totpPop').css('display') == 'block') {
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

    },
    openTotpPop: function(index) {
        var $popElem_1 = $('.totpPop'),
            $popElem_2 = parent.$('.totpPop'),
            $popElem;

        index ? (index == 1 ? $popElem = $popElem_1 : $popElem = $popElem_2) : $popElem = $popElem_2;

        _self.showPop($popElem, index);
        //动态口令input相关操作
        $popElem.find('input[name=code]').val('');
        $popElem.find('.errorTips').hide();
        $popElem.find('input[name=code]').focus();
        //当前时间
        _self.getCurrentTime($popElem);



    },
    getCurrentTime: function($popElem) {
        _self.request({
            url: 'deploy/info'
        }, function(param) {
            $popElem.find('.localTime').html(param.data.time);
        })

    },
    authTotp: function(code, pageName) {
        var dataa = {
                "code": code
            },
            url;
        pageName == 'login' ? url = _self.data.url_totp1 : url = _self.data.url_totp2;
        _self.request({
            url: url,
            data: dataa
        }, function(data) {
            $('.totpPop').hide();
            parent.$('.shade').hide();
            if (data.errno == 0) {
                if (pageName == 'login') {
                    window.location.href = "/";
                }
            } else if (data.errno == -2) {
                _self.showPop($('.aLockHintPop'), 2);
            } else {
                _self.delayHide('error', '动态口令错误');
                var inp = $("#mainFrame").contents().find('.settingPop ').find("input[name=totp_enabled]");

                if (inp.length > 0) {
                    inp.prop('checked', true);
                }
            }
        })
    },
    delayHide: function(errno, text) {
        $(".delayHide").show();
        $(".delayHide .p1").html("<img src='images/unusual.png' class='verticalMiddle'><span class='verticalMiddle'> " + text + "</span>");
        setTimeout(function() {
            $(".delayHide").hide()
        }, 2000);
    },
    showPop: function(elem, index) {
        elem.show();
        if (index == '1') {
            $('.shade').show();
        } else {
            parent.$('.shade').show();
        }
    },
    hidePop: function(elem, index) {
        elem.hide();
        if (index == '1') {
            $('.shade').hide();
        } else if (index == '0') {
            parent.$('.shade').hide();
        } else if (index == '2') {
            $('.shadee').hide();
        }
    },
    getFormatDate: function(param) {
        var nowDate = new Date();
        var year = nowDate.getFullYear();
        var month = nowDate.getMonth() + 1 < 10 ? "0" + (nowDate.getMonth() + 1) : nowDate.getMonth() + 1;
        var date = nowDate.getDate() < 10 ? "0" + nowDate.getDate() : nowDate.getDate();
        var hour = nowDate.getHours() < 10 ? "0" + nowDate.getHours() : nowDate.getHours();
        var minute = nowDate.getMinutes() < 10 ? "0" + nowDate.getMinutes() : nowDate.getMinutes();
        var second = nowDate.getSeconds() < 10 ? "0" + nowDate.getSeconds() : nowDate.getSeconds();
        return year + "-" + month + "-" + date + " " + hour + ":" + minute + ":" + second;
    },
    addInputEvent: function(dom, retrunVal) {
        $(document).delegate(dom, 'input', function(e) {
            var $this = $(this);
            var value = $this.val();
            var _value = parseInt(value);

            if (value === "") {
                return;
            }
            if (!$.isNumeric(_value) || _value <= 0) {
                $this.focus().val(retrunVal);
            } else {

                $this.focus().val(_value);
            }
        })
    },
    keyDownEvent: function(keycodeJson) {
        $(document).off('keydown', keycodeJson.obj).on('keydown', keycodeJson.obj, function(event) {
            var e = event || window.event || arguments.callee.caller.arguments[0],
                val = $(this).val()
            keyCode = e.keyCode;

            //输入空格时
            if (e && keyCode == 32) {
                e.preventDefault ? e.preventDefault() : (e.returnValue = false);
                return false;
            }
        })
    },
    // 获取多终端终端信息
    getClntInfo: function(obj) {
        $.ajax({
            url: '/mgr/clnt/_info',
            data: JSON.stringify({
                client_id: obj.client_id || []
            }),
            type: 'POST',
            contentType: 'text/plain',
            headers: {
                "HTTP-CSRF-TOKEN": getCookie('HRESSCSRF')
            },
            error: function(xhr, textStatus, errorThrown) {
                if (xhr.status == 401) {
                    parent.window.location.href = '/';
                } else {}
            },
            success: function(data) {
                var data = data.data;
                if (!!obj.cb_custom) {
                    obj.cb_custom && obj.cb_custom(data);
                    return false;
                }
                var html = "";
                for (var i = 0, len = data.length; i < len; i++) {
                    var dataItem = data[i];
                    html += "<tr>";
                    html += "<td><span style='width:220px;' class='filePath' title='" + safeStr(dataItem.hostname) + "'>" + safeStr(dataItem.hostname) + "</span></td>";
                    html += "<td width='220'>" + safeStr(dataItem.groupname) + "</td>";
                    html += "<td width='220'>等待任务</td>";
                    html += "</tr>";
                }
                obj.obj.append(html);
            }
        });
    }
}

_self.init();