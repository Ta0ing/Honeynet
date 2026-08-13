WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:calendar', location = '/includes/lib/calendar/Calendar-layerable.js' */
define("jira/libs/calendar-layerable-mixin",["jira/ajs/layer/layer-interactions","jira/ajs/layer/inline-layer","jira/util/events"],function(e,n,i){"use strict";function r(r){e.preventDialogHide(r),e.preventInputBlur(r),e.preventInlineEditCancel(r),i.bind(n.EVENTS.beforeHide,function(e){r.current&&e.preventDefault()})}return{layerableExtensions:r}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'jira.webresources:calendar', location = '/includes/lib/calendar/Calendar-amd.js' */
define("jira/libs/calendar",["require"],function(a){"use strict";var e=window.Calendar;return a("jira/libs/calendar-layerable-mixin").layerableExtensions(e),e}),function(){"use strict";var a=require("jira/libs/calendar");AJS.namespace("window.Calendar",window,a)}();
}catch(e){WRMCB(e)};