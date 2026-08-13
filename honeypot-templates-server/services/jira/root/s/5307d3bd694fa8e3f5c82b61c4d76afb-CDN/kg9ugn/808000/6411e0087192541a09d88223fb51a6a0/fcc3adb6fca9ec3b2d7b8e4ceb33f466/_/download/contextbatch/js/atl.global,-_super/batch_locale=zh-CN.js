WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:user-message-flags', location = '/includes/jira/admin/admin-flags.js' */
require(["jira/util/logger","wrm/data","jquery","jira/flag"],function(a,s,e,i){"use strict";e(function(){var c=s.claim("jira.core:user-message-flags-data.adminLockout")||{};if(c.noprojects){var n=JIRA.Templates.Flags.Admin,r=n.adminIssueAccessFlagTitle({}),l=n.adminIssueAccessFlagBody({manageAccessUrl:c.manageAccessUrl}),g=i.showWarningMsg(r,l,{dismissalKey:c.flagId});e(g).find("a").on("click",function(){g.dismiss()})}a.trace("admin.flags.done")})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:user-message-flags', location = '/includes/jira/admin/admin-flags.soy' */
// This file was automatically generated from admin-flags.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.Flags.Admin.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.Flags == 'undefined') { JIRA.Templates.Flags = {}; }
if (typeof JIRA.Templates.Flags.Admin == 'undefined') { JIRA.Templates.Flags.Admin = {}; }


JIRA.Templates.Flags.Admin.adminIssueAccessFlagTitle = function(opt_data, opt_ignored) {
  return '' + soy.$$escapeHtml('\u5f53\u524d\u9650\u5236');
};
if (goog.DEBUG) {
  JIRA.Templates.Flags.Admin.adminIssueAccessFlagTitle.soyTemplateName = 'JIRA.Templates.Flags.Admin.adminIssueAccessFlagTitle';
}


JIRA.Templates.Flags.Admin.adminIssueAccessFlagBody = function(opt_data, opt_ignored) {
  var output = '';
  var start__soy5 = '<a href="' + soy.$$escapeHtml("") + '/' + soy.$$escapeHtml(opt_data.manageAccessUrl) + '">';
  var end__soy11 = '</a>';
  output += '<p id="admin-noprojects">' + soy.$$escapeHtml('\u60a8\u4ee5\u7ba1\u7406\u5458\u7684\u8eab\u4efd\u767b\u5f55, \u4f46\u662f\u60a8\u4e0d\u5177\u6709\u8bbf\u95ee\u4efb\u4f55 Jira \u7684\u5e94\u7528\u7a0b\u5e8f\u3002\u60a8\u53ef\u4ee5\u6267\u884c\u7ba1\u7406\u4efb\u52a1, \u4f46\u60a8\u5c06\u65e0\u6cd5\u67e5\u770b\u6216\u7f16\u8f91\u95ee\u9898\u3002') + '</p><p><a href="' + soy.$$escapeHtml("") + '/' + soy.$$escapeHtml(opt_data.manageAccessUrl) + '">' + soy.$$escapeHtml('\u7ba1\u7406\u60a8\u7684\u5e94\u7528\u7a0b\u5e8f\u8bbf\u95ee\u6743\u3002') + '</a></p>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.Flags.Admin.adminIssueAccessFlagBody.soyTemplateName = 'JIRA.Templates.Flags.Admin.adminIssueAccessFlagBody';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.codebarrel.addons.automation:manual-trigger-client-resources', location = '/page/viewissue/initManualTriggers.js' */
define("codebarrel/init-manual-triggers",["jquery","jira/flag","jira/util/events","jira/util/events/types","jira/issue","jira/util/browser",],function($,flag,Events,EventTypes,Issue,Browser){const isIssuePage=function(){return JIRA&&JIRA.Issue&&JIRA.Issue.getStalker&&JIRA.Issue.getStalker().length>0};const showSpinner=function(){const spinContainer=$('<div class="toolbar-group" style="width:16px;height:16px;padding-top: 4px;"/>');const issueOperations=$("#stalker .ops-menus .toolbar-split-left");if(issueOperations.length>0){issueOperations.append(spinContainer);spinContainer.spin()}return spinContainer};const executeManualTrigger=function(url){return $.ajax({url:url,type:"POST",contentType:"application/json",dataType:"json",success:function(){flag.showSuccessMsg("Automation","Automation rule was successfully executed.")},error:function(){flag.showErrorMsg("Automation","Error executing automation rule. Please reload the issue and try again.")}})};$(document).on("click",".cb-manual-rule-trigger",function(e){e.preventDefault();const spinner=showSpinner();const url=$(this).attr("href");const urlWithoutReturnUrl=url.split("&returnUrl=")[0];executeManualTrigger(urlWithoutReturnUrl).always(function(){spinner.spinStop()}).done(function(){if(isIssuePage()){Events.trigger(EventTypes.REFRESH_ISSUE_PAGE,[Issue.getIssueId()])}else{Browser.reloadViaWindowLocation()}})})});require("codebarrel/init-manual-triggers");
}catch(e){WRMCB(e)};