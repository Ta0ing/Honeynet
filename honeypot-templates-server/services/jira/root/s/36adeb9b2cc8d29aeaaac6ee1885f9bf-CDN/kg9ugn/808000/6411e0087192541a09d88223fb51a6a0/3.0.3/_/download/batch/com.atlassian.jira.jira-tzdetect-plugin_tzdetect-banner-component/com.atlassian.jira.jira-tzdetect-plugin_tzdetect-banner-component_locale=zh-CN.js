WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'com.atlassian.jira.jira-tzdetect-plugin:tzdetect-banner-component', location = 'banner/tzdetect-banner.soy' */
// This file was automatically generated from tzdetect-banner.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace JIRA.Templates.tzdetect.soy.
 */

if (typeof JIRA == 'undefined') { var JIRA = {}; }
if (typeof JIRA.Templates == 'undefined') { JIRA.Templates = {}; }
if (typeof JIRA.Templates.tzdetect == 'undefined') { JIRA.Templates.tzdetect = {}; }
if (typeof JIRA.Templates.tzdetect.soy == 'undefined') { JIRA.Templates.tzdetect.soy = {}; }


JIRA.Templates.tzdetect.soy.banner = function(opt_data, opt_ignored) {
  return soy.$$escapeHtml(opt_data.message) + '<p><a class="tz-yes" href="' + soy.$$escapeHtml(opt_data.updateLink) + '">' + soy.$$escapeHtml(opt_data.updateLinkText) + '</a></p>';
};
if (goog.DEBUG) {
  JIRA.Templates.tzdetect.soy.banner.soyTemplateName = 'JIRA.Templates.tzdetect.soy.banner';
}


