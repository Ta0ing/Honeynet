WRMCB=function(e){var c=console;if(c&&c.log&&c.error){c.log('Error running batched script.');c.error(e);}}
;
try {
/* module-key = 'confluence.extra.jira:help-dialog-extension', location = '/jira/help-dialog.js' */
if(Confluence.KeyboardShortcuts){Confluence.KeyboardShortcuts.Editor.push({context:"editor.actions",descKey:"Inserisci Issue/Filtro Jira"+":",keys:[["Ctrl+Shift+J"]]})};
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.confluence.plugins.confluence-mentions-plugin:help-dialog-extension', location = 'js/help-dialog.js' */
AJS.toInit(function($) {
    // Add the shortcut to the shortcuts dialog
    Confluence.KeyboardShortcuts && Confluence.KeyboardShortcuts.Autoformat.push(
        {
            action: "@",
            context: "autoformat.autocomplete",
            description: "Riferimento"
        });
});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.confluence.plugins.confluence-dashboard:footer-analytics', location = '/js/confluence-dashboard/utils/footer-analytics.js' */
define("confluence-dashboard/utils/footer-analytics",["jquery","ajs","confluence/meta"],function(e,a,t){"use strict";var n=i(e).default,o=i(a).default,c=i(t).default;function i(e){return e&&e.__esModule?e:{default:e}}o.toInit(function(){var e=c.get("page-id"),a=n("body").hasClass("dashboard");n("#footer").on("click","a",function(){o.trigger("analytics",{name:"confluence.footer.item.click",data:{pageId:e,isDashboard:a}})})})}),require(["confluence-dashboard/utils/footer-analytics"]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.atlassian.auiplugin:split_aui.component.dialog2', location = 'aui.chunk.8dfc581e537092369347--79cee1339050bf73f760.js' */
(window.__auiJsonp=window.__auiJsonp||[]).push([["aui.component.dialog2"],{"0+IH":function(i,e,a){"use strict";Object.defineProperty(e,"__esModule",{value:!0}),e.dialog2=void 0;var t=a("p4ys");Object.defineProperty(e,"dialog2",{enumerable:!0,get:function(){return function(i){return i&&i.__esModule?i:{default:i}}(t).default}}),a("AehQ"),a("ZO0i")},ZO0i:function(i,e,a){},p4ys:function(i,e,a){"use strict";Object.defineProperty(e,"__esModule",{value:!0});var t=s(a("+x/D")),u=s(a("TmQU")),n=s(a("KloK")),o=s(a("jEnx")),l=s(a("6RZY")),d=s(a("+ay7"));function s(i){return i&&i.__esModule?i:{default:i}}var f={"aui-focus":"false","aui-blanketed":"true"};var c='\n        <section role="dialog" class="aui-layer aui-dialog2 aui-dialog2-medium" aria-hidden="true">\n            <header class="aui-dialog2-header">\n                <h2 class="aui-dialog2-header-main"></h2>\n                <a class="aui-dialog2-header-close">\n                    <span class="aui-icon aui-icon-small aui-iconfont-close-dialog">Close</span>\n                </a>\n            </header>\n            <div class="aui-dialog2-content"></div>\n            <footer class="aui-dialog2-footer"></footer>\n        </section>';function r(i){this.$el=i?(0,t.default)(i):(0,t.default)(c),function(i){t.default.each(f,function(e,a){var t="data-"+e;i[0].hasAttribute(t)||i.attr(t,a)})}(this.$el)}r.prototype.on=function(i,e){return(0,o.default)(this.$el).on(i,e),this},r.prototype.off=function(i,e){return(0,o.default)(this.$el).off(i,e),this},r.prototype.show=function(){return(0,o.default)(this.$el).show(),this},r.prototype.hide=function(){return(0,o.default)(this.$el).hide(),this},r.prototype.remove=function(){return(0,o.default)(this.$el).remove(),this},r.prototype.isVisible=function(){return(0,o.default)(this.$el).isVisible()};var h=(0,l.default)("dialog2",r);h.on=function(i,e){return o.default.on(i,".aui-dialog2",e),this},h.off=function(i,e){return o.default.off(i,".aui-dialog2",e),this},(0,t.default)(document).on("click keydown",".aui-dialog2-header-close",function(i){("click"===i.type||i.keyCode===d.default.ENTER||i.keyCode===d.default.SPACE)&&(i.preventDefault(),h((0,t.default)(i.target).closest(".aui-dialog2")).hide())}),h.on("show",function(i,e){e.find(".aui-dialog2-header-close").attr("tabindex",0);var a;[".aui-dialog2-content",".aui-dialog2-footer",".aui-dialog2-header"].some(function(i){return(a=e.find(i).find(":aui-tabbable")).length}),a&&a.first().focus()}),h.on("hide",function(i,e){var a=(0,o.default)(e);e.data("aui-remove-on-hide")&&a.remove()}),(0,u.default)("aui/dialog2",h),(0,n.default)("dialog2",h),e.default=h,i.exports=e.default}},[["0+IH","runtime","aui.splitchunk.vendors--e54c7c7304","aui.splitchunk.172ffb6da7","aui.splitchunk.b0831cc7d0","aui.splitchunk.087aba7911","aui.splitchunk.fbd1a5ab27","aui.splitchunk.54d3f16c20","aui.splitchunk.e54c7c7304","aui.splitchunk.f5828f5ab9","aui.splitchunk.fb15cffa72","aui.splitchunk.f673ef53ac","aui.splitchunk.8659b532c1","aui.splitchunk.d0110a864f","aui.splitchunk.afa5039e04","aui.splitchunk.bff3715233","aui.splitchunk.c750721820","aui.splitchunk.6d6f245ed3","aui.splitchunk.5b8c290363"]]]);
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'confluence.web.resources:page-loading-indicator', location = '/includes/js/page-loading-indicator.js' */
define("confluence/page-loading-indicator",["jquery","underscore","ajs","confluence/templates"],function(c,g,a,h){return function(b){function d(){return c(".confluence-page-loading-blanket",b)}function e(){return c(".confluence-loading-indicator",b)}return{show:function(){0===d().length&&c(b).append(h.pageLoadingIndicator());d().show();e().spin({lines:12,length:8,width:4,radius:10,trail:60,speed:1.5,color:"#F0F0F0"})},hide:function(){e().stop();d().hide()},showUntilResolved:function(c,b){this.show();
c.always(g.bind(this.hide,this));b&&c.fail(function(){a.messages.error(".confluence-page-loading-errors",{body:b})})},showUntilDialogVisible:function(b,d){var f=this,e=d||"Si \u00e8 verificato un errore durante la connessione al server. Riprova pi\u00f9 tardi.",g=c(".aui-dialog:visible"),h=c(".aui-dialog2:visible");!g.length&&!h.length&&f.show();b.always(f.hide).fail(function(){a.messages.error(".confluence-page-loading-errors",{body:e})});a.bind("show.dialog",function i(){a.unbind("show.dialog",i);f.hide()});if(null!=a.dialog2&&void 0!=a.dialog2)a.dialog2.on("show",
function j(){a.dialog2.off("show",j);f.hide()})},destroy:function(){b.remove(".confluence-page-loading-blanket")}}}});require("confluence/module-exporter").exportModuleAsGlobal("confluence/page-loading-indicator","Confluence.PageLoadingIndicator");
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'confluence.web.resources:page-loading-indicator', location = '/includes/soy/page-loading-indicator.soy' */
// This file was automatically generated from page-loading-indicator.soy.
// Please don't edit this file by hand.

/**
 * @fileoverview Templates in namespace Confluence.Templates.
 */

if (typeof Confluence == 'undefined') { var Confluence = {}; }
if (typeof Confluence.Templates == 'undefined') { Confluence.Templates = {}; }


Confluence.Templates.pageLoadingIndicator = function(opt_data, opt_ignored) {
  return '<div class="confluence-page-loading-errors"></div><div class="confluence-page-loading-blanket aui-blanket" aria-hidden="false"><div class="confluence-loading-indicator"></div></div>';
};
if (goog.DEBUG) {
  Confluence.Templates.pageLoadingIndicator.soyTemplateName = 'Confluence.Templates.pageLoadingIndicator';
}

}catch(e){WRMCB(e)};
;
try {
/* module-key = 'confluence.web.resources:deferred-loaders', location = '/includes/js/deferred-dialog-loader.js' */
define("confluence/deferred-dialog-loader",["underscore","jquery","ajs","confluence/page-loading-indicator","confluence/api/event"],function(c,d,i,j,b){return function(){var k=j(d("body")),e,g=!1,h=!1,f=function(){g&&h&&e&&(b.trigger(e),e=void 0)};b.bind("page.move.dialog.ready",function(){g=!0;f()});b.bind("blogpost.move.dialog.ready",function(){h=!0;f()});c.each({movePageDialogTools:{resource:"confluence.web.resources:page-move-resources",selector:"#action-move-page-dialog-link",event:"deferred.page-move.tools-menu"},
movePageDialogEditor:{resource:"confluence.web.resources:page-move-resources",selector:"#rte-button-location",event:"deferred.page-move.editor"},moveBlogDialogTools:{resource:"confluence.web.resources:page-move-resources",selector:"#action-move-blogpost-dialog-link",event:"deferred.blog-move.tools-menu"},availableGadgetsHelp:{resource:"com.atlassian.confluence.plugins.gadgets:gadget-directory-resources",selector:"#gadget-directory-link",event:"deferred.available-gadgets.help-menu"},aboutConfluenceHelp:{resource:"confluence.web.resources:about",
selector:"#confluence-about-link",event:"deferred.about-confluence.help-menu"}},function(a){d("body").on("click",a.selector,function(c){var d=a.resource+".requested",f=WRM.require("wr!"+a.resource,function(){"confluence.web.resources:page-move-resources"!==a.resource?b.trigger(a.event):g&&h?b.trigger(a.event):e=a.event});k.showUntilDialogVisible(f);i.Analytics?i.Analytics.triggerPrivacyPolicySafeEvent(d):b.trigger("analyticsEvent",{name:d});c.preventDefault()})})}});
require("confluence/module-exporter").safeRequire("confluence/deferred-dialog-loader",function(c){require("ajs").toInit(c)});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'confluence.web.resources:deferred-loaders', location = '/includes/js/page-permissions-deferred-loader.js' */
define("confluence/page-permissions-deferred-loader","confluence/dark-features confluence/legacy ajs confluence/page-loading-indicator jquery wrm".split(" "),function(a,b,c,d,e,f){var g=d(e("body"));return function(a){var b=f.require("wr!com.atlassian.confluence.plugins.confluence-page-restrictions-dialog:dialog-resources",function(){c.trigger("deferred.page.permissions")});g.showUntilDialogVisible(b,"Si \u00e8 verificato un errore durante il caricamento delle restrizioni della pagina. Ti invitiamo a riprovare pi\u00f9 tardi.");a.preventDefault()}});
require("confluence/module-exporter").safeRequire("confluence/page-permissions-deferred-loader",function(a){var b=require("ajs");b.toInit(function(c){c("body").on("click","#rte-button-restrictions,#action-page-permissions-link",a);b.bind("system-content-metadata.open-restrictions-dialog",a)})});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'confluence.web.resources:navigator-context', location = '/includes/js/api/navigator-context.js' */
define("confluence/api/navigator-context",["confluence/meta","document"],function(g,b){var h=function(a){a=d(a);return"undefined"!==typeof a&&0!==a},d=function(a){a=g.get(a);if(!isNaN(a))return Number(a)},i=function(){return!!b.querySelector(".page.view")||!!b.querySelector(".blogpost.view")},e=function(){return!!b.querySelector(".page.edit")||!!b.querySelector(".blogpost.edit")},f=function(){return g.get("content-type")},c=function(){return d("page-id")};return{getCurrent:function(){return e()&&
0===c()&&h("draft-id")?{target:"contentcreate",context:{contentId:d("draft-id"),contentType:f()}}:e()&&!i()&&0!==c()?{target:"contentedit",context:{contentId:c(),contentType:f()}}:!e()&&i()&&h("page-id")?{target:"contentview",context:{contentId:c(),contentType:f()}}:{target:"unknown",context:{}}}}});
}catch(e){WRMCB(e)};
;
try {
/* module-key = 'com.mxgraph.confluence.plugins.diagramly:drawio-blueprint-resources-admin', location = 'js/DiagramlyBlueprintInit.js' */
AJS.toInit(function(){if(Confluence&&Confluence.Blueprint&&Confluence.Blueprint.register){Confluence.Blueprint.register("com.mxgraph.confluence.plugins.diagramly:drawio-blueprint-item",function(b,c,a){window.location=Confluence.getContextPath()+"/pages/createpage.action?spaceKey="+encodeURIComponent(c)+"&newSpaceKey="+encodeURIComponent(c)+(b.getParentPageId()?"&fromPageId="+b.getParentPageId():"")+"&diagramlyBlueprint=1"})}});
}catch(e){WRMCB(e)};