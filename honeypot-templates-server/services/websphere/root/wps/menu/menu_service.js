/* Copyright IBM Corp. 2002, 2005  All Rights Reserved.              */

//
// The MenuService is the JavaScript wrapper around a Java applet which
// displays pop-up menus. The applet is accessed via its ID "menuDisplayer".
//
// The main method of this class is:
// 		showMenu - displays a given menu at a given location
// The showMenu method also is passed a JavaScript object to which
// any notifications about menu selection will be routed; this object
// must implement a "menuItemSelected" method.
//
// See also the WPSMenu and WPSMenuContext classes

var CONST_USERAGT = navigator.userAgent.toLowerCase(); 
var CONST_IS_MAJOR = parseInt(navigator.appVersion); 
var CONST_IS_MINOR = parseFloat(navigator.appVersion); 
var CONST_IS_IE = ((CONST_USERAGT.indexOf("msie") != -1) && (CONST_USERAGT.indexOf("opera") == -1));
var CONST_IS_GECKO = (CONST_USERAGT.indexOf('gecko') != -1);
var CONST_IS_NAV  = ((CONST_USERAGT.indexOf('mozilla')!=-1) && (CONST_USERAGT.indexOf('spoofer')==-1) 
            && (CONST_USERAGT.indexOf('compatible') == -1) && (CONST_USERAGT.indexOf('opera')==-1) 
            && (CONST_USERAGT.indexOf('webtv')==-1) && (CONST_USERAGT.indexOf('hotjava')==-1));
var CONST_IS_NAV4 = (CONST_IS_NAV && (CONST_IS_MAJOR == 4));
var CONST_IS_OPERA = (CONST_USERAGT.indexOf("opera") != -1);

var scrollbackx = 0;
var scrollbacky = 0;

function MenuService() {
	// Central cache of parsed XML data objects.
	this.menuDefinitions = new Array();
}

function menuService_scrollback()
{
	if (CONST_IS_GECKO)
	{
		window.setTimeout('window.scrollTo(scrollbackx, scrollbacky)',35);
	}
}

MenuService.prototype.showMenu = menuService_showMenu;
MenuService.prototype.parseMenuData = menuService_parseMenuData;
MenuService.prototype.removeListener = menuService_removeListener;

// labels is String[], enablements is boolean[] (or null)
function menuService_showMenu(listener, labels, enablements, x, y) {
	// This object will be notified if any menu item is selected
	MenuService.currentClient = listener;
	var disp = menuService_getApplet();
	if(!disp)
		return;
	var enablementVector = disp.getNewVector();
	if (enablements != null) {
		for (var a=0; a < enablements.length; a++) 
			enablementVector.addElement(enablements[a] ? "true" : "false");
	}
	
	if (CONST_IS_GECKO)
	{
		scrollbackx = window.pageXOffset;
		scrollbacky = window.pageYOffset;
	}
	
	// Ask the applet to show the menu - menuService_callback will be
	// invoked if the user clicks something.
	disp.showMenu(menuService_convertArrayToVector(labels), enablementVector, "menuService_callback", x, y);
}

// Called by the applet if a menu item has been clicked.
function menuService_callback(ix) 
{
	if(MenuService.currentClient && MenuService.currentClient != null)
		MenuService.currentClient.menuItemSelected(ix);
}

// Called by container service to clear the current listener
function menuService_removeListener()
{
	MenuService.currentClient = null;
}

// Load an XML menu description from the given path
// handler is an event handler to be invoked 
function menuService_parseMenuData(xmlpath, handler) {
	// Use an associative array to minimize duplicate objects.
	var key = xmlpath.toLowerCase();
	var menudef = this.menuDefinitions[key];
	if (!menudef) {
		if(CONST_IS_IE)
		{
			menudef = new ActiveXObject("microsoft.xmldom");
			this.menuDefinitions[key] = menudef;
			menudef.async = false;
		}
		else if(document.implementation)
		{
			menudef = document.implementation.createDocument("", "", null);
			if(handler)
				menudef.addEventListener('load', handler, false);
		}
		if(menudef)
			menudef.load(xmlpath);
	}
	return menudef;	
}

