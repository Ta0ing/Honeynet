function init(e) {
var t, n, a;
for (t = [ "uid", "domain", "password", "dynamicPwd", "verifyCode" ], n = 0; n < t.length; n++) if (a = document.getElementById(t[n]) || document.getElementsByName(t[n])[0], 
null != a) {
if ("text" == a.type && (a.onfocus = function(e) {
var t = e || window.event, n = t.target || t.srcElement;
if (jQ(n).addClass("focus"), "verifyCode" == this.name) {
var a = document.getElementById("vcImageTR");
a && (a.style.display = "");
}
}, a.onblur = function(e) {
var t = e || window.event, n = t.target || t.srcElement;
jQ(n).removeClass("focus");
}), ("uid" == a.name || "password" == a.name) && (a.onkeyup = function() {
uidPasswordChanged();
}), !hasDefaultValue(a)) {
var i = getCookie(a.name);
i && (a.value = i);
}
a.name == e && a.focus();
}
initCommon(), uidPasswordChanged();
}

function hasDefaultValue(e) {
return !e.getAttribute("ignoreIDV") && e.defaultValue || e.getAttribute("dvalue");
}

function switchLoginForm(e) {
for (var t = [ "userLoginTab", "adminLoginTab" ], n = 0; n < t.length; n++) {
var a = document.getElementById(t[n]);
a && (a.id == e ? a.className = "active" :a.className = "inactive");
}
"adminLoginTab" == e ? (xt5Login && jQ("#triangle").css("left", "98px"), jQ("#verifyCodeRow").css("display", "none")) :(xt5Login && jQ("#triangle").css("left", "20px"), 
jQ("#verifyCodeRow").css("display", ""));
}

function initBackground() {
if (xt5Login) {
var e = Math.round(2 * Math.random());
backgroundDiv = jQ("#bg"), jQ("#bg").attr("class", "").addClass("bg" + e), jQ("#blur").attr("class", "").addClass("blur" + e);
} else {
var e = Math.round(Math.random());
backgroundDiv = jQ(".MainBg"), 1 == e && (jQ(".Main").removeClass("Main").addClass("Main1"), jQ(".MainBg").removeClass("MainBg").addClass("MainBg1"), 
backgroundDiv = jQ(".MainBg1")), frostedGlass();
}
}

function initSubmitButtonStyle() {
var e = new RegExp("^Button-.*\\s");
jQ("button[name='action:login']").bind("mouseover", function() {
this.className = "Button-hover " + this.className;
}), jQ("button[name='action:login']").bind("mouseout", function() {
this.className = this.className.replace(e, "");
}), jQ("button[name='action:login']").bind("click", function() {
this.className = this.className.replace(e, ""), this.className = "Button-click " + this.className;
});
}

function frostedGlass() {
!jQ.browser.msie || "6.0" != jQ.browser.version && "7.0" != jQ.browser.version && "8.0" != jQ.browser.version ? jQ(".MainR").blurjs({
source:".MainBg,.MainBg1",
radius:10,
cacheExpired:3e4,
overlay:"rgba(244,244,244,0.5)",
cache:!1
}) :(jQ(".MainR").css({
background:"rgb(244,244,244)",
border:"1px solid black;"
}), jQ(".inpCode,.inpPW,.inpDomain,.inpUser").css({
border:"1px solid #B8B8B8"
}));
}

function initIndex() {
initBackground(), initSubmitButtonStyle();
}

function initXTO(e) {
var t;
t = [ "uid", "domain", "password", "verifyCode" ], initXTOInput(t, e), (document.getElementById("vcImage") || {}).onclick = function() {
this.src = this.src.replace(/&rand=[\w\-\.]+/, "&rand=" + Math.random());
}, (document.getElementById("vcHint") || {}).href = "javascript:document.getElementById('vcImage').onclick();", changeIndexPage2(), 
initEventDelegate(), showErrorMsg(), initWeather(), frostedGlassXTO(), 0 == jQ("#faceVal,#dynamicPasswordLogin").length && jQ(".TLine").hide();
}

