/* (C) Copyright IBM Corp. 2007  All Rights Reserved.                */
/**
 * This is the JS file for the common semantic tagging service
 */










 



 


// Only the methods marked with "public" are available to service implementors and others.
var SemTagSvc = {	

	debug: false,
	trace: false,
	version: "1.0",
	lang: "en",
	bidi: "ltr",
	tagScope: ['*'],
	baseUrl: "http://172.18.2.8/wps_semanticTag",
	service: 




{"entries":[
	 {"id":"com.ibm.portal.action","test":"(node.className.match(SemTagSvc.actionRE))","js":""}
	,{"id":"hcard","test":"(node.className.match(SemTagSvc.hcardRE))","js":"/javascript/semanticTagPerson.js"}
	//,{"id":"mailto","test":"(node.tagName.match(/^a$/i) && node.href.match(/^mailto:/) && !SemTagSvc.getParentByClassName('vcard', node))","js":"/javascript/semanticTagPerson.js"}
	,{"id":"sametime","test":"id:hcard","js":"/javascript/semanticTagAwareness.js"}
	,{"id":"c2a","test":"(node.className.match(/(^|\\s)c2a:(source|target)(\\s|$)/)) ","js":"/javascript/semanticTagC2A.js"}
	//,{"id":"adr","test":"(node.className.match(/(^|\\s)adr(\\s|$)/))","js":"/javascript/semanticTagAddress.js"}
	
]}
,
	scripts: new Array(),
	actionRegistry: null,
	refcntAttr: "semtag_refcnt",
	hoverIdPrefix: "semtag_hover_",
	hoverIdx: 0,
	liveElemPrefix: "semtag_live_",
	reMap: new Array(), 
	actionRE: new RegExp("(^|\\s)com\.ibm\.portal\.action(\\s|$)"),
	hcardRE: new RegExp("(^|\\s)vcard(\\s|$)"),
	specialMenuProviders: new Array(),
	parseElem: null,
	availAttribStr: "",
	ifConnUrlExists: "false",  //isConnUrlExists is defined in com.ibm.wps.semtag.UrlResService class.

	init: function(event) {
		var DELAY = 10; //SemTagUtil.isGecko? 10: 5000;
		window.setTimeout(SemTagMenu.init, DELAY);
		SemTagSvc.parseElem = SemTagSvc.getElementFromEvent(event);
		window.setTimeout(SemTagSvc.parseDom, DELAY);
		SemTagSvc.setCallback("com.ibm.portal.action", SemTagSvc.processActions);
	},

	loadConnScript: function() {
			var scriptElem = document.createElement("script");
			var url = "null";
			var profileUrl = url + "/profiles/portalJS/portalBizCard.js" ;
			
			profileUrl += (profileUrl.indexOf("?")==-1)? "?": "&";
			profileUrl += "lang=" + SemTagSvc.lang;
			scriptElem.src = profileUrl;
			document.body.insertBefore(scriptElem, document.body.firstChild); 
	}, 
 
	watchEvent: function(element, name, observer, useCapture) {
		try {
			if (element.addEventListener) element.addEventListener(name, observer, useCapture);
			else if (element.attachEvent) element.attachEvent('on' + name, observer);
		}
		catch (e) {
			if (SemTagSvc.debug) alert("Svc.watchEvent caught: " + e);
		}
	},

	clearEventWatch: function(element, name, observer, useCapture) {
		try {
			if (element.removeEventListener) element.removeEventListener(name, observer, useCapture);
			else if (element.detachEvent) element.detachEvent('on' + name, observer);
		}
		catch (e) {
			if (SemTagSvc.debug) alert("Svc.clearEventWatch caught: " + e);
		}
	},

	// public
	parseDom: function(/*Event*/ event, /*DomElement|String?*/ element) {
		// summary: parse DOM and search for potential live elements
		// description: 
		//
		// event: Not used in any way (defined just to allow this method to be used as a callback for an event)
		// element: Specifies the DOM node where the parse should start 'downward' (if null, 'document' is assumed, if a string is specified, it's assumed as the ID of the node, and otherwise, it is assumed as a DOM node)


  element = document.getElementsByTagName("span");

/*		if (SemTagSvc.trace) SemTagUtil.log("parseDom(" + event + "," + element + ")");
		if (!element && event) element = SemTagSvc.getElementFromEvent(event);
		if (!element && SemTagSvc.parseElem) element = SemTagSvc.parseElem;
			else if (!element) element = document;
			else if (typeof element == 'string') element = document.getElementById(element);
*/
		var svcEntries = SemTagSvc.service.entries;
		for (var j=0; j<svcEntries.length; j++) {
			var entry = svcEntries[j];
			if (!entry.nodes) entry.nodes = new Array();
				else if (element==document) while (0 < entry.nodes.length) entry.nodes.pop();
		}

//		var startTime = new Date().getTime(); //HMperf
//		SemTagSvc.traverseNodes(element);
//		var endTime = new Date().getTime(); //HMperf
//		alert("parse&test: " + (endTime-startTime)); //HMperf

        if(element || element.length>0){
		   for(var i= 0; i<element.length;i++)
			{
        		var classname = element[i].className;
		        if(classname.indexOf("vcard") > -1)
		        SemTagSvc.traverseNodes(element[i]);
			}
		}
		for (var k=0; k<svcEntries.length; k++) {
			var entry = svcEntries[k];
			var goAhead = false;
			if (0<entry.nodes.length) {
				goAhead = true;
			}
			else if (entry.test.match(/^id:/)) {
				var svcName = entry.test.substr(3);
				var nodes = SemTagSvc.getNodes(svcName);
				if (nodes && 0<nodes.length) {
					goAhead = true;
				}
			}
			if (goAhead) SemTagSvc.loadScript(entry.js);

			if (entry.callback) entry.callback.call(this);
		}

		SemTagSvc.parseElem = null;
	},

	traverseNodes: function(node) {
		//SemTagSvc.traverse_tagscope(node);
		SemTagSvc.traverse_children(node);
	},
	traverse_tagscope: function(node) {
		var svcEntries = SemTagSvc.service.entries;
		var children = node.getElementsByTagName(SemTagSvc.tagScope);
		for (var i = 0; i < children.length; i++) {
			var node = children[i];
			for (var j=0; j<svcEntries.length; j++) {
				var entry = svcEntries[j];
				if (!entry.nodes) entry.nodes = new Array();
				if (eval(entry.test)) entry.nodes.push(node);
			}
		}
	},
	traverse_children: function(node) {
		SemTagSvc.testNode(node);

//		var children = (node && node.childNodes)? node.childNodes: null;
		
		if (node && node.tagName && node.tagName=="IFRAME") {
			try {
				if (typeof ActiveXObject != "undefined") { 
					var childDoc = document.frames[node.id].document;
					if (childDoc) children = childDoc.childNodes;
					
				}
				else {  // Firefox, Safari, Opera
					if (node.contentDocument && node.contentDocument.childNodes) children = node.contentDocument.childNodes;
				}
			} catch (e) { 
					// this could happen when the IFRAME page returned 404/500/etc. 
					// or when we don t have permissions in the iframe doc
					children = null;
			}
			if (SemTagSvc.trace) SemTagUtil.log("IFRAME(id=" + node.id + ",children=" + children.length + ")");
			if (children && 0<children.length) {
				var firstTime = SemTagMenu.includeCSS(children[0].parentNode);
				if (firstTime) {
					SemTagSvc.watchEvent(node, 'load', SemTagSvc.parseDom, false);
					SemTagMenu.registerMenuEventHandlers(children[0].parentNode);
				}
			}
		}
//		if (children==null || children=='undefined') return;
//		for (var i=0; i<children.length; i++) SemTagSvc.traverse_children(children[i]);
	},
	testNode: function(node) {
		if (node==null || node.className==null || node.className=='undefined' || node.tagName==null || node.tagName=='undefined') return;
		var svcEntries = SemTagSvc.service.entries;
		for (var j=0; j<svcEntries.length; j++) {
			var entry = svcEntries[j];
			if (!entry.test.match(/^id:/) && eval(entry.test)) entry.nodes.push(node);
		}
	},
//	traverseDownNode: function(node,svcs) {
//		if (!svcs) svcs = SemTagSvc.getAllServices();
//		if (svcs.length==0) return;
//		SemTagSvc.testNode(node,svcs);
//		var children = (node.tagName=="IFRAME" && SemTagUtil.isGecko)? node.contentDocument.childNodes: node.childNodes;
//		if (!children) return;
//		for (var i=0; i<children.length; i++) SemTagSvc.traverseDownNode(children[i], SemTagSvc.cloneSvcArray(svcs));
//	},
//	testNode: function(node,svcs) {
//		if (!node || !node.className || !node.tagName) return;
//		for (var j=0; j<svcs.length; j++) {
//			var entry = svcs[j];
//			//if (!entry.nodes) entry.nodes = new Array();
//			if (eval(entry.test)) {
//				entry.nodes.push(node);
//				SemTagSvc.dismissService(svcs,entry.id);
//			}
//		}
//	},
//	getAllServices: function() {
//		var svcs = new Array();
//		for (var i=0; i<SemTagSvc.service.entries.length; i++) {
//			var entry = SemTagSvc.service.entries[i];
//			svcs.push(entry);
//		}
//		return svcs;
//	},
//	dismissService: function(svcs,svc2rm) {
//		for (var i=0; i<svcs.length; i++) {
//			if (svcs[i].id==svc2rm) {
//				svcs.splice(i,1);
//				break;
//			}
//		}
//	},
//	cloneSvcArray: function(svcs) {
//		var clone = new Array();
//		for (var i=0; i<svcs.length; i++) {
//			clone.push(svcs[i]);
//		}
//		return clone;
//	},

	getService: function(svcId) {
		var svcEntries = SemTagSvc.service.entries;
		for (var i=0; i<svcEntries.length; i++) {
			var entry = svcEntries[i];
			if (entry.id == svcId) return entry;
		}
	},
	
	// OKOETH
    // public
    setSemanticTagValue: function(/*String*/ svcId, /*Element*/ elem, /*Object*/ value) {
         // summary: store the given semantic tag value that the service has computed for
         //     a DOM node so that it can be used by other services plugging into the menu
         // description: a service can call this method to make semantic tag information 
         //     available that it has computed for a DOM node, before showing the context
         //     menu on this element. the "value" could simply be the information that has 
         //     been parse from the DOM, or it could be augmented by other data about the
         //     semantic tag that the service has retrieved e.g. in a server call.
         //     Note that multiple services may extract different information for a DOM node
         //     that is tagged in multiple ways
         //
         // svcId: specifies the Service that provides this information
         // elem: the DOM node to which the information should be attached
         // value: the computed semantic tag value (typically a JavaScript object)
         if (! elem.semTagValues) elem.semTagValues = {};
         elem.semTagValues[svcId] = value;
    },
	
    getSemanticTagValues: function(/*Element*/ elem) {
         // summary: retrieve the semantic tag values that LOF services have computed for
         //     a DOM node
         // returns: a map (object) from service IDs to values computed by the services. the
         //     return value may also be undefined, null or an empty object
         //
         // elem: the DOM node for which information should be retrieved
         // value: the computed semantic tag value (typically a JavaScript object)
         return elem.semTagValues;
    },
    
	// public
	getNodes: function(/*String*/ svcId) {
		// summary: find the DOM nodes that matches the given Service's criteria
		// description: 
		// returns: Array of DOM nodes

		var service = SemTagSvc.getService(svcId);
		if (service && service.test.match(/^id:/)) service = SemTagSvc.getService(service.test.substr(3));
		return service? service.nodes: null;
	},
	
	// public
	setCallback: function(/*String*/ svcId, /*Function*/ callback) {
		// summary: store the given callback function for the specified Service
		// description: 
		//
		// svcId: specifies the Service that the given callback is for
		// callback: callback function when parseDom finds matching nodes for the Service

		var service = SemTagSvc.getService(svcId);
		if (service) service.callback = callback;
	},
	
	loadScript: function(script) {
		if (!script || script=="") return;
		if (script.indexOf("semanticTagPerson.js") > -1)
		{
			if(SemTagSvc.ifConnUrlExists == "true") SemTagSvc.loadConnScript();
			//availAttrib.retrieveAttribute();
		}

		if (!SemTagSvc.scripts[script]) {
			SemTagSvc.scripts[script] = true;
			var scriptElem = document.createElement("script");
			//var url = script.match(/^http/)? script: SemTagSvc.baseUrl + script;
			var url = script.match(/^http/)? script: "/wps_semanticTag" + script;
			// pass along the language parameter
			url += (url.indexOf("?")==-1)? "?": "&";
			url += "language=" + SemTagSvc.lang;
			scriptElem.src = url;
			document.body.insertBefore(scriptElem, document.body.firstChild);
		}
	},
	
	setSpecialMenuProvider: function(callback) {
		SemTagSvc.specialMenuProviders.push(callback);
	},

	// public
	getElementsByClassName: function(/*String*/ className, /*DomElement*/ element, /*Integer*/ limitCnt, /*String[]*/ tagScope) {
		// summary: find DOM nodes within/under the given element that have the specified value in the class attribute
		// description: 
		// returns: Array of DOM nodes
		//
		// className: class name to look for matches
		// element: element to start traversing down
		// limit: max matches to find (0 means no limit)
		// tagScope: limit searches in the specified tag types

		if (!element) element = document.body; // entire 'body' if not passed in
		if (!limitCnt) limit = 0; // no limit if not passed in
		if (!tagScope) tagScope = ['*']; // no scoping if not passed in

		var regexp = SemTagSvc.getRegExp(className);
		if (element && element.className && element.className.match(regexp)) return new Array(element);

		var elements = new Array();
		for (var t=0; t<tagScope.length; t++) {
			var children = element.getElementsByTagName(tagScope[t]);
			for (var i=0; i<children.length; i++) {
				var child = children[i];
				if (child.className && child.className.match(regexp)) elements.push(child);
				if (0<limitCnt && limitCnt==elements.length) break;
			}
		}
		return elements;
	},

	// public
	getParentByClassName: function(/*String*/ cName, /*DomElement*/ element) {
		// summary: find a nearest parent node from the given element that has the specified value in the class attribute
		// description: 
		// returns: DOM node (null if not found)
		//
		// cName: class name to look for matches
		// element: element to start traversing down

		if (!element) return null;
		var regexp = SemTagSvc.getRegExp(cName);
		if (element.className && element.className.match(regexp)) return element;
		while(element.parentNode) {
			element = element.parentNode;
			if (element.className && element.className.match(regexp)) return element;
		}
		return null;
	},

	// public
	addHover: function(/*DomElement*/ elem, /*JsFunction*/ hoverHandler, /*JsFunction*/ clickHandler, /*String*/ alttext) {
		// summary: add a hover on the given DOM node, associating the given callbacks
		// description: 
		//
		// elem: DOM element to make 'live'
		// hoverHandler: JS event handler for 'mouseover' event to show the hover
		// clickHandler: JS event handler for 'click' event to supply the menu items
		// alttext: alternative text that could be read out by a screen reader software

		while ( typeof SemTagMenu == undefined ) alert( "waiting..." ); //HMtest

		var refcnt = elem.getAttribute(SemTagSvc.refcntAttr);
		if (refcnt) { // this element has already got the basic setup
			elem.setAttribute(SemTagSvc.refcntAttr, parseInt(Number(refcnt)+1)); // increment the ref count

			if (SemTagMenu.staticHover) {
				var img = SemTagMenu.findHoverFromLiveElement(elem);
				if (img) {
					SemTagSvc.watchEvent(img, 'click', clickHandler, false);
                	img.setAttribute("href", "javascript:SemTagMenu.a11y()");
				}
				else if (SemTagSvc.debug) alert("couldn't find the hover for this element!");
			}
			else SemTagSvc.watchEvent(elem, 'mouseover', hoverHandler, false);
		}
		else { // this element needs a new setup
			elem.setAttribute(SemTagSvc.refcntAttr, "1");
			SemTagSvc.hoverIdx++;
			elem.setAttribute(SemTagSvc.liveElemPrefix + "id", SemTagSvc.hoverIdx); //NEEDSWORK "hover" index?

			if (SemTagMenu.staticHover) {
				var img = SemTagSvc.createHoverImage(alttext);
				SemTagSvc.watchEvent(img, 'click', clickHandler, false);
				SemTagSvc.watchEvent(img, 'keydown', SemTagMenu.a11y, false); //A11Y
				img.setAttribute("href", "javascript:void()"); //A11Y make it visibly focusable

				var sibling = elem.nextSibling;
				if (sibling) sibling.parentNode.insertBefore(img,sibling);
					else elem.parentNode.appendChild(img);

				elem.setAttribute(SemTagSvc.hoverIdPrefix + "idx", parseInt(SemTagSvc.hoverIdx));
				elem.id = SemTagSvc.liveElemPrefix + SemTagSvc.hoverIdx; //NEEDSWORK what if the element already had an id...
			}
			else { // dynamic hover == visual indicator as to where to hover
				var classAttrs = SemTagUtil.getNodeClassValue(elem);
				if (classAttrs && 0<classAttrs.length) classAttrs += " hasHover"; //NEEDSATTN decoration
					else classAttrs = "hasHover";
				SemTagUtil.setNodeClassValue(elem, classAttrs);
				SemTagSvc.watchEvent(elem, 'mouseover', hoverHandler, false);
				elem.setAttribute("tabIndex", "0"); //A11Y make it not a tab stop
				SemTagSvc.watchEvent(elem, 'focus', hoverHandler, false); //A11Y
				SemTagSvc.watchEvent(elem, 'blur', SemTagMenu.mouseout, false); //A11Y
				elem.setAttribute("title", alttext);
			}
		}
	},

	removeHover: function(elem,hoverHandler,clickHandler) {
		var refcnt = elem.getAttribute(SemTagSvc.refcntAttr);
		if (refcnt) { // this element has some hover(s)
			var newCnt = Number(refcnt)-1;
			if (newCnt<0 && SemTagSvc.debug) alert("SemTagSvc.removeHover called on an element with refcnt=" + refcnt);
			if (SemTagMenu.staticHover) {
				var hover = SemTagSvc.getHoverElement(elem);
				SemTagSvc.clearEventWatch(hover, 'click', clickHandler, false);
				if (newCnt==0) {
					hover.parentNode.removeChild(hover);
				}
			}
			else { // dynamic
				SemTagSvc.clearEventWatch(elem, 'mouseover', hoverHandler, false);
				SemTagSvc.clearEventWatch(elem, 'focus', hoverHandler, false); //A11Y
				elem.setAttribute(SemTagSvc.refcntAttr, newCnt);
				if (newCnt==0) {
					var classAttrs = SemTagUtil.getNodeClassValue(elem);
					var newAttrs = classAttrs.replace(/hasHover/, "");
					SemTagUtil.setNodeClassValue(elem, newAttrs); //NEEDSWORK only when there is some meaningful values in it
					elem.removeAttribute(SemTagSvc.refcntAttr);
					elem.removeAttribute("tabIndex"); //A11Y make it not a tab stop
				}
			}
		}
	},

	createHoverImage: function(alttext) {
		var img = document.createElement("img");
		img.id = SemTagSvc.hoverIdPrefix + SemTagSvc.hoverIdx;
		img.className = SemTagMenu.iconName;
		//img.setAttribute("src", SemTagSvc.baseUrl +"/ui/menu_selected.gif");
		img.setAttribute("src", "/wps_semanticTag/ui/menu_selected.gif");
		img.setAttribute("border", "0");
		img.setAttribute("alt", alttext); //A11Y

		var link = document.createElement("a");
		SemTagSvc.watchEvent(link,"mouseover",SemTagMenu.activateHover,false);
		SemTagSvc.watchEvent(link,"mouseout",SemTagMenu.deactivateHover,false);
		link.appendChild(img); //A11Y - caller should add 'href' for appropriate event handler

		return link;
	},

	getHoverElement: function(/*DomElement*/ liveElem) { //A11Y
		// summary: find the hover element associated with the given DOM node
		// description: 
		// returns: DOM node that represents the hover (null if not found)
		//
		// liveElem: a DOM element that was previously passed to addHover()

		if (!liveElem) return null;
		if (SemTagMenu.staticHover) {
			var idx = liveElem.getAttribute(SemTagSvc.hoverIdPrefix + "idx");
			var img = document.getElementById(SemTagSvc.hoverIdPrefix + idx);
			if (img) return img.parentNode;
		}
		else { // dynamic hover
			var classAttrs = SemTagUtil.getNodeClassValue(liveElem);
			if (classAttrs.match(/hasHover/)) return liveElem;
		}
		return null;
	},

	// public
	showHover: function(/*Event*/ event, /*Function*/ clickHandler, /*String*/ label) {
		// summary: show the hover for the live element that fired the given event
		// description:
		//
		// event: JS 'mouseover' event object
		// clickHandler: JS function to be called upon click
		// label: Text that the hover should show

		SemTagMenu.showHover(event, clickHandler, label);
	},

	// public
	setMenuData: function(/*Event*/ event, /*MenuItemJson[]?*/ items, /*String?*/ cssClass, /*HeaderJson?*/ header, /*FooterJson?*/ footer) {
		// summary: set menu data that a Service wants to show for the popup
		// description: 
		//
		// event: JS 'click' event object
		// items: array of 'menu item JSON'
		// cssClass: CSS selector for the items
		// header: 'header JSON'
		// footer: 'footer JSON'

		SemTagMenu.setMenuData(event, items, cssClass, header, footer);
	},

	// public
	getMenuItemJson: function(/*String*/ label, /*String*/ href, /*Integer*/ order, /*String*/ icon) {
		// summary: create a JSON object that represents a menu item
		// description: 
		// returns: JSON object that can be passed into setMenuData (MenuItemJson) 
		//
		// label: text label to show in the menu
		// href: hypertext link (could be javascript: as well) to be executed when selected
		// order: indicates the preference as to where in the menu to be shown (the smaller the number, including negative, the higher it's shown in the menu)
		// icon: URL to the icon image to use with the label

		var o = order? order: 0;
		var i = icon? icon: "";
		return {"label": label, "href": href, "order": o, "icon": i};
	},

	// public
	getMenuHeaderJson: function(/*String*/ markup, /*String*/ mimetype, /*Integer*/ order) {
		// summary: create a JSON object that represents a header
		// description: 
		// returns: JSON object that can be passed into setMenuData (HeaderJson) 
		//
		// markup: HTML that could be used in the header section of the popup
		// mimetype: MIME type (eg. "text/html") of the 'markup' value
		// order: indicates how much it wants to be used (the smaller the number, including negative, the more likely it's used)

		if ( mimetype != "text/html" ) return null;
		if ( markup && 0 < markup.length ) return {"markup": markup, "order": order};
		else null;
	},

	// public
	getMenuFooterJson: function(/*String*/ markup, /*String*/ mimetype, /*Integer*/ order) {
		// summary: create a JSON object that represents a footer
		// description: 
		// returns: JSON object that can be passed into setMenuData (FooterJson) 
		//
		// markup: HTML that could be used in the footer section of the popup
		// mimetype: MIME type (eg. "text/html") of the 'markup' value
		// order: indicates how much it wants to be used (the smaller the number, including negative, the more likely it's used)

		if ( mimetype != "text/html" ) return null;
		if ( markup && 0 < markup.length ) return {"markup": markup, "order": order};
		else null;
	},

	getElementFromEvent: function(event) {
		//return event.target? event.target: event.srcElement;
		var element = event.target? event.target: event.srcElement;
		if (element == window)
		{
			element = element.document;
		}
		return element;
	},

	// public
	getLiveElementFromEvent: function(event) {
		return SemTagMenu.findLiveElementFromEventSource(SemTagSvc.getElementFromEvent(event));
	},

	getEventAbsoluteX: function (e) {
		// find out X
		var x=0;
		if (e.pageX) x = e.pageX;
		else if (e.clientX) {
			if (document.body.scrollLeft > document.documentElement.scrollLeft) {
				x = e.clientX + document.body.scrollLeft;
			}
			else {
				x = e.clientX + document.documentElement.scrollLeft;
			}
		}
		return x;
	},
	
	getEventAbsoluteY: function (e) {
		// find out Y
		var y=0;
		if (e.pageY) y = e.pageY;
		else if (e.clientY) {
			if (document.body.scrollTop > document.documentElement.scrollTop) {
				y = e.clientY + document.body.scrollTop;
			}
			else {
				y = e.clientY + document.documentElement.scrollTop;
			}
		}
		return y;
	},
	
	findPosition: function (obj, rightedge) {
		var count=0, posX=0, posY=0;
		var objW = obj.offsetWidth;
		if( obj.offsetParent ) {
			if (rightedge) posX += (SemTagSvc.bidi=='rtl')? 0: objW;
			posY += obj.offsetHeight;
			while (obj != null) {
				posX += obj.offsetLeft;
				posY += obj.offsetTop;
				obj = obj.offsetParent;
				count++;
			}
			return [ posX + (SemTagSvc.bidi=='rtl'? objW: 0), posY ];
		} else {
			return [ obj.x + (SemTagSvc.bidi=='rtl'? objW: 0), obj.y ];
		}
	},
	
	createGroupJson: function(ctx,exts) { // top-level objects in 'actionRegistry' array
		return {"context": ctx, "extenders": exts};
	},

	createActionJson: function(id,js,ctx,label,desc,showif,url,order) { // elements in actionRegistry[x].extenders array
		var i = (id && 0<id.length)? id[0].innerHTML: null;
		var j = (js && 0<js.length)? js[0].innerHTML: null;
		var c = (ctx && 0<ctx.length)? ctx[0].innerHTML: null;
		var l = (label && 0<label.length)? label[0].innerHTML: null;
		var d = (desc && 0<desc.length)? desc[0].innerHTML: null;
		var s = (showif && 0<showif.length)? showif[0].innerHTML: null;
		var u = (url && 0<url.length)? url[0].innerHTML: null;
		//var o = SemTagSvc.parseOrder(order);
		var o = (order && 0<order.length)? parseInt(order[0].innerHTML): 0;
		return {"id": i, "impl": j, "context": c, "label": l, "description": d, "showif": s, "url": u, "order": o};
	},
	parseOrder: function(order) {
		if (order==null || order.length==0) return 0;
		if (order.charAt(0)=='-') return parseInt(order.substr(1)) * -1;
		else return parseInt(order);
	},

	// public
	getActions: function(/*String*/ context) {
		// summary: find menu extensions defined in the page that matches the given context
		// description: 
		// returns: Array of JSON objects that that have the following fields: id, impl, context, label, description, showif, url, order (all fields are of type String)
		//
		// context: specifies the context to search menu extensions for

		if (!SemTagSvc.actionRegistry) SemTagSvc.processActions(); // actionRegistry building - run once (usually)

		for (var i=0; i<SemTagSvc.actionRegistry.length; i++) {
			if (SemTagSvc.actionRegistry[i].context==context) return SemTagSvc.actionRegistry[i].extenders;
		}
		return SemTagSvc.actionRegistry[0].extenders; // no such extenders == 'dummy' array
	},

	processActions: function() {
		if (!SemTagSvc.actionRegistry) {
			SemTagSvc.actionRegistry = new Array();
			SemTagSvc.actionRegistry[0] = SemTagSvc.createGroupJson("dummy", new Array());
		}

		var i;
		var actions = SemTagSvc.getNodes("com.ibm.portal.action");
		while (0 < actions.length) {
			// prepare the Action Json
			var actnNode = actions.pop();
			var id = SemTagSvc.getElementsByClassName("action-id",actnNode);
			var impl = SemTagSvc.getElementsByClassName("action-impl",actnNode);
			var ctx = SemTagSvc.getElementsByClassName("action-context",actnNode);
			var label = SemTagSvc.getElementsByClassName("action-label",actnNode);
			var desc = SemTagSvc.getElementsByClassName("action-description",actnNode);
			var showif = SemTagSvc.getElementsByClassName("action-showif",actnNode);
			var url = SemTagSvc.getElementsByClassName("action-url",actnNode);
			var order = SemTagSvc.getElementsByClassName("action-order",actnNode);
			var actn = SemTagSvc.createActionJson(id,impl,ctx,label,desc,showif,url,order);

			for (i=1; i<SemTagSvc.actionRegistry.length; i++) {
				if (SemTagSvc.actionRegistry[i].context==actn.context) { // add 'actn' if there's already a matching entry in actionRegistry
					var cnt = SemTagSvc.actionRegistry[i].extenders.length;
					for (var j=0; j<cnt; j++) {
						if ( actn.id == SemTagSvc.actionRegistry[i].extenders[j].id ) break;
					}
					if ( j == cnt ) SemTagSvc.actionRegistry[i].extenders.push( actn );
					break;
				}
			}
			if (i==SemTagSvc.actionRegistry.length) { // well, this is new - let's create it
				SemTagSvc.actionRegistry[i] = SemTagSvc.createGroupJson(ctx[0].innerHTML, new Array());
				SemTagSvc.actionRegistry[i].extenders.push( actn );
			}

			if (actn.impl) SemTagSvc.loadScript(actn.impl);

//			if (actn.showif) {
//				for (i=0; i<9; i++) {
//					SemTagMenu.waitCursor();
//					var func = actn.showif.match(/\)$/)? actn.showif: actn.showif + "()";
//					try {
//						var ready2call = (eval(func) != 'undefined');
//						if (ready2call) break;
//					}
//					catch (e) {
//					}
//				}
//				SemTagMenu.defaultCursor();
//			}
		}

		for (i=0; i<SemTagSvc.actionRegistry.length; i++) SemTagSvc.actionRegistry[i].extenders.sort( SemTagSvc.sortByOrder );
	},

	sortByOrder: function(a,b) {
		if (a.order > b.order) return 1;
		else if (a.order < b.order ) return -1;
		else return 0;
	},

	// public
	getTextValue: function(elem) {
		if (!elem) return "";
		return elem.innerHTML.replace(/<[a-zA-Z\/][^>]*>/gi,"");
	},

	// public
	getTypedValue: function(elem,defProp) {
		if (!defProp) defProp = "def"; //NEEDSATTN "def" is not a standard sub-property
		var returnElem = new Array();
		var types = SemTagSvc.getElementsByClassName("type", elem);
		var values = SemTagSvc.getElementsByClassName("value", elem);
		var value = "";
		if (values.length < 1) value = SemTagSvc.getTextValue(elem);
		for (var i=0;i<values.length;i++) {
			value += SemTagSvc.getTextValue(values[i]);
		}
		if (types.length < 1) returnElem[defProp] = value;
		else {
			for (var j=0;j<types.length;j++) {
				var typeElem = types[j];
				var type = typeElem.tagName.match(/^abbr$/i)? typeElem.getAttribute("title"): SemTagSvc.getTextValue(typeElem);
				returnElem[type.toLowerCase()] = value;
			}
		}
		return returnElem;
	},

	// public
	findNameElementInHcard: function(elem) {
		if (elem.className!="vcard") elem = SemTagSvc.getParentByClassName('vcard',elem);
		var nameElem = SemTagSvc.getElementsByClassName("fn",elem)[0];
		if (!nameElem) {
			nameElem = SemTagSvc.getElementsByClassName("n",elem)[0];
		}
		return nameElem;
	},

	// public
	getEmailFromHcard: function(elem) {
		if (!elem) return "";
		if (elem.className!="vcard") elem = SemTagSvc.getParentByClassName('vcard',elem);
		var email = "";
		if (elem) {
			var mailElem = SemTagSvc.getElementsByClassName("email", elem)[0];
			var tvObj = SemTagSvc.getTypedValue(mailElem, "internet");
			email = tvObj.internet;
		}
		return email;
	},

	// public
	findElementByNameInHcard: function(elem,name) {
		if (elem.className!="vcard") elem = SemTagSvc.getParentByClassName('vcard',elem);
		return SemTagSvc.getElementsByClassName(name,elem)[0];
	},

	getRegExp: function(str) { // not so effective for performance gain...
		var regexp = SemTagSvc.reMap[str];
		if (!regexp) {
			regexp = new RegExp("(^|\\s)" + str + "(\\s|$)");
			SemTagSvc.reMap[str] = regexp;
		}
		return regexp;
	}
}

