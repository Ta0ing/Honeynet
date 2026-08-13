WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:calendar-localisation-moment', location = '/includes/lib/calendar/Calendar-localisation-moment.js' */
define("jira/calendar/localisation-moment",["require"],function(a){"use strict";var e=a("jira/moment"),o=a("jira/libs/calendar"),t=e.localeData("jira");null!==t&&"object"==typeof t&&"function"==typeof o&&(o._DN=t._weekdays.concat(t._weekdays[0]),o._SDN=t._weekdaysShort.concat(t._weekdaysShort[0]),o._MN=[].concat(t._months),o._SMN=[].concat(t._monthsShort))}),require("jira/calendar/localisation-moment");
}catch(e){WRMCB(e)};