// These are needed because passing an array from JavaScript to Java or vice-versa
// is broken in IE.

function menuService_convertArrayToVector(theArray) {
	// get a fresh Vector - will use it to pass JavaScript arrays to Java
	var disp = menuService_getApplet();
	if(!disp)
		return null;
	var vector = disp.getNewVector();
	for (var x=0; x < theArray.length; x++) 
		vector.addElement(theArray[x]);
	return vector;
}

// write out the html for the menu displayer applet
function menuService_writeApplet(codebase)
{
	if(window.menuServiceWroteApplet) 
		return;
	var disp = menuService_getApplet();
	if(disp)
	{
		window.menuServiceWroteApplet = true;
		return;
	}
	var base = codebase ? codebase : "/wps/menu";
	var style = CONST_IS_IE ? "style='position:absolute;left:0;top:0;visibility:hidden'" : "";
	var str= "<applet code='com.lotus.cs.menu.MenuDisplayer' id='menuDisplayer' name='menuDisplayer'";
	str += "height='0' width='0' mayscript='true' archive='menudisp.jar' codebase='" + base + "'" + style + ">";
	str += "<param name='useslibrary' value='LotusMenu'>";
	str += "<param name='useslibrarycodebase' value='menudisp.cab'>";
	str += "<param name='useslibraryversion' value='1,0,0,1'>";
	str += "<param name='namespace' value='LotusMenu'></applet>";
	if(CONST_IS_NAV4)
		document.writeln("<layer name='MenuDisplayerLayer'>" + str + "</layer>");
	else if(CONST_IS_IE)
		document.writeln("<span style='position:absolute;left:0;top:0;visibility:hidden'>" + str+ "</span>");
	else if(CONST_IS_GECKO)
		document.writeln("<span style='position:fixed;left:0;top:0;visibility:hidden'>" + str+ "</span>");
	else
		document.writeln(str);
	 window.menuServiceWroteApplet = true;
}

function menuService_getApplet()
{
	var disp = document.applets["menuDisplayer"];
	if(disp)
		return disp;
	else if(CONST_IS_NAV4)
	{
		var menulayer = document.layers["MenuDisplayerLayer"];
		if(menulayer)
			disp = menulayer.document.applets["menuDisplayer"];
	}
	else if (document.getElementById)
		disp = document.getElementById("menuDisplayer"); 
	return disp;
}

// Translate the given point to be relative to the parent window.
function menuService_translateToAbsolute(pt) {
	var dx = CONST_IS_IE ? (window.parent.screenLeft - window.screenLeft) : 0;
	var dy = CONST_IS_IE ? (window.parent.screenTop - window.screenTop) : 0;
	pt.x -= dx;
	pt.y -= dy;
}

function WPSMenu(menuid)
{
	this.id = menuid;
	this.items = new Array();
	this.actions = new Array();
	this.enable = new Array();
}

// label - label of the menu item
// action -	javascript expression executed when menu item is selected 
WPSMenu.prototype.addMenuItem = function (label, action, enable) {
	this.items[this.items.length] = label;
	this.actions[this.actions.length] = action;
	this.enable[this.enable.length] = enable ? enable : null;
}

WPSMenu.prototype.addMenuSeparator = function () {
	this.addMenuItem("-");
}

// initialize a menu context with the given client properties
function WPSMenuContext(props)
{
	if (props)
	{
		for (var p in props) 
			this[p] = props[p];
	}
	this.menuHandler = new Array();
}

WPSMenuContext.prototype.addMenu = function (menu, handler)
{
	var h = handler ? handler : new WPSMenuHandler(menu, this);
	this.setHandler(menu.id, h);
}

WPSMenuContext.prototype.showMenu = function (menuid, event) 
{
	var x = 10;
	var y = 10;
	if (typeof event == "undefined")
		event = window.event;
	if(typeof event != "undefined")
	{
		x = event.clientX ? event.clientX : event.pageX;
		y = event.clientY ? event.clientY : event.pageY;
	}
	this.showMenuAt(menuid, x, y);
}

