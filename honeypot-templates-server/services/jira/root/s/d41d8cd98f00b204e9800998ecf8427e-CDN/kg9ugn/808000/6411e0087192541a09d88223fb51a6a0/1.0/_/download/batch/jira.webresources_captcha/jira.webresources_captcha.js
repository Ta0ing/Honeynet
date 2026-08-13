WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:captcha', location = '/includes/jira/captcha/Captcha.js' */
define("jira/captcha",["jquery"],function(a){var c={setup:function(){a("#captcha").delegate("span.captcha-reload","click",function(a){c.refresh(),a.preventDefault()})},refresh:function(){var c=a(".captcha-image","#captcha .captcha-container"),r=c.attr("src");r=r.indexOf("__r")>=0?r.replace(/__r=([^&]+)/,"__r="+Math.random()):r.indexOf("?")>=0?r+"&__r="+Math.random():r+"?__r="+Math.random(),c.attr("src",r),a("#captcha .captcha-response").focus()}};return c}),AJS.namespace("JIRA.Captcha",null,require("jira/captcha"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:captcha', location = '/includes/jira/captcha/captcha-init.js' */
require(["jira/captcha","jquery"],function(e,r){r(e.setup)});
}catch(e){WRMCB(e)};