/* (C) Copyright IBM Corp. 2007  All Rights Reserved.                */
/**
 * This is the common JS file for menus
 */
 
// Unless specifically noted, everything in this object is private to the semantic tagging / menu generation service core




var mytest = null;
var SemTagMenu = {

	needCss: true,
	staticHover: false, 
	id: "semtagmenu",
	hideDelay: 1500, // '0' requires an explicit event to dismiss the popup
	timeouts: new Array(),
	iconName: "menu_drop_icon",
	showing: false,
	currentElem: null,
	currentHoverLabel: null,
	refCount: -1,
	a11yMode: false,

	svcHandlers: new Array(),
	items: new Array(),
	headers: new Array(),
	footers: new Array(),

	hoverIdRE: new RegExp(SemTagSvc.hoverIdPrefix),
	iconNameRE: new RegExp("(^|\\s)menu_drop_icon(\\s|$)"), //NEEDSWORK why can't you use SemTagMenu.iconName?
	
	init: function() {
		SemTagMenu.includeCSS(document);
		//SemTagMenu.registerMenuEventHandlers(document);
	},

	includeCSS: function(node) {
		var firstTime = false;

		if (typeof(node._JAVLIN_STYLE_) == 'undefined') {
			firstTime = true;

			node._JAVLIN_STYLE_ = "loaded";
		}

		return firstTime;
	},

	registerMenuEventHandlers: function(node) {
		SemTagSvc.watchEvent(node, 'click', SemTagMenu.click, false);
		SemTagSvc.watchEvent(node, 'keydown', SemTagMenu.catchEscape, false); // A11Y
	},

	unregisterMenuEventHandlers: function(node) {
		SemTagSvc.clearEventWatch(node, 'click', SemTagMenu.click, false);
		SemTagSvc.clearEventWatch(node, 'keydown', SemTagMenu.catchEscape, false); // A11Y
	},

	 
 
 
nls: {
 
"hover_label": "Click for options",
"a11y_hover": "more information",
"a11y_photo": "photo",
"a11y_close": "close"
 
},
 
hoverDimension:[14,14],
hoverOffset:[15,-1],
menuOffset:[15,-1],
writeHover: function(out, bidi, label) {
out.write("<div class='" + SemTagMenu.iconName + "' style='cursor:pointer;'>" + label + "</div>");
},
startMenu: function(out, bidi) {
out.write("<div class='personMenu'>");
},
writeHeader: function(out, header, bidi) {
out.write("<div class='semtag_header'>" + header.markup + "</div>");
},
startActionSection: function(out, bidi, selector) {
if (!selector) selector="personMenuActions";
out.write("<table class='" + selector + "' dir='" + bidi + "'><tbody>");
},
writeMenuItem: function(out, item, bidi) {
out.write("<tr><td><a href='" + item.href + "'>" + item.label + "</a></td></tr>");
},
endActionSection: function(out, bidi) {
out.write("</tbody></table>");
},
writeFooter: function(out, footer, bidi) {
out.write("<div class='semtag_footer'>" + footer.markup + "</div>");
},
endMenu: function(out, bidi) {
out.write("<a style='width:0px; height:0px;' onblur=\"SemTagMenu.exitCard();\" onclick='return false;' href='exit from person card' title='exit from person card'></a></div>");
},


	out: function () {
		this.buffer = "";
		this.write = function (str) {
			this.buffer += str;
		}
	},

	activateHover: function(event) {
		var imgElem = SemTagSvc.getElementFromEvent(event);
		if (imgElem) imgElem.src = SemTagSvc.baseUrl +"/ui/menu_selected_hover.gif";
		SemTagMenu.setCurrentElement(SemTagMenu.findLiveElementFromEventSource(imgElem));
	},

	deactivateHover: function(event) {
		var imgElem = SemTagSvc.getElementFromEvent(event);
		if (imgElem) imgElem.src = SemTagSvc.baseUrl +"/ui/menu_selected.gif";
		//SemTagMenu.setCurrentElement(null);
	},

	showHover: function(event, clickHandler, label) {
		if (SemTagMenu.staticHover) return;

		var tag = SemTagMenu.getMenuTag();
		if (SemTagMenu.showing && tag.style.display != "none") return;

		if (SemTagSvc.trace) SemTagUtil.log("Menu.showHover");

		var elem = SemTagSvc.getElementFromEvent(event);
		if (SemTagSvc.trace) SemTagUtil.log("currentElem=" + SemTagMenu.currentElem);
		var origRefCnt = 0;
		if (elem && elem != SemTagMenu.currentElem) { // new hover request
			SemTagMenu.clearAllSvcHandlers(tag);

			origRefCnt = elem.getAttribute(SemTagSvc.refcntAttr);
			if (SemTagSvc.trace) SemTagUtil.log("refcnt=" + origRefCnt);
			if (origRefCnt) {
				SemTagMenu.setCurrentElement(elem);
			}
			else {
				if (SemTagSvc.debug) alert("SemTagMenu.showHover called for a DOM element with no refcnt attribute!");
				SemTagMenu.setCurrentElement(null);
				return;
			}
		}

		SemTagMenu.addSvcHandler(tag, clickHandler);

		if (label && 0<label.length) SemTagMenu.currentHoverLabel = label;

		SemTagMenu.refCount--; // decrement for this current call
		if (0 < SemTagMenu.refCount) return; // need to wait more services to call in

		// everybody called in = time to actually show hover

		SemTagMenu.setRefCount((0<origRefCnt)? origRefCnt: Number(elem.getAttribute(SemTagSvc.refcntAttr)));

		SemTagMenu.showing = false;
		if (event.type=="focus") { //A11Y
			SemTagMenu.a11yMode = true;
			SemTagSvc.watchEvent(elem, 'keydown', SemTagMenu.a11y, false);
		}

		var out = new SemTagMenu.out();
		var label = SemTagMenu.currentHoverLabel? SemTagMenu.currentHoverLabel: SemTagMenu.nls.hover_label;
		SemTagMenu.writeHover(out, SemTagSvc.bidi, label);
    	availAttrib.retrieveAttribute();
		SemTagMenu.offScreen(tag);
		tag.innerHTML = out.buffer;
		var pos = SemTagSvc.findPosition(SemTagMenu.currentElem, false);
		
		var adjustX = (SemTagSvc.bidi=='rtl')? 0 - SemTagMenu.hoverOffset[0]: SemTagMenu.hoverOffset[0];
		SemTagMenu.show(SemTagMenu.id, event,
			pos[0] + adjustX, // x
			pos[1] + SemTagMenu.hoverOffset[1], // y
			tag.offsetWidth, // w
			tag.offsetHeight); // h
	},

	addSvcHandler: function(tag, handler) {
		SemTagMenu.svcHandlers.push(handler);
		SemTagSvc.watchEvent(tag, 'click', handler, false);
	},

	clearAllSvcHandlers: function(tag) {
		while (0 < SemTagMenu.svcHandlers.length) {
			var handler = SemTagMenu.svcHandlers.pop();
			if (handler) SemTagSvc.clearEventWatch(tag, 'click', handler, false);
		}
	},

	setCurrentElement: function(elem) {
//		if (SemTagSvc.debug && elem) {
//			alert("SemTagMenu.setCurrentElement: " + elem);
//		}
		SemTagMenu.currentElem = elem;
		SemTagMenu.setRefCount(elem? Number(elem.getAttribute(SemTagSvc.refcntAttr)): 0);
		SemTagMenu.currentHoverLabel = null;

		while (0 < SemTagMenu.headers.length) SemTagMenu.headers.pop();
		while (0 < SemTagMenu.items.length) SemTagMenu.items.pop();
		while (0 < SemTagMenu.footers.length) SemTagMenu.footers.pop();
	},

	setRefCount: function(cnt) {
		SemTagMenu.refCount = cnt;
		//if (SemTagSvc.debug && cnt<=0) alert("Unless this is from hide(), this is bad!");
	},

	setMenuData: function(event, items, cssClass, header, footer) {
		if (SemTagSvc.debug) window.status = "Menu.setMenuData: items.length=" + (items? items.length: 0);

		if (!SemTagMenu.staticHover && SemTagMenu.refCount < 0) {
			if (SemTagSvc.debug) alert("setMenuData called when refCount=" + SemTagMenu.refCount);
			return;
		}

		var elem = SemTagMenu.findLiveElementFromEventSource(SemTagSvc.getElementFromEvent(event));
		if (!elem) {
			if (SemTagSvc.debug) alert("setMenuData called on a null live element");
			return;
		}
		if (SemTagMenu.staticHover && // need to figure out the current live element and its ref count
			(elem != SemTagMenu.currentElem || SemTagMenu.showing)) {
			SemTagMenu.setCurrentElement(elem);
		}

//		SemTagMenu.waitCursor(); // need to do this only once, but...

		if (items) for (var i=0; i<items.length; i++) SemTagMenu.items.push(items[i]);
		if (cssClass) SemTagMenu.currentMenuCss = cssClass;
		if (header) SemTagMenu.headers.push(header);
		if (footer) SemTagMenu.footers.push(footer);

		SemTagMenu.refCount--; // decrement for this current call

		if (0 < SemTagMenu.refCount) return; // need to wait more services to call in

		// everybody called in = time to actually show popup
		SemTagMenu.stopEvent(event); // eat it always - this method is called upon ENTER as well

		// but wait - we need to call 'special menu providers'
		for (var j=0; j<SemTagSvc.specialMenuProviders.length; j++) {
			var callback = SemTagSvc.specialMenuProviders[j];
			var specialMenus = callback.call(event, SemTagMenu.currentElem);
			if (specialMenus && 0<specialMenus.length) {
				for (var k=0; k<specialMenus.length; k++)
					SemTagMenu.items.push(specialMenus[k]);
			}
		}

		SemTagMenu.showMenu(event);
	},

	// public
	getCurrentElement: function() {
		return SemTagMenu.currentElem;
	},

	showMenu: function(event) {
		if (0 == SemTagMenu.headers.length+SemTagMenu.items.length+SemTagMenu.footers.length) {
			SemTagMenu.hide();
			return;
		}

		if (SemTagSvc.trace) SemTagUtil.log("Menu.showMenu");

		SemTagMenu.items.sort(SemTagSvc.sortByOrder);
		if (1 < SemTagMenu.headers.length) SemTagMenu.headers.sort(SemTagSvc.sortByOrder);
		if (1 < SemTagMenu.footers.length) SemTagMenu.footers.sort(SemTagSvc.sortByOrder);

		var out = new SemTagMenu.out();
		SemTagMenu.startMenu(out, SemTagSvc.bidi);
		if (0 < SemTagMenu.headers.length) SemTagMenu.writeHeader(out, SemTagMenu.headers[0], SemTagSvc.bidi);
		SemTagMenu.startActionSection(out, SemTagSvc.bidi, SemTagMenu.currentMenuCss);
//		for (var i=0; i<SemTagMenu.items.length; i++) {
//			SemTagMenu.writeMenuItem(out, SemTagMenu.items[i], SemTagSvc.bidi);
//		}
		while ( 0 < SemTagMenu.items.length ) {
			SemTagMenu.writeMenuItem(out, SemTagMenu.items.shift(), SemTagSvc.bidi);
		}
		SemTagMenu.endActionSection(out, SemTagSvc.bidi);
		if (0 < SemTagMenu.footers.length) SemTagMenu.writeFooter(out, SemTagMenu.footers[0], SemTagSvc.bidi);
		SemTagMenu.endMenu(out, SemTagSvc.bidi);

		var tag = SemTagMenu.getMenuTag();
		SemTagMenu.clearAllSvcHandlers(tag);
		SemTagMenu.offScreen(tag);
		tag.innerHTML = out.buffer;
		SemTagMenu.showing = true;

		SemTagMenu.defaultCursor();

		if (SemTagMenu.a11yMode) {
			if (SemTagUtil.isGecko) { // SemTagMenu.staticHover?
				var links = tag.getElementsByTagName("li"); //A11Y
				if (0 < links.length) links[0].focus(); //NEEDSWORK this 'focus' remains if you start using mouse to change the menu selection...
			}
			else
				tag.focus(); //A11Y works great for IE, and with dynamic hover in FF
		}
		SemTagMenu.registerMenuEventHandlers(document);

		//var event = {"target": SemTagMenu.currentElem}; // make a fake event
		var pos = SemTagSvc.findPosition(SemTagMenu.currentElem, false);
		var adjustX = (SemTagSvc.bidi=='rtl')? 0 - SemTagMenu.menuOffset[0]: SemTagMenu.menuOffset[0];
		SemTagMenu.show(SemTagMenu.id, event, 
			pos[0] + adjustX, // x
			pos[1] + SemTagMenu.menuOffset[1], // y
			tag.offsetWidth, // w
			tag.offsetHeight); // h
	},

	getMenuTag: function() {
		var tag = document.getElementById(SemTagMenu.id);
		if (!tag) {
			tag = document.createElement("div");
			tag.setAttribute("id", SemTagMenu.id);
			tag.style.position = "absolute";
			tag.style.display = "none";
			tag.style.zIndex = "99999";
			SemTagSvc.watchEvent(tag, 'mouseout', SemTagMenu.mouseout, false);
			SemTagSvc.watchEvent(tag, 'mouseover', SemTagMenu.mouseover, false);
			//document.body.appendChild(tag);
			document.body.insertBefore(tag, document.body.firstChild);
		}
		return tag;
	},

	show: function (pMenu, e, xpos, ypos, width, height) {
		if (SemTagSvc.trace) SemTagUtil.log("SemTagUtil.show");
        var iframeScrollTop;  //Variables defined for scrolling defect 193500
		var iframeScrollLeft; //Variables defined for scrolling defect 193500
		// first clear all previous timeouts since we're showing new menu
		SemTagMenu.clearTimeouts();
		var ptagMenu=document.getElementById(pMenu);
		if (width == null) width=0;
		if (height == null) height=0;

		var top,left;
		if (xpos != null && ypos != null) {
			var top = ypos;
			var left = xpos - (SemTagSvc.bidi=='rtl'? width: 0);
		}
		else {
			var top = SemTagSvc.getEventAbsoluteY(e);
			var left = SemTagSvc.getEventAbsoluteX(e);
		}

		var vSrc = SemTagMenu.currentElem;
		mytest = vSrc.parentNode;
		// find winWidth and winHeight
		var winWidth, winHeight, d=document;
		if (typeof window.innerWidth!='undefined') {
			winWidth = window.innerWidth;
			winHeight = window.innerHeight;
		} else {
			if (d.documentElement && typeof d.documentElement.clientWidth!='undefined' && d.documentElement.clientWidth!=0) {
				winWidth = d.documentElement.clientWidth
				winHeight = d.documentElement.clientHeight
			} else {
				if (d.body && typeof d.body.clientWidth!='undefined') {
					winWidth = d.body.clientWidth
					winHeight = d.body.clientHeight
				}
			}
		}
		
		var scrollX = (document.body.scrollLeft > document.documentElement.scrollLeft)? document.body.scrollLeft: document.documentElement.scrollLeft;
		
		
		

		
		if ((left + width) > (winWidth + scrollX)) {
			var howMuch = (left + width) - winWidth - scrollX;
			left -= howMuch;
		}
		var scrollY = (document.body.scrollTop > document.documentElement.scrollTop)? document.body.scrollTop: document.documentElement.scrollTop;
		

		
		if ((top + height) > (winHeight + scrollY)) {
			var howMuch = (top + height) - winHeight - scrollY;
			top -= howMuch;
		}

		if (SemTagSvc.bidi=='rtl' && left < 0) left = 0;

		//NEEDSWORK ugly, ugly, ugly
		//SemTagUtil.log("before IFRAME adjust (" + left + "," + top + ")"); //HMdebug
		//SemTagUtil.log( "ownerDoc=" + vSrc.ownerDocument ); //HMdebug
		//SemTagUtil.log( "location=" + vSrc.ownerDocument.location ); //HMdebug
		var od = SemTagUtil.getOwnerDocument(vSrc);
		if (od.location != document.location) { // okay, I'm in a child page (IFRAME)...
			var f = SemTagUtil.getFrameElement(vSrc);
			if (f) {

				//fix for defect 201052,now displaying hover wrt live object
				var el =f.offsetParent;
				var x = 0, y = 0;

					while (el)
					{
					x += el.offsetLeft;
					y += el.offsetTop;
					el = el.offsetParent;
					}

					var lof =vSrc.offsetParent;

					var xx = 0, yy = 0;

					while (lof)
					{
					xx += lof.offsetLeft;
					yy += lof.offsetTop;
					lof = lof.offsetParent;
					}

				if (f.id=='wpsFLY_flyoutIFrame') { // ugly, but works
					//top += f.parentNode.offsetTop;
					//left += f.parentNode.offsetLeft;
					top =  y  + yy + vSrc.offsetHeight;
					left = x + xx + vSrc.offsetWidth ;
				}
				else {
					top += f.offsetTop;
					left += f.offsetLeft;
				}
			//Code fix for scrolling defect 193500--Start
			if(SemTagUtil.isGecko && f.id=='wpsFLY_flyoutIFrame'){
				
                  iframeScrollLeft = f.contentWindow.pageXOffset;
                  iframeScrollTop  = f.contentWindow.pageYOffset;
 				  //alert("FF-- iframeScrollLeft : "+iframeScrollLeft + "iframeScrollTop: " +iframeScrollTop);
            }
			if(!SemTagUtil.isGecko && f.id=='wpsFLY_flyoutIFrame'){
                  iframeScrollLeft = f.contentWindow.document.body.parentNode.scrollLeft;
                  iframeScrollTop  = f.contentWindow.document.body.parentNode.scrollTop;
				  //alert("IE-- iframeScrollLeft : "+iframeScrollLeft + "iframeScrollTop: " +iframeScrollTop);
		    }
					
			top -= iframeScrollTop;
            left -= iframeScrollLeft;
			//Code fix for scrolling defect 193500--End
			}
		}
		//alert("ypos=" + ypos + "\nheight=" + height + "\npageY=" + e.pageY + "\nwinHeight=" + winHeight + "\nclientY=" + e.clientY + "\ndocument.body.scrollTop=" + document.body.scrollTop + "\ndocument.documentElement.scrollTop=" + document.documentElement.scrollTop + "\ntop=" + top + "\nleft=" + left); //HMdebug
		ptagMenu.style.top = top + "px";
		if(left < 500){		
		    ptagMenu.style.left = left + 25  + "px";}
		else {
			ptagMenu.style.left = left + "px";}
		ptagMenu.style.display="block";
		
		SemTagMenu.startHideTimer(pMenu);
	},

	hide: function (pMenu,e) {
		SemTagMenu.unregisterMenuEventHandlers(document);

		var tag = SemTagMenu.getMenuTag();
		if (!tag) return false;

		if (SemTagSvc.trace) SemTagUtil.log("Menu.hide");
		if(tag.style.display != "block")
		{
			return false;
		}
		tag.style.display="none";
		SemTagMenu.showing = false;
		// added to set focus on next element to get it working for a11y
		try{
			var nextElem = SemTagMenu.getNextEle(mytest);
   			nextElem.focus(); 		
		}catch(e)  
		{
			document.body.focus();
		}
		SemTagMenu.setCurrentElement(null);
		SemTagMenu.currentHoverLabel = null;
		SemTagMenu.clearAllSvcHandlers(tag);
		SemTagMenu.a11yMode = false;
		SemTagMenu.defaultCursor();

		return true;
	},
	getNextEle:function(ele)
	{//alert("ele : " + ele);
		var par = ele.parentNode;
		//alert("par : " + par);
		while(true){
			if(par == null || par == 'undefined')
			{
				return false;
			}else if( par.nodeName == 'BODY')
			{
				return par;
			}else
			{
				var temp = par.nextSibling;
				if(temp != null)
				{
					return temp;
				}
				par = par.parentNode;
			}
		}
		return false;
	},
	
	offScreen: function(pMenu) {
		pMenu.style.top = "-1000px";
		pMenu.style.left = "-1000px";
		pMenu.style.display="block";
	},

	findLiveElementFromEventSource: function(eventSrc) {
		if (SemTagMenu.staticHover) {
			var id = eventSrc.id;
			if (!id.match(SemTagMenu.hoverIdRE)) {
				var children = eventSrc.getElementsByTagName("img"); //NEEDSATTN
				for (var i=0; i<children.length; i++) {
					if (children[i].id && children[i].id.match(SemTagMenu.hoverIdRE)) {
						id = children[i].id;
						break;
					}
				}
			}
			if ( id && 0 < id.length ) {
				var idx = id.substr(SemTagSvc.hoverIdPrefix.length);
				return document.getElementById(SemTagSvc.liveElemPrefix + idx);
			}
			else return eventSrc; // this saves the "static+inline" case
		}
		else {
			var liveElem = SemTagMenu.getCurrentElement();
			return liveElem? liveElem: eventSrc;
		}
	},

	findHoverFromLiveElement: function(liveElem) {
		var idx = liveElem.getAttribute(SemTagSvc.hoverIdPrefix + "idx");
		return document.getElementById(SemTagSvc.hoverIdPrefix + idx);
	},

	// check to see if the mouse is in the menu during this event...optional buffer parameters
	// can be used to expand the size of the menu
	inMenu: function(menuElem,event,bufferX,bufferY) {
		if (!menuElem) return false;
		if (!SemTagMenu.showing) return false;

		if (!bufferX) bufferX = 0;
		if (!bufferY) bufferY = 0;
		var mouseX = SemTagSvc.getEventAbsoluteX(event);
		var mouseY = SemTagSvc.getEventAbsoluteY(event);
		var elemX = menuElem.style.left.replace(/px$/,'');
		var elemY = menuElem.style.top.replace(/px$/,'');
		// if (debug) alert("mouseX:" + mouseX + ",mouseY:" + mouseY + ";elemX:" + elemX + ",elemY:" + elemY);
		var sumX = parseInt(elemX) + parseInt(menuElem.clientWidth);
		var sumY = parseInt(elemY) + parseInt(menuElem.clientHeight);
		// for some terrible reason, in IE, if you move off the element slowly (to the top or left), 
		// the last event happens just inside the menu, which is why we need to do mouseX-1 and mouseY-1
		if ((mouseX-1 <= (elemX - bufferX)) || (mouseY-1 <= (elemY - bufferY)) ||
			(mouseX >= (sumX + bufferX)) || (mouseY >= (sumY + bufferY))) return false;
		else return true;
	},

	mouseout: function(event) {
		if (SemTagSvc.trace) SemTagUtil.log("Menu.mouseout");
		if (event.type=="blur") { // current live element is losing focus
			var evtSrc = SemTagSvc.getElementFromEvent(event);
			if (SemTagMenu.currentElem) {
				if (SemTagSvc.trace) SemTagUtil.log("clearEventWatch");
				SemTagSvc.clearEventWatch(SemTagMenu.currentElem, "keydown", SemTagMenu.a11y, false);
			}
			if (!SemTagMenu.showing) { // hover is being dismissed
				if (SemTagSvc.trace) SemTagUtil.log("dismissing hover");
				if (!evtSrc.className.match(SemTagMenu.iconNameRE)) SemTagMenu.hide();
			}
			// else we don't hide the popup as we are just popping it up
		}
		else {
			var menuId = SemTagMenu.id;
			menuElem = document.getElementById(menuId);
			if (SemTagMenu.inMenu(menuElem, event)) {
			}
			else {
				SemTagMenu.startHideTimer(menuId);
			}
		}
	},
	
	mouseover: function(event) {
		// if we're over the menu, clear timeouts so it doesn't hide
		SemTagMenu.clearTimeouts();
	},
	
	click: function(event) {
		if (!event) return;

		// on document click, check to see if you are in menu, and hide if you are
		// only hide if we were not invoked by the person tag image
		var srcElement = SemTagSvc.getElementFromEvent(event);
//		if (SemTagMenu.staticHover && srcElement.firstChild && srcElement.firstChild.className==SemTagMenu.iconName) return;
//		if (!srcElement.className.match(SemTagMenu.iconNameRE)) {
//			var menuElem = document.getElementById(SemTagMenu.id);
//			var headerElem = SemTagSvc.getElementsByClassName("photoCard", menuElem, 1)[0]; //HMtest
//			//if (headerElem) menuElem = headerElem;
//			if (!SemTagMenu.inMenu(menuElem, event)) SemTagMenu.hide(SemTagMenu.id);
//		}
		var container = SemTagSvc.getParentByClassName("semtag_header", srcElement);
		if (!container) SemTagMenu.hide(SemTagMenu.id);
	},

	catchEscape: function (event) {
		var key;
		if (window.event) key = event.keyCode; // IE
			else if (event.which) key = event.which; // Netscape, FireFox, Opera
		if (SemTagSvc.trace) SemTagUtil.log("Menu.catchEscape:key=" + key);

		if (key == 27) {
			var nextFocus = null;
			if (SemTagMenu.showing) { // if we are hiding the popup...
				var currElem = SemTagMenu.getCurrentElement();
				if (currElem) { //A11Y
					nextFocus = SemTagSvc.getHoverElement(currElem);
				}
			}
			SemTagMenu.hide();
			if (nextFocus) nextFocus.focus();
		}
	},

	a11y: function(event) { //A11Y catch the ENTER key on hover and fake a click event
		if (event) {
			var key;
			
			if (event.which) key = event.which; // Netscape, FireFox, Opera
				else key = event.keyCode; // IE as well as Iframe in IE (people palette)

			if (SemTagSvc.trace) SemTagUtil.log("Menu.a11y:key=" + key);

			if (key == 13) {
				var elem = SemTagSvc.getElementFromEvent(event);
				if (elem) {
					SemTagMenu.a11yMode = true; //A11Y
					var target = (SemTagMenu.staticHover? elem: SemTagMenu.getMenuTag());
					if (SemTagUtil.isGecko) {
						var evt = document.createEvent("MouseEvents");
						evt.initEvent("click", true, true);
						SemTagUtil.fireEvent(target, evt);
					}
					else { // IE
						target.click();
					}
					SemTagMenu.stopEvent(event);
				}
			}
		}
		else { // this must be the ENTER on the 'fn' elem
			var elem = SemTagMenu.currentElem;
			if (elem) {
				SemTagMenu.a11yMode = true; //A11Y
				var target = (SemTagMenu.staticHover? elem: SemTagMenu.getMenuTag());
				if (SemTagUtil.isGecko) {
					var evt2 = document.createEvent("MouseEvents");
					evt2.initEvent("click", true, true);
					SemTagUtil.fireEvent(target, evt2);
				}
				else { // IE
					target.click();
				}
				SemTagMenu.stopEvent(evt);
			}
		}
		return true; //NEEDSWORK? does this help stop event propagation?
	},

	stopEvent: function(event)
	{
		if (!event) return;

		if (SemTagUtil.isGecko) {
			try {
				event.preventDefault();
				event.stopPropagation();
			}
			catch (e) {
				// this event object may be a fake one, and there's no stopping necessary...
			}
		}
		else {
			try {
				event.returnValue = false;
				event.cancelBubble = true;
			}
			catch (e) {
				// In IE, when Person menu sends a server request out (passing browser 
				// cache), IE will somehow manage to corrupt this click event it had sent 
				// out earlier, and it causes silent error (abort) while doing the above.
				if (SemTagSvc.trace) SemTagUtil.log("stopEvent caught " + e);
			}
		}
	},

	startHideTimer: function(menuId) {
		if (0<SemTagMenu.hideDelay && !SemTagMenu.a11yMode && !SemTagSvc.debug) {
			SemTagMenu.timeouts.push(window.setTimeout('SemTagMenu.endHideTimer("' + menuId + '")',SemTagMenu.hideDelay));
		}
	},
	
	endHideTimer: function(menuId) {
		if (SemTagSvc.trace) SemTagUtil.log("Menu.endHideTimer");
		SemTagMenu.hide(menuId);
	},
	
	clearTimeouts: function() {
		var tos = SemTagMenu.timeouts;
		for (i=0;i<tos.length;i++) {
			window.clearTimeout(tos[i]);
		}
	},

	defaultCursor: function() {
		document.body.style.cursor = "default";
	},

	waitCursor: function() {
		document.body.style.cursor = "progress";
	},

    exitCard:function(){
    var nextFocus = null;
			if (SemTagMenu.showing) { // if we are hiding the popup...
				var currElem = SemTagMenu.getCurrentElement();
				if (currElem) { //A11Y
					nextFocus = SemTagSvc.getHoverElement(currElem);
				}
			}
			SemTagMenu.hide();
			if (nextFocus) nextFocus.focus();
	}
}

