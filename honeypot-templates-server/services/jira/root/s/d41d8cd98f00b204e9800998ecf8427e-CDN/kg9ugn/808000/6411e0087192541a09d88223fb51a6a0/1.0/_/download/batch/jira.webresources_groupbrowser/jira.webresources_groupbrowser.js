WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:groupbrowser', location = '/includes/jira/admin/group-browser/group-browser.js' */
require(["jquery","jira/admin/group-browser/group-label-lozenge"],function(i){i(function(){i(".operations-list .aui-button[disabled]").tooltip({gravity:"e",html:!1})})});
}catch(e){WRMCB(e)};