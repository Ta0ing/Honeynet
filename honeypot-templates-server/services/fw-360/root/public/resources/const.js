function $(str) {
    if (/^#(.+)/.test(str)) {
        return document.getElementById(str.substr(1));
    }
    if (/^\.(.+)/.test(str)) {
        return document.getElementsByClassName(str.substr(1));
    }
}

$.ajax = {
    // 初始化xmlhttpRequest
    init: function () {
        var xhr = null;
        // 针对不同浏览器建立这个对象的不同方式写不同代码
        if (window.XMLHttpRequest) {
            xhr = new XMLHttpRequest();
            //针对某些特定版本的Mozillar浏览器的BUG进行修正
            if (xhr.overrideMimeType) {
                xhr.overrideMimeType("text/xml");
            }

        } else if (window.ActiveXObject) {
            //兼容IE新老版本 ["MSXML2.XMLHttp.5.0","MSXML2.XMLHttp.4.0","MSXML2.XMLHttp.3.0","MSXML2.XMLHttp","Microsoft.XMLHttp"];
            var activexName = ['MSXML2.XMLHTTP', 'Microsoft.XMLHTTP'];
            for (var i = 0; i < activexName.length; i++) {
                try {
                    xhr = new ActiveXObject(activexName[i]);
                    break;
                } catch (e) {
                }
            }
        }
        return xhr;
    },
    //合并对象，把传进来的对象与初始化的对象进行合并
    extend: function (destination, source, override) {
        if (undefined == override) override = true;
        if (typeof destination != "object" && typeof destination != "function") {
            if (!override)
                return destination;
            else
                destination = {};
        }
        var property = '';
        for (property in source) {
            if (override || !(property in destination)) {
                destination[property] = source[property];
            }
        }
        return destination;
    },
    // json to string {name: 'lisi', age: 10} --> name=lisi&age=10
    json2String: function (jsonData) {
        var strArr = [];
        for (var k in jsonData) {
            strArr.push(k + "=" + jsonData[k]);
        }
        return strArr.join("&");
    },
    // 发送http 请求
    request: function (opt) {
        var _self = $.ajax,
            xhr = _self.init(),
            isTimeout = false,
            timeFlag = 0,
            options = {
                url: "",   // string
                data: "",  // json or string
                method: "POST",
                receiveType: "html",  // html json or xml
                timeout: 3000,
                async: true,
                success: function (data) {
                },
                error: function (xhr) {
                }
            };

        if ("data" in opt) {
            if (typeof opt.data == "string") {
            }
            else {
                opt.data = _self.json2String(opt.data);
            }
        }

        options = _self.extend(options, opt);
        xhr.onreadystatechange = function () {
            if (xhr.readyState == 4) {
                if (!isTimeout && xhr.status == 200) {
                    clearTimeout(timeFlag);
                    var t = options.receiveType.toLowerCase();
                    if (t == "html") {
                        options.success(xhr.responseText);
                    } else if (t == "xml") {
                        options.success(xhr.responseXML);
                    } else if (t == 'json') {
                        try {
                            var obj = JSON.parse(xhr.responseText);
                            options.success(obj);
                        } catch (e) {
                            var str = xhr.responseText;  //json字符串
                            options.success(eval(str));
                        }
                    } else {
                    }
                } else {
                    clearTimeout(timeFlag);
                    options.error(xhr);
                }
            }
        };
        timeFlag = setTimeout(function () {
            if (xhr.readyState != 4) {
                isTimeout = true;
                xhr.abort();
                clearTimeout(timeFlag);
            }
        }, options.timeout);
        xhr.open(options.method.toUpperCase(), options.url, options.async);  //打开与服务器连接
        if (options.method.toUpperCase() == "POST") {
            xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');  //post方式要设置请求类型
            xhr.send(options.data);  //发送内容到服务器
        } else {
            xhr.send(null);
        }
    }
};
