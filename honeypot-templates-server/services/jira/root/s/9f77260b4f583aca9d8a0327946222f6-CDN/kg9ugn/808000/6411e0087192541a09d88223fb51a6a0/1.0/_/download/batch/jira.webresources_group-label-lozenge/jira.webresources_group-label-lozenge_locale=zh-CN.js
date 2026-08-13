WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:group-label-lozenge', location = '/includes/jira/admin/group-browser/group-label-lozenge.js' */
define("jira/admin/group-browser/group-label-lozenge",["jquery","jira/skate"],function(e,t){t("group-label-lozenge",{type:t.type.CLASSNAME,attached:function(t){e(t).tooltip({gravity:"w",html:!0})}})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:group-label-lozenge', location = '/includes/jira/admin/group-browser/group-label-lozenge.soy' */
// This file was automatically generated from group-label-lozenge.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }


JIRA.Templates.groupLabelLozenge = function(opt_data, opt_ignored) {
  return '' + aui.lozenges.lozenge({text: opt_data.label.text, title: opt_data.label.title, isSubtle: true, extraClasses: 'group-label-lozenge'});
};
if (goog.DEBUG) {
  JIRA.Templates.groupLabelLozenge.soyTemplateName = 'JIRA.Templates.groupLabelLozenge';
}

}catch(e){WRMCB(e)};