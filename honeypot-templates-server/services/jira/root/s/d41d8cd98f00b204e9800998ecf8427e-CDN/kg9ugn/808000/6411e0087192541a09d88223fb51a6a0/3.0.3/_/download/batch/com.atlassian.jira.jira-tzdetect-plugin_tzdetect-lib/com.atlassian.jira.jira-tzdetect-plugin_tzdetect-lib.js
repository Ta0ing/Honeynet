WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'com.atlassian.jira.jira-tzdetect-plugin:tzdetect-lib', location = 'banner/tzdetect-banner-init.js' */
require(["jira-tzdetect/banner","wrm/data","jquery"],function(A,B,C){C(function(){if(B.claim("tzdetect.pref.auto.detect")==true){A.init()}})});
}catch(e){WRMCB(e)};