WPSMenuContext.prototype.showMenuAt = function (menuid, x, y) 
{
	if(menuid)
	{
		var handler = this.menuHandler[menuid];
		if(handler)
			menuService_showMenu(handler, handler.getLabels(), handler.getEnableArray(), x, y);
	}
}

WPSMenuContext.prototype.setHandler = function (menuid, handler)
{
	this.menuHandler[menuid] = handler;
}

// An object that knows how to handle menu
function WPSMenuHandler(menu, context)
{
	this.menu = menu;
	this.context = context;
}

WPSMenuHandler.prototype.getLabels = function() 
{
	return this.menu.items;
}

WPSMenuHandler.prototype.getEnableArray = function() 
{
	var menuEnable = new Array();
	for(var i=0; i < this.menu.enable.length; i++)
		menuEnable[i] = this.menu.enable[i] ? eval(this.menu.enable[i]) : true;
	return menuEnable;
}

WPSMenuHandler.prototype.menuItemSelected = function(ix) 
{
	if(!this.menu.actions[ix]) return;
	var act = this.menu.actions[ix];
	// Evaluate the function as a method on our context, if such a method exists.
	if (this.context && this.context[act])
	{
		if(CONST_IS_IE)
			(this.context[act])();
		else
			eval("this.context." + act + "()");
	}
	else
		eval(act);
}

// load and initialize a menucontext to handle menus from the given path
// this is only supported on MSIE 5+ and Mozilla
function wpsMenuContext_initialize(xmlpath, props) 
{
	var mc = new WPSMenuContext(props);
	var menusvc = getMenuService();
	var pmenus = menusvc.parseMenuData(xmlpath);
	if(pmenus == null)
		return mc;
	var menuNodes = pmenus.getElementsByTagName("menu");
	for(var n = 0; n < menuNodes.length; n++)
	{
		var mNode = menuNodes.item(n);
		var idNode = mNode.attributes.getNamedItem("id");
		if(idNode != null)
		{
			var themenu = new WPSMenu(idNode.nodeValue);
			var itemnodes = idNode.getElementsByTagName("menuitem");
			for (var i = 0; i < itemnodes.length; i++)
			{
				var menuitem = itemnodes.item(i);
				var label = menuitem.attributes.getNamedItem("label").nodeValue;
				var hideWhenNode = menuitem.attributes.getNamedItem("hidewhen");
				if (hideWhenNode && eval(hideWhenNode.nodeValue))
					continue;
				var actionNode = menuitem.attributes.getNamedItem("action");
				themenu.addMenuItem(label, actionNode);
			}
			mc.addMenu(themenu, mc);
		}
	}
	return mc;
}

function showPortalMenu(menuid, event, context) 
{
	if(context)
		context.showMenu(menuid, event);
}

// Get the pixel position of an HTML element relative to the body.
function kwFindAbsoluteTop(obj) 
{
	var top = 0;
	var theobj = obj;
	var thebody = window.document.body;
	while (theobj && theobj != null && theobj != thebody) {
		top += theobj.offsetTop;
		theobj = theobj.offsetParent;
	}
	return top;
}

// Get the pixel position of an HTML element relative to the body.
function kwFindAbsoluteLeft(obj) 
{
	var left = 0;
	var indexobj = obj;
	var thebody = window.document.body;
	while (indexobj && indexobj != null && indexobj != thebody) {
		left += indexobj.offsetLeft;
		indexobj = indexobj.offsetParent;
	}
	return left;
}

if (window.ServiceRegistry)
	ServiceRegistry.registerService("MenuService", "MenuService");

var menuSvcInst;

// Retrieve the single instance of this service. 
function getMenuService()
{	
	if(menuSvcInst)
		return menuSvcInst;
	if(window.ServiceRegistry)
		menuSvcInst = window.ServiceRegistry.getService("MenuService");
	if(!menuSvcInst)
		menuSvcInst = new MenuService();
	return menuSvcInst;
}
