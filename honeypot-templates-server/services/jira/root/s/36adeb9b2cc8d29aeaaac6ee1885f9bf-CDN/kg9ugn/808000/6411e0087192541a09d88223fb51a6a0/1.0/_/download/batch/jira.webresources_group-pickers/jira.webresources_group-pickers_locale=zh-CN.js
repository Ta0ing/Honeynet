WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:group-pickers', location = '/includes/jira/field/groupPickerUtil.js' */
define("jira/field/group-picker-util",["jira/ajs/list/item-descriptor","jira/ajs/list/group-descriptor","jquery"],function(e,i,r){var t=function(t,a){var n=[],s=a?JIRA.Templates.GroupPickerUtil.formatResponseWithLabels:JIRA.Templates.GroupPickerUtil.formatResponse;return r(t).each(function(t,a){var u=new i({weight:t,label:a.header});r(a.groups).each(function(){u.addItem(new e({value:this.name,label:this.name,title:this.name,html:s(this),highlighted:!0}))}),n.push(u)}),n};return{formatResponseWithLabels:function(e){return t(e,!0)},formatResponse:function(e,i){return t(e,i)}}}),AJS.namespace("JIRA.GroupPickerUtil",null,require("jira/field/group-picker-util"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:group-pickers', location = '/includes/jira/field/templates/groupPickerUtil.soy' */
// This file was automatically generated from groupPickerUtil.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.GroupPickerUtil.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.GroupPickerUtil == 'undefined') { JIRA.Templates.GroupPickerUtil = {}; }


JIRA.Templates.GroupPickerUtil.formatResponseWithLabels = function(opt_data, opt_ignored) {
  return '' + JIRA.Templates.GroupPickerUtil.formatResponse(soy.$$augmentMap(opt_data, {showLabels: true}));
};
if (goog.DEBUG) {
  JIRA.Templates.GroupPickerUtil.formatResponseWithLabels.soyTemplateName = 'JIRA.Templates.GroupPickerUtil.formatResponseWithLabels';
}


JIRA.Templates.GroupPickerUtil.formatResponse = function(opt_data, opt_ignored) {
  var output = '<div class="group-suggestion-item"><span class="group-suggestion-item__name">' + soy.$$filterNoAutoescape(opt_data.html) + '</span>';
  if (opt_data.showLabels) {
    output += '<span class="group-suggestion-item__labels group-labels-lozenges">';
    var labelList13 = opt_data.labels;
    var labelListLen13 = labelList13.length;
    for (var labelIndex13 = 0; labelIndex13 < labelListLen13; labelIndex13++) {
      var labelData13 = labelList13[labelIndex13];
      output += JIRA.Templates.groupLabelLozenge({label: labelData13}) + ' ';
    }
    output += '</span>';
  }
  output += '</div>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.GroupPickerUtil.formatResponse.soyTemplateName = 'JIRA.Templates.GroupPickerUtil.formatResponse';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:group-pickers', location = '/includes/jira/field/initMultiGroupPickers.js' */
define("jira/field/init-multi-group-pickers",["jquery","wrm/context-path","jira/ajs/select/multi-select","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/field/group-picker-util"],function(e,t,i,r,a,s,n){"use strict";function u(t){t.find(".js-default-multi-group-picker").each(function(){var t=e(this),r=!0===t.data("show-labels"),a=t.data("user-name");new i({element:this,itemAttrDisplayed:"label",showDropdownButton:!1,ajaxOptions:{data:function(e){return{userName:a,query:e,exclude:t.val()}},url:l+"/rest/api/2/groups/picker",query:!0,formatResponse:r?n.formatResponseWithLabels:n.formatResponse}})})}var l=t();s.bind(a.NEW_CONTENT_ADDED,function(e,t,i){i!==r.panelRefreshed&&u(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:group-pickers', location = '/includes/jira/field/initSingleGroupPickers.js' */
define("jira/field/init-single-group-pickers",["jira/ajs/select/single-select","jquery","wrm/context-path","jira/ajs/list/item-descriptor","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/field/group-picker-util"],function(e,t,i,a,n,r,s,l){"use strict";function u(i){i.find(".js-default-single-group-picker").each(function(){var i=t(this),n=i.find("option[data-empty]"),r=!0===i.data("show-labels"),s=i.data("user-name");new e({element:this,itemAttrDisplayed:"label",revertOnInvalid:!0,ajaxOptions:{data:function(e){return{userName:s,query:e,exclude:i.val()}},url:o+"/rest/api/2/groups/picker",query:!0,formatResponse:function(e){var t=l.formatResponse(e,r);return n.length&&""!==i.val()&&t.unshift(new a({value:"",label:n.text(),highlighted:!0})),t}}})})}var o=i();s.bind(r.NEW_CONTENT_ADDED,function(e,t,i){i!==n.panelRefreshed&&u(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:group-pickers', location = '/includes/jira/field/init/init-group-pickers-webresource.js' */
require("jira/field/init-multi-group-pickers"),require("jira/field/init-single-group-pickers");
}catch(e){WRMCB(e)};