/**
 * This is the common JS file for utility methods
 */




var SemTagUtil = {

	isGecko: (document.all? false : true),

	getNodeClassValue: function(node) {
		var rv;
		if (SemTagUtil.isGecko) {
			//NEEDSWORK? I don't know how fast this 'undefined check' executes...
			rv = (typeof(node.getAttribute)!='undefined')? node.getAttribute("class"): "";
		}
		else {
			rv = node.className;
		}
		return (typeof(rv)!='undefined' && rv!=null)? rv: "";
	},

	setNodeClassValue: function(node,value) {
		if (SemTagUtil.isGecko)
			node.setAttribute("class", value);
		else
			node.className = value;
	},

	fireEvent: function(element, event) {
		try {
			if (element.fireEvent) element.fireEvent(event);
			else if (element.dispatchEvent) element.dispatchEvent(event);
		}
		catch (e) {
			if (SemTagSvc.debug) alert("Svc.fireEvent caught: " + e);
		}
	},

	getOwnerDocument: function(element) {
		if (!element) return null;

		if (SemTagUtil.isGecko) {
			return element.ownerDocument;
		}
		else {
			var tmp = element;
			while (tmp.parentNode) tmp = tmp.parentNode;
			return tmp;
		}
	},

	getFrameElement: function(element) {
		if (SemTagUtil.isGecko) {
			var oD = element.ownerDocument;
			return oD.defaultView.frameElement;
		}
		else {
			var tmp = element;
			while (tmp.parentNode) tmp = tmp.parentNode;
			return tmp.parentWindow.frameElement;
		}
	},

	getHcardAttributeValue: function(hcardAttr, srcElement) {
		switch(hcardAttr) {
			case("email"):
				return SemTagUtil.getHcardTypedAttribute(srcElement, hcardAttr, "internet");
			case("tel"):
				return SemTagUtil.getHcardTypedAttribute(srcElement, hcardAttr, "voice");
			case("adr"):
				return SemTagUtil.getHcardTypedAttribute(srcElement, hcardAttr, "intl");
		}
		var parentVcard = SemTagSvc.getParentByClassName("vcard", srcElement);
		var elems = SemTagSvc.getElementsByClassName(hcardAttr, parentVcard, 1);
		if (elems.length > 0 && elems[0].tagName.match(/^abbr$/i)) return elems[0].getAttribute("title");
		switch (hcardAttr) {
			case("X-person-display-inline"): // this test is done for every hCard on the page, so, do it first
				if (elems.length > 0) return true;
				else return false;
			case("fn"):
				var fn = SemTagUtil.getSinglePropertyValue(elems[0]);
				if (fn) return fn;
					else return SemTagUtil.getHcardAttributeValue("n", srcElement); // no fn, parse for n
				break;
			case("X-person-header-only"):
			case("X-person-inside-inline"):
				return (elems && 0 < elems.length);
			case("X-sametime-resolve"):
				var stResolve = elems[0];
				return (typeof(stResolve)!='undefined' && stResolve!=null);
			case("n"):
				if (elems.length > 0) {
					var nElem = elems[0];
					var fnStr = "";
					var attrs = ["honorific-prefix","given-name","additional-name","family-name","honorific-suffix"];
					for (var i=0;i<5;i++) {
						var n = SemTagUtil.getSinglePropertyValue(SemTagSvc.getElementsByClassName(attrs[i], nElem, 1)[0]);
						if (n) fnStr += n + " ";
					}
					return fnStr;
				}
				// empty or no n is valid, return blank
				return "";
				break;
			case("photo"):
				var photoElem = elems[0];
				if (photoElem) return photoElem.getAttribute("src");
				else return;
				break;
			case("X-sametime-status"):
				var stStatusElem = elems[0];
				if (!stStatusElem) return "";
				var stStatusValue = stStatusElem.getAttribute("value");
				if (stStatusValue) return stStatusValue;
				else return SemTagUtil.getSinglePropertyValue(stStatusElem);
				break;
			case("street-address"):
			case("post-office-box"):
			case("extended-address"):
			case("locality"):
			case("region"):
			case("postal-code"):
			case("country-name"):
			case("title"):
			case("role"):
			case("org"):
			default:
				return SemTagUtil.getSinglePropertyValue(elems[0]);
			break;
		}
	},

	getHcardTypedAttribute: function(srcElement, hcardAttr, defSubProp) {
		var returnElem = new Object();
		var parentVcard = SemTagSvc.getParentByClassName("vcard", srcElement);
		var typedElems = SemTagSvc.getElementsByClassName(hcardAttr, parentVcard);
		// this means there are multiple email, adr, or tel elements
		for (var i=0; i<typedElems.length; i++) {
			var curElem = typedElems[i];
//			if (curElem.tagName.match(/^abbr$/i)) {
//				returnElem[defSubProp] = curElem.getAttribute("title");
//				continue;
//			}
			if (hcardAttr == "email" && curElem.nodeName.toLowerCase() == 'a' && curElem.href.match(/^mailto:/)) {
				// then we need to use the href value
				var queryPos = curElem.href.indexOf("?");
				if (queryPos > -1) returnElem[defSubProp] = curElem.href.slice(7,queryPos);
				else returnElem[defSubProp] = curElem.href.slice(7);
				continue;
			}
			returnElem = SemTagSvc.getTypedValue(curElem, defSubProp);
		}
		return returnElem;
	},
	
	getSinglePropertyValue: function (prop) {
		if (!prop) return false;
		var returnVal = prop.innerHTML.replace(/<[a-zA-Z\/][^>]*>/gi,"");
		return returnVal;
	},

	/**
	 * Object to issue cross-domain AJAX calls.  This object uses 
	 * a dynamic script generation/removal process.  This is required due 
	 * to the security restrictions placed on XMLHttpRequest
	 */
	crossDomainRequest:	function() {
		var openConnectionMapping = new Array();
		var requestSrcElements = new Array();
		var self = this;

		createTimeoutFunction = function (connectionId) {
			return function() { self.cancelRequest(connectionId); }
		};
	
		this.getScriptId = function(id) {
			return "_JVLN_" + id;
		};
		
		this.getScriptObject = function(id)	{
			var scriptId = SemTagUtil.isGecko? this.getScriptId(id): this.$_getScriptId(id);
			return document.getElementById(scriptId);
		};
		
		this.request = function(url, timeLimit, callback, srcElement, connectionId) {
			var objId = this.getScriptId(connectionId);
			if(callback) {
				openConnectionMapping[connectionId] = callback;
			}
			if (srcElement) {
				requestSrcElements[connectionId] = srcElement;
			}
			var script = document.createElement("script");
			script.id = objId;
			try {
				script.src = url;
			}
			catch(e) {
				if (SemTagSvc.debug) alert("crossDomainRequest.request: " + e);
				return false;
			}
			// Insert at the first position to avoid odd IE behavior	
			document.body.insertBefore(script, document.body.firstChild);

			if(timeLimit) {
				var self = this;
				window.setTimeout(createTimeoutFunction(connectionId), timeLimit);
			}
		};
		
		this.cancelRequest = function(id) {
			if (SemTagSvc.debug) window.status = "crossDomainRequest.cancelRequest";
			// If the connection is still open, gracefully clean it up and report this to the user
			var callback = openConnectionMapping[id];
			openConnectionMapping[id] = null;
			var srcElem = requestSrcElements[id];
			requestSrcElements[id] = null;

			if(callback) {
				try {
					if(callback) {
						var evt = {"target": srcElem}; // I need a fake event
						callback.call(this, false, null, evt);
					}
					var script = this.getScriptObject(id);
					if(script) {
						document.body.removeChild(script);
					}
				}
				catch(e) {
					if (SemTagSvc.trace) SemTagUtil.log("crossDomainRequest.cancelRequest caught: " + e + "(callback=" + callback + ")");
				}
			}
		};
	
		this.dispatch = function(id, data) {
			if(!id) {
				return;
			}
			try {
				// Clean up the entry
				var callback = openConnectionMapping[id];
				if (SemTagSvc.debug && !callback) alert("dispatch got null callback for: " + id);
				if(callback) {
					// Call the specified callback function
					var evt = {"target": requestSrcElements[id]}; // I need a fake event
					callback.call(this, true, data, evt);

					// clean up
					openConnectionMapping[id] = null;
					requestSrcElements[id] = null;
					var script = this.getScriptObject(id);
					if(script) {
						document.body.removeChild(script);
					}
				}
			}
			catch(e) {
				if (SemTagSvc.trace) SemTagUtil.log("crossDomainRequest.dispatch caught: " + e);
			}
		};
	}, 
	
	log: function(msg) {
		var logger = document.getElementById("javlin.logger"); //NEEDSATTN
		if (logger) {
			var txt = document.createTextNode(msg + "..... ");
			logger.appendChild(txt);
		}
	}
}