function changeIndexPage2() {
function e(e) {
function t(e) {
return e && e.length > 0;
}
function n(e) {
return void 0 === e;
}
if (e.indexPageData2) {
var a = e.indexPageData2.real_resource;
if (!a) return !1;
if (a.facade_custom) {
var i = a.facade_custom.logo, o = a.facade_custom.background, l = a.facade_custom.favor_title, r = a.facade_custom.favor;
if (t(i)) {
var c = s + "func=lp:getImg&org_id=" + e.indexPageData2.org + "&img_id=" + i[0];
jQ(".Logo").css("background-image", "url(" + c + ")");
}
if (t(o)) {
var d = s + "func=lp:getImg&org_id=" + e.indexPageData2.org + "&img_id=" + o[0];
jQ(".MainContent").css("background-image", "url(" + d + ")");
var u = "progid:DXImageTransform.Microsoft.AlphaImageLoader(src='" + d + "', sizingMethod='scale')";
jQ(".MainContent").css("filter", u);
}
if (t(r)) {
var r = s + "func=lp:getImg&org_id=" + e.indexPageData2.org + "&img_id=" + r[0] + "&t=" + new Date().getTime();
jQ("#favor-image").attr("href", r);
}
n(l) || (document.title = l), t(a.facade_custom.background_color) && jQ("#mainContainer").css("background-color", a.facade_custom.background_color), 
t(a.facade_custom.submit_button_color) && jQ("#logArea .inptr .button").css("background", a.facade_custom.submit_button_color), 
t(a.facade_custom.submit_button_font_color) && jQ("#logArea .inptr .button").css("color", a.facade_custom.submit_button_font_color), 
n(a.facade_custom.copyright_text) || jQ(".copyright a").html(a.facade_custom.copyright_text), n(a.facade_custom.copyright_link) || jQ(".copyright a").attr("href", a.facade_custom.copyright_link), 
t(a.facade_custom.slogan_color) && jQ("#slogan").css("color", a.facade_custom.slogan_color), n(a.facade_custom.slogan_text) || jQ("#slogan").html(a.facade_custom.slogan_text), 
t(a.facade_custom.slogan_fontsize) && jQ("#slogan").css("font-size", a.facade_custom.slogan_fontsize + "px");
}
if (a.detail_custom) {
if (a.detail_custom.iac_enable && 1 == a.detail_custom.iac_enable) {
if (t(a.detail_custom.iac)) {
var m = s + "func=lp:getImg&org_id=" + e.indexPageData2.org + "&img_id=" + a.detail_custom.iac[0];
jQ(".wechat .icon").css("background", "url(" + m + ") no-repeat center");
}
n(a.detail_custom.iac_text) || jQ(".wechat .word a").html(a.detail_custom.iac_text), n(a.detail_custom.iac_link) || jQ(".wechat .word a").attr("href", a.detail_custom.iac_link);
} else jQ(".wechat").hide(), jQ(".bottom .info").css("border-left", "0px");
if (!n(a.detail_custom.top_link)) {
for (var g = a.detail_custom.top_link, f = "", h = 0; h < g.length; h++) f += '<a target="_blank" href="' + g[h].top_link_href + '">' + g[h].top_link_content + "</a>";
jQ(".Links").html(f);
}
if (!n(a.detail_custom.telephone)) {
for (var p = a.detail_custom.telephone, v = "", h = 0; h < p.length; h++) v += '<div class="item">' + p[h].telephone_content + ":" + p[h].telephone_num + "</div>";
jQ(".phone").html(v);
}
if (!n(a.detail_custom.business_info)) {
var j = a.detail_custom.business_info;
j[0] && j[0].info_href ? jQ(".info .title").html(j[0].info_href) :jQ(".info .title").html(""), j[0] && j[0].info_content ? jQ(".info .text").html(j[0].info_content) :jQ(".info .text").html(""), 
n(j[0]) && (jQ(".info").hide(), "none" == jQ(".wechat").css("display") && "none" == jQ(".info").css("display") || jQ(".slogan").css("border-left", "1px solid #cccccc"));
}
}
a.tool_custom && a.tool_custom.business_notes && 1 != a.tool_custom.business_notes.enable && (jQ(".slogan").hide(), jQ(".bottom .info").css("border-right", "0px")), 
a.facade_custom && t(a.facade_custom.default_lang) && location.href.toString().indexOf("cus=1") < 0 && (setCookie("locale", a.facade_custom.default_lang), 
window.location.href.toString().indexOf("?") > -1 ? window.location = location + "&cus=1" :window.location = location + "?cus=1");
}
}
function t() {
for (var e = document.cookie.split(/\s*;\s*/), t = 0; t < e.length; t++) {
var i = e[t].split("=");
if (a == i[0]) return n(i[1]);
}
return {};
}
function n(e) {
for (var t = decodeURIComponent(e), n = t.split(":"), a = {}, i = 0; i < n.length; i++) {
var o = n[i].split("=");
a[o[0]] = o[1];
}
return a;
}
var a = "Coremail.IndexPageData", i = t(), o = (i.ts, location.protocol + "//" + location.host + "/coremail/index_data.jsp"), s = location.protocol + "//" + location.host + "/coremail/s?";
jQ.ajax({
async:!1,
type:"GET",
url:o,
dataType:"JSON",
success:function(t) {
e(t);
},
error:function(e) {
alert("error");
}
}), jQ("#links>*:last").length > 0 && jQ("#links>*:last").html(jQ("#links>*:last").html().replace("|", ""));
}

