function toggleFlash(e) {
var t = $(e), l = document.getElementsByClassName("title", t)[0], n = document.getElementsByClassName("main", t)[0];
Element.toggle(n), Element.visible(n) ? Element.addClassName(l, "on") :Element.removeClassName(l, "on");
}

!function() {
for (var e = document.getElementsByClassName("item"), t = 0; t < e.length; ++t) {
var l = document.getElementsByClassName("title", e[t])[0], n = document.getElementsByClassName("main", e[t])[0], o = document.getElementsByClassName("close", e[t])[0];
l && n && (l.onclick = toggleFlash.bind(null, e[t]), o && (o.onclick = toggleFlash.bind(null, e[t])));
}
var s = location.hash;
if (s) {
var m = s.substring(1);
document.getElementById(m) && toggleFlash(m);
}
}(), Event.observe(window, "load", function() {
Event.observe(window, "scroll", function() {
var e = document.documentElement.scrollTop || document.body.scrollTop;
e ? ($("toTop").style.display = "block", $("toTop").style.top = e + document.documentElement.clientHeight - 80 + "px") :$("toTop").style.display = "none";
}, !1);
}, !1);
//# sourceMappingURL=help.js.map