/*
Licensed Materials - Property of IBM
 
5724S31

Copyright IBM Corp.  2007    All Rights Reserved.

US Government Users Restricted Rights - Use, duplication or 
disclosure restricted by GSA ADP Schedule Contract with IBM 
Corp.
*/


 

/**
 * This is the JS file for fetching the supported attributes of PUMA
 */
var availAttrib = {

	availAttribUrl: "/wps"+"/um/secure/attributes/users",
	req: "",

	retrieveAttribute: function() {

		if (SemTagSvc.availAttribStr != "")
			return;

		  availAttrib.req= ibm.portal.xml.getXmlHttpRequest();     
		  availAttrib.req.onreadystatechange = availAttrib.processStateChange;

		  try {

			availAttrib.req.open("GET", availAttrib.availAttribUrl, true);

		  } catch (e) {

			alert(e);

		  }

		  availAttrib.req.send(null);


	  },

	getAttributesFromTags: function (elm, tag, prefix) {
		 if ( typeof ActiveXObject != "undefined" ) {
		    return elm.getElementsByTagName(prefix + ":" + tag);
        }
        return elm.getElementsByTagNameNS("*", tag); 
	},

	processStateChange: function () {
		if (availAttrib.req.readyState == 4) { // Complete
		  if (availAttrib.req.status == 200) { // OK response
			var response = availAttrib.req.responseText;
		   if(response.length  > 1)	{
			   try{
			var oXmlDoc=ibm.portal.xml.loadXmlString(response); 
	
/*			if (typeof ActiveXObject != "undefined") 
				var attribs = oXmlDoc.documentElement.getElementsByTagName("atom:title");
			else 
				var attribs = oXmlDoc.documentElement.getElementsByTagName("title");
*/
			var attribs = availAttrib.getAttributesFromTags(oXmlDoc.documentElement, "title", "atom");
			var i;
			for (i=1;i<attribs.length;i++)
			{
				if (( SemTagSvc.availAttribStr == "" ))
				{
					if (( attribs[i].firstChild.nodeValue != null) )
					{
						SemTagSvc.availAttribStr = attribs[i].firstChild.nodeValue;
					}
				}else {

					if (( attribs[i].firstChild.nodeValue != null) )
					{
						SemTagSvc.availAttribStr = attribs[i].firstChild.nodeValue+","+ SemTagSvc.availAttribStr;
					}
				}
			}
			   }catch(e){}
		  }

		  } else {

			//alert("Problem: " + availAttrib.req.statusText);

		  }
		}

	  }

}