function initXTOInput(e, t) {
for (var n, a = 0; a < e.length; a++) if (n = document.getElementById(e[a]) || document.getElementsByName(e[a])[0], null != n) {
if (!hasDefaultValue(n)) {
var i = getCookie(n.name);
i && ("domain" == n.name ? changeDomain(i) :n.value = i);
}
n.name == t && n.focus();
}
}

function initXT5(e) {
xt5Login = !0, initWeather(), adjustFrame(), adjustIcon(e), initLoginTab(), initLogin();
}

function initXT3(e) {
initLogin(e);
}

function initLogin(e) {
var t;
t = [ "uid", "domain", "password", "verifyCode" ], initInputCss(t, e), changeIndexPage(), 0 == jQ("#faceVal,#dynamicPasswordLogin").length && jQ(".TLine").hide(), 
jQ(document).mousedown(function(e) {
e = e || window.event || this.parentWindow.event;
for (var t = e.srcElement || e.target; void 0 !== t && null !== t; ) "language" != t.className && fadeOutElement(jQ(".localePanel").get(0)), 
"inpDomain" != t.id && fadeOutElement(jQ(".domainPanel").get(0)), "faceSelectOption" != t.id && fadeOutElement(jQ(".facePanel").get(0)), 
t = t.parentNode;
}), initCommon(), adjustHeight(), window.onresize = adjustHeight;
}

function getQuery(e) {
var t = new RegExp("(^|&)" + e + "=([^&]*)(&|$)"), n = window.location.search.substr(1).match(t);
return null != n ? unescape(n[2]) :null;
}

