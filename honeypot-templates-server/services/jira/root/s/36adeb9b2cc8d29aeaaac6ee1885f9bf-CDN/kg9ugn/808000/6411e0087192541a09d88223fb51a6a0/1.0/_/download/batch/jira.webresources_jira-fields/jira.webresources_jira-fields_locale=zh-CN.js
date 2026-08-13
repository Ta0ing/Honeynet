WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/field-templates.soy' */
// This file was automatically generated from field-templates.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.Fields.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.Fields == 'undefined') { JIRA.Templates.Fields = {}; }


JIRA.Templates.Fields.username = function(opt_data, opt_ignored) {
  return '' + ((opt_data.escape) ? soy.$$escapeHtml(opt_data.displayName) : soy.$$filterNoAutoescape(opt_data.displayName));
};
if (goog.DEBUG) {
  JIRA.Templates.Fields.username.soyTemplateName = 'JIRA.Templates.Fields.username';
}


JIRA.Templates.Fields.recipientUsername = function(opt_data, opt_ignored) {
  return '<li data-username="' + soy.$$escapeHtml(opt_data.username) + '" title=' + soy.$$escapeHtml(opt_data.username) + '><span><span class="user-hover" rel="' + soy.$$escapeHtml(opt_data.username) + '"><img alt="" src="' + soy.$$escapeHtml(opt_data.icon) + '" title="' + JIRA.Templates.Fields.username(opt_data) + '"><span title="' + JIRA.Templates.Fields.username(opt_data) + '">' + JIRA.Templates.Fields.username(opt_data) + '</span></span>' + ((! opt_data.readOnly) ? '<button type="button" class="remove-recipient item-delete"><span class="icon-default aui-icon aui-icon-small aui-iconfont-delete">' + soy.$$escapeHtml(AJS.format('\u5220\u9664\u76d1\u89c6\u5668: {0}',opt_data.username)) + '</span></button>' : '') + '</span></li>';
};
if (goog.DEBUG) {
  JIRA.Templates.Fields.recipientUsername.soyTemplateName = 'JIRA.Templates.Fields.recipientUsername';
}