/*
Licensed Materials - Property of IBM
 
5724S31

Copyright IBM Corp.  2007    All Rights Reserved.

US Government Users Restricted Rights - Use, duplication or 
disclosure restricted by GSA ADP Schedule Contract with IBM 
Corp.
*/

//Make sure "ibm" and "ibm.portal" objects are defined, as it's conceivable that
//another product could define an "ibm" javascript object. It also is conceivable that
//another portion of the portal product would define an "ibm.portal" object. Hopefully
//they will play nicely as well and not overwrite OUR object definitions.
if(typeof(ibm)=="undefined")ibm = {};

if(typeof(ibm.portal)=="undefined")ibm.portal = {};
if(typeof(ibm.portal.xml)=="undefined")ibm.portal.xml = {};

// The object reference for all of the functions contained in this file
//ibm.portal.xml={};  commented out for defect 193789 and added following line inplace
if(typeof(ibm.portal.xml)=="undefined")ibm.portal.xml = {};

//----------------------------------------------------------------------  General utility methods
ibm.portal.util = {};

ibm.portal.util.cloneObject = function (obj) {
 var returnObj = {};
    for (i in obj) {
        returnObj[i] = obj[i];
    }
    return returnObj;
}

//----------------------------------------------------------------------- resource utility methods
ibm.portal.resource={};
ibm.portal.resource.str = {};
ibm.portal.resource.getString = function (bundle, key) {
    s = bundle[key];
    s = (s==null)?key:s;
    if (arguments.length >= 2) {
        i = 1;
        do {
            s = s.replace("%"+i , arguments[i]);
            i = i +1;
        } while ( i < arguments.length );
     }
     return s;
}