function changeIndexPage() {
function e(e) {
function n(t) {
try {
if ("broadsideColor" === t) return e.indexPageData.loginPageData.background.broadsideColor;
if ("images" === t) return e.indexPageData.loginPageData.background.backgroundImage.images;
if ("isRotator" === t) return e.indexPageData.loginPageData.background.backgroundImage.isRotator;
if ("links" === t) return e.indexPageData.loginPageData.links;
if ("copyright" === t) return e.indexPageData.loginPageData.copyright;
if ("defaultLocale" === t) return e.indexPageData.loginPageData.loginForm.languages.defaultLocale;
if ("domains" === t) return e.indexPageData.loginPageData.loginForm.domains;
if ("submitButtonStyle" === t) return e.indexPageData.loginPageData.loginForm.submitButtonStyle;
if ("extraFunctions" === t) return e.indexPageData.loginPageData.loginForm.extraFunctions;
if ("loginModeForm" === t) return e.indexPageData.loginPageData.loginModeForm.loginMode;
} catch (n) {
return !1;
}
}
if (e.background) {
var a = s + "?key=background&ts=" + e.ts;
backgroundDiv.css("background-image", "url(" + a + ")");
} else if (e.indexPageData) {
if ((temp = n("submitButtonStyle")) ? jQ("#login_button").css({
border:"1px solid " + temp.backgroundColor,
"background-color":temp.backgroundColor,
color:temp.textColor
}) :initSubmitButtonStyle(), (temp = n("defaultLocale")) && location.href.toString().indexOf("cus=1") < 0 && (setCookie("locale", temp), 
window.location.href.toString().indexOf("?") > -1 ? window.location = location + "&cus=1" :window.location = location + "?cus=1"), 
(temp = n("copyright")) && jQ("#copyright").html(t(temp)), (temp = n("domains")) && (temp.defaultDomain && (jQ("#domainVal").html(temp.defaultDomain), 
jQ("#domain").val(temp.defaultDomain)), temp.isDomainSelectable === !1 && jQ("#inpDomain").remove()), temp = n("images")) {
if (getQuery("temp_login_page_data")) var i = "&temp_login_page_data=" + getQuery("temp_login_page_data");
if (n("isRotator")) {
var o = Math.floor(Math.random() * temp.length + 1) - 1, a = s + "?ts=" + e.ts + "&key=" + temp[o] + (i ? i :"");
jQ("#MainBg").css("background-image", "url(" + a + ")");
} else {
var a = s + "?ts=" + e.ts + "&key=" + temp[0] + (i ? i :"");
jQ("#MainBg").css("background-image", "url(" + a + ")");
}
xt5Login || frostedGlass();
} else initBackground();
if ((temp = n("broadsideColor")) && jQ("#Main").css("background-color", temp), temp = n("links")) for (var l = jQ("#links > *").detach(), r = 0; r < temp.length; r++) if (node = l.filter("#" + temp[r].id), 
node.length > 0) jQ("#links").append(node[0]); else {
var c = temp[r].resource[login_page_custom.locale] ? temp[r].resource[login_page_custom.locale] :temp[r].id;
jQ("#links").append('<span id="' + temp[r].id + '"><a target="_blank" href="' + temp[r].url + '">' + c + "</a> | </span>");
}
(temp = n("extraFunctions")) && (temp.enableUseSSL === !1 && jQ("#enableUseSSLLabel").remove(), temp.enableRememberMe === !1 && jQ("#saveUsernameLabel").remove(), 
temp.displayNewFeatures === !1 && jQ("#newFeatures").remove());
} else initIndex();
}
function t(e) {
return e.replace(/&/gi, "&amp;").replace(/</gi, "&lt;").replace(/>/gi, "&gt;").replace(/\"/gi, "&quot;").replace(/'/gi, "&#039;");
}
function n() {
for (var e = document.cookie.split(/\s*;\s*/), t = 0; t < e.length; t++) {
var n = e[t].split("=");
if (i == n[0]) return a(n[1]);
}
return {};
}
function a(e) {
for (var t = decodeURIComponent(e), n = t.split(":"), a = {}, i = 0; i < n.length; i++) {
var o = n[i].split("=");
a[o[0]] = o[1];
}
return a;
}
var i = "Coremail.IndexPageData", o = n(), s = (o.ts, location.protocol + "//" + location.host + "/coremail/index_data.jsp");
if (getQuery("temp_login_page_data")) {
var l;
l = s.indexOf("?") > -1 ? "&temp_login_page_data=" + getQuery("temp_login_page_data") :"?temp_login_page_data=" + getQuery("temp_login_page_data");
}
jQ.ajax({
async:!1,
type:"GET",
url:s + (l ? l :""),
dataType:"JSON",
success:function(t) {
e(t);
},
error:function(e) {
initIndex();
}
}), jQ("#links>*:last").length > 0 && jQ("#links>*:last").html(jQ("#links>*:last").html().replace("|", ""));
}

function initInputCss(e, t) {
for (var n, a = 0; a < e.length; a++) if (n = document.getElementById(e[a]) || document.getElementsByName(e[a])[0], null != n) {
if (("text" == n.type || "password" == n.type) && (xt5Login && "uid" == n.id && "" == n.value && (n.value = n.getAttribute("defaultValue")), 
xt5Login && "password" == n.type && "" == n.value && (n.type = "text", n.value = n.getAttribute("defaultValue")), n.onmouseover = function() {
document.activeElement.name != this.name && (jQ(this).addClass("inpOver"), jQ(this).removeClass("inpFocus"));
}, n.onmouseout = function() {
document.activeElement.name != this.name && (jQ(this).removeClass("inpOver"), jQ(this).removeClass("inpFocus"));
}, n.onblur = function() {
jQ(this).removeClass("inpOver"), jQ(this).removeClass("inpFocus"), jQ(this).removeClass("inped"), jQ.trim(this.value).length > 0 && jQ(this).addClass("inped"), 
"text" == this.type && "password" != this.type && (this.value = jQ.trim(this.value)), "uid" == this.id && (this.value = changePoint(this.value), 
xt5Login && "" == this.value && (this.value = jQ(this).attr("defaultValue"))), xt5Login && "" == this.value && "password" == this.id && (this.type = "text", 
this.value = jQ(this).attr("defaultValue"));
}, n.onfocus = function() {
jQ(this).removeClass("inpOver"), jQ(this).addClass("inpFocus"), xt5Login && "uid" == this.id && this.value == jQ(this).attr("defaultValue") && (this.value = ""), 
xt5Login && "password" == this.id && this.value == jQ(this).attr("defaultValue") && (this.type = "password", this.value = "");
}), !hasDefaultValue(n)) {
var i = getCookie(n.name);
i && ("domain" == n.name ? changeDomain(i) :n.value = i);
}
if (t && n.name == t && n.focus(), jQ.trim(n.value).length > 0) {
if (xt5Login && "uid" == n.id && n.value == n.getAttribute("defaultValue")) continue;
if (xt5Login && "password" == n.id && n.value == n.getAttribute("defaultValue")) continue;
jQ(n).addClass("inped");
}
}
}

function changePoint(e) {
var t = unescape("\\u3002".replace(/\\u/gi, "%u")), n = e.indexOf("@");
if (-1 != n) {
var a = jQ.trim(e.substring(0, n)), i = e.substring(n + 1, e.length);
if (-1 == i.indexOf(".") && -1 != i.indexOf(t)) {
var o = i.lastIndexOf(t);
i = i.substring(0, o) + "." + i.substring(o + 1, i.length);
}
return a + "@" + i;
}
return e;
}

function initCommon() {
(document.getElementById("vcImage") || {}).onclick = function() {
this.src = this.src.replace(/&rand=[\w\-\.]+/, "&rand=" + Math.random());
}, (document.getElementById("vcHint") || {}).href = "javascript:document.getElementById('vcImage').onclick();";
var e = document.getElementById("homepage");
e && (document.all ? (e.href = "javascript:void(0);", e.style.behavior = "url(#default#homepage)", e.onclick = function() {
this.setHomePage(document.location);
}) :e.style.display = "none");
}

function adjustHeight() {
var e = jQ("div[class='Head']").height(), t = 0 == jQ("div[class='Main']").length ? jQ("div[class='Main1']").height() :jQ("div[class='Main']").height(), n = jQ("div[class='MainM']").height(), a = jQ("div[class='footer']").height(), i = parseInt(jQ(".footer").css("padding-bottom")), o = e + t + n + a + i, s = (getClientSize()[1] - o) / 2;
s > 0 ? jQ("body").css("padding-top", s) :jQ("body").css("padding-top", "0px"), adjustElPos(jQ("localeArea").get(0), jQ("#language").get(0), 5, -4), 
adjustElPos(jQ("#domainPanel").get(0), jQ("#inpDomain").get(0));
}

function getCookie(e) {
var t = new RegExp("(^| )" + e + "=([^;]*)(;|$)", "gi"), n = t.exec(document.cookie);
return unescape((n || [])[2] || "");
}

function setCookie(e, t) {
document.cookie = e + "=" + escape(t) + ";expires=" + new Date(1990, 1, 1).toGMTString(), document.cookie = e + "=" + escape(t) + ";path=/;expires=" + new Date(2099, 12, 31).toGMTString();
}

function changeLocale(e, t) {
setCookie("locale", e), window.location.href.toString().indexOf("cus=1") > -1 ? window.location = location :window.location.href.toString().indexOf("?") > -1 ? window.location = location + "&cus=1" :window.location = location + "?cus=1";
}

function changeDomain(e) {
jQ("input[name='domain']").val(e), jQ("#domainVal").html(e), fadeOutElement(jQ(".domainPanel").get(0));
}

function changeFace(e, t, n, a) {
var i = jQ("#supportXT5Msg");
i.css("display", "none"), "XT5" == e && !supportXT5Browser && n && n.length > 0 && a && a.length > 0 && (i.css("display", "block"), 
e = n, t = a), "auto" != e || "XT5" != jQ("input[name='face']").val() || supportXT5Browser || (e = "XT3"), jQ("input[name='face']").val(e), 
xt5Login && jQ("input[name='faceName']").val(t), jQ("#faceVal").html(t), fadeOutElement(jQ(".facePanel").get(0));
}

function fadeInElement(e, t, n, a) {
jQ(e).length > 0 && jQ(e).is(":hidden") && (jQ(e).fadeIn(), adjustElPos(e, t, n, a));
}

function fadeOutElement(e) {
jQ(e).length > 0 && jQ(e).is(":visible") && jQ(e).fadeOut();
}

function displayFacePanel(e) {
fadeInElement(jQ(".facePanel").get(0), e);
var t = 0 - jQ(".facePanel").height() - jQ(e).height();
adjustElPos(jQ(".facePanel").get(0), e, t - 5, -12);
}

function adjustElPos(e, t, n, a) {
n = n || 0, a = a || 0, jQ(e).length > 0 && jQ(e).is(":visible") && (xt5Login ? (jQ(e).hasClass("facePanel") ? jQ(e).css("top", (t ? jQ(t).offset().top + jQ(t).height() :localShowDomainList ? 0 :-41) + n + "px") :jQ(e).css("top", (t ? jQ(t).offset().top + jQ(t).height() :0) + n + "px"), 
jQ(e).css("left", (t ? jQ(t).offset().left :0) + a + "px"), jQ(e).css("z-index", 999)) :(jQ(e).css("top", jQ(t).offset().top + jQ(t).height() + n + "px"), 
jQ(e).css("left", jQ(t).offset().left + a + "px")));
}

function loginSubmit(e, t, n) {
function a(e) {
if ("active" == (document.getElementById("adminLoginTab") || {}).className) {
var t = e.substring(0, e.indexOf(n));
e = t + "/webadmin/index.jsp?submit=true";
}
return e;
}
var i = jQ("input[name='face']").val();
"XT5" != i || supportXT5Browser || jQ("input[name='face']").val("XT3");
var o = document.getElementById("uid"), s = document.getElementById("password");
xt5Login && o.value == o.getAttribute("defaultValue") && (o.value = ""), xt5Login && s.value == s.getAttribute("defaultValue") && (s.value = ""), 
document.getElementById("uid").value = changePoint(o.value), n = n || "/coremail", (document.getElementById("saveUsername") || {
checked:!0
}).checked && (setCookie("uid", document.getElementById("uid").value), document.getElementById("domain") && setCookie("domain", jQ("#domain").val())), 
document.getElementById("locale") && setCookie("locale", document.getElementById("locale").value);
var l = document.getElementById("speedtest");
if (l) {
var r = getCookie("netSpeedServerHost");
if (r) {
var c = location.protocol + "//" + r;
location.pathname && (c += location.pathname), e.action = c;
}
}
var d = (document.getElementsByName("useSSL")[0] || {}).checked;
if ("boolean" == typeof d) {
var u = d ? "http://" :"https://", m = d ? "https://" :"http://";
e.action = function(e) {
if ("/" == e.charAt(0)) return a(location.protocol + "//" != m ? m + location.hostname + e :e);
if (e.substring(0, u.length).toLowerCase() == u) {
var t = e.indexOf("/", u.length), n = e.lastIndexOf(":", t);
return a(t > 0 && n > 0 && e.substring(n + 1, t).match(/^\d+$/) ? m + e.substring(u.length, n) + e.substring(t) :m + e.substring(u.length));
}
return a(e);
}(e.action);
}
return (document.getElementById("face_classic_cgi") || {}).selected ? (document.all && (t.returnValue = !1), document.getElementById("classic_cgi_form").elements.user.value = e.elements.uid.value, 
document.getElementById("classic_cgi_form").elements.pass.value = e.elements.password.value, document.getElementById("classic_cgi_form").submit(), 
!1) :!0;
}

function recoverPwd(e) {
e.href += "?uid=" + document.getElementById("uid").value;
}

function bookmarkMe() {
try {
window.external.AddFavorite(location.href, document.title);
} catch (e) {
try {
window.sidebar.addPanel(location.href, document.title, "");
} catch (e) {
alert(markme_msg);
}
}
}

function uidPasswordChanged() {
var e = document.getElementById("verifyCellCode"), t = document.getElementById("sendVerifyCellCodeLink");
if (null != t && null != t) {
for (var n = [ "uid", "password" ], a = 0, i = n.length; i > a; a++) {
var o = document.getElementById(n[a]) || document.getElementsByName(n[a])[0];
if ("" == o.value) return e.disabled = !0, void (t.onclick = null);
}
e.disabled = !1, t.onclick = submitSendVerifyCellCode;
}
}

function submitSendVerifyCellCode() {
dialog && dialog.close();
var e = document.getElementById("loginForm"), t = document.getElementsByName("action:sendVerifyCellCode")[0];
t && (t.disabled = !1, document.getElementById("verifyCellCode").value = "", e.submit());
}

function initDialog(e, t, n) {
if ("verifyCellCode" == e || "dynamicPwd" == e) {
var a = e + "_d", i = t + "<p><input type='text' class='inpKey' name='" + a + "'/>";
"verifyCellCode" == e && (i = i + "<p><a href='javascript:submitSendVerifyCellCode();' style='padding-left: 10px;'>" + jQ("input[name='action:sendVerifyCellCode']").val() + "</a>");
var o = {
showMask:!0,
body:i,
button:n,
actions:[ function() {
jQ("input[name='" + e + "']").val(jQ("input[name='" + a + "']").val()), jQ("button[name='action:login']").click();
} ]
};
dialog = new Dialog(o), initInputCss([ a ], a);
}
}

function Dialog(e) {
function t(e) {
for (var t = new Array(), n = 0; n < e.length; n++) {
var a = e[n].split(":"), i = a.length > 1 ? a[1] :a[0];
t.push("<button type='button' class='winBotton'>" + i + "</button>");
}
return t;
}
function n(e, t) {
return (e || function() {})(t || g.getElementsByTagName("FORM")[0] || g, c);
}
function a(e) {
n(j[e]) !== !1 && i();
}
function i() {
d == u.parentNode && d.removeChild(u);
var e = (d.lastChild || {}).previousSibling;
e ? e.style.zIndex = 999 :(d.style.display = "none", u.innerHTML = "", d.innerHTML = "");
}
function o() {
if (u.style.cssText = "position:absolute;z-index:999;top:0;left:0;display:none;", null == d.firstChild) {
var e = h ? p :0, t = h ? 100 * p :0;
getIEVersionLt10() ? d.innerHTML = '<div style="position:absolute;z-index:998;top:0;left:0;width:100%;height:100%;background-color:#b8b8b8;filter:alpha(opacity=' + t + ');"><iframe style="position:absolute;z-index:-1;top:0;left:0;width:100%;height:100%;filter:alpha(opacity=0)" frameborder=0 src="javascript:\'\'"></iframe></div>' :d.innerHTML = '<div style="position:absolute;z-index:998;top:0;left:0;width:100%;height:100%;background-color:#b8b8b8;opacity:' + e + ';"></div>';
}
var n = d.lastChild.previousSibling;
n && (n.style.zIndex = 997), d.insertBefore(u, d.lastChild);
}
function s() {
d.style.display = "", u.style.display = "", l(), window.onresize = function() {
l();
};
}
function l() {
var e = u.lastChild, t = e.offsetWidth, n = e.offsetHeight, a = getClientSize(), i = a[0], o = a[1];
u.style.left = Math.floor(Math.max(0, i - t) / 2) + "px", u.style.top = Math.floor(Math.max(0, o - n) / 2) + "px", document.body.screenTop = document.body.screenLeft = 0;
}
function r() {
o(), s();
}
e = e || {};
var c = this, d = document.getElementById("#dialogContainerPanel");
d || (d = document.createElement("div"), d.id = "dialogContainerPanel", document.body.appendChild(d));
var u = document.createElement("div"), m = "<div id='winFrame'><div class='winHead'><div class='closeBtn' id='closeBtn'><b class='ico icoClose'></b></div></div><div class='winMain'>$BODY$</div><div class='winFoot'>$BUTTONS$</div></div>";
u.innerHTML = (document.all ? '<iframe style="position:absolute;z-index:-1;top:0;left:0;width:0;height:0;);" frameborder=0 src="javascript:\'\'"></iframe>' :"") + '<div style="position:absolute;top:0;left:0;">' + m + "</div>";
for (var g, f, h = void 0 === e.showMask ? !1 :e.showMask, p = void 0 === e.maskValue ? .5 :e.maskValue, v = void 0 === e.button ? "OK" :e.button, j = e.actions || [ e.action ], Q = 0, b = u.getElementsByTagName("*"); b[Q]; Q++) if (1 == b[Q].childNodes.length && 3 == b[Q].firstChild.nodeType) {
var y = b[Q].firstChild.data;
b[Q].removeChild(b[Q].firstChild), "$BODY$" == y ? g = b[Q] :"$BUTTONS$" == y && (f = b[Q]);
}
g.innerHTML = e.body, f.innerHTML = t(v.split("_")).join(" ");
for (var _ = jQ(f).children("button"), Q = 0; Q < _.length; Q++) _[Q].onclick = function(e) {
return function() {
a(e);
};
}(Q);
r(), jQ("#closeBtn").click(i), c.close = i;
}

function getClientSize() {
return [ jQ(window).width(), jQ(window).height() ];
}

function getIEVersionLt10() {
var e = navigator.userAgent;
return -1 != e.indexOf("MSIE") && -1 == e.indexOf("MSIE 10") ? !0 :!1;
}

function adjustFrame() {
var e = document.documentElement.clientHeight, t = document.documentElement.clientWidth, n = $("version").clientWidth, a = $("version").clientHeight;
jQ("#bg").css("height", e + "px"), jQ("#version").css("left", .5 * (.64 * t - n) + "px").css("top", .3 * (e - a) + "px"), 
jQ("#blur").css("height", e + "px").css("width", t + "px"), jQ(".MainR:first").css("top", .44 * (e - 337) + "px").css("left", .36 * e * .18 + "px"), 
jQ("#weather").css("left", .36 * e * .18 + "px");
}

function adjustIcon(e) {
localShowDomainList = e, e ? (jQ(".iconlock:first").css("top", "159px"), jQ(".icontie:first, .iconFace:first").css("top", "198px")) :(jQ(".iconlock:first").css("top", "117px"), 
jQ(".icontie:first, .iconFace:first").css("top", "159px"));
}

function initLoginTab() {
jQ("#userLoginTab, #adminLoginTab").click(function() {
switchLoginForm(this.id);
var e = [];
e.push("#faceSelectOption"), e.push(".inplist"), toggle(this.id, e);
});
}

function toggle(e, t) {
if ("userLoginTab" == e) for (var n = 0; n < t.length; n++) {
var a = t[n];
"#" == a.substring(0, 1) ? jQ(a).show() :jQ(a).each(function() {
this.style.display = "block";
});
} else if ("adminLoginTab" == e) for (var n = 0; n < t.length; n++) {
var a = t[n];
"#" == a.substring(0, 1) ? jQ(a).hide() :jQ(a).each(function() {
this.style.display = "none";
});
}
}

function initWeather() {
jQ.ajax({
type:"GET",
url:"/coremail/XT5/jsp/mail.jsp",
data:{
func:"getWeather"
},
dataType:"json",
success:function(e) {
if ("S_OK" == e.code) {
var t = e["var"].weather.status;
if ("success" == t) {
var n = e["var"].weather.results[0], a = n.currentCity, i = n.weather_data[0], o = i.temperature, s = i.dayPictureUrl, l = i.nightPictureUrl;
jQ("#weather #weatherInfo").html(o + " | " + a);
new Date().getHours() >= 18 ? weatherPic = l.substring(l.lastIndexOf("/") + 1, l.lastIndexOf(".")) :weatherPic = s.substring(s.lastIndexOf("/") + 1, s.lastIndexOf(".")), 
jQ("#mainRight #weatherPic").attr("class", "iconfont commonIconfont icon" + weatherPic), jQ("#mainRight #weather").show();
} else jQ("#mainRight #weather").hide();
}
},
error:function(e) {
jQ("#mainRight #weather").hide();
}
});
}

function initEventDelegate() {
jQ(".favourite").click(function() {
try {
window.external.AddFavorite(location.href, document.title), jQ(this).addClass("favourite");
} catch (e) {
try {
window.sidebar.addPanel(location.href, document.title, ""), jQ(this).addClass("favourite");
} catch (e) {
alert(jQ(this).data("msg"));
}
}
}), jQ("#domainMenu").dropdown({
callback:function() {
jQ("label.domainTxt").html(jQ(this).data("domain")), jQ("#domain").val(jQ(this).data("domain"));
}
}), jQ("#faceMenu").dropdown({
callback:function() {
jQ("label.faceTxt").html(jQ(this).data("facetxt")), jQ("#faceName").val(jQ(this).data("facetxt")), jQ("#face").val(jQ(this).data("face"));
}
}), jQ(".refresh").click(function() {
var e = jQ(".securityCode img"), t = e.attr("src"), n = parseInt(1e8 * Math.random());
t = t.replace(/rand=-?\d*/, "rand=" + n), e.attr("src", t);
}), jQ("div.inputArea").on("focus", "input", function() {
jQ(this).prev().addClass("actived");
}).on("err.focus", "input", function() {
jQ(this).trigger("focus"), jQ(this).prev().addClass("error"), "uid" == jQ(this).attr("id") && jQ(".forPassword label").addClass("error");
}).on("blur", "input", function() {
jQ(this).prev().removeClass("error actived");
});
}

function frostedGlassXTO() {
var e = Math.round(3 * Math.random());
jQ(".MainContent").addClass("MainContent-" + e), jQ(".aside-blur").blurjs({
source:".MainContent",
radius:50,
cacheExpired:3e4,
cache:!1
});
}

function showErrorMsg() {
var e = SYS_CONST.focusEleName, t = SYS_CONST.empty_error, n = SYS_CONST.error_other, a = jQ('[name="' + e + '"]');
if ("verifyCellCode" == e) {
jQ("#overlay").show();
var i = $("#dialog");
i.on("click", "a", function() {
jQ("#resendCellCode").prop("disabled", !1), jQ("button.submit").trigger("click");
}), i.dialog({
closeText:'<i class="iconfont iconerror"></i>',
width:386,
buttons:[ {
text:SYS_CONST.confirmText,
click:function() {
jQ(this).dialog("close"), jQ("#verifyCellCode").val($("#cellCode").val().toUpperCase()), jQ("button.submit").trigger("click");
}
}, {
text:SYS_CONST.cancelText,
click:function() {
jQ(this).dialog("close");
}
} ],
close:function() {
jQ("#overlay").hide();
}
});
} else if (0 != a.length) if (t || n) {
a.trigger("err.focus");
var o = a.parents(".u-form-item");
if (jQ.support.leadingWhitespace) o.addClass("animated shake"); else {
for (var s = o.offset() - o.width(), l = s.left, r = s.top, c = 1; 5 >= c; c++) c % 2 == 0 ? o.animate({
left:"+6px"
}, 80) :o.animate({
left:"-6px"
}, 80);
o.animate({
left:0
}, 150), o.offset({
top:r,
left:l
});
}
} else a.trigger("focus");
}

var jQ = jQuery.noConflict(), $ = function(e) {
return document.getElementById(e);
}, xt5Login = !1, localShowDomainList = !1, dialog, supportXT5Browser = jQ.browser.msie && ("7.0" == jQ.browser.version || "6.0" == jQ.browser.version || document.documentMode < 8) ? !1 :!0;

window.SYS_CONST = window.SYS_CONST || {};
//# sourceMappingURL=login.js.map