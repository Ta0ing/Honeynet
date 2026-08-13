WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:bigpipe-init', location = '/includes/jira/common/initBigpipe.js' */
!function(){var e=require("jira/skate");require("jira/bigpipe/element");for(var i=document.querySelectorAll("big-pipe"),r=0,t=i.length;r<t;r++)e.init(i[r])}();
}catch(e){WRMCB(e)};