JIRA.Templates.Fields.recipientEmail = function(opt_data, opt_ignored) {
  return '<li data-email="' + soy.$$escapeHtml(opt_data.email) + '" title=' + soy.$$escapeHtml(opt_data.email) + '><span><img src="' + soy.$$escapeHtml(opt_data.icon) + '" title="' + soy.$$escapeHtml(opt_data.email) + '"><span title="' + soy.$$escapeHtml(opt_data.email) + '">' + soy.$$escapeHtml(opt_data.email) + '</span><span class="remove-recipient item-delete"><span class="icon-default aui-icon aui-icon-small aui-iconfont-delete"></span></span></span></li>';
};
if (goog.DEBUG) {
  JIRA.Templates.Fields.recipientEmail.soyTemplateName = 'JIRA.Templates.Fields.recipientEmail';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/LabelPicker.js' */
define("jira/field/label-picker",["jira/util/formatter","jira/ajs/select/multi-select","jira/ajs/list/group-descriptor","jira/ajs/list/item-descriptor","wrm/context-path","jquery"],function(e,t,i,s,a,l){var n=a();return t.extend({_getDefaultOptions:function(){return l.extend(!0,this._super(),{ajaxOptions:{url:n+"/includes/js/ajs/layer/labeldata.js",query:!0,formatResponse:this._formatResponse},removeDuplicates:!0,removeOnUnSelect:!0,userEnteredOptionsMsg:"\u65b0\u6807\u7b7e"})},isValidItem:function(e){return!/\s/.test(e)},_handleServerSuggestions:function(e){e&&e.token&&l.trim(this.$field.val())!==e.token||this._super(e)},_handleSpace:function(){""!==l.trim(this.$field.val())&&this.handleFreeInput()&&this.hideSuggestions()},keys:{Spacebar:function(e){this._handleSpace(),e.preventDefault()}},_formatResponse:function(t){var a=new i({label:"\u5efa\u8bae",type:"optgroup",styleClass:"labels-suggested"});return t&&t.suggestions&&l.each(t.suggestions,function(){a.addItem(new s({value:this.label,label:this.label,html:this.html,highlighted:!0}))}),[a]},handleFreeInput:function(){var e=l.trim(this.$field.val()).match(/\S+/g);if(e){for(var t,i=0;t=e[i];i++)this.addItem({value:t,label:t});this.model.$element.trigger("change")}this.$field.val("")}})}),AJS.namespace("AJS.LabelPicker",null,require("jira/field/label-picker")),AJS.namespace("JIRA.LabelPicker",null,require("jira/field/label-picker"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/LabelPickerFactory.js' */
define("jira/field/label-picker-factory",["jira/field/label-picker","jira/data/parse-options-from-fieldset","wrm/context-path","jquery","exports"],function(e,t,i,a,r){"use strict";r.createPicker=function(r,s){var d=t(r),l=a("#"+d.id,s),c=d.issueId,o={};/customfield_\d/.test(d.id)&&(o.customFieldId=parseInt(d.id.replace("customfield_",""),10)),new e({element:l,ajaxOptions:{url:i()+"/rest/api/1.0/labels"+(c?"/"+c:"")+"/suggest",data:o}})}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/MultiUserListPicker.Item.js' */
define("jira/field/multi-user-list-picker/item",["jira/util/data/meta","jira/util/strings","jira/ajs/control","jquery"],function(e,t,i,n){"use strict";return i.extend({init:function(e){this._setOptions(e),this.$lozenge=this._render("item"),this.$removeButton=this.$lozenge.find(".remove-recipient"),this._assignEvents("instance",this),this._assignEvents("lozenge",this.$lozenge),this._assignEvents("removeButton",this.$removeButton),this.$lozenge.prependTo(this.options.container)},_getDefaultOptions:function(){return{label:null,title:null,container:null,focusClass:"focused"}},_renders:{item:function(){var i,s=this.options.descriptor;return!0!==s.noExactMatch()?(i={escape:!1,username:s.value(),icon:s.icon(),displayName:t.escapeHtml(s.label())},n(JIRA.Templates.Fields.recipientUsername(i))):(i={email:s.value(),icon:e.get("default-avatar-url")},n(JIRA.Templates.Fields.recipientEmail(i)))}},_events:{instance:{remove:function(){this.$lozenge.remove()}},removeButton:{click:function(e){e.stopPropagation(),this.trigger("remove")}}}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/MultiUserListPicker.js' */
define("jira/field/multi-user-list-picker",["jira/util/formatter","jira/field/multi-user-list-picker/item","jira/ajs/select/suggestions/user-list-suggest-handler","jira/ajs/select/multi-select","jira/ajs/list/item-descriptor","jira/ajs/list/group-descriptor","jira/ajs/group","wrm/context-path","jquery"],function(e,t,i,r,s,n,a,l,u){"use strict";var o=l();return r.extend({init:function(r){function l(e){var t=[];return u(e).each(function(e,i){var r=new n({weight:e,label:i.footer});u(i.users).each(function(){r.addItem(new s({value:this.name,label:this.displayName,html:this.html,icon:this.avatarUrl,allowDuplicate:!1,highlighted:!0}))}),t.push(r)}),t}u.extend(r,{itemAttrDisplayed:"label",userEnteredOptionsMsg:"Email\u5730\u5740",showDropdownButton:!1,removeOnUnSelect:!0,ajaxOptions:{url:o+"/rest/api/1.0/users/picker",query:!0,data:{showAvatar:!0},formatResponse:l},suggestionsHandler:i,itemGroup:new a,itemBuilder:function(e){return new t({descriptor:e,container:this.$selectedItemsContainer})}}),this._super(r)},_createFurniture:function(e){this._super(e),this.options.description&&this._render("description",this.options.description).insertAfter(this.$field)},updateItemsIndent:u.noop,_renders:{selectedItemsWrapper:function(){return u('<div class="recipients" />')},selectedItemsContainer:function(){return u("<ol />")},description:function(e){return u("<div />").addClass("description").text(e)}}})}),AJS.namespace("JIRA.MultiUserListPicker",null,require("jira/field/multi-user-list-picker")),AJS.namespace("JIRA.MultiUserListPicker.Item",null,require("jira/field/multi-user-list-picker/item"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initCascadingSelects.js' */
define("jira/field/init-cascading-selects",["jquery","jira/util/events/reasons","jira/util/events/types","jira/util/events"],function(e,i,n,t){function a(i){e(i||document.body).find("div.aui-field-cascadingselect").add("tr.aui-field-cascadingselect").add("td.aui-field-cascadingselect").each(function(){function i(){var i,n=t.find("option:selected").attr("class");n!==a&&(i=e("<span />").insertAfter(d),d.detach(),d.find("option").each(function(e,i){i.parentNode.removeChild(this)}),d.insertAfter(i),i.remove(),c.filter("."+t.find("option:selected").attr("class")).appendTo(d),s.hasClass(t.find("option:selected").attr("class"))?d.val(s.val()):d.val(d.find("option:first").val()),a=n)}var n=e(this),t=n.find(".cascadingselect-parent"),a=(t.find("option"),""),d=n.find(".cascadingselect-child"),c=d.find("option"),s=d.find(":selected");t.bind("cascadingSelectChanged",i).change(function(){e(this).trigger("cascadingSelectChanged")}).trigger("cascadingSelectChanged")})}t.bind(n.NEW_CONTENT_ADDED,function(e,n,t){t!==i.panelRefreshed&&a(n)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/ComponentPicker.js' */
define("jira/field/component-picker",["jira/util/formatter","jira/util/events","jira/ajs/select/suggestions/only-new-items-suggest-handler","jira/ajs/select/multi-select","underscore","jquery"],function(e,t,n,s,i,r){return s.extend({init:function(e){this._super(e),this.suggestionsHandler=new n(this.options,this.model)},_getDefaultOptions:function(t){var n=!1;return t&&t.element&&(n=r(t.element).data("create-permission")),n?r.extend(!0,this._super(),{userEnteredOptionsMsg:"\u65b0\u7ec4\u4ef6"}):this._super(t)},_selectionHandler:function(e,n){var s=this.model.getDisplayableSelectedDescriptors().concat(this.model.getDisplayableUnSelectedDescriptors()),r=e.data("descriptor");i.find(s,function(e){return e.label()===r.label()})||(r.properties.value="nv_"+r.value(),t.trigger("Issue.Component.new.selected",[r.value()])),this._super(e,n)}})}),AJS.namespace("JIRA.ComponentPicker",null,require("jira/field/component-picker"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initComponentPickers.js' */
define("jira/field/init-component-pickers",["jira/field/component-picker","jira/util/formatter","jira/util/events","jira/util/events/types","jira/util/events/reasons","jquery"],function(e,n,r,t,i,o){function l(r){new e({element:r,itemAttrDisplayed:"label",errorMessage:"{0} \u4e0d\u662f\u4e00\u4e2a\u6709\u6548\u7684\u7ec4\u4ef6\u3002",maxInlineResultsDisplayed:15,expandAllResults:!0})}function a(e){var n=o(e);return n.is("select")?n:n.find("select")}function c(e,n){n=n||s.join(", "),o(n,e).each(function(){var e=a(this);e.length&&l(e)})}var s=["div.aui-field-componentspicker.frother-control-renderer","td.aui-field-componentspicker.frother-control-renderer","tr.aui-field-componentspicker.frother-control-renderer"];r.bind(t.NEW_CONTENT_ADDED,function(e,n,r){r!==i.panelRefreshed&&c(n)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initMultiselectPickers.js' */
define("jira/field/init-multi-select-pickers",["jquery","jira/ajs/select/multi-select","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/util/formatter"],function(e,t,i,n,l,r){function s(e){new t({element:e,itemAttrDisplayed:"label",errorMessage:"{0} \u662f\u65e0\u6548\u7684\u503c\u3002",maxInlineResultsDisplayed:15,submitInputVal:!0,expandAllResults:!0})}function a(t){var i=e(t);return i.is("select")?i:i.find("select")}function u(t,i){i=i||c.join(", "),e(i,t).each(function(){var e=a(this);e.length&&s(e)})}var c=["div.aui-field-multiselectpicker.frother-control-renderer"];l.bind(n.NEW_CONTENT_ADDED,function(e,t,n){n!==i.panelRefreshed&&u(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initDatePickers.js' */
define("jira/field/init-date-pickers",["jquery","jira/data/parse-options-from-fieldset","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/libs/calendar"],function(e,i,t,d,a,n){"use strict";function r(t){e(t||document.body).find("div.aui-field-datepicker").add("tr.aui-field-datepicker").add("td.aui-field-datepicker").each(function(){function d(){r.prop("disabled",s.is(":checked"))}var a=e(this),r=a.find("input:text"),s=a.find("#useCurrentDate"),c=i(a.find("fieldset.datepicker-params"));c.context=t,n.setup(c),s.length&&(d(),s.click(d))})}a.bind(d.NEW_CONTENT_ADDED,function(e,i,d){d!==t.panelRefreshed&&r(i)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initLegacyGroupPickers.js' */
define("jira/field/init-legacy-group-pickers",["jquery","jira/util/events/reasons","jira/util/events/types","jira/util/events"],function(e,i,r,t){function n(i){e(i||document.body).find("div.aui-field-grouppicker").add("tr.aui-field-grouppicker").add("td.aui-field-grouppicker").each(function(){function i(e){e.preventDefault(),window.open(n,"GroupPicker","status=yes,resizable=yes,top=100,left=100,width=800,height=750,scrollbars=yes")}var r=e(this),t=r.find("a.grouppicker-trigger"),n=t.attr("href");t.click(i)})}t.bind(r.NEW_CONTENT_ADDED,function(e,r,t){t!==i.panelRefreshed&&n(r)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initIssuePickers.js' */
define("jira/field/init-issue-pickers",["jira/field/issue-picker","jira/util/events","jira/util/events/types","jira/util/events/reasons","jquery","jira/util/formatter"],function(e,i,n,t,s,r){function u(i){s(i||document.body).find(".aui-field-issuepicker").each(function(){new e({element:s(this),userEnteredOptionsMsg:"\u8f93\u5165\u95ee\u9898\u5173\u952e\u5b57\u3002",uppercaseUserEnteredOnSelect:!0})})}i.bind(n.NEW_CONTENT_ADDED,function(e,i,n){n!==t.panelRefreshed&&u(i)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initLabelPickers.js' */
define("jira/field/init-label-pickers",["jira/field/label-picker-factory","jira/util/events","jira/util/events/types","jira/util/events/reasons","jquery"],function(e,i,n,r,a){function t(e){var i=a(e);return i.is(c)?i:i.find(c)}function l(i,n){n=n||".aui-field-labelpicker",a(n,i).each(function(){var n=t(this);n.length>0&&e.createPicker(n,i)})}var c="fieldset.labelpicker-params";i.bind(n.NEW_CONTENT_ADDED,function(e,i,n){n!==r.panelRefreshed&&l(i)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/VersionPicker.js' */
define("jira/field/version-picker",["jira/util/formatter","jira/util/events","jira/ajs/select/suggestions/only-new-items-suggest-handler","jira/ajs/select/multi-select","underscore","jquery"],function(e,t,s,i,n,r){return i.extend({init:function(e){this._super(e),this.suggestionsHandler=new s(this.options,this.model)},_getDefaultOptions:function(t){var s=!1;return t&&t.element&&(s=r(t.element).data("create-permission")),s?r.extend(!0,this._super(),{userEnteredOptionsMsg:"\u65b0\u7248\u672c"}):this._super(t)},_selectionHandler:function(e,s){var i=this.model.getDisplayableSelectedDescriptors().concat(this.model.getDisplayableUnSelectedDescriptors()),r=e.data("descriptor");n.find(i,function(e){return e.label()===r.label()})||(r.properties.value="nv_"+r.value(),t.trigger("Issue.Version.new.selected",[r.value()])),this._super(e,s)}})}),AJS.namespace("JIRA.VersionPicker",null,require("jira/field/version-picker"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initVersionPickers.js' */
define("jira/field/init-version-pickers",["jira/field/version-picker","jira/util/formatter","jira/util/events","jira/util/events/types","jira/util/events/reasons","jquery"],function(e,i,n,r,t,s){function l(n){new e({element:n,itemAttrDisplayed:"label",removeOnUnSelect:!1,submitInputVal:!0,errorMessage:"{0} \u4e0d\u662f\u4e00\u4e2a\u6709\u6548\u7684\u7248\u672c\u3002",maxInlineResultsDisplayed:15,expandAllResults:!0})}function a(e){var i=s(e);return i.is("select")?i:i.find("select")}function o(e,i){i=i||".aui-field-versionspicker.frother-control-renderer",s(i,e).each(function(){var e=a(this);e.length&&l(e)})}n.bind(r.NEW_CONTENT_ADDED,function(e,i,n){n!==t.panelRefreshed&&o(i)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initCommentControls.js' */
define("jira/field/init-comment-controls",["jira/skate","jira/ajs/select/security-level-select","jira/wikipreview/wiki-preview"],function(e,i,t){e("jira-wikifield",{type:e.type.CLASSNAME,created:function(e){var i={fieldId:e.getAttribute("field-id"),trigger:e.querySelector(".wiki-preview").id,issueKey:e.getAttribute("issue-key"),rendererType:e.getAttribute("renderer-type")};t(i,e).init()}}),e("security-level",{type:e.type.CLASSNAME,created:function(e){var t=e.querySelector("#commentLevel");t&&new i(t)}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initLogWorkControls.js' */
define("jira/field/init-log-work-controls",["jira/util/events","jira/util/events/types","jira/util/events/reasons","jquery"],function(e,t,a,i){function n(e,t){var a=i(e).find("#worklog-logworkcontainer"),n=i(e).find("#worklog-timetrackingcontainer"),d=i(e).find("#log-work-activate");t?(a.removeClass("hidden"),n.addClass("hidden"),d.prop("checked",!0)):(a.addClass("hidden"),n.removeClass("hidden"),d.prop("checked",!1))}function d(e){i("#log-work-adjust-estimate-new-value, #log-work-adjust-estimate-manual-value",e).attr("disabled","disabled"),i("#log-work-adjust-estimate-"+i("input[name=worklog_adjustEstimate]:checked,input[name=adjustEstimate]:checked",e).val()+"-value",e).removeAttr("disabled"),i("input[name=worklog_adjustEstimate],input[name=adjustEstimate]",e).change(function(){i("#log-work-adjust-estimate-new-value,#log-work-adjust-estimate-manual-value",e).attr("disabled","disabled"),i("#log-work-adjust-estimate-"+i(this).val()+"-value",e).removeAttr("disabled")}),i("#delete-log-work-adjust-estimate-new-value").change(function(){i("#delete-log-work-adjust-estimate-new").prop("checked",!0)}),i("#delete-log-work-adjust-estimate-manual-value").change(function(){i("#delete-log-work-adjust-estimate-manual").prop("checked",!0)}),i(e).find("#log-work-activate").change(function(){n(e,i(this).is(":checked"))})}e.bind(t.VALIDATE_TIMETRACKING,function(e,t){n(t,!0)}),e.bind(t.NEW_CONTENT_ADDED,function(e,t,i){i!==a.panelRefreshed&&d(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initPriorityPickers.js' */
define("jira/field/init-priority-pickers",["jira/ajs/select/single-select","jira/util/events/reasons","jira/util/events/types","jira/util/events"],function(e,i,n,t){function r(i){i.find("select#priority").each(function(i,n){new e({element:n,revertOnInvalid:!0})})}t.bind(n.NEW_CONTENT_ADDED,function(e,n,t){t!==i.panelRefreshed&&r(n)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/initInlineAttach.js' */
define("jira/field/init-inline-attach",["jira/util/events","jira/util/events/types","jira/util/events/reasons","jquery","jira/jquery/plugins/attachment/inline-attach"],function(i,n,t){"use strict";function e(i){i.find("input[type=file]:not('.ignore-inline-attach')").inlineAttach()}i.bind(n.NEW_CONTENT_ADDED,function(i,n,a){a!==t.panelRefreshed&&e(n)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-fields', location = '/includes/jira/field/init/init-jira-fields-webresource.js' */
require("jira/field/init-cascading-selects"),require("jira/field/init-component-pickers"),require("jira/field/init-multi-select-pickers"),require("jira/field/init-date-pickers"),require("jira/field/init-legacy-group-pickers"),require("jira/field/init-issue-pickers"),require("jira/field/init-label-pickers"),require("jira/field/init-version-pickers"),require("jira/field/init-comment-controls"),require("jira/field/init-log-work-controls"),require("jira/field/init-priority-pickers"),require("jira/field/init-inline-attach");
}catch(e){WRMCB(e)};