JIRA.Templates.tzdetect.soy.dropdown = function(opt_data, opt_ignored) {
  var output = '<div class="aui-dropdown2 aui-style-default"' + ((opt_data.id) ? ' id="' + soy.$$escapeHtml(opt_data.id) + '"' : '') + '>';
  var sectionList17 = opt_data.sections;
  var sectionListLen17 = sectionList17.length;
  for (var sectionIndex17 = 0; sectionIndex17 < sectionListLen17; sectionIndex17++) {
    var sectionData17 = sectionList17[sectionIndex17];
    output += '<div class="aui-dropdown2-section">';
    if (sectionData17.region) {
      output += '<strong>' + soy.$$escapeHtml(sectionData17.region) + '</strong><ul class="aui-list-truncate">';
      var zoneList24 = sectionData17.zones;
      var zoneListLen24 = zoneList24.length;
      for (var zoneIndex24 = 0; zoneIndex24 < zoneListLen24; zoneIndex24++) {
        var zoneData24 = zoneList24[zoneIndex24];
        output += '<li><a href="' + soy.$$escapeHtml(opt_data.baseHref) + '#zone-' + soy.$$escapeHtml(zoneData24.timeZoneId) + '" data-tzid="' + soy.$$escapeHtml(zoneData24.timeZoneId) + '">' + soy.$$escapeHtml(zoneData24.gmtOffset) + ' ' + soy.$$escapeHtml(zoneData24.city) + '</a></li>';
      }
      output += '</ul>';
    } else if (sectionData17.other) {
      output += '<ul class="aui-list-truncate"><li><a href="' + soy.$$escapeHtml(opt_data.baseHref) + '#zone-other" data-tzother="true">' + soy.$$escapeHtml(sectionData17.other) + '</a></li></ul>';
    }
    output += '</div>';
  }
  output += '</div>';
  return output;
};
if (goog.DEBUG) {
  JIRA.Templates.tzdetect.soy.dropdown.soyTemplateName = 'JIRA.Templates.tzdetect.soy.dropdown';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-tzdetect-plugin:tzdetect-banner-component', location = 'banner/tzdetect-banner-templates.js' */
define("jira-tzdetect/banner/templates",function(){return JIRA.Templates.tzdetect.soy});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.jira.jira-tzdetect-plugin:tzdetect-banner-component', location = 'banner/tzdetect-banner.js' */
define("jira-tzdetect/banner",["jira-tzdetect/banner/templates","jira/util/formatter","jira/ajs/ajax/smart-ajax","jira/analytics","jira/flag","wrm/context-path","wrm/data","jira/jquery/deferred","jquery","underscore"],function(O,M,A,G,H,F,L,I,E,N){function D(P){return(+P||0)*60*1000}function C(P){return Math.floor((+P||0)/1000/60)}function B(P){return function Q(R){if(R.which===1&&!R.shiftKey&&!R.ctrlKey&&!R.metaKey){P.apply(this,arguments)}}}var K=F()+"/rest/tzdetect/1";var J={$tzFlag:null,$link:null,$dropdown:null,prefs:{},detected:{}};J.init=function(){if(J.$tzFlag){return }J.detectTimezoneOffsets();if(J.detected.janOffset!=null&&J.detected.julyOffset!=null){var P=J.detected.janOffset!=J.prefs.janOffset||J.detected.julyOffset!=J.prefs.julyOffset;if(P){J.create()}}};J.detectTimezoneOffsets=function(){J.prefs=J.getPreferences();if(!J.prefs.tzid){return }J.detected.janOffset=J.getTzOffset(0);J.detected.julyOffset=J.getTzOffset(6)};J.create=function(){var T=J.profileLink();var U=M.format("\u60a8\u7684\u8ba1\u7b97\u673a\u65f6\u533a\u4f3c\u4e4e\u5e76\u4e0d\u5339\u914d\u60a8\u7684Jira\u65f6\u533a\u9996\u9009\u9879{0}\u3002",J.prefs.tzname);var Q="\u66f4\u65b0\u60a8\u7684Jira\u9996\u9009\u9879";var S=O.banner({message:U,updateLinkText:Q,updateLink:T+"#zone-set"});var R=[D(J.detected.janOffset),D(J.detected.julyOffset)].join(",");var P=H.showInfoMsg(null,S,{dismissalKey:"com.atlassian.jira.tzdetect."+R});if(P){J.$tzFlag=E(P);J.getPotentialZones();J.$tzFlag.on("aui-flag-close",function(){J.track("clicked.nothanks");J.$tzFlag=undefined});J.$link=J.$tzFlag.find(".tz-yes");J.$link.click(B(function(V){V.preventDefault();J.getPotentialZones().then(J.handleZoneData)}));J.track("shown")}};J.track=function(R,P){var Q={name:"tzdetect.banner."+R};if(P){Q.data=P}G.send(Q)};J.profileLink=function(){return F()+"/secure/ViewProfile.jspa"};J.redirect=function(P){window.location=P};J.getPreferences=function(){var T="tzdetect.pref.";var S=["tzid","tzname","janOffset","julyOffset"];var P={};var R=S.length;var Q,U;while(R--){Q=S[R];U=L.claim(T+Q);if(Q.indexOf("Offset")>-1){U=C(U)}P[Q]=U}return P};J.handleZoneData=function(S){var Q=J.filterZoneData(S.zones);var P=Q.length;Q.regions={};var R=S.regions.length,T;while(R--){T=S.regions[R];Q.regions[T.key]=T.displayName}Q.sort(function(W,V){var Y=W.regionKey;var X=V.regionKey;if(Y==X){return W.city<V.city?-1:1}return Y<X?-1:1});J.track("clicked.update",{matchingZoneCount:P});if(!P){J.redirect(J.$link.attr("href"));return }if(P===1){J.setUserTimeZone(Q[0].timeZoneId,"banner");return }if(!J.$dropdown){var U="timezone-selection-list";J.$link.addClass("aui-dropdown2-trigger").attr("aria-owns",U).attr("aria-haspopup","true");J.$dropdown=E(O.dropdown({id:U,sections:J.getListSections(Q),baseHref:J.profileLink()})).css("z-index",5000).on("click","a",B(function(X){var W=E(this);var Y=W.attr("data-tzid");if(Y){X.preventDefault();J.setUserTimeZone(Y,"menu")}else{if(W.attr("data-tzother")){var V={offsets:[J.detected.janOffset,J.detected.julyOffset].join(",")};J.track("menu.other",V)}}})).appendTo("body")}J.$dropdown.trigger("aui-button-invoke")};J.filterZoneData=function(P){var Q=["Antarctica","Etc"];return N.filter(P,function(R){return !N.contains(Q,R.regionKey)})};J.getListSections=function(P){var V=[];var T,S,U;for(var Q=0,R=P.length;Q<R;Q++){T=P[Q].regionKey;if(T!==S){U={region:P.regions[T]||T,zones:[]};V.push(U);S=T}U.zones.push(P[Q])}V.push({other:"\u5176\u5b83......"});return V};J.getTzOffset=function(R){var Q=new Date().getFullYear();var P=new Date(Q,R,1,12,0,0);return -P.getTimezoneOffset()};J.getPotentialZones=function(){var P=new I();if(J.zoneList){P.resolve(J.zoneList)}else{var Q={janOffset:D(J.detected.janOffset),julyOffset:D(J.detected.julyOffset)};A.makeRequest({url:K+"/zones",type:"GET",data:Q,contentType:"application/json",complete:function(T,R,S){if(R!="abort"&&S.successful){J.zoneList=S.data;P.resolve(J.zoneList)}else{P.reject(S)}}})}return P.promise()};J.setUserTimeZone=function(Q,P){J.track("setzone",{zoneId:Q,source:P});A.makeRequest({url:K+"/update",type:"POST",data:Q,contentType:"application/json",complete:function(V,R,S){if(R!="abort"&&S.successful){if(J.$tzFlag){J.$tzFlag.find(".aui-icon.icon-close").click()}var U=S.data;var T=M.format("\u60a8\u9ed8\u8ba4\u7684\u65f6\u533a\u5df2\u88ab\u8bbe\u7f6e\u4e3a{0}\u3002",U.gmtOffset+" "+U.city);H.showSuccessMsg("",T,{close:"manual"})}}})};return J});
}catch(e){WRMCB(e)};