//---------------------------------------------------------------------- xml related utility methods
/*
 * getXmlHttpRequest()
 * loadXml(sUrl) - returns oDomDoc. parses xml from the url into DOM document object.
 * loadXmlString() - returns oDomDoc. parses xml from the contents of the string into DOM document object.
 * loadXsl()
 * transform()
 */
//ibm.portal.xml = {}; Commented out for defect 193789 and added following line inplace
if(typeof(ibm.portal.xml)=="undefined")ibm.portal.xml = {};
ibm.portal.xml.ie = {};
ibm.portal.xml.gecko = {};

ibm.portal.xml.getXmlHttpRequest = function() {
	var oXml = null;
	if (typeof ActiveXObject != "undefined") {
		oXml = new ActiveXObject("Microsoft.XMLHTTP");
	} else {
		oXml = new XMLHttpRequest();
	}
  return oXml;
}

ibm.portal.xml.loadXml = function(sUrl) {
	if (typeof ActiveXObject != "undefined") 
		return ibm.portal.xml.ie.loadXml(sUrl);
	else 
		return ibm.portal.xml.gecko.loadXml(sUrl);
}

ibm.portal.xml.loadXmlString = function(sXml) {
	if (typeof ActiveXObject != "undefined") 
		return ibm.portal.xml.ie.loadXmlString(sXml);
	else 
		return ibm.portal.xml.gecko.loadXmlString(sXml);
}

