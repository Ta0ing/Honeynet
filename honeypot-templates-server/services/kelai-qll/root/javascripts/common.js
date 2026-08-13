/**
 * @ alert的title，使用js必须在页面初始化字符串invalid_title
 */
var invalid_title;



/**
 * @bref			Set element val
 */
var setVal = function(el, val) {

    $elem = $("[name=" + el + "]");
    if ($elem[0].nodeName == "INPUT" || $elem[0].nodeName == "SELECT") {
        $elem.val(val);
        return;
    }
    $elem.text(val);
}

/**
 * @bref			Set element val
 */
var getVal = function(el) {

    $elem = $("[name=" + el + "]");
    if ($elem[0].nodeName.toLowerCase() == "input") {
        if ($elem.attr("type") == "checkbox" || $elem.attr("type") == "radio") {

            return $elem.attr("checked") ? true : false;
        }
        return $elem.val();
    }
    if ($elem[0].nodeName == "INPUT" || $elem[0].nodeName == "SELECT") {

        return $elem.val();
    }
    return $elem.text();
}

/**
 * @bref     Format bits per second
 */
var formatBitPerSecond = function(bps) {

    var unit = [" bps", " Kbps", " Mbps", " Gbps", " Tbps"];
    var index = 0;

    if (bps == -1) {

        return "N/A";
    }

    while (bps > 1000) {

        index++;
        bps = bps / 1000;
    }
    return bps.toFixed(3) + unit[index];
}

/**
 * @bref     Format byte
 */
var formatByte = function(bytes) {

    var unit = [" B", " KB", " MB", " GB", " TB", " PB", " EB"];
    var index = 0;

    if (bytes == -1) {
        return "N/A";
    }

    while (bytes > 1024) {
        index++;
        bytes = bytes / 1024;
    }
    return Math.floor(bytes) + unit[index];
}

var formatByteFixed = function(bytes) {
    var f = new Number(bytes);
    var unit = [" B", " KB", " MB", " GB", " TB", " PB", " EB"];
    var index = 0;

    if (f == -1) {
        return "N/A";
    }

    while (f > 1024) {
        index++;
        f = f / 1024;
    }
    return f.toFixed(2) + unit[index];
}

var fmtBytes2GB = function(bytes) {
    return parseInt(bytes / 1073741824) + "GB";
}

var fmtBytes2GBToDecimal2 = function(bytes) {
    var f = parseFloat(bytes / 1073741824);
    return Math.round(f * 100) / 100 + "GB";
}

/**
 * @bref integer checker
 **/
var isInteger = function($input, option) {
    var _text = _element.val();
    return (_text.length >= option.min && _text.length <= option.max);
}

/**
 * @bref regular checker
 **/
var regular = function($input) {

    var option = $input.attr("vopt");
    return ((new RegExp(option.reg)).test($input.val()));
}

/**
 * @bref formar, example:"hello{0}".format('world'); return string "hello world"
 **/
String.prototype.format = function() {
    var args = arguments;
    return this.replace(/\{(\d+)\}/g, function(s, i) {
        return args[i];
    });
}




















/********************************************************************************/
/* unified style enable/disable													*/
/********************************************************************************/
function unifieddisable(selector, disable) {
    selector.each(function(i) {
        var dis_class = "unified_" + $(this).attr("type") + "_disabled";
        $(this).attr("disabled", disable).toggleClass(dis_class, disable);
    });
}

/********************************************************************************/
/* map form	fields																*/
/********************************************************************************/
function mapFormFields(selector) {
    var mapFields = {};
    selector.each(function() {
        if (($(this).attr("type") != "checkbox" && $(this).attr("type") != "radio") || $(this).attr("checked")) {
            mapFields[this.name] = this.value;
        }
    });
    return mapFields;
}

/********************************************************************************/
/* map form	fields2																*/
/********************************************************************************/
function mapFormFields2(selector) {
    var mapFields = {};
    selector.find(":input, select, textarea").not(":disabled").each(function() {
        if (($(this).attr("type") != "checkbox" && $(this).attr("type") != "radio") || $(this).attr("checked")) {
            mapFields[this.name] = this.value;
        }
    });
    return mapFields;
}

/********************************************************************************/
/* map form	fields, Include disabled element. 									*/
/********************************************************************************/
function mapFormFields3(selector) {
    var mapFields = {};
    selector.find(":input, select, textarea").each(function() {
        if (this.name != "" && ($(this).attr("type") != "checkbox" && $(this).attr("type") != "radio") || $(this).attr("checked")) {
            mapFields[this.name] = this.value;
        }
    });
    return mapFields;
}

