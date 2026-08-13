WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'jira.webresources:calendar-zh', location = '/includes/lib/calendar/lang/calendar-zh.js' */
// ** I18N

// Calendar ZH language
// Author: muziq, <muziq@sina.com>
// Encoding: GB2312 or GBK
// Distributed under the same terms as the calendar itself.

// full day names
Calendar._DN = new Array
("\u661f\u671f\u65e5",
 "\u661f\u671f\u4e00",
 "\u661f\u671f\u4e8c",
 "\u661f\u671f\u4e09",
 "\u661f\u671f\u56db",
 "\u661f\u671f\u4e94",
 "\u661f\u671f\u516d",
 "\u661f\u671f\u65e5");

// Please note that the following array of short day names (and the same goes
// for short month names, _SMN) isn't absolutely necessary.  We give it here
// for exemplification on how one can customize the short day names, but if
// they are simply the first N letters of the full name you can simply say:
//
//   Calendar._SDN_len = N; // short day name length
//   Calendar._SMN_len = N; // short month name length
//
// If N = 3 then this is not needed either since we assume a value of 3 if not
// present, to be compatible with translation files that were written before
// this feature.

// short day names
Calendar._SDN = new Array
("\u65e5",
 "\u4e00",
 "\u4e8c",
 "\u4e09",
 "\u56db",
 "\u4e94",
 "\u516d",
 "\u65e5");

// full month names
Calendar._MN = new Array
("\u4e00\u6708",
 "\u4e8c\u6708",
 "\u4e09\u6708",
 "\u56db\u6708",
 "\u4e94\u6708",
 "\u516d\u6708",
 "\u4e03\u6708",
 "\u516b\u6708",
 "\u4e5d\u6708",
 "\u5341\u6708",
 "\u5341\u4e00\u6708",
 "\u5341\u4e8c\u6708");

// short month names
Calendar._SMN = new Array
("\u4e00\u6708",
 "\u4e8c\u6708",
 "\u4e09\u6708",
 "\u56db\u6708",
 "\u4e94\u6708",
 "\u516d\u6708",
 "\u4e03\u6708",
 "\u516b\u6708",
 "\u4e5d\u6708",
 "\u5341\u6708",
 "\u5341\u4e00\u6708",
 "\u5341\u4e8c\u6708");

// tooltips
Calendar._TT = {};
Calendar._TT["INFO"] = "\u5e2e\u52a9";

Calendar._TT["ABOUT"] =
"DHTML Date/Time Selector\n" +
"(c) dynarch.com 2002-2005 / Author: Mihai Bazon\n" + // don't translate this this ;-)
"For latest version visit: http://www.dynarch.com/projects/calendar/\n" +
"Distributed under GNU LGPL.  See http://gnu.org/licenses/lgpl.html for details." +
"\n\n" +
"\u9009\u62e9\u65e5\u671f:\n" +
"- \u70b9\u51fb \xab, \xbb \u6309\u94ae\u9009\u62e9\u5e74\u4efd\n" +
"- \u70b9\u51fb " + String.fromCharCode(0x2039) + ", " + String.fromCharCode(0x203a) + " \u6309\u94ae\u9009\u62e9\u6708\u4efd\n" +
"- \u957f\u6309\u4ee5\u4e0a\u6309\u94ae\u53ef\u4ece\u83dc\u5355\u4e2d\u5feb\u901f\u9009\u62e9\u5e74\u4efd\u6216\u6708\u4efd";
Calendar._TT["ABOUT_TIME"] = "\n\n" +
"\u9009\u62e9\u65f6\u95f4:\n" +
"- \u70b9\u51fb\u5c0f\u65f6\u6216\u5206\u949f\u53ef\u4f7f\u6539\u6570\u503c\u52a0\u4e00\n" +
"- \u6309\u4f4fShift\u952e\u70b9\u51fb\u5c0f\u65f6\u6216\u5206\u949f\u53ef\u4f7f\u6539\u6570\u503c\u51cf\u4e00\n" +
"- \u70b9\u51fb\u62d6\u52a8\u9f20\u6807\u53ef\u8fdb\u884c\u5feb\u901f\u9009\u62e9";

Calendar._TT["PREV_YEAR"] = "\u4e0a\u4e00\u5e74 (\u6309\u4f4f\u51fa\u83dc\u5355)";
Calendar._TT["PREV_MONTH"] = "\u4e0a\u4e00\u6708 (\u6309\u4f4f\u51fa\u83dc\u5355)";
Calendar._TT["GO_TODAY"] = "\u8f6c\u5230\u4eca\u65e5";
Calendar._TT["NEXT_MONTH"] = "\u4e0b\u4e00\u6708 (\u6309\u4f4f\u51fa\u83dc\u5355)";
Calendar._TT["NEXT_YEAR"] = "\u4e0b\u4e00\u5e74 (\u6309\u4f4f\u51fa\u83dc\u5355)";
Calendar._TT["SEL_DATE"] = "\u9009\u62e9\u65e5\u671f";
Calendar._TT["DRAG_TO_MOVE"] = "\u62d6\u52a8";
Calendar._TT["PART_TODAY"] = " (\u4eca\u65e5)";

// the following is to inform that "%s" is to be the first day of week
// %s will be replaced with the day name.
Calendar._TT["DAY_FIRST"] = "\u6700\u5de6\u8fb9\u663e\u793a%s";

// This may be locale-dependent.  It specifies the week-end days, as an array
// of comma-separated numbers.  The numbers are from 0 to 6: 0 means Sunday, 1
// means Monday, etc.
Calendar._TT["WEEKEND"] = "0,6";

Calendar._TT["CLOSE"] = "\u5173\u95ed";
Calendar._TT["TODAY"] = "\u4eca\u65e5";
Calendar._TT["TIME_PART"] = "(Shift-)\u70b9\u51fb\u9f20\u6807\u6216\u62d6\u52a8\u6539\u53d8\u503c";

// date formats
Calendar._TT["DEF_DATE_FORMAT"] = "%Y-%m-%d";
Calendar._TT["TT_DATE_FORMAT"] = "%A, %b %e\u65e5";

Calendar._TT["WK"] = "\u5468";
Calendar._TT["TIME"] = "\u65f6\u95f4:";

Calendar._TT["AM"] = "\u4e0a\u5348";
Calendar._TT["PM"] = "\u4e0b\u5348";
Calendar._TT["am"] = "\u4e0a\u5348";
Calendar._TT["pm"] = "\u4e0b\u5348";

}catch(e){WRMCB(e)};