ibm.portal.xml.loadXsl = function(sUrl) {
	if (typeof ActiveXObject != "undefined") 
		return ibm.portal.xml.ie.loadXsl(sUrl);
	else 
		return ibm.portal.xml.gecko.loadXsl(sUrl);
}

ibm.portal.xml.transform = function (xml, xsl, sXslMode, aXslParams, bReturnString)
{
   if (typeof ActiveXObject != "undefined") {
    return ibm.portal.xml.ie.transform(xml, xsl, sXslMode, aXslParams,bReturnString);
  }
  else 
    return ibm.portal.xml.gecko.transform(xml, xsl, sXslMode, aXslParams,bReturnString);
}

ibm.portal.xml.update = function (nodeToUpdate, xml, xsl, sXslMode, aXslParams) {
	if ( typeof ActiveXObject != "undefined" ) {
		var results = ibm.portal.xml.ie.transform(xml, xsl, sXslMode, aXslParams, true );
		//Don't really want to use innerHTML here, but seems to be the only way IE will 
		//take the update.
		ibm.portal.debug.text( "XSLT result: " + results );
		nodeToUpdate.innerHTML += results;
	}
	else {
		results = ibm.portal.xml.gecko.transform(xml,xsl,sXslMode,aXslParams, false);
		nodeToUpdate.appendChild( results.documentElement );
	}
}

