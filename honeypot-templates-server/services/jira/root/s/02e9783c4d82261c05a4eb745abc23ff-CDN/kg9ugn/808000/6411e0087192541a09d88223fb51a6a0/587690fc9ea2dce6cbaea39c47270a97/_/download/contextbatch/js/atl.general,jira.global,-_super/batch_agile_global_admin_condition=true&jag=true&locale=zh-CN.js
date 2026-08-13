WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'com.atlassian.analytics.analytics-client:js-events', location = 'js/store-1.3.1.js' */
(function(){var l={},h=window,k=h.document,c="localStorage",n="globalStorage",d="__storejs__",g;l.disabled=false;l.set=function(e,o){};l.get=function(e){};l.remove=function(e){};l.clear=function(){};l.transact=function(e,o){var p=l.get(e);if(typeof p=="undefined"){p={}}o(p);l.set(e,p)};l.serialize=function(e){return JSON.stringify(e)};l.deserialize=function(e){if(typeof e!="string"){return undefined}return JSON.parse(e)};function b(){try{return(c in h&&h[c])}catch(e){return false}}function m(){try{return(n in h&&h[n]&&h[n][h.location.hostname])}catch(e){return false}}if(b()){g=h[c];l.set=function(e,o){if(o===undefined){return l.remove(e)}g.setItem(e,l.serialize(o))};l.get=function(e){return l.deserialize(g.getItem(e))};l.remove=function(e){g.removeItem(e)};l.clear=function(){g.clear()}}else{if(m()){g=h[n][h.location.hostname];l.set=function(e,o){if(o===undefined){return l.remove(e)}g[e]=l.serialize(o)};l.get=function(e){return l.deserialize(g[e]&&g[e].value)};l.remove=function(e){delete g[e]};l.clear=function(){for(var e in g){delete g[e]}}}else{if(k.documentElement.addBehavior){var j,f;try{f=new ActiveXObject("htmlfile");f.open();f.write('<script>document.w=window<\/script><iframe src="/favicon.ico"></frame>');f.close();j=f.w.frames[0].document;g=j.createElement("div")}catch(i){g=k.createElement("div");j=k.body}function a(e){return function(){var p=Array.prototype.slice.call(arguments,0);p.unshift(g);j.appendChild(g);g.addBehavior("#default#userData");g.load(c);var o=e.apply(l,p);j.removeChild(g);return o}}l.set=a(function(p,e,o){if(o===undefined){return l.remove(e)}p.setAttribute(e,l.serialize(o));p.save(c)});l.get=a(function(o,e){return l.deserialize(o.getAttribute(e))});l.remove=a(function(o,e){o.removeAttribute(e);o.save(c)});l.clear=a(function(q){var o=q.XMLDocument.documentElement.attributes;q.load(c);for(var p=0,e;e=o[p];p++){q.removeAttribute(e.name)}q.save(c)})}}}try{l.set(d,d);if(l.get(d)!=d){l.disabled=true}l.remove(d)}catch(i){l.disabled=true}if(typeof module!="undefined"){module.exports=l}else{if(typeof define==="function"&&define.amd){define(l)}else{this.store=l}}})();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.analytics.analytics-client:js-events', location = 'js/page-visibility.js' */
define("atlassian/analytics/page-visibility",function(){var a=(document.hidden!==undefined);var b={supported:a,isHidden:function(){return a?document.hidden:false}};return b});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.analytics.analytics-client:js-events', location = 'js/user-activity-xhr-header.js' */
define("atlassian/analytics/user-activity-xhr-header",["atlassian/analytics/page-visibility"],function(c){var e="x-atlassian-mau-ignore";var d=XMLHttpRequest.prototype.send;var a=window.fetch;var b=false;return{install:function(){if(!b&&c.supported){XMLHttpRequest.prototype.send=function(){if(c.isHidden()){this.setRequestHeader(e,c.isHidden())}d.apply(this,arguments)};if(a){window.fetch=function(f,i){var g=i||{};var h=g.headers;if(c.isHidden()){g.headers=(h)?new Headers(h):new Headers();g.headers.set(e,c.isHidden())}return a.call(this,f,g)}}b=true}},uninstall:function(){if(b){XMLHttpRequest.prototype.send=d;if(a){window.fetch=a}}b=false}}});require("atlassian/analytics/user-activity-xhr-header").install();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.analytics.analytics-client:js-events', location = 'js/atlassian-analytics.js' */
(function(d){var q=AJS.$.ajax;var l="atlassian-analytics";var j=typeof AJS.contextPath==="function"?AJS.contextPath():"";var p=null;var m=null;var g=null;var r="xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g,function(z){var y=Math.random()*16|0,x=z=="x"?y:(y&3|8);return x.toString(16)});var t=function(){var x="unknown";if(document.body.id=="jira"){x="jira"}else{if(document.body.id=="com-atlassian-confluence"){x="confluence"}}m=l+"."+x;g=m+".lock"};var f=function(){if(store.get(g)){return false}store.set(g,r);return(store.get(g)===r)};var u=function(){store.set(g,null)};var i=function(){var y=[],A,B,x,z;if(AJS.EventQueue.length==0){return}y=store.get(m)||y;for(x=0,z=AJS.EventQueue.length;x<z;++x){B=AJS.EventQueue[x];if(B.name){A={name:B.name,properties:B.properties,time:B.time||0};y.push(A)}}AJS.EventQueue.length=0;store.set(m,y)};var v=0;var a=function(x){return Math.min(5000*Math.pow(2,x),5*60*1000)};var e=0;var h=function(){var A;function y(){setTimeout(h,a(v=0))}function x(){setTimeout(h,a(++v))}if(!f()){++e;if(e<20){return y()}}else{e=0}try{i();A=store.get(m);if(!A||!A.length){return y()}store.remove(m);if(!o(A)){return y()}var B=new Date().getTime();for(var z=0;z<A.length;z++){if(A[z].time>0){A[z].timeDelta=A[z].time-B}else{A[z].timeDelta=z-A.length}delete A[z].time}p=q({type:"POST",url:j+"/rest/analytics/1.0/publish/bulk",data:JSON.stringify(A),contentType:"application/json",dataType:"json"});p.fail(function(){AJS.EventQueue.concat(A);i();x()});p.done(function(){y()})}finally{u()}};var o=function(z){for(var y=z.length-1;y>=0;y--){var B="";var A=z[y];var x=A.properties;if(typeof A.name==="undefined"){B="you must provide a name for the event."}else{if(typeof x!=="undefined"&&x!==null){if(x.constructor!==Object){B="properties must be an object with key value pairs."}else{b(x)}}}if(B!==""){AJS.log("WARN: Invalid analytics event detected and ignored, "+B+"\nEvent: "+JSON.stringify(A));z.splice(y,1)}}return z.length};var b=function(z){for(var y in z){if(z.hasOwnProperty(y)){var x=z[y];if(c(x)&&k(x)){}else{if(c(x)&&x.toString){z[y]=x.toString()}else{z[y]=""}}}}};function c(x){return typeof x!=="undefined"&&x!==null}function k(x){return typeof x==="number"||typeof x==="string"||typeof x==="boolean"}var n=function(){if(p&&!(p.state()==="resolved"||p.state()==="rejected")){p.abort()}};AJS.EventQueue=AJS.EventQueue||[];var s=Array.prototype.push;AJS.EventQueue.push=function(x){x.time=new Date().getTime();s.call(AJS.EventQueue,x)};AJS.toInit(function(){t();setTimeout(h,500);w()});d(window).unload(function(){n();i()});AJS.Analytics={triggerPrivacyPolicySafeEvent:function(x,y){AJS.log("WARN: 'triggerPrivacyPolicySafeEvent' has been deprecated");AJS.EventQueue.push({name:x,properties:y})}};AJS.bind("analytics",function(x,y){AJS.EventQueue.push({name:y.name,properties:y.data})});AJS.bind("analyticsEvent",function(x,y){AJS.EventQueue.push({name:y.name,properties:y.data})});var w=function(){if(window.location.pathname.indexOf("/secure/admin/ViewApplicationProperties")>-1){AJS.$("[data-property-id='analytics-enabled']").remove()}else{if(window.location.pathname.indexOf("/secure/admin/EditApplicationProperties")>-1){var y=AJS.$(":contains(Enable Atlassian analytics)");if(y.size()>0){var x=y[y.size()-2];if(x){x.remove()}}}}}}(AJS.$));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.analytics.analytics-client:policy-update-init', location = 'js/policy-update-init.js' */
AJS.toInit(function(){var a=WRM.data.claim("com.atlassian.analytics.analytics-client:policy-update-init.policy-update-data-provider");if(a){WRM.require("wrc!com.atlassian.analytics.analytics-client:policy-update")}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.analytics.analytics-client:programmatic-analytics-init', location = 'js/programmatic-analytics-init.js' */
(function(){var a=WRM.data.claim("com.atlassian.analytics.analytics-client:programmatic-analytics-init.programmatic-analytics-data-provider");if(a){WRM.require("wrc!com.atlassian.analytics.analytics-client:programmatic-analytics")}}());
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:application-header-administration-cog-resource', location = 'header/cog.js' */
var NavLinks=(function(a){a.ApplicationHeader=function(b){b.Cog=(function(){var c=function(){var d=AJS.$("#system-admin-menu-content");if(d.length>0){return d}var e=AJS.$("#admin-menu-link-content");if(e.length>0){return e}return AJS.$("#bamboo\\.global\\.header-admin\\.menu")};return{getDropdown:c}}());return b}(a.ApplicationHeader||{});return a}(NavLinks||{}));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:administration-shortcuts-resources', location = 'adminshortcuts/adminshortcuts.soy' */
// This file was automatically generated from adminshortcuts.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace navlinks.templates.adminshortcuts.
 */

if (typeof navlinks == 'undefined') { var navlinks = {}; }
if (typeof navlinks.templates == 'undefined') { navlinks.templates = {}; }
if (typeof navlinks.templates.adminshortcuts == 'undefined') { navlinks.templates.adminshortcuts = {}; }


navlinks.templates.adminshortcuts.section = function(opt_data, opt_ignored) {
  var param5 = '<ul class="aui-list-truncate">';
  var linkList7 = opt_data.links;
  var linkListLen7 = linkList7.length;
  for (var linkIndex7 = 0; linkIndex7 < linkListLen7; linkIndex7++) {
    var linkData7 = linkList7[linkIndex7];
    param5 += '<li><a href="' + soy.$$escapeHtml(linkData7.link) + '">' + soy.$$escapeHtml(linkData7.label) + '</a></li>';
  }
  param5 += '</ul>';
  var output = '' + aui.dropdown2.section({id: 'nl-remote-admin-section', label: '\u5176\u5b83\u5e94\u7528\u7a0b\u5e8f', content: param5});
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.adminshortcuts.section.soyTemplateName = 'navlinks.templates.adminshortcuts.section';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:administration-shortcuts-resources', location = 'adminshortcuts/adminnavlinks.js' */
var NavLinks=(function(a){a.AdminShortcuts=(function(){var c=function(){return AJS.$.ajax({url:AJS.contextPath()+"/rest/menu/latest/admin",cache:false,dataType:"json"})};var b=function(){AJS.$("#nl-remote-admin-section").on("click","a",function(){NL.trackEvent("remoteAdminItemSelected",NL.getCurrentApplication(),$(this).attr("href"))})};return{render:function(){c().done(function(e){e=_.reject(e,function(f){return f.local===true});if(e.length){var d=navlinks.templates.adminshortcuts.section({links:e});a.ApplicationHeader.Cog.getDropdown().append(d);b()}})}}}());return a}(NavLinks||{}));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:administration-shortcuts', location = 'adminshortcuts/adminshortcuts-cdn.js' */
AJS.toInit(function(){if(AJS.DarkFeatures&&AJS.DarkFeatures.isEnabled("rotp.admin.shortcuts")){NavLinks.AdminShortcuts.render()}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:atlassian-ui-popup-display-controller', location = 'popups/DisplayController.js' */
AJS.Popups=AJS.Popups||{};AJS.Popups.DisplayController=function(){var c=[];var a=false;var b=false;AJS.toInit(function(){setTimeout(function(){AJS.Popups.DisplayController.render()},0)});return{request:function(d){c.push(d);if(a&&b===false){this.render()}},render:function(){c.sort(function(e,d){return e.weight-d.weight});a=true;if(c.length!==0){b=true;c[0].show()}}}}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugin.jslibs:moment-2.22.2', location = 'libs/moment/2.22.2/moment-2.22.2.js' */
'use strict';(function(g){define("atlassian/libs/moment-2.22.2",function(){var A={};g.call(A);return A.moment})})(function(){(function(g,A){g.moment=A()})(this,function(){function g(){return dc.apply(null,arguments)}function A(a){return a instanceof Array||"[object Array]"===Object.prototype.toString.call(a)}function ma(a){return null!=a&&"[object Object]"===Object.prototype.toString.call(a)}function B(a){return void 0===a}function fa(a){return"number"===typeof a||"[object Number]"===Object.prototype.toString.call(a)}
function wa(a){return a instanceof Date||"[object Date]"===Object.prototype.toString.call(a)}function mb(a,b){var c=[],d;for(d=0;d<a.length;++d)c.push(b(a[d],d));return c}function v(a,b){return Object.prototype.hasOwnProperty.call(a,b)}function S(a,b){for(var c in b)v(b,c)&&(a[c]=b[c]);v(b,"toString")&&(a.toString=b.toString);v(b,"valueOf")&&(a.valueOf=b.valueOf);return a}function K(a,b,c,d){return nb(a,b,c,d,!0).utc()}function n(a){null==a._pf&&(a._pf={empty:!1,unusedTokens:[],unusedInput:[],overflow:-2,
charsLeftOver:0,nullInput:!1,invalidMonth:null,invalidFormat:!1,userInvalidated:!1,iso:!1,parsedDateParts:[],meridiem:null,rfc2822:!1,weekdayMismatch:!1});return a._pf}function Oa(a){if(null==a._isValid){var b=n(a),c=ec.call(b.parsedDateParts,function(a){return null!=a});c=!isNaN(a._d.getTime())&&0>b.overflow&&!b.empty&&!b.invalidMonth&&!b.invalidWeekday&&!b.weekdayMismatch&&!b.nullInput&&!b.invalidFormat&&!b.userInvalidated&&(!b.meridiem||b.meridiem&&c);a._strict&&(c=c&&0===b.charsLeftOver&&0===
b.unusedTokens.length&&void 0===b.bigHour);if(null!=Object.isFrozen&&Object.isFrozen(a))return c;a._isValid=c}return a._isValid}function xa(a){var b=K(NaN);null!=a?S(n(b),a):n(b).userInvalidated=!0;return b}function Pa(a,b){var c;B(b._isAMomentObject)||(a._isAMomentObject=b._isAMomentObject);B(b._i)||(a._i=b._i);B(b._f)||(a._f=b._f);B(b._l)||(a._l=b._l);B(b._strict)||(a._strict=b._strict);B(b._tzm)||(a._tzm=b._tzm);B(b._isUTC)||(a._isUTC=b._isUTC);B(b._offset)||(a._offset=b._offset);B(b._pf)||(a._pf=
n(b));B(b._locale)||(a._locale=b._locale);if(0<Qa.length)for(c=0;c<Qa.length;c++){var d=Qa[c];var f=b[d];B(f)||(a[d]=f)}return a}function na(a){Pa(this,a);this._d=new Date(null!=a._d?a._d.getTime():NaN);this.isValid()||(this._d=new Date(NaN));!1===Ra&&(Ra=!0,g.updateOffset(this),Ra=!1)}function T(a){return a instanceof na||null!=a&&null!=a._isAMomentObject}function E(a){return 0>a?Math.ceil(a)||0:Math.floor(a)}function l(a){a=+a;var b=0;0!==a&&isFinite(a)&&(b=E(a));return b}function ob(a,b,c){var d=
Math.min(a.length,b.length),f=Math.abs(a.length-b.length),U=0,h;for(h=0;h<d;h++)(c&&a[h]!==b[h]||!c&&l(a[h])!==l(b[h]))&&U++;return U+f}function pb(a){!1===g.suppressDeprecationWarnings&&"undefined"!==typeof console&&console.warn&&console.warn("Deprecation warning: "+a)}function F(a,b){var c=!0;return S(function(){null!=g.deprecationHandler&&g.deprecationHandler(null,a);if(c){for(var d=[],f,U=0;U<arguments.length;U++){f="";if("object"===typeof arguments[U]){f+="\n["+U+"] ";for(var h in arguments[0])f+=
h+": "+arguments[0][h]+", ";f=f.slice(0,-2)}else f=arguments[U];d.push(f)}pb(a+"\nArguments: "+Array.prototype.slice.call(d).join("")+"\n"+Error().stack);c=!1}return b.apply(this,arguments)},b)}function qb(a,b){null!=g.deprecationHandler&&g.deprecationHandler(a,b);rb[a]||(pb(b),rb[a]=!0)}function L(a){return a instanceof Function||"[object Function]"===Object.prototype.toString.call(a)}function sb(a,b){var c=S({},a),d;for(d in b)v(b,d)&&(ma(a[d])&&ma(b[d])?(c[d]={},S(c[d],a[d]),S(c[d],b[d])):null!=
b[d]?c[d]=b[d]:delete c[d]);for(d in a)v(a,d)&&!v(b,d)&&ma(a[d])&&(c[d]=S({},c[d]));return c}function Sa(a){null!=a&&this.set(a)}function C(a,b){var c=a.toLowerCase();oa[c]=oa[c+"s"]=oa[b]=a}function G(a){return"string"===typeof a?oa[a]||oa[a.toLowerCase()]:void 0}function Ta(a){var b={},c,d;for(d in a)v(a,d)&&(c=G(d))&&(b[c]=a[d]);return b}function fc(a){var b=[],c;for(c in a)b.push({unit:c,priority:z[c]});b.sort(function(a,b){return a.priority-b.priority});return b}function O(a,b,c){var d=""+Math.abs(a);
return(0<=a?c?"+":"":"-")+Math.pow(10,Math.max(0,b-d.length)).toString().substr(1)+d}function m(a,b,c,d){var f=d;"string"===typeof d&&(f=function(){return this[d]()});a&&(ha[a]=f);b&&(ha[b[0]]=function(){return O(f.apply(this,arguments),b[1],b[2])});c&&(ha[c]=function(){return this.localeData().ordinal(f.apply(this,arguments),a)})}function gc(a){return a.match(/\[[\s\S]/)?a.replace(/^\[|\]$/g,""):a.replace(/\\/g,"")}function hc(a){var b=a.match(tb),c;var d=0;for(c=b.length;d<c;d++)b[d]=ha[b[d]]?ha[b[d]]:
gc(b[d]);return function(d){var f="",h;for(h=0;h<c;h++)f+=L(b[h])?b[h].call(d,a):b[h];return f}}function ya(a,b){if(!a.isValid())return a.localeData().invalidDate();b=ub(b,a.localeData());Ua[b]=Ua[b]||hc(b);return Ua[b](a)}function ub(a,b){function c(a){return b.longDateFormat(a)||a}var d=5;for(za.lastIndex=0;0<=d&&za.test(a);)a=a.replace(za,c),za.lastIndex=0,--d;return a}function k(a,b,c){Va[a]=L(b)?b:function(a,f){return a&&c?c:b}}function ic(a,b){return v(Va,a)?Va[a](b._strict,b._locale):new RegExp(jc(a))}
function jc(a){return aa(a.replace("\\","").replace(/\\(\[)|\\(\])|\[([^\]\[]*)\]|\\(.)/g,function(a,c,d,f,e){return c||d||f||e}))}function aa(a){return a.replace(/[-\/\\^$*+?.()|[\]{}]/g,"\\$\x26")}function r(a,b){var c,d=b;"string"===typeof a&&(a=[a]);fa(b)&&(d=function(a,c){c[b]=l(a)});for(c=0;c<a.length;c++)Wa[a[c]]=d}function pa(a,b){r(a,function(a,d,f,e){f._w=f._w||{};b(a,f._w,f,e)})}function V(a){return 0===a%4&&0!==a%100||0===a%400}function ia(a,b){return function(c){return null!=c?(vb(this,
a,c),g.updateOffset(this,b),this):Aa(this,a)}}function Aa(a,b){return a.isValid()?a._d["get"+(a._isUTC?"UTC":"")+b]():NaN}function vb(a,b,c){if(a.isValid()&&!isNaN(c))if("FullYear"===b&&V(a.year())&&1===a.month()&&29===a.date())a._d["set"+(a._isUTC?"UTC":"")+b](c,a.month(),Ba(c,a.month()));else a._d["set"+(a._isUTC?"UTC":"")+b](c)}function Ba(a,b){if(isNaN(a)||isNaN(b))return NaN;var c=(b%12+12)%12;return 1===c?V(a+(b-c)/12)?29:28:31-c%7%2}function wb(a,b){if(!a.isValid())return a;if("string"===typeof b)if(/^\d+$/.test(b))b=
l(b);else if(b=a.localeData().monthsParse(b),!fa(b))return a;var c=Math.min(a.date(),Ba(a.year(),b));a._d["set"+(a._isUTC?"UTC":"")+"Month"](b,c);return a}function xb(a){return null!=a?(wb(this,a),g.updateOffset(this,!0),this):Aa(this,"Month")}function yb(){function a(a,b){return b.length-a.length}var b=[],c=[],d=[],f;for(f=0;12>f;f++){var e=K([2E3,f]);b.push(this.monthsShort(e,""));c.push(this.months(e,""));d.push(this.months(e,""));d.push(this.monthsShort(e,""))}b.sort(a);c.sort(a);d.sort(a);for(f=
0;12>f;f++)b[f]=aa(b[f]),c[f]=aa(c[f]);for(f=0;24>f;f++)d[f]=aa(d[f]);this._monthsShortRegex=this._monthsRegex=new RegExp("^("+d.join("|")+")","i");this._monthsStrictRegex=new RegExp("^("+c.join("|")+")","i");this._monthsShortStrictRegex=new RegExp("^("+b.join("|")+")","i")}function kc(a,b,c,d,f,e,h){b=new Date(a,b,c,d,f,e,h);100>a&&0<=a&&isFinite(b.getFullYear())&&b.setFullYear(a);return b}function qa(a){var b=new Date(Date.UTC.apply(null,arguments));100>a&&0<=a&&isFinite(b.getUTCFullYear())&&b.setUTCFullYear(a);
return b}function Ca(a,b,c){c=7+b-c;return-((7+qa(a,0,c).getUTCDay()-b)%7)+c-1}function zb(a,b,c,d,f){c=(7+c-d)%7;d=Ca(a,d,f);d=1+7*(b-1)+c+d;0>=d?(b=a-1,a=(V(b)?366:365)+d):d>(V(a)?366:365)?(b=a+1,a=d-(V(a)?366:365)):(b=a,a=d);return{year:b,dayOfYear:a}}function ra(a,b,c){var d=Ca(a.year(),b,c);d=Math.floor((a.dayOfYear()-d-1)/7)+1;1>d?(a=a.year()-1,b=d+ba(a,b,c)):d>ba(a.year(),b,c)?(b=d-ba(a.year(),b,c),a=a.year()+1):(a=a.year(),b=d);return{week:b,year:a}}function ba(a,b,c){var d=Ca(a,b,c);b=Ca(a+
1,b,c);return((V(a)?366:365)-d+b)/7}function lc(a,b,c){var d;a=a.toLocaleLowerCase();if(!this._weekdaysParse)for(this._weekdaysParse=[],this._shortWeekdaysParse=[],this._minWeekdaysParse=[],d=0;7>d;++d){var f=K([2E3,1]).day(d);this._minWeekdaysParse[d]=this.weekdaysMin(f,"").toLocaleLowerCase();this._shortWeekdaysParse[d]=this.weekdaysShort(f,"").toLocaleLowerCase();this._weekdaysParse[d]=this.weekdays(f,"").toLocaleLowerCase()}if(c)b="dddd"===b?w.call(this._weekdaysParse,a):"ddd"===b?w.call(this._shortWeekdaysParse,
a):w.call(this._minWeekdaysParse,a);else if("dddd"===b){b=w.call(this._weekdaysParse,a);if(-1!==b)return b;b=w.call(this._shortWeekdaysParse,a);if(-1!==b)return b;b=w.call(this._minWeekdaysParse,a)}else if("ddd"===b){b=w.call(this._shortWeekdaysParse,a);if(-1!==b)return b;b=w.call(this._weekdaysParse,a);if(-1!==b)return b;b=w.call(this._minWeekdaysParse,a)}else{b=w.call(this._minWeekdaysParse,a);if(-1!==b)return b;b=w.call(this._weekdaysParse,a);if(-1!==b)return b;b=w.call(this._shortWeekdaysParse,
a)}return-1!==b?b:null}function Xa(){function a(a,b){return b.length-a.length}var b=[],c=[],d=[],f=[],e;for(e=0;7>e;e++){var h=K([2E3,1]).day(e);var g=this.weekdaysMin(h,"");var k=this.weekdaysShort(h,"");h=this.weekdays(h,"");b.push(g);c.push(k);d.push(h);f.push(g);f.push(k);f.push(h)}b.sort(a);c.sort(a);d.sort(a);f.sort(a);for(e=0;7>e;e++)c[e]=aa(c[e]),d[e]=aa(d[e]),f[e]=aa(f[e]);this._weekdaysMinRegex=this._weekdaysShortRegex=this._weekdaysRegex=new RegExp("^("+f.join("|")+")","i");this._weekdaysStrictRegex=
new RegExp("^("+d.join("|")+")","i");this._weekdaysShortStrictRegex=new RegExp("^("+c.join("|")+")","i");this._weekdaysMinStrictRegex=new RegExp("^("+b.join("|")+")","i")}function Ya(){return this.hours()%12||12}function Ab(a,b){m(a,0,0,function(){return this.localeData().meridiem(this.hours(),this.minutes(),b)})}function Bb(a,b){return b._meridiemParse}function Cb(a){return a?a.toLowerCase().replace("_","-"):a}function Da(a){var b=null;if(!x[a]&&"undefined"!==typeof module&&module&&module.exports)try{b=
sa._abbr,require("./locale/"+a),ja(b)}catch(c){}return x[a]}function ja(a,b){a&&((b=B(b)?W(a):Za(a,b))?sa=b:"undefined"!==typeof console&&console.warn&&console.warn("Locale "+a+" not found. Did you forget to load it?"));return sa._abbr}function Za(a,b){if(null!==b){var c=Db;b.abbr=a;if(null!=x[a])qb("defineLocaleOverride","use moment.updateLocale(localeName, config) to change an existing locale. moment.defineLocale(localeName, config) should only be used for creating a new locale See http://momentjs.com/guides/#/warnings/define-locale/ for more info."),
c=x[a]._config;else if(null!=b.parentLocale)if(null!=x[b.parentLocale])c=x[b.parentLocale]._config;else if(c=Da(b.parentLocale),null!=c)c=c._config;else return ta[b.parentLocale]||(ta[b.parentLocale]=[]),ta[b.parentLocale].push({name:a,config:b}),null;x[a]=new Sa(sb(c,b));ta[a]&&ta[a].forEach(function(a){Za(a.name,a.config)});ja(a);return x[a]}delete x[a];return null}function W(a){var b;a&&a._locale&&a._locale._abbr&&(a=a._locale._abbr);if(!a)return sa;if(!A(a)){if(b=Da(a))return b;a=[a]}a:{b=0;for(var c,
d,f,e;b<a.length;){e=Cb(a[b]).split("-");c=e.length;for(d=(d=Cb(a[b+1]))?d.split("-"):null;0<c;){if(f=Da(e.slice(0,c).join("-"))){a=f;break a}if(d&&d.length>=c&&ob(e,d,!0)>=c-1)break;c--}b++}a=sa}return a}function $a(a){var b;(b=a._a)&&-2===n(a).overflow&&(b=0>b[P]||11<b[P]?P:1>b[M]||b[M]>Ba(b[H],b[P])?M:0>b[y]||24<b[y]||24===b[y]&&(0!==b[I]||0!==b[Q]||0!==b[ca])?y:0>b[I]||59<b[I]?I:0>b[Q]||59<b[Q]?Q:0>b[ca]||999<b[ca]?ca:-1,n(a)._overflowDayOfYear&&(b<H||b>M)&&(b=M),n(a)._overflowWeeks&&-1===b&&
(b=mc),n(a)._overflowWeekday&&-1===b&&(b=nc),n(a).overflow=b);return a}function ka(a,b,c){return null!=a?a:null!=b?b:c}function ab(a){var b=[];if(!a._d){var c=new Date(g.now());c=a._useUTC?[c.getUTCFullYear(),c.getUTCMonth(),c.getUTCDate()]:[c.getFullYear(),c.getMonth(),c.getDate()];if(a._w&&null==a._a[M]&&null==a._a[P]){var d=a._w;if(null!=d.GG||null!=d.W||null!=d.E){var f=1;var e=4;var h=ka(d.GG,a._a[H],ra(u(),1,4).year);var k=ka(d.W,1);var m=ka(d.E,1);if(1>m||7<m)var l=!0}else if(f=a._locale._week.dow,
e=a._locale._week.doy,k=ra(u(),f,e),h=ka(d.gg,a._a[H],k.year),k=ka(d.w,k.week),null!=d.d){if(m=d.d,0>m||6<m)l=!0}else if(null!=d.e){if(m=d.e+f,0>d.e||6<d.e)l=!0}else m=f;1>k||k>ba(h,f,e)?n(a)._overflowWeeks=!0:null!=l?n(a)._overflowWeekday=!0:(l=zb(h,k,m,f,e),a._a[H]=l.year,a._dayOfYear=l.dayOfYear)}if(null!=a._dayOfYear){l=ka(a._a[H],c[H]);if(a._dayOfYear>(V(l)?366:365)||0===a._dayOfYear)n(a)._overflowDayOfYear=!0;l=qa(l,0,a._dayOfYear);a._a[P]=l.getUTCMonth();a._a[M]=l.getUTCDate()}for(l=0;3>l&&
null==a._a[l];++l)a._a[l]=b[l]=c[l];for(;7>l;l++)a._a[l]=b[l]=null==a._a[l]?2===l?1:0:a._a[l];24===a._a[y]&&0===a._a[I]&&0===a._a[Q]&&0===a._a[ca]&&(a._nextDay=!0,a._a[y]=0);a._d=(a._useUTC?qa:kc).apply(null,b);b=a._useUTC?a._d.getUTCDay():a._d.getDay();null!=a._tzm&&a._d.setUTCMinutes(a._d.getUTCMinutes()-a._tzm);a._nextDay&&(a._a[y]=24);a._w&&"undefined"!==typeof a._w.d&&a._w.d!==b&&(n(a).weekdayMismatch=!0)}}function Eb(a){var b;var c=a._i;var d=oc.exec(c)||pc.exec(c);if(d){n(a).iso=!0;c=0;for(b=
Ea.length;c<b;c++)if(Ea[c][1].exec(d[1])){var f=Ea[c][0];var e=!1!==Ea[c][2];break}if(null==f)a._isValid=!1;else{if(d[3]){c=0;for(b=bb.length;c<b;c++)if(bb[c][1].exec(d[3])){var h=(d[2]||" ")+bb[c][0];break}if(null==h){a._isValid=!1;return}}if(e||null==h){if(d[4])if(qc.exec(d[4]))var g="Z";else{a._isValid=!1;return}a._f=f+(h||"")+(g||"");cb(a)}else a._isValid=!1}}else a._isValid=!1}function Fb(a){var b=rc.exec(a._i.replace(/\([^)]*\)|[\n\t]/g," ").replace(/(\s\s+)/g," ").replace(/^\s\s*/,"").replace(/\s\s*$/,
""));if(b){var c=b[3],d=b[2],f=b[5],e=b[6],h=b[7],g=parseInt(b[4],10);c=[49>=g?2E3+g:999>=g?1900+g:g,Gb.indexOf(c),parseInt(d,10),parseInt(f,10),parseInt(e,10)];h&&c.push(parseInt(h,10));a:{if(h=b[1])if(h=Hb.indexOf(h),d=(new Date(c[0],c[1],c[2])).getDay(),h!==d){n(a).weekdayMismatch=!0;h=a._isValid=!1;break a}h=!0}h&&(a._a=c,(h=b[8])?b=sc[h]:b[9]?b=0:(b=parseInt(b[10],10),h=b%100,b=(b-h)/100*60+h),a._tzm=b,a._d=qa.apply(null,a._a),a._d.setUTCMinutes(a._d.getUTCMinutes()-a._tzm),n(a).rfc2822=!0)}else a._isValid=
!1}function tc(a){var b=uc.exec(a._i);null!==b?a._d=new Date(+b[1]):(Eb(a),!1===a._isValid&&(delete a._isValid,Fb(a),!1===a._isValid&&(delete a._isValid,g.createFromInputFallback(a))))}function cb(a){if(a._f===g.ISO_8601)Eb(a);else if(a._f===g.RFC_2822)Fb(a);else{a._a=[];n(a).empty=!0;var b=""+a._i,c,d,f=b.length,e=0;var h=ub(a._f,a._locale).match(tb)||[];for(c=0;c<h.length;c++){var k=h[c];if(d=(b.match(ic(k,a))||[])[0]){var l=b.substr(0,b.indexOf(d));0<l.length&&n(a).unusedInput.push(l);b=b.slice(b.indexOf(d)+
d.length);e+=d.length}if(ha[k]){if(d?n(a).empty=!1:n(a).unusedTokens.push(k),l=a,null!=d&&v(Wa,k))Wa[k](d,l._a,l,k)}else a._strict&&!d&&n(a).unusedTokens.push(k)}n(a).charsLeftOver=f-e;0<b.length&&n(a).unusedInput.push(b);12>=a._a[y]&&!0===n(a).bigHour&&0<a._a[y]&&(n(a).bigHour=void 0);n(a).parsedDateParts=a._a.slice(0);n(a).meridiem=a._meridiem;b=a._a;c=y;f=a._locale;h=a._a[y];e=a._meridiem;null!=e&&(null!=f.meridiemHour?h=f.meridiemHour(h,e):null!=f.isPM&&((f=f.isPM(e))&&12>h&&(h+=12),f||12!==h||
(h=0)));b[c]=h;ab(a);$a(a)}}function vc(a){if(!a._d){var b=Ta(a._i);a._a=mb([b.year,b.month,b.day||b.date,b.hour,b.minute,b.second,b.millisecond],function(a){return a&&parseInt(a,10)});ab(a)}}function Ib(a){var b=a._i,c=a._f;a._locale=a._locale||W(a._l);if(null===b||void 0===c&&""===b)return xa({nullInput:!0});"string"===typeof b&&(a._i=b=a._locale.preparse(b));if(T(b))return new na($a(b));if(wa(b))a._d=b;else if(A(c))if(0===a._f.length)n(a).invalidFormat=!0,a._d=new Date(NaN);else{for(b=0;b<a._f.length;b++){c=
0;var d=Pa({},a);null!=a._useUTC&&(d._useUTC=a._useUTC);d._f=a._f[b];cb(d);if(Oa(d)&&(c+=n(d).charsLeftOver,c+=10*n(d).unusedTokens.length,n(d).score=c,null==f||c<f)){var f=c;var e=d}}S(a,e||d)}else c?cb(a):wc(a);Oa(a)||(a._d=null);return a}function wc(a){var b=a._i;B(b)?a._d=new Date(g.now()):wa(b)?a._d=new Date(b.valueOf()):"string"===typeof b?tc(a):A(b)?(a._a=mb(b.slice(0),function(a){return parseInt(a,10)}),ab(a)):ma(b)?vc(a):fa(b)?a._d=new Date(b):g.createFromInputFallback(a)}function nb(a,b,
c,d,f){var e={};if(!0===c||!1===c)d=c,c=void 0;var h;if(h=ma(a))a:if(h=a,Object.getOwnPropertyNames)h=0===Object.getOwnPropertyNames(h).length;else{for(var g in h)if(h.hasOwnProperty(g)){h=!1;break a}h=!0}if(h||A(a)&&0===a.length)a=void 0;e._isAMomentObject=!0;e._useUTC=e._isUTC=f;e._l=c;e._i=a;e._f=b;e._strict=d;a=new na($a(Ib(e)));a._nextDay&&(a.add(1,"d"),a._nextDay=void 0);return a}function u(a,b,c,d){return nb(a,b,c,d,!1)}function Jb(a,b){var c;1===b.length&&A(b[0])&&(b=b[0]);if(!b.length)return u();
var d=b[0];for(c=1;c<b.length;++c)if(!b[c].isValid()||b[c][a](d))d=b[c];return d}function Fa(a){var b=Ta(a);a=b.year||0;var c=b.quarter||0,d=b.month||0,f=b.week||0,e=b.day||0,h=b.hour||0,g=b.minute||0,k=b.second||0,m=b.millisecond||0;a:{for(var n in b)if(-1===w.call(ua,n)||null!=b[n]&&isNaN(b[n])){b=!1;break a}n=!1;for(var p=0;p<ua.length;++p)if(b[ua[p]]){if(n){b=!1;break a}parseFloat(b[ua[p]])!==l(b[ua[p]])&&(n=!0)}b=!0}this._isValid=b;this._milliseconds=+m+1E3*k+6E4*g+36E5*h;this._days=+e+7*f;this._months=
+d+3*c+12*a;this._data={};this._locale=W();this._bubble()}function db(a){return a instanceof Fa}function eb(a){return 0>a?-1*Math.round(-1*a):Math.round(a)}function Kb(a,b){m(a,0,0,function(){var a=this.utcOffset(),d="+";0>a&&(a=-a,d="-");return d+O(~~(a/60),2)+b+O(~~a%60,2)})}function fb(a,b){a=(b||"").match(a);if(null===a)return null;a=((a[a.length-1]||[])+"").match(xc)||["-",0,0];b=+(60*a[1])+l(a[2]);return 0===b?0:"+"===a[0]?b:-b}function gb(a,b){return b._isUTC?(b=b.clone(),a=(T(a)||wa(a)?a.valueOf():
u(a).valueOf())-b.valueOf(),b._d.setTime(b._d.valueOf()+a),g.updateOffset(b,!1),b):u(a).local()}function Lb(){return this.isValid()?this._isUTC&&0===this._offset:!1}function J(a,b){var c=a;db(a)?c={ms:a._milliseconds,d:a._days,M:a._months}:fa(a)?(c={},b?c[b]=a:c.milliseconds=a):(b=yc.exec(a))?(c="-"===b[1]?-1:1,c={y:0,d:l(b[M])*c,h:l(b[y])*c,m:l(b[I])*c,s:l(b[Q])*c,ms:l(eb(1E3*b[ca]))*c}):(b=zc.exec(a))?(c="-"===b[1]?-1:1,c={y:da(b[2],c),M:da(b[3],c),w:da(b[4],c),d:da(b[5],c),h:da(b[6],c),m:da(b[7],
c),s:da(b[8],c)}):null==c?c={}:"object"===typeof c&&("from"in c||"to"in c)&&(b=u(c.from),c=u(c.to),b.isValid()&&c.isValid()?(c=gb(c,b),b.isBefore(c)?c=Mb(b,c):(c=Mb(c,b),c.milliseconds=-c.milliseconds,c.months=-c.months),b=c):b={milliseconds:0,months:0},c={},c.ms=b.milliseconds,c.M=b.months);c=new Fa(c);db(a)&&v(a,"_locale")&&(c._locale=a._locale);return c}function da(a,b){a=a&&parseFloat(a.replace(",","."));return(isNaN(a)?0:a)*b}function Mb(a,b){var c={milliseconds:0,months:0};c.months=b.month()-
a.month()+12*(b.year()-a.year());a.clone().add(c.months,"M").isAfter(b)&&--c.months;c.milliseconds=+b-+a.clone().add(c.months,"M");return c}function Nb(a,b){return function(c,d){if(null!==d&&!isNaN(+d)){qb(b,"moment()."+b+"(period, number) is deprecated. Please use moment()."+b+"(number, period). See http://momentjs.com/guides/#/warnings/add-inverted-param/ for more info.");var f=c;c=d;d=f}c=J("string"===typeof c?+c:c,d);Ob(this,c,a);return this}}function Ob(a,b,c,d){var f=b._milliseconds,e=eb(b._days);
b=eb(b._months);a.isValid()&&(d=null==d?!0:d,b&&wb(a,Aa(a,"Month")+b*c),e&&vb(a,"Date",Aa(a,"Date")+e*c),f&&a._d.setTime(a._d.valueOf()+f*c),d&&g.updateOffset(a,e||b))}function hb(a,b){var c=12*(b.year()-a.year())+(b.month()-a.month()),d=a.clone().add(c,"months");0>b-d?(a=a.clone().add(c-1,"months"),b=(b-d)/(d-a)):(a=a.clone().add(c+1,"months"),b=(b-d)/(a-d));return-(c+b)||0}function Pb(a){if(void 0===a)return this._locale._abbr;a=W(a);null!=a&&(this._locale=a);return this}function Qb(){return this._locale}
function Ga(a,b){m(0,[a,a.length],0,b)}function Rb(a,b,c,d,f){if(null==a)return ra(this,d,f).year;var e=ba(a,d,f);b>e&&(b=e);a=zb(a,b,c,d,f);a=qa(a.year,0,a.dayOfYear);this.year(a.getUTCFullYear());this.month(a.getUTCMonth());this.date(a.getUTCDate());return this}function Ac(a,b){b[ca]=l(1E3*("0."+a))}function Sb(a){return a}function Ha(a,b,c,d){var f=W();b=K().set(d,b);return f[c](b,a)}function Tb(a,b,c){fa(a)&&(b=a,a=void 0);a=a||"";if(null!=b)return Ha(a,b,c,"month");var d=[];for(b=0;12>b;b++)d[b]=
Ha(a,b,c,"month");return d}function ib(a,b,c,d){"boolean"!==typeof a&&(c=b=a,a=!1);fa(b)&&(c=b,b=void 0);b=b||"";var f=W();a=a?f._week.dow:0;if(null!=c)return Ha(b,(c+a)%7,d,"day");f=[];for(c=0;7>c;c++)f[c]=Ha(b,(c+a)%7,d,"day");return f}function Ub(a,b,c,d){b=J(b,c);a._milliseconds+=d*b._milliseconds;a._days+=d*b._days;a._months+=d*b._months;return a._bubble()}function Vb(a){return 0>a?Math.floor(a):Math.ceil(a)}function X(a){return function(){return this.as(a)}}function ea(a){return function(){return this.isValid()?
this._data[a]:NaN}}function Bc(a,b,c,d,f){return f.relativeTime(b||1,!!c,a,d)}function la(a){return(0<a)-(0>a)||+a}function Ia(){if(!this.isValid())return this.localeData().invalidDate();var a=jb(this._milliseconds)/1E3,b=jb(this._days),c=jb(this._months);var d=E(a/60);var f=E(d/60);a%=60;d%=60;var e=E(c/12);c%=12;a=a?a.toFixed(3).replace(/\.?0+$/,""):"";var h=this.asSeconds();if(!h)return"P0D";var g=0>h?"-":"",k=la(this._months)!==la(h)?"-":"",l=la(this._days)!==la(h)?"-":"";h=la(this._milliseconds)!==
la(h)?"-":"";return g+"P"+(e?k+e+"Y":"")+(c?k+c+"M":"")+(b?l+b+"D":"")+(f||d||a?"T":"")+(f?h+f+"H":"")+(d?h+d+"M":"")+(a?h+a+"S":"")}var ec=Array.prototype.some?Array.prototype.some:function(a){for(var b=Object(this),c=b.length>>>0,d=0;d<c;d++)if(d in b&&a.call(this,b[d],d,b))return!0;return!1};var Qa=g.momentProperties=[],Ra=!1,rb={};g.suppressDeprecationWarnings=!1;g.deprecationHandler=null;var Cc=Object.keys?Object.keys:function(a){var b,c=[];for(b in a)v(a,b)&&c.push(b);return c};var oa={},z=
{},tb=/(\[[^\[]*\])|(\\)?([Hh]mm(ss)?|Mo|MM?M?M?|Do|DDDo|DD?D?D?|ddd?d?|do?|w[o|w]?|W[o|W]?|Qo?|YYYYYY|YYYYY|YYYY|YY|gg(ggg?)?|GG(GGG?)?|e|E|a|A|hh?|HH?|kk?|mm?|ss?|S{1,9}|x|X|zz?|ZZ?|.)/g,za=/(\[[^\[]*\])|(\\)?(LTS|LT|LL?L?L?|l{1,4})/g,Ua={},ha={},Wb=/\d/,D=/\d\d/,Xb=/\d{3}/,kb=/\d{4}/,Ja=/[+-]?\d{6}/,t=/\d\d?/,Yb=/\d\d\d\d?/,Zb=/\d\d\d\d\d\d?/,Ka=/\d{1,3}/,lb=/\d{1,4}/,La=/[+-]?\d{1,6}/,Dc=/\d+/,Ma=/[+-]?\d+/,Ec=/Z|[+-]\d\d:?\d\d/gi,Na=/Z|[+-]\d\d(?::?\d\d)?/gi,va=/[0-9]{0,256}['a-z\u00A0-\u05FF\u0700-\uD7FF\uF900-\uFDCF\uFDF0-\uFF07\uFF10-\uFFEF]{1,256}|[\u0600-\u06FF\/]{1,256}(\s*?[\u0600-\u06FF]{1,256}){1,2}/i,
Va={},Wa={},H=0,P=1,M=2,y=3,I=4,Q=5,ca=6,mc=7,nc=8;m("Y",0,0,function(){var a=this.year();return 9999>=a?""+a:"+"+a});m(0,["YY",2],0,function(){return this.year()%100});m(0,["YYYY",4],0,"year");m(0,["YYYYY",5],0,"year");m(0,["YYYYYY",6,!0],0,"year");C("year","y");z.year=1;k("Y",Ma);k("YY",t,D);k("YYYY",lb,kb);k("YYYYY",La,Ja);k("YYYYYY",La,Ja);r(["YYYYY","YYYYYY"],H);r("YYYY",function(a,b){b[H]=2===a.length?g.parseTwoDigitYear(a):l(a)});r("YY",function(a,b){b[H]=g.parseTwoDigitYear(a)});r("Y",function(a,
b){b[H]=parseInt(a,10)});g.parseTwoDigitYear=function(a){return l(a)+(68<l(a)?1900:2E3)};var $b=ia("FullYear",!0);var w=Array.prototype.indexOf?Array.prototype.indexOf:function(a){var b;for(b=0;b<this.length;++b)if(this[b]===a)return b;return-1};m("M",["MM",2],"Mo",function(){return this.month()+1});m("MMM",0,0,function(a){return this.localeData().monthsShort(this,a)});m("MMMM",0,0,function(a){return this.localeData().months(this,a)});C("month","M");z.month=8;k("M",t);k("MM",t,D);k("MMM",function(a,
b){return b.monthsShortRegex(a)});k("MMMM",function(a,b){return b.monthsRegex(a)});r(["M","MM"],function(a,b){b[P]=l(a)-1});r(["MMM","MMMM"],function(a,b,c,d){d=c._locale.monthsParse(a,d,c._strict);null!=d?b[P]=d:n(c).invalidMonth=a});var ac=/D[oD]?(\[[^\[\]]*\]|\s)+MMMM?/,Gb="Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec".split(" ");m("w",["ww",2],"wo","week");m("W",["WW",2],"Wo","isoWeek");C("week","w");C("isoWeek","W");z.week=5;z.isoWeek=5;k("w",t);k("ww",t,D);k("W",t);k("WW",t,D);pa(["w","ww",
"W","WW"],function(a,b,c,d){b[d.substr(0,1)]=l(a)});m("d",0,"do","day");m("dd",0,0,function(a){return this.localeData().weekdaysMin(this,a)});m("ddd",0,0,function(a){return this.localeData().weekdaysShort(this,a)});m("dddd",0,0,function(a){return this.localeData().weekdays(this,a)});m("e",0,0,"weekday");m("E",0,0,"isoWeekday");C("day","d");C("weekday","e");C("isoWeekday","E");z.day=11;z.weekday=11;z.isoWeekday=11;k("d",t);k("e",t);k("E",t);k("dd",function(a,b){return b.weekdaysMinRegex(a)});k("ddd",
function(a,b){return b.weekdaysShortRegex(a)});k("dddd",function(a,b){return b.weekdaysRegex(a)});pa(["dd","ddd","dddd"],function(a,b,c,d){d=c._locale.weekdaysParse(a,d,c._strict);null!=d?b.d=d:n(c).invalidWeekday=a});pa(["d","e","E"],function(a,b,c,d){b[d]=l(a)});var Hb="Sun Mon Tue Wed Thu Fri Sat".split(" ");m("H",["HH",2],0,"hour");m("h",["hh",2],0,Ya);m("k",["kk",2],0,function(){return this.hours()||24});m("hmm",0,0,function(){return""+Ya.apply(this)+O(this.minutes(),2)});m("hmmss",0,0,function(){return""+
Ya.apply(this)+O(this.minutes(),2)+O(this.seconds(),2)});m("Hmm",0,0,function(){return""+this.hours()+O(this.minutes(),2)});m("Hmmss",0,0,function(){return""+this.hours()+O(this.minutes(),2)+O(this.seconds(),2)});Ab("a",!0);Ab("A",!1);C("hour","h");z.hour=13;k("a",Bb);k("A",Bb);k("H",t);k("h",t);k("k",t);k("HH",t,D);k("hh",t,D);k("kk",t,D);k("hmm",Yb);k("hmmss",Zb);k("Hmm",Yb);k("Hmmss",Zb);r(["H","HH"],y);r(["k","kk"],function(a,b,c){a=l(a);b[y]=24===a?0:a});r(["a","A"],function(a,b,c){c._isPm=c._locale.isPM(a);
c._meridiem=a});r(["h","hh"],function(a,b,c){b[y]=l(a);n(c).bigHour=!0});r("hmm",function(a,b,c){var d=a.length-2;b[y]=l(a.substr(0,d));b[I]=l(a.substr(d));n(c).bigHour=!0});r("hmmss",function(a,b,c){var d=a.length-4,e=a.length-2;b[y]=l(a.substr(0,d));b[I]=l(a.substr(d,2));b[Q]=l(a.substr(e));n(c).bigHour=!0});r("Hmm",function(a,b,c){c=a.length-2;b[y]=l(a.substr(0,c));b[I]=l(a.substr(c))});r("Hmmss",function(a,b,c){c=a.length-4;var d=a.length-2;b[y]=l(a.substr(0,c));b[I]=l(a.substr(c,2));b[Q]=l(a.substr(d))});
var Fc=ia("Hours",!0),Db={calendar:{sameDay:"[Today at] LT",nextDay:"[Tomorrow at] LT",nextWeek:"dddd [at] LT",lastDay:"[Yesterday at] LT",lastWeek:"[Last] dddd [at] LT",sameElse:"L"},longDateFormat:{LTS:"h:mm:ss A",LT:"h:mm A",L:"MM/DD/YYYY",LL:"MMMM D, YYYY",LLL:"MMMM D, YYYY h:mm A",LLLL:"dddd, MMMM D, YYYY h:mm A"},invalidDate:"Invalid date",ordinal:"%d",dayOfMonthOrdinalParse:/\d{1,2}/,relativeTime:{future:"in %s",past:"%s ago",s:"a few seconds",ss:"%d seconds",m:"a minute",mm:"%d minutes",h:"an hour",
hh:"%d hours",d:"a day",dd:"%d days",M:"a month",MM:"%d months",y:"a year",yy:"%d years"},months:"January February March April May June July August September October November December".split(" "),monthsShort:Gb,week:{dow:0,doy:6},weekdays:"Sunday Monday Tuesday Wednesday Thursday Friday Saturday".split(" "),weekdaysMin:"Su Mo Tu We Th Fr Sa".split(" "),weekdaysShort:Hb,meridiemParse:/[ap]\.?m?\.?/i},x={},ta={},sa,oc=/^\s*((?:[+-]\d{6}|\d{4})-(?:\d\d-\d\d|W\d\d-\d|W\d\d|\d\d\d|\d\d))(?:(T| )(\d\d(?::\d\d(?::\d\d(?:[.,]\d+)?)?)?)([\+\-]\d\d(?::?\d\d)?|\s*Z)?)?$/,
pc=/^\s*((?:[+-]\d{6}|\d{4})(?:\d\d\d\d|W\d\d\d|W\d\d|\d\d\d|\d\d))(?:(T| )(\d\d(?:\d\d(?:\d\d(?:[.,]\d+)?)?)?)([\+\-]\d\d(?::?\d\d)?|\s*Z)?)?$/,qc=/Z|[+-]\d\d(?::?\d\d)?/,Ea=[["YYYYYY-MM-DD",/[+-]\d{6}-\d\d-\d\d/],["YYYY-MM-DD",/\d{4}-\d\d-\d\d/],["GGGG-[W]WW-E",/\d{4}-W\d\d-\d/],["GGGG-[W]WW",/\d{4}-W\d\d/,!1],["YYYY-DDD",/\d{4}-\d{3}/],["YYYY-MM",/\d{4}-\d\d/,!1],["YYYYYYMMDD",/[+-]\d{10}/],["YYYYMMDD",/\d{8}/],["GGGG[W]WWE",/\d{4}W\d{3}/],["GGGG[W]WW",/\d{4}W\d{2}/,!1],["YYYYDDD",/\d{7}/]],bb=
[["HH:mm:ss.SSSS",/\d\d:\d\d:\d\d\.\d+/],["HH:mm:ss,SSSS",/\d\d:\d\d:\d\d,\d+/],["HH:mm:ss",/\d\d:\d\d:\d\d/],["HH:mm",/\d\d:\d\d/],["HHmmss.SSSS",/\d\d\d\d\d\d\.\d+/],["HHmmss,SSSS",/\d\d\d\d\d\d,\d+/],["HHmmss",/\d\d\d\d\d\d/],["HHmm",/\d\d\d\d/],["HH",/\d\d/]],uc=/^\/?Date\((\-?\d+)/i,rc=/^(?:(Mon|Tue|Wed|Thu|Fri|Sat|Sun),?\s)?(\d{1,2})\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s(\d{2,4})\s(\d\d):(\d\d)(?::(\d\d))?\s(?:(UT|GMT|[ECMP][SD]T)|([Zz])|([+-]\d{4}))$/,sc={UT:0,GMT:0,EDT:-240,
EST:-300,CDT:-300,CST:-360,MDT:-360,MST:-420,PDT:-420,PST:-480};g.createFromInputFallback=F("value provided is not in a recognized RFC2822 or ISO format. moment construction falls back to js Date(), which is not reliable across all browsers and versions. Non RFC2822/ISO date formats are discouraged and will be removed in an upcoming major release. Please refer to http://momentjs.com/guides/#/warnings/js-date/ for more info.",function(a){a._d=new Date(a._i+(a._useUTC?" UTC":""))});g.ISO_8601=function(){};
g.RFC_2822=function(){};var Gc=F("moment().min is deprecated, use moment.max instead. http://momentjs.com/guides/#/warnings/min-max/",function(){var a=u.apply(null,arguments);return this.isValid()&&a.isValid()?a<this?this:a:xa()}),Hc=F("moment().max is deprecated, use moment.min instead. http://momentjs.com/guides/#/warnings/min-max/",function(){var a=u.apply(null,arguments);return this.isValid()&&a.isValid()?a>this?this:a:xa()}),ua="year quarter month week day hour minute second millisecond".split(" ");
Kb("Z",":");Kb("ZZ","");k("Z",Na);k("ZZ",Na);r(["Z","ZZ"],function(a,b,c){c._useUTC=!0;c._tzm=fb(Na,a)});var xc=/([\+\-]|\d\d)/gi;g.updateOffset=function(){};var yc=/^(\-|\+)?(?:(\d*)[. ])?(\d+):(\d+)(?::(\d+)(\.\d*)?)?$/,zc=/^(-|\+)?P(?:([-+]?[0-9,.]*)Y)?(?:([-+]?[0-9,.]*)M)?(?:([-+]?[0-9,.]*)W)?(?:([-+]?[0-9,.]*)D)?(?:T(?:([-+]?[0-9,.]*)H)?(?:([-+]?[0-9,.]*)M)?(?:([-+]?[0-9,.]*)S)?)?$/;J.fn=Fa.prototype;J.invalid=function(){return J(NaN)};var Ic=Nb(1,"add"),Jc=Nb(-1,"subtract");g.defaultFormat=
"YYYY-MM-DDTHH:mm:ssZ";g.defaultFormatUtc="YYYY-MM-DDTHH:mm:ss[Z]";var bc=F("moment().lang() is deprecated. Instead, use moment().localeData() to get the language configuration. Use moment().locale() to change languages.",function(a){return void 0===a?this.localeData():this.locale(a)});m(0,["gg",2],0,function(){return this.weekYear()%100});m(0,["GG",2],0,function(){return this.isoWeekYear()%100});Ga("gggg","weekYear");Ga("ggggg","weekYear");Ga("GGGG","isoWeekYear");Ga("GGGGG","isoWeekYear");C("weekYear",
"gg");C("isoWeekYear","GG");z.weekYear=1;z.isoWeekYear=1;k("G",Ma);k("g",Ma);k("GG",t,D);k("gg",t,D);k("GGGG",lb,kb);k("gggg",lb,kb);k("GGGGG",La,Ja);k("ggggg",La,Ja);pa(["gggg","ggggg","GGGG","GGGGG"],function(a,b,c,d){b[d.substr(0,2)]=l(a)});pa(["gg","GG"],function(a,b,c,d){b[d]=g.parseTwoDigitYear(a)});m("Q",0,"Qo","quarter");C("quarter","Q");z.quarter=7;k("Q",Wb);r("Q",function(a,b){b[P]=3*(l(a)-1)});m("D",["DD",2],"Do","date");C("date","D");z.date=9;k("D",t);k("DD",t,D);k("Do",function(a,b){return a?
b._dayOfMonthOrdinalParse||b._ordinalParse:b._dayOfMonthOrdinalParseLenient});r(["D","DD"],M);r("Do",function(a,b){b[M]=l(a.match(t)[0])});var cc=ia("Date",!0);m("DDD",["DDDD",3],"DDDo","dayOfYear");C("dayOfYear","DDD");z.dayOfYear=4;k("DDD",Ka);k("DDDD",Xb);r(["DDD","DDDD"],function(a,b,c){c._dayOfYear=l(a)});m("m",["mm",2],0,"minute");C("minute","m");z.minute=14;k("m",t);k("mm",t,D);r(["m","mm"],I);var Kc=ia("Minutes",!1);m("s",["ss",2],0,"second");C("second","s");z.second=15;k("s",t);k("ss",t,
D);r(["s","ss"],Q);var Lc=ia("Seconds",!1);m("S",0,0,function(){return~~(this.millisecond()/100)});m(0,["SS",2],0,function(){return~~(this.millisecond()/10)});m(0,["SSS",3],0,"millisecond");m(0,["SSSS",4],0,function(){return 10*this.millisecond()});m(0,["SSSSS",5],0,function(){return 100*this.millisecond()});m(0,["SSSSSS",6],0,function(){return 1E3*this.millisecond()});m(0,["SSSSSSS",7],0,function(){return 1E4*this.millisecond()});m(0,["SSSSSSSS",8],0,function(){return 1E5*this.millisecond()});m(0,
["SSSSSSSSS",9],0,function(){return 1E6*this.millisecond()});C("millisecond","ms");z.millisecond=16;k("S",Ka,Wb);k("SS",Ka,D);k("SSS",Ka,Xb);var Y;for(Y="SSSS";9>=Y.length;Y+="S")k(Y,Dc);for(Y="S";9>=Y.length;Y+="S")r(Y,Ac);var Mc=ia("Milliseconds",!1);m("z",0,0,"zoneAbbr");m("zz",0,0,"zoneName");var e=na.prototype;e.add=Ic;e.calendar=function(a,b){a=a||u();var c=gb(a,this).startOf("day");c=g.calendarFormat(this,c)||"sameElse";b=b&&(L(b[c])?b[c].call(this,a):b[c]);return this.format(b||this.localeData().calendar(c,
this,u(a)))};e.clone=function(){return new na(this)};e.diff=function(a,b,c){if(!this.isValid())return NaN;a=gb(a,this);if(!a.isValid())return NaN;var d=6E4*(a.utcOffset()-this.utcOffset());b=G(b);switch(b){case "year":b=hb(this,a)/12;break;case "month":b=hb(this,a);break;case "quarter":b=hb(this,a)/3;break;case "second":b=(this-a)/1E3;break;case "minute":b=(this-a)/6E4;break;case "hour":b=(this-a)/36E5;break;case "day":b=(this-a-d)/864E5;break;case "week":b=(this-a-d)/6048E5;break;default:b=this-
a}return c?b:E(b)};e.endOf=function(a){a=G(a);if(void 0===a||"millisecond"===a)return this;"date"===a&&(a="day");return this.startOf(a).add(1,"isoWeek"===a?"week":a).subtract(1,"ms")};e.format=function(a){a||(a=this.isUtc()?g.defaultFormatUtc:g.defaultFormat);a=ya(this,a);return this.localeData().postformat(a)};e.from=function(a,b){return this.isValid()&&(T(a)&&a.isValid()||u(a).isValid())?J({to:this,from:a}).locale(this.locale()).humanize(!b):this.localeData().invalidDate()};e.fromNow=function(a){return this.from(u(),
a)};e.to=function(a,b){return this.isValid()&&(T(a)&&a.isValid()||u(a).isValid())?J({from:this,to:a}).locale(this.locale()).humanize(!b):this.localeData().invalidDate()};e.toNow=function(a){return this.to(u(),a)};e.get=function(a){a=G(a);return L(this[a])?this[a]():this};e.invalidAt=function(){return n(this).overflow};e.isAfter=function(a,b){a=T(a)?a:u(a);if(!this.isValid()||!a.isValid())return!1;b=G(B(b)?"millisecond":b);return"millisecond"===b?this.valueOf()>a.valueOf():a.valueOf()<this.clone().startOf(b).valueOf()};
e.isBefore=function(a,b){a=T(a)?a:u(a);if(!this.isValid()||!a.isValid())return!1;b=G(B(b)?"millisecond":b);return"millisecond"===b?this.valueOf()<a.valueOf():this.clone().endOf(b).valueOf()<a.valueOf()};e.isBetween=function(a,b,c,d){d=d||"()";return("("===d[0]?this.isAfter(a,c):!this.isBefore(a,c))&&(")"===d[1]?this.isBefore(b,c):!this.isAfter(b,c))};e.isSame=function(a,b){a=T(a)?a:u(a);if(!this.isValid()||!a.isValid())return!1;b=G(b||"millisecond");if("millisecond"===b)return this.valueOf()===a.valueOf();
a=a.valueOf();return this.clone().startOf(b).valueOf()<=a&&a<=this.clone().endOf(b).valueOf()};e.isSameOrAfter=function(a,b){return this.isSame(a,b)||this.isAfter(a,b)};e.isSameOrBefore=function(a,b){return this.isSame(a,b)||this.isBefore(a,b)};e.isValid=function(){return Oa(this)};e.lang=bc;e.locale=Pb;e.localeData=Qb;e.max=Hc;e.min=Gc;e.parsingFlags=function(){return S({},n(this))};e.set=function(a,b){if("object"===typeof a){a=Ta(a);b=fc(a);for(var c=0;c<b.length;c++)this[b[c].unit](a[b[c].unit])}else if(a=
G(a),L(this[a]))return this[a](b);return this};e.startOf=function(a){a=G(a);switch(a){case "year":this.month(0);case "quarter":case "month":this.date(1);case "week":case "isoWeek":case "day":case "date":this.hours(0);case "hour":this.minutes(0);case "minute":this.seconds(0);case "second":this.milliseconds(0)}"week"===a&&this.weekday(0);"isoWeek"===a&&this.isoWeekday(1);"quarter"===a&&this.month(3*Math.floor(this.month()/3));return this};e.subtract=Jc;e.toArray=function(){return[this.year(),this.month(),
this.date(),this.hour(),this.minute(),this.second(),this.millisecond()]};e.toObject=function(){return{years:this.year(),months:this.month(),date:this.date(),hours:this.hours(),minutes:this.minutes(),seconds:this.seconds(),milliseconds:this.milliseconds()}};e.toDate=function(){return new Date(this.valueOf())};e.toISOString=function(a){if(!this.isValid())return null;var b=(a=!0!==a)?this.clone().utc():this;return 0>b.year()||9999<b.year()?ya(b,a?"YYYYYY-MM-DD[T]HH:mm:ss.SSS[Z]":"YYYYYY-MM-DD[T]HH:mm:ss.SSSZ"):
L(Date.prototype.toISOString)?a?this.toDate().toISOString():(new Date(this.valueOf()+6E4*this.utcOffset())).toISOString().replace("Z",ya(b,"Z")):ya(b,a?"YYYY-MM-DD[T]HH:mm:ss.SSS[Z]":"YYYY-MM-DD[T]HH:mm:ss.SSSZ")};e.inspect=function(){if(!this.isValid())return"moment.invalid(/* "+this._i+" */)";var a="moment",b="";this.isLocal()||(a=0===this.utcOffset()?"moment.utc":"moment.parseZone",b="Z");a="["+a+'("]';var c=0<=this.year()&&9999>=this.year()?"YYYY":"YYYYYY";return this.format(a+c+"-MM-DD[T]HH:mm:ss.SSS"+
(b+'[")]'))};e.toJSON=function(){return this.isValid()?this.toISOString():null};e.toString=function(){return this.clone().locale("en").format("ddd MMM DD YYYY HH:mm:ss [GMT]ZZ")};e.unix=function(){return Math.floor(this.valueOf()/1E3)};e.valueOf=function(){return this._d.valueOf()-6E4*(this._offset||0)};e.creationData=function(){return{input:this._i,format:this._f,locale:this._locale,isUTC:this._isUTC,strict:this._strict}};e.year=$b;e.isLeapYear=function(){return V(this.year())};e.weekYear=function(a){return Rb.call(this,
a,this.week(),this.weekday(),this.localeData()._week.dow,this.localeData()._week.doy)};e.isoWeekYear=function(a){return Rb.call(this,a,this.isoWeek(),this.isoWeekday(),1,4)};e.quarter=e.quarters=function(a){return null==a?Math.ceil((this.month()+1)/3):this.month(3*(a-1)+this.month()%3)};e.month=xb;e.daysInMonth=function(){return Ba(this.year(),this.month())};e.week=e.weeks=function(a){var b=this.localeData().week(this);return null==a?b:this.add(7*(a-b),"d")};e.isoWeek=e.isoWeeks=function(a){var b=
ra(this,1,4).week;return null==a?b:this.add(7*(a-b),"d")};e.weeksInYear=function(){var a=this.localeData()._week;return ba(this.year(),a.dow,a.doy)};e.isoWeeksInYear=function(){return ba(this.year(),1,4)};e.date=cc;e.day=e.days=function(a){if(!this.isValid())return null!=a?this:NaN;var b=this._isUTC?this._d.getUTCDay():this._d.getDay();if(null!=a){var c=this.localeData();"string"===typeof a&&(isNaN(a)?(a=c.weekdaysParse(a),a="number"===typeof a?a:null):a=parseInt(a,10));return this.add(a-b,"d")}return b};
e.weekday=function(a){if(!this.isValid())return null!=a?this:NaN;var b=(this.day()+7-this.localeData()._week.dow)%7;return null==a?b:this.add(a-b,"d")};e.isoWeekday=function(a){if(!this.isValid())return null!=a?this:NaN;if(null!=a){var b=this.localeData();a="string"===typeof a?b.weekdaysParse(a)%7||7:isNaN(a)?null:a;return this.day(this.day()%7?a:a-7)}return this.day()||7};e.dayOfYear=function(a){var b=Math.round((this.clone().startOf("day")-this.clone().startOf("year"))/864E5)+1;return null==a?b:
this.add(a-b,"d")};e.hour=e.hours=Fc;e.minute=e.minutes=Kc;e.second=e.seconds=Lc;e.millisecond=e.milliseconds=Mc;e.utcOffset=function(a,b,c){var d=this._offset||0,e;if(!this.isValid())return null!=a?this:NaN;if(null!=a){if("string"===typeof a){if(a=fb(Na,a),null===a)return this}else 16>Math.abs(a)&&!c&&(a*=60);!this._isUTC&&b&&(e=15*-Math.round(this._d.getTimezoneOffset()/15));this._offset=a;this._isUTC=!0;null!=e&&this.add(e,"m");d!==a&&(!b||this._changeInProgress?Ob(this,J(a-d,"m"),1,!1):this._changeInProgress||
(this._changeInProgress=!0,g.updateOffset(this,!0),this._changeInProgress=null));return this}return this._isUTC?d:15*-Math.round(this._d.getTimezoneOffset()/15)};e.utc=function(a){return this.utcOffset(0,a)};e.local=function(a){this._isUTC&&(this.utcOffset(0,a),this._isUTC=!1,a&&this.subtract(15*-Math.round(this._d.getTimezoneOffset()/15),"m"));return this};e.parseZone=function(){if(null!=this._tzm)this.utcOffset(this._tzm,!1,!0);else if("string"===typeof this._i){var a=fb(Ec,this._i);null!=a?this.utcOffset(a):
this.utcOffset(0,!0)}return this};e.hasAlignedHourOffset=function(a){if(!this.isValid())return!1;a=a?u(a).utcOffset():0;return 0===(this.utcOffset()-a)%60};e.isDST=function(){return this.utcOffset()>this.clone().month(0).utcOffset()||this.utcOffset()>this.clone().month(5).utcOffset()};e.isLocal=function(){return this.isValid()?!this._isUTC:!1};e.isUtcOffset=function(){return this.isValid()?this._isUTC:!1};e.isUtc=Lb;e.isUTC=Lb;e.zoneAbbr=function(){return this._isUTC?"UTC":""};e.zoneName=function(){return this._isUTC?
"Coordinated Universal Time":""};e.dates=F("dates accessor is deprecated. Use date instead.",cc);e.months=F("months accessor is deprecated. Use month instead",xb);e.years=F("years accessor is deprecated. Use year instead",$b);e.zone=F("moment().zone is deprecated, use moment().utcOffset instead. http://momentjs.com/guides/#/warnings/zone/",function(a,b){return null!=a?("string"!==typeof a&&(a=-a),this.utcOffset(a,b),this):-this.utcOffset()});e.isDSTShifted=F("isDSTShifted is deprecated. See http://momentjs.com/guides/#/warnings/dst-shifted/ for more information",
function(){if(!B(this._isDSTShifted))return this._isDSTShifted;var a={};Pa(a,this);a=Ib(a);if(a._a){var b=a._isUTC?K(a._a):u(a._a);this._isDSTShifted=this.isValid()&&0<ob(a._a,b.toArray())}else this._isDSTShifted=!1;return this._isDSTShifted});var q=Sa.prototype;q.calendar=function(a,b,c){a=this._calendar[a]||this._calendar.sameElse;return L(a)?a.call(b,c):a};q.longDateFormat=function(a){var b=this._longDateFormat[a],c=this._longDateFormat[a.toUpperCase()];if(b||!c)return b;this._longDateFormat[a]=
c.replace(/MMMM|MM|DD|dddd/g,function(a){return a.slice(1)});return this._longDateFormat[a]};q.invalidDate=function(){return this._invalidDate};q.ordinal=function(a){return this._ordinal.replace("%d",a)};q.preparse=Sb;q.postformat=Sb;q.relativeTime=function(a,b,c,d){var e=this._relativeTime[c];return L(e)?e(a,b,c,d):e.replace(/%d/i,a)};q.pastFuture=function(a,b){a=this._relativeTime[0<a?"future":"past"];return L(a)?a(b):a.replace(/%s/i,b)};q.set=function(a){var b;for(b in a){var c=a[b];L(c)?this[b]=
c:this["_"+b]=c}this._config=a;this._dayOfMonthOrdinalParseLenient=new RegExp((this._dayOfMonthOrdinalParse.source||this._ordinalParse.source)+"|"+/\d{1,2}/.source)};q.months=function(a,b){return a?A(this._months)?this._months[a.month()]:this._months[(this._months.isFormat||ac).test(b)?"format":"standalone"][a.month()]:A(this._months)?this._months:this._months.standalone};q.monthsShort=function(a,b){return a?A(this._monthsShort)?this._monthsShort[a.month()]:this._monthsShort[ac.test(b)?"format":"standalone"][a.month()]:
A(this._monthsShort)?this._monthsShort:this._monthsShort.standalone};q.monthsParse=function(a,b,c){var d;if(this._monthsParseExact){a:{a=a.toLocaleLowerCase();if(!this._monthsParse)for(this._monthsParse=[],this._longMonthsParse=[],this._shortMonthsParse=[],d=0;12>d;++d){var e=K([2E3,d]);this._shortMonthsParse[d]=this.monthsShort(e,"").toLocaleLowerCase();this._longMonthsParse[d]=this.months(e,"").toLocaleLowerCase()}if(c)b="MMM"===b?w.call(this._shortMonthsParse,a):w.call(this._longMonthsParse,a);
else if("MMM"===b){b=w.call(this._shortMonthsParse,a);if(-1!==b)break a;b=w.call(this._longMonthsParse,a)}else{b=w.call(this._longMonthsParse,a);if(-1!==b)break a;b=w.call(this._shortMonthsParse,a)}b=-1!==b?b:null}return b}this._monthsParse||(this._monthsParse=[],this._longMonthsParse=[],this._shortMonthsParse=[]);for(d=0;12>d;d++)if(e=K([2E3,d]),c&&!this._longMonthsParse[d]&&(this._longMonthsParse[d]=new RegExp("^"+this.months(e,"").replace(".","")+"$","i"),this._shortMonthsParse[d]=new RegExp("^"+
this.monthsShort(e,"").replace(".","")+"$","i")),c||this._monthsParse[d]||(e="^"+this.months(e,"")+"|^"+this.monthsShort(e,""),this._monthsParse[d]=new RegExp(e.replace(".",""),"i")),c&&"MMMM"===b&&this._longMonthsParse[d].test(a)||c&&"MMM"===b&&this._shortMonthsParse[d].test(a)||!c&&this._monthsParse[d].test(a))return d};q.monthsRegex=function(a){if(this._monthsParseExact)return v(this,"_monthsRegex")||yb.call(this),a?this._monthsStrictRegex:this._monthsRegex;v(this,"_monthsRegex")||(this._monthsRegex=
va);return this._monthsStrictRegex&&a?this._monthsStrictRegex:this._monthsRegex};q.monthsShortRegex=function(a){if(this._monthsParseExact)return v(this,"_monthsRegex")||yb.call(this),a?this._monthsShortStrictRegex:this._monthsShortRegex;v(this,"_monthsShortRegex")||(this._monthsShortRegex=va);return this._monthsShortStrictRegex&&a?this._monthsShortStrictRegex:this._monthsShortRegex};q.week=function(a){return ra(a,this._week.dow,this._week.doy).week};q.firstDayOfYear=function(){return this._week.doy};
q.firstDayOfWeek=function(){return this._week.dow};q.weekdays=function(a,b){return a?A(this._weekdays)?this._weekdays[a.day()]:this._weekdays[this._weekdays.isFormat.test(b)?"format":"standalone"][a.day()]:A(this._weekdays)?this._weekdays:this._weekdays.standalone};q.weekdaysMin=function(a){return a?this._weekdaysMin[a.day()]:this._weekdaysMin};q.weekdaysShort=function(a){return a?this._weekdaysShort[a.day()]:this._weekdaysShort};q.weekdaysParse=function(a,b,c){var d;if(this._weekdaysParseExact)return lc.call(this,
a,b,c);this._weekdaysParse||(this._weekdaysParse=[],this._minWeekdaysParse=[],this._shortWeekdaysParse=[],this._fullWeekdaysParse=[]);for(d=0;7>d;d++){var e=K([2E3,1]).day(d);c&&!this._fullWeekdaysParse[d]&&(this._fullWeekdaysParse[d]=new RegExp("^"+this.weekdays(e,"").replace(".","\\.?")+"$","i"),this._shortWeekdaysParse[d]=new RegExp("^"+this.weekdaysShort(e,"").replace(".","\\.?")+"$","i"),this._minWeekdaysParse[d]=new RegExp("^"+this.weekdaysMin(e,"").replace(".","\\.?")+"$","i"));this._weekdaysParse[d]||
(e="^"+this.weekdays(e,"")+"|^"+this.weekdaysShort(e,"")+"|^"+this.weekdaysMin(e,""),this._weekdaysParse[d]=new RegExp(e.replace(".",""),"i"));if(c&&"dddd"===b&&this._fullWeekdaysParse[d].test(a)||c&&"ddd"===b&&this._shortWeekdaysParse[d].test(a)||c&&"dd"===b&&this._minWeekdaysParse[d].test(a)||!c&&this._weekdaysParse[d].test(a))return d}};q.weekdaysRegex=function(a){if(this._weekdaysParseExact)return v(this,"_weekdaysRegex")||Xa.call(this),a?this._weekdaysStrictRegex:this._weekdaysRegex;v(this,"_weekdaysRegex")||
(this._weekdaysRegex=va);return this._weekdaysStrictRegex&&a?this._weekdaysStrictRegex:this._weekdaysRegex};q.weekdaysShortRegex=function(a){if(this._weekdaysParseExact)return v(this,"_weekdaysRegex")||Xa.call(this),a?this._weekdaysShortStrictRegex:this._weekdaysShortRegex;v(this,"_weekdaysShortRegex")||(this._weekdaysShortRegex=va);return this._weekdaysShortStrictRegex&&a?this._weekdaysShortStrictRegex:this._weekdaysShortRegex};q.weekdaysMinRegex=function(a){if(this._weekdaysParseExact)return v(this,
"_weekdaysRegex")||Xa.call(this),a?this._weekdaysMinStrictRegex:this._weekdaysMinRegex;v(this,"_weekdaysMinRegex")||(this._weekdaysMinRegex=va);return this._weekdaysMinStrictRegex&&a?this._weekdaysMinStrictRegex:this._weekdaysMinRegex};q.isPM=function(a){return"p"===(a+"").toLowerCase().charAt(0)};q.meridiem=function(a,b,c){return 11<a?c?"pm":"PM":c?"am":"AM"};ja("en",{dayOfMonthOrdinalParse:/\d{1,2}(th|st|nd|rd)/,ordinal:function(a){var b=a%10;b=1===l(a%100/10)?"th":1===b?"st":2===b?"nd":3===b?"rd":
"th";return a+b}});g.lang=F("moment.lang is deprecated. Use moment.locale instead.",ja);g.langData=F("moment.langData is deprecated. Use moment.localeData instead.",W);var R=Math.abs,Nc=X("ms"),Oc=X("s"),Pc=X("m"),Qc=X("h"),Rc=X("d"),Sc=X("w"),Tc=X("M"),Uc=X("y"),Vc=ea("milliseconds"),Wc=ea("seconds"),Xc=ea("minutes"),Yc=ea("hours"),Zc=ea("days"),$c=ea("months"),ad=ea("years"),Z=Math.round,N={ss:44,s:45,m:45,h:22,d:26,M:11},jb=Math.abs,p=Fa.prototype;p.isValid=function(){return this._isValid};p.abs=
function(){var a=this._data;this._milliseconds=R(this._milliseconds);this._days=R(this._days);this._months=R(this._months);a.milliseconds=R(a.milliseconds);a.seconds=R(a.seconds);a.minutes=R(a.minutes);a.hours=R(a.hours);a.months=R(a.months);a.years=R(a.years);return this};p.add=function(a,b){return Ub(this,a,b,1)};p.subtract=function(a,b){return Ub(this,a,b,-1)};p.as=function(a){if(!this.isValid())return NaN;var b=this._milliseconds;a=G(a);if("month"===a||"year"===a){var c=this._days+b/864E5;c=this._months+
4800*c/146097;return"month"===a?c:c/12}c=this._days+Math.round(146097*this._months/4800);switch(a){case "week":return c/7+b/6048E5;case "day":return c+b/864E5;case "hour":return 24*c+b/36E5;case "minute":return 1440*c+b/6E4;case "second":return 86400*c+b/1E3;case "millisecond":return Math.floor(864E5*c)+b;default:throw Error("Unknown unit "+a);}};p.asMilliseconds=Nc;p.asSeconds=Oc;p.asMinutes=Pc;p.asHours=Qc;p.asDays=Rc;p.asWeeks=Sc;p.asMonths=Tc;p.asYears=Uc;p.valueOf=function(){return this.isValid()?
this._milliseconds+864E5*this._days+this._months%12*2592E6+31536E6*l(this._months/12):NaN};p._bubble=function(){var a=this._milliseconds,b=this._days,c=this._months,d=this._data;0<=a&&0<=b&&0<=c||0>=a&&0>=b&&0>=c||(a+=864E5*Vb(146097*c/4800+b),c=b=0);d.milliseconds=a%1E3;a=E(a/1E3);d.seconds=a%60;a=E(a/60);d.minutes=a%60;a=E(a/60);d.hours=a%24;b+=E(a/24);a=E(4800*b/146097);c+=a;b-=Vb(146097*a/4800);a=E(c/12);d.days=b;d.months=c%12;d.years=a;return this};p.clone=function(){return J(this)};p.get=function(a){a=
G(a);return this.isValid()?this[a+"s"]():NaN};p.milliseconds=Vc;p.seconds=Wc;p.minutes=Xc;p.hours=Yc;p.days=Zc;p.weeks=function(){return E(this.days()/7)};p.months=$c;p.years=ad;p.humanize=function(a){if(!this.isValid())return this.localeData().invalidDate();var b=this.localeData();var c=!a;var d=J(this).abs(),e=Z(d.as("s")),g=Z(d.as("m")),h=Z(d.as("h")),k=Z(d.as("d")),l=Z(d.as("M"));d=Z(d.as("y"));e=e<=N.ss&&["s",e]||e<N.s&&["ss",e]||1>=g&&["m"]||g<N.m&&["mm",g]||1>=h&&["h"]||h<N.h&&["hh",h]||1>=
k&&["d"]||k<N.d&&["dd",k]||1>=l&&["M"]||l<N.M&&["MM",l]||1>=d&&["y"]||["yy",d];e[2]=c;e[3]=0<+this;e[4]=b;c=Bc.apply(null,e);a&&(c=b.pastFuture(+this,c));return b.postformat(c)};p.toISOString=Ia;p.toString=Ia;p.toJSON=Ia;p.locale=Pb;p.localeData=Qb;p.toIsoString=F("toIsoString() is deprecated. Please use toISOString() instead (notice the capitals)",Ia);p.lang=bc;m("X",0,0,"unix");m("x",0,0,"valueOf");k("x",Ma);k("X",/[+-]?\d+(\.\d{1,3})?/);r("X",function(a,b,c){c._d=new Date(1E3*parseFloat(a,10))});
r("x",function(a,b,c){c._d=new Date(l(a))});g.version="2.22.2";var dc=u;g.fn=e;g.min=function(){var a=[].slice.call(arguments,0);return Jb("isBefore",a)};g.max=function(){var a=[].slice.call(arguments,0);return Jb("isAfter",a)};g.now=function(){return Date.now?Date.now():+new Date};g.utc=K;g.unix=function(a){return u(1E3*a)};g.months=function(a,b){return Tb(a,b,"months")};g.isDate=wa;g.locale=ja;g.invalid=xa;g.duration=J;g.isMoment=T;g.weekdays=function(a,b,c){return ib(a,b,c,"weekdays")};g.parseZone=
function(){return u.apply(null,arguments).parseZone()};g.localeData=W;g.isDuration=db;g.monthsShort=function(a,b){return Tb(a,b,"monthsShort")};g.weekdaysMin=function(a,b,c){return ib(a,b,c,"weekdaysMin")};g.defineLocale=Za;g.updateLocale=function(a,b){if(null!=b){var c=Db;var d=Da(a);null!=d&&(c=d._config);b=sb(c,b);b=new Sa(b);b.parentLocale=x[a];x[a]=b;ja(a)}else null!=x[a]&&(null!=x[a].parentLocale?x[a]=x[a].parentLocale:null!=x[a]&&delete x[a]);return x[a]};g.locales=function(){return Cc(x)};
g.weekdaysShort=function(a,b,c){return ib(a,b,c,"weekdaysShort")};g.normalizeUnits=G;g.relativeTimeRounding=function(a){return void 0===a?Z:"function"===typeof a?(Z=a,!0):!1};g.relativeTimeThreshold=function(a,b){if(void 0===N[a])return!1;if(void 0===b)return N[a];N[a]=b;"s"===a&&(N.ss=b-1);return!0};g.calendarFormat=function(a,b){a=a.diff(b,"days",!0);return-6>a?"sameElse":-1>a?"lastWeek":0>a?"lastDay":1>a?"sameDay":2>a?"nextDay":7>a?"nextWeek":"sameElse"};g.prototype=e;g.HTML5_FMT={DATETIME_LOCAL:"YYYY-MM-DDTHH:mm",
DATETIME_LOCAL_SECONDS:"YYYY-MM-DDTHH:mm:ss",DATETIME_LOCAL_MS:"YYYY-MM-DDTHH:mm:ss.SSS",DATE:"YYYY-MM-DD",TIME:"HH:mm",TIME_SECONDS:"HH:mm:ss",TIME_MS:"HH:mm:ss.SSS",WEEK:"YYYY-[W]WW",MONTH:"YYYY-MM"};return g})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:momentjs', location = '/static/lib/moment/moment.lib.js' */
define("jira/moment/moment.lib",["atlassian/libs/moment-2.22.2"],function(n){return n});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:momentjs', location = '/static/lib/moment/moment.jira.formatter.js' */
define("jira/moment/moment.jira.formatter",["underscore","exports"],function(t,e){"use strict";function r(e){return u[e]||t.reduce(e,function(t,e){return t+(u[e]||e)},"")}function n(e){var n=!1,u=!1,a="",o=t.reduce(e,function(t,e,o,i){return u?u=!1:"'"===e?(a&&(t+=r(a),a=""),"'"===i[o+1]?(t+=e,u=!0):(t+=n?"]":"[",n=!n)):n?t+=e:/[a-zA-Z]/.test(e)?a&&a[a.length-1]!==e?(t+=r(a),a=e):a+=e:(a&&(t+=r(a),a=""),t+=e),t},"");return a&&(o+=r(a)),o}var u={d:"D",y:"Y",a:"A",E:"d",u:"d",Z:"ZZ",z:"[GMT]ZZ",XX:"ZZ",XXX:"Z"};e.translateSimpleDateFormat=n});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:momentjs', location = '/static/lib/moment/moment.jira.i18n.js' */
!function(){"use strict";var e;define("jira/moment/moment.jira.datetime-formats",["wrm/data"],function(t){return e||(e=t.claim("jira.core:dateFormatProvider.allFormats")),e})}(),define("jira/moment/moment.jira.i18n",["jira/moment/moment.lib","jira/moment/moment.jira.datetime-formats","jira/moment/moment.jira.formatter","jira/util/formatter"],function(e,t,m,a){"use strict";var o=t.dateFormats,r=t.lookAndFeelFormats;return e.locale("jira",{months:o.months,monthsShort:o.monthsShort,weekdays:o.weekdays,weekdaysShort:o.weekdaysShort,weekdaysMin:o.weekdaysShort,longDateFormat:{LT:m.translateSimpleDateFormat(r.time),L:m.translateSimpleDateFormat(r.day),LL:m.translateSimpleDateFormat(r.dmy),LLL:m.translateSimpleDateFormat(r.complete)},meridiem:function(e){return o.meridiem[+(e>11)]},calendar:{sameDay:"LLL",nextDay:"LLL",nextWeek:"LLL",lastDay:"LLL",lastWeek:"LLL",sameElse:"LLL"},relativeTime:{future:a.format("{0}\u5185","%s"),past:a.format("{0}\u4e4b\u524d","%s"),s:"\u51e0\u79d2\u949f",m:"\u4e00\u5206\u949f",mm:a.format("{0}\u5206\u949f","%d"),h:"\u4e00\u5c0f\u65f6",hh:a.format("{0}\u5c0f\u65f6","%d"),d:"\u4e00\u5929",dd:a.format("{0}\u5929","%d"),M:"\u4e00\u6708",MM:a.format("{0}\u6708","%d"),y:"\u4e00\u5e74",yy:a.format("{0}\u5e74","%d")}}),e});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:momentjs', location = '/static/lib/moment/moment.js' */
define("jira/moment",["jira/moment/moment.jira.i18n"],function(t){return t.splitDate=function(t,n){return Date.splitDate(t,n)},t});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:momentjs', location = '/static/lib/moment/moment.legacy.js' */
!function(){AJS.namespace("JIRA.translateSimpleDateFormat",null,require("jira/moment/moment.jira.formatter").translateSimpleDateFormat),AJS.namespace("window.moment",null,require("jira/moment"))}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:shared-utility', location = 'js/ctrlv/time.js' */
define("dndattachment/ctrlv/time",["require"],function(a){var b=a("jira/moment");return b});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:shared-utility', location = 'js/ctrlv/base64decode.js' */
define("dndattachment/ctrlv/base64decode",["exports"],function(a){function c(d){return d>64&&d<91?d-65:d>96&&d<123?d-71:d>47&&d<58?d+4:d===43?62:d===47?63:0}var b=function(f){var m=f.replace(/[^A-Za-z0-9\+\/]/g,""),g=m.length,d=g*3+1>>2,k=new Uint8Array(d);for(var j,h,i=0,e=0,l=0;l<g;l++){h=l&3;i|=c(m.charCodeAt(l))<<18-6*h;if(h===3||g-l===1){for(j=0;j<3&&e<d;j++,e++){k[e]=i>>>(16>>>j&24)&255}i=0}}return k};a.base64decode=b;a.decodeBase64DataUri=function(e){var g=";base64,",f=e.slice(0,e.indexOf(g))+g,d=e.substring(f.length);return b(d)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:shared-utility', location = 'js/ctrlv/utility.js' */
define("dndattachment/ctrlv/utility",["require","exports"],function(b,c){var g=b("jira/jquery/deferred");var d=b("jquery");var k=b("underscore");var i=b("dndattachment/ctrlv/base64decode");var a=b("dndattachment/ctrlv/time");var f=b("dndattachment/ctrlv/version");var j=b("jira/flag");var m=b("jira/util/navigator");c.isKeyPasteEvent=function(q,n){var o=86;var p=((n||"").indexOf("Mac")!==-1)||m.isMac();return q.which===o&&(p?q.metaKey:q.ctrlKey)};c.isImagePasteEvent=function(q){if(q&&q.clipboardData){var o=k.union(k.toArray(q.clipboardData.items),k.toArray(q.clipboardData.files));var p=o.some(function(r){return r.type.indexOf("image")!==-1});var n=o.some(function(r){return r.type.indexOf("rtf")!==-1});return p&&!n}};var h=function(o){if(!o.clipboardData||!o.clipboardData.types||!k.contains(o.clipboardData.types,"text/html")){return}var n=d(o.clipboardData.getData("text/html"));if(n.length===1&&n.children().length===0&&n[0].nodeName.toLowerCase()==="img"){return n[0].src}return null};c.getHtmlImagePaste=h;c.isHtmlImagePasteEvent=function(n){return h(n)!=null};c.isTextPasteEvent=function(n){return n.clipboardData.types&&(k.contains(n.clipboardData.types,"text/plain")||k.contains(n.clipboardData.types,"text/html"))};c.isContentEditable=function(n){return n.contentEditable=="true"||n.contentEditable==""||(n.contentEditable=="inherit"&&d(n).parents().is("[contenteditable=true]"))};c.isValidFileName=function(n){return !k(["\\","/",'"',":","?","*","<","|",">","!"]).any(function(o){return n.indexOf(o)>-1})};c.getTextPasteContent=function(n){return n.clipboardData.types&&k.contains(n.clipboardData.types,"text/plain")&&n.clipboardData.getData("text/plain")||""};c.getTextContent=function(n){if(!n){return""}else{if(m.isIE()&&typeof n.innerText!=="undefined"){return n.innerText}else{return d(n).text()}}};c.normalizePasteEvent=function(n){if(n&&n.originalEvent&&n.originalEvent.clipboardData){n.clipboardData=n.originalEvent.clipboardData}if(n&&!n.clipboardData&&window.clipboardData){n.clipboardData={files:window.clipboardData.files,types:{contains:function(o){if(o=="text/plain"){return !!window.clipboardData.getData("Text")}}},getData:function(o){if(o=="text/plain"){return window.clipboardData.getData("Text")}}}}return n};c.browserIsSupported=function(n){var p=((n||"").indexOf("Mac")!==-1)||m.isMac();var o=m.isSafari();return !(p&&o)&&!(f.isIE8()||f.isIE9()||f.isIE10())};c.browserIsNativePaste=function(){return m.isChrome()||m.isSafari()};c.isWikiTextfield=function(n){return n.is(":input")&&n.hasClass("wiki-textfield")};var e=function(n){return n.is("input#summary")};var l=function(o,v,r,t){var q=d(v);if(!(c.isWikiTextfield(q)||e(q))||!o){return}if(!q.is(":focus")){q.one("focus",function(){var w=q[0];w.selectionStart=w.selectionEnd=r+o.length})}var s=q.val();var n=s.substring(0,r);var u=s.substring(t,s.length);var p=q.data("wikiEditor");if(p&&p.undoRedoEl&&k.isFunction(p.undoRedoEl.recordHistoryItem)){p.undoRedoEl.recordHistoryItem()}q.val(n+o+u);q.trigger("input");v.selectionStart=v.selectionEnd=r+o.length;if(p&&p.undoRedoEl&&k.isFunction(p.undoRedoEl.updateCurrent)){p.undoRedoEl.updateCurrent()}};c.insertToInput=l;c.isThumbnailsAllowed=function(){var n=d("#dnd-metadata-webpanel").data("thumbnails-allowed");var o=true;if(n!=null){o=!!n}return o};c.getMarkup=function(s){var q=["bmp","gif"];var r=["pjpeg","jpeg","jpg","png"];var p=s.split(".");var o=c.isThumbnailsAllowed();var n=p[p.length-1].toLowerCase();if(o&&k.contains(r,n)){return"!"+s+"|thumbnail!"}else{if(k.contains(r,n)||k.contains(q,n)){return"!"+s+"!"}else{return"[^"+s+"]"}}};c.insertWikiMarkup=function(s,p,q,r,n){var o=c.getMarkup(s,n);if(o){o=" "+o+" ";l(o,p,q,r)}};c.loadImage=function(p){var n=new g();var o=new Image();o.setAttribute("crossOrigin","anonymous");o.onload=function(){n.resolve(o)};o.onerror=n.reject.bind(n);o.src=p;if(o.width>0&&o.height>0){n.resolve(o)}return n};c.convertImageToBlob=function(p){var n=d("<canvas>").attr("width",p.width).attr("height",p.height)[0];n.getContext("2d").drawImage(p,0,0);try{if(n.mozGetAsFile){return n.mozGetAsFile("image/png")}if(n.toDataURL){return new Blob([i.decodeBase64DataUri(n.toDataURL("image/png"))],{type:"image/png"})}}catch(o){return null}};c.convertBlobToImage=function(n,p){var o=new Blob([n.slice()],{type:n.type});o.lastModifiedDate=new Date();o.name=p;return o};c.dropFileToElement=function(p,o){var r=p.name;if(!r){r=this.generateFileName();p.name=r+".png"}var q=d.Event("drop",{dataTransfer:{files:[p]}});var n=false;d(document).on("dropHandled.dropFileToElement",function(){n=true});o.trigger(q);d(document).off(".dropFileToElement");return n};c.generateFileName=function(){return"image-"+a().format("YYYY-MM-DD-HH-mm-ss-SSS")};c.getCurrentIssueId=function(){return JIRA.Issues.Api.getSelectedIssueId()};c.showErrorMsg=function(o,n){j.showErrorMsg(o,n)};c.createBlobFromFile=function(n){var o=new Blob([n.slice()],{type:n.type});o.name=n.name;return o};c.dragEventContainsFiles=function(o){if(!o.dataTransfer||!o.dataTransfer.types){return true}var n=o.dataTransfer.types;if(m.isMozilla()){return k.contains(n,"application/x-moz-file")}else{return k.contains(n,"Files")}}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:shared-utility', location = 'js/ctrlv/ie_version.js' */
define("dndattachment/ctrlv/version",["require","exports"],function(e,h){var i=e("jquery");var m=e("underscore");var n=e("jira/util/navigator");var g=document.documentElement.className.split(/\s+/);var c=n.isIE();var l=i.inArray("msie-gt-7",g)>-1;var k=i.inArray("msie-gt-8",g)>-1;var j=i.inArray("msie-gt-9",g)>-1;var b=i.inArray("msie-gt-10",g)>-1;var f=c&&l&&!k;var d=c&&k&&!j;var a=c&&j&&!b;h.isIE8=m.once(function(){return f});h.isIE9=m.once(function(){return d});h.isIE10=m.once(function(){return a})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-utility', location = 'js/util/Configuration.js' */
define("dndattachment/util/Configuration",["require"],function(b){var a=b("wrm/data");var d="com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-issue-drop-zone.";function c(e){return a.claim(e)}return{getWRM:function(e){return c(d+e)}}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-utility', location = 'js/util/DataTransfer.js' */
define("dndattachment/util/DataTransfer",["require"],function(d){var a=d("jira/lib/class");var c=d("underscore");var b=d("jira/jquery/deferred");var e=d("jquery");return a.extend({init:function(f){this._dataTransfer=f},getFiles:function(){var g=this._dataTransfer.files;var f=this._dataTransfer.items;if(f&&f.length>0){if(c.any(f,function(h){return !!h.webkitGetAsEntry})){return this.readEntries(f)}}if(g&&g.length>0){return this.readFiles(this._dataTransfer.files)}else{return new b().reject()}},readFiles:function(g){var f=new b();e.when.apply(window,c.map(g,function(i){var h=new b();var j=!i.name.match(/\.([a-z0-9]+)$/i);if(j&&(i.size<=8192)||i.size==8192||i.size==4096||i.size<=1024){this.readFileAsText(i).fail(function(){g=c(g).without(i)}).always(h.resolve.bind(h))}else{h.resolve()}return h}.bind(this))).always(function(){f.resolve(g)});return f},readFileAsText:function(h){var g=new b();var f=new FileReader();f.onload=function(){g.resolve(this.result)};f.onerror=function(){g.reject(this.error)};f.readAsText(h);return g},readEntries:function(g){var f=new b();g=c.filter(g,function(h){return h.kind==="file"});e.when.apply(window,c.map(g,function(h){return this.readEntry(h.webkitGetAsEntry(),h)}.bind(this))).then(function(){f.resolve(c.union.apply(c,arguments))},f.reject.bind(f));return f},readEntry:function(h,g){var f=new b();if(h.isFile){if(g&&g.getAsFile){f.resolve([g.getAsFile()])}else{h.file(function(i){f.resolve([i])})}}else{if(h.isDirectory){h.createReader().readEntries(function(i){var j=[];e.when.apply(window,c.map(i,function(k){return this.readEntry(k).then(function(l){return j.push.apply(j,l)})}.bind(this))).always(function(){f.resolve(j)})}.bind(this))}}return f}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-utility', location = 'js/util/FileSizeUtil.js' */
define("dndattachment/util/FileSizeUtil",function(){var h=1024;var c=h*h;var d=" kB";var g=" MB";function f(i){if(i>c){return e(i)}else{if(i>h){return a(i)}else{return b(i)}}}function e(j){var i=j/c;return i.toFixed(2)+g}function a(j){var i=Math.round(j/h);return i+d}function b(j){var i=j/h;return i.toFixed(1)+d}return{format:f}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-utility', location = 'js/util/AttachmentEvents.js' */
define("dndattachment/util/events/types",{ATTACHMENT_FOR_PAGE_RECEIVED:"attachmentForPageReceived"});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-upload', location = 'js/upload/UploadHandler.js' */
define("dndattachment/upload/handler",["require","exports"],function(c,e){var i=c("jira/analytics");var j=c("jira/dialog/dialog");var k=c("jira/jquery/deferred");var n=c("jira/util/events");var m=c("dndattachment/util/events/types");var l=c("underscore");var f=c("jquery");var h=c("dndattachment/ctrlv/utility");var d=[];var a;var g=function(r,q){var p=k();var o=j.current;if(o&&o.$form){i.send({name:"attach.screenshot.html5.dialogPaste",data:{}});var t=h.createBlobFromFile(q.files[0]);if(h.dropFileToElement(t,o.$form)){p.resolve([q.files[0].name])}}else{var s=l.find(d,function(u){return u.isValid(r,q)});if(s){p=s.processFiles(q.files,a);if(q.successCallback){p.done(q.successCallback)}if(q.failureCallback){p.fail(q.failureCallback)}if(q.alwaysCallback){p.always(q.alwaysCallback)}}}p.done(function(v,u){if(!u&&q.isWikiTextfieldFocused){var w=q.wikiTextfield;l.each(v,function(x){h.insertWikiMarkup(x,w,w.selectionStart,w.selectionStart)});if(o&&q.isPaste){i.send({name:"attach.screenshot.html5.dialogPaste.insertWikiMarkup",data:{}})}if(!f(q.wikiTextfield).is(":focus")){setTimeout(function(){q.wikiTextfield.focus()},0)}}})};var b=function(s){if(s){var p=typeof s.isValid!=="undefined";var q=typeof s.processFiles!=="undefined";var o=typeof s.weight!=="undefined";var r=typeof s.name!=="undefined";return p&&q&&o&&r}return false};e.registerExecutor=function(q){var p=b(q);if(p){var o=l.reject(d,function(r){return r.name===q.name});o.push(q);d=l.sortBy(o,function(r){return -r.weight})}return p};e.unregisterExecutor=function(o){d=l.reject(d,function(p){return p.name===o.name})};e.initialize=function(){n.bind(m.ATTACHMENT_FOR_PAGE_RECEIVED,g)};e.disable=function(){n.unbind(m.ATTACHMENT_FOR_PAGE_RECEIVED,g)};e.setAttachmentDropZone=function(o){a=o}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-upload', location = 'js/upload/DefaultUploadExecutor.js' */
define("dndattachment/upload/default/executor",["require","exports"],function(b,a){a.name="Default attachment executor";a.weight=0;a.isValid=function(d,c){return true};a.processFiles=function(d,c){return c.uploadFiles(d)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/path.js' */
define("dndattachment/ctrlv/path",["exports"],function(a){var b=function b(c){return c.split(/\//).pop()};a.basename=b;a.dirname=function(c){var d=b(c);return c.substring(0,c.length-d.length)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'templates/soy/attach-screenshot-form.soy' */
// This file was automatically generated from attach-screenshot-form.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.Html5Screenshot.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.Html5Screenshot == 'undefined') { JIRA.Templates.Html5Screenshot = {}; }


JIRA.Templates.Html5Screenshot.attachScreenshotForm = function(opt_data, opt_ignored) {
  return '<h2 class="dialog-title">' + soy.$$escapeHtml('\u9644\u52a0\u9875\u9762\u622a\u56fe') + '</h2><form action="ShowAttachScreenshotFormAction.jspa" class="aui" method="post" id="attach-screenshot-form" data-attach-secure-token="' + soy.$$escapeHtml(opt_data.attachUserToken) + '" data-submit-secure-token="' + soy.$$escapeHtml(opt_data.submitUserToken) + '"><input type="hidden" name="id" value="' + soy.$$escapeHtml(opt_data.id) + '"/><input type="hidden" name="atl_token" value="' + soy.$$escapeHtml(opt_data.atlToken) + '"><input type="hidden" name="formToken" value="' + soy.$$escapeHtml(opt_data.formToken) + '"><input type="hidden" name="filetoconvert" value=""><div class="attach-screenshot-padding"><div class="attach-screenshot-container"><div id="attach-screenshot-inner-container"><div class="mod-content"><ul class="item-details"><li><dl><dt>' + soy.$$escapeHtml('\u622a\u5c4f:') + '</dt>' + ((opt_data.userPlatform == 'pc') ? '<dd>' + soy.$$filterNoAutoescape('\x3ckbd\x3ePrtScn\x3c/kbd\x3e') + '</dd>' : '<dd>' + soy.$$filterNoAutoescape('\x3ckbd\x3e\x26#8963; \u6309\u4f4f Ctrl\x3c/kbd\x3e+\x3ckbd\x3e\x26#8984; Cmd\x3c/kbd\x3e+\x3ckbd\x3e\x26#8679; Shift\x3c/kbd\x3e+\x3ckbd\x3e3\x3c/kbd\x3e') + '</dd>') + '</dl></li><li><dl><dt>' + soy.$$escapeHtml('\u7c98\u8d34\u56fe\u50cf:') + '</dt>' + ((opt_data.userPlatform == 'pc') ? '<dd>' + soy.$$filterNoAutoescape('\x3ckbd\x3eCtrl\x3c/kbd\x3e+\x3ckbd\x3ev\x3c/kbd\x3e') + '</dd>' : '<dd>' + soy.$$filterNoAutoescape('\x3ckbd\x3e\x26#8984; Cmd\x3c/kbd\x3e+\x3ckbd\x3ev\x3c/kbd\x3e') + '</dd>') + '</dl></li></ul></div><input type="text" id="attach-screenshot-fake-input"><div class="attach-screenshot-padding attach-screenshot-padding-inner"><div id="attach-screenshot-image-container" class="attach-screenshot-image-container"><div class="attach-screenshot-placeholder"><div class="mod-content"><ul class="item-details"><li><dl><dt id="attach-screenshot-placeholder-message">' + soy.$$escapeHtml('\u8bf7\u5c06\u56fe\u7247\u7c98\u8d34\u5728\u6b64\u5904') + '</dt></dl></li></ul></div></div></div></div><div id=\'attach-max-size\' class="hidden">' + soy.$$escapeHtml(opt_data.maxSize) + '</div></div></div></div><fieldset><div><legend><span>' + soy.$$escapeHtml('\u9644\u52a0\u9875\u9762\u622a\u56fe') + '</span></legend><div id="attach-screenshot-filename-group" class="field-group"><div id="attach-screenshot-progress-container"></div><label for="attachscreenshotname">' + soy.$$escapeHtml('\u6587\u4ef6\u540d') + ' <span class="aui-icon icon-required">' + soy.$$escapeHtml('\u5fc5\u9009\u9879') + '</span></label><input class="text" type="text" id="attachscreenshotname" name="attachscreenshotname" title="File Name" value="' + soy.$$escapeHtml(opt_data.nextScreenshotName) + '"><div class="description">' + soy.$$escapeHtml('\u6587\u4ef6\u540d\u7528\u4f5c\u9644\u52a0\u56fe\u7247\u540d') + '</div></div></div></fieldset><div class="buttons-container form-footer"><div class="buttons"><button class="aui-button aui-button-primary" id="attach-screenshot-html5-upload">' + soy.$$escapeHtml('\u4e0a\u4f20') + '</button><a href="#" class="cancel">' + soy.$$escapeHtml('\u53d6\u6d88') + '</a></div></div></form>';
};
if (goog.DEBUG) {
  JIRA.Templates.Html5Screenshot.attachScreenshotForm.soyTemplateName = 'JIRA.Templates.Html5Screenshot.attachScreenshotForm';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/html5.js' */
define("dndattachment/ctrlv/html5",["require","exports"],function(b,c){var k=b("jira/util/formatter");var e=b("wrm/context-path");var h=b("jira/jquery/deferred");var d=b("jquery");var m=b("underscore");var i=b("dndattachment/ctrlv/base64decode");var l=b("dndattachment/ctrlv/utility");var g=b("jira/dialog/dialog");var j=b("jira/dialog/dialog-register");var a=b("jira/ajs/ajax/smart-ajax");var f=b("jira/attachment/inline-attach");c.getFileFromEvent=function(o){var n=new h();if(l.isImagePasteEvent(o)){n.resolve(m(o.clipboardData.items).filter(function(p){return p.type.indexOf("image")!==-1}).map(function(p){return p.getAsFile()})[0]||m(o.clipboardData.files).filter(function(p){return p.type.indexOf("image")!==-1})[0])}else{n.reject()}return n.promise()};c.REQUEST_TIMEOUT_MILLIS=5*60*1000;c.screenshotFileUpload={};c.screenshotFileUploadUri=null;c.$document=undefined;c.$window=undefined;c.$container=undefined;c.$fakeInput=undefined;c.uploadError=false;c.uploadErrorMsg="";c.progressView={hidden:false,progress:0,old:0,progressBarContainer:undefined,progressEl:undefined,container:undefined,options:undefined,staticProgress:undefined,initProgress:function(){this.container=this.buildContainer();this.progressEl=this.buildProgress();this.container.append(this.progressEl);this.options={showPercentage:false,height:"5px"};this.progressBarContainer=d("#attach-screenshot-progress-container");this.progressBarContainer.empty();this.progressBarContainer.append(this.container);this.staticProgress=this.container;this.hidden=true},finish:function(){this.value(100)},progressHandler:function(o){var n=Math.round(o.loaded*100/o.total);this.value(n)},value:function(n){if(n>100){n=100}else{if(n<0){n=0}}if(this.hidden){this.progressEl.show();this.hidden=false}if(this.old!==n){this.progressEl.progressBar(n,this.options);if(n>=100){this.progressEl.fadeOut()}this.old=n}},buildContainer:function(){return d("<div>").addClass("file-progress")},buildProgress:function(){return d("<div>").attr("id","attach-screenshot-upload-progress")}};c.dialogView={pasteCatcher:{},presenter:undefined,getMaxSize:function(){return d("#attach-max-size").text()},getFileSize:function(){if(d.isPlainObject(this.presenter.screenshotFileUpload)&&m.isEmpty(this.presenter.screenshotFileUpload)){return 0}return this.presenter.screenshotFileUpload.size||this.presenter.screenshotFileUpload.byteLength||this.presenter.screenshotFileUpload.length},cleanGeneralErrors:function(){d("#error-attach-screenshot-image").closest(".field-group").remove()},cleanFileErrors:function(){d("#error-attach-screenshot-filename").remove()},displayErrors:function(n){if("compatibility" in n){l.showErrorMsg("",n.compatibility)}if("fileName" in n){l.showErrorMsg("",n.fileName)}if("fileUpload" in n){l.showErrorMsg("",n.fileUpload)}},appendBlobImage:function(n){var p=window.URL||window.webkitURL;var o=p.createObjectURL(n);this.presenter.screenshotFileUpload=n;this.createImage(o)},createImage:function(q){var o=new Image();o.onload=function(){};o.src=q;this.presenter.screenshotToUpload=o;var p=d(o);p.addClass("attach-screenshot-pasted-image");var n=d("#attach-screenshot-image-container");n.empty();n.append(p);this.presenter.$fakeInput.focus();m.defer(function(){c.imageCreatedHandler()})},checkInput:function(){var o,n=c.dialogView.pasteCatcher.childNodes[0];if(n){if("IMG"===n.tagName){if(n.src.indexOf("data:")===0){o=n.src}else{}}c.dialogView.pasteCatcher.innerHTML=""}if(!o){c.$fakeInput.focus()}return o},onPaste:function(n){c.dialogView.cleanFileErrors();c.dialogView.cleanGeneralErrors();if(l.isImagePasteEvent(n)){c.getFileFromEvent(n).then(function(o){this.appendBlobImage(o)}.bind(this))}else{setTimeout(function(){var o=this.checkInput();if(o){c.screenshotFileUpload=i.decodeBase64DataUri(o);c.dialogView.createImage(o)}}.bind(this),0)}},getFakeInput:function(){return d("#attach-screenshot-fake-input")},getContainer:function(){return d("#attach-screenshot-image-container")},getIssueKey:function(){return d("input[name='id']").val()},getDocument:function(){return d(document)},getWindow:function(){return d(window)},getFileNameInput:function(){return d("#attachscreenshotname")},hasPngExtension:function(o){var n=/\.png$/i;return n.test(o)},setFileToConvert:function(n){d("input[name='filetoconvert']").val(n)},buildPasteCatcher:function(){if(!document.getElementById("attach-screenshot-form")){return}var n=document.createElement("div");n.setAttribute("contenteditable","true");n.style.width=0;n.style.height=0;n.style.position="absolute";n.style.top="-5000px";document.getElementById("attach-screenshot-form").appendChild(n);return n},_getFormSubmits:function(){return d("#attach-screenshot-form").find("button.aui-button")},disable:function(){this._getFormSubmits().attr("disabled","disabled");return this},enable:function(){this._getFormSubmits().removeAttr("disabled");return this},isEnabled:function(){return this.isVisible()&&!this._getFormSubmits().attr("disabled")},isVisible:function(){return d("#attach-screenshot-form").length>0},initDialog:function(n){this.pasteCatcher={};this.presenter=n;this.pasteCatcher=this.buildPasteCatcher()}};c.initScreenshotPasteHandler=function(){var n=c.dialogView;c.screenshotFileUpload={};c.resetUploadErrors();c.dialogView.initDialog(c);c.$document=n.getDocument();c.$window=n.getWindow();c.$container=n.getContainer();c.$fakeInput=n.getFakeInput();c.bindOnce(c.$container,"click",c.setFocusOnClickHandler);c.bindOnce(c.$fakeInput,"focus",c.showFocusOnFieldHandler);c.bindOnce(c.$fakeInput,"blur",c.hideFocusOnFieldHandler);c.bindOnce(c.$fakeInput,"keydown",c.keyDownHandler);if(!l.browserIsSupported()){n.displayErrors({compatibility:k.format("\u4e0d\u652f\u6301\u60a8\u7684\u6d4f\u89c8\u5668\u3002\u8bf7\u8003\u8651\u4f7f\u7528\u66ff\u4ee3\u7684\u652f\u6301\u6d4f\u89c8\u5668\uff0c\u5982{0}\u8c37\u6b4cchrome{1}\u6216{2}\u706b\u72d0{3}\u3002",'<a href="//www.google.com/chrome/browser/">',"</a>",'<a href="//www.mozilla.org/firefox/">',"</a>")})}else{c.bindOnce(c.$window,"paste",c.pasteHandler)}};c.resetUploadErrors=function(){c.uploadError=false;c.uploadErrorMsg=undefined};c.setUploadError=function(n){c.uploadError=true;c.uploadErrorMsg=n};c.bindOnce=function(p,n,o){p.unbind(n,o);p.bind(n,o)};c.showFocusOnFieldHandler=function(){c.$container.addClass("focus")};c.hideFocusOnFieldHandler=function(){c.$container.addClass("focus")};c.setFocusOnClickHandler=function(){c.$fakeInput.focus()};c.pasteHandler=function(n){if(!c.dialogView.isEnabled()){return}n=l.normalizePasteEvent(n);c.dialogView.onPaste(n)};c.polyPasteHandler=function(o,n,p){if(!c.dialogView.isEnabled()){return}c.screenshotFileUpload={length:-1};c.screenshotFileUploadUri=n;c.dialogView.createImage(p)};c.keyDownHandler=function(n){if(l.isKeyPasteEvent(n)){if(c.dialogView.pasteCatcher.focus){c.dialogView.pasteCatcher.focus()}}};c.imageCreatedHandler=function(){c.doAjaxUpload(this.dialogView.getIssueKey(),this.dialogView.getFileNameInput().val())};c.getMimeType=function(){return"image/png"};c.createData=function(){return c.screenshotFileUpload};c.clipboardDataIsEmpty=function(n){if(window.clipboardData!=null){return false}return !(n&&n.clipboardData&&n.clipboardData.types&&n.clipboardData.types.length>0)};c.validateFileSize=function(q,n,p){var n=n||c.dialogView.getFileSize();var p=p||c.dialogView.getMaxSize();if(n>p){var o=f.Text.fileSize(p,n);q.fileUpload=k.format("\u9644\u52a0\u7684\u622a\u56fe\u592a\u5927\u3002\u9644\u4ef6\u662f{0}\uff0c\u4f46\u5141\u8bb8\u7684\u6700\u5927\u9644\u4ef6\u662f{1}\u3002",o[1],o[0])}};c.validateFormData=function(n,p){var o={};if(d.isPlainObject(n)&&m.isEmpty(n)){o.fileUpload="\u8bf7\u7c98\u8d34\u60a8\u5e0c\u671b\u4e0a\u4f20\u7684\u56fe\u7247"}c.validateFileSize(o);if(c.uploadError){o.fileUpload="\u56fe\u7247\u4e0a\u4f20\u65f6\u51fa\u9519\uff0c\u60a8\u9700\u8981\u518d\u6b21\u7c98\u8d34\u4e0a\u4f20\u56fe\u7247\u3002\u62b1\u6b49\u5e26\u6765\u7684\u4e0d\u4fbf\u3002"}if(""==p){o.fileName="\u60a8\u5fc5\u987b\u6307\u5b9a\u4e00\u4e2a\u56fe\u50cf\u6587\u4ef6\u540d\u3002"}if(!l.isValidFileName(p)){o.fileName="\u6587\u4ef6\u540d\u7981\u6b62\u5305\u542b\u5b57\u7b26\u201c\\\u201d\u3001\u201c/\u201d\u3001\u201c\"\u201d\u3001 \':\"\u3001\u201c?\u201d\u3001\u201c*\u201d\u3001\u201c\u003c\u201d\u3001\u201c|\u201d\u3001\u201c\u003e\u201d\u3001\u201c!\u201d"}return o};c.doAjaxUpload=function(n,u){c.dialogView.disable();var t={};c.validateFileSize(t);if(!m.isEmpty(t)){c.dialogView.displayErrors(t);c.dialogView.enable();return}var r=c.createData(),s=c.getMimeType(),p=d("#attach-screenshot-form").data("attach-secure-token"),o=d("#attach-screenshot-form").find("input[name='formToken']").attr("value");var q=e()+"/rest/internal/1.0/AttachTemporaryFile?size="+c.dialogView.getFileSize()+"&filename="+encodeURIComponent(u)+"&atl_token="+encodeURIComponent(atl_token())+"&issueId="+encodeURIComponent(this.dialogView.getIssueKey())+(p?("&secureToken="+encodeURIComponent(p)):"")+(o?("&formToken="+encodeURIComponent(o)):"");c.resetUploadErrors();c.executeAjaxUpload(r,q,s).progress(function(v){if(v=="init"){c.progressView.initProgress()}}).done(function(v){if(typeof v=="string"){v=JSON.parse(v)}c.dialogView.setFileToConvert(v.id)}).fail(function(x,A,z,v){var w;if(v.statusText=="abort"){w="\u547c\u53eb\u670d\u52a1\u5668\u7684 Jira \u6ca1\u6709\u5b8c\u6210\u5728\u8d85\u65f6\u65f6\u95f4\u6bb5\u5185\u3002\u6211\u4eec\u4e0d\u80fd\u786e\u5b9a\u8fd9\u9879\u884c\u52a8\u7684\u7ed3\u679c\u3002"}else{if(v.hasData){var y=JSON.parse(v.data);if(y.errorMessage){w=y.errorMessage}else{w=a.buildSimpleErrorContent(v,{alert:false})}}}c.setUploadError(w);c.dialogView.displayErrors({fileUpload:w})}).always(function(){c.dialogView.enable();c.progressView.finish()}).progress(function(v){if(v!="init"){c.progressView.progressHandler(v)}})};c.executeAjaxUpload=function(p,o,r){var n=h();var q=a.makeRequest({type:"POST",url:o,contentType:r,processData:false,data:p,timeout:c.REQUEST_TIMEOUT_MILLIS,success:n.resolve.bind(n),error:n.reject.bind(n),xhr:function(){var s=d.ajaxSettings.xhr();n.notify("init");s.upload.addEventListener("progress",n.notify.bind(n));return s}});n.always(function(){d.ajaxSettings.xhr().removeEventListener("progress",c.progressView.progressHandler);d(j.attachScreenshotDialog).off("Dialog.hide",q.abort)});d(j.attachScreenshotDialog).one("Dialog.hide",q.abort);return n.promise()};c.show=function(){var n=new h();var o=d(".issueaction-attach-screenshot-html5");if(g.current!=null){return n.reject().promise()}if(o.length==0){return n.reject().promise()}m.defer(function(){o.trigger("click")});d(document).on("dialogContentReady",function(q,p){if(p===j.attachScreenshotDialog&&c.dialogView.isEnabled()){n.resolve(p)}else{n.reject()}});d(document).on("ajaxComplete.jira.screenshot.dialog",function(r,q,p){if(p.url.indexOf(o.attr("href"))>-1){setTimeout(function(){if(!n.state()==="resolved"){n.reject()}},1000)}});n.always(function(){d(document).off("ajaxComplete.jira.screenshot.dialog")});return n.promise()}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/issue-paste.js' */
define("dndattachment/ctrlv/issue-paste",["require","exports"],function(h,D){var d=h("jquery");var v=h("dndattachment/ctrlv/trace");var e=h("dndattachment/ctrlv/tracking");var r=h("dndattachment/ctrlv/utility");var g=h("dndattachment/ctrlv/html5");var B=h("dndattachment/ctrlv/base64decode");var f=h("jira/dialog/dialog");var o=h("jira/util/events");var x=h("dndattachment/util/events/types");var A=0;var m=1;var c=2;var y=3;var p=4;var t=5;var i=6;var w=7;var q=8;var n=9;function l(G,F){F.stateName=G;var E=function(){var H=F.apply(null,arguments);H.stateName=G;return H};E.stateName=G;return E}function b(F,E,G){return l(F,function(){var J=setTimeout(function(){a(n)},G);var I=E.apply(null,arguments);var H=function(M,L){var K=I(M,L);if(K){clearTimeout(J);return K}};H.stateName=F;return H})}var k=l("idle",function(){function E(F){if(d(F.target).is(":input:not(.wiki-textfield, #summary)")||r.isContentEditable(F.target)){return}return true}return function(G,F){if(G==A&&r.isKeyPasteEvent(F)){if(r.browserIsNativePaste()||!E(F)||!r.browserIsSupported()){return}e.trigger("attach.screenshot.html5.catchClipboard");return new u(F)}if(G==m){if(r.isImagePasteEvent(F)){e.trigger("attach.screenshot.html5.handlePaste");if(!E(F)){v("jira/attach-images-plugin/pasteIgnored");return}if(d(F.target).is(":input.wiki-textfield")){F.preventDefault()}return new s(F)}else{v("jira/attach-images-plugin/pasteIgnoredNotImage")}}}});var u=b("catchClipboard",function(G){var E=document.activeElement;G.selectionStart=G.target.selectionStart;G.selectionEnd=G.target.selectionEnd;var H=d('<div contenteditable="true" class="attach-screenshot-paste-catcher"></div>').appendTo("body");H.focus();if(d(E).is(":focusable:input,:aui-focusable:input")){setTimeout(function(){E.focus()})}function F(){H.remove()}return function(L,I){if(L==m){if(r.isImagePasteEvent(I)){F();I.target=E;return new s(I)}if(r.isHtmlImagePasteEvent(I)){F();I.target=E;return new z(I)}if(r.isTextPasteEvent(I)){F();var O=r.getTextPasteContent(I);r.insertToInput(O,G.target,G.selectionStart,G.selectionEnd);I.preventDefault();return new k()}}if(L==n){F();var K=H.find(">img");if(K.is(":only-child")){var M=K.attr("src");if(M.toLowerCase().indexOf("http")===0){return k()}else{var P=B.decodeBase64DataUri(M);var J=new Blob([P],{type:"image/png"});return new j(J,G)}}else{var N=r.getTextContent(H[0]);r.insertToInput(N,G.target,G.selectionStart,G.selectionEnd);return k()}}}});var z=l("imageLoading",function(E){var F=r.getHtmlImagePaste(E);if(!F){a(t)}else{r.loadImage(F).then(function(G){a(p,G)},a.bind(null,t))}return function(I,G){if(I==p){var H=r.convertImageToBlob(G);if(H){return new j(H,E)}else{return new k()}}if(I==n||I==t){return new k()}}});var s=b("fileLoading",function(E){g.getFileFromEvent(E).done(function(F){a(c,F)}).fail(function(){a(y)});return function(G,F){if(G==c){return new j(F,E)}else{if(G==y||G==n){return new k()}}}},1000);var j=l("attachImage",function(G,F){var H=F.target;var I=r.generateFileName()+".png";var E=r.convertBlobToImage(G,I);o.trigger(x.ATTACHMENT_FOR_PAGE_RECEIVED,{files:[E],isWikiTextfieldFocused:r.isWikiTextfield(d(H)),wikiTextfield:H,isPaste:true,successCallback:function(){a(i)}});return function(K,J){return new k()}});var C=new k();function a(F,E){if(a.eventQueue){a.eventQueue.push({type:F,object:E});return}a.eventQueue=[{type:F,object:E}];while(a.eventQueue.length>0){var G=a.eventQueue.splice(0,1)[0];var H=C(G.type,G.object);if(H){C=H}}delete a.eventQueue}D._getStateMap=function(){return{events:{EVENT_WINDOW_KEYDOWN:A,EVENT_WINDOW_PASTE:m,EVENT_FILE_LOADED:c,EVENT_FILE_LOAD_ERROR:y,EVENT_IMAGE_LOADED:p,EVENT_IMAGE_LOAD_ERROR:t,EVENT_DIALOG_LOADED:i,EVENT_DIALOG_CLOSED:w,EVENT_DIALOG_CANCELED:q,EVENT_TIMEOUT:n},states:{StateIdle:k,StateCatchClipboard:u,StateImageLoading:z,StateFileLoading:s,StateAttachImage:j}}};D.initIssuePaste=function(){d(window).on("keydown",function(E){a(A,E)});d(window).on("paste",function(E){a(m,r.normalizePasteEvent(E))});o.bind("Dialog.hide",function(E,G,F){if(F){a(q,G)}else{a(w,G)}})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/trace.js' */
define("dndattachment/ctrlv/trace",["require"],function(b){var a=b("jira/util/logger");return a.trace});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/tracking.js' */
define("dndattachment/ctrlv/tracking",["require","exports"],function(c,b){var a=c("jira/analytics");b.trigger=function(e,d){a.send({name:e,data:d||{}})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/executor.js' */
define("dndattachment/ctrlv/executor",["require","exports"],function(e,c){var d=e("jira/jquery/deferred");var g=e("jquery");var a=e("jira/util/events");var b=e("dndattachment/ctrlv/html5");var f=e("dndattachment/upload/handler");c.register=function(){var h={name:"JIRA Ctrl+V attachment executor",weight:5,isValid:function(j,i){return !!i.isPaste},processFiles:function(k,j){var i=d();b.show().done(function(n){var m=g("#attach-screenshot-placeholder-message");var l=g.Event("paste");l.clipboardData={files:k};m.focus();setTimeout(function(){m.trigger(l)});a.bind("Dialog.hide",function(o,q,p){if(p){i.reject()}else{i.resolve([q.find("#attachscreenshotname").val()+".png"])}})});return i}};f.registerExecutor(h)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:jira-html5-attach-images-resources', location = 'js/ctrlv/initialize.js' */
define("dndattachment/ctrlv/initialize",["require","exports"],function(b,d){var h=b("jira/jquery/deferred");var e=b("jquery");var k=b("underscore");var l=b("dndattachment/ctrlv/utility");var f=b("dndattachment/ctrlv/html5");var g=b("dndattachment/ctrlv/issue-paste");var j=b("dndattachment/ctrlv/tracking");var c=b("dndattachment/ctrlv/executor");var i=b("jira/dialog/dialog-register");var a=b("jira/dialog/form-dialog");d.init=function(){var m=a.extend({options:{}});window.JIRA.ScreenshotDialog=m;g.initIssuePaste();e(document).ready(function(){var o=h();o.resolve();i.attachScreenshotDialog=new m({id:"attach-screenshot-dialog",trigger:".issueaction-attach-screenshot-html5",isIssueDialog:true,onContentRefresh:function n(){this.$form.bind("before-submit",function p(q){var r=f.validateFormData(f.screenshotFileUpload,e.trim(f.dialogView.getFileNameInput().val()));if(f.dialogView.getFileSize()==0){q.preventDefault();return false}else{if(!k.isEmpty(r)){f.dialogView.displayErrors(r);q.preventDefault();return false}}return true})},delayShowUntil:function(){return o}});e(document).bind("dialogContentReady",function(q,p){if(p===i.attachScreenshotDialog&&document.getElementById("attach-screenshot-form")!==null){j.trigger("attach.screenshot.html5.contentReady");f.initScreenshotPasteHandler()}});e(document).ready(function(){e(document).on("click","#attach-screenshot-html5",function(){j.trigger("attach.screenshot.html5.display")})})});c.register()}});require(["dndattachment/ctrlv/initialize","jquery"],function(b,a){a(function(){b.init()})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-amd', location = 'js/aui-amd.js' */
define("dndattachment/aui/dialog2",function(){return AJS.dialog2});define("dndattachment/aui/escape-html",function(){return AJS.escapeHtml});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:dnd-amd', location = 'js/dnd-template-wrapper.js' */
define("dndattachment/templates",function(){return window.JIRA.Templates.DnDAttachmentPlugin});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-dnd-attachment-plugin:attachment-menu-link-init', location = 'js/menu/initAttachmentMenuLink.js' */
require(["jquery","jira/analytics","jira/util/events","dndattachment/util/events/types"],function(c,f,j,i){var e="issueaction-attach-file";var a=function(k){k.addClass(e);k.click()};var b=function(){var k=c('<input type="file" multiple />');k.change(function(l){j.trigger(i.ATTACHMENT_FOR_PAGE_RECEIVED,{files:k[0].files})});k.click()};var h=function(k){return !!k.data("issuekey")};var g=function(){c(document).on("click",".unified-attach-file",function(m){f.send({name:"issue.dnd.attachment.opsbar.attachFiles.linkClick",data:{}});var k=c(m.target);if(k.hasClass(e)){return}m.preventDefault();var l=c(".issue-drop-zone").length;if(l&&!h(k)){b()}else{a(k)}})};var d=function(){g()};if(c.isReady){d()}else{c(d)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:common-runtime', location = 'runtime.25ce10de75f3804bffe1.js' */
!function(e){function n(n){for(var r,i,u=n[0],l=n[1],c=n[2],f=0,p=[];f<u.length;f++)i=u[f],o[i]&&p.push(o[i][0]),o[i]=0;for(r in l)Object.prototype.hasOwnProperty.call(l,r)&&(e[r]=l[r]);for(s&&s(n);p.length;)p.shift()();return a.push.apply(a,c||[]),t()}function t(){for(var e,n=0;n<a.length;n++){for(var t=a[n],r=!0,u=1;u<t.length;u++){var l=t[u];0!==o[l]&&(r=!1)}r&&(a.splice(n--,1),e=i(i.s=t[0]))}return e}var r={},o={52:0},a=[];function i(n){if(r[n])return r[n].exports;var t=r[n]={i:n,l:!1,exports:{}};return e[n].call(t.exports,t,t.exports,i),t.l=!0,t.exports}i.e=function(e){var n=[];return 0===o[e]?Promise.resolve():o[e]?o[e][2]:(n.push(new Promise(function(n,t){o[e]=[n,t]}),new Promise(function(n,t){WRM.require("wrc!com.atlassian.servicedesk.frontend-webpack-plugin:"+e).then(n,t)})),o[e][2]=Promise.all(n))},i.m=e,i.c=r,i.d=function(e,n,t){i.o(e,n)||Object.defineProperty(e,n,{enumerable:!0,get:t})},i.r=function(e){"undefined"!=typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},i.t=function(e,n){if(1&n&&(e=i(e)),8&n)return e;if(4&n&&"object"==typeof e&&e&&e.__esModule)return e;var t=Object.create(null);if(i.r(t),Object.defineProperty(t,"default",{enumerable:!0,value:e}),2&n&&"string"!=typeof e)for(var r in e)i.d(t,r,function(n){return e[n]}.bind(null,r));return t},i.n=function(e){var n=e&&e.__esModule?function(){return e.default}:function(){return e};return i.d(n,"a",n),n},i.o=function(e,n){return Object.prototype.hasOwnProperty.call(e,n)},i.p="",i.oe=function(e){throw console.error(e),e},function(e){if(e.__webpanel_register=e.__webpanel_register||[],!e.__webpanel_register.WEBPANEL_REGISTER_INITIALIZED){var n=e.__webpanel_register;function t(e){var t=n.callbacks[e],r=n.panelPayload[e];t&&r&&t(r.element,r.payload)}n.WEBPANEL_REGISTER_INITIALIZED=!0,n.callbacks={},n.panelPayload={};var r,o=[];function a(e){var a;n.panelPayload[e.id]={id:e.id,element:e.element,payload:e.payload},a=e.resources,o.push(a),clearTimeout(r),r=setTimeout(function(){WRM.require(o.map(function(e){return"wrc!"+e})),o=[]},42),t(e.id)}function i(e){var n;n=function(){a(e)},"complete"===document.readyState?setTimeout(n):document.addEventListener("DOMContentLoaded",n)}for(n.push=function(e){"sync"===e.type||"on-demand"===e.type?a(e):"async"===e.type&&i(e)};n.length;)n.push(n.pop());n.registerCallback=function(e,r){n.callbacks[e]=r,t(e)}}}("undefined"!=typeof window?window:{}),"undefined"!=typeof AJS&&(i.p=AJS.contextPath()+"/download/resources/com.atlassian.servicedesk.frontend-webpack-plugin:assets-7cee78fb-6fc1-478c-848f-43d8139723d4/");var u=window.jsdsjsonp=window.jsdsjsonp||[],l=u.push.bind(u);u.push=n,u=u.slice();for(var c=0;c<u.length;c++)n(u[c]);var s=l;t()}([]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_approvals-past-web-panel~approvals-pending-web-panel~atl.general~comment_fields_jsd~custom_field_jsd~2f871576', location = 'approvals-past-web-panel~approvals-pending-web-panel~atl.general~comment_fields_jsd~custom_field_jsd~2f871576.2a3dd7b3eef9e76e575a.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[53],{53:function(s,o,n){s.exports=void 0}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_approvals-past-web-panel~approvals-pending-web-panel~atl.general~comment_fields_jsd~custom_field_jsd~aad0728e', location = 'approvals-past-web-panel~approvals-pending-web-panel~atl.general~comment_fields_jsd~custom_field_jsd~aad0728e.09591cfae251c4335eea.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[54],[function(s,n,d){"use strict";d.r(n);d(53);n.default=jQuery}]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_atl.general~comment_fields_jsd~custom_field_jsd-organizations-field~custom_field_jsd-request-partici~3a735481', location = 'atl.general~comment_fields_jsd~custom_field_jsd-organizations-field~custom_field_jsd-request-partici~3a735481.48dd46a09479ed5484c3.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[56],{1:function(s,n,i){"use strict";i.r(n);var r=i(611),t=i.n(r);n.default=t.a},13:function(s,n){s.exports=require("jira/util/data/meta")},611:function(s,n){s.exports=require("atlassian/libs/underscore-1.5.2")}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_atl.general~comment_fields_jsd~custom_field_jsd-organizations-field~custom_field_jsd-request-partici~e23d54a1', location = 'atl.general~comment_fields_jsd~custom_field_jsd-organizations-field~custom_field_jsd-request-partici~e23d54a1.766ff8a29d2c4a523577.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[60],{11:function(e,a){e.exports=require("jira/util/logger")},159:function(e,a){e.exports=require("jira/analytics")},4:function(e,a,t){"use strict";var n=t(11),r=t.n(n),i=t(13),s=t.n(i),o=t(159),d=t.n(o),c=t(1),u=t(0),p=t(410),l={};l.trackPageview=function(e,a){var t=e.replace(/^\/?/,"").replace(/\//g,".");this.trackEvent("page.view."+t,a)},l.trackEvent=function(e,a,t){"string"==typeof a&&(e=[e,a].join("."),a=t),t||0===e.indexOf("servicedesk.")||(e="servicedesk."+e);var n={};if(void 0!==a)if(!c.default.isObject(a)||c.default.isArray(a)){var i="A tracking call was made with invalid properties. Data should be key-value pairs. Event name is '"+e+"' and data is: ";r.a.warn(i,a)}else c.default.extend(n,a);var o=Object(p.a)();n.sdVersion=o||(n.sdVersion||"unknown");var l=s.a&&s.a.get("projectId")?s.a.get("projectId"):Object(u.default)("body").data("projectId")||null;l&&(n.projectId=n.projectId||l);var v={name:e,properties:n};d.a.send(v)},a.a=l},410:function(e,a,t){"use strict";t.d(a,"a",function(){return i});var n=t(13),r=t.n(n),i=function(){return r.a.get("sdPluginVersion")}}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_atl.general~comment_fields_jsd~custom_field_jsd-organizations-field~custom_field_jsd-request-feedbac~43083c7e', location = 'atl.general~comment_fields_jsd~custom_field_jsd-organizations-field~custom_field_jsd-request-feedbac~43083c7e.b04c38a816fcbea15a3d.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[62],{15:function(s,n,i){"use strict";var a=i(610),o=i.n(a);n.a=o.a},610:function(s,n){s.exports=require("atlassian/libs/skate-0.12.6")}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_approvals-past-web-panel~approvals-pending-web-panel~atl.general~customerportal', location = 'approvals-past-web-panel~approvals-pending-web-panel~atl.general~customerportal.186499102e5e30204d0e.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[95],{409:function(s,o,n){s.exports=void 0}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_atl.general~customerportal~sd.project.admin', location = 'atl.general~customerportal~sd.project.admin.f87e825686ddd63ab356.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[111],{111:function(t,a,n){"use strict";n.d(a,"a",function(){return o});var c=n(1),s=n(0),e=n(15),d=n(4),o=c.default.once(function(){Object(e.a)("data-analytics-key",{type:e.a.types.ATTR,attached:function(t){var a=Object(s.default)(t);a.click(function(){var t={},n=a.data("analytics-key");-1!==n.substr(".help.docs.")&&(t.docsUrl=a.attr("href")),d.a.trackEvent(n,t)})}})})}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_atl.general~customerportal', location = 'atl.general~customerportal.4f4ed7a4577bb91efea5.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[129],{186:function(t,e,a){"use strict";a.d(e,"a",function(){return d});a(409);var s=a(1),n=a(0),i=a(15);a(848);var d=function(){Object(i.a)("js-tooltip",{type:i.a.types.CLASS,created:function(t){var e=Object(n.default)(t),a=e.data();t.tipsyClassName=s.default.uniqueId("sd-tipsy-tooltip-"),a.className=t.tipsyClassName,function(t){if(t.is("button")){var e=t.attr("title");t.removeAttr("title");var a=Object(n.default)("<span />").addClass("sd-tipsy-btn-container").insertBefore(t);return t.appendTo(a),t.is(":disabled")?Object(n.default)("<span />").addClass("sd-tipsy-btn-blanket").attr("title",e).appendTo(a):(a.attr("title",e),a)}return t}(e).tooltip(a)},detached:function(t){Object(n.default)("."+t.tipsyClassName).remove()}})}},848:function(t,e,a){}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.servicedesk.frontend-webpack-plugin:split_atl.general', location = 'atl.general.cf00b6a81d2909f96be5.js' */
(window.jsdsjsonp=window.jsdsjsonp||[]).push([[178],{849:function(s,j,n){"use strict";n.r(j);var o=n(111),t=n(186);Object(t.a)(),Object(o.a)()}},[[849,52,53,54,56,60,62,95,111,129]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-analytics', location = '/includes/jira/analytics/analytics-init.js' */
require(["jquery","jira/analytics","jira/util/data/meta"],function(a,t,c){"use strict";function n(a){var t=c.get("analytics-prefix");return t&&(a=t+a),a}function i(){a("[data-track-pageview]").each(function(c,i){t.send({name:n(a(i).data("track-pageview"))})})}a(document).on("click","[data-track-click]",function(){t.send({name:n(a(this).data("track-click"))})}),a(document).on("auxclick","[data-track-auxclick]",function(){t.send({name:n(a(this).data("track-auxclick"))})}),a(i)});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/jquery/plugins/avataror/avataror.js' */
jQuery.fn.avataror=function(i){var e=jQuery,t=e(document);this.each(function(){var r=e(this),a=r.find("img").attr("src");r.css({"-moz-border-radius":"10px","-webkit-border-radius":"10px"}),r.html("<p>Loading?</p>");var g={previewSize:48};g.preview=e("<div/>").addClass("avataror-preview").css({border:"solid 1px #000",float:"left",height:g.previewSize+"px",overflow:"hidden",width:g.previewSize+"px",position:"relative",top:"-9999em",left:"-9999em"}),g.preview.prependTo(i.previewElement),g.img=e('<img alt="Avatar Source"/>'),g.img.load(function(){g.image=e("<div/>").css({background:"url('"+a+"') no-repeat",clear:"left",position:"relative"}),g.marker=e("<div/>").css({cursor:"move",position:"relative"}),g.dash=e("<div/>"),g.shadow=e("<div/>"),g.dash.add(g.shadow).css({cursor:"move",opacity:.5,left:0,top:0,position:"absolute"}),g.image.append(g.shadow).append(g.dash).append(g.marker),r.append(g.image),g.marker.html("<div></div><div></div><div></div><div></div>"),e("div",g.marker).each(function(i){var t=e(this);t.css({background:"#000",border:"solid 1px #fff",width:"10px",height:"10px",position:"absolute","font-size":"1px"}),t.css(["left","right","right","left"][i],"-6px"),t.css(["top","top","bottom","bottom"][i],"-6px"),t.css("cursor",["nw-resize","ne-resize","se-resize","sw-resize"][i]),t.mousedown(function(e){e.preventDefault(),e.stopPropagation(),g.dragging={x:e.pageX,y:e.pageY,ax:g.x,ay:g.y,w:g.width,h:g.height,i:i+1},g.shadow.hide()})}),g.marker.add(g.image).mousedown(function(i){i.preventDefault(),g.dragging={x:i.pageX,y:i.pageY,ax:g.x,ay:g.y,w:g.width,h:g.height},g.shadow.hide()}),t.mouseup(function(i){g.handleMouseUp(i)}),t.mousemove(function(i){g.dragging&&(g.handleMouseMove(i.pageX,i.pageY),i.preventDefault())}),g.imgwidth=g.img[0].naturalWidth,g.imgheight=g.img[0].naturalHeight,g.x=parseInt(e("#avatar-offsetX").val()),g.y=parseInt(e("#avatar-offsetY").val()),g.width=parseInt(e("#avatar-width").val()),g.height=g.width,g.image.css({width:g.imgwidth+"px",height:g.imgheight+"px"}),g.setMarker(),r.css({width:g.imgwidth+"px"}),g.preview.css({position:"static"}),e("p",r).remove(),r.trigger("AvatarImageLoaded"),g.adjustPreview()}),g.img.attr("src",a),g.preview.append(g.img),g.setMarker=function(){g.marker.css("border","dashed 1px #fff"),g.dash.css("border","solid 1px #000"),g.shadow.css("border","solid 1px #000"),g.marker.add(this.dash).css("left",this.x-1+"px"),g.marker.add(g.dash).css("top",g.y-1+"px"),g.shadow.css("border-left-width",g.x+"px"),g.shadow.css("border-right-width",g.imgwidth-g.x-g.width+"px"),g.shadow.css("border-top-width",g.y+"px"),g.shadow.css("border-bottom-width",g.imgheight-g.y-g.height+"px"),g.shadow.css("width",g.width+"px"),g.shadow.css("height",g.height+"px"),g.marker.add(g.dash).css("width",g.width+"px"),g.marker.add(g.dash).css("height",g.height+"px")},g.adjustPreview=function(){g.img.attr("width",g.imgwidth*g.previewSize/g.width),g.img.attr("height",g.imgheight*g.previewSize/g.height),g.img.css("margin-left","-"+g.x*g.previewSize/g.width+"px"),g.img.css("margin-top","-"+g.y*g.previewSize/g.height+"px"),g.preview.select()},g.handleMouseMove=function(i,e){if(g.dragging.nextExec=g.dragging.nextExec||0,0!=g.dragging.nextExec)return void g.dragging.nextExec--;g.dragging.nextExec=3;var t=i-g.dragging.x,r=e-g.dragging.y;if(this.dragging.i){(0,g.resizeHandlers[this.dragging.i-1])(t,r)}else g.x=g.dragging.ax+t,g.y=g.dragging.ay+r,g.x+g.width>g.imgwidth&&(g.x=g.imgwidth-g.width),g.y+g.height>g.imgheight&&(g.y=g.imgheight-g.height),g.x<0&&(g.x=0),g.y<0&&(g.y=0);g.setMarker(),g.adjustPreview()},g.handleMouseUp=function(i){e("#avatar-offsetX").val(g.x),e("#avatar-offsetY").val(g.y),e("#avatar-width").val(g.width),g.dragging=null,g.shadow.show()},g.originX=function(){return g.dragging.ax},g.originY=function(){return g.dragging.ay},g.originBottomX=function(){return g.dragging.ax+g.dragging.w},g.originBottomY=function(){return g.dragging.ay+g.dragging.h},g.originNw=function(){return{x:g.originX(),y:g.originY()}},g.originNe=function(){return{x:g.originBottomX(),y:g.originY()}},g.originSe=function(){return{x:g.originBottomX(),y:g.originBottomY()}},g.originSw=function(){return{x:g.originX(),y:g.originBottomY()}},g.nwHandler=function(i,e){var t=g.originSe(),r={x:g.originX()+i,y:g.originY()+e},a=Math.abs(r.x-t.x),o=Math.abs(r.y-t.y),n=Math.min(a,o);n<20&&(n=20),t.x-n<0&&(n=t.x),t.y-n<0&&(n=t.y),g.x=t.x-n,g.y=t.y-n,g.width=g.height=n},g.neHandler=function(i,e){var t=g.originSw(),r={x:g.originBottomX()+i,y:g.originY()+e},a=Math.abs(r.x-t.x),o=Math.abs(r.y-t.y),n=Math.min(a,o);n<20&&(n=20),t.x+n>g.imgwidth&&(n=g.imgwidth-t.x),t.y-n<0&&(n=t.y),g.y=t.y-n,g.width=g.height=n},g.seHandler=function(i,e){var t=g.originNw(),r={x:g.originBottomX()+i,y:g.originBottomY()+e},a=Math.abs(r.x-t.x),o=Math.abs(r.y-t.y),n=Math.min(a,o);n<20&&(n=20),t.x+n>g.imgwidth&&(n=g.imgwidth-t.x),t.y+n>g.imgheight&&(n=g.imgheight-t.y),g.width=g.height=n},g.swHandler=function(i,e){var t=g.originNe(),r={x:g.originX()+i,y:g.originBottomY()+e},a=Math.abs(r.x-t.x),o=Math.abs(r.y-t.y),n=Math.min(a,o);n<20&&(n=20),t.x-n<0&&(n=t.x),t.y+n>g.imgheight&&(n=g.imgheight-t.y),g.x=t.x-n,g.width=g.height=n},g.resizeHandlers=[g.nwHandler,g.neHandler,g.seHandler,g.swHandler]})};
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/Avatar.js' */
var _typeof="function"==typeof Symbol&&"symbol"==typeof Symbol.iterator?function(t){return typeof t}:function(t){return t&&"function"==typeof Symbol&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t};define("jira/ajs/avatarpicker/avatar",["jira/lib/class"],function(t){return t.extend({init:function(t){this._id=t.id,this._isSystemAvatar=t.isSystemAvatar,this._isSelected=t.isSelected,this._urls=t.urls},setUnSelected:function(){this._isSelected=!1},setSelected:function(){this._isSelected=!0},isSelected:function(){return!!this._isSelected},isSystemAvatar:function(){return this._isSystemAvatar},getId:function(){return this._id},getUrl:function(t){return this._urls[t]},toJSON:function(){return{id:this._id,isSystemAvatar:this._isSystemAvatar,isSelected:this._isSelected,urls:this._urls}}})}),define("jira/ajs/avatarpicker/avatar-factory",["jira/ajs/avatarpicker/avatar","exports"],function(t,e){e.createCustomAvatar=function(e){return e.isSystemAvatar=!1,new t(e)},e.createSystemAvatar=function(e){return e.isSystemAvatar=!0,new t(e)}}),define("jira/ajs/avatarpicker/avatar/sizes",["exports","jquery"],function(t,e){t.getSizeObjectFromName=function(i){if("object"===(void 0===i?"undefined":_typeof(i)))return i;"string"==typeof i&&e.trim(i);return t.LARGE.param===i?t.LARGE:t.MEDIUM.param===i?t.MEDIUM:t.SMALL.param===i?t.SMALL:"xsmall"===i?t.SMALL:t.LARGE},t.LARGE={param:"large",height:48,width:48},t.MEDIUM={param:"medium",width:32,height:32},t.SMALL={param:"small",width:16,height:16}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/AvatarManager.js' */
define("jira/ajs/avatarpicker/avatar-manager",["jira/util/formatter","jira/lib/class","wrm/context-path","jquery"],function(t,a,r,e){"use strict";var s=r();return a.extend({init:function(t){this.store=t.store,this.ownerId=t.ownerId,this.username=t.username,this.anonymousAvatarId=t.anonymousAvatarId,this.avatarSrcBaseUrl=t.avatarSrcBaseUrl},selectAvatar:function(t,a){return this.store.selectAvatar(t,a)},getById:function(t){return this.store.getById(t)},destroy:function(t,a){this.store.destroy(t,a)},update:function(t,a){this.store.update(t,a)},add:function(t,a){this.store._add(t,a)},getAllSystemAvatars:function(){return this.store.getAllSystemAvatars()},getAllCustomAvatars:function(){return this.store.getAllCustomAvatars()},getSelectedAvatar:function(){return this.store.getSelectedAvatar()},getAllAvatars:function(){return this.store.getAllAvatars()},getAllAvatarsRenderData:function(t){var a,r=this,e=this.getAllAvatars(),s={system:[],custom:[]};for(a=0;a<e.system.length;a++)s.system.push(r.getAvatarRenderData(e.system[a],t));for(a=0;a<e.custom.length;a++)s.custom.push(r.getAvatarRenderData(e.custom[a],t));return s},getAvatarRenderData:function(t,a){var r=t.toJSON();return r.src=this.getAvatarSrc(t,a),r.width=a.width,r.height=a.height,r},refreshStore:function(t){this.store.refresh(t)},getAvatarSrc:function(a,r){return this.store.isTempAvatar(a)?s+"/secure/temporaryavatar?"+e.param({cropped:!0,magic:(new Date).getTime(),size:r.param}):a.getUrl(t.format("{0}x{1}",r.height,r.width))},createTemporaryAvatar:function(t,a){this.store.createTemporaryAvatar(t,a)},createAvatarFromTemporary:function(t,a){this.store.createAvatarFromTemporary(t,a)},getAnonymousAvatarId:function(){return this.anonymousAvatarId}})}),define("jira/ajs/avatarpicker/avatar-manager-factory",["jira/ajs/avatarpicker/avatar-store","jira/ajs/avatarpicker/avatar-manager","wrm/context-path","exports"],function(t,a,r,e){"use strict";var s=r();e.createUniversalAvatarManager=function(r){var e,n="",o="",i="",v="";if(r.projectId){var u=s+"/rest/api/latest/universal_avatar/type/"+r.avatarType+"/owner/"+r.projectId;e=u;var c=u+"/avatar";n=null,o=u+"/temp",i=c,v=c}else e=s+"/rest/api/latest/avatar/project/system",o=s+"/rest/api/latest/avatar/project/temporary",i=s+"/rest/api/latest/avatar/project/temporaryCrop";var l=new t({restQueryUrl:e,restUpdateUrl:n,restCreateTempUrl:o,restUpdateTempUrl:i,restSingleAvatarUrl:v,defaultAvatarId:r.defaultAvatarId});return new a({store:l,ownerId:r.projectId,avatarSrcBaseUrl:s+"/secure/projectavatar"})},e.createProjectAvatarManager=function(t){return t.avatarType="project",e.createUniversalAvatarManager(t)},e.createUserAvatarManager=function(r){var e=s+"/rest/api/latest/user",n=new t({restQueryUrl:e+"/avatars",restUpdateUrl:e+"/avatar",restCreateTempUrl:e+"/avatar/temporary",restUpdateTempUrl:e+"/avatar",restSingleAvatarUrl:e+"/avatar",restParams:{username:r.username},defaultAvatarId:r.defaultAvatarId});return new a({store:n,username:r.username,avatarSrcBaseUrl:s+"/secure/useravatar"})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/AvatarStore.js' */
define("jira/ajs/avatarpicker/avatar-store",["jira/util/urls","jira/util/formatter","jira/ajs/avatarpicker/avatar-util","jira/ajs/avatarpicker/avatar-factory","jira/ajs/ajax/smart-ajax","jira/lib/class","jquery"],function(t,e,r,a,s,i,o){return i.extend({TEMP_ID:"TEMP",init:function(e){if(!e.restQueryUrl)throw new Error("JIRA.AvatarStore.init: You must specify [restQueryUrl], The rest url for querying avatars (see class description)");if(!e.restCreateTempUrl)throw new Error("JIRA.AvatarStore.init: You must specify [restCreateTempUrl], The rest url for creating a temporary avatar (see class description)");if(!e.restUpdateTempUrl)throw new Error("JIRA.AvatarStore.init: You must specify [restUpdateTempUrl], The rest url for updating a temporary avatar (see class description)");if(!e.defaultAvatarId)throw new Error("JIRA.AvatarStore.init: You must specify [defaultAvatarId] to the contructor so the store knows what to select if you delete the selected one");this.restQueryUrl=e.restQueryUrl,this.restUpdateUrl=e.restUpdateUrl,this.restCreateTempUrl=e.restCreateTempUrl,this.restUpdateTempUrl=e.restUpdateTempUrl,this.restSingleAvatarUrl=e.restSingleAvatarUrl,this.restParams=e.restParams||{},this.restParams.atl_token=t.atl_token(),this.defaultAvatarId=e.defaultAvatarId,this.avatars={system:[],custom:[]}},_buildCompleteUrl:function(t){var r=t;if(this.restParams){var a="";for(var s in this.restParams)a+=e.format("&{0}={1}",encodeURIComponent(s),encodeURIComponent(this.restParams[s]));r+="?"+a.substr(1)}return r},getById:function(t){var e;return o.each(this.avatars.system,function(r,a){if(this.getId()===t)return e=a,!1}),e||o.each(this.avatars.custom,function(r,a){if(this.getId()===t)return e=a,!1}),e},isTempAvatar:function(t){return t.getId()===this.TEMP_ID},_selectAvatar:function(t){var e=this.getSelectedAvatar();e&&e.setUnSelected(),t.setSelected()},selectAvatar:function(t,e){var r=this;if(!t)throw new Error("JIRA.AvatarStore.selectAvatar: Cannot select Avatar that does not exist");this.restUpdateUrl?s.makeRequest({type:"PUT",contentType:"application/json",dataType:"json",url:this._buildCompleteUrl(this.restUpdateUrl),data:JSON.stringify(t.toJSON()),success:function(){r._selectAvatar(t),e.success&&e.success.call(this,t)},error:e.error}):(r._selectAvatar(t),e.success&&e.success.call(this,t))},_destory:function(t){var e=o.inArray(t,this.avatars.custom);if(-1===e)throw new Error("JIRA.AvatarStore._destroy: Cannot remove avatar ["+t.getId()+"], it might be a system avatar (readonly) or does not exist.");this.avatars.custom.splice(e,1)},destroy:function(t,e){var r=this;if(e=e||{},!t)throw new Error("JIRA.AvatarStore.destroy: Cannot delete Avatar that does not exist");s.makeRequest({type:"DELETE",url:this.getRestUrlForAvatar(t),success:function(){r._destory(t),t.isSelected()?r.selectAvatar(r.getById(r.defaultAvatarId),e):e.success&&e.success.apply(this,arguments)},error:e.error})},getSelectedAvatar:function(){for(var t=0;t<this.avatars.custom.length;t++)if(this.avatars.custom[t].isSelected())return this.avatars.custom[t];for(t=0;t<this.avatars.system.length;t++)if(this.avatars.system[t].isSelected())return this.avatars.system[t]},_update:function(t){var e=this;if(!this.getById(t.getId()))throw new Error("JIRA.AvatarStore._update: Cannot update avatar ["+t.getId()+"], it might be a system avatar (readonly) or does not exist.");o.each(this.avatars.custom,function(r){this.getId()===t.getId()&&(e.avatars.custom[r]=t)})},update:function(t,e){var r=this;e=e||{},s.makeRequest({type:"PUT",url:this.getRestUrlForAvatar(t),error:e.error,success:function(){r._update(t),e.success&&e.success.apply(this,arguments)}})},_add:function(t){t.isSystemAvatar()?this.avatars.system.push(t):this.avatars.custom.push(t)},createAvatarFromTemporary:function(t,e){var r=this;e=e||{},this.restUpdateTempUrl&&s.makeRequest({type:"POST",url:this._buildCompleteUrl(this.restUpdateTempUrl),data:JSON.stringify(t),contentType:"application/json",dataType:"json",success:function(t){t||(t={id:r.TEMP_ID,isSelected:!0});var s=a.createCustomAvatar(t);r._add(s),e.success&&e.success.call(this,t)},error:e.error})},createTemporaryAvatar:function(t,e){e=o.extend(!0,{},e,{params:this.restParams}),r.uploadTemporaryAvatar(this.restCreateTempUrl,t,e)},_refresh:function(t){var e=this;e.avatars.system=[],e.avatars.custom=[],t.system&&o.each(t.system,function(t,r){e.avatars.system.push(a.createSystemAvatar(r))}),t.custom&&o.each(t.custom,function(t,r){e.avatars.custom.push(a.createCustomAvatar(r))})},refresh:function(t){var e=this,r=this.getById(e.TEMP_ID);t=t||{},s.makeRequest({url:this._buildCompleteUrl(this.restQueryUrl),error:t.error,success:function(a){e._refresh(a),r&&e._add(r),t.success&&t.success.apply(this,arguments)}})},getAllAvatars:function(){return this.avatars},getAllSystemAvatars:function(){return this.avatars.system},getAllCustomAvatars:function(){return this.avatars.custom},getRestUrlForAvatar:function(t){return this._buildCompleteUrl(this.restSingleAvatarUrl+"/"+t.getId())}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/AvatarUtil.js' */
define("jira/ajs/avatarpicker/avatar-util",["jira/util/urls","jira/util/logger","jira/attachment/inline-attach","jquery"],function(e,r,a,n){var t={uploadUsingIframeRemoting:function(o,s,i){i=i||{};var l=s.val(),c=new a.Form(new a.FileInput(s,!1)),u=c.addStaticProgress(l),d=c.cloneFileInput();c.fileSelector.clear();var g=new a.Timer(function(){!this.cancelled&&u.show()},this),f=new a.FormUpload({$input:d,url:o,params:n.extend({},i.params,{filename:l,atl_token:e.atl_token()}),scope:this,before:function(){!this.cancelled&&u.start()},success:function(e,r){e.errorMessages&&e.errorMessages.length?c.addErrorWithFileName(e.errorMessages[0],l,t.getErrorTarget(c)):i.success&&i.success(e,r)},error:function(e){r.log(e),this.cancelled||(e.indexOf("SecurityTokenMissing")>=0?c.addError(a.Text.tr("upload.xsrf.timeout",l),t.getErrorTarget(c)):c.addError(a.Text.tr("upload.error.unknown",l),t.getErrorTarget(c)))},after:function(){g.cancel(),u.remove(),this.cancelled||c.enable()}});u.onCancel(function(){f.abort()}),f.upload()},uploadUsingFileApi:function(r,o,s){var i,l,c=o[0].files[0],u=new a.Form(new a.FileInput(o,!1)),d=u.addProgress(c);s=s||{},i=new a.Timer(function(){d.show()}),l=new a.AjaxUpload({file:c,params:n.extend({},s.params,{filename:c.name,size:c.size,atl_token:e.atl_token()}),scope:this,url:r,before:function(){o.hide(),d.start()},progress:function(e){d.progress.$progress.parent().parent().show(),d.update(e)},success:function(e,r){e.errorMessages&&e.errorMessages.length?u.addErrorWithFileName(e.errorMessages[0],c.name,t.getErrorTarget(u)):201===r&&s.success(e,r)},error:function(e,r){r<0?u.addError(e,t.getErrorTarget(u)):u.addError(a.Text.tr("upload.error.unknown",c.name),t.getErrorTarget(u)),s.error&&s.error(e,r)},after:function(){i.cancel(),d.finish().remove(),o.val("").show()}}),l.upload(),d.onCancel(function(){l.abort()})},getErrorTarget:function(e){return{$element:e.$form.find(".error")}},uploadTemporaryAvatar:function(e,r,n){a.AjaxPresenter.isSupported(r)?this.uploadUsingFileApi(e,r,n):this.uploadUsingIframeRemoting(e,r,n)}};return t});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/AvatarPicker.js' */
define("jira/ajs/avatarpicker/avatar-picker",["jira/util/logger","jira/util/formatter","jira/util/strings","jira/ajs/avatarpicker/avatar-picker/image-editor","jira/ajs/ajax/smart-ajax","jira/dialog/dialog","jira/ajs/control","aui/message","jquery"],function(a,e,t,r,i,n,c,o,d){"use strict";return c.extend({init:function(a){this.avatarManager=a.avatarManager,this.avatarRenderer=a.avatarRenderer,this.imageEditor=a.imageEditor,this.size=a.size,this.selectCallback=a.select,this.cropperDialog=null,this.initialSelection=a.initialSelection},render:function(a){var e=this;this.avatarManager.refreshStore({success:function(){e.cropperDialog instanceof n&&(e.cropperDialog.hide(),delete e.cropperDialog),e.element=d('<div id="jira-avatar-picker" />'),e.element.html(JIRA.Templates.AvatarPicker.picker({avatars:e.avatarManager.getAllAvatarsRenderData(e.size)})),e._assignEvents("selectAvatar",e.element.find(".jira-avatar button")),e._assignEvents("deleteAvatar",e.element.find(".jira-delete-avatar")),e._assignEvents("uploader",e.element.find("#jira-avatar-uploader")),void 0!==e.initialSelection&&e.getAvatarElById(e.initialSelection).addClass("jira-selected-avatar"),a(e.element)},error:function(t,r,i,n){e.appendErrorContent(e.element,n),a(e.element)}})},appendErrorContent:function(a,e){try{var r=JSON.parse(e.data);r&&r.errorMessages?d.each(r.errorMessages,function(e,r){o.error(a,{body:t.escapeHtml(r),closeable:!1,shadowed:!1})}):a.append(i.buildDialogErrorContent(e,!0))}catch(t){a.append(i.buildDialogErrorContent(e,!0))}},uploadTempAvatar:function(e){var t=this;this.avatarManager.createTemporaryAvatar(e,{success:function(a){a.id?t.render(function(){t.selectAvatar(a.id)}):(e.val(""),t.cropperDialog=new n({id:"project-avatar-cropper",width:560,content:function(e){function r(){var a=i.find("input[type=submit]"),e=d("<span class='icon throbber loading'></span>");return a.attr("aria-disabled","true").attr("disabled",""),a.before(e),function(){e.remove(),a.removeAttr("aria-disabled").removeAttr("disabled")}}var i=t.imageEditor.render(a);t.imageEditor.edit({confirm:function(a){var e=r();t.avatarManager.createAvatarFromTemporary(a,{success:function(a){t.render(function(){t.selectAvatar(a.id)})},error:e})}}),i.find(".cancel").click(function(){t.cropperDialog.hide()}),e(i)}}),t.cropperDialog.bind("dialogContentReady",function(){d(t).trigger(r.LOADED)}),t.cropperDialog.bind("Dialog.hide",function(){d(t).trigger(r.DISMISSED)}),t.cropperDialog.show())},error:function(){a.log(arguments)}})},getAvatarElById:function(a){return this.element.find(".jira-avatar[data-id='"+a+"']")},selectAvatar:function(a){var e=this.avatarManager.getById(a),t=this;this.avatarManager.selectAvatar(this.avatarManager.getById(a),{error:function(){},success:function(){t.getAvatarElById(a).remove(),t.selectCallback&&t.selectCallback.call(t,e,t.avatarManager.getAvatarSrc(e,t.size))}})},deleteAvatar:function(a){var t=this;confirm("\u786e\u5b9e\u8981\u5220\u9664\u8fd9\u4e2a\u56fe\u50cf\uff1f")&&this.avatarManager.destroy(this.avatarManager.getById(a),{error:function(){},success:function(){var e=t.avatarManager.getSelectedAvatar(),r=t.getAvatarElById(a);r.fadeOut(function(){r.remove()}),e.getId()!==a&&(t.getAvatarElById(e.getId()).addClass("jira-selected-avatar"),t.selectCallback.call(t,e,t.avatarManager.getAvatarSrc(e,t.size),!0))}})},_events:{uploader:{change:function(a,e){this.uploadTempAvatar(e)}},deleteAvatar:{click:function(a,e){this.deleteAvatar(e.attr("data-id"))}},selectAvatar:{click:function(a,e){"button"===e[0].tagName.toLowerCase()&&this.selectAvatar(e.attr("data-id"))}}}})}),define("jira/ajs/avatarpicker/avatar-picker/image-editor",["jira/ajs/control","jquery"],function(a,e){"use strict";var t=a.extend({render:function(a){return this.element=e('<div id="avatar-picker-image-editor"/>').html(JIRA.Templates.AvatarPicker.imageEditor(a)),this.element},edit:function(a){var t=this,r=this.element.find(".avataror");a=a||{},r.unbind(),r.bind("AvatarImageLoaded",function(){a.ready&&a.ready()}),r.find("img").load(function(){r.avataror({previewElement:t.element.find(".jira-avatar-cropper-header"),parent:t.element})}),this.element.find("#avataror").submit(function(t){t.preventDefault(),a.confirm&&a.confirm({cropperOffsetX:e("#avatar-offsetX").val(),cropperOffsetY:e("#avatar-offsetY").val(),cropperWidth:e("#avatar-width").val()})}).find(".cancel").click(function(e){e.preventDefault(),a.cancel&&a.cancel()})}});return t.LOADED="imageEditorLoaded",t.DISMISSED="imageEditorDismissed",t}),define("jira/ajs/avatarpicker/avatar-picker-factory",["jira/util/browser","jira/util/formatter","jira/ajs/avatarpicker/avatar-picker","jira/ajs/avatarpicker/avatar-picker/image-editor","jira/ajs/avatarpicker/avatar-manager-factory","jira/ajs/avatarpicker/avatar/sizes","jira/dialog/form-dialog","wrm/context-path","wrm/data","jquery","exports"],function(a,e,t,r,i,n,c,o,d,l,s){"use strict";s.createUniversalAvatarPicker=function(a){return new t({avatarManager:i.createUniversalAvatarManager({projectKey:a.projectKey,projectId:a.projectId,defaultAvatarId:a.defaultAvatarId,avatarType:a.avatarType}),initialSelection:a.initialSelection,imageEditor:new r,size:a.hasOwnProperty("avatarSize")?a.avatarSize:n.LARGE,select:a.select})},s.createProjectAvatarPicker=function(a){return new t({avatarManager:i.createProjectAvatarManager({projectKey:a.projectKey,projectId:a.projectId,defaultAvatarId:a.defaultAvatarId}),imageEditor:new r,size:n.LARGE,select:a.select})},s.createUserAvatarPicker=function(a){return new t({avatarManager:i.createUserAvatarManager({username:a.username,defaultAvatarId:a.defaultAvatarId}),imageEditor:new r,size:n.LARGE,select:a.select})},s.createUniversalAvatarPickerDialog=function(a){var t=a.initialSelection||a.defaultAvatarId,r=new c({trigger:a.trigger,id:"project-avatar-picker",width:600,stacked:!0,content:function(i){var n,c;c=l('<div id="projectavatar-content-wrapper"/>'),l("<h2 />").text(a.title||"\u9009\u62e9\u4e00\u4e2a\u9879\u76ee\u56fe\u50cf").appendTo(c),n=s.createUniversalAvatarPicker({projectKey:a.projectKey,projectId:a.projectId,defaultAvatarId:a.defaultAvatarId,initialSelection:t,avatarType:a.avatarType,avatarSize:a.avatarSize,select:function(e,i,n){t=String(e.getId()),a.select&&a.select.apply(this,arguments),n||r.hide()}}),n.render(function(a){c.append(a),i(c)})}});r._focusFirstField=function(){}},s.createProjectAvatarPickerDialog=function(a){var t=new c({trigger:a.trigger,id:"project-avatar-picker",width:600,stacked:!0,content:function(r){var i,n;n=l('<div id="projectavatar-content-wrapper"/>'),l("<h2 />").text("\u9009\u62e9\u4e00\u4e2a\u9879\u76ee\u56fe\u50cf").appendTo(n),i=s.createProjectAvatarPicker({projectKey:a.projectKey,projectId:a.projectId,defaultAvatarId:a.defaultAvatarId,select:function(e,r,i){a.select&&a.select.apply(this,arguments),i||t.hide()}}),i.render(function(a){n.append(a),r(n)})}});t._focusFirstField=function(){}};var v=d.claim("jira.core:avatar-picker-data.data");s.createUserAvatarPickerDialog=function(t){if(v&&v.isEnabled)return void l(t.trigger).click(function(e){var t=o()+v.url;t+=(t.indexOf("?")>-1?"&":"?")+"continue="+encodeURIComponent(window.location.href),e.preventDefault(),e.stopPropagation(),a.reloadViaWindowLocation(t)});var r=new c({trigger:t.trigger,id:"user-avatar-picker",width:600,stacked:!0,content:function(a){var i,n;n=l('<div id="useravatar-content-wrapper"/>'),l("<h2 />").text("\u9009\u62e9\u4e00\u4e2a\u7528\u6237\u56fe\u50cf").appendTo(n),i=s.createUserAvatarPicker({username:t.username,defaultAvatarId:t.defaultAvatarId,select:function(a,e,i){t.select&&t.select.apply(this,arguments),l(".avatar-image").attr("src",e),i||r.hide()}}),i.render(function(e){n.append(e),a(n)})}})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/GravatarUtil.js' */
define("jira/ajs/avatarpicker/gravatar-util",["jquery","exports"],function(a,r){r.showGravatarHelp=function(r){void 0!==r&&void 0!==r.entry&&(a(".gravatar-signup-text").addClass("hidden"),a(".gravatar-login-text").removeClass("hidden"))},r.displayGravatarHelp=function(){var t=a("#gravatar_json_url");t.length&&a.ajax(t.val(),{dataType:"jsonp",success:r.showGravatarHelp})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/initAvatarPicker.js' */
!function(){var a=require("jquery"),r=require("jira/ajs/avatarpicker/avatar/sizes"),e=require("jira/ajs/avatarpicker/avatar-factory"),t=require("jira/ajs/avatarpicker/avatar-manager-factory"),c=require("jira/ajs/avatarpicker/avatar-picker-factory"),A=require("jira/ajs/avatarpicker/gravatar-util");AJS.namespace("JIRA.Avatar",null,require("jira/ajs/avatarpicker/avatar")),AJS.namespace("JIRA.Avatar.createCustomAvatar",null,e.createCustomAvatar),AJS.namespace("JIRA.Avatar.createSystemAvatar",null,e.createSystemAvatar),AJS.namespace("JIRA.Avatar.getSizeObjectFromName",null,r.getSizeObjectFromName),AJS.namespace("JIRA.Avatar.LARGE",null,r.LARGE),AJS.namespace("JIRA.Avatar.MEDIUM",null,r.MEDIUM),AJS.namespace("JIRA.Avatar.SMALL",null,r.SMALL),AJS.namespace("JIRA.AvatarManager",null,require("jira/ajs/avatarpicker/avatar-manager")),AJS.namespace("JIRA.AvatarManager.createUniversalAvatarManager",null,t.createUniversalAvatarManager),AJS.namespace("JIRA.AvatarManager.createProjectAvatarManager",null,t.createProjectAvatarManager),AJS.namespace("JIRA.AvatarManager.createUserAvatarManager",null,t.createUserAvatarManager),AJS.namespace("JIRA.AvatarPicker",null,require("jira/ajs/avatarpicker/avatar-picker")),AJS.namespace("JIRA.AvatarPicker.ImageEditor",null,require("jira/ajs/avatarpicker/avatar-picker/image-editor")),AJS.namespace("JIRA.AvatarPicker.createUniversalAvatarPicker",null,c.createUniversalAvatarPicker),AJS.namespace("JIRA.AvatarPicker.createProjectAvatarPicker",null,c.createProjectAvatarPicker),AJS.namespace("JIRA.AvatarPicker.createUserAvatarPicker",null,c.createUserAvatarPicker),AJS.namespace("JIRA.createUniversalAvatarPickerDialog",null,c.createUniversalAvatarPickerDialog),AJS.namespace("JIRA.createProjectAvatarPickerDialog",null,c.createProjectAvatarPickerDialog),AJS.namespace("JIRA.createUserAvatarPickerDialog",null,c.createUserAvatarPickerDialog),AJS.namespace("JIRA.AvatarStore",null,require("jira/ajs/avatarpicker/avatar-store")),AJS.namespace("JIRA.AvatarUtil",null,require("jira/ajs/avatarpicker/avatar-util")),AJS.namespace("JIRA.GravatarUtil.showGravatarHelp",null,A.showGravatarHelp),a(function(){c.createUserAvatarPickerDialog({trigger:"#user_avatar_image",username:a("#avatar-owner-id").text(),defaultAvatarId:a("#default-avatar-id").text()}),a("#gravatar_help_params")&&A.displayGravatarHelp()})}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker', location = '/includes/ajs/avatarpicker/AvatarPicker.soy' */
// This file was automatically generated from AvatarPicker.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.AvatarPicker.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.AvatarPicker == 'undefined') { JIRA.Templates.AvatarPicker = {}; }


JIRA.Templates.AvatarPicker.picker = function(opt_data, opt_ignored) {
  var output = '<form class="jira-avatar-upload-form aui top-label" action="#"><input name="id" value="10000" type="hidden" ><input name="pid" value="10000" type="hidden" ><div class="field-group"><label for="jira-avatar-uploader">' + soy.$$escapeHtml('\u4e0a\u4f20\u4e00\u4e2a\u65b0\u56fe\u50cf') + '</label><input type="file" class="ignore-inline-attach" name="avatar" id="jira-avatar-uploader"/><div class="error"></div></div></form><div class="form-body"><ul class="jira-avatars">';
  var avatarList6 = opt_data.avatars.system;
  var avatarListLen6 = avatarList6.length;
  for (var avatarIndex6 = 0; avatarIndex6 < avatarListLen6; avatarIndex6++) {
    var avatarData6 = avatarList6[avatarIndex6];
    output += '<li class="jira-avatar jira-system-avatar ' + ((avatarData6.isSelected) ? 'jira-selected-avatar' : '') + '" title="Select this Avatar" data-id="' + soy.$$escapeHtml(avatarData6.id) + '"><button data-id="' + soy.$$escapeHtml(avatarData6.id) + '" class="jira-icon-button" title="' + soy.$$escapeHtml('\u9009\u62e9\u8fd9\u4e2a\u56fe\u50cf') + '"><img id="avatar-' + soy.$$escapeHtml(avatarData6.id) + '" src="' + soy.$$escapeHtml(avatarData6.src) + '" width="' + soy.$$escapeHtml(avatarData6.width) + '" height="' + soy.$$escapeHtml(avatarData6.height) + '" alt="' + soy.$$escapeHtml('\u9009\u62e9\u8fd9\u4e2a\u56fe\u50cf') + '"/></button></li>';
  }
  var avatarList29 = opt_data.avatars.custom;
  var avatarListLen29 = avatarList29.length;
  for (var avatarIndex29 = 0; avatarIndex29 < avatarListLen29; avatarIndex29++) {
    var avatarData29 = avatarList29[avatarIndex29];
    output += '<li class="jira-avatar jira-custom-avatar ' + ((avatarData29.isSelected) ? 'jira-selected-avatar' : '') + '" title="Select this avatar" data-id="' + soy.$$escapeHtml(avatarData29.id) + '"><button data-id="' + soy.$$escapeHtml(avatarData29.id) + '" class="jira-icon-button" title="' + soy.$$escapeHtml('\u9009\u62e9\u8fd9\u4e2a\u56fe\u50cf') + '"><img id="avatar-' + soy.$$escapeHtml(avatarData29.id) + '" src="' + soy.$$escapeHtml(avatarData29.src) + '" width="' + soy.$$escapeHtml(avatarData29.width) + '" height="' + soy.$$escapeHtml(avatarData29.height) + '" alt="' + soy.$$escapeHtml('\u9009\u62e9\u8fd9\u4e2a\u56fe\u50cf') + '" /></button><button class="jira-delete-avatar jira-icon-button" data-id="' + soy.$$escapeHtml(avatarData29.id) + '" title="' + soy.$$escapeHtml('\u5220\u9664\u8fd9\u4e2a\u56fe\u50cf') + '">' + soy.$$escapeHtml('\u5220\u9664\u8fd9\u4e2a\u56fe\u50cf') + '</button></li>';
  }
  output += '</ul></div>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.AvatarPicker.picker.soyTemplateName = 'JIRA.Templates.AvatarPicker.picker';
}


JIRA.Templates.AvatarPicker.imageEditor = function(opt_data, opt_ignored) {
  return '<form id="avataror" class="jira-avatar-cropper-form aui" action="/jira/secure/project/AvatarPicker.jspa"><input type="hidden" name="cropperOffsetX" id="avatar-offsetX" value="' + soy.$$escapeHtml(opt_data.cropperOffsetX) + '"><input type="hidden" name="cropperOffsetY" id="avatar-offsetY" value="' + soy.$$escapeHtml(opt_data.cropperOffsetY) + '"><input type="hidden" name="cropperWidth" id="avatar-width"  value="' + soy.$$escapeHtml(opt_data.cropperWidth) + '"><div class="jira-avatar-cropper-header"><p>' + soy.$$escapeHtml('\u8bf7\u622a\u53d6\u9700\u8981\u4e0a\u4f20\u56fe\u50cf\u7684\u533a\u57df\u3002') + '</p></div><div class="form-body"><div class="avataror"><img src="' + soy.$$escapeHtml(opt_data.url) + '" height="300" /></div></div><div class="form-footer buttons-container"><div class="buttons"><input type="submit" class="aui-button aui-button-primary" value="' + soy.$$escapeHtml('\u786e\u8ba4') + '"><a class="aui-button aui-button-link cancel" href="#">' + soy.$$escapeHtml('\u53d6\u6d88') + '</a></div></div></form>';
};
if (goog.DEBUG) {
  JIRA.Templates.AvatarPicker.imageEditor.soyTemplateName = 'JIRA.Templates.AvatarPicker.imageEditor';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker-trigger', location = '/includes/jira/admin/initAvatarPickerTrigger.js' */
define("jira/admin/init/avatar-picker-trigger",["jira/ajs/layer/layer-constants","jira/util/events","jira/util/events/types","jira/util/events/reasons","jira/ajs/layer/inline-layer","jira/ajs/contentretriever/content-retriever","jira/ajs/avatarpicker/avatar","jira/ajs/avatarpicker/avatar-picker-factory","jira/ajs/avatarpicker/avatar-picker","jira/ajs/avatarpicker/avatar-picker/image-editor","jquery"],function(t,a,e,r,i,n,c,s,v,o,d){"use strict";function f(t){var a=d(".jira-avatar-picker-trigger"),e=d(".jira-avatar-picker-trigger img, img.jira-avatar-picker-trigger",t),r=d(t).find("#avatar-picker-avatar-id"),i=d(t).find("#avatar-picker-iconurl"),n=d(t).find("#avatar-type");""!==n.text()&&s.createUniversalAvatarPickerDialog({trigger:a,title:d(t).find("#avatar-dialog-title").text(),projectId:d(t).find("#avatar-owner-id").text(),projectKey:d(t).find("#avatar-owner-key").text(),defaultAvatarId:d(t).find("#default-avatar-id").text(),initialSelection:r.val(),avatarSize:c.getSizeObjectFromName(d(t).find("#avatar-size").text()),avatarType:n.text(),select:function(t,a){e.attr("src",a),i.val(a),r.val(t.getId())}})}function l(a){var e=d(".jira-inline-avatar-picker-trigger",a);e.length&&new j({offsetTarget:e,projectId:d(a).find("#avatar-owner-id").text(),projectKey:d(a).find("#avatar-owner-key").text(),defaultAvatarId:d(a).find("#default-avatar-id").text(),alignment:t.LEFT,width:420,allowDownsize:!0})}var g=n.extend({init:function(t){this.avatarPicker=t},content:function(t){this.avatarPicker.render(function(a){t(d("<div />").html(a))})},cache:function(){return!1},isLocked:function(){},startingRequest:function(){},finishedRequest:function(){}}),j=i.extend({init:function(t){var a=this;this.avatarPicker=v.createProjectAvatarPicker({projectId:t.projectId,projectKey:t.projectKey,defaultAvatarId:t.defaultAvatarId,select:function(e,r,i){t.select&&t.select.apply(this,arguments),i||a.hide(),a.offsetTarget().attr("src",r),a.offsetTarget().trigger("AvatarSelected")}}),t.contentRetriever=new g(this.avatarPicker),d(this.avatarPicker.imageEditor).bind(o.LOADED,function(){a.setWidth(a.layer().attr("scrollWidth"))}),this._super(t);var e=this.offsetTarget(),r=d("<span class='jira-avatar-picker-trigger'></span>");r.insertBefore(e).append(e),this._assignEvents("offsetTarget",r)},_events:{offsetTarget:{click:function(t){this.show()}}}});a.bind(e.NEW_CONTENT_ADDED,function(t,a,e){e!==r.panelRefreshed&&(f(a),l(a))})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:avatar-picker-trigger', location = '/includes/jira/admin/init/init-avatar-picker.js' */
require("jira/admin/init/avatar-picker-trigger");
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:croneditor', location = '/includes/js/cron/croneditor.js' */
define("jira/util/cron",["exports"],function(n){"use strict";var i=window.timesOnce=window.timesOnce||{};n.hideCronEdit=function(n){document.getElementById(n).style.display="none"},n.showCronEdit=function(n){document.getElementById(n).style.display=""},n.toggleFrequencyControl=function(i,e){0==document.getElementById(i+"interval").value?n.switchToOnce(i,e):n.switchToMany(i,e)},n.switchToOnce=function(e,o){n.hideCronEdit(e+"runMany"),n.showCronEdit(e+"runOnce"),o&&(i[e]=!0)},n.switchToMany=function(e,o){n.hideCronEdit(e+"runOnce"),n.showCronEdit(e+"runMany"),o&&(i[e]=!1)},n.switchToDaysOfMonth=function(i){n.hideCronEdit(i+"daysOfWeek"),n.showCronEdit(i+"daysOfMonth"),n.showCronEdit(i+"freqDiv"),n.hideCronEdit(i+"innerFreqDiv"),n.hideCronEdit(i+"advanced"),n.switchToOnce(i,!1)},n.switchToDaysOfWeek=function(i){n.showCronEdit(i+"daysOfWeek"),n.hideCronEdit(i+"daysOfMonth"),n.showCronEdit(i+"freqDiv"),n.showCronEdit(i+"innerFreqDiv"),n.hideCronEdit(i+"advanced"),n.switchToOriginal(i)},n.switchToDaily=function(i){n.hideCronEdit(i+"daysOfWeek"),n.hideCronEdit(i+"daysOfMonth"),n.showCronEdit(i+"freqDiv"),n.showCronEdit(i+"innerFreqDiv"),n.hideCronEdit(i+"advanced"),n.switchToOriginal(i)},n.switchToAdvanced=function(i){n.hideCronEdit(i+"daysOfWeek"),n.hideCronEdit(i+"daysOfMonth"),n.hideCronEdit(i+"runOnce"),n.hideCronEdit(i+"runMany"),n.hideCronEdit(i+"freqDiv"),n.showCronEdit(i+"advanced")},n.switchToOriginal=function(e){i[e]?n.switchToOnce(e,!1):n.switchToMany(e,!1)}}),function(){"use strict";var n=require("jira/util/cron");AJS.namespace("hideCronEdit",null,n.hideCronEdit),AJS.namespace("showCronEdit",null,n.showCronEdit),AJS.namespace("switchToOnce",null,n.switchToOnce),AJS.namespace("switchToMany",null,n.switchToMany),AJS.namespace("switchToDaysOfMonth",null,n.switchToDaysOfMonth),AJS.namespace("switchToDaysOfWeek",null,n.switchToDaysOfWeek),AJS.namespace("switchToDaily",null,n.switchToDaily),AJS.namespace("switchToAdvanced",null,n.switchToAdvanced),AJS.namespace("switchToOriginal",null,n.switchToOriginal),AJS.namespace("toggleFrequencyControl",null,n.toggleFrequencyControl)}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-quick-edit-plugin:loading-indicator', location = 'js/util/loading-indicator.js' */
define("quick-edit/util/loading-indicator",["jira/loading/loading","jira/flag"],function(b,d){var c=AJS.dim;var a=AJS.undim;return{showLoadingIndicator:function(){c();b.showLoadingIndicator()},hideLoadingIndicator:function(e){a();b.hideLoadingIndicator();if(e){d.showErrorMsg(null,"Jira \u670d\u52a1\u5668\u65e0\u6cd5\u8054\u7cfb\u3002\u8fd9\u53ef\u80fd\u662f\u4e00\u4e2a\u4e34\u65f6\u6545\u969c\u6216\u670d\u52a1\u5668\u53ef\u80fd\u5df2\u5173\u95ed\u3002")}}}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-quick-edit-plugin:quick-create-issue-bootstrap', location = 'js/init/create-bootstrap-wrm.js' */
require(["jquery","underscore","wrm/require","quick-edit/util/loading-indicator","jira/ajs/keyboardshortcut/keyboard-shortcut-toggle"],function(l,k,h,i,b){var f=function(){b.disable();i.showLoadingIndicator()};var c=function(){i.hideLoadingIndicator(true);b.enable()};var g;var a=false;l(document).on("simpleClick",".create-issue, .create-issue-type",function(o){o.preventDefault();if(!a){a=true;f();var n="wr!com.atlassian.jira.jira-quick-edit-plugin:quick-create-issue";var m=h([n]);m.done(d.bind(undefined,o)).fail(function p(){c();a=false})}});var d=function(n){var m=require("quick-edit/init/create");m.initIssueDialog(n);b.enable();a=false};l(function(){g=document.getElementById("stqc_show");if(g){g.onclick=null}});var j=false;l(document).on("simpleClick",".issueaction-create-subtask",function(o){o.preventDefault();if(!j){j=true;f();var n="wr!com.atlassian.jira.jira-quick-edit-plugin:quick-create-issue";var m=h([n]);m.done(e).fail(function p(){c();j=false})}});var e=function(n){var m=require("quick-edit/init/create");m.initSubtaskDialog(n);b.enable();j=false}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-page-atl-prefetch', location = '/includes/jira/page/prefetch.js' */
define("jira/page/atl/prefetch",["jira/util/data/meta","jira/data/local-storage","jira/ajs/dark-features","jquery","wrm/context-path"],function(e,t,r,n,i){"use strict";function a(e){n("<link />",{rel:"prefetch",href:e}).appendTo("head")}function s(e){var t=e.sections;if(t){var r=t.filter(function(e){return"issues_history_main"===e.id});if(r.length&&0!==r[0].items.length){return r[0].items[0].url}}}function c(e,t){for(var r;r=t.exec(e);){a(r[1].replace(/&amp;/g,"&"))}}function u(){try{t.setItem(g,d)}catch(e){}}function f(e){c(e,/<script.+?src="(.+?)".+?<\/script>/g),c(e,/<link.+?rel="stylesheet".+?href="(.+?)".+?>/g),u()}function o(){return!!r.isEnabled("jira.issue.prefetch")&&(1===n("#isNavigator").length?(u(),!1):e.get("issue-key")?(u(),!1):d!==t.getItem(g))}function l(e){var t=s(e);t&&h(t)}function h(e){n.get(e,f)}function p(){o()&&n.ajax({url:i()+"/rest/api/1.0/menus/find_link?inAdminMode=false",dataType:"json"}).done(l)}var g="jira.issue.prefetch.last.superbatch",d=function(){var e=new Date,t=e.getFullYear().toString()+e.getMonth().toString()+e.getDate().toString(),r=n("head > script").filter(function(e,t){return t.src.indexOf("/_super")>0});return(r.length>0?r[0].src:"empty")+t}();return{prefetchResourcesForUrl:h,prefetchViewIssueResources:p}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:init-jira-page-atl-prefetch', location = '/includes/jira/page/initPrefetch.js' */
require(["jira/page/atl/prefetch","jquery"],function(e,r){r(window).on("load",e.prefetchViewIssueResources.bind(e))});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.base-hipchat-integration-plugin-api:wrm-amd', location = 'lib/amd/wrm-amd.js' */
define("wrm/context-path",function(){return AJS.contextPath});define("wrm/data",function(){var a=function(){};if(WRM){if(typeof WRM.data==="function"){a=WRM.data}else{if(WRM.data&&typeof WRM.data.claim==="function"){a=WRM.data.claim}}}return{claim:a}});define("wrm",function(){if(WRM.data instanceof Function){WRM.data.claim=WRM.data}return WRM});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.labs.hipchat.hipchat-for-jira-plugin:amd', location = 'js/lib/amd/jira/jira-components-query-amd.js' */
define("hipchat/jira/query-component",function(){return JIRA.Components.Query});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.labs.hipchat.hipchat-for-jira-plugin:amd', location = 'js/lib/amd/jira/jira-issues-amd.js' */
define("hipchat/jira/issues/searcher-collection",function(){return JIRA.Issues.SearcherCollection});define("hipchat/jira/issues/query-state-model",function(){return JIRA.Issues.QueryStateModel});define("hipchat/jira/issues/criteria-model",function(){return JIRA.Issues.CriteriaModel});define("hipchat/jira/issues/criteria-view",function(){return JIRA.Issues.CriteriaView});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.labs.hipchat.hipchat-for-jira-plugin:amd', location = 'js/lib/amd/hipchat-scopes-provider-amd.js' */
define("hipchat/scopes-provider",function(){return HipChat.ScopesProvider});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.deprecated.behaviour.cookie', location = 'aui.chunk.232340e2530e70a15b98--f2430e5473cc8a1940a8.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.deprecated.behaviour.cookie"],{"N/qQ":function(e,r,n){"use strict";Object.defineProperty(r,"__esModule",{value:!0});var t=function(e){if(e&&e.__esModule)return e;var r={};if(null!=e)for(var n in e)Object.prototype.hasOwnProperty.call(e,n)&&(r[n]=e[n]);return r.default=e,r}(n("qHho"));r.default=t,e.exports=r.default},qHho:function(e,r,n){"use strict";Object.defineProperty(r,"__esModule",{value:!0}),r.save=r.read=r.erase=void 0;var t=function(e){if(e&&e.__esModule)return e;var r={};if(null!=e)for(var n in e)Object.prototype.hasOwnProperty.call(e,n)&&(r[n]=e[n]);return r.default=e,r}(n("bPPT")),o=function(e){return e&&e.__esModule?e:{default:e}}(n("KloK"));var u="AJS.conglomerate.cookie",a=/(\\|^"|"$)/g,i=/\|\|+/g,c=/"/g,l=/[.*+?|^$()[\]{\\]/g;function f(e){return e.replace(l,"\\$&")}function s(e){var r=new RegExp("\\b"+f(e)+"=((?:[^\\\\;]+|\\\\.)*)(?:;|$)"),n=document.cookie.match(r);return n&&function(e){return e.replace(a,"")}(n[1])}function p(e,r,n){var t=s(u);t=function(e,r,n){var t=new RegExp("(\\s|\\|)*\\b"+f(e)+"=[^|]*[|]*");if(n=(n=n||"").replace(t,"|"),""!==r){var o=e+"="+r;n.length+o.length<4020&&(n+="|"+o)}return n.replace(i,"|")}(e,r,t),function(e,r,n){var t,o="",u='"'+r.replace(c,'\\"')+'"';n&&((t=new Date).setTime(+t+24*n*60*60*1e3),o="; expires="+t.toGMTString()),document.cookie=e+"="+u+o+";path=/"}(u,t,n||365)}function d(e,r){var n=function(e,r){r=r||"";var n=new RegExp(f(e)+"=([^|]+)"),t=r.match(n);return t&&t[1]}(e,s(u));return null!=n?n:r}function v(e){p(e,"")}var h={erase:v,read:d,save:p};(0,o.default)("cookie",h),(0,o.default)("Cookie",h),t.prop(AJS,"Cookie",{alternativeName:"cookie",sinceVersion:"5.8.0"}),r.erase=v,r.read=d,r.save=p}},[["N/qQ","runtime","aui.splitchunk.0d131bcbf1","aui.splitchunk.444efc83be","aui.splitchunk.78b1c5d416","aui.splitchunk.54d3f16c20"]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.labs.hipchat.hipchat-for-jira-plugin:update-plugin-banner-resources', location = 'js/banner/update-hipchat-link-banner.js' */
require(["jquery","jira/flag","jira/util/users/logged-in-user","wrm/context-path"],function(d,b,e,a){var c=AJS.Cookie;AJS.toInit(function(){var i="HIPCHAT_ADMIN_BANNER_DISMISS_2";var g;if(!h()&&e.isAdmin()&&!k()){d.ajax({url:a()+"/rest/hipchat/integrations/1.0/configuration/status",dataType:"json",cache:false}).done(function(m){if(m.installationCheckResult.required||m.v1){g=f(m)}})}function f(o){var n=JIRA.Templates.HipChat.Banner.updateHipChatLink({removeV1link:!o.v2&&o.v1,reinstallationDiscoveryTitle:o.installationCheckResult.discoveryTitle,reinstallationDiscoveryText:o.installationCheckResult.discoveryText});var m=b.showWarningMsg(null,n);d(document).on("click","#remove-hipchat-v1-link",l);d(document).on("aui-flag-close",".aui-flag",j);return m}function h(){return window.location.toString().indexOf("/plugins/servlet/hipchat/configure")>0}function j(m){if(d(m.target).find(".hipchat-update-banner").length>0){c.save(i,true)}}function k(){return c.read(i)=="true"}function l(m){m.preventDefault();d.ajax({url:a()+"/rest/hipchat/integrations/1.0/configuration/v1",cache:false,type:"DELETE"}).error(function(n){if(g){g.hide()}b.showErrorMsg(null,AJS.format("\u6211\u4eec\u4e0d\u80fd\u627e\u5230\u60a8\u7684\u96c6\u6210\u3002\u5982\u679c\u95ee\u9898\u4ecd\u7136\u5b58\u5728, \u8bf7\u8054\u7cfb Atlassian \u652f\u6301\u3002 \u003cbr/\u003e\u9519\u8bef\u72b6\u6001[{0}]",n.status))}).done(function(n){if(g){g.hide()}b.showSuccessMsg(null,"\u96c6\u6210\u5df2\u7ecf\u88ab\u6210\u529f\u79fb\u9664")})}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.labs.hipchat.hipchat-for-jira-plugin:update-plugin-banner-resources', location = 'soy/banner/update-hipchat-link-banner.soy' */
// This file was automatically generated from update-hipchat-link-banner.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.HipChat.Banner.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.HipChat == 'undefined') { JIRA.Templates.HipChat = {}; }
if (typeof JIRA.Templates.HipChat.Banner == 'undefined') { JIRA.Templates.HipChat.Banner = {}; }


JIRA.Templates.HipChat.Banner.updateHipChatLink = function(opt_data, opt_ignored) {
  return '<div class="hipchat-update-banner"><b>' + soy.$$escapeHtml(opt_data.reinstallationDiscoveryTitle) + '</b><p>' + soy.$$escapeHtml(opt_data.reinstallationDiscoveryText) + '</p><p/><a href="' + soy.$$escapeHtml("" + '/plugins/servlet/hipchat/configure') + '">' + soy.$$escapeHtml('\u73b0\u5728\u66f4\u65b0!') + '</a>' + ((opt_data.removeV1link) ? '<a id="remove-hipchat-v1-link">' + soy.$$escapeHtml('\u79fb\u9664\u96c6\u6210') + '</a>' : '') + '</div>';
};
if (goog.DEBUG) {
  JIRA.Templates.HipChat.Banner.updateHipChatLink.soyTemplateName = 'JIRA.Templates.HipChat.Banner.updateHipChatLink';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.feedback.jira-feedback-plugin:button-resources-init', location = '/js/feedback-dialog-init.js' */
require(["jquery","jira/loading/loading","wrm/require"],function(n,i,e){"use strict";function o(){t=!0,AJS.dim(),i.showLoadingIndicator()}function r(n){t=!1,n&&AJS.dim.$dim.css("transition","none"),AJS.undim(),n&&AJS.dim.$dim&&AJS.dim.$dim.css("transition",""),i.hideLoadingIndicator()}var t=!1;n(document).on("click",".jira-feedback-plugin",function(n){n.preventDefault(),t||(o(),e(["wr!com.atlassian.feedback.jira-feedback-plugin:button-resources"]).done(function(){require("jira/feedback/feedback-dialog").show().then(function(){return r(!0)},function(n){console.error("Something went wrong in an attempt to show the issue collector dialog",n),r(!1)})}).fail(function(n){console.error("Something went wrong when loading feedback plugin resources",n),r(!1)}))})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:newsletter-signup-tip-init', location = 'static/components/newsletter/initNewsletterSignup.js' */
define("jira/newsletter/init-signup",["wrm/data","wrm/require","require"],function(e,i,n){return function(){var r=e.claim("com.atlassian.jira.jira-header-plugin:newsletter-signup-tip-init.newsletterSignup");r&&r.showNewsletterTip&&i("wr!com.atlassian.jira.jira-header-plugin:newsletter-signup-tip").then(function(){n("jira/newsletter/signuptip").render(r)})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:newsletter-signup-tip-init-init', location = 'static/components/newsletter/initInitNewsletterSignup.js' */
require(["jquery","jira/newsletter/init-signup"],function(i,n){i(function(){n()})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-admin-helper-plugin:dialog-resources', location = 'js/dialog.js' */
define("jira/admin-helper/dialog",["jquery","jira/dialog/dialog","jira/focus/set-focus"],function(c,a,b){return a.extend({_getDefaultOptions:function(){return c.extend(this._super(),{cached:false,widthClass:"large",stacked:true})},defineResources:function(){this._super();this.requireResource("com.atlassian.jira.plugins.jira-admin-helper-plugin:whereismycf-resources")},decorateContent:function(){var d=this;this.get$popupContent().find(".cancel").click(function(f){d.hide();f.preventDefault()})},_onShowContent:function(){this._super();if(a.current===this){var d=new b.FocusConfiguration();d.context=this.get$popup()[0];d.parentElementSelectors=[".form-body"];b.pushConfiguration(d);b.triggerFocus()}},hide:function(d){if(this._super(d)===false){return false}b.popConfiguration()}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-admin-helper-plugin:dialog-resources', location = 'js/notification-helper-init.js' */
require(["jquery","jira/admin-helper/dialog","jira/util/data/meta"],function(c,a,b){if(b.get("is-admin")){c(document).delegate(".notificationhelper-trigger","click",function(e){e.preventDefault();new a({id:"notification-helper-dialog",content:function d(g){var f=this;require(["jira/admin-helper/notification-helper/content-loader"],function(h){h.loadContent(f,g)})}}).show()})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-admin-helper-plugin:dialog-resources', location = 'js/permission-helper-init.js' */
require(["jquery","jira/admin-helper/dialog","jira/util/data/meta"],function(c,a,b){if(b.get("is-admin")){c(document).delegate(".permissionhelper-trigger","click",function(e){e.preventDefault();new a({id:"permission-helper-dialog",content:function d(g){var f=this;require(["jira/admin-helper/permission-helper/content-loader"],function(h){h.loadContent(f,g)})}}).show()})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-admin-helper-plugin:dialog-resources', location = 'js/whereismycf-init.js' */
require(["jquery","jira/admin-helper/dialog","jira/util/data/meta","jira/dialog/dialog"],function(e,a,d,b){if(d.get("is-admin")){var c="\u67e5\u627e\u57df\u540d";e(document).delegate(".whereismycf-trigger","click",function(g){g.preventDefault();new a({id:"whereismycf-dialog",content:function f(i){var h=this;require(["jira/admin-helper/whereismycf/content-loader"],function(j){j.loadContentForView(h,i)})}}).show()});e(document).bind("showLayer",function(g,i,h){if(i=="inlineDialog"&&h.popup.attr("id")=="inline-dialog-field_picker_popup"){if(h.popup.find(".whereismycf-qfpicker-link").length==0){var f=e("<a href='#' class='whereismycf-qfpicker-link'>"+c+"</a>").appendTo(h.popup.find(".qf-picker-header dl"));f.click(function(l){l.preventDefault();h.hide();var j=b.current;new a({id:"whereismycf-dialog",content:function k(n){var m=this;require(["jira/admin-helper/whereismycf/content-loader"],function(o){o.loadContentForEditAndCreate(m,j,n)})}}).show()})}}})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:field-templates', location = '/includes/jira/field/templates/singleUserPicker.soy' */
// This file was automatically generated from singleUserPicker.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.Fields.Pickers.User.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.Fields == 'undefined') { JIRA.Templates.Fields = {}; }
if (typeof JIRA.Templates.Fields.Pickers == 'undefined') { JIRA.Templates.Fields.Pickers = {}; }
if (typeof JIRA.Templates.Fields.Pickers.User == 'undefined') { JIRA.Templates.Fields.Pickers.User = {}; }


JIRA.Templates.Fields.Pickers.User.single = function(opt_data, opt_ignored) {
  var output = '<select id="' + soy.$$escapeHtml(opt_data.field.id) + '" name="' + soy.$$escapeHtml(opt_data.field.name) + '" class="single-user-picker js-' + ((opt_data.type) ? soy.$$escapeHtml(opt_data.type) : 'default-user-picker') + '"' + ((opt_data.inputText) ? ' data-input-text="' + soy.$$escapeHtml(opt_data.inputText) + '"' : '') + ((opt_data.editValue) ? ' data-edit-value="' + soy.$$escapeHtml(opt_data.editValue) + '"' : '') + ((opt_data.showDropdownButton) ? ' data-show-dropdown-button="true"' : '') + ((opt_data.userType) ? ' data-user-type="' + soy.$$escapeHtml(opt_data.userType) + '"' : '') + ((opt_data.containerClass) ? ' data-container-class="' + soy.$$escapeHtml(opt_data.containerClass) + '"' : '') + '>';
  var optionList38 = opt_data.options;
  var optionListLen38 = optionList38.length;
  for (var optionIndex38 = 0; optionIndex38 < optionListLen38; optionIndex38++) {
    var optionData38 = optionList38[optionIndex38];
    if (optionData38.optionGroup) {
      output += '<optgroup id="' + soy.$$escapeHtml(opt_data.field.id) + '-group-' + soy.$$escapeHtml(optionData38.id) + '" label="' + soy.$$escapeHtml(optionData38.display) + '"' + ((optionData38.footer) ? ' data-footer-text="' + soy.$$escapeHtml(optionData38.footer) + '"' : '') + ((optionData38.weight != -1) ? ' data-weight="' + soy.$$escapeHtml(optionData38.weight) + '"' : '') + '>';
      var groupOptionList59 = optionData38.groupOptions;
      var groupOptionListLen59 = groupOptionList59.length;
      for (var groupOptionIndex59 = 0; groupOptionIndex59 < groupOptionListLen59; groupOptionIndex59++) {
        var groupOptionData59 = groupOptionList59[groupOptionIndex59];
        output += JIRA.Templates.Fields.Pickers.User.option(groupOptionData59);
      }
      output += '</optgroup>';
    } else {
      output += JIRA.Templates.Fields.Pickers.User.option(optionData38);
    }
  }
  output += '</select>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.Fields.Pickers.User.single.soyTemplateName = 'JIRA.Templates.Fields.Pickers.User.single';
}


JIRA.Templates.Fields.Pickers.User.option = function(opt_data, opt_ignored) {
  return '<option ' + ((opt_data.loggedInUser) ? 'class="current-user" ' : '') + ((opt_data.selected) ? 'selected="selected" ' : '') + 'value="' + soy.$$escapeHtml(opt_data.optionName) + '" data-field-text="' + soy.$$escapeHtml(opt_data.displayName) + '" data-field-label="' + soy.$$escapeHtml(opt_data.displayName) + ((opt_data.emailAddress) ? ' - ' + soy.$$escapeHtml(opt_data.emailAddress) : '') + ((opt_data.optionName && opt_data.optionName != '-1') ? ' (' + soy.$$escapeHtml(opt_data.optionName) + ')' : '') + '" data-icon="' + soy.$$escapeHtml(opt_data.avatarURL) + '" >' + soy.$$escapeHtml(opt_data.displayName) + '</option>';
};
if (goog.DEBUG) {
  JIRA.Templates.Fields.Pickers.User.option.soyTemplateName = 'JIRA.Templates.Fields.Pickers.User.option';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:field-templates', location = '/includes/jira/field/templates/assigneeField.soy' */
// This file was automatically generated from assigneeField.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.Fields.Pickers.User.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.Fields == 'undefined') { JIRA.Templates.Fields = {}; }
if (typeof JIRA.Templates.Fields.Pickers == 'undefined') { JIRA.Templates.Fields.Pickers = {}; }
if (typeof JIRA.Templates.Fields.Pickers.User == 'undefined') { JIRA.Templates.Fields.Pickers.User = {}; }


JIRA.Templates.Fields.Pickers.User.assignee = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.Fields.Pickers.User.single(soy.$$augmentMap(opt_data, {options: opt_data.assigneeOptions.options, showDropdownButton: true, userType: 'assignee', type: 'assignee-picker', containerClass: 'long-field'})) + ((opt_data.isLoggedInUserAssignable) ? '<a href="#' + soy.$$escapeHtml(opt_data.field.id) + '" id="assign-to-me-trigger">' + soy.$$escapeHtml('\u5206\u914d\u7ed9\u6211') + '</a>' : '') + '<fieldset class="hidden parameters"><input type="hidden" title="projectKeys" value="' + soy.$$escapeHtml(opt_data.projectKeys) + '"/>' + ((opt_data.issueKey) ? '<input type="hidden" title="assigneeEditIssueKey" value="' + soy.$$escapeHtml(opt_data.issueKey) + '"/>' : '') + ((opt_data.actionDescriptorId) ? '<input type="hidden" title="actionDescriptorId" value="' + soy.$$escapeHtml(opt_data.actionDescriptorId) + '"/>' : '') + '</fieldset>';
};
if (goog.DEBUG) {
  JIRA.Templates.Fields.Pickers.User.assignee.soyTemplateName = 'JIRA.Templates.Fields.Pickers.User.assignee';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:user-pickers', location = '/includes/jira/field/userPickerUtil.js' */
define("jira/field/user-picker-util",["jira/ajs/list/item-descriptor","jira/ajs/list/group-descriptor","jquery"],function(e,i,r){return{formatResponse:function(t){var a=[];return r(t).each(function(t,l){var s=new i({weight:t,label:l.footer});r(l.users).each(function(){s.addItem(new e({value:this.name,label:this.displayName,html:this.html,icon:this.avatarUrl,allowDuplicate:!1,highlighted:!0,userKey:this.key}))}),a.push(s)}),a}}}),AJS.namespace("JIRA.UserPickerUtil",null,require("jira/field/user-picker-util"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:user-pickers', location = '/includes/jira/field/initSingleUserPickers.js' */
define("jira/field/init-single-user-pickers",["jquery","wrm/context-path","jira/ajs/select/single-select","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/field/user-picker-util","jira/util/formatter"],function(e,t,r,i,a,s,n,u){"use strict";function o(t){e(".js-default-user-picker",t).each(function(){var t=e(this);if(!t.data("aui-ss")){var i={showAvatar:!0},a=t.data("inputValue");new r({element:t,submitInputVal:!0,showDropdownButton:!!t.data("show-dropdown-button"),errorMessage:u.format("\u6ca1\u6709\u7528\u6237 \'\'{0}\'\' \u3002","'{0}'"),ajaxOptions:{url:l+"/rest/api/2/user/picker",query:!0,data:i,formatResponse:n.formatResponse},inputText:a,ariaLabel:"\u9009\u62e9\u4e00\u4e2a\u7528\u6237"})}})}var l=t();s.bind(a.NEW_CONTENT_ADDED,function(e,t,r){r!==i.panelRefreshed&&o(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:user-pickers', location = '/includes/jira/field/NoBrowseUserNamePicker.js' */
define("jira/field/no-browser-user-name-picker",["require"],function(e){"use strict";var t=e("jquery"),r=e("jira/ajs/list/item-descriptor"),i=e("jira/ajs/select/multi-select"),a=e("jira/util/formatter"),n=e("wrm/context-path"),s=n();return i.extend({_getDefaultOptions:function(){return t.extend(!0,this._super(),{errorMessage:"\u5bfb\u6c42\u7684\u7528\u6237\u4e0d\u5b58\u5728\u3002",showDropdownButton:!1,removeOnUnSelect:!0,itemAttrDisplayed:"label"})},_handleCharacterInput:function(){},_setSuggestions:function(){},_handleServerError:function(e){404===e.status?this.showErrorMessage():this._super()},_deactivate:function(){this.validateAndAdd()},validateAndAdd:function(){var e=this;""===t.trim(this.$field.val())?this.hideErrorMessage():t.ajax({url:s+"/rest/api/2/user",data:{username:t.trim(e.getQueryVal())},success:function(t){e.hideErrorMessage(),e.$field.val(""),e.addItem(new r({label:t.displayName,value:t.name}))},error:function(){e.showErrorMessage()}})},_handleSpace:function(){this.validate()},_handleServerSuggestions:function(){this.hideErrorMessage(),this.handleFreeInput()},handleFreeInput:function(){var e=t.trim(this.$field.val());""!==e&&(this.addItem({value:e,label:e}),this.model.$element.trigger("change")),this.$field.val("")},keys:{Return:function(e){e.preventDefault(),this.validateAndAdd()},Spacebar:function(e){e.preventDefault(),this.validateAndAdd()}}})}),AJS.namespace("AJS.NoBrowseUserNamePicker",null,require("jira/field/no-browser-user-name-picker"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:user-pickers', location = '/includes/jira/field/initMultiUserPickers.js' */
define("jira/field/init-multi-user-pickers",["jquery","wrm/context-path","jira/ajs/select/multi-select","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/field/no-browser-user-name-picker","jira/field/user-picker-util"],function(e,t,r,i,n,s,a,u){function l(t){t.find(".js-default-multi-user-picker").each(function(){var t=e(this);AJS.params.currentUserCanBrowseUsers?new r({element:this,itemAttrDisplayed:"label",showDropdownButton:!1,removeOnUnSelect:!0,submitInputVal:!0,ajaxOptions:{url:o+"/rest/api/1.0/users/picker",query:!0,data:function(e){return{showAvatar:!0,query:e,exclude:t.val()}},formatResponse:u.formatResponse}}):new a({element:this})})}var o=t();s.bind(n.NEW_CONTENT_ADDED,function(e,t,r){r!==i.panelRefreshed&&l(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:user-pickers', location = '/includes/jira/field/init/init-user-pickers-webresource.js' */
require("jira/field/init-multi-user-pickers"),require("jira/field/init-single-user-pickers");
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/lib/jquery.dotdotdot-1.5.7.js' */
(function(i){if(i.fn.dotdotdot){return}i.fn.dotdotdot=function(v){if(this.length==0){if(!v||v.debug!==false){b(true,'No element found for "'+this.selector+'".')}return this}if(this.length>1){return this.each(function(){i(this).dotdotdot(v)})}var r=this;if(r.data("dotdotdot")){r.trigger("destroy.dot")}r.bind_events=function(){r.bind("update.dot",function(y,A){y.preventDefault();y.stopPropagation();t.maxHeight=(typeof t.height=="number")?t.height:o(r);t.maxHeight+=t.tolerance;if(typeof A!="undefined"){if(typeof A=="string"||A instanceof HTMLElement){A=i("<div />").append(A).contents()}if(A instanceof i){w=A}}s=r.wrapInner('<div class="dotdotdot" />').children();s.empty().append(w.clone(true)).css({height:"auto",width:"auto",border:"none",padding:0,margin:0});var z=false,x=false;if(q.afterElement){z=q.afterElement.clone(true);q.afterElement.remove()}if(k(s,t)){if(t.wrap=="children"){x=d(s,t,z)}else{x=m(s,r,s,t,z)}}s.replaceWith(s.contents());s=null;if(i.isFunction(t.callback)){t.callback.call(r[0],x,w)}q.isTruncated=x;return x}).bind("isTruncated.dot",function(y,x){y.preventDefault();y.stopPropagation();if(typeof x=="function"){x.call(r[0],q.isTruncated)}return q.isTruncated}).bind("originalContent.dot",function(y,x){y.preventDefault();y.stopPropagation();if(typeof x=="function"){x.call(r[0],w)}return w}).bind("destroy.dot",function(x){x.preventDefault();x.stopPropagation();r.unwatch().unbind_events().empty().append(w).data("dotdotdot",false)});return r};r.unbind_events=function(){r.unbind(".dot");return r};r.watch=function(){r.unwatch();if(t.watch=="window"){var z=i(window),y=z.width(),x=z.height();z.bind("resize.dot"+q.dotId,function(){if(y!=z.width()||x!=z.height()||!t.windowResizeFix){y=z.width();x=z.height();if(p){clearInterval(p)}p=setTimeout(function(){r.trigger("update.dot")},10)}})}else{u=j(r);p=setInterval(function(){var A=j(r);if(u.width!=A.width||u.height!=A.height){r.trigger("update.dot");u=j(r)}},100)}return r};r.unwatch=function(){i(window).unbind("resize.dot"+q.dotId);if(p){clearInterval(p)}return r};var w=r.contents(),t=i.extend(true,{},i.fn.dotdotdot.defaults,v),q={},u={},p=null,s=null;q.afterElement=c(t.after,r);q.isTruncated=false;q.dotId=l++;r.data("dotdotdot",true).bind_events().trigger("update.dot");if(t.watch){r.watch()}return r};i.fn.dotdotdot.defaults={ellipsis:"... ",wrap:"word",lastCharacter:{remove:[" ",",",";",".","!","?"],noEllipsis:[]},tolerance:0,callback:null,after:null,height:null,watch:false,windowResizeFix:true,debug:false};var l=1;function d(s,w,v){var u=s.children(),p=false;s.empty();for(var r=0,q=u.length;r<q;r++){var t=u.eq(r);s.append(t);if(v){s.append(v)}if(k(s,w)){t.remove();p=true;break}else{if(v){v.remove()}}}return p}function m(s,r,A,t,q){var w=s.contents(),x=false;s.empty();var z="table, thead, tbody, tfoot, tr, col, colgroup, object, embed, param, ol, ul, dl, select, optgroup, option, textarea, script, style";for(var y=0,u=w.length;y<u;y++){if(x){break}var v=w[y],p=i(v);if(typeof v=="undefined"){continue}s.append(p);if(q){s[(s.is(z))?"after":"append"](q)}if(v.nodeType==3){if(k(A,t)){x=e(p,r,A,t,q)}}else{x=m(p,r,A,t,q)}if(!x){if(q){q.remove()}}}return x}function e(q,r,p,w,s){var D=false,A=q[0];if(typeof A=="undefined"){return false}var E=(w.wrap=="letter")?"":" ",t=h(A).split(E),F=-1,v=-1,G=0,y=t.length-1;while(G<=y){var x=Math.floor((G+y)/2);if(x==v){break}v=x;a(A,t.slice(0,v+1).join(E)+w.ellipsis);if(!k(p,w)){F=v;G=v}else{y=v}}if(F!=-1&&!(t.length==1&&t[0].length==0)){var u=f(t.slice(0,F+1).join(E),w);D=true;a(A,u)}else{var z=q.parent();q.remove();var C=(s)?s.length:0;if(z.contents().size()>C){var B=z.contents().eq(-1-C);D=e(B,r,p,w,s)}else{var A=z.prev().contents().eq(-1)[0];if(typeof A!="undefined"){var u=f(h(A),w);a(A,u);z.remove();D=true}}}return D}function k(q,p){return q.innerHeight()>p.maxHeight}function f(p,q){while(i.inArray(p.slice(-1),q.lastCharacter.remove)>-1){p=p.slice(0,-1)}if(i.inArray(p.slice(-1),q.lastCharacter.noEllipsis)<0){p+=q.ellipsis}return p}function j(p){return{width:p.innerWidth(),height:p.innerHeight()}}function a(q,p){if(q.innerText){q.innerText=p}else{if(q.nodeValue){q.nodeValue=p}else{if(q.textContent){q.textContent=p}}}}function h(p){if(p.innerText){return p.innerText}else{if(p.nodeValue){return p.nodeValue}else{if(p.textContent){return p.textContent}else{return""}}}}function c(p,q){if(typeof p=="undefined"){return false}if(!p){return false}if(typeof p=="string"){p=i(p,q);return(p.length)?p:false}if(typeof p=="object"){return(typeof p.jquery=="undefined")?false:p}return false}function o(s){var t=s.innerHeight(),r=["paddingTop","paddingBottom"];for(var u=0,q=r.length;u<q;u++){var p=parseInt(s.css(r[u]),10);if(isNaN(p)){p=0}t-=p}return t}function b(q,p){if(!q){return false}if(typeof p=="string"){p="dotdotdot: "+p}else{p=["dotdotdot:",p]}if(typeof window.console!="undefined"){if(typeof window.console.log!="undefined"){window.console.log(p)}}return false}var n=i.fn.html;i.fn.html=function(p){if(typeof p!="undefined"){if(this.data("dotdotdot")){if(typeof p!="function"){return this.trigger("update",[p])}}return n.call(this,p)}return n.call(this)};var g=i.fn.text;i.fn.text=function(q){if(typeof q!="undefined"){if(this.data("dotdotdot")){var p=i("<div />");p.text(q);q=p.html();p.remove();return this.trigger("update",[q])}return g.call(this,q)}return g.call(this)}})(require("jquery"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/soy/ProjectTemplatesList.soy' */
// This file was automatically generated from ProjectTemplatesList.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.ProjectTemplates.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.ProjectTemplates == 'undefined') { JIRA.Templates.ProjectTemplates = {}; }


JIRA.Templates.ProjectTemplates.renderProjectTemplates = function(opt_data, opt_ignored) {
  return '' + ((soy.$$getMapKeys(opt_data.projectTemplates).length == 0) ? JIRA.Templates.errorMsg({closeable: false, msg: '\u6ca1\u6709\u627e\u5230\u9879\u76ee\u6a21\u677f\u3002'}) : JIRA.Templates.ProjectTemplates.renderTemplates(opt_data));
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.renderProjectTemplates.soyTemplateName = 'JIRA.Templates.ProjectTemplates.renderProjectTemplates';
}


JIRA.Templates.ProjectTemplates.renderProjectTemplatesGroupedByType = function(opt_data, opt_ignored) {
  var output = '';
  if (opt_data.projectTemplatesByType.length == 0) {
    output += JIRA.Templates.errorMsg({closeable: false, msg: '\u6ca1\u6709\u627e\u5230\u9879\u76ee\u6a21\u677f\u3002'});
  } else {
    var projectTypeTemplatesGroupList18 = opt_data.projectTemplatesByType;
    var projectTypeTemplatesGroupListLen18 = projectTypeTemplatesGroupList18.length;
    for (var projectTypeTemplatesGroupIndex18 = 0; projectTypeTemplatesGroupIndex18 < projectTypeTemplatesGroupListLen18; projectTypeTemplatesGroupIndex18++) {
      var projectTypeTemplatesGroupData18 = projectTypeTemplatesGroupList18[projectTypeTemplatesGroupIndex18];
      output += '<div class="template-group" id="project-template-group-' + soy.$$escapeHtml(projectTypeTemplatesGroupData18.projectTypeBean.projectTypeKey) + '"><div class="template-group-header"><h6><img class="project-type-icon" src="data:image/svg+xml;base64, ' + soy.$$escapeHtml(projectTypeTemplatesGroupData18.projectTypeBean.icon) + ' "/><span>' + soy.$$escapeHtml(projectTypeTemplatesGroupData18.projectTypeBean.projectTypeDisplayKey) + '</span></h6></div>' + JIRA.Templates.ProjectTemplates.renderTemplates({projectTemplates: projectTypeTemplatesGroupData18.projectTemplates}) + '</div>';
    }
  }
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.renderProjectTemplatesGroupedByType.soyTemplateName = 'JIRA.Templates.ProjectTemplates.renderProjectTemplatesGroupedByType';
}


JIRA.Templates.ProjectTemplates.renderTemplates = function(opt_data, opt_ignored) {
  return '<div class="pt-templates-list">' + JIRA.Templates.ProjectTemplates.renderItems(opt_data) + '</div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.renderTemplates.soyTemplateName = 'JIRA.Templates.ProjectTemplates.renderTemplates';
}


JIRA.Templates.ProjectTemplates.renderItems = function(opt_data, opt_ignored) {
  var output = '<ol class="templates" tabindex="100">';
  var templateList37 = opt_data.projectTemplates;
  var templateListLen37 = templateList37.length;
  for (var templateIndex37 = 0; templateIndex37 < templateListLen37; templateIndex37++) {
    var templateData37 = templateList37[templateIndex37];
    output += '<li class="template"' + ((templateData37.itemModuleCompleteKey) ? 'data-item-module-complete-key="' + soy.$$escapeHtml(templateData37.itemModuleCompleteKey) + '"' : '') + ((templateData37.projectTemplateModuleCompleteKey) ? 'data-project-template-module-complete-key="' + soy.$$escapeHtml(templateData37.projectTemplateModuleCompleteKey) + '"' : '') + ((templateData37.createProject) ? 'data-create-project="' + soy.$$escapeHtml(templateData37.createProject) + '"' : '') + ((templateData37.demoProject) ? 'data-demo-project="' + soy.$$escapeHtml(templateData37.demoProject) + '"' : '') + ((templateData37.backgroundIconUrl) ? 'data-background-icon-url="' + soy.$$escapeHtml(templateData37.backgroundIconUrl) + '"' : '') + ((templateData37.name) ? 'data-name="' + soy.$$escapeHtml(templateData37.name) + '"' : '') + ((templateData37.description) ? 'data-description="' + soy.$$escapeHtml(templateData37.description) + '"' : '') + ((templateData37.longDescriptionContent) ? 'data-long-description-content="' + soy.$$escapeHtml(templateData37.longDescriptionContent) + '"' : '') + ((templateData37.infoSoyPath) ? 'data-info-soy-path="' + soy.$$escapeHtml(templateData37.infoSoyPath) + '"' : '') + '><img class="template-preview" src="' + soy.$$escapeHtml(templateData37.iconUrl) + '" /><div class="template-meta"><div class="template-name" title="' + soy.$$escapeHtml(templateData37.name) + '">' + soy.$$escapeHtml(templateData37.name) + '</div><div class="template-description" title="' + soy.$$escapeHtml(templateData37.description) + '">' + soy.$$escapeHtml(templateData37.description) + '</div></div></li>';
  }
  output += '</ol>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.renderItems.soyTemplateName = 'JIRA.Templates.ProjectTemplates.renderItems';
}


JIRA.Templates.ProjectTemplates.loadingTemplatesList = function(opt_data, opt_ignored) {
  return '<div class="wait-container"><img class="wait-icon" src="' + soy.$$escapeHtml("") + '/images/icons/wait.gif"><span class="wait-text">' + soy.$$escapeHtml('\u4e0a\u4f20\u9879\u76ee\u6a21\u677f') + '&hellip;</span></div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.loadingTemplatesList.soyTemplateName = 'JIRA.Templates.ProjectTemplates.loadingTemplatesList';
}


JIRA.Templates.ProjectTemplates.footerLinks = function(opt_data, opt_ignored) {
  return '<div class="footer-links"><a class="import-project-trigger button-panel-link" href="' + soy.$$escapeHtml("") + '/secure/admin/views/ExternalImport1.jspa">' + soy.$$escapeHtml('\u5bfc\u5165\u4e00\u4e2a\u9879\u76ee') + '</a>|<a class="create-with-shared-config button-panel-link" href="#"><span class="aui-icon aui-icon-create-shared">' + soy.$$escapeHtml('\u521b\u5efa\u4e0e\u5171\u4eab\u914d\u7f6e') + '</span> ' + soy.$$escapeHtml('\u521b\u5efa\u4e0e\u5171\u4eab\u914d\u7f6e') + '</a>' + ((opt_data.showDemoLink) ? '| <a class="add-demo-project-trigger button-panel-link" href="#">' + soy.$$escapeHtml('\u521b\u5efa\u793a\u4f8b\u6570\u636e') + '</a>' : '') + '</div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.footerLinks.soyTemplateName = 'JIRA.Templates.ProjectTemplates.footerLinks';
}


JIRA.Templates.ProjectTemplates.addWorkflowsLink = function(opt_data, opt_ignored) {
  return '<a class="add-workflow-link" href="' + soy.$$escapeHtml(opt_data.baseUrl) + '/plugins/servlet/wfshare-import?src=projecttemplates" tabindex="-1">' + soy.$$escapeHtml('\u67e5\u770b\u5e02\u573a\u5de5\u4f5c\u6d41') + '</a>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.addWorkflowsLink.soyTemplateName = 'JIRA.Templates.ProjectTemplates.addWorkflowsLink';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/soy/AddProject.soy' */
// This file was automatically generated from AddProject.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.ProjectTemplates.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.ProjectTemplates == 'undefined') { JIRA.Templates.ProjectTemplates = {}; }


JIRA.Templates.ProjectTemplates.addProjectForm = function(opt_data, opt_ignored) {
  var output = '';
  if (opt_data.errors.errorMessages) {
    var errorList5 = opt_data.errors.errorMessages;
    var errorListLen5 = errorList5.length;
    for (var errorIndex5 = 0; errorIndex5 < errorListLen5; errorIndex5++) {
      var errorData5 = errorList5[errorIndex5];
      output += JIRA.Templates.errorMsg({closeable: false, msg: errorData5});
    }
  }
  output += '<div class="add-project-wrapper"><div class="add-project-form-wrapper">' + aui.form.form({content: '' + aui.form.fieldset({legendContent: '', content: '' + JIRA.Templates.ProjectTemplates.longTextField({id: 'name', name: 'name', isRequired: false, maxLength: opt_data.maxNameLength, labelContent: '\u540d\u79f0', descriptionText: AJS.format('\u6700\u591a{0}\u4e2a\u5b57\u7b26',opt_data.maxNameLength), value: opt_data.currentName ? opt_data.currentName : '', errorTexts: opt_data.errors.errors.projectName ? [opt_data.errors.errors.projectName] : []}) + JIRA.Templates.ProjectTemplates.textFieldWithHelpIcon({id: 'key', name: 'key', isRequired: false, maxLength: opt_data.maxKeyLength, labelContent: '\u5173\u952e\u5b57', descriptionText: AJS.format('\u6700\u591a{0}\u4e2a\u5b57\u7b26',opt_data.maxKeyLength), helpTitle: '\u4ec0\u4e48\u662f\u9879\u76ee\u952e\uff1f', value: opt_data.currentKey ? opt_data.currentKey : '', errorTexts: opt_data.errors.errors.projectKey ? [opt_data.errors.errors.projectKey] : []}) + ((opt_data.shouldShowLead) ? JIRA.Templates.ProjectTemplates.projectLeadField({field: opt_data.projectLeadPickerField, isRequired: false, errorTexts: opt_data.errors.errors.projectLead ? [opt_data.errors.errors.projectLead] : [], options: opt_data.leadOptions}) : '') + ((opt_data.addUserToLicense.displayCheckbox) ? JIRA.Templates.ProjectTemplates.addUserToLicense({fieldId: opt_data.addUserToLicense.fieldId, fieldName: opt_data.addUserToLicense.fieldName, applicationName: opt_data.addUserToLicense.applicationName, usedSeats: opt_data.addUserToLicense.usedSeats, totalSeats: opt_data.addUserToLicense.totalSeats, disableCheckbox: opt_data.addUserToLicense.disableCheckbox, licensingUrl: opt_data.addUserToLicense.licensingUrl}) : '') + '<input type="hidden" name="keyEdited" id="keyEdited" value="false"><input type="hidden" name="projectTemplateWebItemKey" value="' + soy.$$escapeHtml(opt_data.projectTemplateWebItemKey) + '"><input type="hidden" name="projectTemplateModuleKey" value="' + soy.$$escapeHtml(opt_data.projectTemplateModuleKey) + '"><input type="submit" class="pt-hidden-submit offscreen-left">'}), id: 'add-project-form'}) + '</div>' + ((opt_data.projectTemplateDescriptionContent) ? '<div class="add-project-description-wrapper"><div class="project-template-title"><h3>' + soy.$$escapeHtml(opt_data.projectTemplateTitle) + '</h3></div><div class="project-template-description">' + soy.$$filterNoAutoescape(opt_data.projectTemplateDescriptionContent) + '</div></div>' : '') + '</div>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.addProjectForm.soyTemplateName = 'JIRA.Templates.ProjectTemplates.addProjectForm';
}


JIRA.Templates.ProjectTemplates.keyHelp = function(opt_data, opt_ignored) {
  return '<div class="project-key-help"><p><b>' + soy.$$escapeHtml('\u4ec0\u4e48\u662f\u9879\u76ee\u952e\uff1f') + '</b></p><ul><li>' + soy.$$escapeHtml('\u5b83\u4f1a\u6210\u4e3a\u9879\u76ee\u4e2d\u6bcf\u4e2a\u95ee\u9898\u7684\u524d\u7f00') + '</li><li>' + soy.$$escapeHtml('\u5b83\u53ef\u4ee5\u88ab\u6539\u53d8, \u4f46\u8fd9\u5e76\u4e0d\u662f\u4e00\u4ef6\u7b80\u5355\u7684\u4efb\u52a1') + '</li></ul></div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.keyHelp.soyTemplateName = 'JIRA.Templates.ProjectTemplates.keyHelp';
}


JIRA.Templates.ProjectTemplates.projectLeadField = function(opt_data, opt_ignored) {
  var param77 = '' + aui.form.label({forField: opt_data.field.name + '-field', isRequired: opt_data.isRequired, content: '\u9879\u76ee\u8d1f\u8d23\u4eba'}) + JIRA.Templates.Fields.Pickers.User.single({field: opt_data.field, options: opt_data.options, editValue: opt_data.value}) + aui.form.fieldDescription({message: '\u8f93\u5165\u9879\u76ee\u8d1f\u8d23\u4eba\u7684\u7528\u6237\u540d'});
  if (opt_data.errorTexts) {
    var errorList90 = opt_data.errorTexts;
    var errorListLen90 = errorList90.length;
    for (var errorIndex90 = 0; errorIndex90 < errorListLen90; errorIndex90++) {
      var errorData90 = errorList90[errorIndex90];
      param77 += aui.form.fieldError({message: errorData90});
    }
  }
  var output = '' + aui.form.fieldGroup({content: param77});
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.projectLeadField.soyTemplateName = 'JIRA.Templates.ProjectTemplates.projectLeadField';
}


JIRA.Templates.ProjectTemplates.addUserToLicense = function(opt_data, opt_ignored) {
  var output = '';
  var labelMessage__soy96 = '' + soy.$$escapeHtml(AJS.format('\u7ed9\u6211\u4e00\u4e2a{0}\u8bb8\u53ef\u8bc1\uff0c\u8fd9\u6837\uff0c\u6211\u5c31\u53ef\u4ee5\u8fdb\u5165\u6b64\u9879\u76ee\u3002',opt_data.applicationName)) + ((opt_data.totalSeats > 0) ? '<br/>' + soy.$$escapeHtml(AJS.format('({0}\u7684{1}\u7684\u8bb8\u53ef\u8bc1\u73b0\u5728\u7528)',opt_data.usedSeats,opt_data.totalSeats)) : '');
  output += aui.form.fieldGroup({content: '' + aui.form.label({forField: opt_data.fieldName, content: 'License'}) + ((opt_data.disableCheckbox) ? aui.form.field({id: opt_data.fieldId, name: opt_data.fieldName, value: 'true', type: 'checkbox', isChecked: false, labelContent: labelMessage__soy96, isDisabled: true, descriptionContent: AJS.format('\u6211\u4eec\u4e0d\u80fd\u8ba9\u60a8\u8bbf\u95ee{0}\u81ea\u52a8\u7684\u3002\u60a8\u53ef\u4ee5\u5728{1}\u4e2d\u7ba1\u7406\u5e94\u7528\u7a0b\u5e8f\u8bbf\u95ee{2}\u7684\u5de5\u4f5c{0}\u9879\u76ee\u3002',opt_data.applicationName,'<a href="' + "" + opt_data.licensingUrl + '">','</a>')}) : aui.form.field({id: opt_data.fieldId, name: opt_data.fieldName, value: 'true', type: 'checkbox', isChecked: false, labelContent: labelMessage__soy96, isDisabled: false}))});
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.addUserToLicense.soyTemplateName = 'JIRA.Templates.ProjectTemplates.addUserToLicense';
}


JIRA.Templates.ProjectTemplates.textFieldWithHelpIcon = function(opt_data, opt_ignored) {
  var param128 = '' + aui.form.label({forField: opt_data.name, isRequired: opt_data.isRequired, content: opt_data.labelContent}) + aui.form.input({id: opt_data.id, name: opt_data.name, type: 'text', maxLength: opt_data.maxLength, value: opt_data.value}) + '<a class="help-lnk" title="' + soy.$$escapeHtml(opt_data.helpTitle) + '" id="' + soy.$$escapeHtml(opt_data.id) + '-help-icon"><span class="aui-icon aui-icon-help">' + soy.$$escapeHtml(opt_data.helpTitle) + '</span></a>' + ((opt_data.descriptionText) ? aui.form.fieldDescription({message: opt_data.descriptionText}) : '');
  if (opt_data.errorTexts) {
    var errorList152 = opt_data.errorTexts;
    var errorListLen152 = errorList152.length;
    for (var errorIndex152 = 0; errorIndex152 < errorListLen152; errorIndex152++) {
      var errorData152 = errorList152[errorIndex152];
      param128 += aui.form.fieldError({message: errorData152});
    }
  }
  var output = '' + aui.form.fieldGroup({content: param128});
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.textFieldWithHelpIcon.soyTemplateName = 'JIRA.Templates.ProjectTemplates.textFieldWithHelpIcon';
}


JIRA.Templates.ProjectTemplates.longTextField = function(opt_data, opt_ignored) {
  var param158 = '' + aui.form.label({forField: opt_data.name, isRequired: opt_data.isRequired, content: opt_data.labelContent}) + aui.form.input({id: opt_data.id, name: opt_data.name, type: 'text', maxLength: opt_data.maxLength, value: opt_data.value, extraClasses: 'long-field'}) + ((opt_data.descriptionText) ? aui.form.fieldDescription({message: opt_data.descriptionText}) : '');
  if (opt_data.errorTexts) {
    var errorList176 = opt_data.errorTexts;
    var errorListLen176 = errorList176.length;
    for (var errorIndex176 = 0; errorIndex176 < errorListLen176; errorIndex176++) {
      var errorData176 = errorList176[errorIndex176];
      param158 += aui.form.fieldError({message: errorData176});
    }
  }
  var output = '' + aui.form.fieldGroup({content: param158});
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.longTextField.soyTemplateName = 'JIRA.Templates.ProjectTemplates.longTextField';
}


JIRA.Templates.ProjectTemplates.spinner = function(opt_data, opt_ignored) {
  return '<span id=\'' + soy.$$escapeHtml(opt_data.id) + '\' class=\'icon throbber loading\'/>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.spinner.soyTemplateName = 'JIRA.Templates.ProjectTemplates.spinner';
}


JIRA.Templates.ProjectTemplates.formatAnchor = function(opt_data, opt_ignored) {
  return '<a href="' + soy.$$escapeHtml(opt_data.href) + '"' + ((opt_data.title) ? ' title="' + soy.$$escapeHtml(opt_data.title) + '"' : '') + '>' + soy.$$escapeHtml(opt_data.body) + '</a>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.formatAnchor.soyTemplateName = 'JIRA.Templates.ProjectTemplates.formatAnchor';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/soy/CreateSharedProject.soy' */
// This file was automatically generated from CreateSharedProject.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.ProjectTemplates.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.ProjectTemplates == 'undefined') { JIRA.Templates.ProjectTemplates = {}; }


JIRA.Templates.ProjectTemplates.createSharedProjectForm = function(opt_data, opt_ignored) {
  return '<div class="create-shared-project-wrapper"><div class="create-shared-project-form-wrapper">' + aui.form.form({content: '' + aui.form.fieldset({content: '<p>' + soy.$$escapeHtml('\u9009\u62e9\u60a8\u5e0c\u671b\u5171\u4eab\u914d\u7f6e\u7684\u73b0\u6709\u9879\u76ee\u3002') + '<a class="help-lnk" title="' + soy.$$escapeHtml('\u4ec0\u4e48\u4f1a\u88ab\u5171\u4eab\uff1f') + '" id="shared-help-icon"><span class="aui-icon aui-icon-help">' + soy.$$escapeHtml('\u4ec0\u4e48\u4f1a\u88ab\u5171\u4eab\uff1f') + '</span></a></p>' + aui.form.fieldGroup({extraClasses: 'project-picker-group', content: '' + aui.form.label({forField: 'project-picker', isRequired: false, content: '\u9009\u62e9\u4e00\u4e2a\u9879\u76ee'}) + aui.form.input({id: 'project-picker', name: 'project', type: 'text'}) + '<div id="project-picker-options" data-suggestions="' + soy.$$escapeHtml(opt_data.projectSuggestions) + '"></div>'}) + '<p class="create-shared-info">' + soy.$$escapeHtml('\u5f53\u4e00\u4e2a\u914d\u7f6e\u88ab\u591a\u4e2a\u9879\u76ee\u5171\u4eab\u65f6\uff0c\u8fd9\u610f\u5473\u7740\u5bf9\u914d\u7f6e\u7684\u4efb\u4f55\u6539\u52a8\u90fd\u4f1a\u5f71\u54cd\u6240\u6709\u9879\u76ee\u3002') + '</p>'}), id: 'create-shared-project-form'}) + '</div></div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.createSharedProjectForm.soyTemplateName = 'JIRA.Templates.ProjectTemplates.createSharedProjectForm';
}


JIRA.Templates.ProjectTemplates.loading = function(opt_data, opt_ignored) {
  return '<div><div class="dialog-spinner"></div></div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.loading.soyTemplateName = 'JIRA.Templates.ProjectTemplates.loading';
}


JIRA.Templates.ProjectTemplates.noProjects = function(opt_data, opt_ignored) {
  return '<p>' + soy.$$escapeHtml('\u770b\u8d77\u6765\u6ca1\u6709\u5176\u4ed6\u9879\u76ee\u7684\u5b58\u5728\u3002\u8bf7\u521b\u5efa\u5168\u65b0\u7684\u9879\u76ee\u7b2c\u4e00\u6b21\u5c1d\u8bd5\u4e4b\u524d\u8981\u5171\u4eab\u73b0\u6709\u9879\u76ee\u7684\u914d\u7f6e\u3002') + '</p>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.noProjects.soyTemplateName = 'JIRA.Templates.ProjectTemplates.noProjects';
}


JIRA.Templates.ProjectTemplates.sharedHelp = function(opt_data, opt_ignored) {
  return '<div class="shared-help"><p><strong>' + soy.$$escapeHtml('\u4ec0\u4e48\u4f1a\u88ab\u5171\u4eab\uff1f') + '</strong></p><p>' + soy.$$escapeHtml('\u914d\u7f6e\u5305\u542b\u9879\u76ee\u4f7f\u7528\u7684\u65b9\u6848\u7ec4\u5408') + '</p><ul><li>' + soy.$$escapeHtml('\u6743\u9650\u65b9\u6848') + '</li><li>' + soy.$$escapeHtml('\u901a\u77e5\u65b9\u6848') + '</li><li>' + soy.$$escapeHtml('\u95ee\u9898\u5b89\u5168\u65b9\u6848') + '</li><li>' + soy.$$escapeHtml('\u5de5\u4f5c\u6d41\u65b9\u6848') + '</li><li>' + soy.$$escapeHtml('\u95ee\u9898\u7c7b\u578b\u65b9\u6848') + '</li><li>' + soy.$$escapeHtml('\u95ee\u9898\u7c7b\u578b\u9875\u9762\u65b9\u6848') + '</li><li>' + soy.$$escapeHtml('\u57df\u914d\u7f6e\u65b9\u6848') + '</li></ul></div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.sharedHelp.soyTemplateName = 'JIRA.Templates.ProjectTemplates.sharedHelp';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/Config.js' */
define("jira/project-templates/config",{model:{}});AJS.namespace("JPT.ConfigModel",null,require("jira/project-templates/config"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/ProjectKeyGenerator.js' */
define("jira/project-templates/project-key-generator",["jquery","underscore"],function(d,h){var c={};c.IGNORED_WORDS=["THE","A","AN","AS","AND","OF","OR"];c.CHARACTER_MAP={};c.CHARACTER_MAP[199]="C";c.CHARACTER_MAP[231]="c";c.CHARACTER_MAP[252]="u";c.CHARACTER_MAP[251]="u";c.CHARACTER_MAP[250]="u";c.CHARACTER_MAP[249]="u";c.CHARACTER_MAP[233]="e";c.CHARACTER_MAP[234]="e";c.CHARACTER_MAP[235]="e";c.CHARACTER_MAP[232]="e";c.CHARACTER_MAP[226]="a";c.CHARACTER_MAP[228]="a";c.CHARACTER_MAP[224]="a";c.CHARACTER_MAP[229]="a";c.CHARACTER_MAP[225]="a";c.CHARACTER_MAP[239]="i";c.CHARACTER_MAP[238]="i";c.CHARACTER_MAP[236]="i";c.CHARACTER_MAP[237]="i";c.CHARACTER_MAP[196]="A";c.CHARACTER_MAP[197]="A";c.CHARACTER_MAP[201]="E";c.CHARACTER_MAP[230]="ae";c.CHARACTER_MAP[198]="Ae";c.CHARACTER_MAP[244]="o";c.CHARACTER_MAP[246]="o";c.CHARACTER_MAP[242]="o";c.CHARACTER_MAP[243]="o";c.CHARACTER_MAP[220]="U";c.CHARACTER_MAP[255]="Y";c.CHARACTER_MAP[214]="O";c.CHARACTER_MAP[241]="n";c.CHARACTER_MAP[209]="N";c.desiredKeyLength=4;c.maxKeyLength=10;c.getTotalLength=function a(k){return k.join("").length};c.removeIgnoredWords=function f(k){return h.reject(k,function(l){return d.inArray(l,c.IGNORED_WORDS)!==-1})};c.createAcronym=function b(l){var k="";d.each(l,function(m,n){k+=n.charAt(0)});return k};c.getFirstSyllable=function i(m){var l=false;var k;for(k=0;k<m.length;k++){if(c.isVowelOrY(m[k])){l=true}else{if(l){return m.substring(0,k+1)}}}return m};c.isVowelOrY=function g(k){return k&&k.length===1&&k.search("[AEIOUY]")!==-1};c.init=function j(l,k){c.desiredKeyLength=l;c.maxKeyLength=k};c.generate=function e(l){l=d.trim(l);if(!l){return""}var k=[];for(var n=0,p=l.length;n<p;n++){var o=c.CHARACTER_MAP[l.charCodeAt(n)];k.push(o?o:l[n])}l=k.join("");var r=[];d.each(l.split(/\s+/),function(s,t){if(t){t=t.replace(/[^a-zA-Z]/g,"");t=t.toUpperCase();t.length&&r.push(t)}});if(c.desiredKeyLength&&c.getTotalLength(r)>c.desiredKeyLength){r=c.removeIgnoredWords(r)}var m;if(r.length==0){m=""}else{if(r.length==1){var q=r[0];if(c.desiredKeyLength&&q.length>c.desiredKeyLength){m=c.getFirstSyllable(q)}else{m=q}}else{m=c.createAcronym(r)}}if(c.maxKeyLength&&m.length>c.maxKeyLength){m=m.substr(0,c.maxKeyLength)}return m};return c});AJS.namespace("JPT.ProjectKeyGenerator",null,require("jira/project-templates/project-key-generator"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/DialogView.js' */
define("jira/project-templates/dialog-view-impl",["require","backbone","underscore"],function(e,f,d){var b=AJS.Dialog;var a=AJS.trigger;return f.View.extend({events:{"click   .add-project-trigger":"_onClickAddProject","click   .add-demo-project-trigger":"_onClickAddDemoProject","click   .add-workflow-link":"_onClickViewMarketplaceWorkflows","click   #add-project-dialog .add-demo-project-trigger":"_onCreateProjectAddDemoProjectClick","click   #add-project-dialog .import-project-trigger":"_onClickImportProject","click   #add-project-dialog .create-with-shared-config":"_onClickCreateShared","keydown #add-project-dialog .pt-templates-list":"_onKeydownInTemplatesList"},draw:function(){var g=new b({width:840,height:400,id:"add-project-dialog",closeOnOutsideClick:false,keypressListener:d.bind(this._onKeyPressed,this)});var h=g.popup.element;g.addPage("project-templates-page");g.addHeader("\u521b\u5efa\u9879\u76ee","add-project-dialog-header");g.addPanel("ProjectTemplatesListPanel",JIRA.Templates.ProjectTemplates.loadingTemplatesList(),"pt-content");this._dialog=g;this._$dialogElement=h;g.show();return g},remove:function(){this._dialog&&this._dialog.remove()},showErrorMessage:function(g){this._dialog.getPanel(1,0).html(JIRA.Templates.errorMsg({closable:false,msg:g}))},get$PTContent:function c(){return this.getDialogController().$dialogElement.find(".pt-content")},_onKeyPressed:function(h){var g=27;if(this._dialog){if(h.keyCode===g){this.getDialogController().hideDialogFromNewUser("dismissed");this._dialog.remove();return false}}return true},_onKeydownInTemplatesList:function(h){var g=13;if(this._dialog){if(h.keyCode===g){this._$dialogElement.find(".pt-submit-button:visible").click();return false}}return true},_onClickAddProject:function(g){g.preventDefault();this.getDialogController().handleProjectTemplateTriggered()},_onCreateProjectAddDemoProjectClick:function(){a("analyticsEvent",{name:"jira.project.templates.dialog.create.demo.create.project.clicked"})},_onClickViewMarketplaceWorkflows:function(){a("analyticsEvent",{name:"jira.project.templates.dialog.create.viewmarketplaceworkflows.clicked"})},_onClickAddDemoProject:function(h){h.preventDefault();var g=this.getDialogController().dialog;if(g&&g.popup&&g.popup.element){g.remove()}this.getDialogController().handleDemoProjectTemplateTriggered()},_onClickImportProject:function(g){this.getDialogController().hideDialogFromNewUser("importproject")},_onClickCreateShared:function(g){this.getDialogController().handleCreateShared()},getDialogController:function(){return e("jira/project-templates/dialog-controller")}})});define("jira/project-templates/dialog-view",["jira/project-templates/dialog-view-impl","jquery"],function(b,a){return new b({el:a(document)})});AJS.namespace("JPT.DialogView",null,require("jira/project-templates/dialog-view"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/DialogController.js' */
define("jira/project-templates/dialog-controller",["jira/project-templates/dialog-view","jira/project-templates/config","jira/ajs/ajax/smart-ajax","jira/util/data/meta","jira/lib/class","jquery","underscore","wrm/data"],function(n,e,a,g,b,k,o,m){var d=AJS.contextPath();var f=AJS.isDevMode;var i=AJS.log;function h(p,q){AJS.trigger("analyticsEvent",{name:p,data:q})}var l;var c=function(){if(!l){l=m.claim("com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources.ptAnalyticsData")||{}}return l};var j=b.extend({WELCOME_REST_URL:d+"/rest/welcome/1.0/show/welcome",initialize:function(){o.bindAll(this,"handleProjectTemplateTriggered","hideDialogFromNewUser");if(g.getBoolean("show-welcome-screen")){this.handleProjectTemplateTriggered()}},handleProjectTemplateTriggered:function(){this.demoProject=false;h("jira.project.templates.dialog.create.show");this.openWithFirstProjectTemplateOfTypePreSelected()},handleDemoProjectTemplateTriggered:function(){this.demoProject=true;var p=this.getTemplateController().loadDemoProjectTemplatesData();this.openWithFirstProjectTemplateOfTypePreSelected(null,p);this._addAnalyticsToCreateDemoShow(p)},_addAnalyticsToCreateDemoShow:function(p){p.done(function(r){var s={};var q=function(t){return t.projectTemplatesGroupedByType.map(function(u){return u.projectTemplates.map(function(v){return v.itemModuleCompleteKey}).join()}).join()};s.demoSets=q(r);s.instanceCreatedDate=c().instanceCreatedDate;h("jira.project.templates.dialog.create.demo.show",s)}.bind(this))},openWithFirstProjectTemplateOfTypePreSelected:function(p,q){q=q||this.getTemplateController().loadProjectTemplatesData();this.dialog=n.draw();q.fail(o.bind(function(r){this._handleUnknownErrorOfPTRetrieval()},this));q.done(o.bind(function(r){this.$dialogElement=this.dialog.popup.element;var s;if(!o.isUndefined(e.model.selectedTemplate)){s=e.model.selectedTemplate}e.model=r;e.model.selectedTemplate=s;this.getTemplateController().init(r,p)},this))},handleCreateShared:function(){h("jira.project.templates.dialog.create.shared.clicked");this.demoProject=false;this.openWithCreateShared()},openWithCreateShared:function(){var p=require("jira/project-templates/create-shared-controller");p.initCreateShared()},callbackWithResize:function(p){return o.bind(o.wrap(p,function(q){q.call(this);this.dialog.updateHeight()}),this)},addPage:function(r){var q=this.dialog.addPage(r.name).page[this.dialog.curpage];this.dialog.addHeader(r.title,"add-project-dialog-header");this.dialog.addPanel(r.panelName,"","pt-content");if(r.backButton){this._addBackButton(q)}this.dialog.addButton(r.submitButtonText,this.callbackWithResize(r.submitButtonCallback),r.submitButtonClass+" pt-submit-button");var p=this.$dialogElement.find("."+r.submitButtonClass);p.removeClass("button-panel-button").addClass("aui-button aui-button-primary");p.focus();this.dialog.addCancel("\u53d6\u6d88",o.bind(function(s){this.hideDialogFromNewUser("dismissed");this.dialog.remove()},this));return q},_backButtonOnClickCallback:function(p){return o.bind(function(){if(this.demoProject){h("jira.project.templates.dialog.demo.back")}else{h("jira.project.templates.dialog.create.back")}var q=n.get$PTContent();q.css("background-image","none");this.dialog.prevPage();p.remove();this.dialog.page.pop()},this)},_addBackButton:function(q){this.dialog.addButton("\u56de\u9000",this.callbackWithResize(this._backButtonOnClickCallback(q)),"add-project-back-button");var p=this.$dialogElement.find(".add-project-back-button");p.removeClass("button-panel-button").addClass("aui-button")},_handleUnknownErrorOfPTRetrieval:function(){n.showErrorMessage("\u8bd5\u56fe\u8fde\u63a5Jira\u7684\u65f6\u5019\u53d1\u751f\u4e86\u4e00\u4e2a\u9519\u8bef\u3002")},hideDialogFromNewUser:function(p){if(p==="dismissed"){if(this.demoProject){h("jira.project.templates.dialog.demo.dismissed")}else{h("jira.project.templates.dialog.create.dismissed")}}else{if(p==="importproject"){h("jira.project.templates.dialog.import.clicked")}else{if(p==="templateselected"){if(this.demoProject){h("jira.project.templates.dialog.demo.templateselected")}else{h("jira.project.templates.dialog.create.templateselected")}}}}if(g.getBoolean("show-welcome-screen")){k.ajax({url:this.WELCOME_REST_URL+"/"+p,type:"DELETE",success:function(){if(f&&f()){i("don't show project template dialog anymore")}}})}},getTemplateController:function(){return require("jira/project-templates/select-project-template-controller")}});return new j()});AJS.namespace("JPT.DialogController",null,require("jira/project-templates/dialog-controller"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/AddProjectView.js' */
define("jira/project-templates/add-project-view-impl",["jira/project-templates/dialog-controller","jira/project-templates/config","jira/util/events","jira/util/events/types","require","backbone","jquery","underscore"],function(q,n,t,y,m,e,g,D){var f=AJS.InlineDialog;return e.View.extend({TIMEOUT_MS:100,postDrawCallbacks:[],events:{"submit #add-project-form":"onSubmitForm"},page:undefined,isSubmitting:false,initialize:function(F){D.bindAll(this,"bindHook")},addPostDrawCallback:function(F){this.postDrawCallbacks.push(F)},prepareDialog:function(F){this.page=q.addPage({name:"add-project",title:F,panelName:"add-project",backButton:true,submitButtonText:"\u63d0\u4ea4",submitButtonCallback:this.onSubmitForm.bind(this),submitButtonClass:"add-project-dialog-create-button"})},draw:function(U){function V(X){if(!X.errors){X.errors={}}return X}this.isSubmitting=false;q.dialog.gotoPage(this.page.id);var G=U.webItemData.projectTemplateModuleCompleteKey;var F=D.find(n.model.projectTemplatesGroupedByType,function(X){return D.any(X.projectTemplates,function(Y){return Y.itemModuleCompleteKey==G})});var L=(F&&F.applicationInfo&&F.applicationInfo.applicationName)||"";var K=(F&&F.applicationInfo&&F.applicationInfo.licenseUsedSeats)||0;var H=(F&&F.applicationInfo&&F.applicationInfo.licenseTotalSeats)||0;var T=(F&&F.applicationInfo&&F.applicationInfo.canUserBeAddedToApplication)||false;var J=(F&&F.applicationInfo&&F.applicationInfo.canUserUseApplication)||false;var P=(F&&F.applicationInfo&&F.applicationInfo.ondemand)||false;var I;if(!F||!F.applicationInfo){I=false}else{I=!J}var R={maxNameLength:n.model.maxNameLength,maxKeyLength:n.model.maxKeyLength,shouldShowLead:n.model.shouldShowProjectLead,projectTemplateWebItemKey:n.model.selectedTemplate||U.webItemData.itemModuleCompleteKey,projectTemplateModuleKey:G,projectTemplateTitle:U.webItemData.name,projectTemplateDescriptionContent:U.webItemData.longDescriptionContent,currentKey:U.currentKey||"",currentName:U.currentName||"",errors:V(U.errors),projectLeadPickerField:{id:"lead",name:"lead"},addUserToLicense:{displayCheckbox:I,disableCheckbox:!T,applicationName:L,usedSeats:K,totalSeats:H,fieldId:"licenseUser",fieldName:"licenseUser",licensingUrl:P?"/admin/accessconfig":"/plugins/servlet/applications/versions-licenses"},leadOptions:[{selected:true,displayName:U.currentUserDisplayName,optionName:U.currentUserName,avatarURL:U.currentUserAvatarUrl}]};var N=JIRA.Templates.ProjectTemplates.addProjectForm(R);q.dialog.getPanel(this.page.id,0).html(N);if(U.webItemData&&U.webItemData.backgroundIconUrl){var Q=q.$dialogElement.find(".pt-content");Q.css("background-image",'url("'+U.webItemData.backgroundIconUrl+'")')}var O=g("#add-project-form");t.trigger(y.NEW_CONTENT_ADDED,[O]);this.nameElement=q.$dialogElement.find("#name");this.keyElement=q.$dialogElement.find("#key");this.keyEditedElement=q.$dialogElement.find("#keyEdited");this.leadDisplayElement=q.$dialogElement.find("#lead-field");this.leadValueElement=q.$dialogElement.find("#lead");var W=this.keyElement.parent().find(".aui-icon-help");if(W.length){new f(W,"project-key-help-popup",function(Z,Y,X){Z.html(JIRA.Templates.ProjectTemplates.keyHelp());X()},{width:330,offsetX:-30})}this.keyElement.attr("style","text-transform: uppercase");var S=m("jira/project-templates/add-project-controller");this.nameElement.focus(D.bind(function(X){this.bindHook(X,S.nameTimeout)},this));var M=this;this.nameElement.change(function(X){S.validateName();M.unbindHook(X)});this.nameElement.focus();this.keyElement.focus(D.bind(function(Y){var X=g(Y.target);X.data("lastValue",X.val());this.bindHook(Y,S.keyTimeout)},this));this.keyElement.blur(D.bind(function(X){this.unbindHook(X)},this));this.keyElement.change(function(){S.validateKey();S.autofillKeyIfNeeded()});if(!D.isEmpty(this.postDrawCallbacks)){D.each(this.postDrawCallbacks,function(X){X()})}q.dialog.updateHeight()},onSubmitForm:function(G){var F=m("jira/project-templates/add-project-controller");F.submit();return false},get$SubmitButton:function(){return q.$dialogElement.find(".add-project-dialog-create-button")},get$BackButton:function(){return q.$dialogElement.find(".add-project-back-button")},bindHook:function p(I,G){var F=g(I.target),H;H=D.bind(function(){this.unbindHook(I);G.apply();if(F.is(":visible")){F.data("checkHook",setTimeout(H,this.TIMEOUT_MS))}},this);if(!F.data("checkHook")){F.data("checkHook",setTimeout(H,0))}},unbindHook:function(G){var F=g(G.target);clearTimeout(F.data("checkHook"));F.removeData("checkHook")},showInlineError:function(G,H){if(this.isSubmitting){return}var F=G.parent().find(".error");if(!F.length){F=g("<div class='error'></div>");G.parent().append(F)}F.text(H);F.show()},showInlineErrorForName:function E(F){this.showInlineError(this.nameElement,F)},showInlineErrorForKey:function A(F){this.showInlineError(this.keyElement,F)},hideInlineError:function z(F){F.parent().find(".error").hide()},hideInlineErrorForName:function c(){this.hideInlineError(this.nameElement)},hideInlineErrorForKey:function u(){this.hideInlineError(this.keyElement)},setName:function i(F){this.nameElement.val(F)},getName:function s(){return this.nameElement.val()},setKey:function b(F){this.keyElement.val(F)},getKey:function o(){return this.keyElement.val().toUpperCase()},getLeadDisplayName:function v(){return this.leadDisplayElement.val()},getLeadUserName:function w(){return this.leadValueElement.val()},getAvatarUrlOfSelectedLead:function r(){var F=q.$dialogElement.find("#lead-single-select .aui-ss-entity-icon").css("background-image");if(!D.isUndefined(F)){var G=F.match(/^url\((.+)\)$/);return(G&&G[1])?G[1]:""}else{return""}},setKeyEdited:function l(F){this.keyEditedElement.val(F)},getKeyEdited:function B(){return this.keyEditedElement.val()},setKeyEdited:function l(){var F=this.getKey();if(this.keyElement.data("lastValue")!==F){this.keyEditedElement.val((F)?"true":"false")}this.keyElement.data("lastValue",F)},hasNameErrors:function d(){return this.nameElement.parent().find(".error").size()>0},getAddProjectForm:function h(){return g("#add-project-form")},get$FormFields:function k(){return q.$dialogElement.find(":input")},enterLoadingState:function j(){$submitButton=this.get$SubmitButton();if(!$submitButton.attr("disabled")){$backButton=this.get$BackButton();$submitButton.attr("disabled","disabled");$backButton.attr("disabled","disabled");$backButton.before(JIRA.Templates.ProjectTemplates.spinner({id:"addproject-loading"}));this.get$FormFields().attr("disabled","disabled");this.isSubmitting=true;return true}else{return false}},hideLoadingState:function x(){q.$dialogElement.find("#addproject-loading").remove();this.get$SubmitButton().removeAttr("disabled");this.get$BackButton().removeAttr("disabled");this.get$FormFields().removeAttr("disabled")},avoidDirtyFormWarning:function a(){if(g.fn.removeDirtyWarning){this.getAddProjectForm().removeDirtyWarning()}},hasInlineErrors:function C(){return q.$dialogElement.find(".field-group .error:visible").length!=0}})});define("jira/project-templates/add-project-view",["jira/project-templates/add-project-view-impl","jira/project-templates/dialog-view","jquery"],function(b,a,c){return new b({el:c(document),dialogView:a})});AJS.namespace("JPT.AddProjectView",null,require("jira/project-templates/add-project-view"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/AddProjectController.js' */
define("jira/project-templates/add-project-controller-impl",["jira/project-templates/add-project-view","jira/project-templates/project-key-generator","jira/project-templates/dialog-controller","jira/project-templates/config","jira/lib/class","jquery","underscore"],function(m,i,k,e,b,l,o){var d=AJS.contextPath();var c=AJS.trigger;return b.extend({postProjectCreationCallbacks:[],projectKeyValidationCallbacks:[],projectNamesUpperCased:[],init:function(p){this._window=p.window||window;o.bindAll(this,"projectCreatedHandler","projectValidationFailedHandler","nameTimeout","keyTimeout","submit")},initCreateShared:function(p){this.existingProjectData=p;this._initAddProject("\u521b\u5efa\u4e0e\u5171\u4eab\u914d\u7f6e")},initProjectTemplate:function(p){this.existingProjectData=undefined;this.webItemData=p;this._initAddProject(p.name)},_initAddProject:function(p){this.desiredKeyLength=4;this.lastKeyValidated="";this._getExistingProjects();i.init(this.desiredKeyLength,e.model.maxKeyLength);m.prepareDialog(p);m.draw({webItemData:this.webItemData||{},maxKeyLength:e.model.maxKeyLength,maxNameLength:e.model.maxNameLength,currentUserName:e.model.currentUserName,currentUserDisplayName:e.model.currentUserDisplayName,currentUserAvatarUrl:e.model.currentUserAvatarUrl,errors:{}})},_resetProjectLeadValues:function(){e.model.currentUserDisplayName="";e.model.currentUserName="";e.model.currentUserAvatarUrl=""},_hasFullyConfiguredProjectLead:function(){return m.getLeadDisplayName()&&m.getAvatarUrlOfSelectedLead()!=""},submit:function(p){var q;if(this.existingProjectData){q={key:m.getKey(),name:m.getName(),lead:m.getLeadUserName()&&m.getLeadUserName()[0]}}else{q=jQuery.param(m.getAddProjectForm().serializeArray().map(function(r){if(r.name==="key"){r.value=r.value.toUpperCase()}return r}))}if(m.hasInlineErrors()){return}if(m.enterLoadingState()){if(this._hasFullyConfiguredProjectLead()){e.model.currentUserDisplayName=m.getLeadDisplayName();e.model.currentUserName=e.model.currentUserDisplayName?m.getLeadUserName():"";e.model.currentUserAvatarUrl=m.getAvatarUrlOfSelectedLead()}else{this._resetProjectLeadValues()}if(this.existingProjectData){l.ajax({url:d+"/rest/project-templates/1.0/createshared/"+this.existingProjectData.existingProjectId,type:"POST",contentType:"application/json",data:JSON.stringify(q)}).done(this.projectCreatedHandler).fail(this.projectValidationFailedHandler)}else{l.ajax({url:d+(this.webItemData.demoProject?"/rest/jira-importers-plugin/1.0/demo/create":"/rest/project-templates/1.0/templates"),type:"POST",data:q,headers:{"X-Atlassian-Token":"no-check"}}).done(this.projectCreatedHandler).fail(this.projectValidationFailedHandler)}}},registerPostProjectCreationCallback:function(p){this.postProjectCreationCallbacks.push(p)},registerProjectKeyValidationCallback:function(p){this.projectKeyValidationCallbacks.push(p)},localStoragePrefix:"jira.projecttemplates.",projectCreatedHandler:function(q){if(this.existingProjectData){c("analyticsEvent",{name:"jira.project.templates.dialog.create.shared.project.create.success"})}else{if(this.webItemData.demoProject){c("analyticsEvent",{name:"jira.project.templates.dialog.create.demo.success"})}else{c("analyticsEvent",{name:"jira.project.templates.dialog.create.project.success"})}}m.avoidDirtyFormWarning();var p=this.localStoragePrefix;var r=["confluenceProject","fisheyeProject","crucibleProject","bambooProject"];o.map(r,function(s){if(q.remoteProjectLinks&&q.remoteProjectLinks[s]){localStorage.setItem(p+s,q.remoteProjectLinks[s])}});if(!o.isEmpty(this.postProjectCreationCallbacks)){k.dialog.addPage("post-project-created-page");o.each(this.postProjectCreationCallbacks,function(s){s(k.dialog,q.returnUrl,q.projectId,q.projectKey,q.projectName)})}else{this._window.location=d+q.returnUrl}},projectValidationFailedHandler:function(q,s){if(this.existingProjectData){c("analyticsEvent",{name:"jira.project.templates.dialog.create.shared.project.create.failure"})}else{if(this.webItemData.demoProject){c("analyticsEvent",{name:"jira.project.templates.dialog.create.demo.failure"})}else{c("analyticsEvent",{name:"jira.project.templates.dialog.create.project.failure"})}}var r={};if(this.isBadRequest(q)){r=JSON.parse(q.responseText)}else{if(this.isUnDefinedServerSideError(q)){r={errorMessages:["\u55ef...\u6211\u4eec\u65e0\u6cd5\u521b\u5efa\u60a8\u7684\u9879\u76ee\u7531\u4e8e\u672a\u77e5\u7684\u9519\u8bef\u3002\u8bf7\u5c1d\u8bd5\u5237\u65b0\u9875\u9762\u5e76\u91cd\u65b0\u5f00\u59cb\u3002"]}}else{if(this.isDefinedServerSideError(q)){var p=JSON.parse(q.responseText);JIRA.Messages.showReloadErrorMsg(p.message);m.avoidDirtyFormWarning();this._window.location=d+p.returnUrl;return}else{if(this.isTimeoutError(s)){r={errorMessages:["\u521b\u5efa\u9879\u76ee\u7684\u65f6\u5019\u8bf7\u6c42\u8d85\u65f6"]}}else{r={errorMessages:[AJS.format("\u521b\u5efa\u9879\u76ee {0} \u65f6\u53d1\u751f\u9519\u8bef",q.responseText)]}}}}}m.draw({webItemData:this.webItemData||{},errors:r,currentName:m.getName(),currentKey:m.getKey(),currentUserDisplayName:e.model.currentUserDisplayName,currentUserName:e.model.currentUserName,currentUserAvatarUrl:e.model.currentUserAvatarUrl});m.hideLoadingState()},isBadRequest:function(p){return(p.status===400)},isUnDefinedServerSideError:function(p){if(p.status===500){try{JSON.parse(p.responseText)}catch(q){return true}}return false},isDefinedServerSideError:function(p){return p.status===500&&!o.isUndefined(JSON.parse(p.responseText).message)},isTimeoutError:function(p){return p==="timeout"},_updateAndValidateKey:function a(p){m.setKey(p);this.validateKey()},_shouldUpdateKey:function h(){return(m.getKeyEdited()!="true")},autofillKeyIfNeeded:function n(){if(this._shouldUpdateKey()){var p=i.generate(m.getName());if(p.length>1){this._updateAndValidateKey(p)}else{m.setKey("")}}},_doesProjectNameExists:function(q){var p;for(p in this.projectNamesUpperCased){if(q.toUpperCase()==this.projectNamesUpperCased[p]){return true}}return false},validateName:function(){var p=l.trim(m.getName());if(!p){return}if(p.length<e.model.minNameLength){m.showInlineErrorForName(AJS.format("\u9879\u76ee\u540d\u79f0\u81f3\u5c11\u5e94\u8be5 {0} \u4e2a\u5b57\u7b26\u3002",e.model.minNameLength));return}if(p.length>e.model.maxNameLength){m.showInlineErrorForName(AJS.format("\u9879\u76ee\u540d\u79f0\u957f\u5ea6\u4e0d\u80fd\u8d85\u8fc7 {0} \u4e2a\u5b57\u7b26\u3002",e.model.maxNameLength));return}if(this._doesProjectNameExists(p)){m.showInlineErrorForName("\u5df2\u7ecf\u5b58\u5728\u540c\u540d\u9879\u76ee");return}m.hideInlineErrorForName()},_performKeyValidationChecks:function(p){var q=l.ajax({url:d+"/rest/api/latest/projectvalidate/key?key="+p.toUpperCase()});q.done(o.bind(function(s){if(s.errors&&s.errors.projectKey){m.showInlineErrorForKey(s.errors.projectKey)}else{var r=false;o.each(this.projectKeyValidationCallbacks,function(u){var t=u(p.toUpperCase());if(t.errors&&t.errors.projectKey){r=true;m.showInlineErrorForKey(t.errors.projectKey)}});if(!r){m.hideInlineErrorForKey()}}},this))},validateKey:function g(){var p=m.getKey();if(this.lastKeyValidated===p){return}if(p){this.lastKeyValidated=p;this._performKeyValidationChecks(p)}else{m.hideInlineErrorForKey()}},nameTimeout:function j(){this.autofillKeyIfNeeded()},keyTimeout:function f(){m.setKeyEdited()},_getExistingProjects:function(){if(this.projectNamesUpperCased.length>0){return this.projectNamesUpperCased}var p=l.ajax({url:d+"/rest/api/latest/project"});p.done(o.bind(function(q){this.projectNamesUpperCased=o.map(q,function(r){return r.name.toUpperCase()})},this))}})});define("jira/project-templates/add-project-controller",["jira/project-templates/add-project-controller-impl","jquery"],function(a,b){return new a({el:b(document)})});AJS.namespace("JPT.AddProjectController",null,require("jira/project-templates/add-project-controller"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/TemplateInfoView.js' */
define("jira/project-templates/template-info-view",["jira/project-templates/dialog-controller","require","backbone","jquery","underscore"],function(e,c,f,d,b){var a=AJS.trigger;return f.View.extend({initialize:function(g){b.bindAll(this,"draw","onNext")},draw:function(j,h){var k=e.addPage({name:"template-info",title:h,panelName:"template-info",backButton:true,submitButtonText:"\u9009\u62e9",submitButtonCallback:this.onNext,submitButtonClass:"template-info-dialog-create-button"});var g=this._nameToObj(j)();e.dialog.getPanel(k.id,0).html(g);e.dialog.gotoPage(k.id);var i=b.find(k.button,function(l){return l.item&&l.item.hasClass("pt-submit-button")});if(i){setTimeout(function(){i.item.focus()},0)}},_nameToObj:function(g){return b.reduce(g.split("."),function(i,h){if(i){return i[h]}},window)},onNext:function(h){a("analyticsEvent",{name:"jira.project.templates.dialog.create.templateinfo.next"});var g=c("jira/project-templates/template-info-controller");g.next();return false}})});AJS.namespace("JPT.TemplateInfoView",null,require("jira/project-templates/template-info-view"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/TemplateInfoController.js' */
define("jira/project-templates/template-info-controller",["jira/project-templates/template-info-view","jira/lib/class","require","underscore"],function(e,a,c,b){var d=a.extend({init:function(f){b.bindAll(this,"initTemplateInfo","next")},initTemplateInfo:function(g){this.projectTemplateData=g;var f=new e();f.draw(g.infoSoyPath,g.name)},next:function(){var f=c("jira/project-templates/select-project-template-controller");f.openAddProjectPage(this.projectTemplateData)}});return new d()});AJS.namespace("JPT.TemplateInfoControllerImpl",null,require("jira/project-templates/template-info-controller"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/SelectProjectTemplateView.js' */
define("jira/project-templates/select-project-template-view",["jira/project-templates/dialog-controller","jira/project-templates/config","jira/featureflags/feature-manager","require","backbone","jquery","underscore"],function(d,b,i,c,g,e,h){var a=AJS.contextPath();var f=AJS.preventDefault;return{ROW_LENGTH:2,DIALOG_BODY_CLASS:"select-project-templates-page",DIALOG_WIDTH:800,draw:function(n,m){var j=c("jira/project-templates/select-project-template-controller");this.page=d.addPage({name:this.DIALOG_BODY_CLASS,title:n.demoProjects?"\u4f7f\u7528\u793a\u4f8b\u6570\u636e\u521b\u5efa\u65b0\u9879\u76ee":"\u521b\u5efa\u9879\u76ee",panelName:"ProjectTemplatesListPanel",backButton:false,submitButtonText:"\u4e0b\u4e00\u6b65",submitButtonCallback:j.dialogSubmitted,submitButtonClass:"create-project-dialog-create-button"});function p(){if(j.isProjectTypesEnabled(n)){return JIRA.Templates.ProjectTemplates.renderProjectTemplatesGroupedByType({projectTemplatesByType:n.projectTemplatesGroupedByType})}return JIRA.Templates.ProjectTemplates.renderProjectTemplates({projectTemplates:n.projectTemplates})}d.dialog.getPanel(this.page.id,0).html(p());if(b.model.projectCount>=2&&!n.demoProjects){e(JIRA.Templates.ProjectTemplates.addWorkflowsLink({baseUrl:a})).appendTo(d.$dialogElement.find(".dialog-title"))}var o=d.$dialogElement.find(".create-project-dialog-create-button");var l=this;this.getTemplateItems().click(function(){var q=e(this);q.addClass("selected");l.getTemplateItems().not(this).removeClass("selected")}).dblclick(function(){o.click()}).focus(function(){e(this).click()});var k=this.get$TemplatesContainer();if(!n.demoProjects){e(JIRA.Templates.ProjectTemplates.footerLinks({showDemoLink:i.isFeatureEnabled("jira.onboarding.cyoa")})).prependTo(d.$dialogElement.find(".dialog-button-panel"))}this.bindKeyboardEvents(k);if(j.isProjectTypesEnabled(n)){if((n.projectTemplatesGroupedByType).length==0){o.attr("disabled","disabled")}}else{if((n.projectTemplates).length==0){o.attr("disabled","disabled")}}this.focusOnFirstTemplate(k,m);d.dialog.updateHeight();this.truncateTemplateDescriptions()},get$TemplatesContainer:function(){return d.$dialogElement.find(".templates")},focusOnFirstTemplate:function(j,k){if(k){this.getFirstTemplateItemOfProjectType(k).click()}else{this.getFirstTemplateItem().click()}setTimeout(function(){j.focus()},0)},getSelectedTemplateData:function(){var j=d.$dialogElement.find(".template.selected");return j.data()},getMoveDeltaForKey:function(j){switch(j){case 37:return -1;case 39:return +1;case 38:return -this.ROW_LENGTH;case 40:return +this.ROW_LENGTH}return 0},bindKeyboardEvents:function(j){j.bind("keydown",h.bind(function(k){var l=this.getMoveDeltaForKey(k.which);if(l){this.moveSelection(j,l);return f(k)}},this))},moveSelection:function(l,o){var k=l.find(".template");var n=k.filter(".selected");var m=k.index(n)+o;if(m<k.length&&m>=0){var j=k.eq(m);j.click().focus();this.scrollToSelectedElement(j)}},scrollToSelectedElement:function(l){var o=e(".dialog-panel-body.pt-content");var j=o.offset().top;var k=l.offset().top;var n=k+l.height();var m=j+o.height();if(k<j){o.scrollTop(o.scrollTop()-(j-k))}else{if(n>(j+o.height())){o.scrollTop(o.scrollTop()+n-m)}}},get$NextButton:function(){return d.$dialogElement.find(".create-project-dialog-create-button")},disableNextButton:function(){this.get$NextButton().attr("disabled","disabled")},truncateTemplateDescriptions:function(){var j=AJS.Meta.get("user-locale");var k="word";if(j==="ja_JP"){k="letter"}d.$dialogElement.find(".template-description").each(function(){e(this).dotdotdot({wrap:k,lastCharacter:{remove:[" ",",",";",".","!","?","。"],noEllipsis:[]}})})},getFirstTemplateItem:function(){return this.getTemplateItems().first()},getFirstTemplateItemOfProjectType:function(j){return this.getTemplateItems().filter("#project-template-group-"+j+" *").first()},getTemplateItems:function(){return d.$dialogElement.find(".template")}}});AJS.namespace("JPT.SelectProjectTemplateView",null,require("jira/project-templates/select-project-template-view"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/SelectProjectTemplateController.js' */
define("jira/project-templates/select-project-template-controller",["jira/project-templates/select-project-template-view","jira/project-templates/template-info-controller","jira/project-templates/add-project-controller","jira/project-templates/dialog-controller","jira/project-templates/dialog-view","jira/project-templates/config","wrm/context-path","jquery","underscore"],function(f,h,k,l,o,g,n,m,p){var e=n();var d={init:function q(s,r){f.draw(s,r)},isProjectTypesEnabled:function c(r){return r.projectTemplatesGroupedByType!=null},loadProjectTemplatesData:function j(){return m.ajax({url:e+"/rest/project-templates/1.0/templates",type:"GET"})},loadDemoProjectTemplatesData:function j(){return m.ajax({url:e+"/rest/project-templates/1.0/templates/demo-projects",type:"GET"}).done(function(r){r.demoProjects=true;return r})},dialogSubmitted:function b(){var r=f.getSelectedTemplateData();l.hideDialogFromNewUser("templateselected");d.raiseAtlassianEvent(r.itemModuleCompleteKey);if(!p.isUndefined(r.infoSoyPath)){h.initTemplateInfo(r)}else{d.openAddProjectPage(r)}},openAddProjectPage:function i(r){if(!r){o.showErrorMessage("\u8bd5\u56fe\u8fde\u63a5Jira\u7684\u65f6\u5019\u53d1\u751f\u4e86\u4e00\u4e2a\u9519\u8bef\u3002");return}if(r.createProject){k.initProjectTemplate(r)}else{l.dialog.addPage("blank-template-page");m("body").trigger(r.itemModuleCompleteKey,l.dialog)}},raiseAtlassianEvent:function a(r){g.model.selectedTemplate=r;if(AJS.EventQueue){AJS.EventQueue.push({name:"projecttemplates.templateselected",properties:{selectedTemplate:r}})}}};return d});AJS.namespace("JPT.SelectProjectTemplateController",null,require("jira/project-templates/select-project-template-controller"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/CreateSharedView.js' */
define("jira/project-templates/create-shared-view",["jira/project-templates/dialog-controller","jira/project-templates/dialog-view","jira/ajs/select/scrollable-single-select","jira/ajs/select/suggestion-collection-model","require","backbone","jquery"],function(d,h,g,i,c,j,e){var f=AJS.InlineDialog;var b=AJS.trigger;var a=j.View.extend({events:{"submit #create-shared-project-form":"onSubmitForm"},page:undefined,prepareDialog:function(l){var k=c("jira/project-templates/create-shared-controller");this.page=d.addPage({name:"create-shared-project",title:l,panelName:"create-shared-project",backButton:true,submitButtonText:"\u4e0b\u4e00\u6b65",submitButtonCallback:k.dialogSubmitted,submitButtonClass:"create-shared-dialog-button"})},draw:function(l){d.dialog.gotoPage(this.page.id);var k=JIRA.Templates.ProjectTemplates.createSharedProjectForm({projectSuggestions:JSON.stringify(l.projectSuggestions)});d.dialog.getPanel(this.page.id,0).html(k);this._createProjectPicker();var m=d.$dialogElement.find("#shared-help-icon");if(m.length){new f(m,"shared-project-help-popup",function(p,o,n){p.html(JIRA.Templates.ProjectTemplates.sharedHelp());n()},{width:330,offsetX:-30})}d.$dialogElement.find(".dialog-button-panel button").removeAttr("disabled")},showProjectMissingError:function(){this._clearFormErrors();this._getProjectPickerInput().after(aui.form.fieldError({extraClasses:"project-picker-missing-error",message:"\u8bf7\u5148\u9009\u62e9\u4e00\u4e2a\u73b0\u6709\u7684\u9879\u76ee"}))},drawEmptyInfo:function(){var k=JIRA.Templates.ProjectTemplates.noProjects();d.dialog.getPanel(this.page.id,0).html(k);d.$dialogElement.find(".dialog-button-panel button").hide()},drawError:function(k){d.dialog.getPanel(this.page.id,0).html(JIRA.Templates.errorMsg({closable:false,msg:k}))},drawLoading:function(){d.dialog.gotoPage(this.page.id);var k=JIRA.Templates.ProjectTemplates.loading();d.dialog.getPanel(this.page.id,0).html(k);d.$dialogElement.find(".dialog-spinner").spin();d.$dialogElement.find(".dialog-button-panel button").attr("disabled","disabled")},_clearFormErrors:function(){d.$dialogElement.find(".project-picker-missing-error").remove()},_getProjectPickerInput:function(){return d.$dialogElement.find("#project-picker")},onSubmitForm:function(l){this._clearFormErrors();var k=c("jira/project-templates/create-shared-controller");k.dialogSubmitted();return false},_getExtraInfoMessage:function(){return d.$dialogElement.find(".create-shared-info")},_createProjectPicker:function(){this._getExtraInfoMessage().hide();this.projectSelect=new g({element:this._getProjectPickerInput(),revertOnInvalid:true,pageSize:50,pagingThreshold:100,model:i});this.projectSelect.$field.focus();var k=this;this._getProjectPickerInput().on("selected",function(m,l){if(l.value()){b("analyticsEvent",{name:"jira.project.templates.dialog.create.shared.project.selected"});k._getExtraInfoMessage().show()}})},getSelectedProject:function(){return this.projectSelect.getSelectedDescriptor()&&this.projectSelect.getSelectedDescriptor().value()}});return new a({el:e(document),dialogView:h})});AJS.namespace("JPT.CreateSharedView",null,require("jira/project-templates/create-shared-view"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/CreateSharedController.js' */
define("jira/project-templates/create-shared-controller-impl",["jira/project-templates/add-project-controller","jira/project-templates/create-shared-view","jira/lib/class","jquery","underscore"],function(e,f,a,d,c){var b=AJS.contextPath();return a.extend({init:function(g){this._window=g.window||window;c.bindAll(this,"dialogSubmitted")},initCreateShared:function(){f.prepareDialog("\u521b\u5efa\u4e0e\u5171\u4eab\u914d\u7f6e");var g=this;this._getProjectSuggestions().done(function(h){if(g._hasSuggestions(h)){f.draw({projectSuggestions:h})}else{f.drawEmptyInfo()}}).fail(function(){f.drawError("\u8bd5\u56fe\u8fde\u63a5Jira\u7684\u65f6\u5019\u53d1\u751f\u4e86\u4e00\u4e2a\u9519\u8bef\u3002")});f.drawLoading()},_hasSuggestions:function(h){var g=false;d.each(h,function(j,k){if(k&&k.items&&k.items.length>0){g=true}});return g},_getProjectSuggestions:function(){return d.ajax({url:b+"/rest/project-templates/1.0/createshared"})},dialogSubmitted:function(){var g=f.getSelectedProject();if(g){e.initCreateShared({existingProjectId:g})}else{f.showProjectMissingError()}}})});define("jira/project-templates/create-shared-controller",["jira/project-templates/create-shared-controller-impl","jquery"],function(a,b){return new a({el:b(document)})});AJS.namespace("JPT.CreateSharedController",null,require("jira/project-templates/create-shared-controller"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.project-templates-plugin:project-templates-plugin-resources', location = '/js/RemoteProjectsCreatedMessage.js' */
var $=require("jquery");$(function(){var i=AJS.format;var e=require("jira/project-templates/add-project-controller");var b=localStorage.getItem(e.localStoragePrefix+"confluenceProject");var j=localStorage.getItem(e.localStoragePrefix+"fisheyeProject");var g=localStorage.getItem(e.localStoragePrefix+"crucibleProject");var d=localStorage.getItem(e.localStoragePrefix+"bambooProject");localStorage.removeItem(e.localStoragePrefix+"confluenceProject");localStorage.removeItem(e.localStoragePrefix+"fisheyeProject");localStorage.removeItem(e.localStoragePrefix+"crucibleProject");localStorage.removeItem(e.localStoragePrefix+"bambooProject");var k;function c(m,l){return JIRA.Templates.ProjectTemplates.formatAnchor({href:m,body:l,title:l})}function a(n,o,m,p){var l=[];l.push("Jira \u9879\u76ee");if(n){l.push(c(n,"Confluence \u7a7a\u95f4"))}if(o){l.push(c(o,"FishEye\u5e93"))}if(m){l.push(c(m,"Crucible\u9879\u76ee"))}if(p){l.push(c(p,"Bamboo\u9879\u76ee"))}return l}function h(m,o,l,r){var q="\u4ee5\u4e0b\u5b9e\u4f53\u7684\u521b\u5efa: {0}\u548c{1}\u3002";var n=a(m,o,l,r);var p=n.pop();return(n.length>0)?i(q,n.join(", "),p):null}function f(l){var m=$(l).offset();if(m){window.scrollTo(m.left,m.top)}}if(b||j||g||d){k=h(b,j,g,d);if(k){JIRA.Messages.showSuccessMsg(k,{closeable:true})}f("#project-config-webpanel-summary-settings")}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira-core-project-templates:jira-core-project-templates-resources', location = '/soy/CoreProjectTemplates.soy' */
// This file was automatically generated from CoreProjectTemplates.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.ProjectTemplates.CoreTemplates.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.ProjectTemplates == 'undefined') { JIRA.Templates.ProjectTemplates = {}; }
if (typeof JIRA.Templates.ProjectTemplates.CoreTemplates == 'undefined') { JIRA.Templates.ProjectTemplates.CoreTemplates = {}; }


JIRA.Templates.ProjectTemplates.CoreTemplates.taskManagementInfoDialog = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog({description: '\u521b\u5efa\u7b80\u5355\u7684\u4efb\u52a1, \u7ec4\u7ec7\u4ed6\u4eec, \u8ba9\u4ed6\u4eec\u53bb\u505a\u3002\u60a8\u53ef\u4ee5\u4f7f\u7528\u6b64\u9879\u76ee\u6765\u7ba1\u7406\u60a8\u7684\u4efb\u52a1\u6216\u5c06\u5b83\u4eec\u5206\u914d\u7ed9\u5176\u4ed6\u4eba\u3002', projectTemplate: 'taskManagement'});
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.taskManagementInfoDialog.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.taskManagementInfoDialog';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.projectManagementInfoDialog = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog({description: '\u521b\u5efa\u60a8\u7684\u4efb\u52a1\u3001\u7ec4\u7ec7\u548c\u8ddf\u8e2a\u4ed6\u4eec\u7684\u8fdb\u5c55\u3001\u63d0\u4f9b\u60a8\u7684\u5de5\u4f5c\u65f6\u95f4\u3002\u4f30\u8ba1\u548c\u65f6\u95f4\u8ddf\u8e2a\u8ba9\u60a8\u62a5\u544a\u5728\u54ea\u91cc\u60a8\u7684\u9879\u76ee\u6240\u5904\u7684\u9636\u6bb5\u3002', projectTemplate: 'projectManagement'});
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.projectManagementInfoDialog.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.projectManagementInfoDialog';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.processManagementInfoDialog = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog({description: '\u521b\u5efa\u60a8\u7684\u4efb\u52a1\u548c\u8ddf\u8e2a\u4ed6\u4eec\u7684\u6bcf\u4e2a\u6b65\u9aa4, \u4ece\u5f00\u59cb\u5230\u7ed3\u675f\u3002\u60a8\u53ef\u4ee5\u4f7f\u7528\u6b64\u9879\u76ee, \u5ba1\u67e5\u6587\u6863\u3001\u5ba1\u6279\u8d39\u7528\u3001\u6216\u5176\u4ed6\u8fdb\u7a0b\u3002', projectTemplate: 'processManagement'});
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.processManagementInfoDialog.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.processManagementInfoDialog';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.boardBetaInfoDialog = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog({description: '\u8de8\u56e2\u961f\u677f\u9762\u53ef\u89c6\u5316\u5de5\u4f5c\u3002', projectTemplate: 'projectManagement', issueTypesHtml: '' + JIRA.Templates.ProjectTemplates.CoreTemplates.taskIssueType(null)});
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.boardBetaInfoDialog.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.boardBetaInfoDialog';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog = function(opt_data, opt_ignored) {
  return '<div><div class="top-panel"><p>' + soy.$$escapeHtml(opt_data.description) + '</p></div><div class="left-panel"><h6>' + soy.$$escapeHtml('\u95ee\u9898\u7c7b\u578b') + '</h6><ul class="project-template-issuetype-list">' + ((opt_data.issueTypesHtml != null) ? soy.$$filterNoAutoescape(opt_data.issueTypesHtml) : JIRA.Templates.ProjectTemplates.CoreTemplates.taskIssueType(null) + JIRA.Templates.ProjectTemplates.CoreTemplates.subtaskIssueType(null)) + '</ul></div><div class="right-panel"><h6>' + soy.$$escapeHtml('\u5de5\u4f5c\u6d41') + '</h6><div class="workflow ' + soy.$$escapeHtml(opt_data.projectTemplate) + '"></div></div></div>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.templateInfoDialog';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.issueType = function(opt_data, opt_ignored) {
  return '<li><span class="issuetype-list-label"><span class="issuetype-icon ' + soy.$$escapeHtml(opt_data.iconKey) + '"></span><span class="issuetype-name">' + soy.$$escapeHtml(opt_data.label) + '</span></span></li>';
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.issueType.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.issueType';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.taskIssueType = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.ProjectTemplates.CoreTemplates.issueType({iconKey: 'task', label: '\u4efb\u52a1'});
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.taskIssueType.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.taskIssueType';
}


JIRA.Templates.ProjectTemplates.CoreTemplates.subtaskIssueType = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.ProjectTemplates.CoreTemplates.issueType({iconKey: 'subtask', label: '\u5b50\u4efb\u52a1'});
};
if (goog.DEBUG) {
  JIRA.Templates.ProjectTemplates.CoreTemplates.subtaskIssueType.soyTemplateName = 'JIRA.Templates.ProjectTemplates.CoreTemplates.subtaskIssueType';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-analytics-tracker', location = 'includes/js/rapid/AnalyticsTracker.js' */
define("jira-agile/rapid/analytics-tracker",["jira/util/logger","underscore","jquery"],function(e,a,i){function t(e,i,t){if(a.isUndefined(e))throw new Error("You must specify a category");this.params={},this.params.category=e;var r=0;a.isUndefined(i)||(this.params.action=i,r++,a.isUndefined(t)||(this.params.label=t,r++)),this.requiredArgumentNames=["action","label","value"].slice(r),this.useAsync=!0}return t.prototype.setAsync=function(e){return this.useAsync=e,this},t.prototype.trigger=function(){this.requiredArgumentNames.length<arguments.length&&e.log("Warning: too many arguments passed in. Needed at most "+this.requiredArgumentNames.length+" but got "+arguments.length);for(var t=a.clone(this.params),r=0;r<arguments.length&&!a.isUndefined(this.requiredArgumentNames[r]);r++)t[this.requiredArgumentNames[r]]=arguments[r];this._validateParams(t);var n="gh.analytics.async";i(document).trigger(n,t)},t.prototype._validateParams=function(e){var i=["category","action","label"];a.each(i,function(i){a.isUndefined(e[i])||a.isString(e[i])?a.isUndefined(e[i])&&(e[i]=""):e[i]=e[i].toString()}),a.each(["category","action"],function(a){e[a]=e[a].replace(/\s+/g,"")}),a.isUndefined(e.value)||(e.value=parseInt(e.value,10),isNaN(e.value)&&(e.value=void 0))},t}),AJS.namespace("GH.AnalyticsTracker",null,require("jira-agile/rapid/analytics-tracker"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-help-analytics', location = 'includes/js/rapid/ui/help/HelpAnalytics.js' */
AJS.$(function(){var e=require("jira-agile/rapid/analytics-tracker"),r=function(r){return new e(r).setAsync(!1)};AJS.$("#gh_view_help").click(function(e){return r("gh.gethelp").trigger("usermenu","docs"),n(AJS.$(this))}),AJS.$("#greenhopper-my-jira-home-enablement").click(function(e){return r("gh.myjirahome").trigger("usermenu","enabled"),n(AJS.$(this))}),AJS.$("#greenhopper-my-jira-home-enablement-ondemand").click(function(e){return r("gh.myjirahome").trigger("usermenu","set"),n(AJS.$(this))}),AJS.$("#greenhopper-my-jira-home-disablement-ondemand").click(function(e){return r("gh.myjirahome").trigger("usermenu","disable"),n(AJS.$(this))}),AJS.$("#greenhopper-my-jira-home-set").click(function(e){return r("gh.myjirahome").trigger("usermenu","set"),n(AJS.$(this))}),AJS.$("#beta_gh\\.configuration\\.rapid\\.removal").click(function(e){var n=AJS.$(this),t=n.attr("checked")?"on":"off";return r("gh.labs.feature").trigger("rapid_removal",t),!0}),AJS.$(document).delegate("#js-classic-link_lnk","click",function(e){return r("gh.agile.menu").trigger("classic"),!0});var n=function(e){return"_blank"==e.attr("target")||(setTimeout('document.location = "'+e.attr("href")+'"',100),!1)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-help-analytics', location = 'includes/js/rapid/Analytics.js' */
define("jira-agile/rapid/analytics",["jquery","underscore"],function(n,a){"use strict";function e(){n(document).on(l,function(n,a){i(a)})}function i(n){var e={},i=n.category;a.isUndefined(n.action)||(i=i+"."+n.action),a.isUndefined(n.label)||(e.label=n.label),a.isUndefined(n.value)||(e.value=n.value),AJS.trigger("analytics",{name:i,data:e})}GH.AnalyticsEnabled=!0;var l="gh.analytics.async";e();var t={};return t.isEnabled=function(){return GH.AnalyticsEnabled},t.setEnabled=function(n){return GH.AnalyticsEnabled=n},t}),AJS.namespace("GH.Analytics",null,require("jira-agile/rapid/analytics"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-globalissueoperations', location = 'includes/js/gh-globalissueoperations.js' */
define("jira-agile/gh-globalissueoperations",["jira/dialog/form-dialog","jira/dialog/dialog-register","jquery"],function(r,i,e){var a=require("jira-agile/rapid/analytics-tracker"),o=new a("gh.issueaction.gotorapidboard").setAsync(!1),n=r.extend({_performRedirect:function(r){o.trigger("direct"),window.location.href=r}}),t=new n({id:"gh-rapidboard-dialog",trigger:".issueaction-greenhopper-rapidboard-operation",ajaxOptions:function(){return i.getDefaultAjaxOptions.apply(this,arguments)}}),s=function(r,i){var a="",o=e(r);return o.hasClass("js-rapidboard-operation-sprint")?a="sprint.":o.hasClass("js-rapidboard-operation-issue")&&(a="issue."),a+=i};return{init:function(){e(document).on("simpleClick",".issueaction-greenhopper-rapidboard-operation",function(r){r.preventDefault();var i=s(r.currentTarget,"open");o.trigger(i),t.show()}),e(document).on("simpleClick",".js-select-rapidboard",function(r){var i=s(r.currentTarget,"chosen");o.trigger(i)})}}}),require("jira-agile/gh-globalissueoperations").init();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-notification', location = 'includes/js/rapid/ui/notification/Notification.soy' */
// This file was automatically generated from Notification.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace GH.tpl.rapid.notification.
 */

if (typeof GH == 'undefined') { var GH = {}; }
if (typeof GH.tpl == 'undefined') { GH.tpl = {}; }
if (typeof GH.tpl.rapid == 'undefined') { GH.tpl.rapid = {}; }
if (typeof GH.tpl.rapid.notification == 'undefined') { GH.tpl.rapid.notification = {}; }


GH.tpl.rapid.notification.renderErrorMessages = function(opt_data, opt_ignored) {
  var output = '';
  if (opt_data.errors.length > 1) {
    output += '<ul>';
    var errorList6 = opt_data.errors;
    var errorListLen6 = errorList6.length;
    for (var errorIndex6 = 0; errorIndex6 < errorListLen6; errorIndex6++) {
      var errorData6 = errorList6[errorIndex6];
      output += '<li>' + ((errorData6.noAutoescape) ? soy.$$filterNoAutoescape(errorData6.message) : soy.$$escapeHtml(errorData6.message)) + '</li>';
    }
    output += '</ul>';
  } else {
    var errorList18 = opt_data.errors;
    var errorListLen18 = errorList18.length;
    for (var errorIndex18 = 0; errorIndex18 < errorListLen18; errorIndex18++) {
      var errorData18 = errorList18[errorIndex18];
      output += '<p>' + ((errorData18.noAutoescape) ? soy.$$filterNoAutoescape(errorData18.message) : soy.$$escapeHtml(errorData18.message)) + '</p>';
    }
  }
  return output;
};
if (goog.DEBUG) {
  GH.tpl.rapid.notification.renderErrorMessages.soyTemplateName = 'GH.tpl.rapid.notification.renderErrorMessages';
}


GH.tpl.rapid.notification.renderContextualErrors = function(opt_data, opt_ignored) {
  var output = '';
  var errorList29 = opt_data.errors;
  var errorListLen29 = errorList29.length;
  for (var errorIndex29 = 0; errorIndex29 < errorListLen29; errorIndex29++) {
    var errorData29 = errorList29[errorIndex29];
    output += '<div class="ghx-error">' + soy.$$escapeHtml(errorData29.message) + '</div>';
  }
  return output;
};
if (goog.DEBUG) {
  GH.tpl.rapid.notification.renderContextualErrors.soyTemplateName = 'GH.tpl.rapid.notification.renderContextualErrors';
}


GH.tpl.rapid.notification.renderMessageHolder = function(opt_data, opt_ignored) {
  return '<div class="ghx-body-warning">' + soy.$$escapeHtml(AJS.format('\u6b64\u7248\u672c\u7684{0}\u4e0d\u662f\u7cfb\u7edf\u652f\u6301\u7684\u6d4f\u89c8\u5668\u3002',opt_data.browser)) + ' <a href="' + soy.$$escapeHtml(opt_data.docsUrl) + '">' + soy.$$escapeHtml('\u67e5\u770b\u6587\u6863\u4e86\u89e3\u652f\u6301\u7684\u6d4f\u89c8\u5668\u3002') + '</a></div>';
};
if (goog.DEBUG) {
  GH.tpl.rapid.notification.renderMessageHolder.soyTemplateName = 'GH.tpl.rapid.notification.renderMessageHolder';
}


GH.tpl.rapid.notification.renderAuiMessage = function(opt_data, opt_ignored) {
  return '<div class="aui-message ' + soy.$$escapeHtml(opt_data.type) + ((opt_data.className) ? ' aui-message-' + soy.$$escapeHtml(opt_data.className) : '') + '">' + ((opt_data.title) ? '<p class="title">' + ((opt_data.icon) ? '<span class="aui-icon icon-' + soy.$$escapeHtml(opt_data.type) + '"></span>' : '') + '<strong>' + soy.$$escapeHtml(opt_data.title) + '</strong></p>' : '') + ((opt_data.message) ? '<p>' + ((! opt_data.title && opt_data.icon) ? '<span class="aui-icon icon-' + soy.$$escapeHtml(opt_data.type) + '"></span>' : '') + ((opt_data.messageHtml) ? soy.$$filterNoAutoescape(opt_data.message) : soy.$$escapeHtml(opt_data.message)) + '</p>' : '') + '</div>';
};
if (goog.DEBUG) {
  GH.tpl.rapid.notification.renderAuiMessage.soyTemplateName = 'GH.tpl.rapid.notification.renderAuiMessage';
}


GH.tpl.rapid.notification.renderGHtvMessage = function(opt_data, opt_ignored) {
  return '<div id="ghx-update-message" class="ghx-tv-message"><p>' + soy.$$escapeHtml('\u6b64\u677f\u5df2\u7ecf\u66f4\u65b0') + ': <a href="" class="js-refresh-now" data-track-click="gh.rapidboard.updated.message.refresh.click">' + soy.$$escapeHtml('\u5237\u65b0') + '</a><span class="ghx-divider">&bull;</span><a href="" class="js-ignore-refresh" data-track-click="gh.rapidboard.updated.message.cancel.click">' + soy.$$escapeHtml('\u5ffd\u7565') + '</a></p></div>';
};
if (goog.DEBUG) {
  GH.tpl.rapid.notification.renderGHtvMessage.soyTemplateName = 'GH.tpl.rapid.notification.renderGHtvMessage';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-notification', location = 'includes/js/rapid/ui/notification/Notification.js' */
define("jira-agile/rapid/ui/notification",["require"],function(e){"use strict";function a(e,a,r){return s(i({body:e,title:r.showTitle?a:void 0,close:r.autoHide?"auto":r.closeable?"manual":"never",type:r.type}))}function s(e){var a=r(e).find(".title");return""===r.trim(a.text())&&a.remove(),e}var r=e("jquery"),t=e("underscore"),i=e("aui/flag"),n=e("jira/util/logger"),o=e("jira/analytics"),d={showErrors:function(e){var a=!(arguments.length>1&&void 0!==arguments[1])||arguments[1],s=arguments[2],r="\u9519\u8bef",t=GH.tpl.rapid.notification.renderErrorMessages(e);d.showError(r,t,a,s)},showError:function(e,s){var r=!(arguments.length>2&&void 0!==arguments[2])||arguments[2],i=arguments[3],n={autoHide:!1,closeable:r,showTitle:!0,type:"error"};i=t.extend({},n,i),a(s,e,i)},showWarnings:function(e,s){var r={preRendered:!1,autoHide:!0,closeable:!0,showTitle:!1,type:"warning"};s=t.extend({},r,s);var i=void 0;i=s.preRendered?e:GH.tpl.rapid.notification.renderErrorMessages({errors:e}),a(i,"\u8b66\u544a",s)},showWarning:function(e){d.showWarnings(e,{preRendered:!0})},showSuccess:function(e,s){var r={closeable:!0,autoHide:!0,showTitle:!1,type:"success"};s=t.extend({},r,s),a(e,null,s)},clear:function(){r("#aui-flag-container").empty()},handleDateFormatMismatchError:function(e){var a='<a href="'+GH.Ajax.CONTEXT_PATH+'/secure/admin/AdvancedApplicationProperties.jspa">';a+="\u9ad8\u7ea7\u8bbe\u7f6e",a+="</a>";var s=AJS.format("\u60a8\u7684 Jira \u65e5\u671f\u683c\u5f0f\u8bbe\u7f6e\u4e0d\u6b63\u786e\u3002\u8bf7\u68c0\u67e5 {0} \u5e76\u786e\u4fdd \'\'jira.date.time.picker.java.format\'\' \u53ca \'\'jira.date.time.picker.javascript.format\'\' \u5b58\u5728\u5339\u914d\u8f93\u51fa\u3002",a);d.showWarnings(s,{autoHide:!1,preRendered:!0})},auiMessage:function(){n.warn("[deprecated] GH.Notification.auiMessage has no effect")},showIssueMessage:function(e){var a=JIRA.SessionStorage,s=null,r=null;e||(e=a.getItem("selectedIssueId")),e&&(r=a.getItem("selectedIssueKey"),s=a.getItem("selectedIssueMsg"),d.showIssueMessageImpl(e,s,r)),a.removeItem("selectedIssueId"),a.removeItem("selectedIssueKey"),a.removeItem("selectedIssueMsg")},showIssueMessageImpl:function(e,a,s){a||(a="thanks_issue_updated");var r=AJS.params[a];r&&s&&(r=AJS.format(r,s),d.showSuccess(r))},addPageLoadMessage:function(e,a){var s=GH.storage.get("gh.pageloadmessages",!0)||[];s.push({message:e,type:a||"success"}),GH.storage.put("gh.pageloadmessages",s,!0)},_displayPageLoadMessages:function(){var e=GH.storage.get("gh.pageloadmessages",!0);if(e){var a=e[0].type,s="";switch(t.each(e,function(e){s.length>0&&(s+="<br>"),s+=e.message}),a){case"warning":d.showWarning(s);break;case"success":default:d.showSuccess(s)}GH.storage.put("gh.pageloadmessages",null,!0)}},showBoardUpdatedMessage:function(){d.triggerShowMessageAnalytics(),d.clearBoardUpdatedMessage();var e=r.Deferred(),a=r(GH.tpl.rapid.notification.renderGHtvMessage()).appendTo("#gh");return a.data("deferred",e),a.find(".js-refresh-now").click(function(s){s.preventDefault(),a.removeData("deferred"),d.clearBoardUpdatedMessage(),e.resolve()}),a.find(".js-ignore-refresh").click(function(e){e.preventDefault(),d.clearBoardUpdatedMessage()}),e.promise()},clearBoardUpdatedMessage:function(){var e=r("#ghx-update-message");if(e.length){var a=e.data("deferred");a&&a.reject&&a.reject(),e.remove()}},isBoardUpdatedMessageVisible:function(){var e=r("#ghx-update-message");return e.length>0},triggerShowMessageAnalytics:function(){o.send({name:"gh.rapidboard.updated.message.show",properties:{initial:!d.isBoardUpdatedMessageVisible()}})}};return d}),AJS.namespace("GH.Notification",null,require("jira-agile/rapid/ui/notification")),AJS.$(document).ready(function(){GH.Notification.showIssueMessage(),GH.storage&&GH.Notification._displayPageLoadMessages()});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/gh-namespace.js' */
AJS.namespace("GH.namespace",null,AJS.namespace);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/initFieldPickers.js' */
!function(e){function n(n){e(".js-epic-picker",n).each(function(){var n=e(this),i=new JIRA.EpicPicker({element:n});e(document).trigger(JIRA.EpicPicker.READY_EVENT,i)})}function i(n){e(".js-sprint-picker",n).each(function(){var n=e(this),i=new JIRA.SprintPicker({element:n});e(document).trigger(JIRA.SprintPicker.READY_EVENT,i)})}JIRA.bind(JIRA.Events.NEW_CONTENT_ADDED,function(e,c,t){t!==JIRA.CONTENT_ADDED_REASON.panelRefreshed&&(n(c),i(c))})}(AJS.$);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/EpicPicker.js' */
JIRA.EpicPicker=AJS.SingleSelect.extend({init:function(e){function r(e){var r=!p();return{searchQuery:e,projectKey:AJS.$(g).attr("data-project-key"),maxResults:100,hideDone:r}}function t(e,r,t,i){e=AJS.escapeHTML(String(e)),r=AJS.escapeHTML(String(r)),i=AJS.escapeHTML(String(i));var n=i.toUpperCase(),o=e.toUpperCase(),a=r.toUpperCase();n=(n+"").replace(/([\\\.\+\*\?\[\^\]\$\(\)\{\}\=\!\<\>\|\:])/g,"\\$1");var c,s=new RegExp("("+n+")","gi"),l=o.search(s);l>-1&&(c=e.slice(l,l+i.length),e=e.replace(c,"<em>"+c+"</em>")),l=a.search(s),l>-1&&(c=r.slice(l,l+i.length),r=r.replace(c,"<em>"+c+"</em>"));var p=t?" &#8226; "+"\u5b8c\u6210":"";return e+" - <span class='epic-menu-metadata'>("+r+p+")</span>"}function i(e){if(!e||!e.epicLists||0===e.total)return[];for(var r=[],t=d.getQueryVal(),i=0,a=0;a<e.epicLists.length;a++){var c=e.epicLists[a],s=c.epicNames,l=o(a+1,c.listDescriptor);s.length>0&&(_.each(s,function(e){l.addItem(n(e,t))}),r.push(l)),i+=s.length}if(i>0){var h=new AJS.GroupDescriptor({weight:0,showLabel:!1,items:[new AJS.ItemDescriptor({label:t,highlighted:!0,styleClass:"ghx-epic-menu-header",customItem:!0,html:"<li><h5>"+AJS.format("\u663e\u793a{1}\u4e2d\u7684{0}\u4e2a\u5339\u914d\u53f2\u8bd7",i,e.total)+"</h5><label for='chkShowDoneEpic'><input type='checkbox' id='chkShowDoneEpic'"+(p()?" checked":"")+">"+"\u663e\u793a\u5df2\u5b8c\u6210\u7684\u53f2\u8bd7"+"</label></li>"})]});r.unshift(h)}return r}function n(e,r){var i=e.key,n=e.name,o=e.isDone,a=t(n,i,o,r);return new AJS.ItemDescriptor({value:"key:"+i,fieldText:n,label:a,html:a,allowDuplicate:!1,highlighted:!0})}function o(e,r){return new AJS.GroupDescriptor({weight:e,label:r,replace:!0})}function a(e,r,t){var i=AJS.$(g);i.siblings(".ghx-error").remove(),i.before(GH.tpl.rapid.notification.renderAuiMessage({message:"\u9519\u8bef"+": "+t,type:"error",className:"ghx-error aui-ss"}))}function c(e){var r="",t=!1;return e&&e.length>0&&(r=AJS.format("\u5173\u952e\u5b57\u4e3a\u201c{0}\u201d\u7684Epic\u4e0d\u5b58\u5728\u3002",e),t=!0),{isError:t,message:r}}function s(e){var r,t;u.apply(d,[e]),e=e||AJS.$.trim(d.$field.val()),t=d.model.getDescriptor(e),!t&&d.$container.hasClass("aui-ss-editing")&&(r=c(e),r.isError?(d.options.errorMessage=r.message,d.showErrorMessage(e)):d.hideErrorMessage())}function l(){d._super(e),u=d.handleFreeInput,d.handleFreeInput=s;var r=require("jira/ajs/layer/inline-layer");d.dropdownController.bind(r.EVENTS.show,function(e,r){r.off("click",'.ghx-epic-menu-header label[for="chkShowDoneEpic"]',h),r.on("click",'.ghx-epic-menu-header label[for="chkShowDoneEpic"]',h)})}function p(){return"true"===localStorage.getItem("gh.epicpicker.showdone")}function h(e){e.preventDefault();var r=!p();try{localStorage.setItem("gh.epicpicker.showdone",r)}catch(e){}return d.$field.click(),!1}var u,g=e.element,d=this,m=require("wrm/context-path");AJS.$.extend(e,{submitInputVal:!0,showDropdownButton:!0,removeDuplicates:!0,ajaxOptions:{url:m()+"/rest/greenhopper/1.0/epics",query:!0,data:r,formatResponse:i,error:a}});var f=this.setSelection;this.setSelection=function(e,r){e.properties.customItem||f.apply(this,arguments)},l()}}),JIRA.EpicPicker.READY_EVENT="gh.epic-picker.ready";
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/SprintPicker.js' */
JIRA.SprintPicker=AJS.SingleSelect.extend({init:function(e){function n(e){var n=f.apply(this,arguments);return n.find("a").append(GH.tpl.customfields.sprintpicker.renderSuggestionMeta(e.properties)),n}function t(e){var n,t=this.getSelectedDescriptor();e||t||p(null),o(),this.hideErrorMessage(),n=t?i(t):r(e),n.message&&(this.options.errorMessage=n.message,h.apply(this,arguments),this.$errorMessage.addClass("inline-edit-error")),n.needsScopeChangeWarning&&s()}function r(e){var n=null,t=!1;return e&&e.length>0?v?(n=AJS.format("{0} \u4e0d\u662f\u6709\u6548\u7684Sprint\u540d\u79f0\u4e0d\u5b8c\u6574\u7684\u95ee\u9898\u5c06\u4f1a\u79fb\u5230\u5f85\u529e\u4e8b\u9879\u3002",e),v===m.ACTIVE&&(t=!0)):n=AJS.format("{0} \u4e0d\u662f\u6709\u6548\u7684Sprint\u540d\u79f0",e):v===m.ACTIVE&&(t=!0),{message:n,needsScopeChangeWarning:t}}function i(e){var n=!1;return v===m.ACTIVE?n=!!u()||I!=e.properties.value:v===m.FUTURE?e.properties.stateKey===m.ACTIVE&&(n=!0):v||e.properties.stateKey===m.ACTIVE&&(n=!0),{message:null,needsScopeChangeWarning:n}}function s(){if(!S){u()||(S=AJS.$(GH.tpl.rapid.notification.renderAuiMessage({type:"warning",className:"ghx-sprint-picker-scope-warning",icon:!0,title:"\u8fd9\u4e2a\u64cd\u4f5c\u5c06\u4f1a\u5f71\u54cd\u8fdb\u884c\u4e2dSprint\u7684\u8303\u56f4\uff01"})));var e=J.$container.closest(".inline-edit-fields");e.size()>0?u()?(S=AJS.$('<div class="ghx-estimate-scope-warning"></div>'),S.text("\u8fd9\u4e2a\u64cd\u4f5c\u5c06\u4f1a\u5f71\u54cd\u8fdb\u884c\u4e2dSprint\u7684\u8303\u56f4\uff01"),e.find(".field-group").append(S)):e.closest(".editable-field").append(S):(u()&&(S=AJS.$(GH.tpl.rapid.notification.renderAuiMessage({type:"warning",className:"ghx-sprint-picker-scope-warning",icon:!0,title:"\u8fd9\u4e2a\u64cd\u4f5c\u5c06\u4f1a\u5f71\u54cd\u8fdb\u884c\u4e2dSprint\u7684\u8303\u56f4\uff01"}))),J.$container.closest(".field-group").append(S))}S.show()}function o(){S&&(S.remove(),S=null)}function a(e){if(!e.suggestions&&!e.allMatches)return[];var n=new AJS.GroupDescriptor({weight:0,label:"\u5efa\u8bae"}),t=new AJS.GroupDescriptor({weight:1,label:"\u5168\u90e8 Sprint"});return _.each(l(e.suggestions),function(e){n.addItem(g(e,!0))}),_.each(l(e.allMatches),function(e){t.addItem(g(e,!1))}),[n,t]}function g(e,n){var t=AJS.escapeHTML(String(e.name));return new AJS.ItemDescriptor({value:e.id.toString(),label:e.name,html:t,stateKey:e.stateKey,boardName:e.boardName,allowDuplicate:!1,isSuggestion:n})}function l(e){return _.sortBy(e,function(e){return e.name.toLowerCase()})}function p(e){GH.SprintConfig&&AJS.$(GH).trigger("QuickEdit.fieldChange",{fieldId:GH.SprintConfig.getSprintFieldId(),fieldChangeData:{original:I,updated:e?e.properties.value:null}})}function d(){if(J._super(e),f=J.listController._renders.suggestion,h=J.showErrorMessage,J.listController._renders.suggestion=n,J.showErrorMessage=t,u()){var r=J.$container.closest(".inline-edit-fields"),i=r.size()>0;if(i){var s=w.parent().find("input");s.on("focus",function(){J.showErrorMessage()})}}w.bind("selected",function(e,n,t){u()&&i||t.showErrorMessage(t.getQueryVal()),p(n),J.showErrorMessage()}),J.suggestionsHandler=new x(J.options,J.model)}function c(e,n,t){var r=AJS.$(w);r.siblings(".ghx-error").remove(),r.before(GH.tpl.rapid.notification.renderAuiMessage({message:"\u9519\u8bef"+": "+t,type:"error",className:"ghx-error aui-ss"}))}function u(){return!(!GH||!GH.DetailsObjectFactory)}var f,h,S,m={ACTIVE:"ACTIVE",FUTURE:"FUTURE",CLOSED:"CLOSED"},A=10,w=e.element,v=w.data("saved-state"),I=w.data("saved-id"),J=this,x=AJS.SelectSuggestHandler.extend({formatSuggestions:function(e,n){return e=this._super(e,n),0===n.length&&e[e.length-1].footerText("\u5f00\u59cb\u8f93\u5165Sprint\u4fe1\u606f\u3002"),e}}),E=require("wrm/context-path");AJS.$.extend(e,{submitInputVal:!1,showDropdownButton:!0,removeOnUnSelect:!1,maxInlineResultsDisplayed:A,ajaxOptions:{url:E()+"/rest/greenhopper/1.0/sprint/picker",query:!0,formatResponse:a,error:c}}),d()}}),JIRA.SprintPicker.READY_EVENT="gh.sprint-picker.ready";
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/SprintPicker.soy' */
// This file was automatically generated from SprintPicker.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace GH.tpl.customfields.sprintpicker.
 */

if (typeof GH == 'undefined') { var GH = {}; }
if (typeof GH.tpl == 'undefined') { GH.tpl = {}; }
if (typeof GH.tpl.customfields == 'undefined') { GH.tpl.customfields = {}; }
if (typeof GH.tpl.customfields.sprintpicker == 'undefined') { GH.tpl.customfields.sprintpicker = {}; }


GH.tpl.customfields.sprintpicker.renderSuggestionMeta = function(opt_data, opt_ignored) {
  return '' + ((opt_data.boardName && ! opt_data.isSuggestion) ? '<span class=\'ghx-sprint-picker-meta\'> (' + GH.tpl.customfields.sprintpicker.renderMetaWithBoardName(opt_data) + ')</span>' : '<span class=\'ghx-sprint-picker-meta\'> (' + GH.tpl.customfields.sprintpicker.renderMetaWithoutBoardName(opt_data) + ')</span>');
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintpicker.renderSuggestionMeta.soyTemplateName = 'GH.tpl.customfields.sprintpicker.renderSuggestionMeta';
}


GH.tpl.customfields.sprintpicker.renderMetaWithoutBoardName = function(opt_data, opt_ignored) {
  return '' + ((opt_data.stateKey == 'ACTIVE') ? soy.$$escapeHtml('\u5728\u7528 sprint') : (opt_data.stateKey == 'FUTURE') ? soy.$$escapeHtml('\u672a\u6765 Sprint') : '');
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintpicker.renderMetaWithoutBoardName.soyTemplateName = 'GH.tpl.customfields.sprintpicker.renderMetaWithoutBoardName';
}


GH.tpl.customfields.sprintpicker.renderMetaWithBoardName = function(opt_data, opt_ignored) {
  return '' + ((opt_data.stateKey == 'ACTIVE') ? soy.$$escapeHtml(AJS.format('\u5728\u7528 Sprint \u5728 {0}',opt_data.boardName)) : (opt_data.stateKey == 'FUTURE') ? soy.$$escapeHtml(AJS.format('\u672a\u6765 sprint \u5728 {0}',opt_data.boardName)) : '');
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintpicker.renderMetaWithBoardName.soyTemplateName = 'GH.tpl.customfields.sprintpicker.renderMetaWithBoardName';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/SprintMultiselect.js' */
JIRA.bind(JIRA.Events.NEW_CONTENT_ADDED,function(e,t){function n(e){var t=new AJS.GroupDescriptor({weight:0,label:"\u5efa\u8bae",items:_.map(e.suggestions,s)}),n=new AJS.GroupDescriptor({weight:1,label:"\u5168\u90e8 Sprint",items:_.map(e.allMatches,s)});return[t,n]}function s(e){e.date=moment(e.date,"YYYY-MM-DDTHH:mm:ssZ").format("LL");var t=GH.tpl.customfields.sprintmultiselect.renderSuggestionMeta(e);return new AJS.ItemDescriptor({value:e.id,label:e.name,html:t,title:GH.tpl.customfields.sprintmultiselect.renderTooltip(e),date:e.date,boardName:e.boardName,stateKey:e.stateKey,highlighted:!0,allowDuplicate:!1})}AJS.$(".js-sprint-checkboxmultiselect",t).each(function(){var e=new AJS.CheckboxMultiSelect({element:this,maxInlineResultsDisplayed:5,content:"ajax",removeOnUnSelect:!0,ajaxOptions:{url:AJS.contextPath()+"/rest/greenhopper/1.0/sprint/picker",error:function(){e._setSuggestions([])},data:{excludeCompleted:!1},query:!0,formatResponse:n},suggestionsHandler:GH.SprintSuggestHandler})})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/SprintMultiselect.soy' */
// This file was automatically generated from SprintMultiselect.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace GH.tpl.customfields.sprintmultiselect.
 */

if (typeof GH == 'undefined') { var GH = {}; }
if (typeof GH.tpl == 'undefined') { GH.tpl = {}; }
if (typeof GH.tpl.customfields == 'undefined') { GH.tpl.customfields = {}; }
if (typeof GH.tpl.customfields.sprintmultiselect == 'undefined') { GH.tpl.customfields.sprintmultiselect = {}; }


GH.tpl.customfields.sprintmultiselect.renderSuggestionMeta = function(opt_data, opt_ignored) {
  return '<div>' + soy.$$escapeHtml(opt_data.name) + '</div><div><span class=\'ghx-sprint-multiselect-meta\'> ' + ((opt_data.boardName) ? GH.tpl.customfields.sprintmultiselect.renderMetaWithBoardName(opt_data) : GH.tpl.customfields.sprintmultiselect.renderMetaWithoutBoardName(opt_data)) + '</span></div>';
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintmultiselect.renderSuggestionMeta.soyTemplateName = 'GH.tpl.customfields.sprintmultiselect.renderSuggestionMeta';
}


GH.tpl.customfields.sprintmultiselect.renderMetaWithoutBoardName = function(opt_data, opt_ignored) {
  return '' + ((opt_data.stateKey == 'ACTIVE') ? soy.$$escapeHtml('\u6d3b\u8dc3') : (opt_data.stateKey == 'FUTURE') ? soy.$$escapeHtml('\u672a\u6765') : (opt_data.stateKey == 'CLOSED') ? soy.$$escapeHtml('\u5df2\u5173\u95ed') : '');
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintmultiselect.renderMetaWithoutBoardName.soyTemplateName = 'GH.tpl.customfields.sprintmultiselect.renderMetaWithoutBoardName';
}


GH.tpl.customfields.sprintmultiselect.renderMetaWithBoardName = function(opt_data, opt_ignored) {
  return '' + ((opt_data.stateKey == 'ACTIVE') ? soy.$$escapeHtml(AJS.format('\u6d3b\u52a8\u7684 {0}',opt_data.boardName)) : (opt_data.stateKey == 'FUTURE') ? soy.$$escapeHtml(AJS.format('\u672a\u6765{0}',opt_data.boardName)) : (opt_data.stateKey == 'CLOSED') ? soy.$$escapeHtml(AJS.format('\u5df2\u5173\u95ed{0}',opt_data.boardName)) : '');
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintmultiselect.renderMetaWithBoardName.soyTemplateName = 'GH.tpl.customfields.sprintmultiselect.renderMetaWithBoardName';
}


GH.tpl.customfields.sprintmultiselect.renderTooltip = function(opt_data, opt_ignored) {
  return '' + ((opt_data.stateKey == 'ACTIVE') ? soy.$$escapeHtml(AJS.format('\u5f00\u59cb\u4e8e{0}',opt_data.date)) : (opt_data.stateKey == 'FUTURE') ? soy.$$escapeHtml('\u672a\u6765') : (opt_data.stateKey == 'CLOSED') ? soy.$$escapeHtml(AJS.format('\u5173\u95ed\u4e8e{0}',opt_data.date)) : '');
};
if (goog.DEBUG) {
  GH.tpl.customfields.sprintmultiselect.renderTooltip.soyTemplateName = 'GH.tpl.customfields.sprintmultiselect.renderTooltip';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/SprintSuggestHandler.js' */
define("includes/js/field/sprint-suggest-handler",["jira/ajs/select/suggestions/checkbox-multi-select-suggest-handler"],function(e){var t=function(e){_.each(e,function(e){var t=JSON.parse(e.properties.meta);t.date=moment(t.date).format("LL"),e.properties.title=GH.tpl.customfields.sprintmultiselect.renderTooltip(t),e.properties.html=GH.tpl.customfields.sprintmultiselect.renderSuggestionMeta(t)})};return e.extend({formatSuggestions:function(e,r){var s=this.model.getDisplayableSelectedDescriptors();t(s);var i=function(e){var t;for(t in s)if(s[t].properties.value===e.properties.value.toString())return!1;return!0};return e=_.map(e,function(e){var t=_.filter(e.properties.items,i);return e.properties.items=t,e}),this._super(e,r)}})}),AJS.namespace("GH.SprintSuggestHandler",null,require("includes/js/field/sprint-suggest-handler"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/EpicLinkMultiselect.js' */
require(["require"],function(e){function t(e){return 0!==e.indexOf("project")?"":n(e).trim()}function n(e){function t(e){var t=e.match(/\bproject\s*=\s*(\S+)/i);return t?t[1]:null}function n(e){var t=e.match(/\bproject\s+in\s+\((.*?)\)/i);return t?t[1]:null}return t(e)||n(e)}var i=e("jquery"),r=e("underscore");JIRA.bind(JIRA.Events.NEW_CONTENT_ADDED,function(e,n){var s=i("#jql",this).val();n.find(".js-epiclink-checkboxmultiselect").each(function(){var e=new AJS.CheckboxMultiSelect({element:this,matchingStrategy:"(^|.*?(.*))({0})(.*)",maxInlineResultsDisplayed:5,content:"ajax",removeOnUnSelect:!0,ajaxOptions:{url:AJS.contextPath()+"/rest/greenhopper/1.0/epics",error:function(){e._setSuggestions([])},data:{searchQuery:"",projectKey:t(s),filterEpicsByGivenProjects:!0,maxResults:100},query:!0,minQueryLength:0,formatResponse:function(e){var t=[];if(t.push(new AJS.GroupDescriptor({weight:1,items:[new AJS.ItemDescriptor({highlighted:!0,label:"\u6ca1\u6709 Epic \u94fe\u63a5",value:"EMPTY",title:"EMPTY"})]})),!e||!e.epicLists||0===e.total)return t;for(var n=0;n<e.epicLists.length;n++){var i=e.epicLists[n],s=i.epicNames,c=new AJS.GroupDescriptor({label:i.listDescriptor,weight:n});s.length>0&&(r.each(s,function(e){c.addItem(new AJS.ItemDescriptor({highlighted:!0,label:e.name+" - ("+e.key+")",value:e.key,title:e.key}))}),t.push(c))}return t}},suggestionsHandler:GH.EpicLinkSuggestHandler})})})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:gh-custom-field-pickers', location = 'includes/js/field/EpicLinkSuggestHandler.js' */
define("includes/js/field/epic-link-suggest-handler",["jira/ajs/select/suggestions/checkbox-multi-select-suggest-handler"],function(e){return e.extend({formatSuggestions:function(e,r){var i=this.model.getDisplayableSelectedDescriptors(),t=function(e){var r;for(r in i)if(i[r].properties.value===e.properties.value.toString())return!1;return!0};return e=_.map(e,function(e){var r=_.filter(e.properties.items,t);return e.properties.items=r,e}),this._super(e,r)}})}),AJS.namespace("GH.EpicLinkSuggestHandler",null,require("includes/js/field/epic-link-suggest-handler"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.pyxis.greenhopper.jira:project-template-resources', location = '/projecttemplates/soy/ProjectTemplates.soy' */
// This file was automatically generated from ProjectTemplates.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace GH.tpl.projecttemplates.
 */

if (typeof GH == 'undefined') { var GH = {}; }
if (typeof GH.tpl == 'undefined') { GH.tpl = {}; }
if (typeof GH.tpl.projecttemplates == 'undefined') { GH.tpl.projecttemplates = {}; }


GH.tpl.projecttemplates.kanbanInfoPageJira7 = function(opt_data, opt_ignored) {
  return '' + GH.tpl.projecttemplates.templateInfoDialog({description: '\u4f7f\u7528\u6b64\u9879\u76ee\u53ef\u4ee5\u4f18\u5316\u7684\u5de5\u4f5c\u6d41\u7a0b\u5728\u60a8\u7684\u5f00\u53d1\u9879\u76ee\u3002\u5c06\u7ea6\u675f\u6dfb\u52a0\u5230\u5de5\u4f5c\u4e2d--\u8fdb\u5ea6\u3001\u5206\u6790\u9700\u8981\u591a\u957f\u65f6\u95f4\u624d\u80fd\u5b8c\u6210\u7684\u95ee\u9898\u3001\u627e\u5230\u7684\u74f6\u9888\u60a8\u7684\u8fdb\u7a0b, \u548c\u66f4\u591a\u3002\u8fd9\u4e00\u9879\u76ee\u5305\u62ec\u5728\u770b\u677f\u677f\u3001\u57fa\u672c\u7684\u7075\u6d3b\u7684\u5de5\u4f5c\u6d41\u7a0b\u548c\u95ee\u9898\u7c7b\u578b\u7684\u914d\u7f6e, \u60a8\u53ef\u4ee5\u5728\u4ee5\u540e\u66f4\u6539\u3002', extraIssueTypes: '<li><span class="issuetype-list-label"><span class="issuetype-icon story"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u6545\u4e8b') + '</span></span></li><li><span class="issuetype-list-label"><span class="issuetype-icon epic"></span><span class="issuetype-name">' + soy.$$escapeHtml('Epic') + '</span></span></li>', workflowScreenshotClass: 'kanban-workflow-screenshot'});
};
if (goog.DEBUG) {
  GH.tpl.projecttemplates.kanbanInfoPageJira7.soyTemplateName = 'GH.tpl.projecttemplates.kanbanInfoPageJira7';
}


GH.tpl.projecttemplates.scrumInfoPageJira7 = function(opt_data, opt_ignored) {
  return '' + GH.tpl.projecttemplates.templateInfoDialog({description: '\u4f7f\u7528\u6b64\u9879\u76ee\u53ef\u7ba1\u7406\u60a8\u7684\u654f\u6377\u5f00\u53d1\u7684\u5de5\u4f5c\u3002\u521b\u5efa\u79ef\u538b\u3001\u7b79\u529e\u5de5\u4f5c\u8fdb\u5165\u51b2\u523a\u3001\u68c0\u67e5\u4f7f\u7528\u7684\u8fdb\u5c55\u60c5\u51b5\u7684\u62a5\u544a, \u4ee5\u53ca\u66f4\u591a\u3002\u8fd9\u4e00\u9879\u76ee\u5305\u62ec Scrum \u677f\u3001\u57fa\u672c\u7684\u7075\u6d3b\u5de5\u4f5c\u6d41\u7a0b\u548c\u95ee\u9898\u7c7b\u578b\u7684\u914d\u7f6e, \u60a8\u53ef\u4ee5\u5728\u4ee5\u540e\u66f4\u6539\u3002', extraIssueTypes: '<li><span class="issuetype-list-label"><span class="issuetype-icon story"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u6545\u4e8b') + '</span></span></li><li><span class="issuetype-list-label"><span class="issuetype-icon epic"></span><span class="issuetype-name">' + soy.$$escapeHtml('Epic') + '</span></span></li>', workflowScreenshotClass: 'scrum-workflow-screenshot'});
};
if (goog.DEBUG) {
  GH.tpl.projecttemplates.scrumInfoPageJira7.soyTemplateName = 'GH.tpl.projecttemplates.scrumInfoPageJira7';
}


GH.tpl.projecttemplates.softwareDevelopmentInfoPageJira7 = function(opt_data, opt_ignored) {
  return '' + GH.tpl.projecttemplates.templateInfoDialog({description: '\u4f7f\u7528\u6b64\u9879\u76ee\u5de5\u4f5c\u7684\u65b0\u529f\u80fd\u4e3a\u60a8\u7684\u4ea7\u54c1\u4ee5\u53ca\u8ddf\u8e2a\u7684\u4efb\u4f55 bug\u3002\u672c\u9879\u76ee\u4e3a\u60a8\u63d0\u4f9b\u4e86\u57fa\u672c\u7684\u5de5\u4f5c\u6d41\u7a0b\u548c\u95ee\u9898\u7c7b\u578b\u7684\u914d\u7f6e, \u60a8\u53ef\u4ee5\u5728\u4ee5\u540e\u66f4\u6539\u3002', extraIssueTypes: '<li><span class="issuetype-list-label"><span class="issuetype-icon improvement"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u6539\u8fdb') + '</span></span></li><li><span class="issuetype-list-label"><span class="issuetype-icon newfeature"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u65b0\u529f\u80fd') + '</span></span></li><li><span class="issuetype-list-label"><span class="issuetype-icon epic"></span><span class="issuetype-name">' + soy.$$escapeHtml('Epic') + '</span></span></li>', workflowScreenshotClass: 'basic-development-workflow-screenshot'});
};
if (goog.DEBUG) {
  GH.tpl.projecttemplates.softwareDevelopmentInfoPageJira7.soyTemplateName = 'GH.tpl.projecttemplates.softwareDevelopmentInfoPageJira7';
}


GH.tpl.projecttemplates.templateInfoDialog = function(opt_data, opt_ignored) {
  return '<div><div class="top-panel"><p>' + soy.$$escapeHtml(opt_data.description) + '</p></div><div class="left-panel"><h6>' + soy.$$escapeHtml('\u95ee\u9898\u7c7b\u578b') + '</h6><ul class="project-template-issuetype-list" ><li><span class="issuetype-list-label"><span class="issuetype-icon bug"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u6545\u969c') + '</span></span></li><li><span class="issuetype-list-label"><span class="issuetype-icon task"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u4efb\u52a1') + '</span></span></li><li><span class="issuetype-list-label"><span class="issuetype-icon subtask"></span><span class="issuetype-name">' + soy.$$escapeHtml('\u5b50\u4efb\u52a1') + '</span></span></li>' + soy.$$filterNoAutoescape(opt_data.extraIssueTypes) + '</ul></div><div class="right-panel"><h6>' + soy.$$escapeHtml('\u5de5\u4f5c\u6d41') + '</h6><div class="workflow ' + soy.$$escapeHtml(opt_data.workflowScreenshotClass) + '"></div></div></div>';
};
if (goog.DEBUG) {
  GH.tpl.projecttemplates.templateInfoDialog.soyTemplateName = 'GH.tpl.projecttemplates.templateInfoDialog';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugin.ext.bamboo:bamboo-soy-resources', location = 'templates/plugins/bamboo/components.soy' */
// This file was automatically generated from components.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace bamboo.components.
 */

if (typeof bamboo == 'undefined') { var bamboo = {}; }
if (typeof bamboo.components == 'undefined') { bamboo.components = {}; }


bamboo.components.pageHeader = function(opt_data, opt_ignored) {
  return '' + aui.page.pageHeader({content: '' + ((opt_data.headerImageContent) ? aui.page.pageHeaderImage({content: opt_data.headerImageContent}) : '') + aui.page.pageHeaderMain({content: opt_data.headerMainContent}) + ((opt_data.headerActionsContent) ? aui.page.pageHeaderActions({content: opt_data.headerActionsContent}) : '')});
};
if (goog.DEBUG) {
  bamboo.components.pageHeader.soyTemplateName = 'bamboo.components.pageHeader';
}


bamboo.components.icon = function(opt_data, opt_ignored) {
  return '<span class="bamboo-icon bamboo-icon-' + soy.$$escapeHtml(opt_data.type) + '"' + ((opt_data.text && opt_data.showTitle) ? ' title="' + soy.$$escapeHtml(opt_data.text) + '"' : '') + '>' + ((opt_data.text) ? '<span>' + soy.$$escapeHtml(opt_data.text) + '</span>' : '') + '</span>';
};
if (goog.DEBUG) {
  bamboo.components.icon.soyTemplateName = 'bamboo.components.icon';
}


bamboo.components.buildDetail = function(opt_data, opt_ignored) {
  return '<dl><dt class="' + soy.$$escapeHtml(opt_data.keyClass) + '">' + soy.$$escapeHtml(opt_data.key) + '</dt><dd>' + soy.$$filterNoAutoescape(opt_data.value) + '</dd></dl>';
};
if (goog.DEBUG) {
  bamboo.components.buildDetail.soyTemplateName = 'bamboo.components.buildDetail';
}


bamboo.components.artifacts = function(opt_data, opt_ignored) {
  var output = '<ul id="shared-artifacts">';
  var artifactList42 = opt_data.artifacts;
  var artifactListLen42 = artifactList42.length;
  for (var artifactIndex42 = 0; artifactIndex42 < artifactListLen42; artifactIndex42++) {
    var artifactData42 = artifactList42[artifactIndex42];
    output += bamboo.components.artifactItem(soy.$$augmentMap(artifactData42, {id: artifactData42.name, url: artifactData42.link.href}));
  }
  output += '</ul>';
  return output;
};
if (goog.DEBUG) {
  bamboo.components.artifacts.soyTemplateName = 'bamboo.components.artifacts';
}


bamboo.components.artifactItem = function(opt_data, opt_ignored) {
  return '<li>' + bamboo.components.icon({type: 'artifact-shared'}) + ' <a class="artifact-link" id="artifact-' + soy.$$escapeHtml(opt_data.id) + '" href="' + soy.$$escapeHtml(opt_data.url) + '" title="' + soy.$$escapeHtml(opt_data.name) + '">' + soy.$$escapeHtml(opt_data.name) + '</a><span class="filesize">(' + soy.$$escapeHtml(opt_data.prettySizeDescription) + ')</span></li>';
};
if (goog.DEBUG) {
  bamboo.components.artifactItem.soyTemplateName = 'bamboo.components.artifactItem';
}


bamboo.components.labels = function(opt_data, opt_ignored) {
  var output = '<ul class="labels">';
  var labelList65 = opt_data.labels;
  var labelListLen65 = labelList65.length;
  for (var labelIndex65 = 0; labelIndex65 < labelListLen65; labelIndex65++) {
    var labelData65 = labelList65[labelIndex65];
    output += bamboo.components.labelItem(soy.$$augmentMap(labelData65, {url: opt_data.baseBambooUrl + '/browse/label/' + labelData65.name}));
  }
  output += '</ul>';
  return output;
};
if (goog.DEBUG) {
  bamboo.components.labels.soyTemplateName = 'bamboo.components.labels';
}


bamboo.components.labelItem = function(opt_data, opt_ignored) {
  return '<li><a class="lozenge" href="' + soy.$$escapeHtml(opt_data.url) + '" title="' + soy.$$escapeHtml(opt_data.name) + '"><span>' + soy.$$escapeHtml(opt_data.name) + '</span></a></li>';
};
if (goog.DEBUG) {
  bamboo.components.labelItem.soyTemplateName = 'bamboo.components.labelItem';
}


bamboo.components.errorResponseMessage = function(opt_data, opt_ignored) {
  return '' + bamboo.components.auiWarning({body: '' + ((opt_data.errorMessage.messageBody) ? '<p>' + soy.$$escapeHtml(opt_data.errorMessage.messageBody) + '</p>' : (opt_data.errorMessage.oauthCallback) ? '<p><a href="' + soy.$$filterNoAutoescape(opt_data.errorMessage.oauthCallback) + '&amp;redirectUrl=' + soy.$$filterNoAutoescape(opt_data.oAuthDanceReturnUrl) + '">' + soy.$$escapeHtml('\u767b\u5f55\u548c\u5ba1\u6279') + '</a></p>' : '<p>' + soy.$$escapeHtml(opt_data.errorMessage.message) + '</p>')});
};
if (goog.DEBUG) {
  bamboo.components.errorResponseMessage.soyTemplateName = 'bamboo.components.errorResponseMessage';
}


bamboo.components.auiWarning = function(opt_data, opt_ignored) {
  return '<div class="aui-message aui-message-warning warning">' + soy.$$filterNoAutoescape(opt_data.body) + '</div>';
};
if (goog.DEBUG) {
  bamboo.components.auiWarning.soyTemplateName = 'bamboo.components.auiWarning';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.administration.atlassian-admin-quicksearch-jira:admin-quicksearch-webresources', location = 'com/atlassian/administration/quicksearch/jira/js/adminQuickNav.js' */
/**
 * Shifter group for admin search
 */
require([
    'jquery',
    'underscore',
    'jira/ajs/ajax/smart-ajax',
    'jira/shifter',
    'wrm/context-path'
], function (jQuery,
             _,
             SmartAjax,
             Shifter,
             contextPath) {
    Shifter.register(function () {
        var suggestionsDeferred = jQuery.Deferred();

        function formatItem(item) {
            return {
                label: item.label,
                value: item.linkUrl,
                keywords: item.aliases
            };
        }

        function getItemsInSection(section) {
            return _.map(section.items, formatItem).concat(_.map(section.sections, getItemsInSection));
        }

        function formatResponse(data) {
            return _.flatten(getItemsInSection(data));
        }

        SmartAjax.makeRequest({
            dataType: 'json',
            url: contextPath() + '/rest/adminquicksearch/latest/links/default'
        })
            .pipe(formatResponse)
            .done(function (suggestions) {
                suggestionsDeferred.resolve(suggestions);
            })
            .fail(function () {
                suggestionsDeferred.reject();
            });

        return {
            id: 'admin',
            name: "\u7ba1\u7406",
            weight: 500,
            getSuggestions: function () {
                return suggestionsDeferred;
            },
            onSelection: function (value) {
                window.location = value;
                return jQuery.Deferred();
            }
        };
    });
});

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.browser.metrics.browser-metrics-plugin:api', location = 'probe.js' */
!function(){var n,t,e,r,i,o,u,a,c,s,f,l,h,d,y,p,v,m,w;n=function(){return window}(),t=function(n){return!(!n.performance||!n.performance.now)}(n),e=[],r=function(n){return function(t){n.unshift({addReporter:t})}}(e),i=function(n){return function(t){for(;n.length;)t(n.splice(0,1)[0]);n.unshift=t,n.push=t}}(e),o=function(n,t){return function(e){n.push({end:{key:e.key,timestamp:t.performance.now()}})}}(e,n),u=function(n){return n.document}(n),a=function(n){return n.Promise}(n),c=function(){function n(){this._={}}var t=function(e){var r=e[0],i=e[1];i instanceof n?e.length>=3?Object.keys(i._).forEach(function(n){t([r,i._[n],n].concat(e.slice(2)))}):Object.keys(i._).forEach(function(n){t([r,i._[n],n])}):Array.isArray(i)&&r.apply(null,[i].concat(e.slice(2)))};n.prototype.forEach=function(n){t([n,this])},n.prototype.add=function(){for(var t=this,e=null,r=null,i=0;i<arguments.length;i++){if(r=arguments[i],i===arguments.length-1&&Array.isArray(t)){t.push(r);break}i<arguments.length-2&&!t._.hasOwnProperty(r)?t._[r]=new n:i!==arguments.length-2||t._.hasOwnProperty(r)||(t._[r]=[]),t=t._[r],e=r}};var e=function(n,t){if(0!==n.length){var r=n.pop(),i=r[0],o=r[1];i===t?e(n,i):o._.hasOwnProperty(t)&&delete o._[t],0===Object.keys(o).length&&e(n,i)}};return n.prototype.remove=function(){for(var n,t=!1,r=null,i=this,o=[[r,i]],u=null,a=0;a<arguments.length;a++)if(u=arguments[a],Array.isArray(i))n=i.indexOf(u),n>-1&&(i.splice(n,1),0===i.length&&o.length>1&&e(o,r),t=!0);else{if(!i._.hasOwnProperty(u))break;a===arguments.length-1&&(delete i._[u],0===Object.keys(i).length&&o.length>1&&e(o,r),t=!0),r=u,i=i._[u],o.push([r,i])}return t},n.prototype.get=function(n){return this._.hasOwnProperty(n)?this._[n]:[]},n}(),s=function(n,t,e,r){function i(n){return!n||null==n||"null"===n||"undefined"===n}function o(t,e,r){l||(c.observe(n,{attributes:!0,childList:!0,subtree:!0}),l=!0),s.add(t,e,r)}function u(t,e){var r=n.querySelectorAll(t);return r.length&&(i(e)||Array.prototype.every.call(r,function(n){return!n.querySelector(e)}))}function a(n,e){var r;n.forEach||(n=[n]),!i(e)&&Array.isArray(e)&&(e=e.join(", "));var a=new t(function(i,a){var c=[],f=[];n.forEach(function(n){var r,i;u(n,e)||(r=new t(function(t){o(n,e,t),i=function(){s.remove(n,e,t)},f.push(i)}),c.push(r))});var l=function(){f.forEach(function(n){n()})};t.all(c).then(l).then(i,a),r=function(){l(),a()}});return a.dismiss=r,a}var c,s,f=r.MutationObserver,l=!1;return f&&t?(s=new e,c=new f(function(){s.forEach(function(n,t,e){u(e,t)&&(n.forEach(function(n){n()}),s.remove(e,t))})}),a):void 0}(u,a,c,n),f=function(n){return!!n}(s),l=function(n){function t(){c(),n.body.classList.add(u)}function e(){function e(){n.body.classList.remove(u),n.removeEventListener(i,s),n.removeEventListener(o,c),r=null}if(r)return r;var c,s,f=!1;return r=new Promise(function(e,r){"visible"!==n.visibilityState?r():(s=function(){f=!0},c=function(n){n.animationName===a&&(f?r():e())},n.addEventListener(i,s),n.addEventListener(o,c),t())}),r.then(e,e),r}var r,i="visibilitychange",o="animationend",u="browser-metrics-visibility-test",a="browser-metrics-visibility-animation",c=function(){var t=n.createElement("style"),e=["."+u+" {","-webkit-animation-duration: 0.001s;","animation-duration: 0.001s;","-webkit-animation-name: "+a+";","animation-name: "+a+";","-webkit-animation-iteration-count: 1;","animation-iteration-count: 1;","}","@keyframes "+a+" {}","@-webkit-keyframes "+a+" {","from {}","to {}","}"].join("\n");t.type="text/css",t.styleSheet?t.styleSheet.cssText=e:t.appendChild(n.createTextNode(e)),n.head.appendChild(t),c=function(){}};return e}(u),h=function(n,t,e,r,i){function o(n){return Array.isArray(n)||(n=[n]),n.map(function(n){return"string"==typeof n?{selector:n,hasNone:null}:n})}function u(n){return Array.isArray(n)||"string"==typeof n}function a(n){return u(n)&&(n={rules:n}),n.rules=o(n.rules),n.requirePaint="undefined"==typeof n.requirePaint?!1:n.requirePaint,n}return function(i,o){if(n){i=a(i);var u=function(){},c=new e(function(n,r){var o=[],a=i.rules.map(function(n){var e=new t(n.selector,n.hasNone);return o.push(function(){e.dismiss()}),e});u=function(){o.forEach(function(n){n()}),r()},e.all(a).then(function(n){}).then(n,r)});return c.cancel=u,i.requirePaint&&(c=c.then(r)),"function"==typeof o&&c.then(o),c}}}(f,s,a,l,n),d=function(n,t){function e(){return r}var r=!1;return n.addEventListener("DOMContentLoaded",function(){t.setTimeout(function(){r=!0})}),e}(u,n),y=function(n,t,e,r,i,o,u){function a(){c=null}var c;return function(o){var s="isInitial"in o?o.isInitial:i()===!1,f="threshold"in o?o.threshold:1e3,l="reporters"in o?o.reporters:[];r.push({start:{key:o.key,isInitial:s,threshold:f,timestamp:s?0:u.performance.now(),reporters:Array.isArray(l)?l:[l]}}),c&&(c.cancel(),a()),o.ready&&e&&(c=n(o.ready),c.then(function(){t({key:o.key})}).then(a,a))}}(h,o,f,e,d,a,n),p=function(n){return function(t){n.push({subscribe:t})}}(e),v=function(){return window}(),m=function(n){return n.performance}(v),w=function(n,t,e,r,i,o,u){var a=function(){};return u?{start:n?i:a,end:n?r:a,addReporter:n?t:a,delegateTo:n?e:a,subscribe:n?o:a}:void 0}(t,r,i,o,y,p,m),window["browser-metrics"]=w,window.define&&window.define("internal/browser-metrics",function(){return w})}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.browser.metrics.browser-metrics-plugin:api', location = 'internal/browser-metrics-aa-beacon.js' */
!function(){var e={};e=function(e){function n(e,n){Object.keys(n).forEach(function(r){e[r]=n[r]})}Object.defineProperty(e,"__esModule",{value:!0});var r,t=[],o=[];return e.addUrlCleaner=function(e){o.push(e)},e.cleanUrl=function(e){return o.reduce(function(n,r){var t=r(e);return t.length>n.length?t:n},"")},e.addReportMarshaller=function(e){t.push(e)},e.setEventQueue=function(e){r=e},e.beacon=function(e){var o={};t.forEach(function(r){var t=r(e);"object"==typeof t&&n(o,t)});var a={name:"browser.metrics.navigation",properties:o};(r||AJS.EventQueue).push(a)},e}(e),window["browser-metrics-aa-beacon"]=e,window.define&&window.define("internal/browser-metrics-aa-beacon",function(){return e})}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.browser.metrics.browser-metrics-plugin:api', location = 'loader.js' */
!function(){var n={},r={};r=function(n,r,i){function e(){t===u&&o&&(o(),o=null)}Object.defineProperty(n,"__esModule",{value:!0});var t=0,u=0,o=null,c={install:function(n){t+=1,n(function(){u+=1,e()})}};return r["browser-metrics-plugin"]=c,i.require(["wrc!browser-metrics-plugin.contrib"],function(){r.require(["internal/browser-metrics-plugin/collector"],function(n){o=function(){n.install()},e()})}),n}(r,n=window,n.WRM)}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-browser-metrics:sensors', location = 'sensors.js' */
require(["internal/browser-metrics","jira/util/events"],function(e,i){var a,o={},d={},t={};o=function(e,i,a){function o(){function e(e){return o.hasOwnProperty(e)?o[e]:null}var o={"bonfire_top_menu_dropdown-content":"jira.header.menu.capture","browse_link-content":"jira.header.menu.projects","find_link-content":"jira.header.menu.issues","greenhopper_menu-content":"jira.header.menu.agile","home_link-content":"jira.header.menu.dashboards","plugins-jira-webitem-main-content":"jira.header.menu.portfolio","servicedesk-section-content":"jira.header.menu.service-desk","system-admin-menu-content":"jira.header.menu.admin","system-help-menu-content":"jira.header.menu.help","user-options-content":"jira.header.menu.profile"};i.bind("aui-dropdown2-show-before",function(i){var o=i.target,d=e(o&&o.id);d&&a.start({key:d,isInitial:!1,threshold:250})}),i.bind("aui-dropdown2-show-after",function(i){var o=i.target,d=e(o&&o.id);d&&a.end({key:d})})}return e.applicationMenusSensorInit=o,e}(o,i,e),d=function(e){function i(e){var i;if(a.hasOwnProperty(e))return e;for(i=0;i<o.length;i++)if(e.match(o[i].pattern))return o[i].dialogId;return null}e.safeDialogId=i;var a=["about-dialog","add-basicuser-dialog","add-developer-dialog","add-field-configuration-dialog","add-field-configuration-scheme-dialog","add-field-dialog","add-field-screen-dialog","add-field-screen-scheme-dialog","add-incoming-mail-handler-dialog","add-issue-type-dialog","add-issue-type-field-configuration-association-dialog","add-issue-type-screen-scheme-configuration-association-dialog","add-issue-type-screen-scheme-dialog","add-new-issue-type-to-scheme-dialog","add-screen-scheme-item-dialog","add-subtask-type-dialog","add-workflow-dialog","add_workflowscheme-dialog","assign-dialog","assign-issue-types-dialog","assign-to-me-link-handler","attach-file-dialog","attach-screenshot-dialog","auditing-settings-dialog","clone-issue-dialog","comment-add-dialog","configure_wallboard_dialog","confluence-page-search-dialog","copy_classic default workflow-dialog","copy-filter-dialog","copy_jira-dialog","create-issue-dialog","create-issue-dialog.issueType","create-issue-dialog.projectId","create-request-dialog","create-service-desk-dialog","create-status-dialog","create-story-dialog","create-story-dialog.issueType","create-story-dialog.projectId","create-subtask-dialog","create-subtask-dialog.issueType","create-team-dialog","create_user-dialog","credits-dialog","delete-attachment-dialog","delete-comment-dialog","delete-dshboard","delete-filter-dialog","delete-issue-dialog","delete-issue-link-dialog","delete-log-work-dialog","delete-metric","delete-queue","delete-status-dialog","deleteuser_link-dialog","devstatus-branch-detail-dialog","devstatus-build-detail-dialog","devstatus-commit-detail-dialog","devstatus-cta-dialog","devstatus-deployment-detail-dialog","devstatus-pullrequest-detail-dialog","devstatus-review-detail-dialog","discard-draft-dialog","discard_draft_workflow-dialog","edit-attachments-dialog","edit-comment","editgroups_admin-dialog","editgroups_link-dialog","editgroups_sysadmin-dialog","edit-issue-dialog","edit-labels-dialog","edit-log-work-dialog","edit-status-dialog","edituser_link_admin-dialog","edit-workflow-dialog","gh-rapidboard-dialog","inline-issue-create-dialog","invite_user-dialog","issue-actions-dialog","issue-tab-error-dialog","jim-create-project-dialog","keyboard-shortcuts-dialog","link-issue-dialog","log-work-dialog","manage-attachment-dialog","metric-pre-save-dialog","modal-field-view","permission-helper-dialog","project-avatar-cropper","project-avatar-picker","project-config-details-project-category-dialog","project-config-project-edit-dialog","project-config-project-edit-lead-and-default-assignee-dialog","project-email-dialog","publish_draft_workflow-dialog","queue-dirty-edits","QuickCreateIssue.error","QuickCreateIssue.success","QuickCreateSubtask.error","QuickCreateSubtask.success","QuickEdit.error","QuickEdit.success","remote-jira-search-dialog","rename-filter-dialog","report-dirty-edits","save-filter-dialog","sd-add-default-value","sd-add-remove-agent-dialog","sd-remove-field-dialog","server-error-dialog","report-delete-confirm","tempo-add-hours-issue-dialog","tempo-add-internal-activity-form","tempo-core-medium-form","tempo-delete-form-dialog","tempo-grace-form","tempo-large-form","tempo-medium-form","tempo-move-form-dialog","tempo-pdf-form","tempo-small-form","tempo-split-form-dialog","tempo-user-settings-dialog","user-avatar-picker","user-defaults-edit-dialog","versionsMergeDialog","view-workflow-dialog","view-workflow-dialog-project-admin","view-workflow-dialog-workflow-schemes","wait-migrate-dialog","whereismycf-dialog","workflow-text-view"].reduce(function(e,i){return e[i]=!0,e},{}),o=[{dialogId:"component-delete-dialog",pattern:/^component-\d+-delete-dialog$/i},{dialogId:"version-delete-dialog",pattern:/^version-\d+-delete-dialog$/i},{dialogId:"workflow-transition-dialog",pattern:/^workflow-transition-\d+-dialog$/i}];return e}(d),t=function(e,i,a,o){function d(){i.bind("beforeShow",function(e,i){if("string"==typeof i){var d=(0,o.safeDialogId)(i);d&&a.start({key:"jira.dialog.open."+d,isInitial:!1,threshold:1e3})}}),i.bind("dialogContentReady",function(e,i){if("string"==typeof i.options.id){var d=(0,o.safeDialogId)(i.options.id);d&&a.end({key:"jira.dialog.open."+d})}})}return e.dialogsSensorInit=d,e}(t,i,e,d),a=function(e,i){(0,e.applicationMenusSensorInit)(),(0,i.dialogsSensorInit)()}(o,t)});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.pattern.help', location = 'aui.chunk.97b0b8778b1dab612048--f0ebc3ce920a8bb3519d.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.pattern.help"],{EoRc:function(i,u,n){},H4sa:function(i,u,n){"use strict";n("FStl"),n("Q0fs"),n("nqD9"),n("0FWE"),n("VtiI"),n("EoRc")}},[["H4sa","runtime","aui.splitchunk.vendors--7da3927366","aui.splitchunk.b0831cc7d0","aui.splitchunk.572a8bf5cd","aui.splitchunk.d49cf794d2","aui.splitchunk.d925afe2c0","aui.splitchunk.d727dd207f","aui.splitchunk.7da3927366"]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.helptips.jira-help-tips:analytics', location = 'util/analytics.js' */
define("jira-help-tips/util/analytics",["jira/analytics"],function(t){"use strict";function e(t){return t=""+(t||""),t.replace(/\./g,"-")}function i(i,n){if(n&&n.attributes.id){var r={},a=e(n.attributes.id),s="";n.attributes.eventPrefix&&(s=n.attributes.eventPrefix,"."!==s.charAt(s.length-1)&&(s+=".")),r.name=s+"helptips."+a+"."+i,r.properties={},t.send(r)}}return{clean:e,send:i}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.helptips.jira-help-tips:help-tip-manager', location = 'js/HelpTipManager.js' */
define("jira-help-tips/feature/help-tip-manager",["require"],function(s){"use strict";var e=s("underscore"),i=s("jquery"),t=s("jira/jquery/deferred"),n=s("wrm/context-path"),r=s("wrm/data"),d=n()+"/rest/helptips/1.0/tips",u=r.claim("com.atlassian.plugins.helptips.jira-help-tips:help-tip-manager.JiraHelpTipData");return{dismissedTipIds:[],sequences:[],loaded:new t,url:function(){return d},sync:function(s,e){var n=new t;return s||(s="get"),e||(e=null),"get"===s&&u&&u.dismissed?n.resolve(u.dismissed):i.ajax(this.url(),{type:s,dataType:"json",contentType:"application/json",data:e&&JSON.stringify(e),processData:!1}).done(function(s){n.resolve(s)}).fail(function(){n.reject()}),n.promise()},fetch:function(){var s=this.sync();return s.done(i.proxy(function(s){i.merge(this.dismissedTipIds,s),this.loaded.resolve()},this)),s.promise()},show:function(s){this.loaded.done(s)},dismiss:function(s){var e=s.attributes.id;e?(this.dismissedTipIds.push(e),this.sync("post",{id:e})):s._dismissed=!0},undismiss:function(s){var e=s.attributes.id;e?(this.dismissedTipIds.splice(i.inArray(e,this.dismissedTipIds),1),this.sync("delete",{id:e})):s._dismissed=!1},isDismissed:function(s){var e=s.attributes.id;return e?i.inArray(e,this.dismissedTipIds)>=0:s._dismissed},clearSequences:function(){this.sequences=[]},hideSequences:function(){e.each(this.sequences,function(s){s.view.dismiss()})},showSequences:function(){if(!this._showStarted){var s=this,t=0;this._showStarted=!0,i.when(this.loaded).done(function(){s.sequences.sort(function(s,e){return s.attributes.weight-e.attributes.weight}),s.sequences=e.filter(s.sequences,function(i){var n=-1===e.indexOf(s.dismissedTipIds,i.attributes.id);return n&&(i.attributes.position=t++),n}),s.sequences.length>0&&(e.last(s.sequences).attributes.showCloseButton=!0,s.sequences[0].show({force:!0})),s._showStarted=!1})}}}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.helptips.jira-help-tips:help-tip', location = 'templates/HelpTip.soy' */
// This file was automatically generated from HelpTip.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace AJS.Templates.HelpTip.
 */

if (typeof AJS == 'undefined') { var AJS = {}; }
if (typeof AJS.Templates == 'undefined') { AJS.Templates = {}; }
if (typeof AJS.Templates.HelpTip == 'undefined') { AJS.Templates.HelpTip = {}; }


AJS.Templates.HelpTip.tipContent = function(opt_data, opt_ignored) {
  return ((opt_data.title) ? '<h2 class="helptip-title">' + soy.$$escapeHtml(opt_data.title) + '</h2>' : '') + '<p class="helptip-body">' + soy.$$filterNoAutoescape(opt_data.bodyHtml) + '</p>' + ((opt_data.url) ? '<p><a class="helptip-link" href="' + soy.$$escapeHtml(opt_data.url) + '" target="_blank">' + ((opt_data.linkText) ? soy.$$escapeHtml(opt_data.linkText) : soy.$$escapeHtml('\u4e86\u89e3\u66f4\u591a')) + '</a></p>' : '') + AJS.Templates.HelpTip.tipFooter(opt_data);
};
if (goog.DEBUG) {
  AJS.Templates.HelpTip.tipContent.soyTemplateName = 'AJS.Templates.HelpTip.tipContent';
}


AJS.Templates.HelpTip.tipFooter = function(opt_data, opt_ignored) {
  return '<form class="tip-footer">' + AJS.Templates.HelpTip.nextButton(opt_data) + AJS.Templates.HelpTip.closeButton(opt_data) + AJS.Templates.HelpTip.sequencePaging(opt_data) + '</form>';
};
if (goog.DEBUG) {
  AJS.Templates.HelpTip.tipFooter.soyTemplateName = 'AJS.Templates.HelpTip.tipFooter';
}


AJS.Templates.HelpTip.nextButton = function(opt_data, opt_ignored) {
  return '' + ((opt_data.showNextButton) ? '<button class="aui-button helptip-next" type="button">' + ((opt_data.nextButtonText) ? soy.$$escapeHtml(opt_data.nextButtonText) : soy.$$escapeHtml('\u4e0b\u4e00\u6b65')) + '</button>' : '');
};
if (goog.DEBUG) {
  AJS.Templates.HelpTip.nextButton.soyTemplateName = 'AJS.Templates.HelpTip.nextButton';
}


AJS.Templates.HelpTip.closeButton = function(opt_data, opt_ignored) {
  return '' + ((opt_data.showCloseButton) ? '<button class="aui-button ' + ((opt_data.showNextButton) ? ' aui-button-link ' : '') + ' helptip-close" type="button">' + ((opt_data.closeButtonText) ? soy.$$escapeHtml(opt_data.closeButtonText) : soy.$$escapeHtml('\u5173\u95ed')) + '</button>' : '');
};
if (goog.DEBUG) {
  AJS.Templates.HelpTip.closeButton.soyTemplateName = 'AJS.Templates.HelpTip.closeButton';
}


AJS.Templates.HelpTip.sequencePaging = function(opt_data, opt_ignored) {
  return '' + ((opt_data.isSequence && opt_data.length > 1) ? '<span class="helptip-sequence-paging">' + soy.$$escapeHtml(opt_data.position + 1) + '/' + soy.$$escapeHtml(opt_data.length) + '</span>' : '');
};
if (goog.DEBUG) {
  AJS.Templates.HelpTip.sequencePaging.soyTemplateName = 'AJS.Templates.HelpTip.sequencePaging';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.helptips.jira-help-tips:help-tip', location = 'js/HelpTip.js' */
define("jira-help-tips/feature/help-tip",["require"],function(i){"use strict";function t(){return!1}function e(){return!0}function s(){return"jira-help-tip-"+p+d++}var n=i("jquery"),o=i("underscore"),h=i("jira-help-tips/feature/help-tip-manager"),l=i("jira-help-tips/util/analytics"),r=i("aui/inline-dialog"),a=i("jira/jquery/deferred"),u=AJS.Popups,c=AJS.Templates.HelpTip,d=0,p=(new Date).getTime(),f=function(i,t){this.initialize(i,t)},b=function(i){this.initialize(i)},m=function(i){var t;this.attributes=n.extend({},i),this.attributes.id||(this.attributes.id=!1),this.attributes.callbacks||(this.attributes.callbacks={}),this.attributes.isSequence&&(this.attributes.weight||(this.attributes.weight=Number.MAX_VALUE),h.sequences.push(this)),this.attributes.body&&(this.attributes.bodyHtml=this.attributes.body,delete this.attributes.body),this.cid=s(),t=this.attributes.anchor,delete this.attributes.anchor,this.view=t?new f(this,t):new b(this)};return n.extend(m.prototype,{show:function(i){i=i||{};var t=this,e=new a;if(this.attributes.callbacks.beforeShow){var s=this.attributes.callbacks.beforeShow();s&&o.isFunction(s.done)?s.done(e.resolve):e.resolve()}else e.resolve();e.done(function(){h.show(function(){t.isDismissed()||(!i.force&&u&&u.DisplayController?u.DisplayController.request({name:t.id,weight:1e3,show:function(){t.view.show()}}):t.view.show(),l.send("shown",t))})})},dismiss:function(){var i=l.clean(arguments[0]||"programmatically");this.view.dismiss(),"close-button"===i&&this.attributes.isSequence&&h.clearSequences(),this.isDismissed()||(h.dismiss(this),l.send("dismissed."+i,this))},isVisible:function(){return this.view.$el.is(":visible")},isDismissed:function(){return h.isDismissed(this)},refresh:function(){this.isDismissed()||this.view.refresh()},hide:function(){this.isDismissed()||this.view.dismiss()},showNextHelpTipInSequence:function(){this.view.clickNext()}}),n.extend(f.prototype,{initialize:function(i,e){this.model=i,this.anchorSelector=e,this.anchor=n(e),this._initDialog(e),n(document).bind("showLayer",function(e,s,o){"inlineDialog"===s&&o.id===i.cid&&(r.current=null,n(document.body).unbind("click."+i.cid+".inline-dialog-check"),o._validateClickToClose=t,o.hide=t)})},show:function(){this.beforeHide=t,this.popup.show()},refresh:function(){var i=n(this.anchorSelector);i.is(":visible")?i.get(0)!==this.anchor.get(0)?this.changeAnchor(i):this.isVisible()?this.popup.refresh():this.show():this.dismiss()},changeAnchor:function(i){var t=this.isVisible();this.dismiss(),this.$el.remove(),this.anchor=i,this._initDialog(i),t&&this.show()},dismiss:function(){this.beforeHide=e,this._popupHide()},clickNext:function(){var i=n(this.$el).find(".helptip-next");i.length>0&&i.click()},isVisible:function(){return this.$el.is(":visible")},_initDialog:function(i){var e=this,s=this.model;this.popup=new r(n(i),s.cid,o.bind(this._createDialog,this),o.extend({container:"#content",noBind:!0,preHideCallback:function(){return e.beforeHide()},calculatePositions:function(i,t,e,s){var o=r.opts.calculatePositions(i,t,e,s),h=n(this.container),l=h.offset();return"auto"!==o.popupCss.left&&(o.popupCss.left-=l.left,o.popupCss.right="auto"),o.popupCss.top-=l.top,o},addActiveClass:!1,initCallback:s.attributes.callbacks.init,hideCallback:s.attributes.callbacks.hide,persistent:!0},s.attributes.inlineDialogOpts)),this._popupHide=this.popup.hide,this.popup.hide=t,this.$el=n(this.popup[0]),this.$el.addClass("jira-help-tip aui-help")},_createDialog:function(i,t,e){var s=this,r=h.sequences,a=this.model.attributes.position,u=this.model.attributes.isSequence;i.removeClass("contents"),i.html(n(c.tipContent(o.extend({showNextButton:u&&r.length>1&&a+1<r.length,length:r.length,position:a,showCloseButton:!0},this.model.attributes)))),i.unbind("mouseover mouseout"),i.find(".helptip-link").click(function(){l.send("learn-more.clicked",s.model)}),i.find(".helptip-close").click(function(i){i.preventDefault(),s.model.dismiss("close-button")}),i.find(".helptip-next").click(function(i){i.preventDefault(),s.model.dismiss("next-button");var t=a+1;r[t]&&r[t].show({force:!0})}),e()}}),n.extend(b.prototype,{initialize:function(){this.$el=n("<div></div>"),this.$el.addClass("jira-help-tip aui-help")},show:function(){},dismiss:function(){}}),m});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.helptips.jira-help-tips:initialiser', location = 'js/initialiser.js' */
define("jira-help-tips/page/initialiser",["jira-help-tips/feature/help-tip-manager","jira/util/users/logged-in-user","underscore"],function(e,i,n){"use strict";function r(){i.isAnonymous()||e.fetch()}return{init:n.once(r)}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.helptips.jira-help-tips:common', location = 'js/legacy.js' */
!function(){"use strict";AJS.namespace("AJS.HelpTip",null,require("jira-help-tips/feature/help-tip")),AJS.namespace("AJS.HelpTip.Manager",null,require("jira-help-tips/feature/help-tip-manager")),require("jira-help-tips/page/initialiser").init()}();
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-quicksearch-plugin:commons_common', location = 'frontend/common.js' */
!function(e){function n(r){if(t[r])return t[r].exports;var o=t[r]={i:r,l:!1,exports:{}};return e[r].call(o.exports,o,o.exports,n),o.l=!0,o.exports}var r=window.atlassianWebpackJsonpe5ed4b01bb42b7893d8aef5bda9bf664;window.atlassianWebpackJsonpe5ed4b01bb42b7893d8aef5bda9bf664=function(t,a,i){for(var c,u,s,f=0,l=[];f<t.length;f++)u=t[f],o[u]&&l.push(o[u][0]),o[u]=0;for(c in a)Object.prototype.hasOwnProperty.call(a,c)&&(e[c]=a[c]);for(r&&r(t,a,i);l.length;)l.shift()();if(i)for(f=0;f<i.length;f++)s=n(n.s=i[f]);return s};var t={},o={2:0};n.e=function(e){function r(){c.onerror=c.onload=null,clearTimeout(u);var n=o[e];0!==n&&(n&&n[1](new Error("Loading chunk "+e+" failed.")),o[e]=void 0)}var t=o[e];if(0===t)return new Promise(function(e){e()});if(t)return t[2];var a=new Promise(function(n,r){t=o[e]=[n,r]});t[2]=a;var i=document.getElementsByTagName("head")[0];if({0:!0}[e])return WRM.require("wrc!com.atlassian.jira.plugins.jira-quicksearch-plugin:"+e),a;var c=document.createElement("script");c.type="text/javascript",c.charset="utf-8",c.async=!0,c.timeout=12e4,n.nc&&c.setAttribute("nonce",n.nc),c.src=n.p+"bundle."+e+".js";var u=setTimeout(r,12e4);return c.onerror=c.onload=r,i.appendChild(c),a},n.m=e,n.c=t,n.d=function(e,r,t){n.o(e,r)||Object.defineProperty(e,r,{configurable:!1,enumerable:!0,get:t})},n.n=function(e){var r=e&&e.__esModule?function(){return e.default}:function(){return e};return n.d(r,"a",r),r},n.o=function(e,n){return Object.prototype.hasOwnProperty.call(e,n)},n.p="",n.oe=function(e){throw console.error(e),e},"undefined"!=typeof AJS&&(n.p=AJS.contextPath()+"/download/resources/com.atlassian.jira.plugins.jira-quicksearch-plugin:assets-94230fb9-ee45-4f08-9618-f44c46c380ee/")}([]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.plugins.jira-quicksearch-plugin:entrypoint-jira-quicksearch-plugin', location = 'frontend/bundle.jira-quicksearch-plugin.js' */
atlassianWebpackJsonpe5ed4b01bb42b7893d8aef5bda9bf664([1],[function(e,t){e.exports=require("jquery")},function(e,t,n){"use strict";function i(e){return e&&e.__esModule?e:{default:e}}Object.defineProperty(t,"__esModule",{value:!0}),t.QUICK_SEARCH_RESULTS_EMPTY_STATE=t.QUICK_SEARCH_PARENT_LOADING_CLASS=t.QUICK_SEARCH_SPINNER_LOCATOR_CLASS=t.QUICK_SEARCH_SPINNER_CLASS=t.QUICK_SEARCH_INPUT_LOCATOR=t.QUICK_SEARCH_INPUT_ID=t.QUICK_SEARCH_FEATURE_FLAG=void 0;var r=n(5),u=i(r),a=n(2),o=i(a),s=o.default,c=(t.QUICK_SEARCH_FEATURE_FLAG="jira.quick.search",t.QUICK_SEARCH_INPUT_ID="quickSearchInput"),l=t.QUICK_SEARCH_INPUT_LOCATOR="#"+c,_=t.QUICK_SEARCH_SPINNER_CLASS="quick-search-spinner";t.QUICK_SEARCH_SPINNER_LOCATOR_CLASS=l+" ~ ."+_,t.QUICK_SEARCH_PARENT_LOADING_CLASS="quicksearch-loading",t.QUICK_SEARCH_RESULTS_EMPTY_STATE=Object.freeze([{id:"quick-search-issues",items:[],name:"\u95ee\u9898",url:(0,u.default)()+"/issues/?jql=order+by+lastViewed+DESC",viewAllTitle:"\u663e\u793a\u6240\u6709\u95ee\u9898"},{id:"quick-search-projects",items:[],name:"ultimate.search.group.projects.name",url:(0,u.default)()+"/projects?selectedCategory=all&selectedProjectType=all&contains=",viewAllTitle:"\u67e5\u770b\u6240\u6709\u9879\u76ee"}])},function(e,t){e.exports=require("jira/util/formatter")},function(e,t){e.exports=require("jira/analytics")},function(e,t){e.exports=require("underscore")},function(e,t){e.exports=require("wrm/context-path")},function(e,t,n){"use strict";function i(){return(0,c.default)(_.QUICK_SEARCH_SPINNER_LOCATOR_CLASS)}function r(){return(0,c.default)(_.QUICK_SEARCH_INPUT_LOCATOR)}function u(){return r().parent()}function a(){u().addClass(_.QUICK_SEARCH_PARENT_LOADING_CLASS),i().spin()}function o(){u().removeClass(_.QUICK_SEARCH_PARENT_LOADING_CLASS),i().spinStop()}Object.defineProperty(t,"__esModule",{value:!0}),t.spinStart=a,t.spinStop=o;var s=n(0),c=function(e){return e&&e.__esModule?e:{default:e}}(s),l=n(1),_=function(e){if(e&&e.__esModule)return e;var t={};if(null!=e)for(var n in e)Object.prototype.hasOwnProperty.call(e,n)&&(t[n]=e[n]);return t.default=e,t}(l);n(9)},function(e,t,n){"use strict";function i(e){return e&&e.__esModule?e:{default:e}}var r=n(0),u=i(r),a=n(4),o=i(a),s=n(3),c=i(s);n(8);var l=n(6),_=n(1),d=n(10),f=i(d),C=n(11),S=o.default.once(function(){(0,l.spinStart)(),n.e(0).then(n.bind(null,13))});(0,u.default)(function(){if(f.default.isFeatureEnabled(_.QUICK_SEARCH_FEATURE_FLAG)){c.default.send({name:"quicksearch.enabled"});var e=(0,u.default)(_.QUICK_SEARCH_INPUT_LOCATOR);e.after('<div class="'+_.QUICK_SEARCH_SPINNER_CLASS+'"></div>'),e.on("beforeBlurInput",function(e){e.preventDefault()}),e.focus(S),e.is(":focus")&&S()}}),(0,u.default)(window).load(function(){f.default.isFeatureEnabled(_.QUICK_SEARCH_FEATURE_FLAG)&&(0,C.showOnboardingTip)()})},function(e,t){},function(e,t){e.exports=void 0},function(e,t){e.exports=require("jira/featureflags/feature-manager")},function(e,t,n){"use strict";function i(e){return e&&e.__esModule?e:{default:e}}function r(e){return"<div><p>"+e+"</p></div>"}function u(){if(0!==(0,f.default)(S).length){var e={anchor:S,id:"qs-onboarding-tip",isSequence:!1,showCloseButton:!0,closeButtonText:"\u597d\u7684\uff0c\u77e5\u9053\u4e86",callbacks:{hide:function(){return c.default.send({name:"quicksearch.onboarding.tip.dissmised"})}},title:"\u66f4\u5feb\u627e\u5230\u60a8\u7684\u5de5\u4f5c",bodyHtml:r("\u901a\u8fc7\u60a8\u6240\u6709\u7684\u95ee\u9898\u548c\u9879\u76ee\u8fdb\u884c\u641c\u7d22\uff0c\u6700\u91cd\u8981\u7684\u5de5\u4f5c\u553e\u624b\u53ef\u5f97\u3002\u8f93\u5165\u540e\u5f00\u59cb\u641c\u7d22......")},t=new o.default(e);if(!t.isDismissed()){if(t.show(),!t.isVisible())return void c.default.send({name:"quicksearch.onboarding.tip.notshown"});c.default.send({name:"quicksearch.onboarding.tip.shown"})}}}Object.defineProperty(t,"__esModule",{value:!0}),t.showOnboardingTip=u;var a=n(12),o=i(a),s=n(3),c=i(s),l=n(2),_=i(l),d=n(0),f=i(d),C=_.default,S="#quicksearch-menu:visible"},function(e,t){e.exports=require("jira-help-tips/feature/help-tip")}],[7]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.pattern.label', location = 'aui.chunk.4c4e082d13db69ce6136--46feba69635d4ca59965.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.pattern.label"],{XpI9:function(i,n,u){},Y5aH:function(i,n,u){"use strict";u("FStl"),u("Q0fs"),u("nqD9"),u("XpI9")}},[["Y5aH","runtime","aui.splitchunk.b0831cc7d0","aui.splitchunk.572a8bf5cd","aui.splitchunk.d49cf794d2"]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.splitchunk.16f099a0da', location = 'aui.chunk.b8d538420d4f10a69910--faa970d0e91ff6ec4533.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.splitchunk.16f099a0da"],{jJ4L:function(i,n,o){"use strict";o("FStl"),o("Q0fs"),o("oTK+"),o("S3ao"),o("HNID"),o("YQ7q"),o("xjlV")},xjlV:function(i,n,o){}}]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.pattern.table', location = 'aui.chunk.f811d39e5d935038628f--b52208891d79a8157502.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.pattern.table"],[],[["jJ4L","runtime","aui.splitchunk.b0831cc7d0","aui.splitchunk.572a8bf5cd","aui.splitchunk.d49cf794d2","aui.splitchunk.336ae4f9e7","aui.splitchunk.fb15cffa72","aui.splitchunk.ed80e00f15","aui.splitchunk.b2ecdd4179","aui.splitchunk.16f099a0da"]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.component.dropdown2', location = 'aui.chunk.9b9bd0c99f46c24dfc3c--017c7b769c548ab21ee4.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.component.dropdown2"],[],[["XHZH","runtime","aui.splitchunk.vendors--894c8113d9","aui.splitchunk.vendors--b47d4b691d","aui.splitchunk.vendors--9c48cc20a9","aui.splitchunk.vendors--084821f40b","aui.splitchunk.vendors--d2297af84a","aui.splitchunk.0d131bcbf1","aui.splitchunk.b0831cc7d0","aui.splitchunk.444efc83be","aui.splitchunk.78b1c5d416","aui.splitchunk.54d3f16c20","aui.splitchunk.9bbf750f52","aui.splitchunk.e54c7c7304","aui.splitchunk.fb15cffa72","aui.splitchunk.56dfb54d0c","aui.splitchunk.f1e06f97a4","aui.splitchunk.479fe6ee76","aui.splitchunk.f673ef53ac","aui.splitchunk.9c48cc20a9","aui.splitchunk.8659b532c1","aui.splitchunk.5f851f97df","aui.splitchunk.d0110a864f","aui.splitchunk.afa5039e04","aui.splitchunk.bff3715233","aui.splitchunk.c750721820","aui.splitchunk.6d6f245ed3","aui.splitchunk.ed86a19e01","aui.splitchunk.084821f40b","aui.splitchunk.5b8c290363","aui.splitchunk.baa83dbaf9","aui.splitchunk.36cd9d521c","aui.splitchunk.e5af17b48e","aui.splitchunk.4248e12b20","aui.splitchunk.b548966f06"]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:rotp-menu', location = 'appswitcher/appswitcher.soy' */
// This file was automatically generated from appswitcher.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace navlinks.templates.appswitcher.
 */

if (typeof navlinks == 'undefined') { var navlinks = {}; }
if (typeof navlinks.templates == 'undefined') { navlinks.templates = {}; }
if (typeof navlinks.templates.appswitcher == 'undefined') { navlinks.templates.appswitcher = {}; }


navlinks.templates.appswitcher.linkSection = function(opt_data, opt_ignored) {
  var output = '';
  if (opt_data.list.length > 0) {
    output += '<div class="aui-nav-heading sidebar-section-header">' + soy.$$escapeHtml(opt_data.title) + '</div><ul class="aui-nav nav-links">';
    var linkList8 = opt_data.list;
    var linkListLen8 = linkList8.length;
    for (var linkIndex8 = 0; linkIndex8 < linkListLen8; linkIndex8++) {
      var linkData8 = linkList8[linkIndex8];
      output += navlinks.templates.appswitcher.applicationsItem(linkData8);
    }
    output += '</ul>';
  }
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.linkSection.soyTemplateName = 'navlinks.templates.appswitcher.linkSection';
}


navlinks.templates.appswitcher.applicationsItem = function(opt_data, opt_ignored) {
  return '<li class="nav-link"><a href="' + soy.$$escapeHtml(opt_data.link) + '" ' + ((opt_data.self) ? 'class="checked"' : '') + ' title="' + soy.$$escapeHtml(opt_data.link) + '"><span class="nav-link-label">' + soy.$$escapeHtml(opt_data.label) + '</span></a></li>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.applicationsItem.soyTemplateName = 'navlinks.templates.appswitcher.applicationsItem';
}


navlinks.templates.appswitcher.shortcutsItem = function(opt_data, opt_ignored) {
  return '<li class="nav-link"><a href="' + soy.$$escapeHtml(opt_data.link) + '" ' + ((opt_data.self) ? 'class="checked"' : '') + ' title="' + soy.$$escapeHtml(opt_data.link) + '"><span class="nav-link-label">' + soy.$$escapeHtml(opt_data.label) + '</span>' + ((opt_data.showDescription && opt_data.description) ? '<span class="nav-link-description">' + soy.$$escapeHtml(opt_data.description) + '</span>' : '') + '</a></li>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.shortcutsItem.soyTemplateName = 'navlinks.templates.appswitcher.shortcutsItem';
}


navlinks.templates.appswitcher.error = function(opt_data, opt_ignored) {
  return '<div class="app-switcher-error">' + soy.$$filterNoAutoescape('\u51fa\u4e86\u9519, \u8bf7\x3cspan class\x3d\x22app-switcher-retry\x22\x3e\u518d\u8bd5\u4e00\u6b21\x3c/span \x3e\u3002') + '</div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.error.soyTemplateName = 'navlinks.templates.appswitcher.error';
}


navlinks.templates.appswitcher.sidebarContents = function(opt_data, opt_ignored) {
  return '<div class="aui-page-panel-nav"><nav class="aui-navgroup aui-navgroup-vertical"><div class="app-switcher-section app-switcher-applications"><div class="aui-nav-heading">' + soy.$$escapeHtml('\u5e94\u7528\u7a0b\u5e8f\u94fe\u63a5') + '</div><div class="app-switcher-loading">' + soy.$$filterNoAutoescape('\u6b63\u5728\u4e0a\u4f20\u548c\u4e71\u7801\uff1b') + '</div></div><div class="app-switcher-section app-switcher-shortcuts"><div class="aui-nav-heading">' + soy.$$escapeHtml('\u5feb\u6377\u65b9\u5f0f') + '</div><div class="app-switcher-loading">' + soy.$$filterNoAutoescape('\u6b63\u5728\u4e0a\u4f20\u548c\u4e71\u7801\uff1b') + '</div></div></nav></div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.sidebarContents.soyTemplateName = 'navlinks.templates.appswitcher.sidebarContents';
}


navlinks.templates.appswitcher.trigger = function(opt_data, opt_ignored) {
  return '<span class="aui-icon aui-icon-small aui-iconfont-appswitcher">' + soy.$$escapeHtml('\u5df2\u94fe\u63a5\u5e94\u7528\u7a0b\u5e8f') + '</span>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.trigger.soyTemplateName = 'navlinks.templates.appswitcher.trigger';
}


navlinks.templates.appswitcher.projectHeaderSection = function(opt_data, opt_ignored) {
  return '<div class="app-switcher-title">' + aui.avatar.avatar({size: 'large', avatarImageUrl: opt_data.avatarUrl, isProject: true, title: opt_data.name}) + '<div class="sidebar-project-name">' + soy.$$escapeHtml(opt_data.name) + '</div></div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.projectHeaderSection.soyTemplateName = 'navlinks.templates.appswitcher.projectHeaderSection';
}


navlinks.templates.appswitcher.cogDropdown = function(opt_data, opt_ignored) {
  var output = '';
  var dropdownList__soy74 = '' + navlinks.templates.appswitcher.dropdownList({list: opt_data.links});
  output += aui.dropdown2.dropdown2({menu: {id: opt_data.id, content: dropdownList__soy74, extraClasses: 'aui-style-default sidebar-customize-section'}, trigger: {showIcon: false, content: '<span class="aui-icon aui-icon-small aui-iconfont-configure"></span>', container: '#app-switcher'}});
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.cogDropdown.soyTemplateName = 'navlinks.templates.appswitcher.cogDropdown';
}


navlinks.templates.appswitcher.dropdownList = function(opt_data, opt_ignored) {
  var output = '<ul class="sidebar-admin-links">';
  var linkList82 = opt_data.list;
  var linkListLen82 = linkList82.length;
  for (var linkIndex82 = 0; linkIndex82 < linkListLen82; linkIndex82++) {
    var linkData82 = linkList82[linkIndex82];
    output += '<li class="nav-link"><a href="' + soy.$$escapeHtml(linkData82.href) + '" title="' + soy.$$escapeHtml(linkData82.title) + '"><span class="nav-link-label">' + soy.$$escapeHtml(linkData82.label) + '</span></a></li>';
  }
  output += '</ul>';
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.dropdownList.soyTemplateName = 'navlinks.templates.appswitcher.dropdownList';
}


navlinks.templates.appswitcher.switcher = function(opt_data, opt_ignored) {
  var output = '';
  if (true) {
    if (AJS.DarkFeatures.isEnabled('rotp.sidebar')) {
      var sidebarContents__soy97 = '' + navlinks.templates.appswitcher.sidebarContents(null);
      var triggerContent__soy99 = '' + navlinks.templates.appswitcher.trigger(null);
      output += navlinks.templates.appswitcher.sidebar({sidebar: {id: 'app-switcher', content: sidebarContents__soy97}, trigger: {showIcon: false, content: triggerContent__soy99}});
    } else {
      output += navlinks.templates.appswitcher_old.switcher(null);
    }
  }
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.switcher.soyTemplateName = 'navlinks.templates.appswitcher.switcher';
}


navlinks.templates.appswitcher.sidebar = function(opt_data, opt_ignored) {
  return '<a href="' + soy.$$escapeHtml(opt_data.sidebar.id) + '" class="sidebar-trigger app-switcher-trigger" aria-owns="' + soy.$$escapeHtml(opt_data.sidebar.id) + '" aria-haspopup="true" data-is-user-admin="' + soy.$$escapeHtml(false) + '" data-is-sidebar="true" >' + soy.$$filterNoAutoescape(opt_data.trigger.content) + '</a><div id=' + soy.$$escapeHtml(opt_data.sidebar.id) + ' class="app-switcher-sidebar aui-style-default sidebar-offscreen">' + soy.$$filterNoAutoescape(opt_data.sidebar.content) + '</div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher.sidebar.soyTemplateName = 'navlinks.templates.appswitcher.sidebar';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:rotp-menu', location = 'appswitcher/appswitcher.js' */
(function(e,a){var d="is-user-admin";var c="#app-switcher";a.SideBar=function(f){var g=this;this.$sidebar=null;f=e.extend({sidebarContents:null},f);this.getLinks=function(){return e.ajax({url:AJS.contextPath()+"/rest/menu/latest/appswitcher",cache:false,dataType:"json"}).done(this.updateAppLinks).fail(this.showAppSwitcherError)};this.populateProjectHeader=function(i,h){g.getSidebar().find(".app-switcher-shortcuts .aui-nav-heading").after(navlinks.templates.appswitcher.projectHeaderSection({avatarUrl:h,name:i}))};this.getProjectData=function(){var h=e(".project-shortcut-dialog-trigger"),i=h.data("key"),j=h.data("entity-type");if(h.size()==0||!i||!j){e(".app-switcher-shortcuts").remove();return}var l,k;k=e.ajax({url:AJS.contextPath()+"/rest/project-shortcuts/1.0/local/"+i,cache:false,data:{entityType:j},dataType:"json"});l=e.ajax({url:AJS.contextPath()+"/rest/project-shortcuts/1.0/remote/"+i,cache:false,data:{entityType:j},dataType:"json"});e.when(k,l).then(function(n,m){g.updateProjectShortcuts(n,m,{key:i,entityType:j,name:h.data("name"),avatarUrl:h.find("img").prop("src")})},g.showProjectShortcutsError)};this.getSidebar=function(){if(!this.$sidebar){this.$sidebar=e(f.sidebarContents)}return this.$sidebar};this.addApplicationsCog=function(){e(".app-switcher-applications .aui-nav-heading").before(navlinks.templates.appswitcher.cogDropdown({id:"sidebar-applications-admin-dropdown",links:[{href:AJS.contextPath()+"/plugins/servlet/customize-application-navigator",label:"\u81ea\u5b9a\u4e49\u5bfc\u822a",title:"\u6dfb\u52a0\u65b0\u6761\u76ee\u3001\u9690\u85cf\u6216\u9650\u5236\u8c01\u770b\u89c1\u3002"},{href:AJS.contextPath()+"/plugins/servlet/applinks/listApplicationLinks",label:"\u7ba1\u7406\u5e94\u7528\u7a0b\u5e8f\u94fe\u63a5",title:"\u94fe\u63a5\u5230\u66f4\u591aAtlassian\u5e94\u7528\u7a0b\u5e8f"}]}))};this.addProjectShortcutsCog=function(h,j){var i=[{href:AJS.contextPath()+"/plugins/servlet/custom-content-links-admin?entityKey="+h,label:"\u81ea\u5b9a\u4e49\u5feb\u6377\u952e",title:""}];if(g.entityMappings[j]){i.push({href:g.generateEntityLinksUrl(h,g.entityMappings[j]),label:"\u7ba1\u7406\u4ea7\u54c1\u94fe\u63a5",title:""})}g.getSidebar().find(".app-switcher-shortcuts .aui-nav-heading").before(navlinks.templates.appswitcher.cogDropdown({id:"sidebar-project-shortcuts-admin-dropdown",links:i}))};this.updateAppLinks=function(h){e(function(){g.getSidebar().find(".app-switcher-applications").html(navlinks.templates.appswitcher.linkSection({title:"\u5e94\u7528\u7a0b\u5e8f\u94fe\u63a5",list:h}));if(g.getSidebar().data(d)){g.addApplicationsCog()}g.bindAnalyticsHandlers(g.getSidebar(),h)})};this.updateProjectShortcuts=function(k,i,j){var l=k[0].shortcuts,h=i[0].shortcuts;g.getSidebar().find(".app-switcher-shortcuts").html(navlinks.templates.appswitcher.linkSection({title:"\u5feb\u6377\u65b9\u5f0f",list:l.concat(h)}));if(g.getSidebar().data(d)){g.addProjectShortcutsCog(j.key,j.entityType)}g.populateProjectHeader(j.name,j.avatarUrl);g.bindAnalyticsHandlers(g.getSidebar(),data)};this.entityMappings={"confluence.space":"com.atlassian.applinks.api.application.confluence.ConfluenceSpaceEntityType","jira.project":"com.atlassian.applinks.api.application.jira.JiraProjectEntityType","bamboo.project":"com.atlassian.applinks.api.application.bamboo.BambooProjectEntityType","stash.project":"com.atlassian.applinks.api.application.stash.StashProjectEntityType"};this.generateEntityLinksUrl=function(h,i){if(i===g.entityMappings["confluence.space"]){return AJS.contextPath()+"/spaces/listentitylinks.action?typeId="+i+"&key="+h}else{return AJS.contextPath()+"/plugins/servlet/applinks/listEntityLinks/"+i+"/"+h}};this.showAppSwitcherError=function(){e(function(){var h=g.getSidebar();h.find(".app-switcher-applications .app-switcher-loading").replaceWith(navlinks.templates.appswitcher.error());h.off(".appswitcher").on("click.appswitcher",".app-switcher-retry",e.proxy(g.retryLoading,g))})};this.showProjectShortcutsError=function(){e(function(){var h=g.getSidebar();h.find(".app-switcher-shortcuts .app-switcher-loading").replaceWith(navlinks.templates.appswitcher.error());h.off(".appswitcher").on("click.appswitcher",".app-switcher-retry",e.proxy(g.retryLoading,g))})};this.retryLoading=function(h){this.getSidebar().html(navlinks.templates.appswitcher.sidebarContents());this.getLinks();this.getProjectData();h&&h.stopPropagation()};this.bindAnalyticsHandlers=function(h,i){};this.getLinks();e(this.getProjectData);this.toggleSidebar=function(j){var k=g.getSidebar(),i=e("body"),h=e(window.document);if(!i.hasClass("app-switcher-open")){var m=e("#header");k.css("left",-k.width());k.parent("body").length||k.appendTo("body");b({data:k});k.animate({left:0},300);function l(n){var p=n.target&&e(n.target),o=n.keyCode;if(n.originalEvent===j.originalEvent){return}if(p&&!o&&!(p.closest(k).length||p.closest(m).length)&&j.which==1&&!(n.shiftKey||n.ctrlKey||n.metaKey)){g.toggleSidebar()}else{if(o===27){g.toggleSidebar()}}}h.on("click.appSwitcher",l);h.on("keydown.appSwitcher",l);h.on("scroll.appSwitcher",k,b)}else{h.off(".appSwitcher")}i.toggleClass("app-switcher-open")};e("#header").on("click",".app-switcher-trigger",this.toggleSidebar)};function b(h){var f=e(document).scrollTop(),i=e("#header"),g=(i.height()+i.offset().top)-f;if(g>=0){h.data.css({top:g,position:"fixed"})}else{h.data.css({top:0,left:0,position:"fixed"})}}e(function(){if(e(c).data("is-sidebar")===true){new a.SideBar({sidebarContents:c})}})}(jQuery,window.NL=(window.NL||{})));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:rotp-menu', location = 'appswitcher/appswitcher_old.js' */
var APPSWITCHER_TRIGGER_CLICK="appswitcher.trigger.click";var APPSWITCHER_DROPDOWN_SHOW="appswitcher.dropdown.show";var APPSWITCHER_DROPDOWN_DISPLAY_ERROR="appswitcher.dropdown.display.error";var APPSWITCHER_APP_LINK_CLICK="appswitcher.app.link.click";var APPSWITCHER_CONFIGURE_LINK_CLICK="appswitcher.configure.link.click";(function(c,f){var b="isAppSuggestionAvailable";var d="isSiteAdminUser";var e="isUserAdmin";var a="#app-switcher";f.AppSwitcher=function(j){var l=AJS.contextPath()+"/plugins/servlet/customize-application-navigator";var k="unified.usermanagement";var m=this;this.$dropdown=null;j=c.extend({dropdownContents:null},j);this.getLinks=function(){return c.ajax({url:AJS.contextPath()+"/rest/menu/latest/appswitcher",cache:false,dataType:"json"}).done(this.updateDropdown).fail(this.showError)};this.getDropdown=function(){if(!this.$dropdown){this.$dropdown=c(j.dropdownContents);this.envData=this.$dropdown.data("environment")}return this.$dropdown};this.updateDropdown=function(n){c(function(){m.getDropdown().html(navlinks.templates.appswitcher_old.applications({apps:n,showAdminLink:m.envData[e],adminLink:l}));m.bindAnalyticsHandlers();if(m.envData[b]===true){m.handleSuggestionApps(n)}})};this.bindAnalyticsHandlers=function(){c(".app-switcher-trigger").on("click",function(){AJS.trigger("analyticsEvent",{name:APPSWITCHER_TRIGGER_CLICK})});c("#app-switcher").on("aui-dropdown2-show",function(){AJS.trigger("analyticsEvent",{name:APPSWITCHER_DROPDOWN_SHOW})});c("#app-switcher .nav-link").on("click",function(){var p="custom";var q=c(this).find("a");var o=q.attr("href");var n=window.location.hostname;if(o&&o.indexOf("bitbucket.org")>-1){p="bitbucket-cloud"}else{if(o.indexOf(n+"/wiki")>-1){p="confluence"}else{if(o.indexOf(n+"/build")>-1){p="bamboo"}else{if(o.indexOf(n)>-1){p="jira"}}}}AJS.trigger("analyticsEvent",{name:APPSWITCHER_APP_LINK_CLICK,data:{product:p}})});c(".nav-link-edit-wrapper").on("click",function(){AJS.trigger("analyticsEvent",{name:APPSWITCHER_CONFIGURE_LINK_CLICK})})};this.isBillingSystemEnabled=function(){return(this.envData[d]===true)&&AJS.DarkFeatures.isEnabled(k)};this.handleSuggestionApps=function(q){var r=_.map(q,function(s){return s.applicationType.toLowerCase()});var o=c("<div id='app-switcher-suggestion-apps' class='aui-dropdown2-section'/>");o.html(navlinks.templates.appswitcher_old.suggestionApps);var p=o.find(".suggestion-apps");var n=false;_.each(g,function(s){if(!_.contains(r,s.appName)){n=true;p.append(navlinks.templates.appswitcher_old.suggestionApp({suggestionApp:s,isBillingSystemEnabled:m.isBillingSystemEnabled()}))}});if(!n){return}c("#app-switcher").append(o);c(".app-discovery-suggestion-app").click(function(){var t=c(this).find("a");var s;if(m.envData[d]===true){s="appswitcher.discovery.siteadmin.select.inproduct."}else{s="appswitcher.discovery.user.select."}s=s+t.attr("id").toLowerCase();AJS.trigger("analytics",{name:s})});c(".app-discovery-suggestion-app").hover(function(){c(this).find("a").removeClass("active").removeClass("aui-dropdown2-active")});c(".app-discovery-cancel-button").click(function(){AJS.trigger("analytics",{name:"appswitcher.discovery.nothanks.button.click"});i(h,"true");o.remove()})};this.showError=function(){c(function(){AJS.trigger("analyticsEvent",{name:APPSWITCHER_DROPDOWN_DISPLAY_ERROR});m.getDropdown().html(navlinks.templates.appswitcher_old.error()).off(".appswitcher").on("click.appswitcher",".app-switcher-retry",c.proxy(m.retryLoading,m))})};this.retryLoading=function(n){this.getDropdown().html(navlinks.templates.appswitcher_old.loading());this.getLinks();n&&n.stopPropagation()};this.getLinks()};var h="key-no-thanks";var g=[{appName:"jira",appDesc:"\u95ee\u9898\u548c\u9879\u76ee\u8ddf\u8e2a\u8f6f\u4ef6",discoveryUrl:"https://www.atlassian.com/software/jira",billingSystemDiscoveryUrl:"/admin/billing/addapplication"},{appName:"confluence",appDesc:"\u534f\u4f5c\u548c\u5185\u5bb9\u5171\u4eab",discoveryUrl:"https://www.atlassian.com/software/confluence",billingSystemDiscoveryUrl:"/admin/billing/addapplication?product=confluence.ondemand"},{appName:"bamboo",appDesc:"\u6301\u7eed\u96c6\u6210",discoveryUrl:"https://www.atlassian.com/software/bamboo",billingSystemDiscoveryUrl:"/admin/billing/addapplication?product=bamboo.ondemand"}];var i=function(j,k){c.ajax({url:AJS.contextPath()+"/rest/menu/latest/userdata/",type:"PUT",contentType:"application/json",data:JSON.stringify({key:j,value:k})})};c(function(){if(c(a).data("is-switcher")===true){new f.AppSwitcher({dropdownContents:a})}})}(jQuery,window.NL=(window.NL||{})));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.plugins.atlassian-nav-links-plugin:rotp-menu', location = 'appswitcher/appswitcher_old.soy' */
// This file was automatically generated from appswitcher_old.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace navlinks.templates.appswitcher_old.
 */

if (typeof navlinks == 'undefined') { var navlinks = {}; }
if (typeof navlinks.templates == 'undefined') { navlinks.templates = {}; }
if (typeof navlinks.templates.appswitcher_old == 'undefined') { navlinks.templates.appswitcher_old = {}; }


navlinks.templates.appswitcher_old.applications = function(opt_data, opt_ignored) {
  return '' + navlinks.templates.appswitcher_old.applicationsSection({list: opt_data.apps, listClass: 'nav-links', showDescription: opt_data.showDescription}) + ((opt_data.custom) ? navlinks.templates.appswitcher_old.applicationsSection({list: opt_data.custom, showDescription: opt_data.showDescription}) : '') + ((opt_data.showAdminLink) ? navlinks.templates.appswitcher_old.adminSection(opt_data) : '');
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.applications.soyTemplateName = 'navlinks.templates.appswitcher_old.applications';
}


navlinks.templates.appswitcher_old.applicationsSection = function(opt_data, opt_ignored) {
  var output = '';
  if (opt_data.list.length > 0) {
    var param19 = '<ul' + ((opt_data.listClass) ? ' class="' + soy.$$escapeHtml(opt_data.listClass) + '"' : '') + '>';
    var linkList27 = opt_data.list;
    var linkListLen27 = linkList27.length;
    for (var linkIndex27 = 0; linkIndex27 < linkListLen27; linkIndex27++) {
      var linkData27 = linkList27[linkIndex27];
      param19 += navlinks.templates.appswitcher_old.applicationsItem(soy.$$augmentMap(linkData27, {showDescription: opt_data.showDescription}));
    }
    param19 += '</ul>';
    output += aui.dropdown2.section({content: param19});
  }
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.applicationsSection.soyTemplateName = 'navlinks.templates.appswitcher_old.applicationsSection';
}


navlinks.templates.appswitcher_old.applicationsItem = function(opt_data, opt_ignored) {
  return '<li class="nav-link' + ((opt_data.self) ? ' nav-link-local' : '') + '"><a href="' + soy.$$escapeHtml(opt_data.link) + '" class="aui-dropdown2-radio ' + ((opt_data.self) ? 'aui-dropdown2-checked' : '') + '" title="' + soy.$$escapeHtml(opt_data.link) + '"><span class="nav-link-label">' + soy.$$escapeHtml(opt_data.label) + '</span>' + ((opt_data.showDescription && opt_data.description) ? '<span class="nav-link-description">' + soy.$$escapeHtml(opt_data.description) + '</span>' : '') + '</a></li>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.applicationsItem.soyTemplateName = 'navlinks.templates.appswitcher_old.applicationsItem';
}


navlinks.templates.appswitcher_old.adminSection = function(opt_data, opt_ignored) {
  return '' + aui.dropdown2.section({content: '<ul class="nav-links"><li><a class="nav-link-edit-wrapper" href="' + soy.$$escapeHtml(opt_data.adminLink) + '"><span class="nav-link-edit">' + soy.$$filterNoAutoescape('\u914d\u7f6e...') + '</span></a></li></ul>'});
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.adminSection.soyTemplateName = 'navlinks.templates.appswitcher_old.adminSection';
}


navlinks.templates.appswitcher_old.error = function(opt_data, opt_ignored) {
  return '<div class="app-switcher-error">' + soy.$$filterNoAutoescape('\u51fa\u4e86\u9519, \u8bf7\x3cspan class\x3d\x22app-switcher-retry\x22\x3e\u518d\u8bd5\u4e00\u6b21\x3c/span \x3e\u3002') + '</div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.error.soyTemplateName = 'navlinks.templates.appswitcher_old.error';
}


navlinks.templates.appswitcher_old.loading = function(opt_data, opt_ignored) {
  return '<div class="app-switcher-loading">' + soy.$$filterNoAutoescape('\u6b63\u5728\u4e0a\u4f20\u548c\u4e71\u7801\uff1b') + '</div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.loading.soyTemplateName = 'navlinks.templates.appswitcher_old.loading';
}


navlinks.templates.appswitcher_old.trigger = function(opt_data, opt_ignored) {
  return '<span class="aui-icon aui-icon-small aui-iconfont-appswitcher">' + soy.$$escapeHtml('\u5df2\u94fe\u63a5\u5e94\u7528\u7a0b\u5e8f') + '</span>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.trigger.soyTemplateName = 'navlinks.templates.appswitcher_old.trigger';
}


navlinks.templates.appswitcher_old.switcher = function(opt_data, opt_ignored) {
  var output = '';
  if (true) {
    var loadingContent__soy81 = '' + navlinks.templates.appswitcher_old.loading(null);
    var triggerContent__soy83 = '' + navlinks.templates.appswitcher_old.trigger(null);
    output += aui.dropdown2.dropdown2({menu: {id: 'app-switcher', content: loadingContent__soy81, extraClasses: 'aui-style-default', extraAttributes: {'data-environment': {}, 'data-is-switcher': 'true'}}, trigger: {showIcon: false, content: triggerContent__soy83, extraClasses: 'app-switcher-trigger', extraAttributes: {href: '#app-switcher'}}});
  }
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.switcher.soyTemplateName = 'navlinks.templates.appswitcher_old.switcher';
}


navlinks.templates.appswitcher_old.suggestionApp = function(opt_data, opt_ignored) {
  var output = '';
  var href__soy89 = opt_data.isBillingSystemEnabled == true ? opt_data.suggestionApp.billingSystemDiscoveryUrl : opt_data.suggestionApp.discoveryUrl;
  output += '<li class="app-discovery-suggestion-app"><a id="' + soy.$$escapeHtml(opt_data.suggestionApp.appName) + '" href="' + soy.$$escapeHtml(href__soy89) + '" class="app-discovery-link aui-icon-container app-discovery-' + soy.$$escapeHtml(opt_data.suggestionApp.appName) + '-product-icon" title="' + soy.$$escapeHtml(href__soy89) + '" target="_blank"/><div class="app-discovery-small">' + soy.$$escapeHtml(opt_data.suggestionApp.appDesc) + '</div></li>';
  return output;
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.suggestionApp.soyTemplateName = 'navlinks.templates.appswitcher_old.suggestionApp';
}


navlinks.templates.appswitcher_old.suggestionApps = function(opt_data, opt_ignored) {
  return '<ul class=\'nav-links suggestion-apps\'><li><span class=\'app-discovery-suggest-title nav-link-label\'><h6>' + soy.$$escapeHtml('\u5c1d\u8bd5\u5176\u5b83Atlassian\u5e94\u7528\u7a0b\u5e8f') + '</h6></span></li></ul><div class=\'buttons-container app-discovery-suggest-apps-buttons\'><div class=\'buttons\'><button class=\'aui-button aui-button-link app-discovery-cancel-button\' name=\'cancel\' accesskey=\'c\' href=\'#\'>' + soy.$$escapeHtml('\u4e0d\u518d\u663e\u793a\u6b64\u9879') + '</button></div></div>';
};
if (goog.DEBUG) {
  navlinks.templates.appswitcher_old.suggestionApps.soyTemplateName = 'navlinks.templates.appswitcher_old.suggestionApps';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:header-templates', location = 'soy/headerDropdown.soy' */
// This file was automatically generated from headerDropdown.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.Menu.Dropdowns.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.Menu == 'undefined') { JIRA.Templates.Menu = {}; }
if (typeof JIRA.Templates.Menu.Dropdowns == 'undefined') { JIRA.Templates.Menu.Dropdowns = {}; }


JIRA.Templates.Menu.Dropdowns.dropdown2Fragment = function(opt_data, opt_ignored) {
  var output = '';
  var sectionList3 = opt_data.sections;
  var sectionListLen3 = sectionList3.length;
  for (var sectionIndex3 = 0; sectionIndex3 < sectionListLen3; sectionIndex3++) {
    var sectionData3 = sectionList3[sectionIndex3];
    output += JIRA.Templates.Menu.Dropdowns.dropdown2Section(sectionData3);
  }
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.Menu.Dropdowns.dropdown2Fragment.soyTemplateName = 'JIRA.Templates.Menu.Dropdowns.dropdown2Fragment';
}


JIRA.Templates.Menu.Dropdowns.dropdown2Section = function(opt_data, opt_ignored) {
  opt_data = opt_data || {};
  var output = '';
  var hasItems__soy7 = opt_data.items && opt_data.items.length > 0;
  if (hasItems__soy7) {
    output += '<div' + ((opt_data.id) ? ' id="' + soy.$$escapeHtml(opt_data.id) + '"' : '') + ' class="aui-dropdown2-section' + ((opt_data.style) ? soy.$$escapeHtml(opt_data.style) : '') + '" >' + ((opt_data.label) ? '<strong>' + soy.$$escapeHtml(opt_data.label) + '</strong>' : '') + '<ul class="aui-list-truncate">';
    var itemList27 = opt_data.items;
    var itemListLen27 = itemList27.length;
    for (var itemIndex27 = 0; itemIndex27 < itemListLen27; itemIndex27++) {
      var itemData27 = itemList27[itemIndex27];
      output += '<li' + ((itemData27.id) ? ' id="' + soy.$$escapeHtml(itemData27.id) + '"' : '') + ((itemData27.style) ? ' class="' + soy.$$escapeHtml(itemData27.style) + '"' : '') + '>' + JIRA.Templates.Menu.Dropdowns.dropdown2Item(itemData27) + '</li>';
    }
    output += '</ul></div>';
  }
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.Menu.Dropdowns.dropdown2Section.soyTemplateName = 'JIRA.Templates.Menu.Dropdowns.dropdown2Section';
}


JIRA.Templates.Menu.Dropdowns.dropdown2Item = function(opt_data, opt_ignored) {
  var output = '<a href="' + soy.$$escapeHtml(opt_data.url) + '"' + ((opt_data.parameters && opt_data.parameters.trackClick) ? ' data-track-click="' + soy.$$escapeHtml(opt_data.parameters.trackClick) + '"' : '') + ((opt_data.id) ? ' id="' + soy.$$escapeHtml(opt_data.id) + '_lnk"' : '') + ' class="' + ((opt_data.iconUrl) ? 'aui-icon-container' : '') + ((opt_data.parameters && opt_data.parameters['class']) ? ' ' + soy.$$escapeHtml(opt_data.parameters['class']) : '') + '"';
  if (opt_data.parameters) {
    var keyList69 = soy.$$getMapKeys(opt_data.parameters);
    var keyListLen69 = keyList69.length;
    for (var keyIndex69 = 0; keyIndex69 < keyListLen69; keyIndex69++) {
      var keyData69 = keyList69[keyIndex69];
      output += (keyData69 != 'class' && keyData69 != 'iconurl') ? ' ' + soy.$$escapeHtml(keyData69) + '="' + soy.$$escapeHtml(opt_data.parameters[keyData69]) + '"' : '';
    }
  }
  output += '>' + ((opt_data.iconUrl) ? '<img class="icon" src="' + soy.$$escapeHtml(opt_data.iconUrl) + '" alt="" />' : '') + soy.$$escapeHtml(opt_data.label) + '</a>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.Menu.Dropdowns.dropdown2Item.soyTemplateName = 'JIRA.Templates.Menu.Dropdowns.dropdown2Item';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:analytics-header', location = 'js/header-analytics.js' */
define("jira/header/init-header-analytics",["jira/analytics","jira/util/logger","jquery"],function(n,t,e){"use strict";function i(n,t){var e=void 0;return n.forEach(function(n){n.id===t&&(e=n.analyticEventKey)}),e}function o(o){var m=e(o);m.on("click","#logo",function(){n.send({name:a+"home"}),t.trace(r)}),m.on("aui-dropdown2-show",l,function(){var n=e(this).attr("id");c(u,n)&&d(n,u,".open")}),m.on("aui-dropdown2-hide",l,function(){var n=e(this).attr("id");c(u,n)&&d(n,u,".close")}),m.on("aui-dropdown2-show",s,function(){var n=e(this);n.trigger("aui-dropdown2-show-before"),n.trigger("aui-dropdown2-show-after")}),m.on("click",y,function(){var n=this.getAttribute("aria-controls");c(u,n)&&d(n,u,".click")}),m.on("click",".aui-dropdown2 .aui-dropdown2-section a",function(o){var c=e(o.target).closest(l).first();if(null!==c){var d=c.find("a").index(o.target),s=c.attr("id"),y=i(u,s);n.send({name:a+y+".item.click",itemIndex:d}),t.trace(r)}})}var a="jira.navigation.header.",r="jira.header.analytics.event",c=function(n,t){return n.some(function(n){return n.id===t})},d=function(e,o,c){var d=i(o,e),u=a+d+c;n.send({name:u}),t.trace(r)},u=[{id:"home_link-content",analyticEventKey:"dashboards"},{id:"browse_link-content",analyticEventKey:"projects"},{id:"find_link-content",analyticEventKey:"issues"},{id:"tempo_menu-content",analyticEventKey:"tempo"},{id:"bonfire_top_menu_dropdown-content",analyticEventKey:"capture"},{id:"greenhopper_menu-content",analyticEventKey:"boards"},{id:"plugins-jira-webitem-main-content",analyticEventKey:"portfolio"},{id:"system-help-menu-content",analyticEventKey:"help"},{id:"system-admin-menu-content",analyticEventKey:"admin"}],s="#system-help-menu-content, #user-options-content, #system-admin-menu-content",l=u.map(function(n){return"#"+n.id}).join(", "),y=u.reduce(function(n,t){return n+"a.aui-dropdown2-trigger[aria-owns="+t.id+"]"},"");return{init:o}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:jira-header-assets', location = 'js/init-dropdown2.js' */
define("jira/header/dropdown/element",["jira/ajs/ajax/smart-ajax","jira/jquery/deferred","jira/skate","jira/util/data/meta","wrm/context-path","jquery"],function(e,a,t,o,n,r){"use strict";return t("aui-dropdown2-ajax",{type:t.type.CLASSNAME,attached:function(t){function d(){document.activeElement===s.get(0)&&u.get(0).focusItem(0)}function i(){s.data("kbd-interacted",!1)}var u=r(document.getElementById(t.getAttribute("aria-controls"))),s=r(t),c=u.data("aui-dropdown2-ajax-key");s.on("keydown.aui-dropdown2-ajax",function(){s.data("kbd-interacted",!0),setTimeout(i,4)}),u.on("aui-dropdown2-show-after.aui-dropdown2-ajax",i),i(),u.on("aui-dropdown2-show.aui-dropdown2-ajax",function(){if(u.trigger("aui-dropdown2-show-before"),c){var t=new a,i=new a;setTimeout(function(){t.resolve(s.data("kbd-interacted"))},0),u.empty(),u.addClass("aui-dropdown2-loading"),s.attr("aria-busy","true"),e.makeRequest({url:n()+"/rest/api/1.0/menus/"+c,data:{inAdminMode:o.getBoolean("in-admin-mode")},dataType:"json",cache:!1,success:function(e){i.resolve(e)},error:i.reject}),r.when(t,i).done(function(e,a){u.html(JIRA.Templates.Menu.Dropdowns.dropdown2Fragment(a)),e&&setTimeout(d,0)}).always(function(){s.attr("aria-busy","false"),u.removeClass("aui-dropdown2-loading"),u.trigger("aui-dropdown2-show-after")})}else u.trigger("aui-dropdown2-show-after")})},detached:function(e){var a=r(document.getElementById(e.getAttribute("aria-controls"))),t=r(e);a.off(".aui-dropdown2-ajax"),t.off(".aui-dropdown2-ajax")}})}),function(e){"use strict";e(document).on("aui-responsive-menu-item-created",".aui-header.aui-dropdown2-trigger-group",function(a){var t=a.originalEvent.detail;e(t.originalItem).find("> .aui-dropdown2-ajax").length>0&&e(t.newItem).find("> .aui-dropdown2-trigger").addClass("aui-dropdown2-ajax")})}(require("jquery"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:jira-header-assets', location = 'js/app-switcher-analytics.js' */
define("jira/header/init-app-switcher-analytics",["jira/analytics","jquery"],function(i,n){"use strict";function a(a){var e=n(a);e.on("aui-dropdown2-show-before","#app-switcher",function(){i.send({name:"navigation.header.appswitcher.open"})}),e.on("click","#app-switcher a, #app-switcher button",function(a){var e,t=n(a.target),o=t.closest("li").index();t.is(".nav-link-edit")?e="configure":t.is(".app-discovery-link")?e=this.id?this.id:"unknown.app.discovery":t.is(".app-discovery-cancel-button")?(e="dont.show",o=0):e="application",i.send({name:"navigation.header.appswitcher.click",data:{position:o,linkType:e}})}),e.on("aui-dropdown2-hide","#app-switcher",function(){i.send({name:"navigation.header.appswitcher.close"})})}return{init:a}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-header-plugin:jira-header', location = 'js/init-header.js' */
require(["jira/header/dropdown/element","jira/header/init-header-analytics","jira/header/init-app-switcher-analytics"],function(e,i,t){"use strict";i.init(document),t.init(document)});
}catch(e){WRMCB(e)};