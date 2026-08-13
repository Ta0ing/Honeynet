WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:jira-project-issuetype-fields', location = '/includes/jira/field/ProjectIssueTypeSelect.js' */
define("jira/field/project-issue-type-select",["jira/lib/class"],function(e){return e.extend({init:function(e){var s,t=this;this.$project=e.project,this.issueTypeSelect=e.issueTypeSelect,this.$projectIssueTypesSchemes=e.projectIssueTypesSchemes,this.$issueTypeSchemeIssueDefaults=e.issueTypeSchemeIssueDefaults,this.projectIssueTypeSchemes=JSON.parse(this.$projectIssueTypesSchemes.html()),this.issueTypesSchemeDefaults=JSON.parse(this.$issueTypeSchemeIssueDefaults.html()||"[]"),t.$project.length>0&&(s=t.$project.val(),t.setIssueTypeScheme(t.getIssueTypeSchemeForProject(s)),this.$project.change(function(){var e=t.$project.val();t.setIssueTypeScheme(t.getIssueTypeSchemeForProject(e))}))},getIssueTypeSchemeForProject:function(e){return this.projectIssueTypeSchemes[e]},getDefaultIssueTypeForScheme:function(e){return this.issueTypesSchemeDefaults[e]},setIssueTypeScheme:function(e){var s=this.issueTypeSelect.model.getValue();this.issueTypeSelect.model.setFilterGroup(e),this.issueTypeSelect.model.setSelected(s,!1)||this.setDefaultIssueType(e),this.issueTypeSelect.model.$element.data("project",this.$project.val())},setDefaultIssueType:function(e){var s=this.getDefaultIssueTypeForScheme(e),t=this.issueTypeSelect.model.getDescriptor(s);t||(t=this.issueTypeSelect.model.getAllDescriptors()[0]),this.issueTypeSelect.setSelection(t,!1)}})}),AJS.namespace("JIRA.ProjectIssueTypeSelect",null,require("jira/field/project-issue-type-select"));
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-project-issuetype-fields', location = '/includes/jira/field/initProjectIssueTypeSelect.js' */
define("jira/field/init-project-issue-type-select",["jira/ajs/select/suggestion-collection-model","jquery","jira/ajs/select/single-select","jira/util/events/reasons","jira/util/events/types","jira/util/events","jira/field/project-issue-type-select"],function(e,t,i,s,n,l,r){function c(s){var n=s||t("body");n.find(".issuetype-field").each(function(s){var l=n.find(".project-field, .project-field-readonly"),c=t(this),d=n.find("#"+c.attr("id")+"-projects"),o=n.find("#"+c.attr("id")+"-defaults"),a=new i({element:c,revertOnInvalid:!0,model:e});a.model.remove(""),l.length>0&&new r({project:l.eq(s),issueTypeSelect:a,projectIssueTypesSchemes:d,issueTypeSchemeIssueDefaults:o})})}l.bind(n.NEW_CONTENT_ADDED,function(e,t,i){i!==s.panelRefreshed&&c(t)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-project-issuetype-fields', location = '/includes/jira/field/initProjectPickers.js' */
define("jira/field/init-project-pickers",["jira/ajs/select/scrollable-single-select","jira/ajs/select/suggestion-collection-model","jira/util/events/reasons","jira/util/events/types","jira/util/events"],function(e,i,n,t,l){function s(n){n.find(".project-field").each(function(){new e({element:this,revertOnInvalid:!0,pageSize:50,pagingThreshold:100,model:i})})}l.bind(t.NEW_CONTENT_ADDED,function(e,i,t){t!==n.panelRefreshed&&s(i)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:jira-project-issuetype-fields', location = '/includes/jira/field/init/init-jira-project-issuetype-fields-webresource.js' */
require("jira/field/init-project-issue-type-select"),require("jira/field/init-project-pickers");
}catch(e){WRMCB(e)};