/********************************************************************************/
/* map form	fields, Include disabled element. 									*/
/********************************************************************************/
function mapFormFields4(selector, unchecked) {
    var mapFields = {};
    selector.find(":input, select, textarea").each(function() {

        if (($(this).attr("type") != "checkbox" && $(this).attr("type") != "radio")) {

            mapFields[this.name] = this.value;
        } else {

            if ($(this).attr("checked")) {

                mapFields[this.name] = 1;

            } else {

                if (unchecked) {

                    mapFields[this.name] = 0;
                }
            }
        }
    });
    return mapFields;
}

/********************************************************************************/
/* number checker																*/
/********************************************************************************/
function numberChecker(sPort, nmin, nmax) {
    var parten = /^(\d)+$/g;
    if (parten.test(sPort) && parseInt(sPort) >= nmin && parseInt(sPort) <= nmax) {
        return true;
    }
    return false;
}

/********************************************************************************/
/* IP address checker															*/
/********************************************************************************/
function isIPa(strIP) {
    var exp = /^([1-9]|[1-9]\d|1\d{2}|2[0-1]\d|22[0-3])(\.(\d|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])){3}$/;
    return exp.test(strIP);
}

/********************************************************************************/
/* MASK address checker															*/
/********************************************************************************/
function isMaska(strMask) {
    var exp = /^(254|252|248|240|224|192|128|0)\.0\.0\.0|255\.(254|252|248|240|224|192|128|0)\.0\.0|255\.255\.(254|252|248|240|224|192|128|0)\.0|255\.255\.255\.(254|252|248|240|224|192|128|0)$/;
    return exp.test(strMask);
}

/********************************************************************************/
/* DNS address checker															*/
/********************************************************************************/
function isDnsa(strDns) {
    var isOK = true;
    $.each(strDns.split(","), function(key, value) {
        isOK &= isIPa(Trim(value));
    });
    return isOK;
}


function Trim(str) {
    str = TrimLeft(str);
    str = TrimRight(str);
    return str;
}

function TrimLeft(str) {
    var nPos = 0;
    for (var i = 0; i < str.length; i++) {
        if (str.charAt(i) != ' ' && str.charAt(i) != '\r' && str.charAt(i) != '\n') {
            nPos = i;
            break;
        }
    }
    return str.substr(nPos);
}

function TrimRight(str) {
    if (str.length == 0)
        return str;

    var nPos = str.length - 1;
    for (var i = nPos; i >= 0; i--) {
        if (str.charAt(i) != ' ' && str.charAt(i) != '\r' && str.charAt(i) != '\n') {
            nPos = i;
            break;
        }
    }
    return str.substr(0, nPos + 1);
}

var _st = window.setTimeout;
window.setTimeout = function(fRef, mDelay) {
    if (typeof fRef == 'function') {
        var argu = Array.prototype.slice.call(arguments, 2);
        var f = (function() {
            fRef.apply(null, argu);
        });
        return _st(f, mDelay);
    }
    return _st(fRef, mDelay);
};

function IsValidNumber(str) {
    if (str.length == 0)
        return false;

    for (var i = 0; i < str.length; i++) {
        if (str.charAt(i) < '0' || str.charAt(i) > '9')
            return false;
    }
    return true;
}

function IsValidLoginName(str) {
    // login name: number, charactor, symobl
    if (str.length == 0)
        return false;

    for (var i = 0; i < str.length; i++) {
        if (str.charAt(i) < '0' || str.charAt(i) > '9') {
            if (str.charAt(i) < 'a' || str.charAt(i) > 'z') {
                if (str.charAt(i) < 'A' || str.charAt(i) > 'Z') {
                    if (str.charAt(i) != '.' && str.charAt(i) != '_')
                        return false;
                }
            }
        }
    }
    return true;
}

/****************************************************************************/
/*	get url params															*/
/****************************************************************************/
function getUrlParams(str) {
    var url = window.location.toString();
    var rs = new RegExp("(^|)" + str + "=([^\&]*)(\&|$)", "gi").exec(url),
        tmp;
    if (tmp = rs)
        return tmp[2];
    return "";
}

function getUrlParams2(url, str) {
    var rs = new RegExp("(^|)" + str + "=([^\&]*)(\&|$)", "gi").exec(url),
        tmp;
    if (tmp = rs)
        return tmp[2];
    return "";
}

/****************************************************************************/
/*	get session key														*/
/****************************************************************************/
function getSessionKey() {
    var rs = new RegExp("(^|)session=([^;]*)(;|$)", "gi").exec(document.cookie),
        tmp;
    if (tmp = rs)
        return tmp[2];
    return "";
}