//---------------------------------------------------------------------- IE xml related utility methods

ibm.portal.xml.ie.loadXml = function(sUrl) {
	var oXmlDoc = new ActiveXObject("MSXML2.DOMDocument");
	oXmlDoc.async=0;
	oXmlDoc.resolveExternals = 0;
  	if(!oXmlDoc.load(sUrl))
  	{
  		//Callers should catch this and can substitute their own error message
  		throw new Error("Error loading xml file " + sUrl);
  	}
	return oXmlDoc;

}

ibm.portal.xml.ie.loadXmlString = function(sXml) {
	var oXmlDoc = new ActiveXObject("MSXML2.DOMDocument");
	oXmlDoc.async=0;
	oXmlDoc.resolveExternals = 0;
  	if(!oXmlDoc.loadXML(sXml))
  	{
  	    //Callers should catch this and can substitute their own error message
  		throw new Error("Error loading xml string " + sXml); 
  	}
	return oXmlDoc;
}

ibm.portal.xml.ie.loadXsl = function(sUrl) {
	//we need to use MSXML2.FreeThreadedDOMDocument interface in order to support 
	//mode and parameters in XSL transformation.
	var oXslDoc = new ActiveXObject("MSXML2.FreeThreadedDOMDocument");
	oXslDoc.async=0;
	oXslDoc.resolveExternals = 0;
  	if(!oXslDoc.load(sUrl))
  	{
  		//Callers should catch this and can substitute their own error message
  		throw new Error("Error loading xsl file " + sUrl);
  	}	
	return oXslDoc;
}

ibm.portal.xml.ie.transform = function(xmlDoc, xsl, sXslMode, aXslParams,bReturnString) {
	var oXml = xmlDoc;
	var oXsl = xsl;

     try {		
		if(!oXsl.documentElement) oXsl = this.loadXsl(xsl);
	 }
	 catch(e) {
		var sMsg = e.message;
		throw new Error(""+sMsg, ""+sMsg);
	 }
	//create the xsl processor and apply the transformation
	var oXslt = new ActiveXObject("Msxml2.XSLTemplate");
	oXslt.stylesheet = oXsl;
	var oXslProc = oXslt.createProcessor();
	oXslProc.input = oXml;
	
	//set paramaters if any
	if(aXslParams) {
		for(var p in aXslParams) {
			oXslProc.addParameter(p, aXslParams[p]);
		}
	}
	if (sXslMode) oXslProc.addParameter("mode", sXslMode);
	
	if (bReturnString) {
		if(!oXslProc.transform()) {
			//Callers should catch this and can substitute their own error message
  			throw new Error("Error transforming xml doc " + oXml); 
  		}
		return oXslProc.output;
	} else {
		var oHtmlDoc = new ActiveXObject("MSXML2.DOMDocument");
		oHtmlDoc.async = 0;
		oHtmlDoc.validateOnParse = 1;
		oXml.transformNodeToObject(oXsl,oHtmlDoc);
		return oHtmlDoc;	
	}
}

//---------------------------------------------------------------------- GECKO xml related utility methods

ibm.portal.xml.gecko.loadXml = function(sUrl) {
    //var oXmlResponse = NG.ServerRequest.postRequest(sUrl);
//    if (oXmlResponse) return xmlLoadString(oXmlResponse.responseText);
//    else return null;
	var oDomDoc = document.implementation.createDocument('','',null); 
//	oDomDoc.async = 0; // this is the important part
//	oDomDoc.load(sUrl);
// to support for Safari browser
		var xhr = new XMLHttpRequest();
		xhr.open( "GET", sUrl, false );
		xhr.send( null );
		  
		if ( xhr.status == 200 ) {
			oDomDoc = xhr.responseXML;
		}
	
	return oDomDoc;
}

ibm.portal.xml.gecko.loadXmlString = function(sXml) {
    var parser = new DOMParser();
    try { oXmlDoc = parser.parseFromString(sXml, "text/xml"); }
    catch (exc) {
	    //Callers should catch this and can substitute their own error message
  		throw new Error("Error loading xml string " + sXml); 
    }
	return oXmlDoc;
}

ibm.portal.xml.gecko.loadXsl = function(sUrl) {
	//This is done through createDocument because of anchor(#) we have in portal url. 
	//Do not change the code without testing the case.
	var oDomDoc = document.implementation.createDocument('','',null); 
//	oDomDoc.async = 0; // this is the important part
//	oDomDoc.load(sUrl);
	var xhr = new XMLHttpRequest();
		xhr.open( "GET", sUrl, false );
		xhr.send( null );
		  
		if ( xhr.status == 200 ) {
			oDomDoc = xhr.responseXML;
		}
	
	return oDomDoc;
}

ibm.portal.xml.gecko.transform = function(xmlDoc, xsl, sXslMode, aXslParams,bReturnString) {
	try {
	  var xslDoc = xsl;
      if(!xslDoc.documentElement) 
      {
          alert("xslDoc is not a Document, loading it...");
          xslDoc = this.loadXsl(xsl);
      }        
      var proc = new XSLTProcessor();
      proc.importStylesheet(xslDoc); 
 
      //set parameters if any
      if(aXslParams) {
	  	for(var p in aXslParams)  {
		    proc.setParameter(null, p, aXslParams[p]);
	    }
      }
      if (sXslMode) proc.setParameter(null, "mode", sXslMode);
    
      var resultDoc = proc.transformToDocument(xmlDoc);
      if (!bReturnString) {
      	return resultDoc;
      }
      resultStr = resultDoc.documentElement.textContent;
    } 
    catch (exc)
    {
        //alert("error transforming document: "+exc)
		//Callers should catch this and can substitute their own error message
  		throw new Error("Error transforming xml doc " + exc);
	}
    return resultStr;  
}

/* This method sets the content of a layer within the HTML page
 * to the result of transforming the xml parameter by the xsl
 * parameter. The xml and xsl parameters may be of any form
 * supported by the transformXml() method. The layer parameter
 * may be either a DOM object or the name of a DOM object that
 * can be found using the findObject() method.
 */

ibm.portal.xml.setLayerContentByXml = function (layer, xml, xsl, xslparam,bReturnString) {
    var result = ibm.portal.xml.transform(xml,xsl,null,xslparam,bReturnString);
    if (layer.innerHTML) layer.innerHTML = result;
    else {
        var obj = document.getElementById(layer);
        obj.innerHTML=result;
    }
}
			   
//----------------------------------------------------------------------  temporary ajax request method 
ibm.portal.io ={};

/* Accept-Language property */
ibm.portal.io.sAcceptLanguage = "";

/*
 * Mozilla throws an exception instead of returning null when the
 * response header is not there. This is a wrapper function that
 * handles the exception and returns null if necessary.
 */
ibm.portal.io.getResponseHeader = function (oXml, sHeader) {
  var result;
  try { result = oXml.getResponseHeader(sHeader); }
  catch (exc) { result = null; }
  return result;
}

ibm.portal.io.checkForError = function (oXml) {
    var sErrMsg = null;
    var sError = ibm.portal.io.getResponseHeader(oXml, "X-IBM-REST-Error");
    if(sError)
    {
        // sArg is the string argument to be replaced in the message if needed
        // Note that RFC 822 headers must contain only US-ASCII characters.
        // non US-ASCII characters must be encoded/decoded.
        var sErrMsg = '', sArg = ibm.portal.io.getResponseHeader(oXml, "X-IBM-REST-MsgArg1");
        if (sArg)
        {
            sErrMsg = ibm.portal.resource.getString(sError, window.decodeURIComponent(sArg));
        } else {
            sErrMsg = ibm.portal.resource.getString(sError);
        }
    }
    return sErrMsg;
}

/*
 * Makes an asynchronous request, allowing the calling object to react
 * at each stage. The calling object MUST define the functions startUpdate(),
 * handleError(), and handleData().
 */
ibm.portal.io.asyncRequest = function(method, sActionUrl, sParameter, oCaller) {
    try
    {
        ibm.portal.io.setAsync(method, sActionUrl, sParameter, oCaller, true);
    } catch(e) {
        ibm.portal.io.setAsync(method, sActionUrl, sParameter, oCaller, false);
    }
}

ibm.portal.io.setAsync = function(method, sUrl, sParameter, oCaller, bfirstime) {

    var oXml = ibm.portal.xml.getXmlHttpRequest();
    oXml.open(method, sUrl, true);

    if(! bfirstime)
    {
        oXml.setRequestHeader("Accept-language", "*");
    }

    try {
        oXml.onreadystatechange = function()  {
            if (oXml.readyState == 4)
            {
                //if (NG.ServerRequest.checkSessionExpiration(oXml))
                //{
                //    return;
                //}
                    oCaller.startUpdate(); //Let the caller know the request is done
                    sError = ibm.portal.io.checkForError(oXml);

                if (sError)
                {
                        oCaller.handleError(""+ibm.portal.resource.getString(sError));
                } else {
                    var data = oXml.responseText;
                    oCaller.handleData(data);
                    }
            }
        }
        oXml.send(sParameter);
    } catch(e) {
        throw new Error("", "");
    }
}


SemTagSvc.watchEvent(window, 'load', SemTagSvc.init, false);