/****************************************************************************/
/*	alert dialog															*/
/****************************************************************************/
$.alert = function(title, text, href) {
        href = (typeof(href) == "undefined") ? "#" : href;
        var OK = "确定";
        $("#modaltoggle").remove();
        $(".modal-backdrop").remove();
        $("BODY").append('<div id="modaltoggle" class="modal hide" data-backdrop="static" style="display: none;z-index:9999">' +
            '<div class="modal-header">' +
            ' <a href="#" class="close" data-dismiss="modal">×</a>' +
            '<h4>' + title + '</h4>' +
            ' </div>' +
            '<div class="modal-body">' +
            text +
            '    </div>' +
            '   <div class="modal-footer">' +
            '     <a href="#" class="btn" data-dismiss="modal">' + OK + '</a>' +
            '    </div>' +
            '  </div>');

        if (href != '#') {

            $(".modal-header A").bind("click", function() {

                window.location.href = href;
            });

            $(".modal-footer A").bind("click", function() {

                window.location.href = href;
            });
        }
        $("#modaltoggle").modal('toggle');
    }

    /****************************************************************************/
    /*	simple validate frame													*/
    /****************************************************************************/
    ! function($) {

        "use strict"; // jshint ;_;

        $.fn.jsValidator = function(options) {
            var methods = $.extend({}, $.fn.jsValidator.defaults, options);

            function _getMethods(_element) {
                var validate = _element.attr("validate");
                var fn = validate.substr(0, validate.indexOf(":"));
                var op = eval("(" + validate.substr(validate.indexOf(":") + 1) + ")");
                return {
                    'fn': methods[fn],
                    'option': op
                };
            }

            var _ret = true;
            $(this).find('[validate]').not(":disabled").each(function() {
                var _element = $(this);
                var _method = _getMethods(_element);
                if (_method.fn(_element, _method.option)) {
                    return true;
                }

                if (_method.option.failed != undefined)
                    $.alert(invalid_title, _method.option.failed);

                if (_element.attr('msg') != undefined)
                    $.alert(invalid_title, _element.attr('msg'));
                _ret = false;
                return false;
            });

            $(this).find("[data-checker]").not(":disabled").each(function(index) {

                var $input = $(this);
                var stringFunction = $input.attr("data-checker") + "($input)";
                if (!eval(stringFunction)) {

                    if ($input.attr("data-error") != undefined) {

                        $.alert(invalid_title, $input.attr("data-error"));
                    }
                    _ret = false;
                }
            });

            return _ret;
        }

        $.fn._length = function(_element, option) {
            var _text = _element.val();
            return (_text.length >= option.min && _text.length <= option.max);
        }

        $.fn._repeat = function(_element, option) {
            return (_element.val() == $("#" + option.repeatid).val());
        }

        $.fn._integer = function(_element, option) {
            var integer = parseInt(_element.val());
            return (integer >= option.min && integer <= option.max);
        }

        $.fn._regular = function(_element, option) {
            return ((new RegExp(option.reg)).test(_element.val()));
        }

        $.fn._nochange = function(_element, option) {
            return (_element.find(":input,select,textarea, :checkbox, :password").serialize() != option.init);
        }

        $.fn._nochangead = function(_element, option) {
            var invalid = false;
            $("#" + option.id).find(":text, input:hidden, select, textarea, :password").not(":disabled").each(function() {
                if (($(this).val() != $(this).attr("init"))) invalid = true;
            });

            $("#" + option.id).find(":checkbox, :radio").not(":disabled").each(function() {
                var check = $(this).attr("checked") ? "checked" : "";
                if (check != $(this).attr("init")) invalid = true;
            });
            return invalid;
        }

        $.fn.jsValidator.defaults = {
            '_regular': $.fn._regular,
            '_length': $.fn._length,
            '_repeat': $.fn._repeat,
            '_integer': $.fn._integer,
            '_nochange': $.fn._nochange,
            '_nochangead': $.fn._nochangead
        };

    }(window.jQuery);

/****************************************************************************/
/*	dropdown control init													*/
/****************************************************************************/
$(document).ready(function() {
    try {
        var processTime = $("#processtime").attr("time");
        $("#processtime").text("处理时间： {0} 秒".format(processTime.toString()));
    } catch (e) {
        console.log(e);
    }

    if ($(".page-top-logo").attr("sample") == "true") {
        $(".page-top-logo").find("img").attr("src", "/lang/images/sample/inp-top-logo.gif");
    }

    $.ajaxSetup({
        cache: false
    });
    $(".dropdown").each(function() {
        $(this).find("a").live("click", function(event) {
            ($(event.delegateTarget).find(".dropdown-input")).val($(this).text()).attr("trans", $(this).attr("option"));
            ($(event.delegateTarget).find(".dropdown-span")).text($(this).text()).attr("trans", $(this).attr("option"));
        });
    });
});

/****************************************************************************/
/*	byte format																*/
/****************************************************************************/
formatFileSize = function(bytes) {
    if (typeof bytes !== 'number') {
        return 'unknown size';
    }
    if (bytes >= 1000000000) {
        return (bytes / 1000000000).toFixed(2) + 'GB';
    }
    if (bytes >= 1000000) {
        return (bytes / 1000000).toFixed(2) + 'MB';
    }
    return (bytes / 1000).toFixed(2) + 'KB';
}