 
 
 
 
 

 




 
/*********************************************************** {COPYRIGHT-TOP} ***
* Licensed Materials - Property of IBM
* Tivoli Presentation Services
*
* (C) Copyright IBM Corp. 2002,2003 All Rights Reserved.
*
* US Government Users Restricted Rights - Use, duplication or
* disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
************************************************************ {COPYRIGHT-END} ***
* Change Activity on 6/20/03 version 1.17:
* @00=WCL, V3R0, 04/14/2002, JCP: Initial version
* @01=D96484, V3R2, 06/14/2002, bcourt: hide select/iframe elements
* @02=D99067, V3R2, 06/25/2002, bcourt: hide listbox scrollbar
* @03=D97043, V3R3, 09/03/2002, JCP: fix launch menu item on linux NS6
* @04=D104656, V3R3, 09/16/2002, JCP: form submit instead of triggers, mozilla compatibility
* @05=D107029, V3R4, 12/03/2002, Mark Rebuck:  Added support for timed menu hiding
* @06=D110173, V3R4, 03/24/2003, JCP: selection sometimes gets stuck
* @07=D113641, V3R4, 04/29/2003, LSR: Requirement #258 Shorten CSS Names
* @08=D113626, V3R4, 06/20/2003, JCP: clicking on text doesn't launch action on linux Moz13
*******************************************************************************/

var visibleMenu_ = null;
var padding_ = 10;

var transImg_ = "transparent.gif";

var arrowNorm_ = "contextArrowDefault.gif";
var arrowSel_ = "contextArrowSelected.gif";
var arrowDis_ = "contextArrowDisabled.gif";
var launchNorm_ = "contextLauncherDefault.gif";
var launchSel_ = "contextLauncherSelected.gif";

var arrowNormRTL_ = "contextArrowDefault.gif";
var arrowSelRTL_ = "contextArrowSelected.gif";
var arrowDisRTL_ = "contextArrowDisabled.gif";
var launchNormRTL_ = "contextLauncherDefault.gif";
var launchSelRTL_ = "contextLauncherSelected.gif";

var wclIsOpera_ = /Opera/.test(navigator.userAgent);

//ARC CHANGES FOR SPECIFYING STYLES - BEGIN
var defaultContextMenuBorderStyle_ = "lwpShadowBorder";
var defaultContextMenuTableStyle_ = "lwpBorderAll";
//ARC CHANGES FOR SPECIFYING STYLES - END

var arrowWidth_ = "12";
var arrowHeight_ = "12";

var submenuAltText_ = "+";

//ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - BEGIN
var defaultNoActionsText_ = "(0)";
var defaultNoActionsTextStyle_ = "lwpMenuItemDisabled";
//ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - END

var hideCurrentMenuTimer_ = null;

var onmousedown_ = document.onmousedown;

function clearMenuTimer( ) { //@05
   if (null != hideCurrentMenuTimer_) {
      clearTimeout( hideCurrentMenuTimer_ );
      hideCurrentMenuTimer_ = null;
   }
}

function setMenuTimer( ) { // @05
   clearMenuTimer( );
   hideCurrentMenuTimer_ = setTimeout( 'hideCurrentContextMenu( )', 2000);
}

function debug( str ) {
   /*
   if ( xbDEBUG != null ) {
      xbDEBUG.dump( str );
   }
   */
}

// constructor
function UilContextMenu( name, isLTR, width, borderStyle, tableStyle, emptyMenuText, emptyMenuTextStyle, positionUnder ) {
   // member variables
   this.name = name;
   this.items = new Array();
   this.isVisible = false;
   this.isDismissable = true;
   this.selectedItem = null;
   this.isDynamic = false;
   this.isCacheable = false;
   this.isEmpty = true;
   this.isLTR = isLTR;
   this.hiddenItems = new Array(); //@01A
   this.isHyperlinkChild = true; //  We will reset later if needed.
   
   this.bottomPositioned = positionUnder;

   // html variables
   this.launcher = null;
   this.menuTag = null;

   //ARC CHANGES FOR SPECIFYING STYLES - BEGIN
   //styles for menu
   if ( borderStyle != null )
   {
       this.menuBorderStyle = borderStyle;
   }
   else
   {
       this.menuBorderStyle = defaultContextMenuBorderStyle_;
   }
   
   if ( tableStyle != null )
   {
       this.menuTableStyle = tableStyle;
   }
   else
   {
       this.menuTableStyle = defaultContextMenuTableStyle_;
   }
   //ARC CHANGES FOR SPECIFYING STYLES - END

   //ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - BEGIN
   if ( emptyMenuText != null )
   {
   	   this.noActionsText = emptyMenuText;
   }
   else
   {
   	   this.noActionsText = defaultNoActionsText_;
   }
   
   if ( emptyMenuTextStyle != null )
   {
   	   this.noActionsTextStyle = emptyMenuTextStyle;
   }
   else
   {
   	   this.noActionsTextStyle = defaultNoActionsTextStyle_;
   }
   //ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - END

   // external methods
   this.add = UilContextMenuAdd;
   this.addSeparator = UilContextMenuAddSeparator;
   this.show = UilContextMenuShow;
   this.hide = UilContextMenuHide;

   // internal methods
   this.create = UilContextMenuCreate;
   this.getMenuItem = UilContextMenuGetMenuItem;
   this.getSelectedItem = UilContextMenuGetSelectedItem;

   if ( this.name == null ) {
      this.name = "UilContextMenu_" + allMenus_.length;
   }
}

// adds a menu item to the context menu
function UilContextMenuAdd( item ) {
   this.items[ this.items.length ] = item;
   this.isEmpty = false;
}

function UilContextMenuAddSeparator() {
   var sep = new UilMenuItem();
   sep.isSeparator = true;
   this.add( sep );
}

// shows the context menu
// launcher- html element (anchor) that is launching the menu
// launchItem- menu item that is launching the menu
function UilContextMenuShow( launcher, launchItem ) {
   if ( this.items.length == 0 ) {
      // empty context menu
      debug( 'menu is empty!' );
      //ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - BEGIN
      this.add( new UilMenuItem( this.noActionsText, false, "javascript:void(0);", null, null, null, null, this.noActionsTextStyle ) );
      //ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - END
      this.isEmpty = true;
   }

   if ( this.menuTag == null ) {
      // create the context menu html
      this.create();
   } else {
      this.menuTag.style.left = ""; //196195 //Reset
      this.menuTag.style.top = ""; //196195 //Reset
      this.menuTag.style.width = ""; //"0px"; //196195 //Reset
      this.menuTag.style.height = ""; //196195 //Reset
      this.menuTag.style.overflow = "visible"; //196195 //Reset, No horizontal and vertical scrollbars
   }

   if ( this.menuTag != null) {
      // store the launcher for later
      this.launcher = launcher;
      if ( this.launcher.tagName == "IMG" ) {
         this.isHyperlinkChild = false;

         // we want the anchor tag
         this.launcher = this.launcher.parentNode;
      }

      // boundaries of window
      var bd = new ContextMenuBrowserDimensions();
      var maxX = bd.getScrollFromLeft() + bd.getViewableAreaWidth();
      var maxY = bd.getScrollFromTop() + bd.getViewableAreaHeight();
      var minX = bd.getScrollFromLeft();
      var minY = bd.getScrollFromTop();

      debug( 'max: ' + maxX + ', ' + maxY );

      var menuWidth = getWidth( this.menuTag );
      var menuHeight = getHeight( this.menuTag );

      // move the context menu to the right of the launcher
      var posX = 0;
      var posY = 0;
      var fUseUpperY = false; //196195
      var maxUpperPosY = 0; //196195

      if ( launchItem != null ) {
         // launched from submenu
         var launchTag = launchItem.itemTag;
         var launchTagWidth = getWidth( launchTag );
         var parentTag = launchItem.parentMenu.menuTag; //@04A
         var launchOffsetX = getLeft( parentTag ); //@04C
         var launchOffsetY = getTop( parentTag ); //@04C

         posX = launchOffsetX + getLeft( launchTag ) + launchTagWidth; //@04C
         posY = launchOffsetY + getTop( launchTag ); //@04C

         if ( !this.isLTR ) {
            posX -= launchTagWidth;
            posX -= menuWidth;
         }

         // try to keep it in the window
         if ( this.isLTR ) {
            if ( posX + menuWidth > maxX ) {
               // try to show it to the left of the parent menu
               var posX1 = launchOffsetX - menuWidth;
               var posX2 = maxX - menuWidth;
               if ( 0 <= posX1 ) {
                  posX = posX1;
               }
               else {
                  posX = Math.max( minX, posX2 );
               }
            }
         }
         else {
            if ( posX < 0 ) {
               // try to show it to the right of the parent menu
               var posX1 = launchOffsetX + launchTagWidth;
               if ( posX1 + menuWidth < maxX ) {
                  posX = posX1;
               }
               else {
                  posX = Math.min( maxX, maxX - menuWidth );
               }
            }
         }

         if ( posY + menuHeight > maxY ) {
            var posY1 = maxY - menuHeight;
            posY = Math.max( minY, posY1 );
         }
      }
      else {
         // launched from menu link
         var launcherLeft = getLeft( this.launcher, true )
         if ( this.launcher.tagName == "BUTTON" || this.bottomPositioned ) {
			
             posX = launcherLeft;
             
             // bidi
             if ( !this.isLTR ) {
                 //196195 posX += getWidth( this.launcher ) - getWidth( this.menuTag );
                 posX += getWidth( this.launcher ) - menuWidth; //196195
             }

             if (this.isLTR) {
                 if ((posX + menuWidth) > maxX) {
                     //196195 begins
                     if ((posX + getWidth(this.launcher)) > maxX) {
                         posX = Math.max(minX, maxX - menuWidth);
                     }
                     else 
                     //196195 ends
                         posX = Math.max(minX, posX + getWidth( this.launcher ) - menuWidth);
                 }
                 //196195 begins 
                 else if (posX < minX) {
                     posX = minX;
                 }
                 //196195 ends 
             }
             else{
                 if (posX < minX) {
                     //196195 if ((launcherLeft + menuWidth) < maxX) {
                     if ((launcherLeft > minX) && ((launcherLeft + menuWidth) < maxX)) { //196195
                         posX = launcherLeft;
                     }
                     else{
                         posX = Math.min(minX, maxX - menuWidth);
                     }
                 }
                 //196195 begins
                 else if ( (posX + menuWidth) > maxX) {
                     if (Math.min(posX, maxX - menuWidth) >= minX)
                         posX = Math.min(posX, maxX - menuWidth);
                 }
                 //196195 ends  
             }
             
             maxUpperPosY = getTop( this.launcher, true ); //196195
             var upperVisibleHeight = maxUpperPosY - minY; //196195
             posY = getTop( this.launcher, true ) + getHeight( this.launcher );
             var lowerVisibleHeight = maxY - posY; //196195
             //196195 if ( posY + menuHeight > maxY ) {
             if ( (posY + menuHeight > maxY) && (lowerVisibleHeight < upperVisibleHeight) ) { //196195
                // top
                posY -= (menuHeight + getHeight( this.launcher ));
                fUseUpperY = true; //196195
             }

             if ( posY < minY ) {
                posY = minY;
             }
         }
         else {

             // left-right
             posX = launcherLeft + this.launcher.offsetWidth;
             posY = getTop( this.launcher, true );

             if ( !this.isLTR ) {
                posX -= this.launcher.offsetWidth;
                posX -= menuWidth;
             }

             // keep it in the window
             if ( this.isLTR ) {
                if ( posX + menuWidth > maxX ) {
                   // try to show it on the left side of the launcher
                   var posX1 = launcherLeft - menuWidth;
                   if ( posX1 > 0 ) {
                      posX = posX1;
                   }
                   else {
                      posX = Math.max( minX, maxX - menuWidth );
                   }
                }
             }
             else {
                if ( posX < minX ) {
                   // try to show it on the right side of the launcher
                   var posX1 = launcherLeft + this.launcher.offsetWidth;
                   if ( posX1 + menuWidth < maxX ) {
                      posX = posX1;
                   }
                   else {
                      posX = Math.min( minX, maxX - menuWidth );
                   }
                }
             }

             if ( posY + menuHeight > maxY ) {
                posY = Math.max( minY, maxY - menuHeight );
             }
         }
         if ( ((posX + menuWidth) > maxX) ||
              (((posY + menuHeight) > maxY) && (fUseUpperY == false)) || 
              (((posY + menuHeight) > maxUpperPosY) && (fUseUpperY == true)) ) {
             if (posX + menuWidth > maxX) {
                 this.menuTag.style.width = (maxX - posX) + "px";
             }
             else{
                 this.menuTag.style.width = menuWidth + "px";
             }

             if (fUseUpperY == false) {
                 if (posY + menuHeight > maxY) {
                     this.menuTag.style.height = (maxY - posY) + "px";
                 }
                 else {
                     this.menuTag.style.height = menuHeight + "px";
                 }
             } else {
                 if (posY + menuHeight > maxUpperPosY) {
                     this.menuTag.style.height = (maxUpperPosY - posY) + "px";
                 }
                 else {
                     this.menuTag.style.height = menuHeight + "px";
                 }
             }

             this.menuTag.style.overflow = "auto";
         } else { //196195 begins
             this.menuTag.style.width = menuWidth + "px";
             this.menuTag.style.height = menuHeight + "px";
             this.menuTag.style.overflow = "visible"; //196195
         } //196196 ends
      }

      debug( 'show ' + this.name + ': ' + posX + ', ' + posY );
      this.menuTag.style.left = posX + "px";
      this.menuTag.style.top = posY + "px";

      // make the context menu visible
      this.menuTag.style.visibility = "visible";
      this.isVisible = true;

      // set focus on the first menu item
      this.items[0].setSelected( true );
      this.items[0].anchorTag.focus();

	  /*
	  // no longer needed since fixed in Opera 9, and no other non-IE browsers need this
      // @01A - Hide any items that intersect this menu
      var coll = document.getElementsByTagName("SELECT");
      if (coll!=null)
      {
         for (i=0; i<coll.length; i++)
         {
            //Hide the element
            if (intersect(this.menuTag,coll[i]) == true ) {
               if (coll[i].style.visibility != "collapse") //@02C
               {
                  coll[i].style.visibility = "collapse"; //@02C
                  this.hiddenItems.push(coll[i]);
               }
            }
         }
      }
      coll = document.getElementsByTagName("IFRAME");
      if (coll!=null)
      {
         for (i=0; i<coll.length; i++)
         {
            //Hide the element
            if (intersect(this.menuTag,coll[i]) == true ) {
               if (coll[i].style.visibility != "hidden")
               {
                  coll[i].style.visibility = "hidden";
                  this.hiddenItems.push(coll[i]);
               }
            }
         }
      }
	  */
      // save old & set new hide action for this menu     
      onmousedown_ = document.onmousedown;
      document.onmousedown = hideCurrentContextMenu;
   }
}

// Test whether two objects overlap
function intersect(obj1 , obj2) //@01A
{
   var left1 =  parseInt(document.defaultView.getComputedStyle(obj1, '').getPropertyValue("left"));
   var right1 = left1 + parseInt(document.defaultView.getComputedStyle(obj1, '').getPropertyValue("width"));
   var top1 = parseInt(document.defaultView.getComputedStyle(obj1, '').getPropertyValue("top"));
   var bottom1 = top1 + parseInt(document.defaultView.getComputedStyle(obj1, '').getPropertyValue("height"));

   var left2 =  parseInt(document.defaultView.getComputedStyle(obj2, '').getPropertyValue("left"));
   var right2 = left2 + parseInt(document.defaultView.getComputedStyle(obj2, '').getPropertyValue("width"));
   var top2 = parseInt(document.defaultView.getComputedStyle(obj2, '').getPropertyValue("top"));
   var bottom2 = top2 + parseInt(document.defaultView.getComputedStyle(obj2, '').getPropertyValue("height"));

    //alert("Comparing: " +left1 + ", " + right1+ ", " +top1 + ", " + bottom1+ " to "  +left2 + ", "  +right2 + ", " +top2 + ", " +bottom2);
   if (lineIntersect(left1,right1, left2,right2)== true &&
       lineIntersect(top1,bottom1,top2,bottom2) == true) {
      return true;
   }
   return false;
}

// Test whether the two line segments a--b and c--d intersect.
function lineIntersect(a, b, c, d) //@01A
{
   //alert (a+"--"+b + "   " + c + "--" + d);
   //Assume that a < b && c < d
   if ( (a <= c  && c <= b) ||
        (a <= d && d <= b) ||
        (c <= a && d >= b ) )
   {
      return true;
   } else {
      return false;
   }
}

// hides the context menu
function UilContextMenuHide() {
   if ( this.menuTag != null ) {
      debug( 'hide ' + this.name );

      // hide any visible submenus first
      for ( var i=0; i<this.items.length; i++ ) {
         if ( this.items[i].submenu != null &&
              this.items[i].submenu.isVisible ) {
            this.items[i].submenu.hide();
         }
      }

      // clear selection
      if ( this.selectedItem != null ) {
         this.selectedItem.setSelected( false );
      }

      // make the context menu hidden
      this.menuTag.style.visibility = "hidden";
      this.isVisible = false;
      this.isDismissable = true;

      // @01A - Show any items that were hidden by this menu
      var itemCount = this.hiddenItems.length;
      for (i=0; i<itemCount; i++)
      {
         var item = this.hiddenItems.pop();
         item.style.visibility = "visible";
      }

      // clear the launcher
      this.launcher = null;
      
      // reset action      
      document.onmousedown = onmousedown_;
   }
}

// creates the context menu html element
function UilContextMenuCreate( recurse ) {
   if ( this.menuTag == null ) {
      this.menuTag = document.createElement( "DIV" );
      this.menuTag.style.position = "absolute";
      this.menuTag.style.cursor = "default";
      this.menuTag.style.visibility = "hidden";
      // following line does not work on Mozilla. SPR #PDIK66SJZ4
      //this.menuTag.style.width = "0px"; // this causes dynamic sizing
      //if (!this.isLTR) this.menuTag.dir = "RTL"; //196195
      this.menuTag.onmouseover = contextMenuDismissDisable;
      this.menuTag.onmouseout = contextMenuDismissEnable;
      this.menuTag.oncontextmenu = contextMenuOnContextMenu;

      var numItems = this.items.length;

      // check if this context menu contains icons or submenus
      var hasIcon = false;
      var hasSubmenu = false;
      for ( var i=0; i<numItems; i++ ) {
         if ( !this.items[i].isSeparator ) {
            if ( !hasSubmenu && this.items[i].submenu != null ) {
               hasSubmenu = true;
            }
            if ( !hasIcon && this.items[i].icon != null ) {
               hasIcon = true;
            }
            if ( hasSubmenu && hasIcon ) {
               break;
            }
         }
      }

      // create the menu items
      for ( var i=0; i<numItems; i++ ) {
         this.items[i].isFirst = ( i == 0 );
         this.items[i].isLast = ( i+1 == numItems );
         this.items[i].parentMenu = this;
         this.items[i].create( hasIcon, hasSubmenu );
      }

      // create the context menu border
      var border = document.createElement( "TABLE" );
      if (!this.isLTR) border.dir = "RTL";
      //border.className = "wclPopupMenuBorder"; //@04D
      border.rules = "none";
      border.cellPadding = 0;
      border.cellSpacing = 0;
      //border.width = "100%";
      border.border = 0;
      var borderBody = document.createElement( "TBODY" );
      var borderRow = document.createElement( "TR" );
      var borderData = document.createElement( "TD" );
      var borderDiv = document.createElement( "DIV" ); //@04A
      //borderDiv.className = "pop2"; //@04A   //@06C1
      
      //ARC CHANGES FOR SPECIFYING STYLES - BEGIN
      borderDiv.className = this.menuBorderStyle; 
      //ARC CHANGES FOR SPECIFYING STYLES - END

      borderData.appendChild( borderDiv ); //@04A
      borderRow.appendChild( borderData );
      borderBody.appendChild( borderRow );
      border.appendChild( borderBody );

      // create the context menu
      var table = document.createElement( "TABLE" );
      if (!this.isLTR) table.dir = "RTL";
      //table.className = "wclPopupMenu"; //@04D
      table.rules = "none";
      table.cellPadding = 0;
      table.cellSpacing = 0;
      table.width = "100%";
      table.border = 0;

      // add the menu items
      var tableBody = document.createElement( "TBODY" );
      table.appendChild( tableBody );

      // @04A - create another table for mozilla fix
      // (border styles not hidden correctly if set on table tag)
      var table2 = document.createElement( "TABLE" );
      if (!this.isLTR) table2.dir = "RTL";
      table2.rules = "none";
      table2.cellPadding = 0;
      table2.cellSpacing = 0;
      table2.width = "100%";
      table2.border = 0;
      var tableRow = document.createElement( "TR" );
      var tableData = document.createElement( "TD" );
      var tableDiv = document.createElement( "DIV" );
      //tableDiv.className = "pop1";     //@06C1

      //ARC CHANGES FOR SPECIFYING STYLES - BEGIN
      tableDiv.className = this.menuTableStyle;     //@06C1
      //ARC CHANGES FOR SPECIFYING STYLES - END

      var tableBody2 = document.createElement( "TBODY" );
      tableBody.appendChild( tableRow );
      tableRow.appendChild( tableData );
      tableData.appendChild( tableDiv );
      tableDiv.appendChild( table2 );
      table2.appendChild( tableBody2 );

      for ( var i=0; i<numItems; i++ ) {
         if ( this.items[i].isSeparator ) {
            this.items[i].createSeparator( tableBody2, hasSubmenu ); //@04C
         }
         else {
            tableBody2.appendChild( this.items[i].itemTag ); //@04C
         }
      }

      borderDiv.appendChild( table ); //@04C
      this.menuTag.appendChild( border );

      // add to document
      document.body.appendChild( this.menuTag );
   }

   if ( recurse ) {
      // this is used when cloning dynamic menus
      for ( var i=0; i<this.items.length; i++ ) {
         if ( this.items[i].submenu != null ) {
            this.items[i].submenu.create( recurse );
         }
      }
   }
}

// returns the menu item object associated with the html element
function UilContextMenuGetMenuItem( htmlElement ) {
   if ( htmlElement != null ) {
      if ( htmlElement.nodeType == 3 ) { //@06A
          // text node
          htmlElement = htmlElement.parentNode; //@06A
      }
      var tagName = htmlElement.tagName;
      var menuItemTag = null;
      if ( tagName == "IMG" ||
           tagName == "A" ) {
         menuItemTag = htmlElement.parentNode.parentNode;
      }
      else if ( tagName == "TD" ) {
         menuItemTag = htmlElement.parentNode;
      }
      for ( var i=0; i<this.items.length; i++ ) {
         if ( this.items[i].itemTag != null &&
              this.items[i].itemTag == menuItemTag ) {
            // found the item
            return this.items[i];
         }
         else if ( this.items[i].submenu != null &&
                   this.items[i].submenu.isVisible ) {
            // recurse through any visible submenus
            var item = this.items[i].submenu.getMenuItem( htmlElement );
            if ( item != null ) {
               // found the item
               return item;
            }
         }
      }
   }
   return null;
}

// returns the deepest selected menu item
function UilContextMenuGetSelectedItem() {
   var item = this.selectedItem;
   if ( item != null && item.submenu != null && item.submenu.isVisible ) {
      // recurse through the item's visible submenu
      return item.submenu.getSelectedItem();
   }
   return item;
}

// method called by an event handler (onmouseover for context menu div)
function contextMenuDismissEnable() {
   if ( visibleMenu_ != null ) {
      visibleMenu_.isDismissable = true;
      if (visibleMenu_.isHyperlinkChild) {
         setMenuTimer( ); // @05
      }
   }
}

// method called by an event handler (onmouseout for context menu div)
function contextMenuDismissDisable() {
   if ( visibleMenu_ != null ) {
      visibleMenu_.isDismissable = false;
      clearMenuTimer( ); // @05
   }
}

// method called by an event handler (oncontextmenu for context menu div)
function contextMenuOnContextMenu() {
   return false;
}

// constructor
function UilMenuItem( text, enabled, action, clientAction, submenu, icon, defItem,menuStyle, selectedMenuStyle ) {
   // member variables
   this.text = text;
   this.icon = icon;
   this.action = action;
   this.clientAction = clientAction;
   this.submenu = submenu;
   this.isSeparator = false;
   this.isSelected = false;
   this.isEnabled = ( enabled != null ) ? enabled : true;
   this.isDefault = ( defItem != null ) ? defItem : false;
   this.isFirst = false;
   this.isLast = false;
   this.parentMenu = null;
   
   if ( menuStyle != null ) {
      this.menuStyle = menuStyle; 
   }
   else {
      this.menuStyle = ( this.isEnabled ) ? "lwpMenuItem" : "lwpMenuItemDisabled"; 
   }

   this.selectedMenuStyle = ( selectedMenuStyle != null) ? selectedMenuStyle : "lwpSelectedMenuItem";

   // html variables
   this.itemTag = null;
   this.anchorTag = null;
   this.arrowTag = null;

   // internal methods
   this.create = UilMenuItemCreate;
   this.createSeparator = UilMenuItemCreateSeparator;
   this.setSelected = UilMenuItemSetSelected;
   this.updateStyle = UilMenuItemUpdateStyle;
   this.getNextItem = UilMenuItemGetNextItem;
   this.getPrevItem = UilMenuItemGetPrevItem;

   if ( this.submenu != null ) {
      // menu items with submenus do not have actions
      this.action = null;
   }
}

// creates the menu item html element
function UilMenuItemCreate( menuHasIcon, menuHasSubmenu ) {
   if ( !this.isSeparator ) {
      this.anchorTag = document.createElement( "A" );
      if ( this.action != null ) {
         this.anchorTag.href = "javascript:menuItemLaunchAction();";
         if ( this.clientAction != null && !wclIsOpera_ ) {
            this.anchorTag.onclick = this.clientAction;
         }
      }
      else if ( this.submenu != null ) {
         this.anchorTag.href = "javascript:void(0);";
         this.anchorTag.onclick = menuItemShowSubmenu;
      }
      else if ( this.clientAction != null ) {
         this.anchorTag.href = "javascript:menuItemLaunchAction();";
         if(!wclIsOpera_) this.anchorTag.onclick = this.clientAction;
      }
      this.anchorTag.onfocus = menuItemFocus;
      this.anchorTag.onblur = menuItemBlur;
      this.anchorTag.onkeydown = menuItemKeyDown;
      this.anchorTag.innerHTML = this.text;
//      this.anchorTag.title = this.text; 
      this.anchorTag.title = this.anchorTag.innerHTML; 
      //this.anchorTag.className = "pop5";    //@06C1
      this.anchorTag.className = this.menuStyle;    //@06C1
      if ( this.isDefault ) {
         this.anchorTag.style.fontWeight = "bold";
      }

      var td = document.createElement( "TD" );
      td.noWrap = true;
      td.style.padding = "3px";
      td.appendChild( this.anchorTag );

      // left padding or icon
      var leftPad = document.createElement( "TD" );
      leftPad.noWrap = true;
      leftPad.innerHTML = "&nbsp;";
      leftPad.style.padding = "3px";
      if ( this.icon != null ) {
         var imgTag = document.createElement( "IMG" );
         imgTag.src = this.icon;
         if (imgTag.src == "" || imgTag.src == null) {
             var imgTag1 = "<img src=\"" + this.icon + "\"";
             if ( this.text != null ) {
                imgTag1 += " alt=\'" + this.text + "\'";
                imgTag1 += " title=\'" + this.text + "\'";
             }
             imgTag1 += "/>";
             leftPad.innerHTML = imgTag1;
         } else {
             if ( this.text != null ) {
                imgTag.alt = this.text;
                imgTag.title = this.text;
             }
             leftPad.appendChild( imgTag );
         }
      }
      else {
         leftPad.width = padding_;
      }

      // right padding
      var rightPad = document.createElement( "TD" );
      rightPad.noWrap = true;
      rightPad.width = padding_;
      rightPad.innerHTML = "&nbsp;";
      rightPad.style.padding = "3px";

      this.itemTag = document.createElement( "TR" );
      this.itemTag.onmousemove = menuItemMouseMove;
      this.itemTag.onmousedown = menuItemLaunchAction;
      //this.itemTag.className = "pop5";  //@06C1
      this.itemTag.className = this.menuStyle;
      // put together the table row
      this.itemTag.appendChild( leftPad );
      this.itemTag.appendChild( td );
      this.itemTag.appendChild( rightPad );
      if ( menuHasSubmenu ) {
         // submenu arrow
         var submenuArrow = document.createElement( "TD" );
         submenuArrow.noWrap = true;
         submenuArrow.style.padding = "3px";
         if ( this.submenu != null ) {
            var submenuImg = document.createElement( "IMG" );
            submenuImg.alt = submenuAltText_;
            submenuImg.title = submenuAltText_;
            submenuImg.width = arrowWidth_;
            submenuImg.height = arrowHeight_;
            if (this.parentMenu.isLTR) submenuImg.src = arrowNorm_;
            else submenuImg.src = arrowNormRTL_;
            submenuArrow.appendChild( submenuImg );
            this.arrowTag = submenuImg;
         }
         else {
            submenuArrow.innerHTML = "&nbsp;";
         }
         this.itemTag.appendChild( submenuArrow );
      }

      // update the style of the menu item
      this.updateStyle( this.itemTag );
   }
}

// create the context menu separator html elements
function UilMenuItemCreateSeparator( tableBody, menuHasSubmenu ) {
   // create the context menu separator
   var numCols = ( menuHasSubmenu ) ? 4 : 3;

   for ( var i=0; i<4; i++ ) {
      var tr = document.createElement( "TR" );
      if ( i == 1 ) {
         //tr.className = "pop3";   //@06C1
         tr.className = "portlet-separator";   //@06C1
      }
      else if ( i == 2 ) {
         //tr.className = "pop4";   //@06C1
         tr.className = "lwpMenuBackground";   //@06C1
      }
      else {
         //tr.className = "pop5";   //@06C1
         tr.className = "lwpMenuItem";   //@06C1
      }

      var td = document.createElement( "TD" );
      td.noWrap = true;
      td.width = "100%";
      td.height = "1px";
      td.colSpan = numCols;

      //mmd - 06/09/06 - commenting out this section of code because it causes many additional requests
      //to the server everytime a context menu is opened.  after unit testing, removing piece of code
      //does not appear to affect the function of the menus
      /*var img = document.createElement( "IMG" );
      img.src = transImg_;
      img.width = 1;
      img.height = 1;
      img.style.display = "block";
      td.appendChild( img );*/

      tr.appendChild( td );
      tableBody.appendChild( tr );
   }
}

// changes the selected state for menu item
function UilMenuItemSetSelected( isSelected ) {

   if ( isSelected && !this.isSelected ) {
      debug( 'selected: ' + this.text );
      // handle the previous selection first
      if ( this.parentMenu != null &&
           this.parentMenu.isVisible &&
           this.parentMenu.selectedItem != null &&
           this.parentMenu.selectedItem != this ) {
         // hide previous selection's submenu
         if ( this.parentMenu.selectedItem.submenu != null ) {
            this.parentMenu.selectedItem.submenu.hide();
         }
         // unselect previous selection from parent menu
         this.parentMenu.selectedItem.setSelected( false );
      }

      // select this menu item
      this.isSelected = true;
      if ( this.parentMenu != null && this.parentMenu.isVisible ) {
         this.parentMenu.selectedItem = this;
      }

      // update the styles
      this.updateStyle( this.itemTag );
   }
   else if ( !isSelected && this.isSelected ) {
      debug( 'deselected: ' + this.text );
      // menu item cannot be unselected if its submenu is visible
      if ( this.submenu == null || ( this.submenu != null && !this.submenu.isVisible ) ) {
         // unselect this menu item
         this.isSelected = false;
         if ( this.parentmenu != null ) {
            this.parentmenu.selectedItem = null;
         }

         // update the styles
         this.updateStyle( this.itemTag );
      }
   }
}

// recursively set the style of the menu item html element
function UilMenuItemUpdateStyle( tag, styleID ) {
   if ( tag != null ) {
      if ( styleID == null ) {
         //styleID = "pop5";  //@06C1
         styleID = this.menuStyle;
         if ( !this.isEnabled ) {
            //styleID = "pop7"; //@06C1
            styleID = this.menuStyle; 
         }
         else if ( this.isSelected ) {
            //styleID = "pop6";  //@06C1
            styleID = this.selectedMenuStyle;  //@06C1
         }
         if ( this.arrowTag != null ) {
            if ( this.isEnabled && this.isSelected ) {
               if (this.parentMenu.isLTR) this.arrowTag.src = arrowSel_;
               else this.arrowTag.src = arrowSelRTL_;
            }
            else if ( !this.isEnabled ) {
               if (this.parentMenu.isLTR) this.arrowTag.src = arrowDis_;
               else this.arrowTag.src = arrowDisRTL_;
            }
            else {
               if (this.parentMenu.isLTR) this.arrowTag.src = arrowNorm_;
               else this.arrowTag.src = arrowNormRTL_;
            }
         }
      }
      tag.className = styleID;
      if ( tag.childNodes != null ) {
         for ( var i=0; i<tag.childNodes.length; i++ ) {
            this.updateStyle( tag.childNodes[i], styleID );
         }
      }
   }
}

// returns the next enabled, non-separator menu item after the given item
function UilMenuItemGetNextItem() {
   var menu = this.parentMenu;
   if ( menu != null ) {
      for ( var i=0; i<menu.items.length; i++ ) {
         if ( menu.items[i] == this ) {
            for ( var j=i+1; j<menu.items.length; j++ ) {
               if ( !menu.items[j].isSeparator && menu.items[j].isEnabled ) {
                  return menu.items[j];
               }
            }
            // no next item
            return null;
         }
      }
   }
   return null;
}

// returns the previous enabled, non-separator menu item before the given item
function UilMenuItemGetPrevItem() {
   var menu = this.parentMenu;
   if ( menu != null ) {
      for ( var i=menu.items.length-1; i>=0; i-- ) {
         if ( menu.items[i] == this ) {
            for ( var j=i-1; j>=0; j-- ) {
               if ( !menu.items[j].isSeparator && menu.items[j].isEnabled ) {
                  return menu.items[j];
               }
            }
            // no previous item
            return null;
         }
      }
   }
   return null;
}

// launches the action for a menu item
// method called by an event handler (href for anchor tag)
function menuItemLaunchAction() {
   if ( visibleMenu_ != null ) {
      //var evt = window.event;
      //var item = visibleMenu_.getMenuItem( evt.target );
      var item = visibleMenu_.getSelectedItem();
      if ( item != null && item.isEnabled ) {
         hideCurrentContextMenu( true );
         if ( item.clientAction != null ) {
            eval( item.clientAction );
         }
         if ( item.action != null ) {
            if ( item.action.indexOf( "javascript:" ) == 0 ) {
               eval( item.action );
            }
            else {
               //window.location.href = item.action; //@04D
            }
         }
      }
   }
}

// shows the submenu for a menu item
// method called by an event handler (onclick for anchor tag)
function menuItemShowSubmenu(evt) {
   if ( visibleMenu_ != null ) {
      var item = visibleMenu_.getMenuItem( evt.target );
      if ( item != null && item.isEnabled ) {
         var menu = item.submenu;
         if ( menu != null ) {
            menu.show( item.anchorTag, item );
         }
      }
   }
}

// focus handler for menu item
// method called by an event handler (onfocus for anchor tag)
function menuItemFocus(evt) {
   if ( visibleMenu_ != null ) {
      var item = visibleMenu_.getMenuItem( evt.target );
      if ( item != null ) {
         // select the focused menu item
         //item.anchorTag.hideFocus = item.isEnabled;
         item.setSelected( true );
      }
   }
}

// blur handler for menu item
// method called by an event handler (onblur for anchor tag)
function menuItemBlur(evt) {
   if ( visibleMenu_ != null ) {
      var item = visibleMenu_.getMenuItem( evt.target );
      if ( item != null ) {
         /* //jcp
         if ( item.isFirst && evt.shiftKey ) {
            debug( 'blur = ' + item.text );
            // user is shift tabbing off the beginning of the menu
            // set focus on the launcher
            item.parentMenu.launcher.focus();
            // hide the menu
            item.parentMenu.hide();
         }
         */
      }
   }
}

// key press handler for menu item
// method called by an event handler (onkeydown for anchor tag)
function menuItemKeyDown(evt) {
   var item = null;
   if ( visibleMenu_ != null ) {
      item = visibleMenu_.getMenuItem( evt.target );
   }
   if ( item != null ) {
      var next = null;
      switch ( evt.keyCode ) {
      case 38: // up key
         next = item.getPrevItem();
         if ( next != null ) {
            next.anchorTag.focus();
         }
         else if ( item.parentMenu != visibleMenu_ ) {
            item.parentMenu.launcher.focus();
            item.parentMenu.hide();
         }
         else {
            visibleMenu_.launcher.focus();
            hideCurrentContextMenu( true );
         }
         return false;
         break;
      case 40: // down key
         next = item.getNextItem();
         if ( next != null ) {
            next.anchorTag.focus();
         }
         else if ( item.parentMenu != visibleMenu_ ) {
            item.parentMenu.launcher.focus();
            item.parentMenu.hide();
         }
         else {
            visibleMenu_.launcher.focus();
            hideCurrentContextMenu( true );
         }
         return false;
         break;
      case 39: // right key
         if ( visibleMenu_.isLTR ) {
            if ( item.submenu != null ) {
               menuItemShowSubmenu(evt);
               item.submenu.items[0].anchorTag.focus();
            }
         }
         else {
            if ( item.parentMenu != visibleMenu_ ) {
               item.parentMenu.launcher.focus();
               item.parentMenu.hide();
            }
         }
         return false;
         break;
      case 37: // left key
         if ( visibleMenu_.isLTR ) {
            if ( item.parentMenu != visibleMenu_ ) {
               item.parentMenu.launcher.focus();
               item.parentMenu.hide();
            }
         }
         else {
            if ( item.submenu != null ) {
               menuItemShowSubmenu(evt);
               item.submenu.items[0].anchorTag.focus();
            }
         }
         return false;
         break;
      case 9: // tab key
         visibleMenu_.launcher.focus();
         hideCurrentContextMenu( true );
         break;
      case 27: // escape key
         visibleMenu_.launcher.focus();
         hideCurrentContextMenu( true );
         break;
      case 13: // enter key
         break;
      default:
         break;
      }
   }
}

// handle mouse move for menu item
// method called by an event handler (onmousemove for item tag)
function menuItemMouseMove(evt) {
   if ( visibleMenu_ != null ) {
      var item = visibleMenu_.getMenuItem( evt.target );
      if ( item != null ) {
         if ( !item.isSelected ) {
            // set focus on the anchor and select the menu item
            item.anchorTag.focus();
         }
         if ( item.submenu != null && !item.submenu.isVisible && item.isEnabled ) {
            // show the submenu
            item.submenu.show( item.anchorTag, item );
         }
      }
   }
}

// handle mouse down event for menu item
// method called by an event handler (onmousedown for item tag)
function menuItemMouseDown(evt) {
   /* //@08D
   if ( visibleMenu_ != null ) {
      var item = visibleMenu_.getMenuItem( evt.target );
      if ( item != null ) {
         item.setSelected( true );
      }
      else { //@03A
         item = visibleMenu_.getSelectedItem();
      }
      if ( item != null && item.anchorTag != evt.target ) {
         //item.anchorTag.click();
         menuItemLaunchAction();
      }
   }
   */
   menuItemLaunchAction(); //@08A
}

var allMenus_ = new Array();

//ARC CHANGES FOR SPECIFYING STYLES - BEGIN
//ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - BEGIN
function createContextMenu( name, isLTR, width, borderStyle, tableStyle, noActionsText, noActionsTextStyle, positionUnder ) {
   var menu = new UilContextMenu( name, isLTR, width, borderStyle, tableStyle, noActionsText, noActionsTextStyle, positionUnder );
   allMenus_[ allMenus_.length ] = menu;
   return menu;
}
//ARC CHANGES FOR SPECIFIYING EMPTY MENU TEXT - END
//ARC CHANGES FOR SPECIFYING STYLES - END

function getContextMenu( name ) {
   for ( var i=0; i<allMenus_.length; i++ ) {
      if ( allMenus_[i].name == name ) {
         return allMenus_[i];
      }
   }
   return null;
}

function showContextMenu( name, isDynamic, isCacheable ) {
   contextMenuShow( name, isDynamic, isCacheable, event.target, true );
}

function showContextMenu( name, launcher ) {
   contextMenuShow( name, false, false, launcher, true );
}

function contextMenuShow( name, isDynamic, isCacheable, launcher, doLoad ) {
   debug( "***** showContextMenu: " + name )

   //  We mnight need to find a suitable launcher...
   var oldLauncher = launcher;
   while ((null != launcher) && (null == launcher.tagName)) {
      launcher = launcher.parentNode;
   }
   if (null == launcher) { //  shouldn't happen, but...
      launcher = oldLauncher;
   }

   if ( eval( isDynamic ) ) {
      debug( 'showContextMenu: dynamic=true, load=' + eval( doLoad ) + ', cache=' + eval( isCacheable ) );
      // dynamically loaded menu
      if ( eval( doLoad ) ) {
         // load the url into hidden frame
         loadDynamicMenu( name );
      }

      // clone the dynamic menu from hidden frame
      menu = getDynamicMenu( name, eval( isCacheable ) );

      if ( menu == null && top.isContextMenuManager_ != null ) {
         // menu not done loading yet
         debug( 'showContextMenu: ' + name + ' added to queue' );
         top.contextMenuManagerRequest( name, window, launcher, isCacheable );
      }
   }
   else {
      debug( 'showContextMenu: static context menu' );
      // statically defined menu
      menu = getContextMenu( name );
      if ( menu == null ) {
         menu = createContextMenu( name, 150 );
      }
   }

   if ( menu != null ) {
      hideCurrentContextMenu( true );
      menu.show( launcher );
      visibleMenu_ = menu;
   }
   else {
      debug( 'showContextMenu: ' + name + ' unavailable' );
   }
   clearMenuTimer( ); // @05
}

// method called by an event handler (onmousedown for document)
function hideCurrentContextMenu( forceHide ) {
   if ( visibleMenu_ != null && ( forceHide == true || visibleMenu_.isDismissable ) ) {
//      contextMenuDismissEnable();
      if ( visibleMenu_.isVisible ) {
         visibleMenu_.hide();
      }
      if ( visibleMenu_.isDynamic && !visibleMenu_.isCacheable ) {
         uncacheContextMenu( visibleMenu_ );
      }
      visibleMenu_ = null;
   }
}

function uncacheContextMenu( menu ) {
   debug( 'uncache menu: ' + menu.name );
   // recurse
   for ( var i=0; i<menu.items.length; i++ ) {
      if ( menu.items[i].submenu != null ) {
         uncacheContextMenu( menu.items[i].submenu );
      }
   }

   // remove from all menus array
   for ( var i=0; i<allMenus_.length; i++ ) {
      if ( allMenus_[i] == menu ) {
         var temp = new Array();
         var index = 0;
         for ( var j=0; j<allMenus_.length; j++ ) {
            if ( j != i ) {
               temp[ index ] = allMenus_[ j ];
               index++;
            }
         }
         allMenus_ = temp;
         break;
      }
   }
}

function contextMenuSetIcons( transparentImage,
                              arrowDefault, arrowSelected, arrowDisabled,
                              launcherDefault, launcherSelected,
                              arrowDefaultRTL, arrowSelectedRTL, arrowDisabledRTL,
                              launcherDefaultRTL, launcherSelectedRTL ) {
   transImg_ = transparentImage;

   arrowNorm_ = arrowDefault;
   arrowSel_ = arrowSelected;
   arrowDis_ = arrowDisabled;
   launchNorm_ = launcherDefault;
   launchSel_ = launcherSelected;

   arrowNormRTL_ = arrowDefaultRTL;
   arrowSelRTL_ = arrowSelectedRTL;
   arrowDisRTL_ = arrowDisabledRTL;
   launchNormRTL_ = launcherDefaultRTL;
   launchSelRTL_ = launcherSelectedRTL;

   contextMenuPreloadImage( transImg_ );

   contextMenuPreloadImage( arrowNorm_ );
   contextMenuPreloadImage( arrowSel_ );
   contextMenuPreloadImage( arrowDis_ );
   contextMenuPreloadImage( launchNorm_ );
   contextMenuPreloadImage( launchSel_ );

   contextMenuPreloadImage( arrowNormRTL_ );
   contextMenuPreloadImage( arrowSelRTL_ );
   contextMenuPreloadImage( arrowDisRTL_ );
   contextMenuPreloadImage( launchNormRTL_ );
   contextMenuPreloadImage( launchSelRTL_ );
}

function contextMenuSetArrowIconDimensions( width, height ) {
   arrowWidth_ = width;
   arrowHeight_ = height;
}

function contextMenuPreloadImage( imgsrc ) {
   var preload = new Image();
   preload.src = imgsrc;
}

function toggleLauncherIcon( popupID, selected, isLTR ) {
   if ( selected ) {
      if ( isLTR ) {
         document.images[ popupID ].src = launchSel_;
      }
      else {
         document.images[ popupID ].src = launchSelRTL_;
      }
   }
   else {
      if ( isLTR ) {
         document.images[ popupID ].src = launchNorm_;
      }
      else {
         document.images[ popupID ].src = launchNormRTL_;
      }
   }
   return true;
}

function contextMenuSetNoActionsText( noActionsText, submenuAltText ) {
   noActionsText_ = noActionsText;
   submenuAltText_ = submenuAltText;
}

function contextMenuGetNoActionsText() {
   return noActionsText_;
}

function getWidth( tag ) {
   return tag.offsetWidth;
}

function getHeight( tag ) {
   return tag.offsetHeight;
}

function getLeft( tag, recurse ) {
   var size = 0;
   if ( tag != null ) {
      //size = parseInt( document.defaultView.getComputedStyle(tag, null).getPropertyValue("left") ); //@04D

      //@04A
      if ( recurse && tag.offsetParent != null ) {
         size += getLeft( tag.offsetParent, recurse );
      }
      if ( tag != null ) {
         size += tag.offsetLeft;
      }

   }
   return size;
}

function getTop( tag, recurse ) {
   var size = 0;
   if ( tag != null ) {
      //size = parseInt( document.defaultView.getComputedStyle(tag, null).getPropertyValue("top") ); //@04D

      //@04A
      if ( recurse && tag.offsetParent != null ) {
         size += getTop( tag.offsetParent, recurse );
      }
      if ( tag != null ) {
         size += tag.offsetTop;
      }

   }
   return size;
}

/*****************************************************
* code for dynamically loaded menus
*****************************************************/
function loadDynamicMenu( menuURL ) {
   debug( '* loadDynamicMenu: ' + menuURL );
   var menu = getContextMenu( menuURL );
   if ( menu != null ) {
      if ( menu.isVisible ) {
         // dynamic menu requested, but it's currently showing
         menu.hide();
      }
      if ( !menu.isCacheable ) {
         // make sure it's not in the cache
         uncacheContextMenu( menu );
      }
   }
   if ( getContextMenu( menuURL ) == null ) {
      if ( top.isContextMenuManager_ != null ) {
         debug( 'loadDynamicMenu: loading' );
         top.contextMenuManagerLoadDynamicMenu( menuURL );
      }
   }
}

function getDynamicMenu( menuURL, cache ) {
   debug( '* getDynamicMenu: ' + menuURL );
   var clone = getContextMenu( menuURL );
   if ( clone == null ) {
      if ( top.isContextMenuManager_ != null ) {
         if ( top.contextMenuManagerIsDynamicMenuLoaded() ) {
            var menu = top.contextMenuManagerGetDynamicMenu();
            debug( 'getDynamicMenu: fetched menu from other frame' );
            clone = cloneMenu( menu, menuURL, cache );
            if ( clone.items.length == 0 ) {
               contextMenuSetNoActionsText( top.contextMenuManagerGetNoActionsText() );
            }
         }
         else {
            debug( 'getDynamicMenu: menu not loaded' );
         }
      }
      else {
         debug( 'getDynamicMenu: menu manager not present' );
      }
   }
   else {
      debug( 'getDynamicMenu: menu previously loaded' );
   }
   return clone;
}

function cloneMenu( menu, name, cache ) {
   var clone = getContextMenu( name );
   if ( clone == null ) {
      if ( menu != null ) {
         clone = createContextMenu( name, menu.width );
      }
      else {
         clone = createContextMenu( name, 150 );
      }
      clone.isDynamic = true;
      clone.isCacheable = cache;
      if ( menu != null ) {
         for ( var i=0; i<menu.items.length; i++ ) {
            clone.add( cloneMenuItem( menu.items[i], name + "_sub" + i, cache ) );
         }
      }
   }
   return clone;
}

function cloneMenuItem( item, submenuName, cache ) {
   var submenu = null;
   if ( item.submenu != null ) {
      submenu = cloneMenu( item.submenu, submenuName, cache );
   }
   var clone = new UilMenuItem( item.text, item.isEnabled, item.action, item.clientAction, submenu, item.icon, null, null, null );
   clone.isEnabled = item.isEnabled;
   clone.isSelected = item.isSelected;
   clone.isSeparator = item.isSeparator;
   return clone;
}


//////////////////////////////////////////////////////////////////
// begin BrowserDimensions object definition
ContextMenuBrowserDimensions.prototype				= new Object();
ContextMenuBrowserDimensions.prototype.constructor = ContextMenuBrowserDimensions;
ContextMenuBrowserDimensions.superclass			= null;

function ContextMenuBrowserDimensions(){

    this.body = document.body;
    if (this.isStrictDoctype() && !this.isSafari()) {
        this.body = document.documentElement;
    }

}

ContextMenuBrowserDimensions.prototype.getScrollFromLeft = function(){
    return this.body.scrollLeft ;
}

ContextMenuBrowserDimensions.prototype.getScrollFromTop = function(){
    return this.body.scrollTop ;
}

ContextMenuBrowserDimensions.prototype.getViewableAreaWidth = function(){
    return this.body.clientWidth ;
}

ContextMenuBrowserDimensions.prototype.getViewableAreaHeight = function(){
    return this.body.clientHeight ;
}

ContextMenuBrowserDimensions.prototype.getHTMLElementWidth = function(){
    return this.body.scrollWidth ;
}

ContextMenuBrowserDimensions.prototype.getHTMLElementHeight = function(){
    return this.body.scrollHeight ;
}

ContextMenuBrowserDimensions.prototype.isStrictDoctype = function(){

    return (document.compatMode && document.compatMode != "BackCompat");

}

ContextMenuBrowserDimensions.prototype.isSafari = function(){

    return (navigator.userAgent.toLowerCase().indexOf("safari") >= 0);

}

ContextMenuBrowserDimensions.prototype.isOpera = function(){

    return /Opera/.test(navigator.userAgent);;

}

// end BrowserDimensions object definition
//////////////////////////////////////////////////////////////////

 
var wptheme_DebugUtils = {
    // summary: Collection of utilities for logging debug messages.
    enabled: false, 
    log: function ( /*String*/className, /*String*/message ) {
            // summary: Logs a debugging message, if debugging is enabled.
            // className: the javascript class name or function name which is logging the message
            // message: the message to log
            if ( this.enabled ) {
                message = className + " ==> " + message;
                if ( typeof( console ) == "undefined" ) {
                        console.debug( message );
                }
                else {
                        //better alternative for browsers that don't support console????
                        alert( message );
                }
            }    
    } 
}
var wptheme_HTMLElementUtils = {
	// summary: Collection of utility functions useful for manipulating HTML elements in a cross-browser fashion.
	className: "wptheme_HTMLElementUtils",
	_debugUtils: wptheme_DebugUtils,
    _uniqueIdCounter: 0,
	getUniqueId: function () {
		// summary: Generates a unique identifier (to the current page) by appending a counter to a set prefix.
		//		A page refresh resets the unique identifier counter.
		// returns: a unique identifier
		var retVal = "wptheme_unique_" + this._uniqueIdCounter;
		this._uniqueIdCounter++;
		return retVal;	// String
	},
	sizeToViewableArea: function ( /*HTMLElement*/element ) {
		// summary: Sizes the given element to the viewable area of the browser in a cross-browser fashion.
		// element: the html element to size
		var browserDimensions = new BrowserDimensions();
		element.style.height = browserDimensions.getViewableAreaHeight() + "px";
		element.style.width = browserDimensions.getViewableAreaWidth() + "px";
		element.style.top = browserDimensions.getScrollFromTop() + "px";
		element.style.left = browserDimensions.getScrollFromLeft() + "px";
	},
	sizeToEntireArea: function ( /*HTMLElement*/element ) {
		// summary: Sizes the given element to the viewable area plus the scroll area.
		// element: the html element to size
		var browserDimensions = new BrowserDimensions();
		//The getHTMLElement*() functions return the exact size of the body element in IE even if the viewable area is
		//larger (i.e. no scroll bars). In Firefox, the dimensions of the viewable area plus the scrollable area is returned.
		//So in IE, we take the viewable area if that is larger and the HTMLElement* area if that is larger.
		element.style.height = Math.max( browserDimensions.getHTMLElementHeight(), browserDimensions.getViewableAreaHeight() ) + "px";
		element.style.width = Math.max( browserDimensions.getHTMLElementWidth(), browserDimensions.getViewableAreaWidth() ) + "px";
	},
	sizeRelativeToViewableArea: function ( /*HTMLElement*/element, /*float*/heightFactor, /*float*/widthFactor ) {
		// summary: Sizes the given element to a certain multiple of the viewable area. For example, if heightFactor is 0.5,
		//		the height of the given element will be set to half of the viewable area height.
		// element: the html element to size
		// heightFactor: the factor to multiply the viewable area height by
		// widthFactor: the factor to multiply the viewable area width by
		var browserDimensions = new BrowserDimensions();
		element.style.height = ( browserDimensions.getViewableAreaHeight() * heightFactor ) + "px";
		element.style.width = ( browserDimensions.getViewableAreaWidth() * widthFactor ) + "px";
	},
	positionRelativeToViewableArea: function ( /*HTMLElement*/element, /*float*/heightFactor, /*float*/widthFactor ) {
		// summary: Positions the given element relative to the viewable area. For example, if the heightFactor is 0.5, the
		// 		top of the element will be positioned (Y axis) halfway down the viewable area. Note that this means the element
		//		will be ABSOLUTELY positioned.
		// element: the html element to position
		// heightFactor: the factor to multiply the viewable area height by
		// widthFactor: the factor to multiply the viewable area width by
		var browserDimensions = new BrowserDimensions();
		element.style.position = "absolute";
		if ( this._debugUtils.enabled ) { 
			this._debugUtils.log( this.className, "Browser's viewable height: " + browserDimensions.getViewableAreaHeight() );
			this._debugUtils.log( this.className, "Browser's viewable width: " + browserDimensions.getViewableAreaWidth() );
			this._debugUtils.log( this.className, "Browser's scroll from top: " + browserDimensions.getScrollFromTop() ); 
			this._debugUtils.log( this.className, "Browser's scroll from left: " + browserDimensions.getScrollFromLeft() );
		}
		element.style.top = ( ( browserDimensions.getViewableAreaHeight() * heightFactor ) + browserDimensions.getScrollFromTop() ) + "px";
		//Scroll left behaves differently in FF & IE in RTL languages. The "correct" behavior is up for debate. In FF, it will return the "correct" value 
		//(scrollLeft switches when the page is rendered right-to-left). In IE, scroll left will basically return the scroll width for the body element.
		//There's no real "capability" to test for here so the window.attachEvent is a cheap trick to check for IE.
		//bidiSupport is defined in the theme.
		if ( bidiSupport.isRTL && window.attachEvent ) {
			if ( this._debugUtils.enabled ) {
				this._debugUtils.log( this.className, "scrollWidth = " + browserDimensions.getHTMLElementWidth() );
				this._debugUtils.log( this.className, "clientWidth = " + browserDimensions.getViewableAreaWidth() );
				this._debugUtils.log( this.className, "Scroll Offset should be: " + ( browserDimensions.getHTMLElementWidth() - browserDimensions.getViewableAreaWidth() - browserDimensions.getScrollFromLeft() ) );
			}
			element.style.left = ( ( browserDimensions.getViewableAreaWidth() * widthFactor) + ( browserDimensions.getHTMLElementWidth() - browserDimensions.getViewableAreaWidth() - browserDimensions.getScrollFromLeft() ) ) + "px";
		}
		else {
			element.style.left = ( ( browserDimensions.getViewableAreaWidth() * widthFactor ) + browserDimensions.getScrollFromLeft() ) + "px";
		}	
	},
	positionOutsideElementTopRight: function ( /*HTMLElement*/elementToPosition, /*HTMLElement*/relativeElement ) {
		// summary: Positions the given element just outside (to the top and lining up with the right edge) of the
		//		relative element.
		// description: Sets the top (Y-axis) position of the given element to the top of the relative element minus the
		//		height of element being positioned. Sets the left (X-axis) position of the given element to the left position of
		//		the relative element plus the width of the relative element (to get the right edge) minus the width of the element 
		// 		being positioned (to line the end of the element being positioned up with the right edge of the relative element). 
		elementToPosition.style.position = "absolute";
		elementToPosition.style.top = ( this.stripUnits( relativeElement.style.top ) - elementToPosition.offsetHeight ) + "px";
		if ( bidiSupport.isRTL ) {
			elementToPosition.style.left = ( this.stripUnits( relativeElement.style.left ) ) + "px";
		}
		else {
			elementToPosition.style.left = ( this.stripUnits( relativeElement.style.left ) + relativeElement.offsetWidth - elementToPosition.offsetWidth) + "px";	
		}
	},
	stripUnits: function ( /*String*/cssProp ) {
		// summary: Strips any units (i.e. "px") from a CSS style property.
		// returns: the number value minus any units
		return parseInt( cssProp.substring( 0, cssProp.length - 2 ));	//integer
	},
	addClassName: function ( /*HTMLElement*/element, /*String*/className ) {
		// summary: Adds the given className to the element's style definitions.
		// element: the HTMLElement to add the class name to
		// className: the className to add
		var clazz = element.className;
		if ( clazz.indexOf( className ) < 0 ) {
			element.className += (" " + className);
		}	
	},
	removeClassName: function ( /*HTMLElement*/element, /*String*/className ) {
		// summary: Removes the given className from the element's style definitions.
		// element: the HTMLElement to remove the class name from
		// className: the className to remove
		var clazz = element.className;
		var startIndex = clazz.indexOf( className );
		if ( startIndex >= 0 ) {
			clazz = clazz.substring(0, startIndex) + clazz.substring( startIndex + className.length + 1 );
			element.className = clazz;
		}
	},
	hideElementsByTagName: function ( /*String 1...N*/) {
		// summary: Hides every element of a given tag name. Stores the old visibility style so it can be 
		//		restored by the showElementsByTagName function.
		for ( var i = 0; i < arguments.length; i++ ) {
			var elements = document.getElementsByTagName( arguments[i] );
			for ( var j = 0; j < elements.length; j++ ) {
				if ( elements[j] && elements[j].style ) {
					elements[j]._oldVisibilityStyle = elements[j].style.visibility;
					elements[j].style.visibility = "hidden";
				}
			}
		}
	},
	showElementsByTagName: function ( /*String 1...N*/) {
		// summary: Shows every element of a given tag name. Uses the old visibility style so that elements hidden
		//		by hideElementsByTagName are properly restored.
		for ( var i = 0; i < arguments.length; i++ ) {
			var elements = document.getElementsByTagName( arguments[i] );
			for ( var j = 0; j < elements.length; j++ ) {
				if ( elements[j] && elements[j].style ) {
					if ( elements[j]._oldVisibility ) {
						elements[j].style.visibility = elements[j]._oldVisibility;
						elements[j]._oldVisibility = null;
					}
					else {
						elements[j].style.visibility = "visible";
					}
				}
			}
		}	
	},
	addOnload: function ( /*Function*/func ) {
		// summary: Adds a function to be called on page load.
		// func: the function to call 
		if ( window.addEventListener ) {
			window.addEventListener( "load", func, false );
		}
		else if ( window.attachEvent ) {
			window.attachEvent( "onload", func );
		}
	},
	getEventObject: function ( /*Event?*/event ) {
		// summary: Cross-browser function to retrieve the event object.
		// event: In W3C-compliant browsers, this object will just simply be
		//		returned
		// returns: the event object
		var result = event;
		if ( !event && window.event ) {
			result = window.event;
		}
		return result;	// Event
	}
}

var wptheme_CookieUtils = {
	// summary: Various utility functions for dealing with cookies on the client.
	_deleteDate: new Date( "1/1/2003" ),
	_undefinedOrNull: function ( /*Object*/variable ) {
            // summary: Determines if a given variable is undefined or NULL.
            // returns: true if undefined OR NULL, false otherwise.
            return ( typeof ( variable ) == "undefined" || variable == null );  // boolean
	},
	debug: wptheme_DebugUtils,
    className: "wptheme_CookieUtils",
	getCookie: function ( /*String*/cookieName ) {
		// summary: Gets the value for a given cookie name. If no value is found, returns NULL.
		if ( this.debug.enabled ) { this.debug.log( this.className, "getCookie( " + cookieName + " )" ); }
		
		cookieName = cookieName + "="
		var retVal = null;
		
		if ( this.debug.enabled ) { 
			this.debug.log( this.className, "document.cookie=" + document.cookie );
			this.debug.log( this.className, "indexOf cookieName: " + document.cookie.indexOf( cookieName ) );  
		}
		
		if ( document.cookie.indexOf( cookieName ) >= 0 )
		{
		    var cookies = document.cookie.split(";");
		
		    var c = 0;
		    if ( this.debug.enabled && cookies.length > 0 ) {
		    	this.debug.log( this.className, "cookies[0] = " + cookies[0] );
		    }
		    while ( c < cookies.length && ( cookies[c].indexOf( cookieName ) == -1 ) )
		    {
		        if ( this.debug.enabled ) { this.debug.log( this.className, "cookies[" + c + "] = " + cookies[0] ); }
		        c=c+1;
		    }
			
			//Make sure there's no leading or trailing spaces on our cookie name/value pair.
			var cookieNVP = cookies[c].replace( /^[ \s]+|[ \s]+$/, '' );
			
			if ( this.debug.enabled ) { 
				this.debug.log( this.className, "cookieName=\"" + cookieName + "\"." );
				this.debug.log( this.className, "cookieName.length=\"" + cookieName.length + "\"." ); 
				this.debug.log( this.className, "cookieNVP=\"" + cookieNVP + "\"." );
				this.debug.log( this.className, "cookieNVP.length=\"" + cookieNVP.length + "\"." );
			}
			
			var cookieValue = cookieNVP.substring( cookieName.length );
		    if ( this.debug.enabled ) { this.debug.log( this.className, "cookie value =\"" + cookieValue + "\"."); }
		    
		    if ( cookieValue != "null" )
		    {
		        retVal = cookieValue;
		    }
		}
		
		if ( this.debug.enabled ) { this.debug.log( this.className, "getCookie( " + cookieName + " ) return " + retVal ); }
		return retVal; // String
	},
	setCookie: function ( /*String*/name, /*String*/value, /*Date?*/expiration, /*String?*/path ) {
		// summary: Creates the cookie based on the given information.
		// name: the name of the cookie
		// value: the value for the cookie
		// expiration: OPTIONAL -- when the cookie should expire
		// path: OPTIONAL -- the url path the cookie applies to
		if ( this.debug.enabled ) { this.debug.log( this.className, "set cookie (" + [ name, value, expiration, path ]  + ")"); }
		
		if ( this._undefinedOrNull( name ) ) { throw Error( "Unable to set cookie! No name given!" ); }
		if ( this._undefinedOrNull( value ) ) { throw Error( "Unable to set cookie! No value given!" ); }
		if ( this._undefinedOrNull( expiration ) ) { 
			expiration = ""; 
		}
		else {
			expiration = "expiration=" + expiration.toUTCString() + ";";
		}
		if ( this._undefinedOrNull( path ) ) { 
			path = "path=/;"; 
		}
		else {
			path = "path=" + path + ";";
		}
		
		document.cookie=name + '=' + value + ';' + expiration +  path;
		if ( this.debug.enabled ) { this.debug.log( this.className, "document.cookie after setting the cookie=" + document.cookie ); }		
	},
	deleteCookie: function ( /*String*/cookieName ) {
		// summary: Deletes a given cookie by setting the value to "null" and setting the expiration
		//		value to expire completely.
		if ( this.debug.enabled ) { this.debug.log( this.className, "delete cookie (" + [ cookieName ] + ") "); }
		if(wpsFLY_isIE){
            this.setCookie( cookieName, "null", this._deleteDate);
        }else{
            this.setCookie( cookieName, "");
        }
	}
} 
// Populates and shows a context menu asynchronously. 
//
// uniqueID             - some unique identifier describing the context of the menu (i.e. portlet window id)
// urlToMenuContents    - url target for the iFrame
// isLTR                - indicates if the page orientation is Left-to-Right
//
//
// This function creates a context menu using the WCL context menu javascript library. It populates this menu
// by creating a hidden DIV ( the ID consists of the unique identifier with "_DIV" appended ) which contains
// a hidden IFRAME ( the ID consists of the DIV identifier with "_IFRAME" appended ). The IFRAME loads the 
// specified URL and calls the buildAndDisplayMenu() function upon completion of loading the IFRAME. The document
// returned by the specified URL must contain a javascript function called "getMenuContents()" which returns 
// an array. The contents of the array must be in the following format ( array[i] = <menu-item-display-name>; 
// array[i+1] = <menu-item-action-url> ). The menu is attached to an HTML element with the id equal to the 
// unique identifier. So, in the portlet context menu case, the image associated with the context menu must have
// an ID equal to the portlet window ID. The dynamically created DIV and IFRAME are deleted after the menu 
// contents are populated and the same menu is returned for the duration of the request in which it was created.
//

//Control debugging. 
// -1 - no debugging
//  0 - minimal debugging ( adding items to menus )
//  1 - medium debugging ( function entry/exit )
//  2 - maximum debugging ( makes iframe visible )
// 999 - make iframe visible only
var asynchContextMenuDebug = -1;

var asynchContextMenuMouseOverIndicator = "";

var portletIdMap = new Object();

function asynchContextMenuOnMouseClickHandler( uniqueID, isLTR, urlToMenuContents, menuBorderStyle, menuTableStyle, menuItemStyle, menuItemSelectedStyle, emptyMenuText, loadingImage, renderBelow )
{
	var menuID = "contextMenu_" + uniqueID;
    
    var menu = getContextMenu( menuID );
    
    if (menu == null) 
    { 
    	asynchContextMenu_menuCurrentlyLoading = uniqueID;
    	
    	if ( loadingImage )
    	{
			setLoadingImage( loadingImage );
	    }

        menu = createContextMenu( menuID, isLTR, null, menuBorderStyle, menuTableStyle, emptyMenuText, null, renderBelow );
        loadAsynchContextMenu( uniqueID, urlToMenuContents, isLTR, menuItemStyle, menuItemSelectedStyle, '', true );
    }
    else
    {
    	if ( asynchContextMenu_menuCurrentlyLoading == uniqueID )
		{
			return;	
		}
    	showContextMenu( menuID, document.getElementById( uniqueID ) );
    }	
}

var asynchContextMenu_originalMenuImgElementSrc;

function setLoadingImage( img )
{
	asynchContextMenu_originalMenuImgElementSrc = document.getElementById( asynchContextMenu_menuCurrentlyLoading + "_img" ).src;
	document.getElementById( asynchContextMenu_menuCurrentlyLoading + "_img" ).src = img;
}

function clearLoadingImage()
{
	document.getElementById( asynchContextMenu_menuCurrentlyLoading + "_img" ).src = asynchContextMenu_originalMenuImgElementSrc;
}

function loadAsynchContextMenu( uniqueID, url, isLTR, menuItemStyle, menuItemSelectedStyle, emptyMenuText, showMenu, onMenuAffordanceShowHandler )
{
    asynchDebug( 'ENTRY loadAsynchContextMenu p1=' + uniqueID + '; p2=' + url + '; p3=' + isLTR + '; p4=' + isLTR);
	
	var menuID = "contextMenu_" + uniqueID;

    var dialogTag = null;
    var ID = uniqueID + '_DIV';
			
	//an iframe wasn't cleaned up properly
	if ( document.getElementById( ID ) != null )
	{
		closeMenu( ID );
        return;
	}
	
	//create the div tag and assign the styles to it		
	dialogTag = document.createElement( "DIV" );
	dialogTag.style.position="absolute";
	
	if ( asynchContextMenuDebug < 2 )
	{
		dialogTag.style.left = "0px";
		dialogTag.style.top  = "-9999px";
    	dialogTag.style.visibility = "hidden";
	}
	
	if ( asynchContextMenuDebug >= 2 || asynchContextMenuDebug == 999 )
	{
		dialogTag.style.left = "100px";
		dialogTag.style.top  = "100px";
    	dialogTag.style.visibility = "visible";
	}		
	
	dialogTag.id=ID;
	
	var styleString = 'null';
	
	if ( menuItemStyle != null )
	{
		styleString = "'" + menuItemStyle + "'";
	}
	
	if ( menuItemSelectedStyle != null ) 
	{
		styleString = styleString + ", '" + menuItemSelectedStyle + "'";
	}
	else
	{
		styleString = styleString + ", null";
	}
	
	//alert( 'buildAndDisplayMenu( this.id, this.name, ' + styleString + ', ' + showMenu + ' , ' + callbackFn + ' );' );
	
	//create the iframe this way because onload handlers attached when creating dynamically don't seem to fire
	dialogTag.innerHTML='<iframe id="' + menuID + '" name="' + ID + '_IFRAME" src="' + url + '" onload="buildAndDisplayMenu( this.id, this.name, ' + styleString + ', ' + showMenu + ' , \''+ onMenuAffordanceShowHandler + '\' ); return false;" style="height: 1px; width: 1px;" ></iframe>';
	
	//append the div tag to the document body		
	document.body.appendChild( dialogTag );
	

    asynchDebug( 'EXIT createDynamicElements' );

}



//Builds and displays the menu from the contents of the IFRAME.
function buildAndDisplayMenu( menuID, iframeID, menuItemStyle, menuItemSelectedStyle, showMenu, onMenuAffordanceShowHandler )
{
    asynchDebug( 'ENTRY buildAndDisplayMenu p1=' + menuID + '; p2=' + iframeID + '; p3=' + showMenu + '; p4=' + onMenuAffordanceShowHandler );
    
    //get the context menu, should have already been created.
    var menu = getContextMenu( menuID );

	//clear out our loading indicator
    clearLoadingImage();
   	asynchContextMenu_menuCurrentlyLoading = null;

    //if the menu doesn't exist, we shouldn't even be here....but just in case.
    if ( menu == null )
    {
        return false;
    }

    //strip the _IFRAME from the id to come up with the DIV id
    index = iframeID.indexOf( "_IFRAME" );
    var divID = iframeID.substring( 0, index );

    //strip the _DIV from the id to come up with the portlet id
    index2 = divID.indexOf( "_DIV" );
    var uniqueID = divID.substring( 0, index2 );
    
    asynchDebug( 'divID = ' + divID );
    asynchDebug( 'uniqueID = ' + uniqueID );

    var frame, c=-1, done=false;

    //In IE, referencing the iFrame via the name in the window.frames[] array
    //does not appear to work in this case, so we have to cycle through all the 
    //frames and compare the names to find the correct one.
    while ( ( c + 1 ) < window.frames.length && !done )
    {  
        c=c+1;

		//We have to surround this with a try/catch block because there are
		//cases where attempting to access the 'name' property of the current
		//frame in the array will generate an access denied exception. This is 
		//OK to ignore because any frame that generates this exception shouldn't
		//be the one we are looking for.
        try 
        {
            done = ( window.frames[c].name == iframeID );
        }
        catch ( e )
        {
            //do nothing.
        }
    }

    //Check for the existence of the function we are looking to call. 
    //If not, don't bother creating the menu. 
    if ( window.frames[c].getMenuContents )
    {
        contents = window.frames[c].getMenuContents();
    }
    else
    {
        //we were unable to load the context menu for whatever reason
	asynchDebug2( "menu function does not exist...refreshing whole page");
	setTimeout(function(){
		window.top.location.reload();
	}, 0);
	return false;
    }
    
    
    //Cycle through the array created by the getMenuContents()
    //function. The structure of the array should be [url, name].
    for ( i=0; i < contents.length; i=i+3 ) 
    {
        asynchDebug2( 'Adding item: ' + contents[i+1] );
        asynchDebug2( 'URL: ' + contents[i] );
        if ( contents[i] )
        {
        	asynchDebug2( 'url length: ' + contents[i].length );
        }
        asynchDebug2( 'icon: ' + contents[i+2] );

        if ( contents[i] && contents[i].length != 0 )
        {
        	var icon = null;
        	
        	if ( contents[i+2] && contents[i+2].length != 0 )
        	{
        		icon = contents[i+2];
        	}
        
            menu.add( new UilMenuItem( contents[i+1], true, '', contents[i], null, icon, null, menuItemStyle, menuItemSelectedStyle ) );
        }
    }

    //our target image should have an ID of the uniqueID
    var target = document.getElementById( uniqueID );
    //remove our iframe since we've created the menu, we don't need the iframe on this request anymore.
    // (148004) deleting the elements causes the status bar to spin forever on mozilla
    //deleteDynamicElements( divID );

    asynchDebug( 'EXIT buildAndDisplayMenu' );

	//asynchContextMenuOnLoadCheck( menuID, uniqueID, target, onMenuAffordanceShowHandler );

    //...and display!
    if ( showMenu == null || showMenu == true )
    {
    	return showContextMenu( menuID, target ); 
    }
}

function asynchDebug( str ) 
{
	if ( asynchContextMenuDebug >= 1 && asynchContextMenuDebug != 999 )
	{
	    alert( str );
	}
}

function asynchDebug2( str )
{
	if ( asynchContextMenuDebug >= 0 && asynchContextMenuDebug != 999 )
	{
    	alert( str) ;
    }
}

//MMD - this function is used so that relative URLs may be used with the context menus.
function asynchDoFormSubmit( url ){

    var formElem = document.createElement("form");
    document.body.appendChild(formElem);

    formElem.setAttribute("method", "GET");

    var delimLocation = url.indexOf("?");
    
    if (delimLocation >= 0) {
        var newUrl = url.substring(0, delimLocation);
        
        var paramsEnd = url.length;
        // test to see if a # fragment identifier (the layout node id) is appended to the end of the URL
        var layoutNodeLocation = url.indexOf("#");
        if (layoutNodeLocation >= 0 && layoutNodeLocation > delimLocation) {
            paramsEnd = layoutNodeLocation;
            newUrl = newUrl + url.substring(layoutNodeLocation, url.length);
        }
        
        var params = url.substring(delimLocation + 1, paramsEnd);
        var paramArray = params.split("&");

        for (var i = 0; i < paramArray.length; i++) {
            var name = paramArray[i].substring(0, paramArray[i].indexOf("="));
            var value = paramArray[i].substring(paramArray[i].indexOf("=") + 1, paramArray[i].length);

            var inputElem = document.createElement("input");
            inputElem.setAttribute("type", "hidden");
            inputElem.setAttribute("name", name);
            inputElem.setAttribute("value", value);
            formElem.appendChild(inputElem);
        }
        
        url = newUrl;

    }

    formElem.setAttribute("action", url);
    
    formElem.submit();

}

var asynchContextMenu_menuCurrentlyLoading = null;

function menuMouseOver( id, selectedImage )
{
	if ( asynchContextMenu_menuCurrentlyLoading != null )
		return;

	portletIdMap[id] = 'menu_'+id+'_img';
	showAffordance(id, selectedImage);
}

function menuMouseOut( id, disabledImage )
{
	if ( asynchContextMenu_menuCurrentlyLoading != null )
		return;
	
   	hideAffordance(id , disabledImage);
	portletIdMap[id] = "";
}

function showAffordance( id, selectedImage )
{
	document.getElementById( 'menu_'+id ).style.cursor='pointer';
	document.getElementById( 'menu_'+id+'_img').src=selectedImage;
}

function hideAffordance( id, disabledImage )
{
	document.getElementById( 'menu_'+id ).style.cursor='default';
	document.getElementById( 'menu_'+id+'_img').src=disabledImage;	
}

function menuMouseOverThinSkin(id, selectedImage, minimized)
{
	if ( asynchContextMenu_menuCurrentlyLoading != null )
		return;

	portletIdMap[id] = 'menu_'+id+'_img';
	showAffordanceThinSkin(id, selectedImage, minimized);
}

function menuMouseOutThinSkin(id, disabledImage, minimized )
{
	if ( asynchContextMenu_menuCurrentlyLoading != null)
		return;

   	hideAffordanceThinSkin(id , disabledImage, minimized);
	portletIdMap[id] = "";
}

function showAffordanceThinSkin(id, selectedImage, minimized)
{
	document.getElementById( 'menu_'+id ).style.cursor='pointer';
	document.getElementById( 'portletTitleBar_'+id ).className='wpsThinSkinContainerBar wpsThinSkinContainerBarBorder';
	document.getElementById( 'title_'+id ).className='wpsThinSkinDragZoneContainer wpsThinSkinVisible';
	document.getElementById( 'menu_'+id+'_img' ).src=selectedImage;
}

function hideAffordanceThinSkin(id, disabledImage, minimized)
{
	document.getElementById( 'menu_'+id ).style.cursor='default';
	/* when minimized, the titlebar should always be displayed so it can be found by the user, so we don't hide it */
	if (minimized == null || minimized == false){
	document.getElementById( 'portletTitleBar_'+id ).className='wpsThinSkinContainerBar';
	}
	document.getElementById( 'title_'+id ).className='wpsThinSkinDragZoneContainer wpsThinSkinInvisible';
	document.getElementById( 'menu_'+id+'_img' ).src=disabledImage;	
}

var onmousedownold_;

function closeMenu(id, disabledImage)
{
	hideCurrentContextMenu();

	if (  portletIdMap[id] == "")
	{
		hideAffordance( id, disabledImage );
	}
	
	document.onmousedown = onmousedownold_;
}

function showPortletMenu( id, portletNoActionsText, isRTL, menuPortletURL, disabledImage, loadingImage )
{
	if ( portletIdMap[id].indexOf( id ) < 0  )
		return;
		
	asynchContextMenuOnMouseClickHandler('menu_'+id,!isRTL,menuPortletURL, null, null, null, null, portletNoActionsText, loadingImage );
   	onmousedownold_ = document.onmousedown;
	document.onmousedown = closeMenu;
}

function accessibleShowMenu( event , id , portletNoActionsText, isRTL, menuPortletURL, loadingImage )
{
	if ( event.which == 13 )
	{
	    asynchContextMenuOnMouseClickHandler( 'menu_'+id,!isRTL,menuPortletURL, null, null, null, null, portletNoActionsText, loadingImage );
	}
	else
	{
	 	return true;
	}
}

wptheme_AsyncMenuAffordance = function ( /*String*/anchorId, /*String*/imageId, /*String*/showingImgUrl, /*String*/hidingImgUrl ) {
	// summary: Representation of an asynchronous menu's affordance (UI element which triggers the menu to show). Manages the details
	//		of showing/hiding the affordance, if appropriate.
	// description: In the Portal theme, we want the menu affordance to only show during certain events (e.g. mouseover the page name). The details
	//		of the showing/hiding is a little more complicated than changing the css on an HTML element due to various rendering/accessibility concerns. This 
	//		object manages these details. 
	this.anchorId = anchorId;
	this.imageId = imageId;
	this.showingImgUrl = showingImgUrl;
	this.hidingImgUrl = hidingImgUrl;
	
	this.show = function () {
		// summary: Shows the affordance.		
		if (document.getElementById( this.anchorId ) != null) {
			document.getElementById( this.anchorId ).style.cursor = 'pointer';
			document.getElementById( this.imageId ).src=this.showingImgUrl;
		}	
	}
	this.hide = function () {
		// summary: Hides the affordance.
		if (document.getElementById( this.anchorId ) != null) {
			document.getElementById( this.anchorId ).style.cursor = 'default';
			document.getElementById( this.imageId ).src=this.hidingImgUrl;
		}
	}
}

wptheme_AsyncMenu = function ( /*String*/id, /*String*/menuBorderStyle, /*String*/menuStyle, /*String*/menuItemStyle, /*String*/selectedMenuItemStyle ) {
		// summary: Representation of an asynchronous context menu. Manages showing/hiding the menu as well as showing/hiding the menu's affordance (UI element
		//		which opens the menu). 
		// id: the menu's id
		// menuBorderStyle: the style name to be applied to the menu's border
		// menuStyle: the style name to be applied to the general menu
		// menuItemStyle: the style name to be applied to the menu item
		// selectedMenuItemStyle: the style name to be applied to a selected menu item
		
		//global utilities
		this._htmlUtils = wptheme_HTMLElementUtils;
		
		//properties passed in at construction time
		this.id = id;
		this.menuBorderStyle = menuBorderStyle; 
		this.menuStyle = menuStyle;
		this.menuItemStyle = menuItemStyle;
		this.selectedMenuItemStyle = selectedMenuItemStyle;
		
		//properties that have to be initialized in the theme
		this.url = null;
		this.isRTL = false;
		this.emptyMenuText = null;
		this.loadingImgUrl = null;
		this.affordance = null;
		this.init = function ( /*String*/ url, /*boolean*/isRTL, /*String*/ emptyMenuText, /*String*/ loadingImgUrl, /*wptheme_MenuAffordance*/affordance, /*boolean*/renderBelow ) {
			// summary: Convenience function for setting up the required variables for showing the page menu.
			// url: the url to load page menu contents (usually created with <portal-navgation:url themeTemplate="pageContextMenu" />)
			// isRTL: is the current locale a right-to-left locale
			// emptyMenuText: the text to display if the user has no valid options
			// loadingImgUrl: the url to the image to display while the menu is loading
			this.url = url;
			this.isRTL = isRTL;
			this.emptyMenuText = emptyMenuText;
			this.loadingImgUrl = loadingImgUrl;
			this.affordance = affordance;
			this.renderBelow = renderBelow;
		}
		this.show = function ( /*Event?*/evt ) {
			// summary: Shows the page menu for the selected page. 
			// description: Typically triggered by 2 types of events: click and keypress. On a click event, we just want to show the menu. On a keypress
			//		event, we want to make sure the ENTER/RETURN key was pressed before showing the menu.
			// event: Event object passed in when triggered from a key press event.
			
			evt = this._htmlUtils.getEventObject( evt );
			var show = false;
			var result;
			//On a keypress event, we want to make sure the ENTER/RETURN key was pressed before showing the menu.
			if ( evt && evt.type == "keypress" ) { 
				var keyCode = -1;
				if ( evt && evt.which ){	
					keyCode = evt.which;
				}
				else {
					keyCode = evt.keyCode
				}	
		
				//Enter/Return was the key that triggered this keypress event.
				if ( keyCode == 13 ) {
					show = true;
				}						
			}
			else {
				//Some other kind of event, just show the menu already...
				show = true;
			}
			
			//Show the menu if necessary.
			if ( show ) {
				result = asynchContextMenuOnMouseClickHandler( this.id, !this.isRTL, this.url, this.menuBorderStyle, this.menuStyle, this.menuItemStyle, this.selectedMenuItemStyle, this.emptyMenuText, this.loadingImgUrl, this.renderBelow );
			} 
			
			return result;
		}
		this.showAffordance = function () {
			// summary: Shows the affordance associated with the given asynchronous menu. 
			if ( asynchContextMenu_menuCurrentlyLoading == null ) {
				this.affordance.show();
			}	
		}
		this.hideAffordance = function () {
			// summary: Hides the affordance associated with the given asynchronous menu.
			if ( asynchContextMenu_menuCurrentlyLoading == null ) {
				this.affordance.hide();
			}
		}
	}

wptheme_ContextMenuUtils = {
	// summary: Utility object for managing the different context menus in the theme. Constructs the wptheme_AsyncMenu objects here, initialization must take place in
	//		the head section of the HTML document (usually the initialization values require the usage of JSP tags). 
	moreMenu: new wptheme_AsyncMenu( "wptheme_more_menu", "wptheme-more-menu-border", "wptheme-more-menu", "wptheme-more-menu-item", "wptheme-more-menu-item-selected", true ),
	topNavPageMenu: new wptheme_AsyncMenu( "wptheme_selected_page_menu", "wptheme-page-menu-border", "wptheme-page-menu", "wptheme-page-menu-item", "wptheme-page-menu-item-selected" ),
	sideNavPageMenu: new wptheme_AsyncMenu( "wptheme_selected_page_menu", "wptheme-page-menu-border", "wptheme-page-menu", "wptheme-page-menu-item", "wptheme-page-menu-item-selected" )
}



 
//////////////////////////////////////////////////////////////////
// begin BrowserDimensions object definition
BrowserDimensions.prototype				= new Object();
BrowserDimensions.prototype.constructor = BrowserDimensions;
BrowserDimensions.superclass			= null;

function BrowserDimensions(){

    this.body = document.body;
    if (this.isStrictDoctype() && !this.isSafari()) {
        this.body = document.documentElement;
    }

}

BrowserDimensions.prototype.getScrollFromLeft = function(){
    return this.body.scrollLeft ;
}

BrowserDimensions.prototype.getScrollFromTop = function(){
    return this.body.scrollTop ;
}

BrowserDimensions.prototype.getViewableAreaWidth = function(){
    return this.body.clientWidth ;
}

BrowserDimensions.prototype.getViewableAreaHeight = function(){
	if(this.isSafari()) return document.documentElement.clientHeight;
    return this.body.clientHeight ;
}

BrowserDimensions.prototype.getHTMLElementWidth = function(){
    return this.body.scrollWidth ;
}

BrowserDimensions.prototype.getHTMLElementHeight = function(){
    return this.body.scrollHeight ;
}

BrowserDimensions.prototype.isStrictDoctype = function(){

    return (document.compatMode && document.compatMode != "BackCompat");

}

BrowserDimensions.prototype.isSafari = function(){

    return (navigator.userAgent.toLowerCase().indexOf("safari") >= 0);

}

BrowserDimensions.prototype.isOpera = function(){

    return (navigator.userAgent.toLowerCase().indexOf("opera") >= 0);

}

// end BrowserDimensions object definition
//////////////////////////////////////////////////////////////////

 
//Provides a controller for enabling and disabling javascript events
//on a particular HTML element. Elements must register with the controller
//in order to be enabled/disabled. The act of registering with the controller
//disables the element, unless otherwise specified. 
//
//
// **The main purpose of this controller is to disable the javascript 
//actions of certain elements that, if executed prior to the page completely 
//loading, cause problems.

//Object definition for ElementJavascriptEventController
function ElementJavascriptEventController()
{
	//Registered elements to disable and enable upon page load.
	this.elements = new Array();
	this.arrayPosition = 0;
	
	//Function mappings
	this.enableAll = enableRegisteredElementsInternal;
	this.disableAll = disableRegisteredElementsInternal;
	this.register = registerElementInternal;
	this.enable = enableRegisteredElementInternal;
	this.disable = disableRegisteredElementInternal;
	
	//Enables all registered items.
	function enableRegisteredElementsInternal()
	{
		for ( var c=0; c < this.arrayPosition; c=c+1 )
		{
			this.elements[c].enable();	
		}
	}
	
	function enableRegisteredElementInternal( id )
	{
		for ( var c=0; c < this.arrayPosition; c=c+1 )
		{
			if ( this.elements[c].ID == id )
			{
				this.elements[c].enable();
			}
		}
	}
	
	//Disables all registered items.
	function disableRegisteredElementsInternal()
	{
		for ( var c=0; c < this.arrayPosition; c=c+1 )
		{
			this.elements[c].disable();
		}
	}
	
	function disableRegisteredElementInternal( id )
	{
		for ( var c=0; c < this.arrayPosition; c=c+1 )
		{
			if ( this.elements[c].ID == id )
			{
				this.elements[c].disable();
			}
		}
	}
	
	//Registers an item with the controller.
	function registerElementInternal( HTMLElementID, doNotDisable, optionalOnEnableJavascriptAction )
	{
		this.elements[ this.arrayPosition ] = new RegisteredElement( HTMLElementID, doNotDisable, optionalOnEnableJavascriptAction );
		this.arrayPosition = this.arrayPosition + 1;
	}
}

//Object definition for an element registered with the controller.
//These objects should only be created by the controller.
function RegisteredElement( ElementID, doNotDisable, optionalOnEnableJavascriptAction )
{
	//Information about the element.
	this.ID = ElementID;
	this.oldCursor = "normal";
	this.ItemOnMouseDown = null;
	this.ItemOnMouseUp = null;
	this.ItemOnMouseOver = null;
	this.ItemOnMouseOut = null;
	this.ItemOnMouseClick = null;
	this.ItemOnBlur = null;
	this.ItemOnFocus = null;
	this.ItemOnChange = null;
	this.onEnableJS = optionalOnEnableJavascriptAction;
	
	//Function mappings
	this.enable = enableInternal;
	this.disable = disableInternal;
	
	//Enables an element. Enabling consists of changing the cursor
	//style back to the original style, and returning all the stored
	//javascript events to their original state. If the HTML element
	//is a button, the disabled property is simply set to false.
	function enableInternal()
	{
		if ( document.getElementById( this.ID ) ) {
			//Return the old cursor style.
			document.getElementById( this.ID ).style.cursor = this.oldCursor;
			
			//If it's a button, re-enable it.
			if ( document.getElementById( this.ID ).tagName == "BUTTON" )
			{
				document.getElementById( this.ID ).disabled = false;
			}
			else
			{
				//Return all the events.
				document.getElementById( this.ID ).onmousedown = this.ItemOnMouseDown;
				document.getElementById( this.ID ).onmouseup = this.ItemOnMouseUp;
				document.getElementById( this.ID ).onmouseover = this.ItemOnMouseOver;
				document.getElementById( this.ID ).onmouseout = this.ItemOnMouseOut;
				document.getElementById( this.ID ).onclick = this.ItemOnMouseClick;
				document.getElementById( this.ID ).onblur = this.ItemOnBlur;
				document.getElementById( this.ID ).onfocus = this.ItemOnFocus;
				document.getElementById( this.ID ).onchange = this.ItemOnChange;	
			}
			
			//Execute the onEnable Javascript, if specified.
			if ( this.onEnableJS != null )
			{
				eval( this.onEnableJS );
			}
		}	
	}
	
	//Disables an element. Disabling consists of changing the cursor
	//style to "not-allowed", and setting all the javascript events to
	//do nothing. If the HTML element is a button, the "disabled" property
	//is simply set to true.	
	function disableInternal() 
	{
		if ( document.getElementById( this.ID ) ) {
		
			//Set the cursor style to point out that you can't do anything yet
			this.oldCursor = document.getElementById( this.ID ).style.cursor;
			document.getElementById( this.ID ).style.cursor = "not-allowed";
		
			//If the HTML element is a BUTTON, we can easily disable it by
			//setting the disabled property to true.
			if ( document.getElementById( this.ID ).tagName == "BUTTON" )
			{
				document.getElementById( this.ID ).disabled = true;
			}
			else
			{
				//Store all the current events registered to the item.
				this.ItemOnMouseDown = document.getElementById( this.ID ).onmousedown;
				this.ItemOnMouseUp = document.getElementById( this.ID ).onmouseup;
				this.ItemOnMouseOver = document.getElementById( this.ID ).onmouseover;
				this.ItemOnMouseOut = document.getElementById( this.ID ).onmouseout;
				this.ItemOnMouseClick = document.getElementById( this.ID ).onclick;
				this.ItemOnBlur = document.getElementById( this.ID ).onblur;
				this.ItemOnFocus = document.getElementById( this.ID ).onfocus;
				this.ItemOnChange = document.getElementById( this.ID ).onchange;
				
				//Now set all the current events to do nothing.
				document.getElementById( this.ID ).onmousedown = function () { void(0); return false; };
				document.getElementById( this.ID ).onmouseup = function () { void(0); return false; };
				document.getElementById( this.ID ).onmouseover = function () { void(0); return false; };
				document.getElementById( this.ID ).onmouseout = function () { void(0); return false; };
				document.getElementById( this.ID ).onclick = function () { void(0); return false; };
				document.getElementById( this.ID ).onblur = function () { void(0); return false; };
				document.getElementById( this.ID ).onfocus = function () { void(0); return false; };
				document.getElementById( this.ID ).onchange = function () { void(0); return false; };
			}
		}	
	}		
	
	//Disable the element
	if ( !doNotDisable )
	{
		this.disable();	
	}
	
} 
 
// Global variables 
var wpsFLY_isIE = document.all?1:0;
var wpsFLY_isNetscape=document.layers?1:0;                             
var wpsFLY_isMoz = document.getElementById && !document.all;

// This sets how many pixels of the tab should show when collapsed was 11
var wpsFLY_minFlyout=0;

// How many pixels should it move every step? 
var wpsFLY_move=15;
if (wpsFLY_isIE)
   wpsFLY_move=12;

// Specify the scroll speed in milliseconds
var wpsFLY_scrollSpeed=1;

// Timeout ID for flyout
var wpsFLY_timeoutID=1;

// How from from top of screen for scrolling
var wpsFLY_fromTop=100;
var wpsFLY_leftResize;

//Cross browser access to required dimensions 
var wpsFLY_browserDimensions = new BrowserDimensions();

var wpsFLY_initFlyoutExpanded = wpsFLY_getInitialFlyoutState();

// Current state of the flyout for the life of the request (true=in, false=out)
var wpsFLY_state = true;

var wpsFLY_currIndex = -1;
// -----------------------------------------------------------------
// Initialize the Flyout
// -----------------------------------------------------------------
function wpsFLY_initFlyout(showHidden)
{
   wpsFLY_Flyout=new wpsFLY_makeFlyout('wpsFLYflyout');
   wpsFLY_Flyout.setWidth(wpsFLY_minFlyout);
   wpsFLY_Flyout.css.overflow = 'hidden';
   wpsFLY_Flyout.setLeft( wpsFLY_Flyout.pageWidth() - wpsFLY_minFlyout-1 );

   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      scrolled="window.pageYOffset";
   else if (wpsFLY_isIE)
      scrolled="document.body.scrollTop";

   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      wpsFLY_fromTop=wpsFLY_Flyout.css.top;
   else if (wpsFLY_isIE)
      wpsFLY_fromTop=wpsFLY_Flyout.css.pixelTop;

   if (wpsFLY_isIE) {
      window.onscroll=wpsFLY_internalScroll;
      window.onresize=wpsFLY_internalScroll;
   }
   else {
     window.onscroll=wpsFLY_internalScroll();
   }

   if (showHidden)
      wpsFLY_Flyout.css.visibility="hidden";
   else
      wpsFLY_Flyout.css.visibility="visible";

   //Open or close the flyout depending on the init state.
   if ( wpsFLY_initFlyoutExpanded != null )
   {
       wpsFLY_toggleFlyout( wpsFLY_initFlyoutExpanded, true );
   }

   return;
}

// -----------------------------------------------------------------
// Initialize the Flyout on left
// -----------------------------------------------------------------
function wpsFLY_initFlyoutLeft(showHidden)
{
   wpsFLY_FlyoutLeft=new wpsFLY_makeFlyoutLeft('wpsFLYflyout');

   if (wpsFLY_isIE) {
      wpsFLY_FlyoutLeft.setWidth(wpsFLY_minFlyout);
      wpsFLY_FlyoutLeft.css.overflow = 'hidden';
	   wpsFLY_FlyoutLeft.setLeft(0);
   } else {
      //  Mozilla does not move the scroll to the left for bidi languages
	   wpsFLY_FlyoutLeft.setLeft(wpsFLY_minFlyout - wpsFLY_FlyoutLeft.getWidth()- 4);
   }

   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      scrolled="window.pageYOffset";
   else if (wpsFLY_isIE)
      scrolled="document.body.scrollTop";

   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      wpsFLY_fromTop=wpsFLY_FlyoutLeft.css.top;
   else if (wpsFLY_isIE)
      wpsFLY_fromTop=wpsFLY_FlyoutLeft.css.pixelTop;

   if (wpsFLY_isIE) {
      window.onscroll=wpsFLY_internalScrollLeft;
      window.onresize=wpsFLY_internalResizeLeft;
   } else
      window.onscroll=wpsFLY_internalScrollLeft();

   if (showHidden)
      wpsFLY_FlyoutLeft.css.visibility="hidden";
   else
      wpsFLY_FlyoutLeft.css.visibility="visible";
      
   //Open or close the flyout depending on the init state.
   if ( wpsFLY_initFlyoutExpanded != null )
   {
       wpsFLY_toggleFlyout( wpsFLY_initFlyoutExpanded, true );
   }   
}


// -----------------------------------------------------------------
// Constructs flyout (default on right)
// -----------------------------------------------------------------
function wpsFLY_makeFlyout(obj)
{
   this.origObject=document.getElementById(obj);
   //get the css for the DIV tag, need it later
   if (wpsFLY_isNetscape)
      this.css=eval('document.'+obj);
   else if (wpsFLY_isMoz)
      this.css=document.getElementById(obj).style;
   else if (wpsFLY_isIE)
      this.css=eval(obj+'.style');

   //initialize the expand state
   wpsFLY_state=1;
   this.go=0;

   //get the width
   if (wpsFLY_isNetscape)
      this.width=this.css.document.width;
   else if (wpsFLY_isMoz)
      this.width=document.getElementById(obj).offsetWidth;
   else if (wpsFLY_isIE)
      this.width=eval(obj+'.offsetWidth');

   this.setWidth=wpsFLY_internalSetWidth;
   this.getWidth=wpsFLY_internalGetWidth;
   
   //set a left method to make it common across browsers
   this.left=wpsFLY_internalGetLeft;
   this.pageWidth=wpsFLY_internalGetPageWidth;
   this.setLeft = wpsFLY_internalSetLeft;

   this.obj = obj + "Object";   
   eval(this.obj + "=this");  
}

// -----------------------------------------------------------------
// Constructs flyout (on left)
// -----------------------------------------------------------------
function wpsFLY_makeFlyoutLeft(obj)        
{
   this.origObject=document.getElementById(obj);
	//get the css for the DIV tag, need it later
    if (wpsFLY_isNetscape) 
		this.css=eval('document.'+obj);
    else if (wpsFLY_isMoz) 
		this.css=document.getElementById(obj).style;
    else if (wpsFLY_isIE) 
		this.css=eval(obj+'.style');

	//initialize the expand state
	wpsFLY_state=1;
	this.go=0;
	
	//get the width
	if (wpsFLY_isNetscape) 
		this.width=this.css.document.width;
    else if (wpsFLY_isMoz) 
		this.width=document.getElementById(obj).offsetWidth;
    else if (wpsFLY_isIE) 
		this.width=eval(obj+'.offsetWidth');

   this.setWidth=wpsFLY_internalSetWidthLeft;
   this.getWidth=wpsFLY_internalGetWidthLeft;

	//set a left method to make it common across browsers
	this.left=wpsFLY_internalGetLeft;
   this.pageWidth=wpsFLY_internalGetPageWidth;
   this.setLeft = wpsFLY_internalSetLeft;

   this.obj = obj + "Object"; 	
   eval(this.obj + "=this");	
}


// -----------------------------------------------------------------
// The internal api to get the page width value that is cross browser
// -----------------------------------------------------------------
function wpsFLY_internalGetPageWidth()
{
   //get the width
   return wpsFLY_browserDimensions.getViewableAreaWidth();
}

function wpsFLY_internalSetLeft( value )
{
    this.css.left=value + "px";
}

// -----------------------------------------------------------------
// The internal api to set the width value that is cross browser
// -----------------------------------------------------------------
function wpsFLY_internalSetWidth(value)
{
   this.css.width = value + "px";
   if (navigator.userAgent.indexOf ("Opera") != -1) {
      var operaIframe=document.getElementById('wpsFLY_flyoutIFrame');
      operaIframe.style.width = (value-wpsFLY_minFlyout) + "px" ;
   }
}

// -----------------------------------------------------------------
// The internal api to set the width value that is cross browser
// -----------------------------------------------------------------
function wpsFLY_internalSetWidthLeft(value)
{
   this.css.width = value + "px";
   if (navigator.userAgent.indexOf ("Opera") != -1) {
      var operaIframe=document.getElementById('wpsFLY_flyoutIFrame');
      operaIframe.style.width = (value-wpsFLY_minFlyout) + "px" ;
   }
}

// -----------------------------------------------------------------
// The internal api to get the width value that is cross browser
// -----------------------------------------------------------------
function wpsFLY_internalGetWidth()
{
   //get the width
   if (wpsFLY_isNetscape)
      return eval(this.css.document.width);
   else if (wpsFLY_isMoz||wpsFLY_isIE)
      return eval(this.origObject.offsetWidth);
}

// -----------------------------------------------------------------
// The internal api to get the width value that is cross browser
// -----------------------------------------------------------------
function wpsFLY_internalGetWidthLeft()
{
   var width;
   if (wpsFLY_isNetscape)
      width = eval(this.css.document.width);
   else if (wpsFLY_isMoz||wpsFLY_isIE)
      width = eval(this.origObject.offsetWidth);
   return width;
}

// -----------------------------------------------------------------
// The internal api to get the left value that is cross browser
// -----------------------------------------------------------------
function wpsFLY_internalGetLeft()
{
   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      leftfunc=parseInt(this.css.left);
   else if (wpsFLY_isIE)
      leftfunc=eval(this.css.pixelLeft);
   return leftfunc;
}

// -----------------------------------------------------------------
// The internal fly out function, called my real function, only 
// wpsFLY_moveOutFlyout should call this function.
// -----------------------------------------------------------------
function wpsFLY_internalMoveOut()
{
	document.getElementById('wpsFLYflyout').className = "portalFlyoutExpanded";

   if (wpsFLY_Flyout.left() - wpsFLY_move > wpsFLY_Flyout.pageWidth()+ wpsFLY_browserDimensions.getScrollFromLeft() - wpsFLY_Flyout.width ) {
      var newwidth= wpsFLY_Flyout.getWidth()+wpsFLY_move;
      wpsFLY_Flyout.setWidth(newwidth);
      wpsFLY_Flyout.setLeft(wpsFLY_Flyout.left() - wpsFLY_move);
      wpsFLY_timeoutID=setTimeout("wpsFLY_internalMoveOut()",wpsFLY_scrollSpeed);
      wpsFLY_Flyout.go=1;
   } else {
      wpsFLY_Flyout.setLeft(wpsFLY_Flyout.pageWidth() + wpsFLY_browserDimensions.getScrollFromLeft() - wpsFLY_Flyout.width);
      wpsFLY_Flyout.setWidth(wpsFLY_Flyout.width);
      wpsFLY_Flyout.go=0;
      wpsFLY_state=0;
   }
}

// -----------------------------------------------------------------
// The internal slide out function, called my real function, only 
// wpsFLY_moveOutFlyoutLeft should call this function. For left sided
// flyout.
// -----------------------------------------------------------------
function wpsFLY_internalMoveOutLeft()
{
	document.getElementById('wpsFLYflyout').className = "portalFlyoutExpanded";

   if (wpsFLY_isIE) {
	   if (wpsFLY_FlyoutLeft.getWidth() + wpsFLY_move < wpsFLY_FlyoutLeft.width) {
         var newwidth= wpsFLY_FlyoutLeft.getWidth()+wpsFLY_move;
         wpsFLY_FlyoutLeft.setWidth(newwidth);
		   wpsFLY_timeoutID=setTimeout("wpsFLY_internalMoveOutLeft()",wpsFLY_scrollSpeed);
		   wpsFLY_FlyoutLeft.go=1;
	   } else {
         wpsFLY_FlyoutLeft.setLeft( wpsFLY_FlyoutLeft.left());
         wpsFLY_FlyoutLeft.setWidth(wpsFLY_FlyoutLeft.width);
		   wpsFLY_FlyoutLeft.go=0;
		   wpsFLY_state=0;
	   }   
   } else {
   //  Mozilla browsers don't scroll left
      if( wpsFLY_FlyoutLeft.left()+wpsFLY_move < wpsFLY_browserDimensions.getScrollFromLeft()) {
		   wpsFLY_FlyoutLeft.go=1;
		   wpsFLY_FlyoutLeft.setLeft(wpsFLY_FlyoutLeft.left()+wpsFLY_move);
		   wpsFLY_timeoutID=setTimeout("wpsFLY_internalMoveOutLeft()",wpsFLY_scrollSpeed);
	   } else {
         wpsFLY_FlyoutLeft.setLeft( wpsFLY_browserDimensions.getScrollFromLeft());
		   wpsFLY_FlyoutLeft.go=0;
		   wpsFLY_state=0;
	   }
   }  
}

// -----------------------------------------------------------------
// The internal fly in function, called my real function, only 
// wpsFLY_moveInFlyout should call this function.
// -----------------------------------------------------------------
function wpsFLY_internalMoveIn()
{
   if ( wpsFLY_Flyout.left() + wpsFLY_move < wpsFLY_Flyout.pageWidth() + wpsFLY_browserDimensions.getScrollFromLeft() - wpsFLY_minFlyout  ) {
      wpsFLY_Flyout.go=1;
      var newwidth= wpsFLY_Flyout.getWidth()-wpsFLY_move;
      wpsFLY_Flyout.setWidth(newwidth);
      wpsFLY_Flyout.setLeft(wpsFLY_Flyout.left()+wpsFLY_move);
      wpsFLY_timeoutID=setTimeout("wpsFLY_internalMoveIn()",wpsFLY_scrollSpeed);
   } else {
      wpsFLY_Flyout.setWidth(wpsFLY_minFlyout);
      wpsFLY_Flyout.setLeft(wpsFLY_Flyout.pageWidth() + wpsFLY_browserDimensions.getScrollFromLeft() - wpsFLY_minFlyout);
      wpsFLY_Flyout.go=0;
      wpsFLY_state=1;
   }  
}

// -----------------------------------------------------------------
// The internal slide in function, called my real function, only 
// wpsFLY_moveInFlyoutLeft should call this function. For left sided
// flyout.
// -----------------------------------------------------------------
function wpsFLY_internalMoveInLeft()
{
   if (wpsFLY_isIE) {
      if (wpsFLY_FlyoutLeft.getWidth() - wpsFLY_move >  wpsFLY_minFlyout) {
         var newwidth= wpsFLY_FlyoutLeft.getWidth() - wpsFLY_move;
         wpsFLY_FlyoutLeft.setWidth(newwidth);
         wpsFLY_timeoutID=setTimeout("wpsFLY_internalMoveInLeft()",wpsFLY_scrollSpeed);
         wpsFLY_FlyoutLeft.go=1;
      } else {
         wpsFLY_FlyoutLeft.setWidth(wpsFLY_minFlyout);
         wpsFLY_FlyoutLeft.setLeft( wpsFLY_FlyoutLeft.left());
         wpsFLY_FlyoutLeft.go=0;
         wpsFLY_state=1;
      }
   } else {
      if(wpsFLY_FlyoutLeft.left()>-wpsFLY_FlyoutLeft.width+wpsFLY_minFlyout) {
		   wpsFLY_FlyoutLeft.go=1;
		   wpsFLY_FlyoutLeft.setLeft(wpsFLY_FlyoutLeft.left()-wpsFLY_move);
		   wpsFLY_timeoutID=setTimeout("wpsFLY_internalMoveInLeft()",wpsFLY_scrollSpeed);
	   } else {
         wpsFLY_FlyoutLeft.setLeft( wpsFLY_minFlyout - wpsFLY_FlyoutLeft.getWidth()- 4 );
		   wpsFLY_FlyoutLeft.go=0;
		   wpsFLY_state=1;
      }
	}
}

// -----------------------------------------------------------------
// The internal scroll function.
// -----------------------------------------------------------------
function wpsFLY_internalScroll() {
   if (!wpsFLY_Flyout.go) {

      //wpsFLY_Flyout.css.top=eval(scrolled)+parseInt(wpsFLY_fromTop);
      if (wpsFLY_state==1) {
         wpsFLY_Flyout.setLeft(wpsFLY_browserDimensions.getScrollFromLeft() + wpsFLY_browserDimensions.getViewableAreaWidth() - wpsFLY_minFlyout);
      } else {
         wpsFLY_Flyout.setLeft(wpsFLY_browserDimensions.getScrollFromLeft() + wpsFLY_browserDimensions.getViewableAreaWidth() - wpsFLY_Flyout.width);
      }
   }
   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      setTimeout('wpsFLY_internalScroll()',20);
}

// -----------------------------------------------------------------
// The internal scroll left function.
// -----------------------------------------------------------------
function wpsFLY_internalScrollLeft() {
   if (!wpsFLY_FlyoutLeft.go) {
      //wpsFLY_FlyoutLeft.css.top=eval(scrolled)+parseInt(wpsFLY_fromTop);

      // scroll horizontally for flyoutin 
      if (wpsFLY_state==1) {
         if (wpsFLY_isIE) {
            if (wpsFLY_leftResize == null) {
               wpsFLY_leftResize = wpsFLY_browserDimensions.getScrollFromLeft();
            }
            wpsFLY_FlyoutLeft.setWidth(wpsFLY_minFlyout);
            wpsFLY_FlyoutLeft.css.overflow = 'hidden';
            wpsFLY_FlyoutLeft.setLeft(wpsFLY_browserDimensions.getScrollFromLeft() - wpsFLY_leftResize);
         } else {
            wpsFLY_FlyoutLeft.setLeft(wpsFLY_minFlyout + wpsFLY_browserDimensions.getScrollFromLeft() - wpsFLY_FlyoutLeft.getWidth() - 4);
         }
      }
   }

   if (wpsFLY_isNetscape||wpsFLY_isMoz)
      setTimeout('wpsFLY_internalScrollLeft()',20);

}

// -----------------------------------------------------------------
// The internal resize left function.
// -----------------------------------------------------------------
function wpsFLY_internalResizeLeft(){
      
   if (wpsFLY_isIE) {
      wpsFLY_leftResize = wpsFLY_browserDimensions.getScrollFromLeft(); - wpsFLY_browserDimensions.getViewableAreaWidth();
   }

}

// -----------------------------------------------------------------
// Expand the flyout. The parameter skipSlide indicates whether or not
// the flyout should simply be rendered without the slide-out effect.
// ----------------------------------------------------------------- 
function wpsFLY_moveOutFlyout( skipSlide )
{
   if (this.wpsFLY_Flyout != null)
   {
      if ( wpsFLY_state && !skipSlide ) {
         clearTimeout(wpsFLY_timeoutID);
         wpsFLY_internalMoveOut(); 
      }
      
      if ( wpsFLY_state && skipSlide )
      {
          wpsFLY_Flyout.setLeft(wpsFLY_Flyout.pageWidth() + document.body.scrollLeft - wpsFLY_Flyout.width);
          wpsFLY_Flyout.setWidth(wpsFLY_Flyout.width);
          wpsFLY_Flyout.go=0;
          wpsFLY_state=0;
          document.getElementById('wpsFLYflyout').className = "portalFlyoutExpanded";
      }
      
   }

   if (this.wpsFLY_FlyoutLeft != null)
   {      
      if ( wpsFLY_state && !skipSlide ) {
         clearTimeout(wpsFLY_timeoutID);
         wpsFLY_internalMoveOutLeft(); 
      }
      
      if ( wpsFLY_state && skipSlide )
      {
      	 if ( wpsFLY_isIE )
      	 {
		     wpsFLY_FlyoutLeft.setLeft( wpsFLY_FlyoutLeft.left());
    	     wpsFLY_FlyoutLeft.setWidth(wpsFLY_FlyoutLeft.width);
			 wpsFLY_FlyoutLeft.go=0;
			 wpsFLY_state=0;	
		 }
		 else
		 {
		 	 wpsFLY_FlyoutLeft.setLeft( document.body.scrollLeft);
   		     wpsFLY_FlyoutLeft.go=0;
		     wpsFLY_state=0;
		 }
		 document.getElementById('wpsFLYflyout').className = "portalFlyoutExpanded";
	  }
   }
}

// -----------------------------------------------------------------
// Called to close the flyout. This is the method that the function
// external to the flyout should call.
// ----------------------------------------------------------------- 
function wpsFLY_moveInFlyout()
{
   if (this.wpsFLY_Flyout != null)
   {
      if (!wpsFLY_state) {
         clearTimeout(wpsFLY_timeoutID);
         wpsFLY_internalMoveIn();
      }
   }

   if (this.wpsFLY_FlyoutLeft != null)
   {
      if (!wpsFLY_state) {
         clearTimeout(wpsFLY_timeoutID);
         wpsFLY_internalMoveInLeft();
      }
   }
   
   document.getElementById('wpsFLYflyout').className = "portalFlyoutCollapsed";
}

// -----------------------------------------------------------------
// Called to toggle the flyout. This is the method that the function
// external to the flyout should call.
// -----------------------------------------------------------------
function wpsFLY_toggleFlyout(index, skipSlide)
{
	if(flyOut[index] != null){
	var checkIndex = index;
	var prevIndex=wpsFLY_getCurrIndex();
	
	if(checkIndex==prevIndex){
        
		if(flyOut[index].active==true){
			flyOut[index].active=false;
            /*
			document.getElementById("toolBarIcon"+prevIndex).src = flyOut[prevIndex].icon;	
			document.getElementById("toolBarIcon"+prevIndex).alt = flyOut[prevIndex].altText;
			document.getElementById("toolBarIcon"+prevIndex).title = flyOut[prevIndex].altText;
			*/
		}
		else{
			flyOut[index].active=true;
            /*
			document.getElementById("toolBarIcon"+index).src = flyOut[index].activeIcon;
			document.getElementById("toolBarIcon"+index).alt = flyOut[index].activeAltText;
			document.getElementById("toolBarIcon"+index).title = flyOut[index].activeAltText;
			*/
		}
        //Closing flyout, clear the state cookie.
        wpsFLY_clearStateCookie();
        wpsFLY_moveInFlyout();
	}else{
		if(prevIndex > -1){
			flyOut[prevIndex].active=false;
            /*
			document.getElementById("toolBarIcon"+prevIndex).src = flyOut[prevIndex].icon;	
			document.getElementById("toolBarIcon"+prevIndex).alt = flyOut[prevIndex].altText;
			document.getElementById("toolBarIcon"+prevIndex).title = flyOut[prevIndex].altText;
			*/
		}
		
		flyOut[index].active=true;
        /*
		document.getElementById("toolBarIcon"+index).src = flyOut[index].activeIcon;
		document.getElementById("toolBarIcon"+index).alt = flyOut[index].activeAltText;
		document.getElementById("toolBarIcon"+index).title = flyOut[index].activeAltText;
		*/
		wpsFLY_setCurrIndex(index);
		document.getElementById("wpsFLY_flyoutIFrame").src = flyOut[index].url;
	}
	
	if(wpsFLY_state){
		
        //Expanding flyout, store the open flyout index in the state cookie.
        wpsFLY_setStateCookie( index );
        wpsFLY_moveOutFlyout( skipSlide );
		}
	}
}

function wpsFLY_getCurrIndex()
{
	return wpsFLY_currIndex;
}

function wpsFLY_setCurrIndex(index)
{
	wpsFLY_currIndex = index;
}

// -----------------------------------------------------------------
// Create a cookie to track the state of the flyout. The value of the 
// cookie is the index of the open flyout. The cookie is marked for 
// deletion when the browser closes and the path is set to ensure the 
// cookie is valid for the entire site.
// -----------------------------------------------------------------
function wpsFLY_setStateCookie( index )
{
    document.cookie='portalOpenFlyout=' + index + '; path=/;';
}

// -----------------------------------------------------------------
// Clear the cookie by changing the value to null. By setting the expiration date in 
// the past, the cookie is marked for deletion when the browser closes.
// -----------------------------------------------------------------
function wpsFLY_clearStateCookie()
{
    document.cookie='portalOpenFlyout=null; expires=Wed, 1 Jan 2003 11:11:11 UTC; path=/;';
}

// -----------------------------------------------------------------
// Check which side of the page the flyout should show on
// -----------------------------------------------------------------
function wpsFLY_onloadShow( isRTL )
{
  if (this.wpsFLY_minFlyout != null) {
    var bodyObj = document.getElementById("FLYParent");
    if (bodyObj != null) {
      var showHidden = false;
      if (isRTL) { 
           bodyObj.onload = wpsFLY_initFlyoutLeft(showHidden);
       } else { 
      	   bodyObj.onload = wpsFLY_initFlyout(showHidden);
       } 	 
    }
  }
 }
// -----------------------------------------------------------------
// Write markup out to document for all flyout items
// -----------------------------------------------------------------
function wpsFLY_markupLoop( flyOut)
{	
 	for(arrayIndex = 0; arrayIndex < flyOut.length; arrayIndex++){
		if(flyOut[arrayIndex].url != "" && flyOut[arrayIndex].url != null){
			document.write('<li><a id="globalActionLink'+arrayIndex+'" href="javascript:void(0);" onclick="wpsFLY_toggleFlyout('+arrayIndex+'); return false;" >');
			document.write(flyOut[arrayIndex].altText);
            //document.write('<img src="'+flyOut[arrayIndex].icon+'" id="toolBarIcon'+arrayIndex+'" title="'+flyOut[arrayIndex].altText+'" border="0" alt="'+flyOut[arrayIndex].altText+'" onmouseover="this.src=flyOut['+arrayIndex+'].hoverIcon;" onmouseout="if (flyOut['+arrayIndex+'].active) {this.src=flyOut['+arrayIndex+'].activeIcon;} else this.src=flyOut['+arrayIndex+'].icon;" />');            
			document.write('</a></li>');
		}
		
		if ( javascriptEventController )
		{
			javascriptEventController.register( "globalActionLink" + arrayIndex );
			//javascriptEventController.register( "toolBarIcon" + arrayIndex );
		}
	}
}

// -----------------------------------------------------------------
// If we have an empty expanded flyout (via the back button), load 
// the previously open flyout.
// -----------------------------------------------------------------
function wpsFLY_checkForEmptyExpandedFlyout()
{
	var index = wpsFLY_getInitialFlyoutState();
	
	if ( index != null && flyOut[index] != null)
	{
		document.getElementById("wpsFLY_flyoutIFrame").src = flyOut[index].url;
	}
}

// -----------------------------------------------------------------
// Determine if the flyout should initially open and which flyout 
// should be loaded. 
// -----------------------------------------------------------------
function wpsFLY_getInitialFlyoutState()
{
	// Determine if the flyout's initial state is open or closed.
	if ( document.cookie.indexOf( "portalOpenFlyout=" ) >= 0 )
	{
	    var cookies = document.cookie.split(";");
	
	    var c = 0;
	    while ( c < cookies.length && ( cookies[c].indexOf( "portalOpenFlyout=" ) == -1 ) )
	    {
	        c=c+1;
	    }
	
	    initCookieValue = cookies[c].substring( 18, cookies[c].length );
	    
	    if ( initCookieValue != "null" )
	    {
	        return initCookieValue;
	    }
	    else
	    {
	        return null;
	    }
	}
	else
	{
		return null;
	}
	
}
 
var wpsInlineShelf_initShelfExpanded = wpsInlineShelf_getInitialShelfState();

// Current state of the flyout for the life of the request (true=expanded, false=collapsed)
var wpsInlineShelf_stateExpanded = false;

var wpsInlineShelf_currIndex = -1;

var wpsInlineShelf_loadingMsg = null;

function wpsInlineShelf_markupLoop( shelves )
{	
    document.write('<ul>');
 	for(arrayIndex = 0; arrayIndex < shelves.length; arrayIndex++){
		if(shelves[arrayIndex].url != "" && shelves[arrayIndex].url != null){
			document.write('<li><a id="globalActionLinkInlineShelf'+arrayIndex+'" href="javascript:void(0);" onclick="wpsInlineShelf_toggleShelf('+arrayIndex+'); return false;" >');
			document.write(shelves[arrayIndex].altText+" ");
            document.write('<img src="'+shelves[arrayIndex].icon+'" id="toolBarIcon'+arrayIndex+'" title="'+shelves[arrayIndex].altText+'" border="0" alt="'+shelves[arrayIndex].altText+'" onmouseover="this.src=wptheme_InlineShelves['+arrayIndex+'].hoverIcon;" onmouseout="if (wptheme_InlineShelves['+arrayIndex+'].active) {this.src=wptheme_InlineShelves['+arrayIndex+'].activeIcon;} else this.src=wptheme_InlineShelves['+arrayIndex+'].icon;" />');            
			document.write('</a></li>');
		}
		
		if ( javascriptEventController )
		{
			javascriptEventController.register( "globalActionLinkInlineShelf" + arrayIndex );
		}
	}
    document.write('</ul>');
}

// -----------------------------------------------------------------
// Called to toggle the shelf. This is the method that the function
// external to the shelf should call.
// -----------------------------------------------------------------
function wpsInlineShelf_toggleShelf(index, skipZoom)
{
	if(wptheme_InlineShelves[index] != null) {
    	var checkIndex = index;
    	var prevIndex=wpsInlineShelf_getCurrIndex();
        var newIframeUrl = null;

    	if(checkIndex==prevIndex){
            
    		if(wptheme_InlineShelves[index].active==true){
    			wptheme_InlineShelves[index].active=false;
                /*
    			document.getElementById("globalActionLinkInlineShelf"+prevIndex).src = wptheme_InlineShelves[prevIndex].icon;	
    			document.getElementById("globalActionLinkInlineShelf"+prevIndex).alt = wptheme_InlineShelves[prevIndex].altText;
    			document.getElementById("globalActionLinkInlineShelf"+prevIndex).title = wptheme_InlineShelves[prevIndex].altText;
    			*/
                wpsInlineShelf_stateExpanded = false;
    		}else{
    			wptheme_InlineShelves[index].active=true;
                /*
    			document.getElementById("globalActionLinkInlineShelf"+index).src = wptheme_InlineShelves[index].activeIcon;
    			document.getElementById("globalActionLinkInlineShelf"+index).alt = wptheme_InlineShelves[index].activeAltText;
    			document.getElementById("globalActionLinkInlineShelf"+index).title = wptheme_InlineShelves[index].activeAltText;
    			*/
                wpsInlineShelf_stateExpanded = true;
    		}
    	}else{
    		if(prevIndex > -1){
    			wptheme_InlineShelves[prevIndex].active=false;
                /*
    			document.getElementById("globalActionLinkInlineShelf"+prevIndex).src = wptheme_InlineShelves[prevIndex].icon;	
    			document.getElementById("globalActionLinkInlineShelf"+prevIndex).alt = wptheme_InlineShelves[prevIndex].altText;
    			document.getElementById("globalActionLinkInlineShelf"+prevIndex).title = wptheme_InlineShelves[prevIndex].altText;
    			*/
    		}
    		
    		wptheme_InlineShelves[index].active=true;
            /*
    		document.getElementById("globalActionLinkInlineShelf"+index).src = wptheme_InlineShelves[index].activeIcon;
    		document.getElementById("globalActionLinkInlineShelf"+index).alt = wptheme_InlineShelves[index].activeAltText;
    		document.getElementById("globalActionLinkInlineShelf"+index).title = wptheme_InlineShelves[index].activeAltText;
    		*/
    		wpsInlineShelf_setCurrIndex(index);
            wpsInlineShelf_stateExpanded = true;

            newIframeUrl = wptheme_InlineShelves[index].url;
//    		document.getElementById("wpsInlineShelf_shelfIFrame").src = wptheme_InlineShelves[index].url;
    	}
    	
    	if(wpsInlineShelf_stateExpanded){
            //Expanding flyout, store the open shelf index in the state cookie.
            wpsInlineShelf_setStateCookie( index );
            wpsInlineShelf_expandShelf( skipZoom, newIframeUrl );
		} else {
            //Closing shelf, clear the state cookie.
            wpsInlineShelf_clearStateCookie();
            wpsInlineShelf_collapseShelf();
        }
	}
}


function wpsInlineShelf_getCurrIndex()
{
	return wpsInlineShelf_currIndex;
}

function wpsInlineShelf_setCurrIndex(index)
{
	wpsInlineShelf_currIndex = index;
}

// -----------------------------------------------------------------
// Create a cookie to track the state of the shelf. The value of the 
// cookie is the index of the open shelf. The cookie is marked for 
// deletion when the browser closes and the path is set to ensure the 
// cookie is valid for the entire site.
// -----------------------------------------------------------------
function wpsInlineShelf_setStateCookie( index )
{
    document.cookie='portalOpenInlineShelf=' + index + '; path=/;';
}

// -----------------------------------------------------------------
// Clear the cookie by changing the value to null. By setting the expiration date in 
// the past, the cookie is marked for deletion when the browser closes.
// -----------------------------------------------------------------
function wpsInlineShelf_clearStateCookie()
{
    document.cookie='portalOpenInlineShelf=null; expires=Wed, 1 Jan 2003 11:11:11 UTC; path=/;';
}

// -----------------------------------------------------------------
// Determine if the shelf should initially open and which shelf 
// should be loaded. 
// -----------------------------------------------------------------
function wpsInlineShelf_getInitialShelfState()
{
	// Determine if the shelf's initial state is expanded or collapsed.
	if ( document.cookie.indexOf( "portalOpenInlineShelf=" ) >= 0 )
	{
	    var cookies = document.cookie.split(";");
	
	    var c = 0;
	    while ( c < cookies.length && ( cookies[c].indexOf( "portalOpenInlineShelf=" ) == -1 ) )
	    {
	        c=c+1;
	    }
	
	    initCookieValue = cookies[c].substring( ("portalOpenInlineShelf=".length)+1, cookies[c].length );
	    
	    if ( initCookieValue != "null" )
	    {
	        return initCookieValue;
	    }
	    else
	    {
	        return null;
	    }
	}
	else
	{
		return null;
	}
	
}


// -----------------------------------------------------------------
// Expand the shelf. The parameter skipZoom indicates whether or not
// the shelf should simply be rendered without the zoom-out effect.
// NOTE: The zoom-out effect is not implemented yet.
// ----------------------------------------------------------------- 
function wpsInlineShelf_expandShelf( skipZoom, newIframeUrl )
{
    var shelf = document.getElementById("wpsInlineShelf");
    
    wpsInlineShelf_stateExpanded = false;

    // attach event listeners so when the URL loads or reloads, the iframe will be shown and resized.
    wpsInlineShelf_AttachIframeEventListeners("wpsInlineShelf_shelfIFrame");

    // show the shelf... but not the iframe yet...
    shelf.style.display = "block";

    // We change the URL AFTER the event listeners are hooked up.
    // If we are not changing the URL, we need to manually resize the iframe.
    if (null != newIframeUrl) {
        // when loading a new URL, display the spinning loading graphic
        wpsInlineShelf_loadingMsg.show(document.getElementById("wpsInlineShelf"));
        document.getElementById("wpsInlineShelf_shelfIFrame").src = newIframeUrl;
    } else {
        wpsInlineShelf_resizeIframe("wpsInlineShelf_shelfIFrame");
    }

}


// -----------------------------------------------------------------
// Collapse the shelf. The parameter skipZoom indicates whether or not
// the shelf should simply be rendered without the zoom-in effect.
// NOTE: The zoom-in effect is not implemented yet yet.
// ----------------------------------------------------------------- 
function wpsInlineShelf_collapseShelf( skipZoom )
{
    var shelf = document.getElementById("wpsInlineShelf");
    var iframe = document.getElementById("wpsInlineShelf_shelfIFrame");

    shelf.style.display = "none";
    iframe.style.display = "none";
    wpsInlineShelf_loadingMsg.hide();
    wpsInlineShelf_stateExpanded = true;
}

// -----------------------------------------------------------------
// Check which side of the page the shelf should show on
// -----------------------------------------------------------------
function wpsInlineShelf_onloadShow( isRTL )
{
    if ( wpsInlineShelf_initShelfExpanded != null )
    {
        wpsInlineShelf_toggleShelf( wpsInlineShelf_initShelfExpanded, true );
    }   
}


var wpsInlineShelf_getFFVersion=navigator.userAgent.substring(navigator.userAgent.indexOf("Firefox")).split("/")[1]
var wpsInlineShelf_FFextraHeight=parseFloat(wpsInlineShelf_getFFVersion)>=0.1? 16 : 0 //extra height in px to add to iframe in FireFox 1.0+ browsers

function wpsInlineShelf_resizeIframe(iframeID){
    var iframe=document.getElementById(iframeID)

    iframe.style.display = "block";
    wpsInlineShelf_loadingMsg.hide();

    if (iframe && !window.opera) {

/*
		// put the background color style onto the body of the document in the iframe
		if (iframe.contentDocument) {
			var iframeDocBody = iframe.contentDocument.body;
			if (iframeDocBody.className.indexOf("wpsInlineShelfIframeDocBody",0) == -1) {
				iframeDocBody.className = iframeDocBody.className + " wpsInlineShelfIframeDocBody";
			}
		}
*/
        if (iframe.contentDocument && iframe.contentDocument.body.offsetHeight) //ns6 syntax
            iframe.height = iframe.contentDocument.body.offsetHeight+wpsInlineShelf_FFextraHeight;
        else if (iframe.Document && iframe.Document.body.scrollHeight) //ie5+ syntax
            iframe.height = iframe.Document.body.scrollHeight;
    }
}

function wpsInlineShelf_AttachIframeEventListeners(iframeID) {
    var iframe=document.getElementById(iframeID)
    if (iframe && !window.opera) {

        if (iframe.addEventListener){
            iframe.addEventListener("load", wpsInlineShelf_IframeOnloadEventListener, false)
        } else if (iframe.attachEvent){
            iframe.detachEvent("onload", wpsInlineShelf_IframeOnloadEventListener) // Bug fix line
            iframe.attachEvent("onload", wpsInlineShelf_IframeOnloadEventListener)
        }
    }
}

function wpsInlineShelf_IframeOnloadEventListener(loadevt) {
    var crossevt=(window.event)? event : loadevt
    var iframeroot=(crossevt.currentTarget)? crossevt.currentTarget : crossevt.srcElement
    if (iframeroot) {
        wpsInlineShelf_resizeIframe(iframeroot.id);
    }
}

// -----------------------------------------------------------------
// If we have an empty expanded shelf (via the back button), load 
// the previously open shelf.
// -----------------------------------------------------------------
function wpsInlineShelf_checkForEmptyExpandedShelf() {
	var index = wpsInlineShelf_getInitialShelfState();
    
	if ( index != null && wptheme_InlineShelves[index] != null)
	{
		document.getElementById("wpsInlineShelf_shelfIFrame").src = wptheme_InlineShelves[index].url;
	}
}

wptheme_QuickLinksShelf = {
	_cookieUtils: wptheme_CookieUtils,
	cookieName: null,
	expand: function () {
		// summary: Expand the quick links shelf.
		document.getElementById("wptheme-expandedQuickLinksShelf").style.display='block';
        document.getElementById("wptheme-collapsedQuickLinksShelf").style.display='none';
        if ( this.cookieName != null ) {
        	this._cookieUtils.deleteCookie( this.cookieName );
        }	
        return false;
	},
	collapse: function () {
		// summary: Collapse the quick links shelf.
		document.getElementById("wptheme-expandedQuickLinksShelf").style.display='none';
        document.getElementById("wptheme-collapsedQuickLinksShelf").style.display='block';
        if ( this.cookieName != null ) {
        	var expires = new Date();
        	expires.setDate( expires.getDate() + 5 );
        	this._cookieUtils.setCookie( this.cookieName, "small", expires );
        }
        return false;
	}
}

var wpsInlineShelf_LoadingImage = function ( /*String*/cssClassName, /*String*/imageURL, /*String*/imageText ) {
    // summary: creates loading image for inline shelf
    // cssClassName: the class name to apply to the overlaid div
    // imageURL: the url to the image to display in the center of the overlaid div
    // imageText: the text to display next to the image    

    var elem = document.createElement( "DIV" );
    elem.className = cssClassName;
    elem.id = cssClassName;

    if ( imageURL && imageURL != "" && imageText ) {
        elem.innerHTML = "<img src='"+ imageURL + "' border=\"0\" alt=\"\" />&nbsp;" + imageText;
    }   
    
    var appended = false;
    
    this.show = function ( refNode ) { 
        if ( !appended ) {
            refNode.appendChild( elem );
            appended= true;
        }
        
        elem.style.display = 'block';
    }

    this.hide = function () {
        elem.style.display = 'none';
    }
}
 
var wptheme_InlinePalettes = {
    // summary: Manages the inline palette(s). Tracks the iframe used to display the palettes, as well as the persistent state.
    // description: Palettes managed by this object must be added using the addPalette method. The addPalette method expects a
    //      PaletteContext object which has the following structure:
    //      {
    //          url: the url to set the iframe to (for CSA, only used as the initial url)
    //          page: the unique id of the page the url will be pointing to
    //          portlet: the unique id of the portlet control the url will be pointing to
    //          newWindow: indicates whether or not the url should be rendered using the Plain theme template
    //          portletWindowState: the window state the portlet should be in
    //          selectionDependent: indicates whether or not the url changes based on the current page selection
    //      }
    //      This PaletteContext object is required to support client-side aggregation. Since, in client-side aggregation,
    //      the page does not always reload in between page changes, the palette may need to be refreshed as the selection
    //      changes in client-side aggregation. This context object gives the client-side aggregation engine all the info
    //      it needs to create the appropriate url for the palette, as needed.
    // paletteStatus: indicates whether the palette is open or closed (0 = closed, 1 = open)
    // iframeID: the ID/NAME of the iframe to use
    // loadingDecorator: the decoration to display while the iframe is loading
    // currentIndex: the index (into the paletteContextArray) of the currently displaying palette
    // cookieName: the cookie used to store the state of the palette
    // paletteContextArray: the array of PaletteContext objects
    // urlFactory: only used in client-side aggregation -- set during the CSA theme's bootstrap, this function takes the PaletteContext
    //      object described above as the only parameter and returns the url to use to display the palette
    
    // INITIALIZATION
    className: "wptheme_InlinePalettes",

    //HTML element IDs
    iframeID: "wpsFLY_flyoutIFrame",    
    
    //Persistent state of the palette
    paletteStatus: 0,  // 0 = closed, 1 = open
    currentIndex: -1,   // the index of the paletteContextArray which is currently displaying
    cookieName: "portalOpenFlyout", 
    
    //Decoration to display while the palette is loading
    loadingDecorator: null, //needs to provide 2 functions: show and hide. show will receive a single parameter -- the node which should be covered with the decoration
    
    //Instance variables
    urlFactory: null,   //expected to be set in the bootstrap of the CSA theme. takes the context as a single parameter and returns the url as the output.
    paletteContextArray: [],    // Array of palettes which can be displayed
    
    //Debug
    debug: wptheme_DebugUtils,
    
    init: function ( /*Document?*/doc) {
        // summary: Initializes the inline palettes. Usually executed on page load.
        // description: Checks to see if the persisted state indicates the palette should be open or closed. If open, the proper location
        //      should be loaded into the iframe and displayed.
        // doc: OPTIONAL -- specifies the document to use when initializing (for use when called from within an iframe, for example).
        if ( this.debug.enabled ) { this.debug.log( this.className, "init( " + [ doc ] + ")" ); }
        
        if ( !doc ) { doc = document; }
        
        //Retrieve the persisted value. This will be the index into the PaletteContextArray.
        var value = this.getPersistedValue();
        if ( this.debug.enabled ) { this.debug.log( this.className, "retrieved value: " + value ); }
        if ( value != null && this.paletteContextArray[ value ] ) {
            this.show( value, true );
        } else {
            this.hide();
        }
        
        if ( this.debug.enabled ) { this.debug.log( this.className, "return init" ); }
    },
    
    // DISPLAY CONTROL 
    
    showCurrent: function () {
        // summary: Displays the current index or auto selects an index if no current index is selected.
        var indexToShow = 0;
        if ( this.currentIndex > -1 ) {
            indexToShow = this.currentIndex;
        }
        
        this.show( indexToShow );
    },
    
    show: function (/*int*/index, /*boolean?*/skipAnimation) {
        // summary: Displays the specified url in the palette.
        // url: the url for the iframe.
        // skipAnimation: OPTIONAL -- skips the loading decorator show/hide steps (used for the case where the palette is open on an initial page load
        if ( this.debug.enabled ) { this.debug.log( this.className, "show( " + [ index, skipAnimation ] + ")" ); }
        
        var iframe = this._getIframeElement();
        if ( !iframe ) { return false; }
        
        var url = this.getURL( index );
        var iframeLocation = window.frames[this.iframeID].location;
        
        if ( this.debug.enabled ) {
        	this.debug.log( this.className, "Url returned from getUrl is: " + url );
        	this.debug.log( this.className, "Current iframe url is: " + iframeLocation.href );
        	this.debug.log( this.className, "Are they equal? " + ( url == iframeLocation.href ) );
        }
        
        iframe.parentNode.style.display = "block";
        //If we have to load the iframe, call postShow onload. Otherwise, call it immediately since the
        //iframe is already loaded.
        //In CSA, the state serialization service returns the url without the protocol, host, and port while
        //the iframe url includes this information. So we compare the complete href value AND the pathname value of
        //the iframe location.
        if ( iframeLocation.href != url && iframeLocation.pathname != url  ) {
            if ( !skipAnimation && this.loadingDecorator != null && this.loadingDecorator.show ) {
                this.loadingDecorator.show( iframe.parentNode.parentNode );
            }
            iframe.src = url;
        }
        else {
            //The location hasn't changed so go ahead and call the post show behavior. Normally, the post show 
            //behavior executes once the iframe is loaded. 
            this._doPostShow();
        }
        
        this.persist( index );
        this.paletteStatus = 1;
        this.currentIndex = index;
    },
    hide: function ( doc ) {
        // summary: Hides the active palette.
        if ( this.debug.enabled ) { this.debug.log( this.className, "hide( " + [ doc ] + ")" ); }
        var iframe = this._getIframeElement( doc );
        if ( !iframe ) { return false; }
        
        iframe.parentNode.style.display = "none";
        this.paletteStatus = 0;
        this.currentIndex = -1;
        
        //Execute the post hide behavior.
        this._doPostHide();
    },
    _doPostShow: function () {
        // summary: Called after the iframe is loaded and ready to display.
        // description: Performs any sizing adjustments necessary (possibly IE) and hides the loading decoration.
        if ( this.debug.enabled ) { this.debug.log( this.className, "_doPostShow()" ); }
        var iframe = this._getIframeElement();
        if ( iframe.parentNode.style.display == "none" ) { return false; }
        iframe.style.visibility = "visible";
        
        if ( typeof ( dojo ) != "undefined" ) {
            var size = dojo.contentBox( iframe );
            if ( size.h < 300) {
                //IE doesn't correctly size the iframe when height is set to 100%. So if the height
                //is still 0 (IE 6) or small (IE7)after setting the display and visibility, set it manually to the height
                //of the TD element.
                var size = dojo.contentBox( iframe.parentNode.parentNode );
                iframe.style.height = size.h + "px";
            }
        }   
        
        if ( this.loadingDecorator != null && this.loadingDecorator.hide ) {
            this.loadingDecorator.hide();
        }
    },
    _doPostHide: function () {
        // summary: Execute any actions that need to occur after the palette is hidden from view.
        if ( this.debug.enabled ) { this.debug.log( this.className, "_doPostHide()" ); }
        var iframe = this._getIframeElement();
        iframe.style.visibility = "hidden";
    },
    
    // PERSISTENT STATE CONTROL
    
    persist: function ( /*String*/value ) {
        // summary: Persist the given value in a cookie.
        if ( this.debug.enabled ) { this.debug.log( this.className, "persist(" + [ value ] + ")" ); }
        wptheme_CookieUtils.setCookie( this.cookieName, value );
    },
    getPersistedValue: function () {
        // summary: Retrieve the persisted state for the inline palettes, if one exists.
        // description: Looks for the "portalOpenFlyout" cookie and parses out it's value.
        // returns: the persisted value for the portalOpenFlyout cookie or NULL if no value exists.
        if ( this.debug.enabled ) { this.debug.log( this.className, "getPersistedValue()" ); }
        return wptheme_CookieUtils.getCookie( this.cookieName );
    },
    unpersist: function () {
        // summary: Clears out the persisted value.
        // description: Sets the cookie's value to NULL and sets it to expire in the past.
        // returns: the index of the persisted value
        if ( this.debug.enabled ) { this.debug.log( this.className, "unpersist()" ); }
        var value = this.getPersistedValue();
        wptheme_CookieUtils.deleteCookie( this.cookieName );
        return value;
    },
    
    // UTILITY
    
    _getIframeElement: function ( /*Document?*/doc ) {
        // summary: Retrieves the iframe HTML element from the HTML document specified. If no HTML document is specified,
        //      the global HTML document is used.
        // doc: OPTIONAL -- specify the HTML document in which to look up the IFRAME object.
        // returns: the iframe HTML element
        if ( this.debug.enabled ) { this.debug.log( this.className,  "_getIframeElement( " + [ doc ] + ")" ); }
        if ( !doc ) { doc = document; }
        return doc.getElementById( this.iframeID );     // the IFRAME HTML element
    },
    addPalette: function ( /*PaletteContext*/context ) {
        if ( this.debug.enabled ){ this.debug.log( this.className, "addPalette( " + [ context ] + ")" ); }
        this.paletteContextArray.push( context );
    },
    
    getURL: function ( /*int*/value ) {
        if ( this.debug.enabled ) { this.debug.log( this.className, "getURL( " + [ value ] + ")" ); }
        var url = this.paletteContextArray[ value ].url;
        if ( document.isCSA && this.urlFactory != null ) {
            url = this.urlFactory( this.paletteContextArray[ value ] );
        }
        return url;
    }
    
    
}

var wptheme_DarkTransparentLoadingDecorator = function ( /*String*/cssClassName, /*String*/imageURL, /*String*/imageText ) {
    // summary: Displays a partially opaque overlay with a centered image and text to partially obscure and prevent
    //      interaction with a loading portion of the HTML page.
    // cssClassName: the class name to apply to the overlaid div
    // imageURL: the url to the image to display in the center of the overlaid div
    // imageText: the text to display next to the image
    this.className = "wptheme_DarkTransparentLoadingDecorator";
    
    var elem = document.createElement( "DIV" );
    elem.className = cssClassName;
    elem.style.position = "absolute";
    
    if ( imageURL && imageURL != "" && imageText ) {
        var text = document.createElement( "DIV" );
        text.style.position = "relative";
        text.style.top = "50%";
        text.style.left = "40%";
        text.innerHTML = "<img src='"+ imageURL + "' alt='" + imageText + "' />&nbsp;" + imageText;
        elem.appendChild( text );
    }   
    
    var appended = false;
    
    this.show = function ( refNode ) { 
        if ( !appended ) {
            document.body.appendChild( elem );
            appended= true;
        }
        
        elem.style.display = 'block';
        elem.style.top = (dojo.coords( refNode, true ).y + 1) + "px";
        elem.style.left = (dojo.coords( refNode, true ).x + 1)  + "px";
      
        var size = dojo.contentBox( refNode );
        elem.style.height = (size.h - 2) + "px";
        elem.style.width = (size.w - 2) + "px";
    }

    this.hide = function () {
        elem.style.display = 'none';
    }
}

var wptheme_InlinePalettesContainer = { 
    // summary: Manages the inline palettes container.
    // description: Manages the container which holds the palettes. Made up of two parts: the first is the container.
    //      The container holds the links to select a palette, as well as, the actual iframe which displays
    //      the palette once a palette is selected. The second is the toggle element. The toggle element is
    //      the html element which actually opens and closes the container element.
    // containerStatus: indicates whether the container is open or closed (0 = closed, 1 = open)
    // openCssClassName: indicates the CSS class name which should be applied when the container is open
    // closedCssClassName: indicates the CSS class name which should be applied when the container is closed
    // containerElementID: the id of the html element which actually holds the palettes
    // toggleElementID: the id of the html element which is the toggle element
    // lastIndex: the index of the last palette that was opened
    // cookieName: the name of the cookie used to store the container's last state
    // cookieUtils: the utility object used to set and unset cookies - default is wptheme_CookieUtils
    // htmlUtils: the utility object used for adding/removing css classnames - default is wptheme_HTMLElementUtils
    // paletteManager: the object which contains the information about the palettes to display inside the container
    //      default is wptheme_InlinePalettes
    className: "wptheme_InlinePalettesContainer",
    debug: wptheme_DebugUtils,
    
    containerStatus: 0,     //0 = closed, 1 = open
    openCssClassName: "wptheme-flyoutExpanded",
    closedCssClassName: "wptheme-flyoutCollapsed",
    toggleElementID: "wptheme-flyoutToggle",
    containerElementID: "wptheme-flyout",
    lastIndex: null,
    cookieName: "portalFlyoutIsOpen",
    cookieUtils: wptheme_CookieUtils,
    htmlUtils: wptheme_HTMLElementUtils,
    paletteManager: wptheme_InlinePalettes,
    
    //Main functions.
    init: function ( /*HTMLDocument?*/doc ) {
        // summary: Initializes and sets the appropriate visibilites for the container and the
        //      palettes inside.
        // doc: OPTIONAL -- used when called from an iframe
        if ( this.debug.enabled ) { this.debug.log( this.className, "init( " + [ doc ] + ")" ); }
        
        var cookie = this.cookieUtils.getCookie( this.cookieName );
        if ( cookie && cookie != "null" ) {
            this.containerStatus = parseInt( cookie );
        }
        
        if ( this.debug.enabled ) { this.debug.log( this.className, "containerStatus is " + this.containerStatus ); }
        
        if ( this.paletteManager.paletteContextArray.length == 0 ) {
            this.disable();
        }
        else {
            if ( this.containerStatus ) {
                this.paletteManager.init();
                this._show();
            }
            else {
                this._hide();
            }
            this._makeVisible();
        }
    },
    toggle: function () {
        // summary: Toggles the container between open and closed state.
        if ( this.debug.enabled ) { this.debug.log( this.className, "toggle()" ); }
        
        if ( this.containerStatus ) {
            this.containerStatus = 0;
            this._hide();           
        }   
        else {
            this.containerStatus = 1;
            this._show();
        }
    },
    persist: function () {
        // summary: Sets the cookie with the current container status.
        if ( this.debug.enabled ) { this.debug.log( this.className, "persist()" ); }
        
        this.cookieUtils.setCookie( this.cookieName, this.containerStatus );
        if ( this.paletteManager.currentIndex == this.lastIndex ) {
            this.paletteManager.persist( this.lastIndex );
        }
    },
    unpersist: function () {
        // summary: Removes the cookie which holds the state of the flyout.
        if ( this.debug.enabled ) { this.debug.log( this.className, "unpersist()" ); }
        
        this.cookieUtils.deleteCookie( this.cookieName );
        this.lastIndex = this.paletteManager.unpersist();
    },
    _makeVisible: function () {
        // summary: Shows the toggle element AND the container element.
        //Retrieve the applicable DOM elements.
        if ( this.debug.enabled ) { this.debug.log( this.className, "_makeVisible()" ); }
        
        var toggleElement = document.getElementById( this.toggleElementID );
        toggleElement.style.visibility = 'visible';
    },
    disable: function () {
        // summary: Hides the toggle element AND the container element.
        //Retrieve the applicable DOM elements.
        if ( this.debug.enabled ) { this.debug.log( this.className, "disable()" ); }
        
        var toggleElement = document.getElementById( this.toggleElementID );
        var containerElement = document.getElementById( this.containerElementID );
        

        if (toggleElement != null) {
            toggleElement.style.display = 'none';
        }
        if (containerElement != null) {
            containerElement.style.display = 'none';
        }
    },
    _hide: function () {
        //Retrieve the applicable DOM elements.
        if ( this.debug.enabled ) { this.debug.log( this.className, "_hide()" ); }
        
        var toggleElement = document.getElementById( this.toggleElementID );
        var containerElement = document.getElementById( this.containerElementID );
        
        if ( !toggleElement || !containerElement ) {
            throw Error( "Unable to retrieve the necessary markup elements! The palettes may not function correctly.");
        }
        
        //Closing the container.
        this.htmlUtils.removeClassName( toggleElement, this.openCssClassName );
        this.htmlUtils.addClassName( toggleElement, this.closedCssClassName );
        containerElement.style.display = 'none';
        this.paletteManager.hide( document );
        
        //Persistence cleanup.
        this.unpersist();       
    },
    _show: function () {
        //Retrieve the applicable DOM elements.
        if ( this.debug.enabled ) { this.debug.log( this.className, "_show()" ); }
        
        var toggleElement = document.getElementById( this.toggleElementID );
        var containerElement = document.getElementById( this.containerElementID );
        
        if ( !toggleElement || !containerElement ) {
            throw Error( "Unable to retrieve the necessary markup elements! The palettes may not function correctly.");
        }
        
        //Opening the container.
        this.htmlUtils.removeClassName( toggleElement, this.closedCssClassName );
        this.htmlUtils.addClassName( toggleElement, this.openCssClassName );
        containerElement.style.display = 'block';
        
        this.paletteManager.showCurrent();
        
        //Persistence cleanup.
        this.persist();
    }
}
//If we aren't inside an iframe, go ahead and register the init function so it's called on page load. This is where we check for 
//a palette's persistent state and handle other startup tasks.
if ( top.location == self.location ) {
    wptheme_InlinePalettesContainer.htmlUtils.addOnload( function () { wptheme_InlinePalettesContainer.init(); } );
};
 
//Shows an IFRAME inside a lightbox which blocks access to the page.
var wptheme_IFrameLightbox = function ( /*String*/disabledBackgroundClassname, /*String*/borderBoxClassname, /*String*/closeLinkClassname, /*String*/closeString ) {
	// summary: Creates a "lightbox" effect where a partially opaque div is set to cover the entire viewable area of the browser and the content
	//		is displayed in an iframe in approximately the middle of the viewable area.
	// description: Creates a div the size of the viewable area of the browser which is styled using the given "disabledBackgroundClassname". The iframe is
	//		displayed inside another div which is approximately centered and styled according to the given "borderBoxClassname". The content of the iframe is
	//		set using the "setURL" function. The "lightbox" is closed via a text anchor link which is positioned above the top right edge of the border box. The
	//		text displayed is controlled using the "closeString" parameter and the link is styled according to the "closeLinkClassname".
	// disabledBackgroundClassname: the CSS class name to apply to the background div displayed when the lightbox is showing
	// borderBoxClassname: the CSS class name to apply to the border box in the center of the page
	// closeString: the string which will be displayed as the link to close the lightbox
	this.className = "wptheme_IFrameLightbox";
	
	//Declare this here so that any dependency error (e.g. wptheme_HTMLElementUtils not yet being defined)
	//is clear from the beginning (throws an error at construction time instead of runtime). Also, allows
	//for easy substitution of alternate implementations (as long as function names & signatures are the same).
	this._htmlUtils = wptheme_HTMLElementUtils;
	this._debugUtils = wptheme_DebugUtils;
	
	this._initialized = false;
	this.showing = false;
	
	var uniquePrefix = this._htmlUtils.getUniqueId();
	this._backgroundDivId = uniquePrefix + "_lightboxPageBackgroundDiv";
	this._borderDivId = uniquePrefix + "_lightboxBorderDiv";
	this._closeLinkId = uniquePrefix + "_lightboxCloseLink";
	this._iframeId = uniquePrefix + "_lightboxIframe";
	
	// ****************************************************************
	// * Dynamically created DOM elements. 
	// ****************************************************************

	function createDiv(idStr, className, parent ) {
		// summary: Creates a div with the given ID, class, and appends to the given parent node. The display property is set to none by default.
		var div = document.createElement( "DIV" ); 
		div.id = idStr;
		div.className = className;
		div.style.display = "none";
		parent.appendChild( div );
		return div;
	}
	var me = this;
	function createLink(idStr, className, text, parent) {
		// summary: Creates a link with the given ID, class, textContent, and appends it to the given parent node. The display property is set to none
		//		by default. The onclick is set to hide the lightbox.
		var a = document.createElement( "A" );
		a.id = idStr;
		a.className = className;
		a.href = "javascript:void(0);";
		a.onclick = function () { me.hide() };
		a.style.display = "none";
		a.appendChild( document.createTextNode( text ) );
		parent.appendChild( a );
		return a;
	}
	
	function createIFrame( idStr, parent ) {
		// summary: Creates an iframe with the given ID (also used for the name) and appends it to the given parent node.
		var iframe = document.createElement( "IFRAME" );
		iframe.name = idStr;
		iframe.id = idStr;
		//iframe.style.display = "none";
		parent.appendChild( iframe );
		return iframe;
	}
	
	// ****************************************************************
	// * Initialization.
	// ****************************************************************
	
	this._init = function () {
		this._initialized = true;
		//Create the background div.
		createDiv( this._backgroundDivId, disabledBackgroundClassname, document.body );
		//Create the border box div
		createIFrame( this._iframeId, createDiv( this._borderDivId, borderBoxClassname, document.body ));
		//Create the close link.
		createLink( this._closeLinkId, closeLinkClassname, closeString, document.body );
	}
	
	// ****************************************************************
	// * Handling the browser scrolling and resizing dynamically.
	// ****************************************************************

	//Make sure to call any existing onscroll handler.
	var oldScrollFunc = window.onscroll;
	window.onscroll = function (e) {
		if ( me.showing ) {
			me.sizeAndPositionBorderBox();
			//me.sizeBackgroundDisablingDiv();
		}
		
		if ( oldScrollFunc ) {
			if (e) {
				oldScrollFunc(e);
			}
			else  {
				oldScrollFunc();
			}
		}
	}
	
	//Make sure to call any existing onresize handler.
	var oldResizeFunc = window.onresize;
	window.onresize = function (e) {
		if ( me.showing ) {
			me.sizeAndPositionBorderBox();
			me.sizeBackgroundDisablingDiv();
		}
		
		if ( oldResizeFunc ) {
			if (e) {
				oldResizeFunc(e);
			}
			else  {
				oldResizeFunc();
			}
		}
	}

	// ****************************************************************
	// * Main functions for use in the theme.
	// ****************************************************************
	
	this.setURL = function ( /*String*/url ) {
		// summary: Sets the URL displayed by the IFRAME in the lightbox.
		// url: the url to the resource to display
		window.frames[this._iframeId].location = url;
	}
	
	
	
	this.show = function ( /*String?*/url ) {
		// summary: Shows the lightbox above the disabled background div. 
		// url: OPTIONAL -- the url to display in the iframe in the center of the screen 
		if ( !this._initialized ) {
			this._init();
		}
		
		this.showing = true;
		
		this.disableBackground();
		this.showBorderBox();
		if ( url ) { 
			this.setURL( url );
		}		
	}
	
	this.hide = function() {
		// summary: Hides the lightbox and the disabled background div.
		if ( !this._initialized ) {
			this._init();
		}
		
		this.showing = false;
		
		this.enableBackground();
		this.hideBorderBox();
	}
	
	// ****************************************************************
	// * Content border box 
	// ****************************************************************
	this.showBorderBox = function () {
		// summary: Shows and positions the border box which contains the IFRAME.
		var div = document.getElementById( this._borderDivId );
		div.style.display = "block";
		var link = document.getElementById( this._closeLinkId );
		link.style.display = "block";
		
		this.sizeAndPositionBorderBox();
	}
	
	this.sizeAndPositionBorderBox = function () {
		// summary: Sizes and positions the border box which contains the IFRAME.
		var div = document.getElementById( this._borderDivId );
		this._htmlUtils.sizeRelativeToViewableArea( div, 0.60, 0.75 );
		this._htmlUtils.positionRelativeToViewableArea( div, 0.20, 0.12 );
		var link = document.getElementById( this._closeLinkId );
		this._htmlUtils.positionOutsideElementTopRight( link, div );
	}
	
	this.hideBorderBox = function () {
		// summary: hides the border box and IFRAME.
		document.getElementById( this._borderDivId ).style.display = "none";
		document.getElementById( this._closeLinkId ).style.display = "none";
	}
	
	// **************************************************************** 
	// * Transparent background controls
	// ****************************************************************
	this.disableBackground = function () {
		// summary: Disables the background by laying a transparent div over top of the document body.
		var div = document.getElementById( this._backgroundDivId );
		div.style.display = "block";
		this.sizeBackgroundDisablingDiv();
		this._htmlUtils.hideElementsByTagName( "select" );
	}
	
	this.sizeBackgroundDisablingDiv = function () {
		// summary: Sizes the transparent div appropriately.
		var div = document.getElementById( this._backgroundDivId );
		//dynamically size the div to the inner browser window
		this._htmlUtils.sizeToEntireArea( div );
	}
	
	this.enableBackground=function () {
		// summary: Enables the background by hiding the overlaid div.
		this._htmlUtils.showElementsByTagName( "select" );
		document.getElementById( this._backgroundDivId ).style.display = "none";
	}
};
 
/*
	Copyright (c) 2004-2009, The Dojo Foundation All Rights Reserved.
	Available via Academic Free License >= 2.1 OR the modified BSD license.
	see: http://dojotoolkit.org/license for details
*/

/*
	This is a compiled version of Dojo, built for deployment and not for
	development. To get an editable version, please visit:

		http://dojotoolkit.org

	for documentation and information on getting the source.
*/

(function(){var _1=null;if((_1||(typeof djConfig!="undefined"&&djConfig.scopeMap))&&(typeof window!="undefined")){var _2="",_3="",_4="",_5={},_6={};_1=_1||djConfig.scopeMap;for(var i=0;i<_1.length;i++){var _8=_1[i];_2+="var "+_8[0]+" = {}; "+_8[1]+" = "+_8[0]+";"+_8[1]+"._scopeName = '"+_8[1]+"';";_3+=(i==0?"":",")+_8[0];_4+=(i==0?"":",")+_8[1];_5[_8[0]]=_8[1];_6[_8[1]]=_8[0];}eval(_2+"dojo._scopeArgs = ["+_4+"];");dojo._scopePrefixArgs=_3;dojo._scopePrefix="(function("+_3+"){";dojo._scopeSuffix="})("+_4+")";dojo._scopeMap=_5;dojo._scopeMapRev=_6;}(function(){if(typeof this["loadFirebugConsole"]=="function"){this["loadFirebugConsole"]();}else{if(this["navigator"]){if(/3[\.0-9]+.*Safari/i.test(navigator.appVersion)&&this["console"]){this.console={_c:this.console,log:function(s){this._c.log(s);},info:function(s){this._c.info(s);},error:function(s){this._c.error(s);},warn:function(s){this._c.warn(s);}};}}this.console=this.console||{};var cn=["assert","count","debug","dir","dirxml","error","group","groupEnd","info","profile","profileEnd","time","timeEnd","trace","warn","log"];var i=0,tn;while((tn=cn[i++])){if(!console[tn]){(function(){var tcn=tn+"";console[tcn]=("log" in console)?function(){var a=Array.apply({},arguments);a.unshift(tcn+":");console["log"](a.join(" "));}:function(){};})();}}}if(typeof dojo=="undefined"){this.dojo={_scopeName:"dojo",_scopePrefix:"",_scopePrefixArgs:"",_scopeSuffix:"",_scopeMap:{},_scopeMapRev:{}};}var d=dojo;if(typeof dijit=="undefined"){this.dijit={_scopeName:"dijit"};}if(typeof dojox=="undefined"){this.dojox={_scopeName:"dojox"};}if(!d._scopeArgs){d._scopeArgs=[dojo,dijit,dojox];}d.global=this;d.config={isDebug:false,debugAtAllCosts:false};if(typeof djConfig!="undefined"){for(var opt in djConfig){d.config[opt]=djConfig[opt];}}dojo.locale=d.config.locale;var rev="$Rev: 18832 $".match(/\d+/);dojo.version={major:1,minor:3,patch:2,flag:"_IBM",revision:rev?+rev[0]:NaN,toString:function(){with(d.version){return major+"."+minor+"."+patch+flag+" ("+revision+")";}}};if(typeof OpenAjax!="undefined"){OpenAjax.hub.registerLibrary(dojo._scopeName,"http://dojotoolkit.org",d.version.toString());}var _15={};dojo._mixin=function(obj,_17){for(var x in _17){if(_15[x]===undefined||_15[x]!=_17[x]){obj[x]=_17[x];}}if(d.isIE&&_17){var p=_17.toString;if(typeof p=="function"&&p!=obj.toString&&p!=_15.toString&&p!="\nfunction toString() {\n    [native code]\n}\n"){obj.toString=_17.toString;}}return obj;};dojo.mixin=function(obj,_1b){if(!obj){obj={};}for(var i=1,l=arguments.length;i<l;i++){d._mixin(obj,arguments[i]);}return obj;};dojo._getProp=function(_1e,_1f,_20){var obj=_20||d.global;for(var i=0,p;obj&&(p=_1e[i]);i++){if(i==0&&this._scopeMap[p]){p=this._scopeMap[p];}obj=(p in obj?obj[p]:(_1f?obj[p]={}:undefined));}return obj;};dojo.setObject=function(_24,_25,_26){var _27=_24.split("."),p=_27.pop(),obj=d._getProp(_27,true,_26);return obj&&p?(obj[p]=_25):undefined;};dojo.getObject=function(_2a,_2b,_2c){return d._getProp(_2a.split("."),_2b,_2c);};dojo.exists=function(_2d,obj){return !!d.getObject(_2d,false,obj);};dojo["eval"]=function(_2f){return d.global.eval?d.global.eval(_2f):eval(_2f);};d.deprecated=d.experimental=function(){};})();(function(){var d=dojo;d.mixin(d,{_loadedModules:{},_inFlightCount:0,_hasResource:{},_modulePrefixes:{dojo:{name:"dojo",value:"."},doh:{name:"doh",value:"../util/doh"},tests:{name:"tests",value:"tests"}},_moduleHasPrefix:function(_31){var mp=this._modulePrefixes;return !!(mp[_31]&&mp[_31].value);},_getModulePrefix:function(_33){var mp=this._modulePrefixes;if(this._moduleHasPrefix(_33)){return mp[_33].value;}return _33;},_loadedUrls:[],_postLoad:false,_loaders:[],_unloaders:[],_loadNotifying:false});dojo._loadPath=function(_35,_36,cb){var uri=((_35.charAt(0)=="/"||_35.match(/^\w+:/))?"":this.baseUrl)+_35;try{return !_36?this._loadUri(uri,cb):this._loadUriAndCheck(uri,_36,cb);}catch(e){console.error(e);return false;}};dojo._loadUri=function(uri,cb){if(d._loadedUrls[uri]){return true;}d._inFlightCount++;var _3b=d._getText(uri,true);if(_3b){d._loadedUrls[uri]=true;d._loadedUrls.push(uri);if(cb){_3b="("+_3b+")";}else{_3b=d._scopePrefix+_3b+d._scopeSuffix;}if(d.isMoz){_3b+="\r\n//@ sourceURL="+uri;}var _3c=d["eval"](_3b);if(cb){cb(_3c);}}if(--d._inFlightCount==0&&d._postLoad&&d._loaders.length){setTimeout(function(){if(d._inFlightCount==0){d._callLoaded();}},0);}return !!_3b;};dojo._loadUriAndCheck=function(uri,_3e,cb){var ok=false;try{ok=this._loadUri(uri,cb);}catch(e){console.error("failed loading "+uri+" with error: "+e);}return !!(ok&&this._loadedModules[_3e]);};dojo.loaded=function(){this._loadNotifying=true;this._postLoad=true;var mll=d._loaders;this._loaders=[];for(var x=0;x<mll.length;x++){mll[x]();}this._loadNotifying=false;if(d._postLoad&&d._inFlightCount==0&&mll.length){d._callLoaded();}};dojo.unloaded=function(){var mll=d._unloaders;while(mll.length){(mll.pop())();}};d._onto=function(arr,obj,fn){if(!fn){arr.push(obj);}else{if(fn){var _47=(typeof fn=="string")?obj[fn]:fn;arr.push(function(){_47.call(obj);});}}};dojo.addOnLoad=function(obj,_49){d._onto(d._loaders,obj,_49);if(d._postLoad&&d._inFlightCount==0&&!d._loadNotifying){d._callLoaded();}};var dca=d.config.addOnLoad;if(dca){d.addOnLoad[(dca instanceof Array?"apply":"call")](d,dca);}dojo._modulesLoaded=function(){if(d._postLoad){return;}if(d._inFlightCount>0){console.warn("files still in flight!");return;}d._callLoaded();};dojo._callLoaded=function(){if(typeof setTimeout=="object"||(dojo.config.useXDomain&&d.isOpera)){if(dojo.isAIR){setTimeout(function(){dojo.loaded();},0);}else{setTimeout(dojo._scopeName+".loaded();",0);}}else{d.loaded();}};dojo._getModuleSymbols=function(_4b){var _4c=_4b.split(".");for(var i=_4c.length;i>0;i--){var _4e=_4c.slice(0,i).join(".");if((i==1)&&!this._moduleHasPrefix(_4e)){_4c[0]="../"+_4c[0];}else{var _4f=this._getModulePrefix(_4e);if(_4f!=_4e){_4c.splice(0,i,_4f);break;}}}return _4c;};dojo._global_omit_module_check=false;dojo.loadInit=function(_50){_50();};dojo._loadModule=dojo.require=function(_51,_52){_52=this._global_omit_module_check||_52;var _53=this._loadedModules[_51];if(_53){return _53;}var _54=this._getModuleSymbols(_51).join("/")+".js";var _55=(!_52)?_51:null;var ok=this._loadPath(_54,_55);if(!ok&&!_52){throw new Error("Could not load '"+_51+"'; last tried '"+_54+"'");}if(!_52&&!this._isXDomain){_53=this._loadedModules[_51];if(!_53){throw new Error("symbol '"+_51+"' is not defined after loading '"+_54+"'");}}return _53;};dojo.provide=function(_57){_57=_57+"";return (d._loadedModules[_57]=d.getObject(_57,true));};dojo.platformRequire=function(_58){var _59=_58.common||[];var _5a=_59.concat(_58[d._name]||_58["default"]||[]);for(var x=0;x<_5a.length;x++){var _5c=_5a[x];if(_5c.constructor==Array){d._loadModule.apply(d,_5c);}else{d._loadModule(_5c);}}};dojo.requireIf=function(_5d,_5e){if(_5d===true){var _5f=[];for(var i=1;i<arguments.length;i++){_5f.push(arguments[i]);}d.require.apply(d,_5f);}};dojo.requireAfterIf=d.requireIf;dojo.registerModulePath=function(_61,_62){d._modulePrefixes[_61]={name:_61,value:_62};};dojo.requireLocalization=function(_63,_64,_65,_66){d.require("dojo.i18n");d.i18n._requireLocalization.apply(d.hostenv,arguments);};var ore=new RegExp("^(([^:/?#]+):)?(//([^/?#]*))?([^?#]*)(\\?([^#]*))?(#(.*))?$");var ire=new RegExp("^((([^\\[:]+):)?([^@]+)@)?(\\[([^\\]]+)\\]|([^\\[:]*))(:([0-9]+))?$");dojo._Url=function(){var n=null;var _a=arguments;var uri=[_a[0]];for(var i=1;i<_a.length;i++){if(!_a[i]){continue;}var _6d=new d._Url(_a[i]+"");var _6e=new d._Url(uri[0]+"");if(_6d.path==""&&!_6d.scheme&&!_6d.authority&&!_6d.query){if(_6d.fragment!=n){_6e.fragment=_6d.fragment;}_6d=_6e;}else{if(!_6d.scheme){_6d.scheme=_6e.scheme;if(!_6d.authority){_6d.authority=_6e.authority;if(_6d.path.charAt(0)!="/"){var _6f=_6e.path.substring(0,_6e.path.lastIndexOf("/")+1)+_6d.path;var _70=_6f.split("/");for(var j=0;j<_70.length;j++){if(_70[j]=="."){if(j==_70.length-1){_70[j]="";}else{_70.splice(j,1);j--;}}else{if(j>0&&!(j==1&&_70[0]=="")&&_70[j]==".."&&_70[j-1]!=".."){if(j==(_70.length-1)){_70.splice(j,1);_70[j-1]="";}else{_70.splice(j-1,2);j-=2;}}}}_6d.path=_70.join("/");}}}}uri=[];if(_6d.scheme){uri.push(_6d.scheme,":");}if(_6d.authority){uri.push("//",_6d.authority);}uri.push(_6d.path);if(_6d.query){uri.push("?",_6d.query);}if(_6d.fragment){uri.push("#",_6d.fragment);}}this.uri=uri.join("");var r=this.uri.match(ore);this.scheme=r[2]||(r[1]?"":n);this.authority=r[4]||(r[3]?"":n);this.path=r[5];this.query=r[7]||(r[6]?"":n);this.fragment=r[9]||(r[8]?"":n);if(this.authority!=n){r=this.authority.match(ire);this.user=r[3]||n;this.password=r[4]||n;this.host=r[6]||r[7];this.port=r[9]||n;}};dojo._Url.prototype.toString=function(){return this.uri;};dojo.moduleUrl=function(_73,url){var loc=d._getModuleSymbols(_73).join("/");if(!loc){return null;}if(loc.lastIndexOf("/")!=loc.length-1){loc+="/";}var _76=loc.indexOf(":");if(loc.charAt(0)!="/"&&(_76==-1||_76>loc.indexOf("/"))){loc=d.baseUrl+loc;}return new d._Url(loc,url);};})();if(typeof window!="undefined"){dojo.isBrowser=true;dojo._name="browser";(function(){var d=dojo;if(document&&document.getElementsByTagName){var _78=document.getElementsByTagName("script");var _79=/dojo(\.xd)?\.js(\W|$)/i;for(var i=0;i<_78.length;i++){var src=_78[i].getAttribute("src");if(!src){continue;}var m=src.match(_79);if(m){if(!d.config.baseUrl){d.config.baseUrl=src.substring(0,m.index);}var cfg=_78[i].getAttribute("djConfig");if(cfg){var _7e=eval("({ "+cfg+" })");for(var x in _7e){dojo.config[x]=_7e[x];}}break;}}}d.baseUrl=d.config.baseUrl;var n=navigator;var dua=n.userAgent,dav=n.appVersion,tv=parseFloat(dav);if(dua.indexOf("Opera")>=0){d.isOpera=tv;}if(dua.indexOf("AdobeAIR")>=0){d.isAIR=1;}d.isKhtml=(dav.indexOf("Konqueror")>=0)?tv:0;d.isWebKit=parseFloat(dua.split("WebKit/")[1])||undefined;d.isChrome=parseFloat(dua.split("Chrome/")[1])||undefined;var _84=Math.max(dav.indexOf("WebKit"),dav.indexOf("Safari"),0);if(_84&&!dojo.isChrome){d.isSafari=parseFloat(dav.split("Version/")[1]);if(!d.isSafari||parseFloat(dav.substr(_84+7))<=419.3){d.isSafari=2;}}if(dua.indexOf("Gecko")>=0&&!d.isKhtml&&!d.isWebKit){d.isMozilla=d.isMoz=tv;}if(d.isMoz){d.isFF=parseFloat(dua.split("Firefox/")[1]||dua.split("Minefield/")[1]||dua.split("Shiretoko/")[1])||undefined;}if(document.all&&!d.isOpera){d.isIE=parseFloat(dav.split("MSIE ")[1])||undefined;if(d.isIE>=8&&document.documentMode!=5){d.isIE=document.documentMode;}}if(dojo.isIE&&window.location.protocol==="file:"){dojo.config.ieForceActiveXXhr=true;}var cm=document.compatMode;d.isQuirks=cm=="BackCompat"||cm=="QuirksMode"||d.isIE<6;d.locale=dojo.config.locale||(d.isIE?n.userLanguage:n.language).toLowerCase();d._XMLHTTP_PROGIDS=["Msxml2.XMLHTTP","Microsoft.XMLHTTP","Msxml2.XMLHTTP.4.0"];d._xhrObj=function(){var _86,_87;if(!dojo.isIE||!dojo.config.ieForceActiveXXhr){try{_86=new XMLHttpRequest();}catch(e){}}if(!_86){for(var i=0;i<3;++i){var _89=d._XMLHTTP_PROGIDS[i];try{_86=new ActiveXObject(_89);}catch(e){_87=e;}if(_86){d._XMLHTTP_PROGIDS=[_89];break;}}}if(!_86){throw new Error("XMLHTTP not available: "+_87);}return _86;};d._isDocumentOk=function(_8a){var _8b=_8a.status||0;return (_8b>=200&&_8b<300)||_8b==304||_8b==1223||(!_8b&&(location.protocol=="file:"||location.protocol=="chrome:"));};var _8c=window.location+"";var _8d=document.getElementsByTagName("base");var _8e=(_8d&&_8d.length>0);d._getText=function(uri,_90){var _91=this._xhrObj();if(!_8e&&dojo._Url){uri=(new dojo._Url(_8c,uri)).toString();}if(d.config.cacheBust){uri+="";uri+=(uri.indexOf("?")==-1?"?":"&")+String(d.config.cacheBust).replace(/\W+/g,"");}_91.open("GET",uri,false);try{_91.send(null);if(!d._isDocumentOk(_91)){var err=Error("Unable to load "+uri+" status:"+_91.status);err.status=_91.status;err.responseText=_91.responseText;throw err;}}catch(e){if(_90){return null;}throw e;}return _91.responseText;};var _w=window;var _94=function(_95,fp){var _97=_w[_95]||function(){};_w[_95]=function(){fp.apply(_w,arguments);_97.apply(_w,arguments);};};d._windowUnloaders=[];d.windowUnloaded=function(){var mll=d._windowUnloaders;while(mll.length){(mll.pop())();}};var _99=0;d.addOnWindowUnload=function(obj,_9b){d._onto(d._windowUnloaders,obj,_9b);if(!_99){_99=1;_94("onunload",d.windowUnloaded);}};var _9c=0;d.addOnUnload=function(obj,_9e){d._onto(d._unloaders,obj,_9e);if(!_9c){_9c=1;_94("onbeforeunload",dojo.unloaded);}};})();dojo._initFired=false;dojo._loadInit=function(e){dojo._initFired=true;var _a0=e&&e.type?e.type.toLowerCase():"load";if(arguments.callee.initialized||(_a0!="domcontentloaded"&&_a0!="load")){return;}arguments.callee.initialized=true;if("_khtmlTimer" in dojo){clearInterval(dojo._khtmlTimer);delete dojo._khtmlTimer;}if(dojo._inFlightCount==0){dojo._modulesLoaded();}};if(!dojo.config.afterOnLoad){if(document.addEventListener){if(dojo.isWebKit>525||dojo.isOpera||dojo.isFF>=3||(dojo.isMoz&&dojo.config.enableMozDomContentLoaded===true)){document.addEventListener("DOMContentLoaded",dojo._loadInit,null);}window.addEventListener("load",dojo._loadInit,null);}if(dojo.isAIR){window.addEventListener("load",dojo._loadInit,null);}else{if((dojo.isWebKit<525)||dojo.isKhtml){dojo._khtmlTimer=setInterval(function(){if(/loaded|complete/.test(document.readyState)){dojo._loadInit();}},10);}}}if(dojo.isIE){if(!dojo.config.afterOnLoad){document.write("<scr"+"ipt defer src=\"//:\" "+"onreadystatechange=\"if(this.readyState=='complete'){"+dojo._scopeName+"._loadInit();}\">"+"</scr"+"ipt>");}try{document.namespaces.add("v","urn:schemas-microsoft-com:vml");document.createStyleSheet().addRule("v\\:*","behavior:url(#default#VML);  display:inline-block");}catch(e){}}}(function(){var mp=dojo.config["modulePaths"];if(mp){for(var _a2 in mp){dojo.registerModulePath(_a2,mp[_a2]);}}})();if(dojo.config.isDebug){dojo.require("dojo._firebug.firebug");}if(dojo.config.debugAtAllCosts){dojo.config.useXDomain=true;dojo.require("dojo._base._loader.loader_xd");dojo.require("dojo._base._loader.loader_debug");}if(!dojo._hasResource["dojo._base.lang"]){dojo._hasResource["dojo._base.lang"]=true;dojo.provide("dojo._base.lang");dojo.isString=function(it){return !!arguments.length&&it!=null&&(typeof it=="string"||it instanceof String);};dojo.isArray=function(it){return it&&(it instanceof Array||typeof it=="array");};dojo.isFunction=(function(){var _a5=function(it){var t=typeof it;return it&&(t=="function"||it instanceof Function);};return dojo.isSafari?function(it){if(typeof it=="function"&&it=="[object NodeList]"){return false;}return _a5(it);}:_a5;})();dojo.isObject=function(it){return it!==undefined&&(it===null||typeof it=="object"||dojo.isArray(it)||dojo.isFunction(it));};dojo.isArrayLike=function(it){var d=dojo;return it&&it!==undefined&&!d.isString(it)&&!d.isFunction(it)&&!(it.tagName&&it.tagName.toLowerCase()=="form")&&(d.isArray(it)||isFinite(it.length));};dojo.isAlien=function(it){return it&&!dojo.isFunction(it)&&/\{\s*\[native code\]\s*\}/.test(String(it));};dojo.extend=function(_ad,_ae){for(var i=1,l=arguments.length;i<l;i++){dojo._mixin(_ad.prototype,arguments[i]);}return _ad;};dojo._hitchArgs=function(_b1,_b2){var pre=dojo._toArray(arguments,2);var _b4=dojo.isString(_b2);return function(){var _b5=dojo._toArray(arguments);var f=_b4?(_b1||dojo.global)[_b2]:_b2;return f&&f.apply(_b1||this,pre.concat(_b5));};};dojo.hitch=function(_b7,_b8){if(arguments.length>2){return dojo._hitchArgs.apply(dojo,arguments);}if(!_b8){_b8=_b7;_b7=null;}if(dojo.isString(_b8)){_b7=_b7||dojo.global;if(!_b7[_b8]){throw (["dojo.hitch: scope[\"",_b8,"\"] is null (scope=\"",_b7,"\")"].join(""));}return function(){return _b7[_b8].apply(_b7,arguments||[]);};}return !_b7?_b8:function(){return _b8.apply(_b7,arguments||[]);};};dojo.delegate=dojo._delegate=(function(){function TMP(){};return function(obj,_ba){TMP.prototype=obj;var tmp=new TMP();if(_ba){dojo._mixin(tmp,_ba);}return tmp;};})();(function(){var _bc=function(obj,_be,_bf){return (_bf||[]).concat(Array.prototype.slice.call(obj,_be||0));};var _c0=function(obj,_c2,_c3){var arr=_c3||[];for(var x=_c2||0;x<obj.length;x++){arr.push(obj[x]);}return arr;};dojo._toArray=dojo.isIE?function(obj){return ((obj.item)?_c0:_bc).apply(this,arguments);}:_bc;})();dojo.partial=function(_c7){var arr=[null];return dojo.hitch.apply(dojo,arr.concat(dojo._toArray(arguments)));};dojo.clone=function(o){if(!o){return o;}if(dojo.isArray(o)){var r=[];for(var i=0;i<o.length;++i){r.push(dojo.clone(o[i]));}return r;}if(!dojo.isObject(o)){return o;}if(o.nodeType&&o.cloneNode){return o.cloneNode(true);}if(o instanceof Date){return new Date(o.getTime());}r=new o.constructor();for(i in o){if(!(i in r)||r[i]!=o[i]){r[i]=dojo.clone(o[i]);}}return r;};dojo.trim=String.prototype.trim?function(str){return str.trim();}:function(str){return str.replace(/^\s\s*/,"").replace(/\s\s*$/,"");};}if(!dojo._hasResource["dojo._base.declare"]){dojo._hasResource["dojo._base.declare"]=true;dojo.provide("dojo._base.declare");dojo.declare=function(_ce,_cf,_d0){var dd=arguments.callee,_d2;if(dojo.isArray(_cf)){_d2=_cf;_cf=_d2.shift();}if(_d2){dojo.forEach(_d2,function(m,i){if(!m){throw (_ce+": mixin #"+i+" is null");}_cf=dd._delegate(_cf,m);});}var _d5=dd._delegate(_cf);_d0=_d0||{};_d5.extend(_d0);dojo.extend(_d5,{declaredClass:_ce,_constructor:_d0.constructor});_d5.prototype.constructor=_d5;return dojo.setObject(_ce,_d5);};dojo.mixin(dojo.declare,{_delegate:function(_d6,_d7){var bp=(_d6||0).prototype,mp=(_d7||0).prototype,dd=dojo.declare;var _db=dd._makeCtor();dojo.mixin(_db,{superclass:bp,mixin:mp,extend:dd._extend});if(_d6){_db.prototype=dojo._delegate(bp);}dojo.extend(_db,dd._core,mp||0,{_constructor:null,preamble:null});_db.prototype.constructor=_db;_db.prototype.declaredClass=(bp||0).declaredClass+"_"+(mp||0).declaredClass;return _db;},_extend:function(_dc){var i,fn;for(i in _dc){if(dojo.isFunction(fn=_dc[i])&&!0[i]){fn.nom=i;fn.ctor=this;}}dojo.extend(this,_dc);},_makeCtor:function(){return function(){this._construct(arguments);};},_core:{_construct:function(_df){var c=_df.callee,s=c.superclass,ct=s&&s.constructor,m=c.mixin,mct=m&&m.constructor,a=_df,ii,fn;if(a[0]){if(((fn=a[0].preamble))){a=fn.apply(this,a)||a;}}if((fn=c.prototype.preamble)){a=fn.apply(this,a)||a;}if(ct&&ct.apply){ct.apply(this,a);}if(mct&&mct.apply){mct.apply(this,a);}if((ii=c.prototype._constructor)){ii.apply(this,_df);}if(this.constructor.prototype==c.prototype&&(ct=this.postscript)){ct.apply(this,_df);}},_findMixin:function(_e8){var c=this.constructor,p,m;while(c){p=c.superclass;m=c.mixin;if(m==_e8||(m instanceof _e8.constructor)){return p;}if(m&&m._findMixin&&(m=m._findMixin(_e8))){return m;}c=p&&p.constructor;}},_findMethod:function(_ec,_ed,_ee,has){var p=_ee,c,m,f;do{c=p.constructor;m=c.mixin;if(m&&(m=this._findMethod(_ec,_ed,m,has))){return m;}if((f=p[_ec])&&(has==(f==_ed))){return p;}p=c.superclass;}while(p);return !has&&(p=this._findMixin(_ee))&&this._findMethod(_ec,_ed,p,has);},inherited:function(_f4,_f5,_f6){var a=arguments;if(!dojo.isString(a[0])){_f6=_f5;_f5=_f4;_f4=_f5.callee.nom;}a=_f6||_f5;var c=_f5.callee,p=this.constructor.prototype,fn,mp;if(this[_f4]!=c||p[_f4]==c){mp=(c.ctor||0).superclass||this._findMethod(_f4,c,p,true);if(!mp){throw (this.declaredClass+": inherited method \""+_f4+"\" mismatch");}p=this._findMethod(_f4,c,mp,false);}fn=p&&p[_f4];if(!fn){throw (mp.declaredClass+": inherited method \""+_f4+"\" not found");}return fn.apply(this,a);}}});}if(!dojo._hasResource["dojo._base.connect"]){dojo._hasResource["dojo._base.connect"]=true;dojo.provide("dojo._base.connect");dojo._listener={getDispatcher:function(){return function(){var ap=Array.prototype,c=arguments.callee,ls=c._listeners,t=c.target;var r=t&&t.apply(this,arguments);var lls;lls=[].concat(ls);for(var i in lls){if(!(i in ap)){lls[i].apply(this,arguments);}}return r;};},add:function(_103,_104,_105){_103=_103||dojo.global;var f=_103[_104];if(!f||!f._listeners){var d=dojo._listener.getDispatcher();d.target=f;d._listeners=[];f=_103[_104]=d;}return f._listeners.push(_105);},remove:function(_108,_109,_10a){var f=(_108||dojo.global)[_109];if(f&&f._listeners&&_10a--){delete f._listeners[_10a];}}};dojo.connect=function(obj,_10d,_10e,_10f,_110){var a=arguments,args=[],i=0;args.push(dojo.isString(a[0])?null:a[i++],a[i++]);var a1=a[i+1];args.push(dojo.isString(a1)||dojo.isFunction(a1)?a[i++]:null,a[i++]);for(var l=a.length;i<l;i++){args.push(a[i]);}return dojo._connect.apply(this,args);};dojo._connect=function(obj,_116,_117,_118){var l=dojo._listener,h=l.add(obj,_116,dojo.hitch(_117,_118));return [obj,_116,h,l];};dojo.disconnect=function(_11b){if(_11b&&_11b[0]!==undefined){dojo._disconnect.apply(this,_11b);delete _11b[0];}};dojo._disconnect=function(obj,_11d,_11e,_11f){_11f.remove(obj,_11d,_11e);};dojo._topics={};dojo.subscribe=function(_120,_121,_122){return [_120,dojo._listener.add(dojo._topics,_120,dojo.hitch(_121,_122))];};dojo.unsubscribe=function(_123){if(_123){dojo._listener.remove(dojo._topics,_123[0],_123[1]);}};dojo.publish=function(_124,args){var f=dojo._topics[_124];if(f){f.apply(this,args||[]);}};dojo.connectPublisher=function(_127,obj,_129){var pf=function(){dojo.publish(_127,arguments);};return (_129)?dojo.connect(obj,_129,pf):dojo.connect(obj,pf);};}if(!dojo._hasResource["dojo._base.Deferred"]){dojo._hasResource["dojo._base.Deferred"]=true;dojo.provide("dojo._base.Deferred");dojo.Deferred=function(_12b){this.chain=[];this.id=this._nextId();this.fired=-1;this.paused=0;this.results=[null,null];this.canceller=_12b;this.silentlyCancelled=false;};dojo.extend(dojo.Deferred,{_nextId:(function(){var n=1;return function(){return n++;};})(),cancel:function(){var err;if(this.fired==-1){if(this.canceller){err=this.canceller(this);}else{this.silentlyCancelled=true;}if(this.fired==-1){if(!(err instanceof Error)){var res=err;var msg="Deferred Cancelled";if(err&&err.toString){msg+=": "+err.toString();}err=new Error(msg);err.dojoType="cancel";err.cancelResult=res;}this.errback(err);}}else{if((this.fired==0)&&(this.results[0] instanceof dojo.Deferred)){this.results[0].cancel();}}},_resback:function(res){this.fired=((res instanceof Error)?1:0);this.results[this.fired]=res;this._fire();},_check:function(){if(this.fired!=-1){if(!this.silentlyCancelled){throw new Error("already called!");}this.silentlyCancelled=false;return;}},callback:function(res){this._check();this._resback(res);},errback:function(res){this._check();if(!(res instanceof Error)){res=new Error(res);}this._resback(res);},addBoth:function(cb,cbfn){var _135=dojo.hitch.apply(dojo,arguments);return this.addCallbacks(_135,_135);},addCallback:function(cb,cbfn){return this.addCallbacks(dojo.hitch.apply(dojo,arguments));},addErrback:function(cb,cbfn){return this.addCallbacks(null,dojo.hitch.apply(dojo,arguments));},addCallbacks:function(cb,eb){this.chain.push([cb,eb]);if(this.fired>=0){this._fire();}return this;},_fire:function(){var _13c=this.chain;var _13d=this.fired;var res=this.results[_13d];var self=this;var cb=null;while((_13c.length>0)&&(this.paused==0)){var f=_13c.shift()[_13d];if(!f){continue;}var func=function(){var ret=f(res);if(typeof ret!="undefined"){res=ret;}_13d=((res instanceof Error)?1:0);if(res instanceof dojo.Deferred){cb=function(res){self._resback(res);self.paused--;if((self.paused==0)&&(self.fired>=0)){self._fire();}};this.paused++;}};if(dojo.config.debugAtAllCosts){func.call(this);}else{try{func.call(this);}catch(err){_13d=1;res=err;}}}this.fired=_13d;this.results[_13d]=res;if((cb)&&(this.paused)){res.addBoth(cb);}}});}if(!dojo._hasResource["dojo._base.json"]){dojo._hasResource["dojo._base.json"]=true;dojo.provide("dojo._base.json");dojo.fromJson=function(json){return eval("("+json+")");};dojo._escapeString=function(str){return ("\""+str.replace(/(["\\])/g,"\\$1")+"\"").replace(/[\f]/g,"\\f").replace(/[\b]/g,"\\b").replace(/[\n]/g,"\\n").replace(/[\t]/g,"\\t").replace(/[\r]/g,"\\r");};dojo.toJsonIndentStr="\t";dojo.toJson=function(it,_148,_149){if(it===undefined){return "undefined";}var _14a=typeof it;if(_14a=="number"||_14a=="boolean"){return it+"";}if(it===null){return "null";}if(dojo.isString(it)){return dojo._escapeString(it);}var _14b=arguments.callee;var _14c;_149=_149||"";var _14d=_148?_149+dojo.toJsonIndentStr:"";var tf=it.__json__||it.json;if(dojo.isFunction(tf)){_14c=tf.call(it);if(it!==_14c){return _14b(_14c,_148,_14d);}}if(it.nodeType&&it.cloneNode){throw new Error("Can't serialize DOM nodes");}var sep=_148?" ":"";var _150=_148?"\n":"";if(dojo.isArray(it)){var res=dojo.map(it,function(obj){var val=_14b(obj,_148,_14d);if(typeof val!="string"){val="undefined";}return _150+_14d+val;});return "["+res.join(","+sep)+_150+_149+"]";}if(_14a=="function"){return null;}var _154=[],key;for(key in it){var _156,val;if(typeof key=="number"){_156="\""+key+"\"";}else{if(typeof key=="string"){_156=dojo._escapeString(key);}else{continue;}}val=_14b(it[key],_148,_14d);if(typeof val!="string"){continue;}_154.push(_150+_14d+_156+":"+sep+val);}return "{"+_154.join(","+sep)+_150+_149+"}";};}if(!dojo._hasResource["dojo._base.array"]){dojo._hasResource["dojo._base.array"]=true;dojo.provide("dojo._base.array");(function(){var _158=function(arr,obj,cb){return [dojo.isString(arr)?arr.split(""):arr,obj||dojo.global,dojo.isString(cb)?new Function("item","index","array",cb):cb];};dojo.mixin(dojo,{indexOf:function(_15c,_15d,_15e,_15f){var step=1,end=_15c.length||0,i=0;if(_15f){i=end-1;step=end=-1;}if(_15e!=undefined){i=_15e;}if((_15f&&i>end)||i<end){for(;i!=end;i+=step){if(_15c[i]==_15d){return i;}}}return -1;},lastIndexOf:function(_162,_163,_164){return dojo.indexOf(_162,_163,_164,true);},forEach:function(arr,_166,_167){if(!arr||!arr.length){return;}var _p=_158(arr,_167,_166);arr=_p[0];for(var i=0,l=arr.length;i<l;++i){_p[2].call(_p[1],arr[i],i,arr);}},_everyOrSome:function(_16b,arr,_16d,_16e){var _p=_158(arr,_16e,_16d);arr=_p[0];for(var i=0,l=arr.length;i<l;++i){var _172=!!_p[2].call(_p[1],arr[i],i,arr);if(_16b^_172){return _172;}}return _16b;},every:function(arr,_174,_175){return dojo._everyOrSome(true,arr,_174,_175);},some:function(arr,_177,_178){return dojo._everyOrSome(false,arr,_177,_178);},map:function(arr,_17a,_17b){var _p=_158(arr,_17b,_17a);arr=_p[0];var _17d=(arguments[3]?(new arguments[3]()):[]);for(var i=0,l=arr.length;i<l;++i){_17d.push(_p[2].call(_p[1],arr[i],i,arr));}return _17d;},filter:function(arr,_181,_182){var _p=_158(arr,_182,_181);arr=_p[0];var _184=[];for(var i=0,l=arr.length;i<l;++i){if(_p[2].call(_p[1],arr[i],i,arr)){_184.push(arr[i]);}}return _184;}});})();}if(!dojo._hasResource["dojo._base.Color"]){dojo._hasResource["dojo._base.Color"]=true;dojo.provide("dojo._base.Color");(function(){var d=dojo;dojo.Color=function(_188){if(_188){this.setColor(_188);}};dojo.Color.named={black:[0,0,0],silver:[192,192,192],gray:[128,128,128],white:[255,255,255],maroon:[128,0,0],red:[255,0,0],purple:[128,0,128],fuchsia:[255,0,255],green:[0,128,0],lime:[0,255,0],olive:[128,128,0],yellow:[255,255,0],navy:[0,0,128],blue:[0,0,255],teal:[0,128,128],aqua:[0,255,255]};dojo.extend(dojo.Color,{r:255,g:255,b:255,a:1,_set:function(r,g,b,a){var t=this;t.r=r;t.g=g;t.b=b;t.a=a;},setColor:function(_18e){if(d.isString(_18e)){d.colorFromString(_18e,this);}else{if(d.isArray(_18e)){d.colorFromArray(_18e,this);}else{this._set(_18e.r,_18e.g,_18e.b,_18e.a);if(!(_18e instanceof d.Color)){this.sanitize();}}}return this;},sanitize:function(){return this;},toRgb:function(){var t=this;return [t.r,t.g,t.b];},toRgba:function(){var t=this;return [t.r,t.g,t.b,t.a];},toHex:function(){var arr=d.map(["r","g","b"],function(x){var s=this[x].toString(16);return s.length<2?"0"+s:s;},this);return "#"+arr.join("");},toCss:function(_194){var t=this,rgb=t.r+", "+t.g+", "+t.b;return (_194?"rgba("+rgb+", "+t.a:"rgb("+rgb)+")";},toString:function(){return this.toCss(true);}});dojo.blendColors=function(_197,end,_199,obj){var t=obj||new d.Color();d.forEach(["r","g","b","a"],function(x){t[x]=_197[x]+(end[x]-_197[x])*_199;if(x!="a"){t[x]=Math.round(t[x]);}});return t.sanitize();};dojo.colorFromRgb=function(_19d,obj){var m=_19d.toLowerCase().match(/^rgba?\(([\s\.,0-9]+)\)/);return m&&dojo.colorFromArray(m[1].split(/\s*,\s*/),obj);};dojo.colorFromHex=function(_1a0,obj){var t=obj||new d.Color(),bits=(_1a0.length==4)?4:8,mask=(1<<bits)-1;_1a0=Number("0x"+_1a0.substr(1));if(isNaN(_1a0)){return null;}d.forEach(["b","g","r"],function(x){var c=_1a0&mask;_1a0>>=bits;t[x]=bits==4?17*c:c;});t.a=1;return t;};dojo.colorFromArray=function(a,obj){var t=obj||new d.Color();t._set(Number(a[0]),Number(a[1]),Number(a[2]),Number(a[3]));if(isNaN(t.a)){t.a=1;}return t.sanitize();};dojo.colorFromString=function(str,obj){var a=d.Color.named[str];return a&&d.colorFromArray(a,obj)||d.colorFromRgb(str,obj)||d.colorFromHex(str,obj);};})();}if(!dojo._hasResource["dojo._base"]){dojo._hasResource["dojo._base"]=true;dojo.provide("dojo._base");}if(!dojo._hasResource["dojo._base.window"]){dojo._hasResource["dojo._base.window"]=true;dojo.provide("dojo._base.window");dojo.doc=window["document"]||null;dojo.body=function(){return dojo.doc.body||dojo.doc.getElementsByTagName("body")[0];};dojo.setContext=function(_1ad,_1ae){dojo.global=_1ad;dojo.doc=_1ae;};dojo.withGlobal=function(_1af,_1b0,_1b1,_1b2){var _1b3=dojo.global;try{dojo.global=_1af;return dojo.withDoc.call(null,_1af.document,_1b0,_1b1,_1b2);}finally{dojo.global=_1b3;}};dojo.withDoc=function(_1b4,_1b5,_1b6,_1b7){var _1b8=dojo.doc,_1b9=dojo._bodyLtr;try{dojo.doc=_1b4;delete dojo._bodyLtr;if(_1b6&&dojo.isString(_1b5)){_1b5=_1b6[_1b5];}return _1b5.apply(_1b6,_1b7||[]);}finally{dojo.doc=_1b8;if(_1b9!==undefined){dojo._bodyLtr=_1b9;}}};}if(!dojo._hasResource["dojo._base.event"]){dojo._hasResource["dojo._base.event"]=true;dojo.provide("dojo._base.event");(function(){var del=(dojo._event_listener={add:function(node,name,fp){if(!node){return;}name=del._normalizeEventName(name);fp=del._fixCallback(name,fp);var _1be=name;if(!dojo.isIE&&(name=="mouseenter"||name=="mouseleave")){var ofp=fp;name=(name=="mouseenter")?"mouseover":"mouseout";fp=function(e){if(dojo.isFF<=2){try{e.relatedTarget.tagName;}catch(e2){return;}}if(!dojo.isDescendant(e.relatedTarget,node)){return ofp.call(this,e);}};}node.addEventListener(name,fp,false);return fp;},remove:function(node,_1c2,_1c3){if(node){_1c2=del._normalizeEventName(_1c2);if(!dojo.isIE&&(_1c2=="mouseenter"||_1c2=="mouseleave")){_1c2=(_1c2=="mouseenter")?"mouseover":"mouseout";}node.removeEventListener(_1c2,_1c3,false);}},_normalizeEventName:function(name){return name.slice(0,2)=="on"?name.slice(2):name;},_fixCallback:function(name,fp){return name!="keypress"?fp:function(e){return fp.call(this,del._fixEvent(e,this));};},_fixEvent:function(evt,_1c9){switch(evt.type){case "keypress":del._setKeyChar(evt);break;}return evt;},_setKeyChar:function(evt){evt.keyChar=evt.charCode?String.fromCharCode(evt.charCode):"";evt.charOrCode=evt.keyChar||evt.keyCode;},_punctMap:{106:42,111:47,186:59,187:43,188:44,189:45,190:46,191:47,192:96,219:91,220:92,221:93,222:39}});dojo.fixEvent=function(evt,_1cc){return del._fixEvent(evt,_1cc);};dojo.stopEvent=function(evt){evt.preventDefault();evt.stopPropagation();};var _1ce=dojo._listener;dojo._connect=function(obj,_1d0,_1d1,_1d2,_1d3){var _1d4=obj&&(obj.nodeType||obj.attachEvent||obj.addEventListener);var lid=_1d4?(_1d3?2:1):0,l=[dojo._listener,del,_1ce][lid];var h=l.add(obj,_1d0,dojo.hitch(_1d1,_1d2));return [obj,_1d0,h,lid];};dojo._disconnect=function(obj,_1d9,_1da,_1db){([dojo._listener,del,_1ce][_1db]).remove(obj,_1d9,_1da);};dojo.keys={BACKSPACE:8,TAB:9,CLEAR:12,ENTER:13,SHIFT:16,CTRL:17,ALT:18,PAUSE:19,CAPS_LOCK:20,ESCAPE:27,SPACE:32,PAGE_UP:33,PAGE_DOWN:34,END:35,HOME:36,LEFT_ARROW:37,UP_ARROW:38,RIGHT_ARROW:39,DOWN_ARROW:40,INSERT:45,DELETE:46,HELP:47,LEFT_WINDOW:91,RIGHT_WINDOW:92,SELECT:93,NUMPAD_0:96,NUMPAD_1:97,NUMPAD_2:98,NUMPAD_3:99,NUMPAD_4:100,NUMPAD_5:101,NUMPAD_6:102,NUMPAD_7:103,NUMPAD_8:104,NUMPAD_9:105,NUMPAD_MULTIPLY:106,NUMPAD_PLUS:107,NUMPAD_ENTER:108,NUMPAD_MINUS:109,NUMPAD_PERIOD:110,NUMPAD_DIVIDE:111,F1:112,F2:113,F3:114,F4:115,F5:116,F6:117,F7:118,F8:119,F9:120,F10:121,F11:122,F12:123,F13:124,F14:125,F15:126,NUM_LOCK:144,SCROLL_LOCK:145};if(dojo.isIE){var _1dc=function(e,code){try{return (e.keyCode=code);}catch(e){return 0;}};var iel=dojo._listener;var _1e0=(dojo._ieListenersName="_"+dojo._scopeName+"_listeners");if(!dojo.config._allow_leaks){_1ce=iel=dojo._ie_listener={handlers:[],add:function(_1e1,_1e2,_1e3){_1e1=_1e1||dojo.global;var f=_1e1[_1e2];if(!f||!f[_1e0]){var d=dojo._getIeDispatcher();d.target=f&&(ieh.push(f)-1);d[_1e0]=[];f=_1e1[_1e2]=d;}return f[_1e0].push(ieh.push(_1e3)-1);},remove:function(_1e7,_1e8,_1e9){var f=(_1e7||dojo.global)[_1e8],l=f&&f[_1e0];if(f&&l&&_1e9--){delete ieh[l[_1e9]];delete l[_1e9];}}};var ieh=iel.handlers;}dojo.mixin(del,{add:function(node,_1ed,fp){if(!node){return;}_1ed=del._normalizeEventName(_1ed);if(_1ed=="onkeypress"){var kd=node.onkeydown;if(!kd||!kd[_1e0]||!kd._stealthKeydownHandle){var h=del.add(node,"onkeydown",del._stealthKeyDown);kd=node.onkeydown;kd._stealthKeydownHandle=h;kd._stealthKeydownRefs=1;}else{kd._stealthKeydownRefs++;}}return iel.add(node,_1ed,del._fixCallback(fp));},remove:function(node,_1f2,_1f3){_1f2=del._normalizeEventName(_1f2);iel.remove(node,_1f2,_1f3);if(_1f2=="onkeypress"){var kd=node.onkeydown;if(--kd._stealthKeydownRefs<=0){iel.remove(node,"onkeydown",kd._stealthKeydownHandle);delete kd._stealthKeydownHandle;}}},_normalizeEventName:function(_1f5){return _1f5.slice(0,2)!="on"?"on"+_1f5:_1f5;},_nop:function(){},_fixEvent:function(evt,_1f7){if(!evt){var w=_1f7&&(_1f7.ownerDocument||_1f7.document||_1f7).parentWindow||window;evt=w.event;}if(!evt){return (evt);}evt.target=evt.srcElement;evt.currentTarget=(_1f7||evt.srcElement);evt.layerX=evt.offsetX;evt.layerY=evt.offsetY;var se=evt.srcElement,doc=(se&&se.ownerDocument)||document;var _1fb=((dojo.isIE<6)||(doc["compatMode"]=="BackCompat"))?doc.body:doc.documentElement;var _1fc=dojo._getIeDocumentElementOffset();evt.pageX=evt.clientX+dojo._fixIeBiDiScrollLeft(_1fb.scrollLeft||0)-_1fc.x;evt.pageY=evt.clientY+(_1fb.scrollTop||0)-_1fc.y;if(evt.type=="mouseover"){evt.relatedTarget=evt.fromElement;}if(evt.type=="mouseout"){evt.relatedTarget=evt.toElement;}evt.stopPropagation=del._stopPropagation;evt.preventDefault=del._preventDefault;return del._fixKeys(evt);},_fixKeys:function(evt){switch(evt.type){case "keypress":var c=("charCode" in evt?evt.charCode:evt.keyCode);if(c==10){c=0;evt.keyCode=13;}else{if(c==13||c==27){c=0;}else{if(c==3){c=99;}}}evt.charCode=c;del._setKeyChar(evt);break;}return evt;},_stealthKeyDown:function(evt){var kp=evt.currentTarget.onkeypress;if(!kp||!kp[_1e0]){return;}var k=evt.keyCode;var _202=k!=13&&k!=32&&k!=27&&(k<48||k>90)&&(k<96||k>111)&&(k<186||k>192)&&(k<219||k>222);if(_202||evt.ctrlKey){var c=_202?0:k;if(evt.ctrlKey){if(k==3||k==13){return;}else{if(c>95&&c<106){c-=48;}else{if((!evt.shiftKey)&&(c>=65&&c<=90)){c+=32;}else{c=del._punctMap[c]||c;}}}}var faux=del._synthesizeEvent(evt,{type:"keypress",faux:true,charCode:c});kp.call(evt.currentTarget,faux);evt.cancelBubble=faux.cancelBubble;evt.returnValue=faux.returnValue;_1dc(evt,faux.keyCode);}},_stopPropagation:function(){this.cancelBubble=true;},_preventDefault:function(){this.bubbledKeyCode=this.keyCode;if(this.ctrlKey){_1dc(this,0);}this.returnValue=false;}});dojo.stopEvent=function(evt){evt=evt||window.event;del._stopPropagation.call(evt);del._preventDefault.call(evt);};}del._synthesizeEvent=function(evt,_207){var faux=dojo.mixin({},evt,_207);del._setKeyChar(faux);faux.preventDefault=function(){evt.preventDefault();};faux.stopPropagation=function(){evt.stopPropagation();};return faux;};if(dojo.isOpera){dojo.mixin(del,{_fixEvent:function(evt,_20a){switch(evt.type){case "keypress":var c=evt.which;if(c==3){c=99;}c=c<41&&!evt.shiftKey?0:c;if(evt.ctrlKey&&!evt.shiftKey&&c>=65&&c<=90){c+=32;}return del._synthesizeEvent(evt,{charCode:c});}return evt;}});}if(dojo.isWebKit){del._add=del.add;del._remove=del.remove;dojo.mixin(del,{add:function(node,_20d,fp){if(!node){return;}var _20f=del._add(node,_20d,fp);if(del._normalizeEventName(_20d)=="keypress"){_20f._stealthKeyDownHandle=del._add(node,"keydown",function(evt){var k=evt.keyCode;var _212=k!=13&&k!=32&&k!=27&&(k<48||k>90)&&(k<96||k>111)&&(k<186||k>192)&&(k<219||k>222);if(_212||evt.ctrlKey){var c=_212?0:k;if(evt.ctrlKey){if(k==3||k==13){return;}else{if(c>95&&c<106){c-=48;}else{if(!evt.shiftKey&&c>=65&&c<=90){c+=32;}else{c=del._punctMap[c]||c;}}}}var faux=del._synthesizeEvent(evt,{type:"keypress",faux:true,charCode:c});fp.call(evt.currentTarget,faux);}});}return _20f;},remove:function(node,_216,_217){if(node){if(_217._stealthKeyDownHandle){del._remove(node,"keydown",_217._stealthKeyDownHandle);}del._remove(node,_216,_217);}},_fixEvent:function(evt,_219){switch(evt.type){case "keypress":if(evt.faux){return evt;}var c=evt.charCode;c=c>=32?c:0;return del._synthesizeEvent(evt,{charCode:c,faux:true});}return evt;}});}})();if(dojo.isIE){dojo._ieDispatcher=function(args,_21c){var ap=Array.prototype,h=dojo._ie_listener.handlers,c=args.callee,ls=c[dojo._ieListenersName],t=h[c.target];var r=t&&t.apply(_21c,args);var lls=[].concat(ls);for(var i in lls){var f=h[lls[i]];if(!(i in ap)&&f){f.apply(_21c,args);}}return r;};dojo._getIeDispatcher=function(){return new Function(dojo._scopeName+"._ieDispatcher(arguments, this)");};dojo._event_listener._fixCallback=function(fp){var f=dojo._event_listener._fixEvent;return function(e){return fp.call(this,f(e,this));};};}}if(!dojo._hasResource["dojo._base.html"]){dojo._hasResource["dojo._base.html"]=true;dojo.provide("dojo._base.html");try{document.execCommand("BackgroundImageCache",false,true);}catch(e){}if(dojo.isIE||dojo.isOpera){dojo.byId=function(id,doc){if(dojo.isString(id)){var _d=doc||dojo.doc;var te=_d.getElementById(id);if(te&&(te.attributes.id.value==id||te.id==id)){return te;}else{var eles=_d.all[id];if(!eles||eles.nodeName){eles=[eles];}var i=0;while((te=eles[i++])){if((te.attributes&&te.attributes.id&&te.attributes.id.value==id)||te.id==id){return te;}}}}else{return id;}};}else{dojo.byId=function(id,doc){return dojo.isString(id)?(doc||dojo.doc).getElementById(id):id;};}(function(){var d=dojo;var _232=null;d.addOnWindowUnload(function(){_232=null;});dojo._destroyElement=dojo.destroy=function(node){node=d.byId(node);try{if(!_232||_232.ownerDocument!=node.ownerDocument){_232=node.ownerDocument.createElement("div");}_232.appendChild(node.parentNode?node.parentNode.removeChild(node):node);_232.innerHTML="";}catch(e){}};dojo.isDescendant=function(node,_235){try{node=d.byId(node);_235=d.byId(_235);while(node){if(node===_235){return true;}node=node.parentNode;}}catch(e){}return false;};dojo.setSelectable=function(node,_237){node=d.byId(node);if(d.isMozilla){node.style.MozUserSelect=_237?"":"none";}else{if(d.isKhtml||d.isWebKit){node.style.KhtmlUserSelect=_237?"auto":"none";}else{if(d.isIE){var v=(node.unselectable=_237?"":"on");d.query("*",node).forEach("item.unselectable = '"+v+"'");}}}};var _239=function(node,ref){var _23c=ref.parentNode;if(_23c){_23c.insertBefore(node,ref);}};var _23d=function(node,ref){var _240=ref.parentNode;if(_240){if(_240.lastChild==ref){_240.appendChild(node);}else{_240.insertBefore(node,ref.nextSibling);}}};dojo.place=function(node,_242,_243){_242=d.byId(_242);if(d.isString(node)){node=node.charAt(0)=="<"?d._toDom(node,_242.ownerDocument):d.byId(node);}if(typeof _243=="number"){var cn=_242.childNodes;if(!cn.length||cn.length<=_243){_242.appendChild(node);}else{_239(node,cn[_243<0?0:_243]);}}else{switch(_243){case "before":_239(node,_242);break;case "after":_23d(node,_242);break;case "replace":_242.parentNode.replaceChild(node,_242);break;case "only":d.empty(_242);_242.appendChild(node);break;case "first":if(_242.firstChild){_239(node,_242.firstChild);break;}default:_242.appendChild(node);}}return node;};dojo.boxModel="content-box";if(d.isIE){var _dcm=document.compatMode;d.boxModel=_dcm=="BackCompat"||_dcm=="QuirksMode"||d.isIE<6?"border-box":"content-box";}var gcs;if(d.isWebKit){gcs=function(node){var s;if(node.nodeType==1){var dv=node.ownerDocument.defaultView;s=dv.getComputedStyle(node,null);if(!s&&node.style){node.style.display="";s=dv.getComputedStyle(node,null);}}return s||{};};}else{if(d.isIE){gcs=function(node){return node.nodeType==1?node.currentStyle:{};};}else{gcs=function(node){return node.nodeType==1?node.ownerDocument.defaultView.getComputedStyle(node,null):{};};}}dojo.getComputedStyle=gcs;if(!d.isIE){d._toPixelValue=function(_24c,_24d){return parseFloat(_24d)||0;};}else{d._toPixelValue=function(_24e,_24f){if(!_24f){return 0;}if(_24f=="medium"){return 4;}if(_24f.slice&&_24f.slice(-2)=="px"){return parseFloat(_24f);}with(_24e){var _250=style.left;var _251=runtimeStyle.left;runtimeStyle.left=currentStyle.left;try{style.left=_24f;_24f=style.pixelLeft;}catch(e){_24f=0;}style.left=_250;runtimeStyle.left=_251;}return _24f;};}var px=d._toPixelValue;var astr="DXImageTransform.Microsoft.Alpha";var af=function(n,f){try{return n.filters.item(astr);}catch(e){return f?{}:null;}};dojo._getOpacity=d.isIE?function(node){try{return af(node).Opacity/100;}catch(e){return 1;}}:function(node){return gcs(node).opacity;};dojo._setOpacity=d.isIE?function(node,_25a){var ov=_25a*100;node.style.zoom=1;af(node,1).Enabled=!(_25a==1);if(!af(node)){node.style.filter+=" progid:"+astr+"(Opacity="+ov+")";}else{af(node,1).Opacity=ov;}if(node.nodeName.toLowerCase()=="tr"){d.query("> td",node).forEach(function(i){d._setOpacity(i,_25a);});}return _25a;}:function(node,_25e){return node.style.opacity=_25e;};var _25f={left:true,top:true};var _260=/margin|padding|width|height|max|min|offset/;var _261=function(node,type,_264){type=type.toLowerCase();if(d.isIE){if(_264=="auto"){if(type=="height"){return node.offsetHeight;}if(type=="width"){return node.offsetWidth;}}if(type=="fontweight"){switch(_264){case 700:return "bold";case 400:default:return "normal";}}}if(!(type in _25f)){_25f[type]=_260.test(type);}return _25f[type]?px(node,_264):_264;};var _265=d.isIE?"styleFloat":"cssFloat",_266={"cssFloat":_265,"styleFloat":_265,"float":_265};dojo.style=function(node,_268,_269){var n=d.byId(node),args=arguments.length,op=(_268=="opacity");_268=_266[_268]||_268;if(args==3){return op?d._setOpacity(n,_269):n.style[_268]=_269;}if(args==2&&op){return d._getOpacity(n);}var s=gcs(n);if(args==2&&!d.isString(_268)){for(var x in _268){d.style(node,x,_268[x]);}return s;}return (args==1)?s:_261(n,_268,s[_268]||n.style[_268]);};dojo._getPadExtents=function(n,_270){var s=_270||gcs(n),l=px(n,s.paddingLeft),t=px(n,s.paddingTop);return {l:l,t:t,w:l+px(n,s.paddingRight),h:t+px(n,s.paddingBottom)};};dojo._getBorderExtents=function(n,_275){var ne="none",s=_275||gcs(n),bl=(s.borderLeftStyle!=ne?px(n,s.borderLeftWidth):0),bt=(s.borderTopStyle!=ne?px(n,s.borderTopWidth):0);return {l:bl,t:bt,w:bl+(s.borderRightStyle!=ne?px(n,s.borderRightWidth):0),h:bt+(s.borderBottomStyle!=ne?px(n,s.borderBottomWidth):0)};};dojo._getPadBorderExtents=function(n,_27b){var s=_27b||gcs(n),p=d._getPadExtents(n,s),b=d._getBorderExtents(n,s);return {l:p.l+b.l,t:p.t+b.t,w:p.w+b.w,h:p.h+b.h};};dojo._getMarginExtents=function(n,_280){var s=_280||gcs(n),l=px(n,s.marginLeft),t=px(n,s.marginTop),r=px(n,s.marginRight),b=px(n,s.marginBottom);if(d.isWebKit&&(s.position!="absolute")){r=l;}return {l:l,t:t,w:l+r,h:t+b};};dojo._getMarginBox=function(node,_287){var s=_287||gcs(node),me=d._getMarginExtents(node,s);var l=node.offsetLeft-me.l,t=node.offsetTop-me.t,p=node.parentNode;if(d.isMoz){var sl=parseFloat(s.left),st=parseFloat(s.top);if(!isNaN(sl)&&!isNaN(st)){l=sl,t=st;}else{if(p&&p.style){var pcs=gcs(p);if(pcs.overflow!="visible"){var be=d._getBorderExtents(p,pcs);l+=be.l,t+=be.t;}}}}else{if(d.isOpera||(d.isIE>7&&!d.isQuirks)){if(p){be=d._getBorderExtents(p);l-=be.l;t-=be.t;}}}return {l:l,t:t,w:node.offsetWidth+me.w,h:node.offsetHeight+me.h};};dojo._getContentBox=function(node,_292){var s=_292||gcs(node),pe=d._getPadExtents(node,s),be=d._getBorderExtents(node,s),w=node.clientWidth,h;if(!w){w=node.offsetWidth,h=node.offsetHeight;}else{h=node.clientHeight,be.w=be.h=0;}if(d.isOpera){pe.l+=be.l;pe.t+=be.t;}return {l:pe.l,t:pe.t,w:w-pe.w-be.w,h:h-pe.h-be.h};};dojo._getBorderBox=function(node,_299){var s=_299||gcs(node),pe=d._getPadExtents(node,s),cb=d._getContentBox(node,s);return {l:cb.l-pe.l,t:cb.t-pe.t,w:cb.w+pe.w,h:cb.h+pe.h};};dojo._setBox=function(node,l,t,w,h,u){u=u||"px";var s=node.style;if(!isNaN(l)){s.left=l+u;}if(!isNaN(t)){s.top=t+u;}if(w>=0){s.width=w+u;}if(h>=0){s.height=h+u;}};dojo._isButtonTag=function(node){return node.tagName=="BUTTON"||node.tagName=="INPUT"&&node.getAttribute("type").toUpperCase()=="BUTTON";};dojo._usesBorderBox=function(node){var n=node.tagName;return d.boxModel=="border-box"||n=="TABLE"||d._isButtonTag(node);};dojo._setContentSize=function(node,_2a8,_2a9,_2aa){if(d._usesBorderBox(node)){var pb=d._getPadBorderExtents(node,_2aa);if(_2a8>=0){_2a8+=pb.w;}if(_2a9>=0){_2a9+=pb.h;}}d._setBox(node,NaN,NaN,_2a8,_2a9);};dojo._setMarginBox=function(node,_2ad,_2ae,_2af,_2b0,_2b1){var s=_2b1||gcs(node),bb=d._usesBorderBox(node),pb=bb?_2b5:d._getPadBorderExtents(node,s);if(d.isWebKit){if(d._isButtonTag(node)){var ns=node.style;if(_2af>=0&&!ns.width){ns.width="4px";}if(_2b0>=0&&!ns.height){ns.height="4px";}}}var mb=d._getMarginExtents(node,s);if(_2af>=0){_2af=Math.max(_2af-pb.w-mb.w,0);}if(_2b0>=0){_2b0=Math.max(_2b0-pb.h-mb.h,0);}d._setBox(node,_2ad,_2ae,_2af,_2b0);};var _2b5={l:0,t:0,w:0,h:0};dojo.marginBox=function(node,box){var n=d.byId(node),s=gcs(n),b=box;return !b?d._getMarginBox(n,s):d._setMarginBox(n,b.l,b.t,b.w,b.h,s);};dojo.contentBox=function(node,box){var n=d.byId(node),s=gcs(n),b=box;return !b?d._getContentBox(n,s):d._setContentSize(n,b.w,b.h,s);};var _2c2=function(node,prop){if(!(node=(node||0).parentNode)){return 0;}var val,_2c6=0,_b=d.body();while(node&&node.style){if(gcs(node).position=="fixed"){return 0;}val=node[prop];if(val){_2c6+=val-0;if(node==_b){break;}}node=node.parentNode;}return _2c6;};dojo._docScroll=function(){var _b=d.body(),_w=d.global,de=d.doc.documentElement;return {y:(_w.pageYOffset||de.scrollTop||_b.scrollTop||0),x:(_w.pageXOffset||d._fixIeBiDiScrollLeft(de.scrollLeft)||_b.scrollLeft||0)};};dojo._isBodyLtr=function(){return "_bodyLtr" in d?d._bodyLtr:d._bodyLtr=(d.body().dir||d.doc.documentElement.dir||"ltr").toLowerCase()=="ltr";};dojo._getIeDocumentElementOffset=function(){var de=d.doc.documentElement;if(d.isIE<7){return {x:d._isBodyLtr()||window.parent==window?de.clientLeft:de.offsetWidth-de.clientWidth-de.clientLeft,y:de.clientTop};}else{if(d.isIE<8){return {x:de.getBoundingClientRect().left,y:de.getBoundingClientRect().top};}else{return {x:0,y:0};}}};dojo._fixIeBiDiScrollLeft=function(_2cc){var dd=d.doc;if(d.isIE<8&&!d._isBodyLtr()){var de=dd.compatMode=="BackCompat"?dd.body:dd.documentElement;return _2cc+de.clientWidth-de.scrollWidth;}return _2cc;};dojo._abs=function(node,_2d0){var db=d.body(),dh=d.body().parentNode,ret;if(node["getBoundingClientRect"]){var _2d4=node.getBoundingClientRect();ret={x:_2d4.left,y:_2d4.top};if(d.isFF>=3){var cs=gcs(dh);ret.x-=px(dh,cs.marginLeft)+px(dh,cs.borderLeftWidth);ret.y-=px(dh,cs.marginTop)+px(dh,cs.borderTopWidth);}if(d.isIE){var _2d6=d._getIeDocumentElementOffset();ret.x-=_2d6.x+(d.isQuirks?db.clientLeft:0);ret.y-=_2d6.y+(d.isQuirks?db.clientTop:0);}}else{ret={x:0,y:0};if(node["offsetParent"]){ret.x-=_2c2(node,"scrollLeft");ret.y-=_2c2(node,"scrollTop");var _2d7=node;do{var n=_2d7.offsetLeft,t=_2d7.offsetTop;ret.x+=isNaN(n)?0:n;ret.y+=isNaN(t)?0:t;cs=gcs(_2d7);if(_2d7!=node){if(d.isFF){ret.x+=2*px(_2d7,cs.borderLeftWidth);ret.y+=2*px(_2d7,cs.borderTopWidth);}else{ret.x+=px(_2d7,cs.borderLeftWidth);ret.y+=px(_2d7,cs.borderTopWidth);}}if(d.isFF&&cs.position=="static"){var _2da=_2d7.parentNode;while(_2da!=_2d7.offsetParent){var pcs=gcs(_2da);if(pcs.position=="static"){ret.x+=px(_2d7,pcs.borderLeftWidth);ret.y+=px(_2d7,pcs.borderTopWidth);}_2da=_2da.parentNode;}}_2d7=_2d7.offsetParent;}while((_2d7!=dh)&&_2d7);}else{if(node.x&&node.y){ret.x+=isNaN(node.x)?0:node.x;ret.y+=isNaN(node.y)?0:node.y;}}}if(_2d0){var _2dc=d._docScroll();ret.x+=_2dc.x;ret.y+=_2dc.y;}return ret;};dojo.coords=function(node,_2de){var n=d.byId(node),s=gcs(n),mb=d._getMarginBox(n,s);var abs=d._abs(n,_2de);mb.x=abs.x;mb.y=abs.y;return mb;};var _2e3=d.isIE<8;var _2e4=function(name){switch(name.toLowerCase()){case "tabindex":return _2e3?"tabIndex":"tabindex";case "readonly":return "readOnly";case "class":return "className";case "for":case "htmlfor":return _2e3?"htmlFor":"for";default:return name;}};var _2e6={colspan:"colSpan",enctype:"enctype",frameborder:"frameborder",method:"method",rowspan:"rowSpan",scrolling:"scrolling",shape:"shape",span:"span",type:"type",valuetype:"valueType",classname:"className",innerhtml:"innerHTML"};dojo.hasAttr=function(node,name){node=d.byId(node);var _2e9=_2e4(name);_2e9=_2e9=="htmlFor"?"for":_2e9;var attr=node.getAttributeNode&&node.getAttributeNode(_2e9);return attr?attr.specified:false;};var _2eb={},_ctr=0,_2ed=dojo._scopeName+"attrid",_2ee={col:1,colgroup:1,table:1,tbody:1,tfoot:1,thead:1,tr:1,title:1};dojo.attr=function(node,name,_2f1){node=d.byId(node);var args=arguments.length;if(args==2&&!d.isString(name)){for(var x in name){d.attr(node,x,name[x]);}return;}name=_2e4(name);if(args==3){if(d.isFunction(_2f1)){var _2f4=d.attr(node,_2ed);if(!_2f4){_2f4=_ctr++;d.attr(node,_2ed,_2f4);}if(!_2eb[_2f4]){_2eb[_2f4]={};}var h=_2eb[_2f4][name];if(h){d.disconnect(h);}else{try{delete node[name];}catch(e){}}_2eb[_2f4][name]=d.connect(node,name,_2f1);}else{if(typeof _2f1=="boolean"){node[name]=_2f1;}else{if(name==="style"&&!d.isString(_2f1)){d.style(node,_2f1);}else{if(name=="className"){node.className=_2f1;}else{if(name==="innerHTML"){if(d.isIE&&node.tagName.toLowerCase() in _2ee){d.empty(node);node.appendChild(d._toDom(_2f1,node.ownerDocument));}else{node[name]=_2f1;}}else{node.setAttribute(name,_2f1);}}}}}}else{var prop=_2e6[name.toLowerCase()];if(prop){return node[prop];}var _2f7=node[name];return (typeof _2f7=="boolean"||typeof _2f7=="function")?_2f7:(d.hasAttr(node,name)?node.getAttribute(name):null);}};dojo.removeAttr=function(node,name){d.byId(node).removeAttribute(_2e4(name));};dojo.create=function(tag,_2fb,_2fc,pos){var doc=d.doc;if(_2fc){_2fc=d.byId(_2fc);doc=_2fc.ownerDocument;}if(d.isString(tag)){tag=doc.createElement(tag);}if(_2fb){d.attr(tag,_2fb);}if(_2fc){d.place(tag,_2fc,pos);}return tag;};d.empty=d.isIE?function(node){node=d.byId(node);for(var c;c=node.lastChild;){d.destroy(c);}}:function(node){d.byId(node).innerHTML="";};var _302={option:["select"],tbody:["table"],thead:["table"],tfoot:["table"],tr:["table","tbody"],td:["table","tbody","tr"],th:["table","thead","tr"],legend:["fieldset"],caption:["table"],colgroup:["table"],col:["table","colgroup"],li:["ul"]},_303=/<\s*([\w\:]+)/,_304={},_305=0,_306="__"+d._scopeName+"ToDomId";for(var _307 in _302){var tw=_302[_307];tw.pre=_307=="option"?"<select multiple=\"multiple\">":"<"+tw.join("><")+">";tw.post="</"+tw.reverse().join("></")+">";}d._toDom=function(frag,doc){doc=doc||d.doc;var _30b=doc[_306];if(!_30b){doc[_306]=_30b=++_305+"";_304[_30b]=doc.createElement("div");}frag+="";var _30c=frag.match(_303),tag=_30c?_30c[1].toLowerCase():"",_30e=_304[_30b],wrap,i,fc,df;if(_30c&&_302[tag]){wrap=_302[tag];_30e.innerHTML=wrap.pre+frag+wrap.post;for(i=wrap.length;i;--i){_30e=_30e.firstChild;}}else{_30e.innerHTML=frag;}if(_30e.childNodes.length==1){return _30e.removeChild(_30e.firstChild);}df=doc.createDocumentFragment();while(fc=_30e.firstChild){df.appendChild(fc);}return df;};var _312="className";dojo.hasClass=function(node,_314){return ((" "+d.byId(node)[_312]+" ").indexOf(" "+_314+" ")>=0);};dojo.addClass=function(node,_316){node=d.byId(node);var cls=node[_312];if((" "+cls+" ").indexOf(" "+_316+" ")<0){node[_312]=cls+(cls?" ":"")+_316;}};dojo.removeClass=function(node,_319){node=d.byId(node);var t=d.trim((" "+node[_312]+" ").replace(" "+_319+" "," "));if(node[_312]!=t){node[_312]=t;}};dojo.toggleClass=function(node,_31c,_31d){if(_31d===undefined){_31d=!d.hasClass(node,_31c);}d[_31d?"addClass":"removeClass"](node,_31c);};})();}if(!dojo._hasResource["dojo._base.NodeList"]){dojo._hasResource["dojo._base.NodeList"]=true;dojo.provide("dojo._base.NodeList");(function(){var d=dojo;var ap=Array.prototype,aps=ap.slice,apc=ap.concat;var tnl=function(a){a.constructor=d.NodeList;dojo._mixin(a,d.NodeList.prototype);return a;};var _324=function(f,a,o){a=[0].concat(aps.call(a,0));if(!a.sort){a=aps.call(a,0);}o=o||d.global;return function(node){a[0]=node;return f.apply(o,a);};};var _329=function(f,o){return function(){this.forEach(_324(f,arguments,o));return this;};};var _32c=function(f,o){return function(){return this.map(_324(f,arguments,o));};};var _32f=function(f,o){return function(){return this.filter(_324(f,arguments,o));};};var _332=function(f,g,o){return function(){var a=arguments,body=_324(f,a,o);if(g.call(o||d.global,a)){return this.map(body);}this.forEach(body);return this;};};var _338=function(a){return a.length==1&&d.isString(a[0]);};var _33a=function(node){var p=node.parentNode;if(p){p.removeChild(node);}};dojo.NodeList=function(){return tnl(Array.apply(null,arguments));};var nl=d.NodeList,nlp=nl.prototype;nl._wrap=tnl;nl._adaptAsMap=_32c;nl._adaptAsForEach=_329;nl._adaptAsFilter=_32f;nl._adaptWithCondition=_332;d.forEach(["slice","splice"],function(name){var f=ap[name];nlp[name]=function(){return tnl(f.apply(this,arguments));};});d.forEach(["indexOf","lastIndexOf","every","some"],function(name){var f=d[name];nlp[name]=function(){return f.apply(d,[this].concat(aps.call(arguments,0)));};});d.forEach(["attr","style"],function(name){nlp[name]=_332(d[name],_338);});d.forEach(["connect","addClass","removeClass","toggleClass","empty"],function(name){nlp[name]=_329(d[name]);});dojo.extend(dojo.NodeList,{concat:function(item){var t=d.isArray(this)?this:aps.call(this,0),m=d.map(arguments,function(a){return a&&!d.isArray(a)&&(a.constructor===NodeList||a.constructor==nl)?aps.call(a,0):a;});return tnl(apc.apply(t,m));},map:function(func,obj){return tnl(d.map(this,func,obj));},forEach:function(_34b,_34c){d.forEach(this,_34b,_34c);return this;},coords:_32c(d.coords),place:function(_34d,_34e){var item=d.query(_34d)[0];return this.forEach(function(node){d.place(node,item,_34e);});},orphan:function(_351){return (_351?d._filterQueryResult(this,_351):this).forEach(_33a);},adopt:function(_352,_353){return d.query(_352).place(this[0],_353);},query:function(_354){if(!_354){return this;}var ret=this.map(function(node){return d.query(_354,node).filter(function(_357){return _357!==undefined;});});return tnl(apc.apply([],ret));},filter:function(_358){var a=arguments,_35a=this,_35b=0;if(d.isString(_358)){_35a=d._filterQueryResult(this,a[0]);if(a.length==1){return _35a;}_35b=1;}return tnl(d.filter(_35a,a[_35b],a[_35b+1]));},addContent:function(_35c,_35d){var c=d.isString(_35c)?d._toDom(_35c,this[0]&&this[0].ownerDocument):_35c,i,l=this.length-1;for(i=0;i<l;++i){d.place(c.cloneNode(true),this[i],_35d);}if(l>=0){d.place(c,this[l],_35d);}return this;},instantiate:function(_360,_361){var c=d.isFunction(_360)?_360:d.getObject(_360);_361=_361||{};return this.forEach(function(node){new c(_361,node);});},at:function(){var t=new dojo.NodeList();d.forEach(arguments,function(i){if(this[i]){t.push(this[i]);}},this);return t;}});d.forEach(["blur","focus","change","click","error","keydown","keypress","keyup","load","mousedown","mouseenter","mouseleave","mousemove","mouseout","mouseover","mouseup","submit"],function(evt){var _oe="on"+evt;nlp[_oe]=function(a,b){return this.connect(_oe,a,b);};});})();}if(!dojo._hasResource["dojo._base.query"]){dojo._hasResource["dojo._base.query"]=true;if(typeof dojo!="undefined"){dojo.provide("dojo._base.query");}(function(d){var trim=d.trim;var each=d.forEach;var qlc=d._queryListCtor=d.NodeList;var _36e=d.isString;var _36f=function(){return d.doc;};var _370=((d.isWebKit||d.isMozilla)&&((_36f().compatMode)=="BackCompat"));var _371=!!_36f().firstChild["children"]?"children":"childNodes";var _372=">~+";var _373=false;var _374=function(){return true;};var _375=function(_376){if(_372.indexOf(_376.slice(-1))>=0){_376+=" * ";}else{_376+=" ";}var ts=function(s,e){return trim(_376.slice(s,e));};var _37a=[];var _37b=-1,_37c=-1,_37d=-1,_37e=-1,_37f=-1,inId=-1,_381=-1,lc="",cc="",_384;var x=0,ql=_376.length,_387=null,_cp=null;var _389=function(){if(_381>=0){var tv=(_381==x)?null:ts(_381,x);_387[(_372.indexOf(tv)<0)?"tag":"oper"]=tv;_381=-1;}};var _38b=function(){if(inId>=0){_387.id=ts(inId,x).replace(/\\/g,"");inId=-1;}};var _38c=function(){if(_37f>=0){_387.classes.push(ts(_37f+1,x).replace(/\\/g,""));_37f=-1;}};var _38d=function(){_38b();_389();_38c();};var _38e=function(){_38d();if(_37e>=0){_387.pseudos.push({name:ts(_37e+1,x)});}_387.loops=(_387.pseudos.length||_387.attrs.length||_387.classes.length);_387.oquery=_387.query=ts(_384,x);_387.otag=_387.tag=(_387["oper"])?null:(_387.tag||"*");if(_387.tag){_387.tag=_387.tag.toUpperCase();}if(_37a.length&&(_37a[_37a.length-1].oper)){_387.infixOper=_37a.pop();_387.query=_387.infixOper.query+" "+_387.query;}_37a.push(_387);_387=null;};for(;lc=cc,cc=_376.charAt(x),x<ql;x++){if(lc=="\\"){continue;}if(!_387){_384=x;_387={query:null,pseudos:[],attrs:[],classes:[],tag:null,oper:null,id:null,getTag:function(){return (_373)?this.otag:this.tag;}};_381=x;}if(_37b>=0){if(cc=="]"){if(!_cp.attr){_cp.attr=ts(_37b+1,x);}else{_cp.matchFor=ts((_37d||_37b+1),x);}var cmf=_cp.matchFor;if(cmf){if((cmf.charAt(0)=="\"")||(cmf.charAt(0)=="'")){_cp.matchFor=cmf.slice(1,-1);}}_387.attrs.push(_cp);_cp=null;_37b=_37d=-1;}else{if(cc=="="){var _390=("|~^$*".indexOf(lc)>=0)?lc:"";_cp.type=_390+cc;_cp.attr=ts(_37b+1,x-_390.length);_37d=x+1;}}}else{if(_37c>=0){if(cc==")"){if(_37e>=0){_cp.value=ts(_37c+1,x);}_37e=_37c=-1;}}else{if(cc=="#"){_38d();inId=x+1;}else{if(cc=="."){_38d();_37f=x;}else{if(cc==":"){_38d();_37e=x;}else{if(cc=="["){_38d();_37b=x;_cp={};}else{if(cc=="("){if(_37e>=0){_cp={name:ts(_37e+1,x),value:null};_387.pseudos.push(_cp);}_37c=x;}else{if((cc==" ")&&(lc!=cc)){_38e();}}}}}}}}}return _37a;};var _391=function(_392,_393){if(!_392){return _393;}if(!_393){return _392;}return function(){return _392.apply(window,arguments)&&_393.apply(window,arguments);};};var _394=function(i,arr){var r=arr||[];if(i){r.push(i);}return r;};var _398=function(n){return (1==n.nodeType);};var _39a="";var _39b=function(elem,attr){if(!elem){return _39a;}if(attr=="class"){return elem.className||_39a;}if(attr=="for"){return elem.htmlFor||_39a;}if(attr=="style"){return elem.style.cssText||_39a;}return (_373?elem.getAttribute(attr):elem.getAttribute(attr,2))||_39a;};var _39e={"*=":function(attr,_3a0){return function(elem){return (_39b(elem,attr).indexOf(_3a0)>=0);};},"^=":function(attr,_3a3){return function(elem){return (_39b(elem,attr).indexOf(_3a3)==0);};},"$=":function(attr,_3a6){var tval=" "+_3a6;return function(elem){var ea=" "+_39b(elem,attr);return (ea.lastIndexOf(_3a6)==(ea.length-_3a6.length));};},"~=":function(attr,_3ab){var tval=" "+_3ab+" ";return function(elem){var ea=" "+_39b(elem,attr)+" ";return (ea.indexOf(tval)>=0);};},"|=":function(attr,_3b0){var _3b1=" "+_3b0+"-";return function(elem){var ea=" "+_39b(elem,attr);return ((ea==_3b0)||(ea.indexOf(_3b1)==0));};},"=":function(attr,_3b5){return function(elem){return (_39b(elem,attr)==_3b5);};}};var _3b7=(typeof _36f().firstChild.nextElementSibling=="undefined");var _ns=!_3b7?"nextElementSibling":"nextSibling";var _ps=!_3b7?"previousElementSibling":"previousSibling";var _3ba=(_3b7?_398:_374);var _3bb=function(node){while(node=node[_ps]){if(_3ba(node)){return false;}}return true;};var _3bd=function(node){while(node=node[_ns]){if(_3ba(node)){return false;}}return true;};var _3bf=function(node){var root=node.parentNode;var i=0,tret=root[_371],ci=(node["_i"]||-1),cl=(root["_l"]||-1);if(!tret){return -1;}var l=tret.length;if(cl==l&&ci>=0&&cl>=0){return ci;}root["_l"]=l;ci=-1;for(var te=root["firstElementChild"]||root["firstChild"];te;te=te[_ns]){if(_3ba(te)){te["_i"]=++i;if(node===te){ci=i;}}}return ci;};var _3c8=function(elem){return !((_3bf(elem))%2);};var _3ca=function(elem){return ((_3bf(elem))%2);};var _3cc={"checked":function(name,_3ce){return function(elem){return !!d.attr(elem,"checked");};},"first-child":function(){return _3bb;},"last-child":function(){return _3bd;},"only-child":function(name,_3d1){return function(node){if(!_3bb(node)){return false;}if(!_3bd(node)){return false;}return true;};},"empty":function(name,_3d4){return function(elem){var cn=elem.childNodes;var cnl=elem.childNodes.length;for(var x=cnl-1;x>=0;x--){var nt=cn[x].nodeType;if((nt===1)||(nt==3)){return false;}}return true;};},"contains":function(name,_3db){var cz=_3db.charAt(0);if(cz=="\""||cz=="'"){_3db=_3db.slice(1,-1);}return function(elem){return (elem.innerHTML.indexOf(_3db)>=0);};},"not":function(name,_3df){var p=_375(_3df)[0];var _3e1={el:1};if(p.tag!="*"){_3e1.tag=1;}if(!p.classes.length){_3e1.classes=1;}var ntf=_3e3(p,_3e1);return function(elem){return (!ntf(elem));};},"nth-child":function(name,_3e6){var pi=parseInt;if(_3e6=="odd"){return _3ca;}else{if(_3e6=="even"){return _3c8;}}if(_3e6.indexOf("n")!=-1){var _3e8=_3e6.split("n",2);var pred=_3e8[0]?((_3e8[0]=="-")?-1:pi(_3e8[0])):1;var idx=_3e8[1]?pi(_3e8[1]):0;var lb=0,ub=-1;if(pred>0){if(idx<0){idx=(idx%pred)&&(pred+(idx%pred));}else{if(idx>0){if(idx>=pred){lb=idx-idx%pred;}idx=idx%pred;}}}else{if(pred<0){pred*=-1;if(idx>0){ub=idx;idx=idx%pred;}}}if(pred>0){return function(elem){var i=_3bf(elem);return (i>=lb)&&(ub<0||i<=ub)&&((i%pred)==idx);};}else{_3e6=idx;}}var _3ef=pi(_3e6);return function(elem){return (_3bf(elem)==_3ef);};}};var _3f1=(d.isIE)?function(cond){var clc=cond.toLowerCase();if(clc=="class"){cond="className";}return function(elem){return (_373?elem.getAttribute(cond):elem[cond]||elem[clc]);};}:function(cond){return function(elem){return (elem&&elem.getAttribute&&elem.hasAttribute(cond));};};var _3e3=function(_3f7,_3f8){if(!_3f7){return _374;}_3f8=_3f8||{};var ff=null;if(!("el" in _3f8)){ff=_391(ff,_398);}if(!("tag" in _3f8)){if(_3f7.tag!="*"){ff=_391(ff,function(elem){return (elem&&(elem.tagName==_3f7.getTag()));});}}if(!("classes" in _3f8)){each(_3f7.classes,function(_3fb,idx,arr){var re=new RegExp("(?:^|\\s)"+_3fb+"(?:\\s|$)");ff=_391(ff,function(elem){return re.test(elem.className);});ff.count=idx;});}if(!("pseudos" in _3f8)){each(_3f7.pseudos,function(_400){var pn=_400.name;if(_3cc[pn]){ff=_391(ff,_3cc[pn](pn,_400.value));}});}if(!("attrs" in _3f8)){each(_3f7.attrs,function(attr){var _403;var a=attr.attr;if(attr.type&&_39e[attr.type]){_403=_39e[attr.type](a,attr.matchFor);}else{if(a.length){_403=_3f1(a);}}if(_403){ff=_391(ff,_403);}});}if(!("id" in _3f8)){if(_3f7.id){ff=_391(ff,function(elem){return (!!elem&&(elem.id==_3f7.id));});}}if(!ff){if(!("default" in _3f8)){ff=_374;}}return ff;};var _406=function(_407){return function(node,ret,bag){while(node=node[_ns]){if(_3b7&&(!_398(node))){continue;}if((!bag||_40b(node,bag))&&_407(node)){ret.push(node);}break;}return ret;};};var _40c=function(_40d){return function(root,ret,bag){var te=root[_ns];while(te){if(_3ba(te)){if(bag&&!_40b(te,bag)){break;}if(_40d(te)){ret.push(te);}}te=te[_ns];}return ret;};};var _412=function(_413){_413=_413||_374;return function(root,ret,bag){var te,x=0,tret=root[_371];while(te=tret[x++]){if(_3ba(te)&&(!bag||_40b(te,bag))&&(_413(te,x))){ret.push(te);}}return ret;};};var _41a=function(node,root){var pn=node.parentNode;while(pn){if(pn==root){break;}pn=pn.parentNode;}return !!pn;};var _41e={};var _41f=function(_420){var _421=_41e[_420.query];if(_421){return _421;}var io=_420.infixOper;var oper=(io?io.oper:"");var _424=_3e3(_420,{el:1});var qt=_420.tag;var _426=("*"==qt);var ecs=_36f()["getElementsByClassName"];if(!oper){if(_420.id){_424=(!_420.loops&&_426)?_374:_3e3(_420,{el:1,id:1});_421=function(root,arr){var te=d.byId(_420.id,(root.ownerDocument||root));if(!te||!_424(te)){return;}if(9==root.nodeType){return _394(te,arr);}else{if(_41a(te,root)){return _394(te,arr);}}};}else{if(ecs&&/\{\s*\[native code\]\s*\}/.test(String(ecs))&&_420.classes.length&&!_370){_424=_3e3(_420,{el:1,classes:1,id:1});var _42b=_420.classes.join(" ");_421=function(root,arr,bag){var ret=_394(0,arr),te,x=0;var tret=root.getElementsByClassName(_42b);while((te=tret[x++])){if(_424(te,root)&&_40b(te,bag)){ret.push(te);}}return ret;};}else{if(!_426&&!_420.loops){_421=function(root,arr,bag){var ret=_394(0,arr),te,x=0;var tret=root.getElementsByTagName(_420.getTag());while((te=tret[x++])){if(_40b(te,bag)){ret.push(te);}}return ret;};}else{_424=_3e3(_420,{el:1,tag:1,id:1});_421=function(root,arr,bag){var ret=_394(0,arr),te,x=0;var tret=root.getElementsByTagName(_420.getTag());while((te=tret[x++])){if(_424(te,root)&&_40b(te,bag)){ret.push(te);}}return ret;};}}}}else{var _441={el:1};if(_426){_441.tag=1;}_424=_3e3(_420,_441);if("+"==oper){_421=_406(_424);}else{if("~"==oper){_421=_40c(_424);}else{if(">"==oper){_421=_412(_424);}}}}return _41e[_420.query]=_421;};var _442=function(root,_444){var _445=_394(root),qp,x,te,qpl=_444.length,bag,ret;for(var i=0;i<qpl;i++){ret=[];qp=_444[i];x=_445.length-1;if(x>0){bag={};ret.nozip=true;}var gef=_41f(qp);while(te=_445[x--]){gef(te,ret,bag);}if(!ret.length){break;}_445=ret;}return ret;};var _44e={},_44f={};var _450=function(_451){var _452=_375(trim(_451));if(_452.length==1){var tef=_41f(_452[0]);return function(root){var r=tef(root,new qlc());if(r){r.nozip=true;}return r;};}return function(root){return _442(root,_452);};};var nua=navigator.userAgent;var wk="WebKit/";var _459=(d.isWebKit&&(nua.indexOf(wk)>0)&&(parseFloat(nua.split(wk)[1])>528));var _45a=d.isIE?"commentStrip":"nozip";var qsa="querySelectorAll";var _45c=(!!_36f()[qsa]&&(!d.isSafari||(d.isSafari>3.1)||_459));var _45d=function(_45e,_45f){if(_45c){var _460=_44f[_45e];if(_460&&!_45f){return _460;}}var _461=_44e[_45e];if(_461){return _461;}var qcz=_45e.charAt(0);var _463=(-1==_45e.indexOf(" "));if((_45e.indexOf("#")>=0)&&(_463)){_45f=true;}var _464=(_45c&&(!_45f)&&(_372.indexOf(qcz)==-1)&&(!d.isIE||(_45e.indexOf(":")==-1))&&(!(_370&&(_45e.indexOf(".")>=0)))&&(_45e.indexOf(":contains")==-1)&&(_45e.indexOf("|=")==-1));if(_464){var tq=(_372.indexOf(_45e.charAt(_45e.length-1))>=0)?(_45e+" *"):_45e;return _44f[_45e]=function(root){try{if(!((9==root.nodeType)||_463)){throw "";}var r=root[qsa](tq);r[_45a]=true;return r;}catch(e){return _45d(_45e,true)(root);}};}else{var _468=_45e.split(/\s*,\s*/);return _44e[_45e]=((_468.length<2)?_450(_45e):function(root){var _46a=0,ret=[],tp;while((tp=_468[_46a++])){ret=ret.concat(_450(tp)(root));}return ret;});}};var _46d=0;var _46e=d.isIE?function(node){if(_373){return (node.getAttribute("_uid")||node.setAttribute("_uid",++_46d)||_46d);}else{return node.uniqueID;}}:function(node){return (node._uid||(node._uid=++_46d));};var _40b=function(node,bag){if(!bag){return 1;}var id=_46e(node);if(!bag[id]){return bag[id]=1;}return 0;};var _474="_zipIdx";var _zip=function(arr){if(arr&&arr.nozip){return (qlc._wrap)?qlc._wrap(arr):arr;}var ret=new qlc();if(!arr||!arr.length){return ret;}if(arr[0]){ret.push(arr[0]);}if(arr.length<2){return ret;}_46d++;if(d.isIE&&_373){var _478=_46d+"";arr[0].setAttribute(_474,_478);for(var x=1,te;te=arr[x];x++){if(arr[x].getAttribute(_474)!=_478){ret.push(te);}te.setAttribute(_474,_478);}}else{if(d.isIE&&arr.commentStrip){try{for(var x=1,te;te=arr[x];x++){if(_398(te)){ret.push(te);}}}catch(e){}}else{if(arr[0]){arr[0][_474]=_46d;}for(var x=1,te;te=arr[x];x++){if(arr[x][_474]!=_46d){ret.push(te);}te[_474]=_46d;}}}return ret;};d.query=function(_47b,root){qlc=d._queryListCtor;if(!_47b){return new qlc();}if(_47b.constructor==qlc){return _47b;}if(!_36e(_47b)){return new qlc(_47b);}if(_36e(root)){root=d.byId(root);if(!root){return new qlc();}}root=root||_36f();var od=root.ownerDocument||root.documentElement;_373=(root.contentType&&root.contentType=="application/xml")||(d.isOpera&&(root.doctype||od.toString()=="[object XMLDocument]"))||(!!od)&&(d.isIE?od.xml:(root.xmlVersion||od.xmlVersion));var r=_45d(_47b)(root);if(r&&r.nozip&&!qlc._wrap){return r;}return _zip(r);};d.query.pseudos=_3cc;d._filterQueryResult=function(_47f,_480){var _481=new d._queryListCtor();var _482=_3e3(_375(_480)[0]);for(var x=0,te;te=_47f[x];x++){if(_482(te)){_481.push(te);}}return _481;};})(this["queryPortability"]||this["acme"]||dojo);}if(!dojo._hasResource["dojo._base.xhr"]){dojo._hasResource["dojo._base.xhr"]=true;dojo.provide("dojo._base.xhr");(function(){var _d=dojo;function setValue(obj,name,_488){var val=obj[name];if(_d.isString(val)){obj[name]=[val,_488];}else{if(_d.isArray(val)){val.push(_488);}else{obj[name]=_488;}}};dojo.formToObject=function(_48a){var ret={};var _48c="file|submit|image|reset|button|";_d.forEach(dojo.byId(_48a).elements,function(item){var _in=item.name;var type=(item.type||"").toLowerCase();if(_in&&type&&_48c.indexOf(type)==-1&&!item.disabled){if(type=="radio"||type=="checkbox"){if(item.checked){setValue(ret,_in,item.value);}}else{if(item.multiple){ret[_in]=[];_d.query("option",item).forEach(function(opt){if(opt.selected){setValue(ret,_in,opt.value);}});}else{setValue(ret,_in,item.value);if(type=="image"){ret[_in+".x"]=ret[_in+".y"]=ret[_in].x=ret[_in].y=0;}}}}});return ret;};dojo.objectToQuery=function(map){var enc=encodeURIComponent;var _493=[];var _494={};for(var name in map){var _496=map[name];if(_496!=_494[name]){var _497=enc(name)+"=";if(_d.isArray(_496)){for(var i=0;i<_496.length;i++){_493.push(_497+enc(_496[i]));}}else{_493.push(_497+enc(_496));}}}return _493.join("&");};dojo.formToQuery=function(_499){return _d.objectToQuery(_d.formToObject(_499));};dojo.formToJson=function(_49a,_49b){return _d.toJson(_d.formToObject(_49a),_49b);};dojo.queryToObject=function(str){var ret={};var qp=str.split("&");var dec=decodeURIComponent;_d.forEach(qp,function(item){if(item.length){var _4a1=item.split("=");var name=dec(_4a1.shift());var val=dec(_4a1.join("="));if(_d.isString(ret[name])){ret[name]=[ret[name]];}if(_d.isArray(ret[name])){ret[name].push(val);}else{ret[name]=val;}}});return ret;};dojo._blockAsync=false;dojo._contentHandlers={text:function(xhr){return xhr.responseText;},json:function(xhr){return _d.fromJson(xhr.responseText||null);},"json-comment-filtered":function(xhr){if(!dojo.config.useCommentedJson){console.warn("Consider using the standard mimetype:application/json."+" json-commenting can introduce security issues. To"+" decrease the chances of hijacking, use the standard the 'json' handler and"+" prefix your json with: {}&&\n"+"Use djConfig.useCommentedJson=true to turn off this message.");}var _4a7=xhr.responseText;var _4a8=_4a7.indexOf("/*");var _4a9=_4a7.lastIndexOf("*/");if(_4a8==-1||_4a9==-1){throw new Error("JSON was not comment filtered");}return _d.fromJson(_4a7.substring(_4a8+2,_4a9));},javascript:function(xhr){return _d.eval(xhr.responseText);},xml:function(xhr){var _4ac=xhr.responseXML;if(_d.isIE&&(!_4ac||!_4ac.documentElement)){var ms=function(n){return "MSXML"+n+".DOMDocument";};var dp=["Microsoft.XMLDOM",ms(6),ms(4),ms(3),ms(2)];_d.some(dp,function(p){try{var dom=new ActiveXObject(p);dom.async=false;dom.loadXML(xhr.responseText);_4ac=dom;}catch(e){return false;}return true;});}return _4ac;}};dojo._contentHandlers["json-comment-optional"]=function(xhr){var _4b3=_d._contentHandlers;if(xhr.responseText&&xhr.responseText.indexOf("/*")!=-1){return _4b3["json-comment-filtered"](xhr);}else{return _4b3["json"](xhr);}};dojo._ioSetArgs=function(args,_4b5,_4b6,_4b7){var _4b8={args:args,url:args.url};var _4b9=null;if(args.form){var form=_d.byId(args.form);var _4bb=form.getAttributeNode("action");_4b8.url=_4b8.url||(_4bb?_4bb.value:null);_4b9=_d.formToObject(form);}var _4bc=[{}];if(_4b9){_4bc.push(_4b9);}if(args.content){_4bc.push(args.content);}if(args.preventCache){_4bc.push({"dojo.preventCache":new Date().valueOf()});}_4b8.query=_d.objectToQuery(_d.mixin.apply(null,_4bc));_4b8.handleAs=args.handleAs||"text";var d=new _d.Deferred(_4b5);d.addCallbacks(_4b6,function(_4be){return _4b7(_4be,d);});var ld=args.load;if(ld&&_d.isFunction(ld)){d.addCallback(function(_4c0){return ld.call(args,_4c0,_4b8);});}var err=args.error;if(err&&_d.isFunction(err)){d.addErrback(function(_4c2){return err.call(args,_4c2,_4b8);});}var _4c3=args.handle;if(_4c3&&_d.isFunction(_4c3)){d.addBoth(function(_4c4){return _4c3.call(args,_4c4,_4b8);});}d.ioArgs=_4b8;return d;};var _4c5=function(dfd){dfd.canceled=true;var xhr=dfd.ioArgs.xhr;var _at=typeof xhr.abort;if(_at=="function"||_at=="object"||_at=="unknown"){xhr.abort();}var err=dfd.ioArgs.error;if(!err){err=new Error("xhr cancelled");err.dojoType="cancel";}return err;};var _4ca=function(dfd){var ret=_d._contentHandlers[dfd.ioArgs.handleAs](dfd.ioArgs.xhr);return ret===undefined?null:ret;};var _4cd=function(_4ce,dfd){console.error(_4ce);return _4ce;};var _4d0=null;var _4d1=[];var _4d2=function(){var now=(new Date()).getTime();if(!_d._blockAsync){for(var i=0,tif;i<_4d1.length&&(tif=_4d1[i]);i++){var dfd=tif.dfd;var func=function(){if(!dfd||dfd.canceled||!tif.validCheck(dfd)){_4d1.splice(i--,1);}else{if(tif.ioCheck(dfd)){_4d1.splice(i--,1);tif.resHandle(dfd);}else{if(dfd.startTime){if(dfd.startTime+(dfd.ioArgs.args.timeout||0)<now){_4d1.splice(i--,1);var err=new Error("timeout exceeded");err.dojoType="timeout";dfd.errback(err);dfd.cancel();}}}}};if(dojo.config.debugAtAllCosts){func.call(this);}else{try{func.call(this);}catch(e){dfd.errback(e);}}}}if(!_4d1.length){clearInterval(_4d0);_4d0=null;return;}};dojo._ioCancelAll=function(){try{_d.forEach(_4d1,function(i){try{i.dfd.cancel();}catch(e){}});}catch(e){}};if(_d.isIE){_d.addOnWindowUnload(_d._ioCancelAll);}_d._ioWatch=function(dfd,_4db,_4dc,_4dd){var args=dfd.ioArgs.args;if(args.timeout){dfd.startTime=(new Date()).getTime();}_4d1.push({dfd:dfd,validCheck:_4db,ioCheck:_4dc,resHandle:_4dd});if(!_4d0){_4d0=setInterval(_4d2,50);}if(args.sync){_4d2();}};var _4df="application/x-www-form-urlencoded";var _4e0=function(dfd){return dfd.ioArgs.xhr.readyState;};var _4e2=function(dfd){return 4==dfd.ioArgs.xhr.readyState;};var _4e4=function(dfd){var xhr=dfd.ioArgs.xhr;if(_d._isDocumentOk(xhr)){dfd.callback(dfd);}else{var err=new Error("Unable to load "+dfd.ioArgs.url+" status:"+xhr.status);err.status=xhr.status;err.responseText=xhr.responseText;dfd.errback(err);}};dojo._ioAddQueryToUrl=function(_4e8){if(_4e8.query.length){_4e8.url+=(_4e8.url.indexOf("?")==-1?"?":"&")+_4e8.query;_4e8.query=null;}};dojo.xhr=function(_4e9,args,_4eb){var dfd=_d._ioSetArgs(args,_4c5,_4ca,_4cd);dfd.ioArgs.xhr=_d._xhrObj(dfd.ioArgs.args);if(_4eb){if("postData" in args){dfd.ioArgs.query=args.postData;}else{if("putData" in args){dfd.ioArgs.query=args.putData;}}}else{_d._ioAddQueryToUrl(dfd.ioArgs);}var _4ed=dfd.ioArgs;var xhr=_4ed.xhr;xhr.open(_4e9,_4ed.url,args.sync!==true,args.user||undefined,args.password||undefined);if(args.headers){for(var hdr in args.headers){if(hdr.toLowerCase()==="content-type"&&!args.contentType){args.contentType=args.headers[hdr];}else{xhr.setRequestHeader(hdr,args.headers[hdr]);}}}xhr.setRequestHeader("Content-Type",args.contentType||_4df);if(!args.headers||!args.headers["X-Requested-With"]){xhr.setRequestHeader("X-Requested-With","XMLHttpRequest");}if(dojo.config.debugAtAllCosts){xhr.send(_4ed.query);}else{try{xhr.send(_4ed.query);}catch(e){dfd.ioArgs.error=e;dfd.cancel();}}_d._ioWatch(dfd,_4e0,_4e2,_4e4);xhr=null;return dfd;};dojo.xhrGet=function(args){return _d.xhr("GET",args);};dojo.rawXhrPost=dojo.xhrPost=function(args){return _d.xhr("POST",args,true);};dojo.rawXhrPut=dojo.xhrPut=function(args){return _d.xhr("PUT",args,true);};dojo.xhrDelete=function(args){return _d.xhr("DELETE",args);};})();}if(!dojo._hasResource["dojo._base.fx"]){dojo._hasResource["dojo._base.fx"]=true;dojo.provide("dojo._base.fx");(function(){var d=dojo;var _4f5=d.mixin;dojo._Line=function(_4f6,end){this.start=_4f6;this.end=end;};dojo._Line.prototype.getValue=function(n){return ((this.end-this.start)*n)+this.start;};d.declare("dojo._Animation",null,{constructor:function(args){_4f5(this,args);if(d.isArray(this.curve)){this.curve=new d._Line(this.curve[0],this.curve[1]);}},duration:350,repeat:0,rate:10,_percent:0,_startRepeatCount:0,_fire:function(evt,args){if(this[evt]){if(dojo.config.debugAtAllCosts){this[evt].apply(this,args||[]);}else{try{this[evt].apply(this,args||[]);}catch(e){console.error("exception in animation handler for:",evt);console.error(e);}}}return this;},play:function(_4fc,_4fd){var _t=this;if(_t._delayTimer){_t._clearTimer();}if(_4fd){_t._stopTimer();_t._active=_t._paused=false;_t._percent=0;}else{if(_t._active&&!_t._paused){return _t;}}_t._fire("beforeBegin");var de=_4fc||_t.delay,_p=dojo.hitch(_t,"_play",_4fd);if(de>0){_t._delayTimer=setTimeout(_p,de);return _t;}_p();return _t;},_play:function(_501){var _t=this;if(_t._delayTimer){_t._clearTimer();}_t._startTime=new Date().valueOf();if(_t._paused){_t._startTime-=_t.duration*_t._percent;}_t._endTime=_t._startTime+_t.duration;_t._active=true;_t._paused=false;var _503=_t.curve.getValue(_t._percent);if(!_t._percent){if(!_t._startRepeatCount){_t._startRepeatCount=_t.repeat;}_t._fire("onBegin",[_503]);}_t._fire("onPlay",[_503]);_t._cycle();return _t;},pause:function(){var _t=this;if(_t._delayTimer){_t._clearTimer();}_t._stopTimer();if(!_t._active){return _t;}_t._paused=true;_t._fire("onPause",[_t.curve.getValue(_t._percent)]);return _t;},gotoPercent:function(_505,_506){var _t=this;_t._stopTimer();_t._active=_t._paused=true;_t._percent=_505;if(_506){_t.play();}return _t;},stop:function(_508){var _t=this;if(_t._delayTimer){_t._clearTimer();}if(!_t._timer){return _t;}_t._stopTimer();if(_508){_t._percent=1;}_t._fire("onStop",[_t.curve.getValue(_t._percent)]);_t._active=_t._paused=false;return _t;},status:function(){if(this._active){return this._paused?"paused":"playing";}return "stopped";},_cycle:function(){var _t=this;if(_t._active){var curr=new Date().valueOf();var step=(curr-_t._startTime)/(_t._endTime-_t._startTime);if(step>=1){step=1;}_t._percent=step;if(_t.easing){step=_t.easing(step);}_t._fire("onAnimate",[_t.curve.getValue(step)]);if(_t._percent<1){_t._startTimer();}else{_t._active=false;if(_t.repeat>0){_t.repeat--;_t.play(null,true);}else{if(_t.repeat==-1){_t.play(null,true);}else{if(_t._startRepeatCount){_t.repeat=_t._startRepeatCount;_t._startRepeatCount=0;}}}_t._percent=0;_t._fire("onEnd");_t._stopTimer();}}return _t;},_clearTimer:function(){clearTimeout(this._delayTimer);delete this._delayTimer;}});var ctr=0,_50e=[],_50f=null,_510={run:function(){}};dojo._Animation.prototype._startTimer=function(){if(!this._timer){this._timer=d.connect(_510,"run",this,"_cycle");ctr++;}if(!_50f){_50f=setInterval(d.hitch(_510,"run"),this.rate);}};dojo._Animation.prototype._stopTimer=function(){if(this._timer){d.disconnect(this._timer);this._timer=null;ctr--;}if(ctr<=0){clearInterval(_50f);_50f=null;ctr=0;}};var _511=d.isIE?function(node){var ns=node.style;if(!ns.width.length&&d.style(node,"width")=="auto"){ns.width="auto";}}:function(){};dojo._fade=function(args){args.node=d.byId(args.node);var _515=_4f5({properties:{}},args),_516=(_515.properties.opacity={});_516.start=!("start" in _515)?function(){return +d.style(_515.node,"opacity")||0;}:_515.start;_516.end=_515.end;var anim=d.animateProperty(_515);d.connect(anim,"beforeBegin",d.partial(_511,_515.node));return anim;};dojo.fadeIn=function(args){return d._fade(_4f5({end:1},args));};dojo.fadeOut=function(args){return d._fade(_4f5({end:0},args));};dojo._defaultEasing=function(n){return 0.5+((Math.sin((n+1.5)*Math.PI))/2);};var _51b=function(_51c){this._properties=_51c;for(var p in _51c){var prop=_51c[p];if(prop.start instanceof d.Color){prop.tempColor=new d.Color();}}};_51b.prototype.getValue=function(r){var ret={};for(var p in this._properties){var prop=this._properties[p],_523=prop.start;if(_523 instanceof d.Color){ret[p]=d.blendColors(_523,prop.end,r,prop.tempColor).toCss();}else{if(!d.isArray(_523)){ret[p]=((prop.end-_523)*r)+_523+(p!="opacity"?prop.units||"px":0);}}}return ret;};dojo.animateProperty=function(args){args.node=d.byId(args.node);if(!args.easing){args.easing=d._defaultEasing;}var anim=new d._Animation(args);d.connect(anim,"beforeBegin",anim,function(){var pm={};for(var p in this.properties){if(p=="width"||p=="height"){this.node.display="block";}var prop=this.properties[p];prop=pm[p]=_4f5({},(d.isObject(prop)?prop:{end:prop}));if(d.isFunction(prop.start)){prop.start=prop.start();}if(d.isFunction(prop.end)){prop.end=prop.end();}var _529=(p.toLowerCase().indexOf("color")>=0);function getStyle(node,p){var v={height:node.offsetHeight,width:node.offsetWidth}[p];if(v!==undefined){return v;}v=d.style(node,p);return (p=="opacity")?+v:(_529?v:parseFloat(v));};if(!("end" in prop)){prop.end=getStyle(this.node,p);}else{if(!("start" in prop)){prop.start=getStyle(this.node,p);}}if(_529){prop.start=new d.Color(prop.start);prop.end=new d.Color(prop.end);}else{prop.start=(p=="opacity")?+prop.start:parseFloat(prop.start);}}this.curve=new _51b(pm);});d.connect(anim,"onAnimate",d.hitch(d,"style",anim.node));return anim;};dojo.anim=function(node,_52e,_52f,_530,_531,_532){return d.animateProperty({node:node,duration:_52f||d._Animation.prototype.duration,properties:_52e,easing:_530,onEnd:_531}).play(_532||0);};})();}if(!dojo._hasResource["dojo._base.browser"]){dojo._hasResource["dojo._base.browser"]=true;dojo.provide("dojo._base.browser");dojo.forEach(dojo.config.require,function(i){dojo["require"](i);});}if(!dojo._hasResource["dojo.date"]){dojo._hasResource["dojo.date"]=true;dojo.provide("dojo.date");dojo.date.getDaysInMonth=function(_534){var _535=_534.getMonth();var days=[31,28,31,30,31,30,31,31,30,31,30,31];if(_535==1&&dojo.date.isLeapYear(_534)){return 29;}return days[_535];};dojo.date.isLeapYear=function(_537){var year=_537.getFullYear();return !(year%400)||(!(year%4)&&!!(year%100));};dojo.date.getTimezoneName=function(_539){var str=_539.toString();var tz="";var _53c;var pos=str.indexOf("(");if(pos>-1){tz=str.substring(++pos,str.indexOf(")"));}else{var pat=/([A-Z\/]+) \d{4}$/;if((_53c=str.match(pat))){tz=_53c[1];}else{str=_539.toLocaleString();pat=/ ([A-Z\/]+)$/;if((_53c=str.match(pat))){tz=_53c[1];}}}return (tz=="AM"||tz=="PM")?"":tz;};dojo.date.compare=function(_53f,_540,_541){_53f=new Date(Number(_53f));_540=new Date(Number(_540||new Date()));if(_541!=="undefined"){if(_541=="date"){_53f.setHours(0,0,0,0);_540.setHours(0,0,0,0);}else{if(_541=="time"){_53f.setFullYear(0,0,0);_540.setFullYear(0,0,0);}}}if(_53f>_540){return 1;}if(_53f<_540){return -1;}return 0;};dojo.date.add=function(date,_543,_544){var sum=new Date(Number(date));var _546=false;var _547="Date";switch(_543){case "day":break;case "weekday":var days,_549;var mod=_544%5;if(!mod){days=(_544>0)?5:-5;_549=(_544>0)?((_544-5)/5):((_544+5)/5);}else{days=mod;_549=parseInt(_544/5);}var strt=date.getDay();var adj=0;if(strt==6&&_544>0){adj=1;}else{if(strt==0&&_544<0){adj=-1;}}var trgt=strt+days;if(trgt==0||trgt==6){adj=(_544>0)?2:-2;}_544=(7*_549)+days+adj;break;case "year":_547="FullYear";_546=true;break;case "week":_544*=7;break;case "quarter":_544*=3;case "month":_546=true;_547="Month";break;case "hour":case "minute":case "second":case "millisecond":_547="UTC"+_543.charAt(0).toUpperCase()+_543.substring(1)+"s";}if(_547){sum["set"+_547](sum["get"+_547]()+_544);}if(_546&&(sum.getDate()<date.getDate())){sum.setDate(0);}return sum;};dojo.date.difference=function(_54e,_54f,_550){_54f=_54f||new Date();_550=_550||"day";var _551=_54f.getFullYear()-_54e.getFullYear();var _552=1;switch(_550){case "quarter":var m1=_54e.getMonth();var m2=_54f.getMonth();var q1=Math.floor(m1/3)+1;var q2=Math.floor(m2/3)+1;q2+=(_551*4);_552=q2-q1;break;case "weekday":var days=Math.round(dojo.date.difference(_54e,_54f,"day"));var _558=parseInt(dojo.date.difference(_54e,_54f,"week"));var mod=days%7;if(mod==0){days=_558*5;}else{var adj=0;var aDay=_54e.getDay();var bDay=_54f.getDay();_558=parseInt(days/7);mod=days%7;var _55d=new Date(_54e);_55d.setDate(_55d.getDate()+(_558*7));var _55e=_55d.getDay();if(days>0){switch(true){case aDay==6:adj=-1;break;case aDay==0:adj=0;break;case bDay==6:adj=-1;break;case bDay==0:adj=-2;break;case (_55e+mod)>5:adj=-2;}}else{if(days<0){switch(true){case aDay==6:adj=0;break;case aDay==0:adj=1;break;case bDay==6:adj=2;break;case bDay==0:adj=1;break;case (_55e+mod)<0:adj=2;}}}days+=adj;days-=(_558*2);}_552=days;break;case "year":_552=_551;break;case "month":_552=(_54f.getMonth()-_54e.getMonth())+(_551*12);break;case "week":_552=parseInt(dojo.date.difference(_54e,_54f,"day")/7);break;case "day":_552/=24;case "hour":_552/=60;case "minute":_552/=60;case "second":_552/=1000;case "millisecond":_552*=_54f.getTime()-_54e.getTime();}return Math.round(_552);};}if(!dojo._hasResource["dojo.i18n"]){dojo._hasResource["dojo.i18n"]=true;dojo.provide("dojo.i18n");dojo.i18n.getLocalization=function(_55f,_560,_561){_561=dojo.i18n.normalizeLocale(_561);var _562=_561.split("-");var _563=[_55f,"nls",_560].join(".");var _564=dojo._loadedModules[_563];if(_564){var _565;for(var i=_562.length;i>0;i--){var loc=_562.slice(0,i).join("_");if(_564[loc]){_565=_564[loc];break;}}if(!_565){_565=_564.ROOT;}if(_565){var _568=function(){};_568.prototype=_565;return new _568();}}throw new Error("Bundle not found: "+_560+" in "+_55f+" , locale="+_561);};dojo.i18n.normalizeLocale=function(_569){var _56a=_569?_569.toLowerCase():dojo.locale;if(_56a=="root"){_56a="ROOT";}return _56a;};dojo.i18n._requireLocalization=function(_56b,_56c,_56d,_56e){var _56f=dojo.i18n.normalizeLocale(_56d);var _570=[_56b,"nls",_56c].join(".");var _571="";if(_56e){var _572=_56e.split(",");for(var i=0;i<_572.length;i++){if(_56f["indexOf"](_572[i])==0){if(_572[i].length>_571.length){_571=_572[i];}}}if(!_571){_571="ROOT";}}var _574=_56e?_571:_56f;var _575=dojo._loadedModules[_570];var _576=null;if(_575){if(dojo.config.localizationComplete&&_575._built){return;}var _577=_574.replace(/-/g,"_");var _578=_570+"."+_577;_576=dojo._loadedModules[_578];}if(!_576){_575=dojo["provide"](_570);var syms=dojo._getModuleSymbols(_56b);var _57a=syms.concat("nls").join("/");var _57b;dojo.i18n._searchLocalePath(_574,_56e,function(loc){var _57d=loc.replace(/-/g,"_");var _57e=_570+"."+_57d;var _57f=false;if(!dojo._loadedModules[_57e]){dojo["provide"](_57e);var _580=[_57a];if(loc!="ROOT"){_580.push(loc);}_580.push(_56c);var _581=_580.join("/")+".js";_57f=dojo._loadPath(_581,null,function(hash){var _583=function(){};_583.prototype=_57b;_575[_57d]=new _583();for(var j in hash){_575[_57d][j]=hash[j];}});}else{_57f=true;}if(_57f&&_575[_57d]){_57b=_575[_57d];}else{_575[_57d]=_57b;}if(_56e){return true;}});}if(_56e&&_56f!=_571){_575[_56f.replace(/-/g,"_")]=_575[_571.replace(/-/g,"_")];}};(function(){var _585=dojo.config.extraLocale;if(_585){if(!_585 instanceof Array){_585=[_585];}var req=dojo.i18n._requireLocalization;dojo.i18n._requireLocalization=function(m,b,_589,_58a){req(m,b,_589,_58a);if(_589){return;}for(var i=0;i<_585.length;i++){req(m,b,_585[i],_58a);}};}})();dojo.i18n._searchLocalePath=function(_58c,down,_58e){_58c=dojo.i18n.normalizeLocale(_58c);var _58f=_58c.split("-");var _590=[];for(var i=_58f.length;i>0;i--){_590.push(_58f.slice(0,i).join("-"));}_590.push(false);if(down){_590.reverse();}for(var j=_590.length-1;j>=0;j--){var loc=_590[j]||"ROOT";var stop=_58e(loc);if(stop){break;}}};dojo.i18n._preloadLocalizations=function(_595,_596){function preload(_597){_597=dojo.i18n.normalizeLocale(_597);dojo.i18n._searchLocalePath(_597,true,function(loc){for(var i=0;i<_596.length;i++){if(_596[i]==loc){dojo["require"](_595+"_"+loc);return true;}}return false;});};preload();var _59a=dojo.config.extraLocale||[];for(var i=0;i<_59a.length;i++){preload(_59a[i]);}};}if(!dojo._hasResource["dojo.cldr.supplemental"]){dojo._hasResource["dojo.cldr.supplemental"]=true;dojo.provide("dojo.cldr.supplemental");dojo.cldr.supplemental.getFirstDayOfWeek=function(_59c){var _59d={mv:5,ae:6,af:6,bh:6,dj:6,dz:6,eg:6,er:6,et:6,iq:6,ir:6,jo:6,ke:6,kw:6,lb:6,ly:6,ma:6,om:6,qa:6,sa:6,sd:6,so:6,tn:6,ye:6,as:0,au:0,az:0,bw:0,ca:0,cn:0,fo:0,ge:0,gl:0,gu:0,hk:0,ie:0,il:0,is:0,jm:0,jp:0,kg:0,kr:0,la:0,mh:0,mo:0,mp:0,mt:0,nz:0,ph:0,pk:0,sg:0,th:0,tt:0,tw:0,um:0,us:0,uz:0,vi:0,za:0,zw:0,et:0,mw:0,ng:0,tj:0,sy:4};var _59e=dojo.cldr.supplemental._region(_59c);var dow=_59d[_59e];return (dow===undefined)?1:dow;};dojo.cldr.supplemental._region=function(_5a0){_5a0=dojo.i18n.normalizeLocale(_5a0);var tags=_5a0.split("-");var _5a2=tags[1];if(!_5a2){_5a2={de:"de",en:"us",es:"es",fi:"fi",fr:"fr",he:"il",hu:"hu",it:"it",ja:"jp",ko:"kr",nl:"nl",pt:"br",sv:"se",zh:"cn"}[tags[0]];}else{if(_5a2.length==4){_5a2=tags[2];}}return _5a2;};dojo.cldr.supplemental.getWeekend=function(_5a3){var _5a4={eg:5,il:5,sy:5,"in":0,ae:4,bh:4,dz:4,iq:4,jo:4,kw:4,lb:4,ly:4,ma:4,om:4,qa:4,sa:4,sd:4,tn:4,ye:4};var _5a5={ae:5,bh:5,dz:5,iq:5,jo:5,kw:5,lb:5,ly:5,ma:5,om:5,qa:5,sa:5,sd:5,tn:5,ye:5,af:5,ir:5,eg:6,il:6,sy:6};var _5a6=dojo.cldr.supplemental._region(_5a3);var _5a7=_5a4[_5a6];var end=_5a5[_5a6];if(_5a7===undefined){_5a7=6;}if(end===undefined){end=0;}return {start:_5a7,end:end};};}if(!dojo._hasResource["dojo.regexp"]){dojo._hasResource["dojo.regexp"]=true;dojo.provide("dojo.regexp");dojo.regexp.escapeString=function(str,_5aa){return str.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g,function(ch){if(_5aa&&_5aa.indexOf(ch)!=-1){return ch;}return "\\"+ch;});};dojo.regexp.buildGroupRE=function(arr,re,_5ae){if(!(arr instanceof Array)){return re(arr);}var b=[];for(var i=0;i<arr.length;i++){b.push(re(arr[i]));}return dojo.regexp.group(b.join("|"),_5ae);};dojo.regexp.group=function(_5b1,_5b2){return "("+(_5b2?"?:":"")+_5b1+")";};}if(!dojo._hasResource["dojo.string"]){dojo._hasResource["dojo.string"]=true;dojo.provide("dojo.string");dojo.string.rep=function(str,num){if(num<=0||!str){return "";}var buf=[];for(;;){if(num&1){buf.push(str);}if(!(num>>=1)){break;}str+=str;}return buf.join("");};dojo.string.pad=function(text,size,ch,end){if(!ch){ch="0";}var out=String(text),pad=dojo.string.rep(ch,Math.ceil((size-out.length)/ch.length));return end?out+pad:pad+out;};dojo.string.substitute=function(_5bc,map,_5be,_5bf){_5bf=_5bf||dojo.global;_5be=(!_5be)?function(v){return v;}:dojo.hitch(_5bf,_5be);return _5bc.replace(/\$\{([^\s\:\}]+)(?:\:([^\s\:\}]+))?\}/g,function(_5c1,key,_5c3){var _5c4=dojo.getObject(key,false,map);if(_5c3){_5c4=dojo.getObject(_5c3,false,_5bf).call(_5bf,_5c4,key);}return _5be(_5c4,key).toString();});};dojo.string.trim=String.prototype.trim?dojo.trim:function(str){str=str.replace(/^\s+/,"");for(var i=str.length-1;i>=0;i--){if(/\S/.test(str.charAt(i))){str=str.substring(0,i+1);break;}}return str;};}if(!dojo._hasResource["dojo.date.locale"]){dojo._hasResource["dojo.date.locale"]=true;dojo.provide("dojo.date.locale");(function(){function formatPattern(_5c7,_5c8,_5c9,_5ca){return _5ca.replace(/([a-z])\1*/ig,function(_5cb){var s,pad;var c=_5cb.charAt(0);var l=_5cb.length;var _5d0=["abbr","wide","narrow"];switch(c){case "G":s=_5c8[(l<4)?"eraAbbr":"eraNames"][_5c7.getFullYear()<0?0:1];break;case "y":s=_5c7.getFullYear();switch(l){case 1:break;case 2:if(!_5c9){s=String(s);s=s.substr(s.length-2);break;}default:pad=true;}break;case "Q":case "q":s=Math.ceil((_5c7.getMonth()+1)/3);pad=true;break;case "M":var m=_5c7.getMonth();if(l<3){s=m+1;pad=true;}else{var _5d2=["months","format",_5d0[l-3]].join("-");s=_5c8[_5d2][m];}break;case "w":var _5d3=0;s=dojo.date.locale._getWeekOfYear(_5c7,_5d3);pad=true;break;case "d":s=_5c7.getDate();pad=true;break;case "D":s=dojo.date.locale._getDayOfYear(_5c7);pad=true;break;case "E":var d=_5c7.getDay();if(l<3){s=d+1;pad=true;}else{var _5d5=["days","format",_5d0[l-3]].join("-");s=_5c8[_5d5][d];}break;case "a":var _5d6=(_5c7.getHours()<12)?"am":"pm";s=_5c8[_5d6];break;case "h":case "H":case "K":case "k":var h=_5c7.getHours();switch(c){case "h":s=(h%12)||12;break;case "H":s=h;break;case "K":s=(h%12);break;case "k":s=h||24;break;}pad=true;break;case "m":s=_5c7.getMinutes();pad=true;break;case "s":s=_5c7.getSeconds();pad=true;break;case "S":s=Math.round(_5c7.getMilliseconds()*Math.pow(10,l-3));pad=true;break;case "v":case "z":s=dojo.date.getTimezoneName(_5c7);if(s){break;}l=4;case "Z":var _5d8=_5c7.getTimezoneOffset();var tz=[(_5d8<=0?"+":"-"),dojo.string.pad(Math.floor(Math.abs(_5d8)/60),2),dojo.string.pad(Math.abs(_5d8)%60,2)];if(l==4){tz.splice(0,0,"GMT");tz.splice(3,0,":");}s=tz.join("");break;default:throw new Error("dojo.date.locale.format: invalid pattern char: "+_5ca);}if(pad){s=dojo.string.pad(s,l);}return s;});};dojo.date.locale.format=function(_5da,_5db){_5db=_5db||{};var _5dc=dojo.i18n.normalizeLocale(_5db.locale);var _5dd=_5db.formatLength||"short";var _5de=dojo.date.locale._getGregorianBundle(_5dc);var str=[];var _5e0=dojo.hitch(this,formatPattern,_5da,_5de,_5db.fullYear);if(_5db.selector=="year"){var year=_5da.getFullYear();if(_5dc.match(/^zh|^ja/)){year+="年";}return year;}if(_5db.selector!="time"){var _5e2=_5db.datePattern||_5de["dateFormat-"+_5dd];if(_5e2){str.push(_processPattern(_5e2,_5e0));}}if(_5db.selector!="date"){var _5e3=_5db.timePattern||_5de["timeFormat-"+_5dd];if(_5e3){str.push(_processPattern(_5e3,_5e0));}}var _5e4=str.join(" ");return _5e4;};dojo.date.locale.regexp=function(_5e5){return dojo.date.locale._parseInfo(_5e5).regexp;};dojo.date.locale._parseInfo=function(_5e6){_5e6=_5e6||{};var _5e7=dojo.i18n.normalizeLocale(_5e6.locale);var _5e8=dojo.date.locale._getGregorianBundle(_5e7);var _5e9=_5e6.formatLength||"short";var _5ea=_5e6.datePattern||_5e8["dateFormat-"+_5e9];var _5eb=_5e6.timePattern||_5e8["timeFormat-"+_5e9];var _5ec;if(_5e6.selector=="date"){_5ec=_5ea;}else{if(_5e6.selector=="time"){_5ec=_5eb;}else{_5ec=_5ea+" "+_5eb;}}var _5ed=[];var re=_processPattern(_5ec,dojo.hitch(this,_buildDateTimeRE,_5ed,_5e8,_5e6));return {regexp:re,tokens:_5ed,bundle:_5e8};};dojo.date.locale.parse=function(_5ef,_5f0){var info=dojo.date.locale._parseInfo(_5f0);var _5f2=info.tokens,_5f3=info.bundle;var re=new RegExp("^"+info.regexp+"$",info.strict?"":"i");var _5f5=re.exec(_5ef);if(!_5f5){return null;}var _5f6=["abbr","wide","narrow"];var _5f7=[1970,0,1,0,0,0,0];var amPm="";var _5f9=dojo.every(_5f5,function(v,i){if(!i){return true;}var _5fc=_5f2[i-1];var l=_5fc.length;switch(_5fc.charAt(0)){case "y":if(l!=2&&_5f0.strict){_5f7[0]=v;}else{if(v<100){v=Number(v);var year=""+new Date().getFullYear();var _5ff=year.substring(0,2)*100;var _600=Math.min(Number(year.substring(2,4))+20,99);var num=(v<_600)?_5ff+v:_5ff-100+v;_5f7[0]=num;}else{if(_5f0.strict){return false;}_5f7[0]=v;}}break;case "M":if(l>2){var _602=_5f3["months-format-"+_5f6[l-3]].concat();if(!_5f0.strict){v=v.replace(".","").toLowerCase();_602=dojo.map(_602,function(s){return s.replace(".","").toLowerCase();});}v=dojo.indexOf(_602,v);if(v==-1){return false;}}else{v--;}_5f7[1]=v;break;case "E":case "e":var days=_5f3["days-format-"+_5f6[l-3]].concat();if(!_5f0.strict){v=v.toLowerCase();days=dojo.map(days,function(d){return d.toLowerCase();});}v=dojo.indexOf(days,v);if(v==-1){return false;}break;case "D":_5f7[1]=0;case "d":_5f7[2]=v;break;case "a":var am=_5f0.am||_5f3.am;var pm=_5f0.pm||_5f3.pm;if(!_5f0.strict){var _608=/\./g;v=v.replace(_608,"").toLowerCase();am=am.replace(_608,"").toLowerCase();pm=pm.replace(_608,"").toLowerCase();}if(_5f0.strict&&v!=am&&v!=pm){return false;}amPm=(v==pm)?"p":(v==am)?"a":"";break;case "K":if(v==24){v=0;}case "h":case "H":case "k":if(v>23){return false;}_5f7[3]=v;break;case "m":_5f7[4]=v;break;case "s":_5f7[5]=v;break;case "S":_5f7[6]=v;}return true;});var _609=+_5f7[3];if(amPm==="p"&&_609<12){_5f7[3]=_609+12;}else{if(amPm==="a"&&_609==12){_5f7[3]=0;}}var _60a=new Date(_5f7[0],_5f7[1],_5f7[2],_5f7[3],_5f7[4],_5f7[5],_5f7[6]);if(_5f0.strict){_60a.setFullYear(_5f7[0]);}var _60b=_5f2.join(""),_60c=_60b.indexOf("d")!=-1,_60d=_60b.indexOf("M")!=-1;if(!_5f9||(_60d&&_60a.getMonth()>_5f7[1])||(_60c&&_60a.getDate()>_5f7[2])){return null;}if((_60d&&_60a.getMonth()<_5f7[1])||(_60c&&_60a.getDate()<_5f7[2])){_60a=dojo.date.add(_60a,"hour",1);}return _60a;};function _processPattern(_60e,_60f,_610,_611){var _612=function(x){return x;};_60f=_60f||_612;_610=_610||_612;_611=_611||_612;var _614=_60e.match(/(''|[^'])+/g);var _615=_60e.charAt(0)=="'";dojo.forEach(_614,function(_616,i){if(!_616){_614[i]="";}else{_614[i]=(_615?_610:_60f)(_616);_615=!_615;}});return _611(_614.join(""));};function _buildDateTimeRE(_618,_619,_61a,_61b){_61b=dojo.regexp.escapeString(_61b);if(!_61a.strict){_61b=_61b.replace(" a"," ?a");}return _61b.replace(/([a-z])\1*/ig,function(_61c){var s;var c=_61c.charAt(0);var l=_61c.length;var p2="",p3="";if(_61a.strict){if(l>1){p2="0"+"{"+(l-1)+"}";}if(l>2){p3="0"+"{"+(l-2)+"}";}}else{p2="0?";p3="0{0,2}";}switch(c){case "y":s="\\d{2,4}";break;case "M":s=(l>2)?"\\S+?":p2+"[1-9]|1[0-2]";break;case "D":s=p2+"[1-9]|"+p3+"[1-9][0-9]|[12][0-9][0-9]|3[0-5][0-9]|36[0-6]";break;case "d":s="[12]\\d|"+p2+"[1-9]|3[01]";break;case "w":s=p2+"[1-9]|[1-4][0-9]|5[0-3]";break;case "E":s="\\S+";break;case "h":s=p2+"[1-9]|1[0-2]";break;case "k":s=p2+"\\d|1[01]";break;case "H":s=p2+"\\d|1\\d|2[0-3]";break;case "K":s=p2+"[1-9]|1\\d|2[0-4]";break;case "m":case "s":s="[0-5]\\d";break;case "S":s="\\d{"+l+"}";break;case "a":var am=_61a.am||_619.am||"AM";var pm=_61a.pm||_619.pm||"PM";if(_61a.strict){s=am+"|"+pm;}else{s=am+"|"+pm;if(am!=am.toLowerCase()){s+="|"+am.toLowerCase();}if(pm!=pm.toLowerCase()){s+="|"+pm.toLowerCase();}if(s.indexOf(".")!=-1){s+="|"+s.replace(/\./g,"");}}s=s.replace(/\./g,"\\.");break;default:s=".*";}if(_618){_618.push(_61c);}return "("+s+")";}).replace(/[\xa0 ]/g,"[\\s\\xa0]");};})();(function(){var _624=[];dojo.date.locale.addCustomFormats=function(_625,_626){_624.push({pkg:_625,name:_626});};dojo.date.locale._getGregorianBundle=function(_627){var _628={};dojo.forEach(_624,function(desc){var _62a=dojo.i18n.getLocalization(desc.pkg,desc.name,_627);_628=dojo.mixin(_628,_62a);},this);return _628;};})();dojo.date.locale.addCustomFormats("dojo.cldr","gregorian");dojo.date.locale.getNames=function(item,type,_62d,_62e){var _62f;var _630=dojo.date.locale._getGregorianBundle(_62e);var _631=[item,_62d,type];if(_62d=="standAlone"){var key=_631.join("-");_62f=_630[key];if(_62f[0]==1){_62f=undefined;}}_631[1]="format";return (_62f||_630[_631.join("-")]).concat();};dojo.date.locale.isWeekend=function(_633,_634){var _635=dojo.cldr.supplemental.getWeekend(_634);var day=(_633||new Date()).getDay();if(_635.end<_635.start){_635.end+=7;if(day<_635.start){day+=7;}}return day>=_635.start&&day<=_635.end;};dojo.date.locale._getDayOfYear=function(_637){return dojo.date.difference(new Date(_637.getFullYear(),0,1,_637.getHours()),_637)+1;};dojo.date.locale._getWeekOfYear=function(_638,_639){if(arguments.length==1){_639=0;}var _63a=new Date(_638.getFullYear(),0,1).getDay();var adj=(_63a-_639+7)%7;var week=Math.floor((dojo.date.locale._getDayOfYear(_638)+adj-1)/7);if(_63a==_639){week++;}return week;};}if(!dojo._hasResource["dojo.date.stamp"]){dojo._hasResource["dojo.date.stamp"]=true;dojo.provide("dojo.date.stamp");dojo.date.stamp.fromISOString=function(_63d,_63e){if(!dojo.date.stamp._isoRegExp){dojo.date.stamp._isoRegExp=/^(?:(\d{4})(?:-(\d{2})(?:-(\d{2}))?)?)?(?:T(\d{2}):(\d{2})(?::(\d{2})(.\d+)?)?((?:[+-](\d{2}):(\d{2}))|Z)?)?$/;}var _63f=dojo.date.stamp._isoRegExp.exec(_63d);var _640=null;if(_63f){_63f.shift();if(_63f[1]){_63f[1]--;}if(_63f[6]){_63f[6]*=1000;}if(_63e){_63e=new Date(_63e);dojo.map(["FullYear","Month","Date","Hours","Minutes","Seconds","Milliseconds"],function(prop){return _63e["get"+prop]();}).forEach(function(_642,_643){if(_63f[_643]===undefined){_63f[_643]=_642;}});}_640=new Date(_63f[0]||1970,_63f[1]||0,_63f[2]||1,_63f[3]||0,_63f[4]||0,_63f[5]||0,_63f[6]||0);var _644=0;var _645=_63f[7]&&_63f[7].charAt(0);if(_645!="Z"){_644=((_63f[8]||0)*60)+(Number(_63f[9])||0);if(_645!="-"){_644*=-1;}}if(_645){_644-=_640.getTimezoneOffset();}if(_644){_640.setTime(_640.getTime()+_644*60000);}}return _640;};dojo.date.stamp.toISOString=function(_646,_647){var _=function(n){return (n<10)?"0"+n:n;};_647=_647||{};var _64a=[];var _64b=_647.zulu?"getUTC":"get";var date="";if(_647.selector!="time"){var year=_646[_64b+"FullYear"]();date=["0000".substr((year+"").length)+year,_(_646[_64b+"Month"]()+1),_(_646[_64b+"Date"]())].join("-");}_64a.push(date);if(_647.selector!="date"){var time=[_(_646[_64b+"Hours"]()),_(_646[_64b+"Minutes"]()),_(_646[_64b+"Seconds"]())].join(":");var _64f=_646[_64b+"Milliseconds"]();if(_647.milliseconds){time+="."+(_64f<100?"0":"")+_(_64f);}if(_647.zulu){time+="Z";}else{if(_647.selector!="time"){var _650=_646.getTimezoneOffset();var _651=Math.abs(_650);time+=(_650>0?"-":"+")+_(Math.floor(_651/60))+":"+_(_651%60);}}_64a.push(time);}return _64a.join("T");};}if(!dojo._hasResource["dojo.parser"]){dojo._hasResource["dojo.parser"]=true;dojo.provide("dojo.parser");dojo.parser=new function(){var d=dojo;this._attrName=d._scopeName+"Type";this._query="["+this._attrName+"]";var _653=0,_654={};var _655=function(_656,_657){var nso=_657||_654;if(dojo.isIE){var cn=_656["__dojoNameCache"];if(cn&&nso[cn]===_656){return cn;}}var name;do{name="__"+_653++;}while(name in nso);nso[name]=_656;return name;};function val2type(_65b){if(d.isString(_65b)){return "string";}if(typeof _65b=="number"){return "number";}if(typeof _65b=="boolean"){return "boolean";}if(d.isFunction(_65b)){return "function";}if(d.isArray(_65b)){return "array";}if(_65b instanceof Date){return "date";}if(_65b instanceof d._Url){return "url";}return "object";};function str2obj(_65c,type){switch(type){case "string":return _65c;case "number":return _65c.length?Number(_65c):NaN;case "boolean":return typeof _65c=="boolean"?_65c:!(_65c.toLowerCase()=="false");case "function":if(d.isFunction(_65c)){_65c=_65c.toString();_65c=d.trim(_65c.substring(_65c.indexOf("{")+1,_65c.length-1));}try{if(_65c.search(/[^\w\.]+/i)!=-1){_65c=_655(new Function(_65c),this);}return d.getObject(_65c,false);}catch(e){return new Function();}case "array":return _65c?_65c.split(/\s*,\s*/):[];case "date":switch(_65c){case "":return new Date("");case "now":return new Date();default:return d.date.stamp.fromISOString(_65c);}case "url":return d.baseUrl+_65c;default:return d.fromJson(_65c);}};var _65e={};dojo.connect(dojo,"extend",function(){_65e={};});function getClassInfo(_65f){if(!_65e[_65f]){var cls=d.getObject(_65f);if(!d.isFunction(cls)){throw new Error("Could not load class '"+_65f+"'. Did you spell the name correctly and use a full path, like 'dijit.form.Button'?");}var _661=cls.prototype;var _662={},_663={};for(var name in _661){if(name.charAt(0)=="_"){continue;}if(name in _663){continue;}var _665=_661[name];_662[name]=val2type(_665);}_65e[_65f]={cls:cls,params:_662};}return _65e[_65f];};this._functionFromScript=function(_666){var _667="";var _668="";var _669=_666.getAttribute("args");if(_669){d.forEach(_669.split(/\s*,\s*/),function(part,idx){_667+="var "+part+" = arguments["+idx+"]; ";});}var _66c=_666.getAttribute("with");if(_66c&&_66c.length){d.forEach(_66c.split(/\s*,\s*/),function(part){_667+="with("+part+"){";_668+="}";});}return new Function(_667+_666.innerHTML+_668);};this.instantiate=function(_66e,_66f){var _670=[],dp=dojo.parser;_66f=_66f||{};d.forEach(_66e,function(node){if(!node){return;}var type=dp._attrName in _66f?_66f[dp._attrName]:node.getAttribute(dp._attrName);if(!type||!type.length){return;}var _674=getClassInfo(type),_675=_674.cls,ps=_675._noScript||_675.prototype._noScript;var _677={},_678=node.attributes;for(var name in _674.params){var item=name in _66f?{value:_66f[name],specified:true}:_678.getNamedItem(name);if(!item||(!item.specified&&(!dojo.isIE||name.toLowerCase()!="value"))){continue;}var _67b=item.value;switch(name){case "class":_67b="className" in _66f?_66f.className:node.className;break;case "style":_67b="style" in _66f?_66f.style:(node.style&&node.style.cssText);}var _67c=_674.params[name];if(typeof _67b=="string"){_677[name]=str2obj(_67b,_67c);}else{_677[name]=_67b;}}if(!ps){var _67d=[],_67e=[];d.query("> script[type^='dojo/']",node).orphan().forEach(function(_67f){var _680=_67f.getAttribute("event"),type=_67f.getAttribute("type"),nf=d.parser._functionFromScript(_67f);if(_680){if(type=="dojo/connect"){_67d.push({event:_680,func:nf});}else{_677[_680]=nf;}}else{_67e.push(nf);}});}var _682=_675["markupFactory"];if(!_682&&_675["prototype"]){_682=_675.prototype["markupFactory"];}var _683=_682?_682(_677,node,_675):new _675(_677,node);_670.push(_683);var _684=node.getAttribute("jsId");if(_684){d.setObject(_684,_683);}if(!ps){d.forEach(_67d,function(_685){d.connect(_683,_685.event,null,_685.func);});d.forEach(_67e,function(func){func.call(_683);});}});d.forEach(_670,function(_687){if(_687&&_687.startup&&!_687._started&&(!_687.getParent||!_687.getParent())){_687.startup();}});return _670;};this.parse=function(_688){var list=d.query(this._query,_688);var _68a=this.instantiate(list);return _68a;};}();(function(){var _68b=function(){if(dojo.config["parseOnLoad"]==true){dojo.parser.parse();}};if(dojo.exists("dijit.wai.onload")&&(dijit.wai.onload===dojo._loaders[0])){dojo._loaders.splice(1,0,_68b);}else{dojo._loaders.unshift(_68b);}})();}if(!dojo._hasResource["dojo.cookie"]){dojo._hasResource["dojo.cookie"]=true;dojo.provide("dojo.cookie");dojo.cookie=function(name,_68d,_68e){var c=document.cookie;if(arguments.length==1){var _690=c.match(new RegExp("(?:^|; )"+dojo.regexp.escapeString(name)+"=([^;]*)"));return _690?decodeURIComponent(_690[1]):undefined;}else{_68e=_68e||{};var exp=_68e.expires;if(typeof exp=="number"){var d=new Date();d.setTime(d.getTime()+exp*24*60*60*1000);exp=_68e.expires=d;}if(exp&&exp.toUTCString){_68e.expires=exp.toUTCString();}_68d=encodeURIComponent(_68d);var _693=name+"="+_68d,_694;for(_694 in _68e){_693+="; "+_694;var _695=_68e[_694];if(_695!==true){_693+="="+_695;}}document.cookie=_693;}};dojo.cookie.isSupported=function(){if(!("cookieEnabled" in navigator)){this("__djCookieTest__","CookiesAllowed");navigator.cookieEnabled=this("__djCookieTest__")=="CookiesAllowed";if(navigator.cookieEnabled){this("__djCookieTest__","",{expires:-1});}}return navigator.cookieEnabled;};}if(!dojo._hasResource["dojo.io.script"]){dojo._hasResource["dojo.io.script"]=true;dojo.provide("dojo.io.script");dojo.io.script={get:function(args){var dfd=this._makeScriptDeferred(args);var _698=dfd.ioArgs;dojo._ioAddQueryToUrl(_698);if(this._canAttach(_698)){this.attach(_698.id,_698.url,args.frameDoc);}dojo._ioWatch(dfd,this._validCheck,this._ioCheck,this._resHandle);return dfd;},attach:function(id,url,_69b){var doc=(_69b||dojo.doc);var _69d=doc.createElement("script");_69d.type="text/javascript";_69d.src=url;_69d.id=id;_69d.charset="utf-8";doc.getElementsByTagName("head")[0].appendChild(_69d);},remove:function(id,_69f){dojo.destroy(dojo.byId(id,_69f));if(this["jsonp_"+id]){delete this["jsonp_"+id];}},_makeScriptDeferred:function(args){var dfd=dojo._ioSetArgs(args,this._deferredCancel,this._deferredOk,this._deferredError);var _6a2=dfd.ioArgs;_6a2.id=dojo._scopeName+"IoScript"+(this._counter++);_6a2.canDelete=false;if(args.callbackParamName){_6a2.query=_6a2.query||"";if(_6a2.query.length>0){_6a2.query+="&";}_6a2.query+=args.callbackParamName+"="+(args.frameDoc?"parent.":"")+dojo._scopeName+".io.script.jsonp_"+_6a2.id+"._jsonpCallback";_6a2.frameDoc=args.frameDoc;_6a2.canDelete=true;dfd._jsonpCallback=this._jsonpCallback;this["jsonp_"+_6a2.id]=dfd;}return dfd;},_deferredCancel:function(dfd){dfd.canceled=true;if(dfd.ioArgs.canDelete){dojo.io.script._addDeadScript(dfd.ioArgs);}},_deferredOk:function(dfd){if(dfd.ioArgs.canDelete){dojo.io.script._addDeadScript(dfd.ioArgs);}if(dfd.ioArgs.json){return dfd.ioArgs.json;}else{return dfd.ioArgs;}},_deferredError:function(_6a5,dfd){if(dfd.ioArgs.canDelete){if(_6a5.dojoType=="timeout"){dojo.io.script.remove(dfd.ioArgs.id,dfd.ioArgs.frameDoc);}else{dojo.io.script._addDeadScript(dfd.ioArgs);}}console.log("dojo.io.script error",_6a5);return _6a5;},_deadScripts:[],_counter:1,_addDeadScript:function(_6a7){dojo.io.script._deadScripts.push({id:_6a7.id,frameDoc:_6a7.frameDoc});_6a7.frameDoc=null;},_validCheck:function(dfd){var _6a9=dojo.io.script;var _6aa=_6a9._deadScripts;if(_6aa&&_6aa.length>0){for(var i=0;i<_6aa.length;i++){_6a9.remove(_6aa[i].id,_6aa[i].frameDoc);_6aa[i].frameDoc=null;}dojo.io.script._deadScripts=[];}return true;},_ioCheck:function(dfd){if(dfd.ioArgs.json){return true;}var _6ad=dfd.ioArgs.args.checkString;if(_6ad&&eval("typeof("+_6ad+") != 'undefined'")){return true;}return false;},_resHandle:function(dfd){if(dojo.io.script._ioCheck(dfd)){dfd.callback(dfd);}else{dfd.errback(new Error("inconceivable dojo.io.script._resHandle error"));}},_canAttach:function(_6af){return true;},_jsonpCallback:function(json){this.ioArgs.json=json;}};}if(!dojo._hasResource["dojo.io.iframe"]){dojo._hasResource["dojo.io.iframe"]=true;dojo.provide("dojo.io.iframe");dojo.io.iframe={create:function(_6b1,_6b2,uri){if(window[_6b1]){return window[_6b1];}if(window.frames[_6b1]){return window.frames[_6b1];}var _6b4=null;var turi=uri;if(!turi){if(dojo.config["useXDomain"]&&!dojo.config["dojoBlankHtmlUrl"]){console.warn("dojo.io.iframe.create: When using cross-domain Dojo builds,"+" please save dojo/resources/blank.html to your domain and set djConfig.dojoBlankHtmlUrl"+" to the path on your domain to blank.html");}turi=(dojo.config["dojoBlankHtmlUrl"]||dojo.moduleUrl("dojo","resources/blank.html"));}var _6b6=dojo.isIE?"<iframe name=\""+_6b1+"\" src=\""+turi+"\" onload=\""+_6b2+"\">":"iframe";_6b4=dojo.doc.createElement(_6b6);with(_6b4){name=_6b1;setAttribute("name",_6b1);id=_6b1;}dojo.body().appendChild(_6b4);window[_6b1]=_6b4;with(_6b4.style){if(!(dojo.isSafari<3)){position="absolute";}left=top="1px";height=width="1px";visibility="hidden";}if(!dojo.isIE){this.setSrc(_6b4,turi,true);_6b4.onload=new Function(_6b2);}return _6b4;},setSrc:function(_6b7,src,_6b9){try{if(!_6b9){if(dojo.isWebKit){_6b7.location=src;}else{frames[_6b7.name].location=src;}}else{var idoc;if(dojo.isIE||dojo.isWebKit>521){idoc=_6b7.contentWindow.document;}else{if(dojo.isSafari){idoc=_6b7.document;}else{idoc=_6b7.contentWindow;}}if(!idoc){_6b7.location=src;return;}else{idoc.location.replace(src);}}}catch(e){console.log("dojo.io.iframe.setSrc: ",e);}},doc:function(_6bb){var doc=_6bb.contentDocument||(((_6bb.name)&&(_6bb.document)&&(document.getElementsByTagName("iframe")[_6bb.name].contentWindow)&&(document.getElementsByTagName("iframe")[_6bb.name].contentWindow.document)))||((_6bb.name)&&(document.frames[_6bb.name])&&(document.frames[_6bb.name].document))||null;return doc;},send:function(args){if(!this["_frame"]){this._frame=this.create(this._iframeName,dojo._scopeName+".io.iframe._iframeOnload();");}var dfd=dojo._ioSetArgs(args,function(dfd){dfd.canceled=true;dfd.ioArgs._callNext();},function(dfd){var _6c1=null;try{var _6c2=dfd.ioArgs;var dii=dojo.io.iframe;var ifd=dii.doc(dii._frame);var _6c5=_6c2.handleAs;_6c1=ifd;if(_6c5!="html"){if(_6c5=="xml"){if(dojo.isIE){dojo.query("a",dii._frame.contentWindow.document.documentElement).orphan();var _6c6=(dii._frame.contentWindow.document).documentElement.innerText;_6c6=_6c6.replace(/>\s+</g,"><");_6c6=dojo.trim(_6c6);var _6c7={responseText:_6c6};_6c1=dojo._contentHandlers["xml"](_6c7);}}else{_6c1=ifd.getElementsByTagName("textarea")[0].value;if(_6c5=="json"){_6c1=dojo.fromJson(_6c1);}else{if(_6c5=="javascript"){_6c1=dojo.eval(_6c1);}}}}}catch(e){_6c1=e;}finally{_6c2._callNext();}return _6c1;},function(_6c8,dfd){dfd.ioArgs._hasError=true;dfd.ioArgs._callNext();return _6c8;});dfd.ioArgs._callNext=function(){if(!this["_calledNext"]){this._calledNext=true;dojo.io.iframe._currentDfd=null;dojo.io.iframe._fireNextRequest();}};this._dfdQueue.push(dfd);this._fireNextRequest();dojo._ioWatch(dfd,function(dfd){return !dfd.ioArgs["_hasError"];},function(dfd){return (!!dfd.ioArgs["_finished"]);},function(dfd){if(dfd.ioArgs._finished){dfd.callback(dfd);}else{dfd.errback(new Error("Invalid dojo.io.iframe request state"));}});return dfd;},_currentDfd:null,_dfdQueue:[],_iframeName:dojo._scopeName+"IoIframe",_fireNextRequest:function(){try{if((this._currentDfd)||(this._dfdQueue.length==0)){return;}var dfd=this._currentDfd=this._dfdQueue.shift();var _6ce=dfd.ioArgs;var args=_6ce.args;_6ce._contentToClean=[];var fn=dojo.byId(args["form"]);var _6d1=args["content"]||{};if(fn){if(_6d1){var _6d2=function(name,_6d4){var tn;if(dojo.isIE){tn=dojo.doc.createElement("<input type='hidden' name='"+name+"'>");}else{tn=dojo.doc.createElement("input");tn.type="hidden";tn.name=name;}tn.value=_6d4;fn.appendChild(tn);_6ce._contentToClean.push(name);};for(var x in _6d1){var val=_6d1[x];if(dojo.isArray(val)&&val.length>1){var i;for(i=0;i<val.length;i++){_6d2(x,val[i]);}}else{if(!fn[x]){_6d2(x,val);}else{fn[x].value=val;}}}}var _6d9=fn.getAttributeNode("action");var _6da=fn.getAttributeNode("method");var _6db=fn.getAttributeNode("target");if(args["url"]){_6ce._originalAction=_6d9?_6d9.value:null;if(_6d9){_6d9.value=args.url;}else{fn.setAttribute("action",args.url);}}if(!_6da||!_6da.value){if(_6da){_6da.value=(args["method"])?args["method"]:"post";}else{fn.setAttribute("method",(args["method"])?args["method"]:"post");}}_6ce._originalTarget=_6db?_6db.value:null;if(_6db){_6db.value=this._iframeName;}else{fn.setAttribute("target",this._iframeName);}fn.target=this._iframeName;fn.submit();}else{var _6dc=args.url+(args.url.indexOf("?")>-1?"&":"?")+_6ce.query;this.setSrc(this._frame,_6dc,true);}}catch(e){dfd.errback(e);}},_iframeOnload:function(){var dfd=this._currentDfd;if(!dfd){this._fireNextRequest();return;}var _6de=dfd.ioArgs;var args=_6de.args;var _6e0=dojo.byId(args.form);if(_6e0){var _6e1=_6de._contentToClean;for(var i=0;i<_6e1.length;i++){var key=_6e1[i];for(var j=0;j<_6e0.childNodes.length;j++){var _6e5=_6e0.childNodes[j];if(_6e5.name==key){dojo.destroy(_6e5);break;}}}if(_6de["_originalAction"]){_6e0.setAttribute("action",_6de._originalAction);}if(_6de["_originalTarget"]){_6e0.setAttribute("target",_6de._originalTarget);_6e0.target=_6de._originalTarget;}}_6de._finished=true;}};}dojo.i18n._preloadLocalizations("dojo.nls.dojo",["ROOT","aa","aa-dj","aa-er","aa-et","af","af-za","am","am-et","ar","ar-dz","ar-jo","ar-lb","ar-ma","ar-qa","ar-sa","ar-sy","ar-tn","ar-ye","as","as-in","az","be","be-by","bg","bg-bg","bn","bn-bd","bn-in","bs","byn","byn-er","ca","ca-es","cs","cs-cz","cy","cy-gb","da","da-dk","de","de-at","de-be","de-ch","de-de","de-li","de-lu","dv","dz","dz-bt","el","el-cy","el-gr","el-polytoni","en","en-as","en-au","en-be","en-bw","en-ca","en-gb","en-gu","en-hk","en-ie","en-in","en-mh","en-mp","en-mt","en-nz","en-ph","en-pk","en-sg","en-um","en-us","en-us-posix","en-vi","en-za","en-zw","eo","es","es-ar","es-bo","es-cl","es-co","es-cr","es-do","es-ec","es-es","es-gt","es-hn","es-mx","es-ni","es-pa","es-pr","es-py","es-sv","es-us","es-uy","es-ve","et","et-ee","eu","eu-es","fa","fa-af","fi","fi-fi","fo","fo-fo","fr","fr-be","fr-ca","fr-ch","fr-lu","fur","ga","ga-ie","gez","gez-er","gez-et","gl","gl-es","gu","gu-in","gv","gv-gb","haw","haw-us","he","he-il","hi","hi-in","hr","hu","hu-hu","hy","hy-am","hy-am-revised","ia","id","id-id","is","is-is","it","it-ch","it-it","ja","ja-jp","ka","kk","kk-kz","kl","kl-gl","km","km-kh","kn","kn-in","ko","ko-kr","kok","kok-in","kw","kw-gb","ky","ln","lo","lo-la","lt","lt-lt","lv","lv-lv","mk","mk-mk","ml","ml-in","mn","mr","mr-in","ms","ms-bn","ms-my","mt","mt-mt","nb","nb-no","ne","nl","nl-be","nl-nl","nn","nn-no","om","om-et","om-ke","or","or-in","pa","pa-arab","pa-in","pl","pl-pl","ps","ps-af","pt","pt-br","pt-pt","ro","ro-ro","ru","ru-ua","rw","sa","se","sh","sh-ba","sid","sid-et","sk","sk-sk","sl","sl-si","so","so-dj","so-et","so-ke","so-so","sq","sq-al","sr","sr-ba","sr-cyrl","sr-cyrl-ba","sr-latn","sr-latn-ba","sv","sv-fi","sv-se","sw","sw-ke","sw-tz","syr","syr-sy","ta","ta-in","te","te-in","th","th-th","ti","ti-er","ti-et","tig","tig-er","tr","tt","tt-ru","uk","uk-ua","uz","uz-arab","vi","wal","wal-et","xh","xx","zh","zh-cn","zh-hans","zh-hans-cn","zh-hans-sg","zh-hant","zh-hant-hk","zh-hant-mo","zh-hant-tw","zh-hk","zh-mo","zh-sg","zh-tw","zu"]);if(dojo.config.afterOnLoad&&dojo.isBrowser){window.setTimeout(dojo._loadInit,1000);}})();
/*
	Copyright (c) 2004-2009, The Dojo Foundation All Rights Reserved.
	Available via Academic Free License >= 2.1 OR the modified BSD license.
	see: http://dojotoolkit.org/license for details
*/

/*
	This is a compiled version of Dojo, built for deployment and not for
	development. To get an editable version, please visit:

		http://dojotoolkit.org

	for documentation and information on getting the source.
*/

dojo.provide("dijit.dijit");if(!dojo._hasResource["dijit._base.focus"]){dojo._hasResource["dijit._base.focus"]=true;dojo.provide("dijit._base.focus");dojo.mixin(dijit,{_curFocus:null,_prevFocus:null,isCollapsed:function(){var _1=dojo.doc;if(_1.selection&&!dojo.isOpera){var s=_1.selection;if(s.type=="Text"){return !s.createRange().htmlText.length;}else{return !s.createRange().length;}}else{var _3=dojo.global;var _4=_3.getSelection();if(dojo.isString(_4)){return !_4;}else{return !_4||_4.isCollapsed||!_4.toString();}}},getBookmark:function(){var _5,_6=dojo.doc.selection;if(_6&&!dojo.isOpera){var _7=_6.createRange();if(_6.type.toUpperCase()=="CONTROL"){if(_7.length){_5=[];var i=0,_9=_7.length;while(i<_9){_5.push(_7.item(i++));}}else{_5=null;}}else{_5=_7.getBookmark();}}else{if(window.getSelection){_6=dojo.global.getSelection();if(_6){_7=_6.getRangeAt(0);_5=_7.cloneRange();}}else{console.warn("No idea how to store the current selection for this browser!");}}return _5;},moveToBookmark:function(_a){var _b=dojo.doc;if(_b.selection&&!dojo.isOpera){var _c;if(dojo.isArray(_a)){_c=_b.body.createControlRange();dojo.forEach(_a,function(n){_c.addElement(n);});}else{_c=_b.selection.createRange();_c.moveToBookmark(_a);}_c.select();}else{var _e=dojo.global.getSelection&&dojo.global.getSelection();if(_e&&_e.removeAllRanges){_e.removeAllRanges();_e.addRange(_a);}else{console.warn("No idea how to restore selection for this browser!");}}},getFocus:function(_f,_10){return {node:_f&&dojo.isDescendant(dijit._curFocus,_f.domNode)?dijit._prevFocus:dijit._curFocus,bookmark:!dojo.withGlobal(_10||dojo.global,dijit.isCollapsed)?dojo.withGlobal(_10||dojo.global,dijit.getBookmark):null,openedForWindow:_10};},focus:function(_11){if(!_11){return;}var _12="node" in _11?_11.node:_11,_13=_11.bookmark,_14=_11.openedForWindow;if(_12){var _15=(_12.tagName.toLowerCase()=="iframe")?_12.contentWindow:_12;if(_15&&_15.focus){try{_15.focus();}catch(e){}}dijit._onFocusNode(_12);}if(_13&&dojo.withGlobal(_14||dojo.global,dijit.isCollapsed)){if(_14){_14.focus();}try{dojo.withGlobal(_14||dojo.global,dijit.moveToBookmark,null,[_13]);}catch(e){}}},_activeStack:[],registerIframe:function(_16){dijit.registerWin(_16.contentWindow,_16);},registerWin:function(_17,_18){dojo.connect(_17.document,"onmousedown",function(evt){dijit._justMouseDowned=true;setTimeout(function(){dijit._justMouseDowned=false;},0);dijit._onTouchNode(_18||evt.target||evt.srcElement);});var doc=_17.document;if(doc){if(dojo.isIE){doc.attachEvent("onactivate",function(evt){if(evt.srcElement.tagName.toLowerCase()!="#document"){dijit._onFocusNode(_18||evt.srcElement);}});doc.attachEvent("ondeactivate",function(evt){dijit._onBlurNode(_18||evt.srcElement);});}else{doc.addEventListener("focus",function(evt){dijit._onFocusNode(_18||evt.target);},true);doc.addEventListener("blur",function(evt){dijit._onBlurNode(_18||evt.target);},true);}}doc=null;},_onBlurNode:function(_1f){dijit._prevFocus=dijit._curFocus;dijit._curFocus=null;if(dijit._justMouseDowned){return;}if(dijit._clearActiveWidgetsTimer){clearTimeout(dijit._clearActiveWidgetsTimer);}dijit._clearActiveWidgetsTimer=setTimeout(function(){delete dijit._clearActiveWidgetsTimer;dijit._setStack([]);dijit._prevFocus=null;},100);},_onTouchNode:function(_20){if(dijit._clearActiveWidgetsTimer){clearTimeout(dijit._clearActiveWidgetsTimer);delete dijit._clearActiveWidgetsTimer;}var _21=[];try{while(_20){if(_20.dijitPopupParent){_20=dijit.byId(_20.dijitPopupParent).domNode;}else{if(_20.tagName&&_20.tagName.toLowerCase()=="body"){if(_20===dojo.body()){break;}_20=dijit.getDocumentWindow(_20.ownerDocument).frameElement;}else{var id=_20.getAttribute&&_20.getAttribute("widgetId");if(id){_21.unshift(id);}_20=_20.parentNode;}}}}catch(e){}dijit._setStack(_21);},_onFocusNode:function(_23){if(!_23){return;}if(_23.nodeType==9){return;}dijit._onTouchNode(_23);if(_23==dijit._curFocus){return;}if(dijit._curFocus){dijit._prevFocus=dijit._curFocus;}dijit._curFocus=_23;dojo.publish("focusNode",[_23]);},_setStack:function(_24){var _25=dijit._activeStack;dijit._activeStack=_24;for(var _26=0;_26<Math.min(_25.length,_24.length);_26++){if(_25[_26]!=_24[_26]){break;}}for(var i=_25.length-1;i>=_26;i--){var _28=dijit.byId(_25[i]);if(_28){_28._focused=false;_28._hasBeenBlurred=true;if(_28._onBlur){_28._onBlur();}if(_28._setStateClass){_28._setStateClass();}dojo.publish("widgetBlur",[_28]);}}for(i=_26;i<_24.length;i++){_28=dijit.byId(_24[i]);if(_28){_28._focused=true;if(_28._onFocus){_28._onFocus();}if(_28._setStateClass){_28._setStateClass();}dojo.publish("widgetFocus",[_28]);}}}});dojo.addOnLoad(function(){dijit.registerWin(window);});}if(!dojo._hasResource["dijit._base.manager"]){dojo._hasResource["dijit._base.manager"]=true;dojo.provide("dijit._base.manager");dojo.declare("dijit.WidgetSet",null,{constructor:function(){this._hash={};},add:function(_29){if(this._hash[_29.id]){throw new Error("Tried to register widget with id=="+_29.id+" but that id is already registered");}this._hash[_29.id]=_29;},remove:function(id){delete this._hash[id];},forEach:function(_2b){for(var id in this._hash){_2b(this._hash[id]);}},filter:function(_2d){var res=new dijit.WidgetSet();this.forEach(function(_2f){if(_2d(_2f)){res.add(_2f);}});return res;},byId:function(id){return this._hash[id];},byClass:function(cls){return this.filter(function(_32){return _32.declaredClass==cls;});}});dijit.registry=new dijit.WidgetSet();dijit._widgetTypeCtr={};dijit.getUniqueId=function(_33){var id;do{id=_33+"_"+(_33 in dijit._widgetTypeCtr?++dijit._widgetTypeCtr[_33]:dijit._widgetTypeCtr[_33]=0);}while(dijit.byId(id));return id;};dijit.findWidgets=function(_35){var _36=[];function getChildrenHelper(_37){var _38=dojo.isIE?_37.children:_37.childNodes,i=0,_3a;while(_3a=_38[i++]){if(_3a.nodeType!=1){continue;}var _3b=_3a.getAttribute("widgetId");if(_3b){var _3c=dijit.byId(_3b);_36.push(_3c);}else{getChildrenHelper(_3a);}}};getChildrenHelper(_35);return _36;};if(dojo.isIE){dojo.addOnWindowUnload(function(){dojo.forEach(dijit.findWidgets(dojo.body()),function(_3d){if(_3d.destroyRecursive){_3d.destroyRecursive();}else{if(_3d.destroy){_3d.destroy();}}});});}dijit.byId=function(id){return (dojo.isString(id))?dijit.registry.byId(id):id;};dijit.byNode=function(_3f){return dijit.registry.byId(_3f.getAttribute("widgetId"));};dijit.getEnclosingWidget=function(_40){while(_40){if(_40.getAttribute&&_40.getAttribute("widgetId")){return dijit.registry.byId(_40.getAttribute("widgetId"));}_40=_40.parentNode;}return null;};dijit._tabElements={area:true,button:true,input:true,object:true,select:true,textarea:true};dijit._isElementShown=function(_41){var _42=dojo.style(_41);return (_42.visibility!="hidden")&&(_42.visibility!="collapsed")&&(_42.display!="none")&&(dojo.attr(_41,"type")!="hidden");};dijit.isTabNavigable=function(_43){if(dojo.hasAttr(_43,"disabled")){return false;}var _44=dojo.hasAttr(_43,"tabindex");var _45=dojo.attr(_43,"tabindex");if(_44&&_45>=0){return true;}var _46=_43.nodeName.toLowerCase();if(((_46=="a"&&dojo.hasAttr(_43,"href"))||dijit._tabElements[_46])&&(!_44||_45>=0)){return true;}return false;};dijit._getTabNavigable=function(_47){var _48,_49,_4a,_4b,_4c,_4d;var _4e=function(_4f){dojo.query("> *",_4f).forEach(function(_50){var _51=dijit._isElementShown(_50);if(_51&&dijit.isTabNavigable(_50)){var _52=dojo.attr(_50,"tabindex");if(!dojo.hasAttr(_50,"tabindex")||_52==0){if(!_48){_48=_50;}_49=_50;}else{if(_52>0){if(!_4a||_52<_4b){_4b=_52;_4a=_50;}if(!_4c||_52>=_4d){_4d=_52;_4c=_50;}}}}if(_51&&_50.nodeName.toUpperCase()!="SELECT"){_4e(_50);}});};if(dijit._isElementShown(_47)){_4e(_47);}return {first:_48,last:_49,lowest:_4a,highest:_4c};};dijit.getFirstInTabbingOrder=function(_53){var _54=dijit._getTabNavigable(dojo.byId(_53));return _54.lowest?_54.lowest:_54.first;};dijit.getLastInTabbingOrder=function(_55){var _56=dijit._getTabNavigable(dojo.byId(_55));return _56.last?_56.last:_56.highest;};dijit.defaultDuration=dojo.config["defaultDuration"]||200;}if(!dojo._hasResource["dojo.AdapterRegistry"]){dojo._hasResource["dojo.AdapterRegistry"]=true;dojo.provide("dojo.AdapterRegistry");dojo.AdapterRegistry=function(_57){this.pairs=[];this.returnWrappers=_57||false;};dojo.extend(dojo.AdapterRegistry,{register:function(_58,_59,_5a,_5b,_5c){this.pairs[((_5c)?"unshift":"push")]([_58,_59,_5a,_5b]);},match:function(){for(var i=0;i<this.pairs.length;i++){var _5e=this.pairs[i];if(_5e[1].apply(this,arguments)){if((_5e[3])||(this.returnWrappers)){return _5e[2];}else{return _5e[2].apply(this,arguments);}}}throw new Error("No match found");},unregister:function(_5f){for(var i=0;i<this.pairs.length;i++){var _61=this.pairs[i];if(_61[0]==_5f){this.pairs.splice(i,1);return true;}}return false;}});}if(!dojo._hasResource["dijit._base.place"]){dojo._hasResource["dijit._base.place"]=true;dojo.provide("dijit._base.place");dijit.getViewport=function(){var _62=(dojo.doc.compatMode=="BackCompat")?dojo.body():dojo.doc.documentElement;var _63=dojo._docScroll();return {w:_62.clientWidth,h:_62.clientHeight,l:_63.x,t:_63.y};};dijit.placeOnScreen=function(_64,pos,_66,_67){var _68=dojo.map(_66,function(_69){var c={corner:_69,pos:{x:pos.x,y:pos.y}};if(_67){c.pos.x+=_69.charAt(1)=="L"?_67.x:-_67.x;c.pos.y+=_69.charAt(0)=="T"?_67.y:-_67.y;}return c;});return dijit._place(_64,_68);};dijit._place=function(_6b,_6c,_6d){var _6e=dijit.getViewport();if(!_6b.parentNode||String(_6b.parentNode.tagName).toLowerCase()!="body"){dojo.body().appendChild(_6b);}var _6f=null;dojo.some(_6c,function(_70){var _71=_70.corner;var pos=_70.pos;if(_6d){_6d(_6b,_70.aroundCorner,_71);}var _73=_6b.style;var _74=_73.display;var _75=_73.visibility;_73.visibility="hidden";_73.display="";var mb=dojo.marginBox(_6b);_73.display=_74;_73.visibility=_75;var _77=(_71.charAt(1)=="L"?pos.x:Math.max(_6e.l,pos.x-mb.w)),_78=(_71.charAt(0)=="T"?pos.y:Math.max(_6e.t,pos.y-mb.h)),_79=(_71.charAt(1)=="L"?Math.min(_6e.l+_6e.w,_77+mb.w):pos.x),_7a=(_71.charAt(0)=="T"?Math.min(_6e.t+_6e.h,_78+mb.h):pos.y),_7b=_79-_77,_7c=_7a-_78,_7d=(mb.w-_7b)+(mb.h-_7c);if(_6f==null||_7d<_6f.overflow){_6f={corner:_71,aroundCorner:_70.aroundCorner,x:_77,y:_78,w:_7b,h:_7c,overflow:_7d};}return !_7d;});_6b.style.left=_6f.x+"px";_6b.style.top=_6f.y+"px";if(_6f.overflow&&_6d){_6d(_6b,_6f.aroundCorner,_6f.corner);}return _6f;};dijit.placeOnScreenAroundNode=function(_7e,_7f,_80,_81){_7f=dojo.byId(_7f);var _82=_7f.style.display;_7f.style.display="";var _83=_7f.offsetWidth;var _84=_7f.offsetHeight;var _85=dojo.coords(_7f,true);_7f.style.display=_82;return dijit._placeOnScreenAroundRect(_7e,_85.x,_85.y,_83,_84,_80,_81);};dijit.placeOnScreenAroundRectangle=function(_86,_87,_88,_89){return dijit._placeOnScreenAroundRect(_86,_87.x,_87.y,_87.width,_87.height,_88,_89);};dijit._placeOnScreenAroundRect=function(_8a,x,y,_8d,_8e,_8f,_90){var _91=[];for(var _92 in _8f){_91.push({aroundCorner:_92,corner:_8f[_92],pos:{x:x+(_92.charAt(1)=="L"?0:_8d),y:y+(_92.charAt(0)=="T"?0:_8e)}});}return dijit._place(_8a,_91,_90);};dijit.placementRegistry=new dojo.AdapterRegistry();dijit.placementRegistry.register("node",function(n,x){return typeof x=="object"&&typeof x.offsetWidth!="undefined"&&typeof x.offsetHeight!="undefined";},dijit.placeOnScreenAroundNode);dijit.placementRegistry.register("rect",function(n,x){return typeof x=="object"&&"x" in x&&"y" in x&&"width" in x&&"height" in x;},dijit.placeOnScreenAroundRectangle);dijit.placeOnScreenAroundElement=function(_97,_98,_99,_9a){return dijit.placementRegistry.match.apply(dijit.placementRegistry,arguments);};}if(!dojo._hasResource["dijit._base.window"]){dojo._hasResource["dijit._base.window"]=true;dojo.provide("dijit._base.window");dijit.getDocumentWindow=function(doc){if(dojo.isIE&&window!==document.parentWindow&&!doc._parentWindow){doc.parentWindow.execScript("document._parentWindow = window;","Javascript");var win=doc._parentWindow;doc._parentWindow=null;return win;}return doc._parentWindow||doc.parentWindow||doc.defaultView;};}if(!dojo._hasResource["dijit._base.popup"]){dojo._hasResource["dijit._base.popup"]=true;dojo.provide("dijit._base.popup");dijit.popup=new function(){var _9d=[],_9e=1000,_9f=1;this.prepare=function(_a0){var s=_a0.style;s.visibility="hidden";s.position="absolute";s.top="-9999px";if(s.display=="none"){s.display="";}dojo.body().appendChild(_a0);};this.open=function(_a2){var _a3=_a2.popup,_a4=_a2.orient||{"BL":"TL","TL":"BL"},_a5=_a2.around,id=(_a2.around&&_a2.around.id)?(_a2.around.id+"_dropdown"):("popup_"+_9f++);var _a7=dojo.create("div",{id:id,"class":"dijitPopup",style:{zIndex:_9e+_9d.length,visibility:"hidden"}},dojo.body());dijit.setWaiRole(_a7,"presentation");_a7.style.left=_a7.style.top="0px";if(_a2.parent){_a7.dijitPopupParent=_a2.parent.id;}var s=_a3.domNode.style;s.display="";s.visibility="";s.position="";s.top="0px";_a7.appendChild(_a3.domNode);var _a9=new dijit.BackgroundIframe(_a7);var _aa=_a5?dijit.placeOnScreenAroundElement(_a7,_a5,_a4,_a3.orient?dojo.hitch(_a3,"orient"):null):dijit.placeOnScreen(_a7,_a2,_a4=="R"?["TR","BR","TL","BL"]:["TL","BL","TR","BR"],_a2.padding);_a7.style.visibility="visible";var _ab=[];var _ac=function(){for(var pi=_9d.length-1;pi>0&&_9d[pi].parent===_9d[pi-1].widget;pi--){}return _9d[pi];};_ab.push(dojo.connect(_a7,"onkeypress",this,function(evt){if(evt.charOrCode==dojo.keys.ESCAPE&&_a2.onCancel){dojo.stopEvent(evt);_a2.onCancel();}else{if(evt.charOrCode===dojo.keys.TAB){dojo.stopEvent(evt);var _af=_ac();if(_af&&_af.onCancel){_af.onCancel();}}}}));if(_a3.onCancel){_ab.push(dojo.connect(_a3,"onCancel",null,_a2.onCancel));}_ab.push(dojo.connect(_a3,_a3.onExecute?"onExecute":"onChange",null,function(){var _b0=_ac();if(_b0&&_b0.onExecute){_b0.onExecute();}}));_9d.push({wrapper:_a7,iframe:_a9,widget:_a3,parent:_a2.parent,onExecute:_a2.onExecute,onCancel:_a2.onCancel,onClose:_a2.onClose,handlers:_ab});if(_a3.onOpen){_a3.onOpen(_aa);}return _aa;};this.close=function(_b1){while(dojo.some(_9d,function(_b2){return _b2.widget==_b1;})){var top=_9d.pop(),_b4=top.wrapper,_b5=top.iframe,_b6=top.widget,_b7=top.onClose;if(_b6.onClose){_b6.onClose();}dojo.forEach(top.handlers,dojo.disconnect);if(!_b6||!_b6.domNode){return;}this.prepare(_b6.domNode);_b5.destroy();dojo.destroy(_b4);if(_b7){_b7();}}};}();dijit._frames=new function(){var _b8=[];this.pop=function(){var _b9;if(_b8.length){_b9=_b8.pop();_b9.style.display="";}else{if(dojo.isIE){var _ba=dojo.config["dojoBlankHtmlUrl"]||(dojo.moduleUrl("dojo","resources/blank.html")+"")||"javascript:\"\"";var _bb="<iframe src='"+_ba+"'"+" style='position: absolute; left: 0px; top: 0px;"+"z-index: -1; filter:Alpha(Opacity=\"0\");'>";_b9=dojo.doc.createElement(_bb);}else{_b9=dojo.create("iframe");_b9.src="javascript:\"\"";_b9.className="dijitBackgroundIframe";}_b9.tabIndex=-1;dojo.body().appendChild(_b9);}return _b9;};this.push=function(_bc){_bc.style.display="none";if(dojo.isIE){_bc.style.removeExpression("width");_bc.style.removeExpression("height");}_b8.push(_bc);};}();dijit.BackgroundIframe=function(_bd){if(!_bd.id){throw new Error("no id");}if(dojo.isIE<7||(dojo.isFF<3&&dojo.hasClass(dojo.body(),"dijit_a11y"))){var _be=dijit._frames.pop();_bd.appendChild(_be);if(dojo.isIE){_be.style.setExpression("width",dojo._scopeName+".doc.getElementById('"+_bd.id+"').offsetWidth");_be.style.setExpression("height",dojo._scopeName+".doc.getElementById('"+_bd.id+"').offsetHeight");}this.iframe=_be;}};dojo.extend(dijit.BackgroundIframe,{destroy:function(){if(this.iframe){dijit._frames.push(this.iframe);delete this.iframe;}}});}if(!dojo._hasResource["dijit._base.scroll"]){dojo._hasResource["dijit._base.scroll"]=true;dojo.provide("dijit._base.scroll");dijit.scrollIntoView=function(_bf){try{_bf=dojo.byId(_bf);var doc=_bf.ownerDocument||dojo.doc;var _c1=doc.body||dojo.body();var _c2=doc.documentElement||_c1.parentNode;if((!(dojo.isFF>=3||dojo.isIE||dojo.isWebKit)||_bf==_c1||_bf==_c2)&&(typeof _bf.scrollIntoView=="function")){_bf.scrollIntoView(false);return;}var ltr=dojo._isBodyLtr();var _c4=dojo.isIE>=8&&!_c5;var rtl=!ltr&&!_c4;var _c7=_c1;var _c5=doc.compatMode=="BackCompat";if(_c5){_c2._offsetWidth=_c2._clientWidth=_c1._offsetWidth=_c1.clientWidth;_c2._offsetHeight=_c2._clientHeight=_c1._offsetHeight=_c1.clientHeight;}else{if(dojo.isWebKit){_c1._offsetWidth=_c1._clientWidth=_c2.clientWidth;_c1._offsetHeight=_c1._clientHeight=_c2.clientHeight;}else{_c7=_c2;}_c2._offsetHeight=_c2.clientHeight;_c2._offsetWidth=_c2.clientWidth;}function isFixedPosition(_c8){var ie=dojo.isIE;return ((ie<=6||(ie>=7&&_c5))?false:(dojo.style(_c8,"position").toLowerCase()=="fixed"));};function addPseudoAttrs(_ca){var _cb=_ca.parentNode;var _cc=_ca.offsetParent;if(_cc==null||isFixedPosition(_ca)){_cc=_c2;_cb=(_ca==_c1)?_c2:null;}_ca._offsetParent=_cc;_ca._parent=_cb;var bp=dojo._getBorderExtents(_ca);_ca._borderStart={H:(_c4&&!ltr)?(bp.w-bp.l):bp.l,V:bp.t};_ca._borderSize={H:bp.w,V:bp.h};_ca._scrolledAmount={H:_ca.scrollLeft,V:_ca.scrollTop};_ca._offsetSize={H:_ca._offsetWidth||_ca.offsetWidth,V:_ca._offsetHeight||_ca.offsetHeight};_ca._offsetStart={H:(_c4&&!ltr)?_cc.clientWidth-_ca.offsetLeft-_ca._offsetSize.H:_ca.offsetLeft,V:_ca.offsetTop};_ca._clientSize={H:_ca._clientWidth||_ca.clientWidth,V:_ca._clientHeight||_ca.clientHeight};if(_ca!=_c1&&_ca!=_c2&&_ca!=_bf){for(var dir in _ca._offsetSize){var _cf=_ca._offsetSize[dir]-_ca._clientSize[dir]-_ca._borderSize[dir];var _d0=_ca._clientSize[dir]>0&&_cf>0;if(_d0){_ca._offsetSize[dir]-=_cf;if(dojo.isIE&&rtl&&dir=="H"){_ca._offsetStart[dir]+=_cf;}}}}};var _d1=_bf;while(_d1!=null){if(isFixedPosition(_d1)){_bf.scrollIntoView(false);return;}addPseudoAttrs(_d1);_d1=_d1._parent;}if(dojo.isIE&&_bf._parent){var _d2=_bf._offsetParent;_bf._offsetStart.H+=_d2._borderStart.H;_bf._offsetStart.V+=_d2._borderStart.V;}if(dojo.isIE>=7&&_c7==_c2&&rtl&&_c1._offsetStart&&_c1._offsetStart.H==0){var _d3=_c2.scrollWidth-_c2._offsetSize.H;if(_d3>0){_c1._offsetStart.H=-_d3;}}if(dojo.isIE<=6&&!_c5){_c2._offsetSize.H+=_c2._borderSize.H;_c2._offsetSize.V+=_c2._borderSize.V;}if(rtl&&_c1._offsetStart&&_c7==_c2&&_c2._scrolledAmount){var ofs=_c1._offsetStart.H;if(ofs<0){_c2._scrolledAmount.H+=ofs;_c1._offsetStart.H=0;}}_d1=_bf;while(_d1){var _d5=_d1._parent;if(!_d5){break;}if(_d5.tagName=="TD"){var _d6=_d5._parent._parent._parent;if(_d5!=_d1._offsetParent&&_d5._offsetParent!=_d1._offsetParent){_d5=_d6;}}var _d7=_d1._offsetParent==_d5;for(var dir in _d1._offsetStart){var _d9=dir=="H"?"V":"H";if(rtl&&dir=="H"&&(_d5!=_c2)&&(_d5!=_c1)&&(dojo.isIE||dojo.isWebKit)&&_d5._clientSize.H>0&&_d5.scrollWidth>_d5._clientSize.H){var _da=_d5.scrollWidth-_d5._clientSize.H;if(_da>0){_d5._scrolledAmount.H-=_da;}}if(_d5._offsetParent.tagName=="TABLE"){if(dojo.isIE){_d5._offsetStart[dir]-=_d5._offsetParent._borderStart[dir];_d5._borderStart[dir]=_d5._borderSize[dir]=0;}else{_d5._offsetStart[dir]+=_d5._offsetParent._borderStart[dir];}}if(dojo.isIE){_d5._offsetStart[dir]+=_d5._offsetParent._borderStart[dir];}var _db=_d1._offsetStart[dir]-_d5._scrolledAmount[dir]-(_d7?0:_d5._offsetStart[dir])-_d5._borderStart[dir];var _dc=_db+_d1._offsetSize[dir]-_d5._offsetSize[dir]+_d5._borderSize[dir];var _dd=(dir=="H")?"scrollLeft":"scrollTop";var _de=dir=="H"&&rtl;var _df=_de?-_dc:_db;var _e0=_de?-_db:_dc;var _e1=(_df*_e0<=0)?0:Math[(_df<0)?"max":"min"](_df,_e0);if(_e1!=0){var _e2=_d5[_dd];_d5[_dd]+=(_de)?-_e1:_e1;var _e3=_d5[_dd]-_e2;}if(_d7){_d1._offsetStart[dir]+=_d5._offsetStart[dir];}_d1._offsetStart[dir]-=_d5[_dd];}_d1._parent=_d5._parent;_d1._offsetParent=_d5._offsetParent;}_d5=_bf;var _e4;while(_d5&&_d5.removeAttribute){_e4=_d5.parentNode;_d5.removeAttribute("_offsetParent");_d5.removeAttribute("_parent");_d5=_e4;}}catch(error){console.error("scrollIntoView: "+error);_bf.scrollIntoView(false);}};}if(!dojo._hasResource["dijit._base.sniff"]){dojo._hasResource["dijit._base.sniff"]=true;dojo.provide("dijit._base.sniff");(function(){var d=dojo,_e6=d.doc.documentElement,ie=d.isIE,_e8=d.isOpera,maj=Math.floor,ff=d.isFF,_eb=d.boxModel.replace(/-/,""),_ec={dj_ie:ie,dj_ie6:maj(ie)==6,dj_ie7:maj(ie)==7,dj_iequirks:ie&&d.isQuirks,dj_opera:_e8,dj_opera8:maj(_e8)==8,dj_opera9:maj(_e8)==9,dj_khtml:d.isKhtml,dj_webkit:d.isWebKit,dj_safari:d.isSafari,dj_gecko:d.isMozilla,dj_ff2:maj(ff)==2,dj_ff3:maj(ff)==3};_ec["dj_"+_eb]=true;for(var p in _ec){if(_ec[p]){if(_e6.className){_e6.className+=" "+p;}else{_e6.className=p;}}}dojo._loaders.unshift(function(){if(!dojo._isBodyLtr()){_e6.className+=" dijitRtl";for(var p in _ec){if(_ec[p]){_e6.className+=" "+p+"-rtl";}}}});})();}if(!dojo._hasResource["dijit._base.typematic"]){dojo._hasResource["dijit._base.typematic"]=true;dojo.provide("dijit._base.typematic");dijit.typematic={_fireEventAndReload:function(){this._timer=null;this._callback(++this._count,this._node,this._evt);this._currentTimeout=(this._currentTimeout<0)?this._initialDelay:((this._subsequentDelay>1)?this._subsequentDelay:Math.round(this._currentTimeout*this._subsequentDelay));this._timer=setTimeout(dojo.hitch(this,"_fireEventAndReload"),this._currentTimeout);},trigger:function(evt,_f0,_f1,_f2,obj,_f4,_f5){if(obj!=this._obj){this.stop();this._initialDelay=_f5||500;this._subsequentDelay=_f4||0.9;this._obj=obj;this._evt=evt;this._node=_f1;this._currentTimeout=-1;this._count=-1;this._callback=dojo.hitch(_f0,_f2);this._fireEventAndReload();}},stop:function(){if(this._timer){clearTimeout(this._timer);this._timer=null;}if(this._obj){this._callback(-1,this._node,this._evt);this._obj=null;}},addKeyListener:function(_f6,_f7,_f8,_f9,_fa,_fb){if(_f7.keyCode){_f7.charOrCode=_f7.keyCode;dojo.deprecated("keyCode attribute parameter for dijit.typematic.addKeyListener is deprecated. Use charOrCode instead.","","2.0");}else{if(_f7.charCode){_f7.charOrCode=String.fromCharCode(_f7.charCode);dojo.deprecated("charCode attribute parameter for dijit.typematic.addKeyListener is deprecated. Use charOrCode instead.","","2.0");}}return [dojo.connect(_f6,"onkeypress",this,function(evt){if(evt.charOrCode==_f7.charOrCode&&(_f7.ctrlKey===undefined||_f7.ctrlKey==evt.ctrlKey)&&(_f7.altKey===undefined||_f7.altKey==evt.ctrlKey)&&(_f7.shiftKey===undefined||_f7.shiftKey==evt.ctrlKey)){dojo.stopEvent(evt);dijit.typematic.trigger(_f7,_f8,_f6,_f9,_f7,_fa,_fb);}else{if(dijit.typematic._obj==_f7){dijit.typematic.stop();}}}),dojo.connect(_f6,"onkeyup",this,function(evt){if(dijit.typematic._obj==_f7){dijit.typematic.stop();}})];},addMouseListener:function(_fe,_ff,_100,_101,_102){var dc=dojo.connect;return [dc(_fe,"mousedown",this,function(evt){dojo.stopEvent(evt);dijit.typematic.trigger(evt,_ff,_fe,_100,_fe,_101,_102);}),dc(_fe,"mouseup",this,function(evt){dojo.stopEvent(evt);dijit.typematic.stop();}),dc(_fe,"mouseout",this,function(evt){dojo.stopEvent(evt);dijit.typematic.stop();}),dc(_fe,"mousemove",this,function(evt){dojo.stopEvent(evt);}),dc(_fe,"dblclick",this,function(evt){dojo.stopEvent(evt);if(dojo.isIE){dijit.typematic.trigger(evt,_ff,_fe,_100,_fe,_101,_102);setTimeout(dojo.hitch(this,dijit.typematic.stop),50);}})];},addListener:function(_109,_10a,_10b,_10c,_10d,_10e,_10f){return this.addKeyListener(_10a,_10b,_10c,_10d,_10e,_10f).concat(this.addMouseListener(_109,_10c,_10d,_10e,_10f));}};}if(!dojo._hasResource["dijit._base.wai"]){dojo._hasResource["dijit._base.wai"]=true;dojo.provide("dijit._base.wai");dijit.wai={onload:function(){var div=dojo.create("div",{id:"a11yTestNode",style:{cssText:"border: 1px solid;"+"border-color:red green;"+"position: absolute;"+"height: 5px;"+"top: -999px;"+"background-image: url(\""+(dojo.config.blankGif||dojo.moduleUrl("dojo","resources/blank.gif"))+"\");"}},dojo.body());var cs=dojo.getComputedStyle(div);if(cs){var _112=cs.backgroundImage;var _113=(cs.borderTopColor==cs.borderRightColor)||(_112!=null&&(_112=="none"||_112=="url(invalid-url:)"));dojo[_113?"addClass":"removeClass"](dojo.body(),"dijit_a11y");if(dojo.isIE){div.outerHTML="";}else{dojo.body().removeChild(div);}}}};if(dojo.isIE||dojo.isMoz){dojo._loaders.unshift(dijit.wai.onload);}dojo.mixin(dijit,{_XhtmlRoles:/banner|contentinfo|definition|main|navigation|search|note|secondary|seealso/,hasWaiRole:function(elem,role){var _116=this.getWaiRole(elem);return role?(_116.indexOf(role)>-1):(_116.length>0);},getWaiRole:function(elem){return dojo.trim((dojo.attr(elem,"role")||"").replace(this._XhtmlRoles,"").replace("wairole:",""));},setWaiRole:function(elem,role){var _11a=dojo.attr(elem,"role")||"";if(dojo.isFF<3||!this._XhtmlRoles.test(_11a)){dojo.attr(elem,"role",dojo.isFF<3?"wairole:"+role:role);}else{if((" "+_11a+" ").indexOf(" "+role+" ")<0){var _11b=dojo.trim(_11a.replace(this._XhtmlRoles,""));var _11c=dojo.trim(_11a.replace(_11b,""));dojo.attr(elem,"role",_11c+(_11c?" ":"")+role);}}},removeWaiRole:function(elem,role){var _11f=dojo.attr(elem,"role");if(!_11f){return;}if(role){var _120=dojo.isFF<3?"wairole:"+role:role;var t=dojo.trim((" "+_11f+" ").replace(" "+_120+" "," "));dojo.attr(elem,"role",t);}else{elem.removeAttribute("role");}},hasWaiState:function(elem,_123){if(dojo.isFF<3){return elem.hasAttributeNS("http://www.w3.org/2005/07/aaa",_123);}return elem.hasAttribute?elem.hasAttribute("aria-"+_123):!!elem.getAttribute("aria-"+_123);},getWaiState:function(elem,_125){if(dojo.isFF<3){return elem.getAttributeNS("http://www.w3.org/2005/07/aaa",_125);}return elem.getAttribute("aria-"+_125)||"";},setWaiState:function(elem,_127,_128){if(dojo.isFF<3){elem.setAttributeNS("http://www.w3.org/2005/07/aaa","aaa:"+_127,_128);}else{elem.setAttribute("aria-"+_127,_128);}},removeWaiState:function(elem,_12a){if(dojo.isFF<3){elem.removeAttributeNS("http://www.w3.org/2005/07/aaa",_12a);}else{elem.removeAttribute("aria-"+_12a);}}});}if(!dojo._hasResource["dijit._base"]){dojo._hasResource["dijit._base"]=true;dojo.provide("dijit._base");}if(!dojo._hasResource["dijit._Widget"]){dojo._hasResource["dijit._Widget"]=true;dojo.provide("dijit._Widget");dojo.require("dijit._base");dojo.connect(dojo,"connect",function(_12b,_12c){if(_12b&&dojo.isFunction(_12b._onConnect)){_12b._onConnect(_12c);}});dijit._connectOnUseEventHandler=function(_12d){};dijit._lastKeyDownNode=null;if(dojo.isIE){dojo.doc.attachEvent("onkeydown",function(evt){dijit._lastKeyDownNode=evt.srcElement;});}else{dojo.doc.addEventListener("keydown",function(evt){dijit._lastKeyDownNode=evt.target;},true);}(function(){var _130={};var _131=function(dc){if(!_130[dc]){var r=[];var _134;var _135=dojo.getObject(dc).prototype;for(var _136 in _135){if(dojo.isFunction(_135[_136])&&(_134=_136.match(/^_set([a-zA-Z]*)Attr$/))&&_134[1]){r.push(_134[1].charAt(0).toLowerCase()+_134[1].substr(1));}}_130[dc]=r;}return _130[dc]||[];};dojo.declare("dijit._Widget",null,{id:"",lang:"",dir:"","class":"",style:"",title:"",srcNodeRef:null,domNode:null,containerNode:null,attributeMap:{id:"",dir:"",lang:"","class":"",style:"",title:""},_deferredConnects:{onClick:"",onDblClick:"",onKeyDown:"",onKeyPress:"",onKeyUp:"",onMouseMove:"",onMouseDown:"",onMouseOut:"",onMouseOver:"",onMouseLeave:"",onMouseEnter:"",onMouseUp:""},onClick:dijit._connectOnUseEventHandler,onDblClick:dijit._connectOnUseEventHandler,onKeyDown:dijit._connectOnUseEventHandler,onKeyPress:dijit._connectOnUseEventHandler,onKeyUp:dijit._connectOnUseEventHandler,onMouseDown:dijit._connectOnUseEventHandler,onMouseMove:dijit._connectOnUseEventHandler,onMouseOut:dijit._connectOnUseEventHandler,onMouseOver:dijit._connectOnUseEventHandler,onMouseLeave:dijit._connectOnUseEventHandler,onMouseEnter:dijit._connectOnUseEventHandler,onMouseUp:dijit._connectOnUseEventHandler,_blankGif:(dojo.config.blankGif||dojo.moduleUrl("dojo","resources/blank.gif")),postscript:function(_137,_138){this.create(_137,_138);},create:function(_139,_13a){this.srcNodeRef=dojo.byId(_13a);this._connects=[];this._deferredConnects=dojo.clone(this._deferredConnects);for(var attr in this.attributeMap){delete this._deferredConnects[attr];}for(attr in this._deferredConnects){if(this[attr]!==dijit._connectOnUseEventHandler){delete this._deferredConnects[attr];}}if(this.srcNodeRef&&(typeof this.srcNodeRef.id=="string")){this.id=this.srcNodeRef.id;}if(_139){this.params=_139;dojo.mixin(this,_139);}this.postMixInProperties();if(!this.id){this.id=dijit.getUniqueId(this.declaredClass.replace(/\./g,"_"));}dijit.registry.add(this);this.buildRendering();if(this.domNode){this._applyAttributes();var _13c=this.srcNodeRef;if(_13c&&_13c.parentNode){_13c.parentNode.replaceChild(this.domNode,_13c);}for(attr in this.params){this._onConnect(attr);}}if(this.domNode){this.domNode.setAttribute("widgetId",this.id);}this.postCreate();if(this.srcNodeRef&&!this.srcNodeRef.parentNode){delete this.srcNodeRef;}this._created=true;},_applyAttributes:function(){var _13d=function(attr,_13f){if((_13f.params&&attr in _13f.params)||_13f[attr]){_13f.attr(attr,_13f[attr]);}};for(var attr in this.attributeMap){_13d(attr,this);}dojo.forEach(_131(this.declaredClass),function(a){if(!(a in this.attributeMap)){_13d(a,this);}},this);},postMixInProperties:function(){},buildRendering:function(){this.domNode=this.srcNodeRef||dojo.create("div");},postCreate:function(){},startup:function(){this._started=true;},destroyRecursive:function(_142){this.destroyDescendants(_142);this.destroy(_142);},destroy:function(_143){this.uninitialize();dojo.forEach(this._connects,function(_144){dojo.forEach(_144,dojo.disconnect);});dojo.forEach(this._supportingWidgets||[],function(w){if(w.destroy){w.destroy();}});this.destroyRendering(_143);dijit.registry.remove(this.id);},destroyRendering:function(_146){if(this.bgIframe){this.bgIframe.destroy(_146);delete this.bgIframe;}if(this.domNode){if(_146){dojo.removeAttr(this.domNode,"widgetId");}else{dojo.destroy(this.domNode);}delete this.domNode;}if(this.srcNodeRef){if(!_146){dojo.destroy(this.srcNodeRef);}delete this.srcNodeRef;}},destroyDescendants:function(_147){dojo.forEach(this.getChildren(),function(_148){if(_148.destroyRecursive){_148.destroyRecursive(_147);}});},uninitialize:function(){return false;},onFocus:function(){},onBlur:function(){},_onFocus:function(e){this.onFocus();},_onBlur:function(){this.onBlur();},_onConnect:function(_14a){if(_14a in this._deferredConnects){var _14b=this[this._deferredConnects[_14a]||"domNode"];this.connect(_14b,_14a.toLowerCase(),_14a);delete this._deferredConnects[_14a];}},_setClassAttr:function(_14c){var _14d=this[this.attributeMap["class"]||"domNode"];dojo.removeClass(_14d,this["class"]);this["class"]=_14c;dojo.addClass(_14d,_14c);},_setStyleAttr:function(_14e){var _14f=this[this.attributeMap["style"]||"domNode"];if(dojo.isObject(_14e)){dojo.style(_14f,_14e);}else{if(_14f.style.cssText){_14f.style.cssText+="; "+_14e;}else{_14f.style.cssText=_14e;}}this["style"]=_14e;},setAttribute:function(attr,_151){dojo.deprecated(this.declaredClass+"::setAttribute() is deprecated. Use attr() instead.","","2.0");this.attr(attr,_151);},_attrToDom:function(attr,_153){var _154=this.attributeMap[attr];dojo.forEach(dojo.isArray(_154)?_154:[_154],function(_155){var _156=this[_155.node||_155||"domNode"];var type=_155.type||"attribute";switch(type){case "attribute":if(dojo.isFunction(_153)){_153=dojo.hitch(this,_153);}if(/^on[A-Z][a-zA-Z]*$/.test(attr)){attr=attr.toLowerCase();}dojo.attr(_156,attr,_153);break;case "innerHTML":_156.innerHTML=_153;break;case "class":dojo.removeClass(_156,this[attr]);dojo.addClass(_156,_153);break;}},this);this[attr]=_153;},attr:function(name,_159){var args=arguments.length;if(args==1&&!dojo.isString(name)){for(var x in name){this.attr(x,name[x]);}return this;}var _15c=this._getAttrNames(name);if(args==2){if(this[_15c.s]){return this[_15c.s](_159)||this;}else{if(name in this.attributeMap){this._attrToDom(name,_159);}this[name]=_159;}return this;}else{if(this[_15c.g]){return this[_15c.g]();}else{return this[name];}}},_attrPairNames:{},_getAttrNames:function(name){var apn=this._attrPairNames;if(apn[name]){return apn[name];}var uc=name.charAt(0).toUpperCase()+name.substr(1);return apn[name]={n:name+"Node",s:"_set"+uc+"Attr",g:"_get"+uc+"Attr"};},toString:function(){return "[Widget "+this.declaredClass+", "+(this.id||"NO ID")+"]";},getDescendants:function(){if(this.containerNode){var list=dojo.query("[widgetId]",this.containerNode);return list.map(dijit.byNode);}else{return [];}},getChildren:function(){if(this.containerNode){return dijit.findWidgets(this.containerNode);}else{return [];}},nodesWithKeyClick:["input","button"],connect:function(obj,_162,_163){var d=dojo;var dc=dojo.connect;var _166=[];if(_162=="ondijitclick"){if(!this.nodesWithKeyClick[obj.tagName.toLowerCase()]){var m=d.hitch(this,_163);_166.push(dc(obj,"onkeydown",this,function(e){if((e.keyCode==d.keys.ENTER||e.keyCode==d.keys.SPACE)&&!e.ctrlKey&&!e.shiftKey&&!e.altKey&&!e.metaKey){dijit._lastKeyDownNode=e.target;d.stopEvent(e);}}),dc(obj,"onkeyup",this,function(e){if((e.keyCode==d.keys.ENTER||e.keyCode==d.keys.SPACE)&&e.target===dijit._lastKeyDownNode&&!e.ctrlKey&&!e.shiftKey&&!e.altKey&&!e.metaKey){dijit._lastKeyDownNode=null;return m(e);}}));}_162="onclick";}_166.push(dc(obj,_162,this,_163));this._connects.push(_166);return _166;},disconnect:function(_16a){for(var i=0;i<this._connects.length;i++){if(this._connects[i]==_16a){dojo.forEach(_16a,dojo.disconnect);this._connects.splice(i,1);return;}}},isLeftToRight:function(){return dojo._isBodyLtr();},isFocusable:function(){return this.focus&&(dojo.style(this.domNode,"display")!="none");},placeAt:function(_16c,_16d){if(_16c["declaredClass"]&&_16c["addChild"]){_16c.addChild(this,_16d);}else{dojo.place(this.domNode,_16c,_16d);}return this;}});})();}if(!dojo._hasResource["dijit._Templated"]){dojo._hasResource["dijit._Templated"]=true;dojo.provide("dijit._Templated");dojo.declare("dijit._Templated",null,{templateString:null,templatePath:null,widgetsInTemplate:false,_skipNodeCache:false,_stringRepl:function(tmpl){var _16f=this.declaredClass,_170=this;return dojo.string.substitute(tmpl,this,function(_171,key){if(key.charAt(0)=="!"){_171=dojo.getObject(key.substr(1),false,_170);}if(typeof _171=="undefined"){throw new Error(_16f+" template:"+key);}if(_171==null){return "";}return key.charAt(0)=="!"?_171:_171.toString().replace(/"/g,"&quot;");},this);},buildRendering:function(){var _173=dijit._Templated.getCachedTemplate(this.templatePath,this.templateString,this._skipNodeCache);var node;if(dojo.isString(_173)){node=dojo._toDom(this._stringRepl(_173));}else{node=_173.cloneNode(true);}this.domNode=node;this._attachTemplateNodes(node);if(this.widgetsInTemplate){var _175=dojo.parser,qry,attr;if(_175._query!="[dojoType]"){qry=_175._query;attr=_175._attrName;_175._query="[dojoType]";_175._attrName="dojoType";}var cw=(this._supportingWidgets=dojo.parser.parse(node));if(qry){_175._query=qry;_175._attrName=attr;}this._attachTemplateNodes(cw,function(n,p){return n[p];});}this._fillContent(this.srcNodeRef);},_fillContent:function(_17b){var dest=this.containerNode;if(_17b&&dest){while(_17b.hasChildNodes()){dest.appendChild(_17b.firstChild);}}},_attachTemplateNodes:function(_17d,_17e){_17e=_17e||function(n,p){return n.getAttribute(p);};var _181=dojo.isArray(_17d)?_17d:(_17d.all||_17d.getElementsByTagName("*"));var x=dojo.isArray(_17d)?0:-1;for(;x<_181.length;x++){var _183=(x==-1)?_17d:_181[x];if(this.widgetsInTemplate&&_17e(_183,"dojoType")){continue;}var _184=_17e(_183,"dojoAttachPoint");if(_184){var _185,_186=_184.split(/\s*,\s*/);while((_185=_186.shift())){if(dojo.isArray(this[_185])){this[_185].push(_183);}else{this[_185]=_183;}}}var _187=_17e(_183,"dojoAttachEvent");if(_187){var _188,_189=_187.split(/\s*,\s*/);var trim=dojo.trim;while((_188=_189.shift())){if(_188){var _18b=null;if(_188.indexOf(":")!=-1){var _18c=_188.split(":");_188=trim(_18c[0]);_18b=trim(_18c[1]);}else{_188=trim(_188);}if(!_18b){_18b=_188;}this.connect(_183,_188,_18b);}}}var role=_17e(_183,"waiRole");if(role){dijit.setWaiRole(_183,role);}var _18e=_17e(_183,"waiState");if(_18e){dojo.forEach(_18e.split(/\s*,\s*/),function(_18f){if(_18f.indexOf("-")!=-1){var pair=_18f.split("-");dijit.setWaiState(_183,pair[0],pair[1]);}});}}}});dijit._Templated._templateCache={};dijit._Templated.getCachedTemplate=function(_191,_192,_193){var _194=dijit._Templated._templateCache;var key=_192||_191;var _196=_194[key];if(_196){if(!_196.ownerDocument||_196.ownerDocument==dojo.doc){return _196;}dojo.destroy(_196);}if(!_192){_192=dijit._Templated._sanitizeTemplateString(dojo.trim(dojo._getText(_191)));}_192=dojo.string.trim(_192);if(_193||_192.match(/\$\{([^\}]+)\}/g)){return (_194[key]=_192);}else{return (_194[key]=dojo._toDom(_192));}};dijit._Templated._sanitizeTemplateString=function(_197){if(_197){_197=_197.replace(/^\s*<\?xml(\s)+version=[\'\"](\d)*.(\d)*[\'\"](\s)*\?>/im,"");var _198=_197.match(/<body[^>]*>\s*([\s\S]+)\s*<\/body>/im);if(_198){_197=_198[1];}}else{_197="";}return _197;};if(dojo.isIE){dojo.addOnWindowUnload(function(){var _199=dijit._Templated._templateCache;for(var key in _199){var _19b=_199[key];if(!isNaN(_19b.nodeType)){dojo.destroy(_19b);}delete _199[key];}});}dojo.extend(dijit._Widget,{dojoAttachEvent:"",dojoAttachPoint:"",waiRole:"",waiState:""});}if(!dojo._hasResource["com.ibm.widgets.TemplateCleaner"]){dojo._hasResource["com.ibm.widgets.TemplateCleaner"]=true;dojo.provide("com.ibm.widgets.TemplateCleaner");dojo.extend(dijit._Templated,{destroyRecursive:function(_19c){if(this._destroyed){return;}this.inherited("destroyRecursive",arguments);},destroy:function(_19d,_19e){if(this._destroyed){return;}this._destroyed=true;this.inherited("destroy",arguments);dojo.forEach(this._attachPoints,function(_19f){this[_19f]=null;},this);this._attachPoints=[];},_oldAttachTemplateNodes:dijit._Templated.prototype._attachTemplateNodes,_attachTemplateNodes:function(_1a0,func){var res=this._oldAttachTemplateNodes.apply(this,arguments);func=func?func:function(node,attr){return node.getAttribute(attr);};var _1a5=dojo.isArray(_1a0)?_1a0:[_1a0];if(!this._attachPoints){this._attachPoints=[];}dojo.forEach(_1a5,function(node){var _1a7=null;if(node.domNode){_1a7=[node];}else{_1a7=dojo.query("[dojoAttachPoint]",node);}dojo.forEach(_1a7,function(_1a8){var _1a9=func(_1a8,"dojoAttachPoint"),_1aa=_1a9.split(/\s*,\s*/mg);dojo.forEach(_1aa,function(_1ab){this._attachPoints.push(_1ab);},this);},this);},this);return res;}});}if(!dojo._hasResource["dijit._Container"]){dojo._hasResource["dijit._Container"]=true;dojo.provide("dijit._Container");dojo.declare("dijit._Container",null,{isContainer:true,buildRendering:function(){this.inherited(arguments);if(!this.containerNode){this.containerNode=this.domNode;}},addChild:function(_1ac,_1ad){var _1ae=this.containerNode;if(_1ad&&typeof _1ad=="number"){var _1af=this.getChildren();if(_1af&&_1af.length>=_1ad){_1ae=_1af[_1ad-1].domNode;_1ad="after";}}dojo.place(_1ac.domNode,_1ae,_1ad);if(this._started&&!_1ac._started){_1ac.startup();}},removeChild:function(_1b0){if(typeof _1b0=="number"&&_1b0>0){_1b0=this.getChildren()[_1b0];}if(!_1b0||!_1b0.domNode){return;}var node=_1b0.domNode;node.parentNode.removeChild(node);},_nextElement:function(node){do{node=node.nextSibling;}while(node&&node.nodeType!=1);return node;},_firstElement:function(node){node=node.firstChild;if(node&&node.nodeType!=1){node=this._nextElement(node);}return node;},getChildren:function(){return dojo.query("> [widgetId]",this.containerNode).map(dijit.byNode);},hasChildren:function(){return !!this._firstElement(this.containerNode);},destroyDescendants:function(_1b4){dojo.forEach(this.getChildren(),function(_1b5){_1b5.destroyRecursive(_1b4);});},_getSiblingOfChild:function(_1b6,dir){var node=_1b6.domNode;var _1b9=(dir>0?"nextSibling":"previousSibling");do{node=node[_1b9];}while(node&&(node.nodeType!=1||!dijit.byNode(node)));return node?dijit.byNode(node):null;},getIndexOfChild:function(_1ba){var _1bb=this.getChildren();for(var i=0,c;c=_1bb[i];i++){if(c==_1ba){return i;}}return -1;}});}if(!dojo._hasResource["dijit._Contained"]){dojo._hasResource["dijit._Contained"]=true;dojo.provide("dijit._Contained");dojo.declare("dijit._Contained",null,{getParent:function(){for(var p=this.domNode.parentNode;p;p=p.parentNode){var id=p.getAttribute&&p.getAttribute("widgetId");if(id){var _1c0=dijit.byId(id);return _1c0.isContainer?_1c0:null;}}return null;},_getSibling:function(_1c1){var node=this.domNode;do{node=node[_1c1+"Sibling"];}while(node&&node.nodeType!=1);if(!node){return null;}var id=node.getAttribute("widgetId");return dijit.byId(id);},getPreviousSibling:function(){return this._getSibling("previous");},getNextSibling:function(){return this._getSibling("next");},getIndexInParent:function(){var p=this.getParent();if(!p||!p.getIndexOfChild){return -1;}return p.getIndexOfChild(this);}});}if(!dojo._hasResource["dijit.layout._LayoutWidget"]){dojo._hasResource["dijit.layout._LayoutWidget"]=true;dojo.provide("dijit.layout._LayoutWidget");dojo.declare("dijit.layout._LayoutWidget",[dijit._Widget,dijit._Container,dijit._Contained],{baseClass:"dijitLayoutContainer",isLayoutContainer:true,postCreate:function(){dojo.addClass(this.domNode,"dijitContainer");dojo.addClass(this.domNode,this.baseClass);},startup:function(){if(this._started){return;}dojo.forEach(this.getChildren(),function(_1c5){_1c5.startup();});if(!this.getParent||!this.getParent()){this.resize();this._viewport=dijit.getViewport();this.connect(dojo.global,"onresize",function(){var _1c6=dijit.getViewport();if(_1c6.w!=this._viewport.w||_1c6.h!=this._viewport.h){this._viewport=_1c6;this.resize();}});}this.inherited(arguments);},resize:function(_1c7,_1c8){var node=this.domNode;if(_1c7){dojo.marginBox(node,_1c7);if(_1c7.t){node.style.top=_1c7.t+"px";}if(_1c7.l){node.style.left=_1c7.l+"px";}}var mb=_1c8||{};dojo.mixin(mb,_1c7||{});if(!("h" in mb)||!("w" in mb)){mb=dojo.mixin(dojo.marginBox(node),mb);}var cs=dojo.getComputedStyle(node);var me=dojo._getMarginExtents(node,cs);var be=dojo._getBorderExtents(node,cs);var bb=(this._borderBox={w:mb.w-(me.w+be.w),h:mb.h-(me.h+be.h)});var pe=dojo._getPadExtents(node,cs);this._contentBox={l:dojo._toPixelValue(node,cs.paddingLeft),t:dojo._toPixelValue(node,cs.paddingTop),w:bb.w-pe.w,h:bb.h-pe.h};this.layout();},layout:function(){},_setupChild:function(_1d0){dojo.addClass(_1d0.domNode,this.baseClass+"-child");if(_1d0.baseClass){dojo.addClass(_1d0.domNode,this.baseClass+"-"+_1d0.baseClass);}},addChild:function(_1d1,_1d2){this.inherited(arguments);if(this._started){this._setupChild(_1d1);}},removeChild:function(_1d3){dojo.removeClass(_1d3.domNode,this.baseClass+"-child");if(_1d3.baseClass){dojo.removeClass(_1d3.domNode,this.baseClass+"-"+_1d3.baseClass);}this.inherited(arguments);}});dijit.layout.marginBox2contentBox=function(node,mb){var cs=dojo.getComputedStyle(node);var me=dojo._getMarginExtents(node,cs);var pb=dojo._getPadBorderExtents(node,cs);return {l:dojo._toPixelValue(node,cs.paddingLeft),t:dojo._toPixelValue(node,cs.paddingTop),w:mb.w-(me.w+pb.w),h:mb.h-(me.h+pb.h)};};(function(){var _1d9=function(word){return word.substring(0,1).toUpperCase()+word.substring(1);};var size=function(_1dc,dim){_1dc.resize?_1dc.resize(dim):dojo.marginBox(_1dc.domNode,dim);dojo.mixin(_1dc,dojo.marginBox(_1dc.domNode));dojo.mixin(_1dc,dim);};dijit.layout.layoutChildren=function(_1de,dim,_1e0){dim=dojo.mixin({},dim);dojo.addClass(_1de,"dijitLayoutContainer");_1e0=dojo.filter(_1e0,function(item){return item.layoutAlign!="client";}).concat(dojo.filter(_1e0,function(item){return item.layoutAlign=="client";}));dojo.forEach(_1e0,function(_1e3){var elm=_1e3.domNode,pos=_1e3.layoutAlign;var _1e6=elm.style;_1e6.left=dim.l+"px";_1e6.top=dim.t+"px";_1e6.bottom=_1e6.right="auto";dojo.addClass(elm,"dijitAlign"+_1d9(pos));if(pos=="top"||pos=="bottom"){size(_1e3,{w:dim.w});dim.h-=_1e3.h;if(pos=="top"){dim.t+=_1e3.h;}else{_1e6.top=dim.t+dim.h+"px";}}else{if(pos=="left"||pos=="right"){size(_1e3,{h:dim.h});dim.w-=_1e3.w;if(pos=="left"){dim.l+=_1e3.w;}else{_1e6.left=dim.l+dim.w+"px";}}else{if(pos=="client"){size(_1e3,dim);}}}});};})();}if(!dojo._hasResource["dijit.form._FormWidget"]){dojo._hasResource["dijit.form._FormWidget"]=true;dojo.provide("dijit.form._FormWidget");dojo.declare("dijit.form._FormWidget",[dijit._Widget,dijit._Templated],{baseClass:"",name:"",alt:"",value:"",type:"text",tabIndex:"0",disabled:false,readOnly:false,intermediateChanges:false,scrollOnFocus:true,attributeMap:dojo.delegate(dijit._Widget.prototype.attributeMap,{value:"focusNode",disabled:"focusNode",readOnly:"focusNode",id:"focusNode",tabIndex:"focusNode",alt:"focusNode"}),postMixInProperties:function(){this.nameAttrSetting=this.name?("name='"+this.name+"'"):"";this.inherited(arguments);},_setDisabledAttr:function(_1e7){this.disabled=_1e7;dojo.attr(this.focusNode,"disabled",_1e7);dijit.setWaiState(this.focusNode,"disabled",_1e7);if(_1e7){this._hovering=false;this._active=false;this.focusNode.removeAttribute("tabIndex");}else{this.focusNode.setAttribute("tabIndex",this.tabIndex);}this._setStateClass();},setDisabled:function(_1e8){dojo.deprecated("setDisabled("+_1e8+") is deprecated. Use attr('disabled',"+_1e8+") instead.","","2.0");this.attr("disabled",_1e8);},_onFocus:function(e){if(this.scrollOnFocus){dijit.scrollIntoView(this.domNode);}this.inherited(arguments);},_onMouse:function(_1ea){var _1eb=_1ea.currentTarget;if(_1eb&&_1eb.getAttribute){this.stateModifier=_1eb.getAttribute("stateModifier")||"";}if(!this.disabled){switch(_1ea.type){case "mouseenter":case "mouseover":this._hovering=true;this._active=this._mouseDown;break;case "mouseout":case "mouseleave":this._hovering=false;this._active=false;break;case "mousedown":this._active=true;this._mouseDown=true;var _1ec=this.connect(dojo.body(),"onmouseup",function(){if(this._mouseDown&&this.isFocusable()){this.focus();}this._active=false;this._mouseDown=false;this._setStateClass();this.disconnect(_1ec);});break;}this._setStateClass();}},isFocusable:function(){return !this.disabled&&!this.readOnly&&this.focusNode&&(dojo.style(this.domNode,"display")!="none");},focus:function(){dijit.focus(this.focusNode);},_setStateClass:function(){var _1ed=this.baseClass.split(" ");function multiply(_1ee){_1ed=_1ed.concat(dojo.map(_1ed,function(c){return c+_1ee;}),"dijit"+_1ee);};if(this.checked){multiply("Checked");}if(this.state){multiply(this.state);}if(this.selected){multiply("Selected");}if(this.disabled){multiply("Disabled");}else{if(this.readOnly){multiply("ReadOnly");}else{if(this._active){multiply(this.stateModifier+"Active");}else{if(this._focused){multiply("Focused");}if(this._hovering){multiply(this.stateModifier+"Hover");}}}}var tn=this.stateNode||this.domNode,_1f1={};dojo.forEach(tn.className.split(" "),function(c){_1f1[c]=true;});if("_stateClasses" in this){dojo.forEach(this._stateClasses,function(c){delete _1f1[c];});}dojo.forEach(_1ed,function(c){_1f1[c]=true;});var _1f5=[];for(var c in _1f1){_1f5.push(c);}tn.className=_1f5.join(" ");this._stateClasses=_1ed;},compare:function(val1,val2){if((typeof val1=="number")&&(typeof val2=="number")){return (isNaN(val1)&&isNaN(val2))?0:(val1-val2);}else{if(val1>val2){return 1;}else{if(val1<val2){return -1;}else{return 0;}}}},onChange:function(_1f9){},_onChangeActive:false,_handleOnChange:function(_1fa,_1fb){this._lastValue=_1fa;if(this._lastValueReported==undefined&&(_1fb===null||!this._onChangeActive)){this._resetValue=this._lastValueReported=_1fa;}if((this.intermediateChanges||_1fb||_1fb===undefined)&&((typeof _1fa!=typeof this._lastValueReported)||this.compare(_1fa,this._lastValueReported)!=0)){this._lastValueReported=_1fa;if(this._onChangeActive){this.onChange(_1fa);}}},create:function(){this.inherited(arguments);this._onChangeActive=true;this._setStateClass();},destroy:function(){if(this._layoutHackHandle){clearTimeout(this._layoutHackHandle);}this.inherited(arguments);},setValue:function(_1fc){dojo.deprecated("dijit.form._FormWidget:setValue("+_1fc+") is deprecated.  Use attr('value',"+_1fc+") instead.","","2.0");this.attr("value",_1fc);},getValue:function(){dojo.deprecated(this.declaredClass+"::getValue() is deprecated. Use attr('value') instead.","","2.0");return this.attr("value");},_layoutHack:function(){if(dojo.isFF==2&&!this._layoutHackHandle){var node=this.domNode;var old=node.style.opacity;node.style.opacity="0.999";this._layoutHackHandle=setTimeout(dojo.hitch(this,function(){this._layoutHackHandle=null;node.style.opacity=old;}),0);}}});dojo.declare("dijit.form._FormValueWidget",dijit.form._FormWidget,{attributeMap:dojo.delegate(dijit.form._FormWidget.prototype.attributeMap,{value:""}),postCreate:function(){if(dojo.isIE||dojo.isWebKit){this.connect(this.focusNode||this.domNode,"onkeydown",this._onKeyDown);}if(this._resetValue===undefined){this._resetValue=this.value;}},_setValueAttr:function(_1ff,_200){this.value=_1ff;this._handleOnChange(_1ff,_200);},_getValueAttr:function(_201){return this._lastValue;},undo:function(){this._setValueAttr(this._lastValueReported,false);},reset:function(){this._hasBeenBlurred=false;this._setValueAttr(this._resetValue,true);},_onKeyDown:function(e){if(e.keyCode==dojo.keys.ESCAPE&&!e.ctrlKey&&!e.altKey){var te;if(dojo.isIE){e.preventDefault();te=document.createEventObject();te.keyCode=dojo.keys.ESCAPE;te.shiftKey=e.shiftKey;e.srcElement.fireEvent("onkeypress",te);}else{if(dojo.isWebKit){te=document.createEvent("Events");te.initEvent("keypress",true,true);te.keyCode=dojo.keys.ESCAPE;te.shiftKey=e.shiftKey;e.target.dispatchEvent(te);}}}}});}if(!dojo._hasResource["dijit.dijit"]){dojo._hasResource["dijit.dijit"]=true;dojo.provide("dijit.dijit");}
/** Licensed Materials - Property of IBM, 5724-E76 and 5724-E77, (C) Copyright IBM Corp. 2009 - All Rights reserved.  **/
dojo.provide("ibm.ibmClientModel");if(!dojo._hasResource["dojox.xml.parser"]){dojo._hasResource["dojox.xml.parser"]=true;dojo.provide("dojox.xml.parser");dojox.xml.parser.parse=function(_1,_2){var _3=dojo.doc;var _4;_2=_2||"text/xml";if(_1&&dojo.trim(_1)&&"DOMParser" in dojo.global){var _5=new DOMParser();_4=_5.parseFromString(_1,_2);var de=_4.documentElement;var _7="http://www.mozilla.org/newlayout/xml/parsererror.xml";if(de.nodeName=="parsererror"&&de.namespaceURI==_7){var _8=de.getElementsByTagNameNS(_7,"sourcetext")[0];if(!_8){_8=_8.firstChild.data;}throw new Error("Error parsing text "+nativeDoc.documentElement.firstChild.data+" \n"+_8);}return _4;}else{if("ActiveXObject" in dojo.global){var ms=function(n){return "MSXML"+n+".DOMDocument";};var dp=["Microsoft.XMLDOM",ms(6),ms(4),ms(3),ms(2)];dojo.some(dp,function(p){try{_4=new ActiveXObject(p);}catch(e){return false;}return true;});if(_1&&_4){_4.async=false;_4.loadXML(_1);var pe=_4.parseError;if(pe.errorCode!==0){throw new Error("Line: "+pe.line+"\n"+"Col: "+pe.linepos+"\n"+"Reason: "+pe.reason+"\n"+"Error Code: "+pe.errorCode+"\n"+"Source: "+pe.srcText);}}if(_4){return _4;}}else{if(_3.implementation&&_3.implementation.createDocument){if(_1&&dojo.trim(_1)&&_3.createElement){var _e=_3.createElement("xml");_e.innerHTML=_1;var _f=_3.implementation.createDocument("foo","",null);dojo.forEach(_e.childNodes,function(_10){_f.importNode(_10,true);});return _f;}else{return _3.implementation.createDocument("","",null);}}}}return null;};dojox.xml.parser.textContent=function(_11,_12){if(arguments.length>1){var _13=_11.ownerDocument||dojo.doc;dojox.xml.parser.replaceChildren(_11,_13.createTextNode(_12));return _12;}else{if(_11.textContent!==undefined){return _11.textContent;}var _14="";if(_11){dojo.forEach(_11.childNodes,function(_15){switch(_15.nodeType){case 1:case 5:_14+=dojox.xml.parser.textContent(_15);break;case 3:case 2:case 4:_14+=_15.nodeValue;}});}return _14;}};dojox.xml.parser.replaceChildren=function(_16,_17){var _18=[];if(dojo.isIE){dojo.forEach(_16.childNodes,function(_19){_18.push(_19);});}dojox.xml.parser.removeChildren(_16);dojo.forEach(_18,dojo.destroy);if(!dojo.isArray(_17)){_16.appendChild(_17);}else{dojo.forEach(_17,function(_1a){_16.appendChild(_1a);});}};dojox.xml.parser.removeChildren=function(_1b){var _1c=_1b.childNodes.length;while(_1b.hasChildNodes()){_1b.removeChild(_1b.firstChild);}return _1c;};dojox.xml.parser.innerXML=function(_1d){if(_1d.innerXML){return _1d.innerXML;}else{if(_1d.xml){return _1d.xml;}else{if(typeof XMLSerializer!="undefined"){return (new XMLSerializer()).serializeToString(_1d);}}}return null;};}if(!dojo._hasResource["dojox.data.dom"]){dojo._hasResource["dojox.data.dom"]=true;dojo.provide("dojox.data.dom");dojo.deprecated("dojox.data.dom","Use dojox.xml.parser instead.","2.0");dojox.data.dom.createDocument=function(str,_1f){dojo.deprecated("dojox.data.dom.createDocument()","Use dojox.xml.parser.parse() instead.","2.0");try{return dojox.xml.parser.parse(str,_1f);}catch(e){return null;}};dojox.data.dom.textContent=function(_20,_21){dojo.deprecated("dojox.data.dom.textContent()","Use dojox.xml.parser.textContent() instead.","2.0");if(arguments.length>1){return dojox.xml.parser.textContent(_20,_21);}else{return dojox.xml.parser.textContent(_20);}};dojox.data.dom.replaceChildren=function(_22,_23){dojo.deprecated("dojox.data.dom.replaceChildren()","Use dojox.xml.parser.replaceChildren() instead.","2.0");dojox.xml.parser.replaceChildren(_22,_23);};dojox.data.dom.removeChildren=function(_24){dojo.deprecated("dojox.data.dom.removeChildren()","Use dojox.xml.parser.removeChildren() instead.","2.0");return dojox.xml.parser.removeChildren(_24);};dojox.data.dom.innerXML=function(_25){dojo.deprecated("dojox.data.dom.innerXML()","Use dojox.xml.parser.innerXML() instead.","2.0");return dojox.xml.parser.innerXML(_25);};}if(!dojo._hasResource["com.ibm.portal.xpath"]){dojo._hasResource["com.ibm.portal.xpath"]=true;dojo.provide("com.ibm.portal.xpath");com.ibm.portal.xpath.evaluateXPath=function(_26,doc,_28){if(typeof ActiveXObject!="undefined"){return com.ibm.portal.xpath.ie.evaluateXPath(_26,doc,_28);}else{return com.ibm.portal.xpath.gecko.evaluateXPath(_26,doc,_28);}};dojo.provide("com.ibm.portal.xpath.ie");com.ibm.portal.xpath.ie.evaluateXPath=function(_29,doc,_2b){if(_2b){var ns="";for(var _2d in _2b){ns+="xmlns:"+_2d+"='"+_2b[_2d]+"' ";}if(doc.ownerDocument){doc.ownerDocument.setProperty("SelectionNamespaces",ns);}else{doc.setProperty("SelectionNamespaces",ns);}}var _2e=doc.selectNodes(_29);var _2f;var _30=new Array();var len=0;for(var i=0;i<_2e.length;i++){_2f=_2e[i];if(_2f){_30[len]=_2f;len++;}}return _30;};dojo.provide("com.ibm.portal.xpath.gecko");com.ibm.portal.xpath.gecko.evaluateXPath=function(_33,doc,_35){var _36;try{var _37=doc;if(!_37.evaluate){_37=doc.ownerDocument;}_36=_37.evaluate(_33,doc,function(_38){return _35[_38]||null;},XPathResult.ANY_TYPE,null);}catch(exc){throw new Error("Error with xpath expression"+exc);}var _39;var _3a=new Array();var len=0;do{_39=_36.iterateNext();if(_39){_3a[len]=_39;len++;}}while(_39);return _3a;};}if(!dojo._hasResource["ibm.portal.xml.xpath"]){dojo._hasResource["ibm.portal.xml.xpath"]=true;dojo.provide("ibm.portal.xml.xpath");dojo.require("com.ibm.portal.xpath");ibm.portal.xml.xpath.evaluateXPath=function(_3c,doc,_3e){return com.ibm.portal.xpath.evaluateXPath(_3c,doc,_3e);};dojo.provide("ibm.portal.xml.xpath.ie");ibm.portal.xml.xpath.ie.evaluateXPath=function(_3f,doc,_41){return com.ibm.portal.xpath.ie.evaluateXPath(_3f,doc,_41);};dojo.provide("ibm.portal.xml.xpath.gecko");ibm.portal.xml.xpath.gecko.evaluateXPath=function(_42,doc,_44){return com.ibm.portal.xpath.gecko.evaluateXPath(_42,doc,_44);};}if(!dojo._hasResource["com.ibm.portal.xslt"]){dojo._hasResource["com.ibm.portal.xslt"]=true;dojo.provide("com.ibm.portal.xslt");dojo.require("dojox.data.dom");dojo.declare("com.ibm.portal.xslt.TransformerFactory",null,{constructor:function(){this._xsltMap=new Array();},newTransformer:function(_45){ibm.portal.debug.entry("newTransformer",[_45]);var trf=this._getCached(_45);if(trf==null){trf=new com.ibm.portal.xslt.Transformer(_45);this._xsltMap.push({url:_45,transformer:trf});}return trf;},_getCached:function(_47){var _48=null;for(i=0;i<this._xsltMap.length;i++){var _49=this._xsltMap[i];if(_47==_49.url){_48=_49.transformer;break;}}return _48;}});dojo.declare("com.ibm.portal.xslt.Transformer",null,{constructor:function(_4a){this._xslt=com.ibm.portal.xslt.loadXsl(_4a);},transformToRegion:function(_4b,_4c,_4d,doc){if(dojo.isIE){var _4f=com.ibm.portal.xslt.transform(_4b,this._xslt,null,_4c,true);_4d.innerHTML=dojo.string.trim(_4f);}else{var _50=com.ibm.portal.xslt.gecko._transformToFragment(_4b,this._xslt,null,_4c,doc);_4d.innerHTML="";_4d.appendChild(_50);}},transformToDocument:function(_51,_52,_53){var _54=com.ibm.portal.xslt.transform(_51,this._xslt,null,_52,_53);return _54;}});com.ibm.portal.xslt.TRANSFORMER_FACTORY=new com.ibm.portal.xslt.TransformerFactory();com.ibm.portal.xslt.ie={};com.ibm.portal.xslt.gecko={};com.ibm.portal.xslt.getXmlHttpRequest=function(){var _55=null;if(typeof ActiveXObject!="undefined"){_55=new ActiveXObject("Microsoft.XMLHTTP");}else{_55=new XMLHttpRequest();}return _55;};com.ibm.portal.xslt.loadXml=function(_56){if(typeof ActiveXObject!="undefined"){return com.ibm.portal.xslt.ie.loadXml(_56);}else{return com.ibm.portal.xslt.gecko.loadXml(_56);}};com.ibm.portal.xslt.loadXmlString=function(_57){if(typeof ActiveXObject!="undefined"){return com.ibm.portal.xslt.ie.loadXmlString(_57);}else{return com.ibm.portal.xslt.gecko.loadXmlString(_57);}};com.ibm.portal.xslt.loadXsl=function(_58){if(typeof ActiveXObject!="undefined"){return com.ibm.portal.xslt.ie.loadXsl(_58);}else{return com.ibm.portal.xslt.gecko.loadXsl(_58);}};com.ibm.portal.xslt.transform=function(xml,xsl,_5b,_5c,_5d){ibm.portal.debug.entry("transform",[xml,xsl,_5b,_5c,_5d]);if(typeof ActiveXObject!="undefined"){return com.ibm.portal.xslt.ie.transform(xml,xsl,_5b,_5c,_5d);}else{return com.ibm.portal.xslt.gecko.transform(xml,xsl,_5b,_5c,_5d);}};com.ibm.portal.xslt.transformAndUpdate=function(_5e,xml,xsl,_61,_62){ibm.portal.debug.entry("transformAndUpdate",[_5e,xml,xsl,_61,_62]);if(typeof ActiveXObject!="undefined"){var _63=com.ibm.portal.xslt.transform(xml,xsl,_61,_62,true);_5e.innerHTML=dojo.string.trim(_63);}else{var doc=_5e.ownerDocument?_5e.ownerDocument:document;var _65=com.ibm.portal.xslt.gecko._transformToFragment(xml,xsl,_61,_62,doc);_5e.innerHTML="";_5e.appendChild(_65);}ibm.portal.debug.exit("transformAndUpdate");};com.ibm.portal.xslt.ie.XSLT_PROG_IDS=["Msxml2.XSLTemplate.6.0","Msxml2.XSLTemplate.4.0","MSXML2.XSLTemplate.3.0","MSXML2.XSLTemplate"];com.ibm.portal.xslt.ie.DOM_PROG_IDS=["Msxml2.DOMDocument.6.0","Msxml2.DOMDocument.4.0","MSXML2.DOMDocument","MSXML.DOMDocument","Microsoft.XMLDOM"];com.ibm.portal.xslt.ie.FTDOM_PROG_IDS=["Msxml2.FreeThreadedDOMDocument.6.0","Msxml2.FreeThreadedDOMDocument.4.0","MSXML2.FreeThreadedDOMDocument","MSXML.FreeThreadedDOMDocument","Microsoft.FreeThreadedXMLDOM"];com.ibm.portal.xslt.ie._getMSXMLImpl=function(_66){while(_66.length>0){try{var _67=new ActiveXObject(_66[0]);if(_67){return _67;}}catch(err){}_66.splice(0,1);}throw new Error("No MSXML implementation exists");};com.ibm.portal.xslt.ie.loadXml=function(_68){var _69=this._getMSXMLImpl(this.DOM_PROG_IDS);_69.async=0;_69.resolveExternals=0;if(!_69.load(_68)){throw new Error("Error loading xml file "+_68);}return _69;};com.ibm.portal.xslt.ie.loadXmlString=function(_6a){var _6b=this._getMSXMLImpl(this.DOM_PROG_IDS);_6b.async=0;_6b.resolveExternals=0;if(!_6b.loadXML(_6a)){throw new Error("Error loading xml string "+_6a);}return _6b;};com.ibm.portal.xslt.ie.loadXsl=function(_6c){var _6d=this._getMSXMLImpl(this.FTDOM_PROG_IDS);_6d.async=0;_6d.resolveExternals=0;if(!_6d.load(_6c)){throw new Error("Error loading xsl file "+_6c);}return _6d;};com.ibm.portal.xslt.ie.transform=function(_6e,xsl,_70,_71,_72){var _73=_6e;var _74=xsl;try{if(!_74.documentElement){_74=this.loadXsl(xsl);}}catch(e){var _75=e.message;throw new Error(""+_75,""+_75);}var _76=this._getMSXMLImpl(this.XSLT_PROG_IDS);_76.stylesheet=_74;var _77=_76.createProcessor();_77.input=_73;if(_71){for(var p in _71){_77.addParameter(p,_71[p]);}}if(_70){_77.addParameter("mode",_70);}if(_72){if(!_77.transform()){throw new Error("Error transforming xml doc "+_73);}return _77.output;}else{var _79=this._getMSXMLImpl(this.DOM_PROG_IDS);_79.async=false;_79.validateOnParse=false;_73.transformNodeToObject(_74,_79);return _79;}};com.ibm.portal.xslt.gecko.loadXml=function(_7a){var _7b=null;if(dojo.isSafari){var xhr=new XMLHttpRequest();xhr.open("GET",_7a,false);xhr.send(null);if(xhr.status==200){_7b=xhr.responseXML;}}else{_7b=document.implementation.createDocument("","",null);_7b.async=0;_7b.load(_7a);}return _7b;};com.ibm.portal.xslt.gecko.loadXmlString=function(_7d){var _7e=new DOMParser();try{oXmlDoc=_7e.parseFromString(_7d,"text/xml");}catch(exc){throw new Error("Error loading xml string "+_7d);}return oXmlDoc;};com.ibm.portal.xslt.gecko.loadXsl=function(_7f){var _80=null;if(dojo.isSafari){var xhr=new XMLHttpRequest();xhr.open("GET",_7f,false);xhr.send(null);if(xhr.status==200){_80=xhr.responseXML;}}else{_80=document.implementation.createDocument("","",null);_80.async=0;_80.load(_7f);}return _80;};com.ibm.portal.xslt.gecko._getXSLTProc=function(_82,xsl,_84,_85){var _86=xsl;if(!_86.documentElement){_86=this.loadXsl(xsl);}var _87=new XSLTProcessor();_87.importStylesheet(_86);if(_85){for(var p in _85){_87.setParameter(null,p,_85[p]);}}if(_84){_87.setParameter(null,"mode",_84);}return _87;};com.ibm.portal.xslt.gecko._transformToFragment=function(_89,xsl,_8b,_8c,doc){var _8e=com.ibm.portal.xslt.gecko._getXSLTProc(_89,xsl,_8b,_8c);var _8f=null;_8f=_8e.transformToFragment(_89,doc);_8e.clearParameters();return _8f;};com.ibm.portal.xslt.gecko.transform=function(_90,xsl,_92,_93,_94){try{var _95=null;if(!_94){var _96=com.ibm.portal.xslt.gecko._getXSLTProc(_90,xsl,_92,_93);_95=_96.transformToDocument(_90);return _95;}else{_95=com.ibm.portal.xslt.gecko._transformToFragment(_90,xsl,_92,_93,document);}var _97=new XMLSerializer();var _98=dojo.string.trim(_97.serializeToString(_95));if(dojo.isOpera&&_95.firstChild&&_95.firstChild.nodeName=="result"){var _99=_98.indexOf("<result>")+8;var end=_98.lastIndexOf("</result>");_98=dojo.string.trim(_98.substring(_99,end));}return _98;}catch(exc){throw new Error("Error transforming xml doc "+exc);}};com.ibm.portal.xslt.setLayerContentByXml=function(_9b,xml,xsl,_9e,_9f){var _a0=com.ibm.portal.xslt.transform(xml,xsl,null,_9e,_9f);if(_9b.innerHTML){_9b.innerHTML=_a0;}else{var obj=document.getElementById(_9b);obj.innerHTML=_a0;}};}if(!dojo._hasResource["ibm.portal.xml.xslt"]){dojo._hasResource["ibm.portal.xml.xslt"]=true;dojo.provide("ibm.portal.xml.xslt");dojo.require("com.ibm.portal.xslt");ibm.portal.xml.xslt.ie={};ibm.portal.xml.xslt.gecko={};ibm.portal.xml.xslt.getXmlHttpRequest=function(){return com.ibm.portal.xslt.getXmlHttpRequest();};ibm.portal.xml.xslt.loadXml=function(_a2){return com.ibm.portal.xslt.loadXml(_a2);};ibm.portal.xml.xslt.loadXmlString=function(_a3){return com.ibm.portal.xslt.loadXmlString(_a3);};ibm.portal.xml.xslt.loadXsl=function(_a4){return com.ibm.portal.xslt.loadXsl(_a4);};ibm.portal.xml.xslt.transform=function(xml,xsl,_a7,_a8,_a9){ibm.portal.debug.entry("transform",[xml,xsl,_a7,_a8,_a9]);return com.ibm.portal.xslt.transform(xml,xsl,_a7,_a8,_a9);};ibm.portal.xml.xslt.transformAndUpdate=function(_aa,xml,xsl,_ad,_ae){ibm.portal.debug.entry("transformAndUpdate",[_aa,xml,xsl,_ad,_ae]);com.ibm.portal.xslt.transformAndUpdate(_aa,xml,xsl,_ad,_ae);ibm.portal.debug.exit("transformAndUpdate");};ibm.portal.xml.xslt.ie.loadXml=function(_af){return com.ibm.portal.xslt.ie.loadXml(_af);};ibm.portal.xml.xslt.ie.loadXmlString=function(_b0){return com.ibm.portal.xslt.ie.loadXmlString(_b0);};ibm.portal.xml.xslt.ie.loadXsl=function(_b1){return com.ibm.portal.xslt.ie.loadXsl(_b1);};ibm.portal.xml.xslt.ie.transform=function(_b2,xsl,_b4,_b5,_b6){return com.ibm.portal.xslt.ie.transform(_b2,xsl,_b4,_b5,_b6);};ibm.portal.xml.xslt.gecko.loadXml=function(_b7){return com.ibm.portal.xslt.gecko.loadXml(_b7);};ibm.portal.xml.xslt.gecko.loadXmlString=function(_b8){return com.ibm.portal.xslt.gecko.loadXmlString(_b8);};ibm.portal.xml.xslt.gecko.loadXsl=function(_b9){return com.ibm.portal.xslt.gecko.loadXsl(_b9);};ibm.portal.xml.xslt.gecko.transform=function(_ba,xsl,_bc,_bd,_be){return com.ibm.portal.xslt.gecko.transform(_ba,xsl,_bc,_bd,_be);};ibm.portal.xml.xslt.setLayerContentByXml=function(_bf,xml,xsl,_c2,_c3){com.ibm.portal.xslt.setLayerContentByXml(_bf,xml,xsl,_c2,_c3);};}if(!dojo._hasResource["com.ibm.portal.state"]){dojo._hasResource["com.ibm.portal.state"]=true;dojo.provide("com.ibm.portal.state");dojo.declare("com.ibm.portal.state.StateManager",null,{constructor:function(_c4){this.stateDOM=null;this.stateNode=null;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};this.serializationManager=new com.ibm.portal.state.SerializationManager(_c4);},getState:function(){return this.stateDOM;},newState:function(_c5,_c6,_c7){var _c8=null;if(_c5==null){_c8=dojox.data.dom.createDocument();}else{if(_c6==null){_c8=dojox.data.dom.createDocument(dojox.data.dom.innerXML(_c5));}else{var _c9=com.ibm.portal.xslt;var _ca=_c9.transform(_c5,_c6,null,_c7,true);_c8=dojox.data.dom.createDocument(_ca);}}return _c8;},reset:function(_cb){this.stateDOM=_cb;this.stateNode=this._getStateNode(_cb);},getSerializationManager:function(){return this.serializationManager;},newPortletAccessor:function(_cc,_cd){var _ce;var _cf;if(_cd==null||this.stateDOM==_cd){_ce=this.stateNode;_cf=this.stateDOM;}else{_ce=this._getStateNode(_cd);_cf=_cd;}var _d0="state:portlet[@id='"+_cc+"']";var _d1=this._getSpecificStateNode("portlet",_d0,_ce,_cf);_d1.setAttribute("id",_cc);return new com.ibm.portal.state.PortletAccessor(_d1,_cf);},newPortletListAccessor:function(_d2){var _d3;var _d4;if(_d2==null||this.stateDOM==_d2){_d3=this.stateNode;_d4=this.stateDOM;}else{_d3=this._getStateNode(_d2);_d4=_d2;}return new com.ibm.portal.state.PortletListAccessor(_d3,_d4);},newSelectionAccessor:function(_d5){var _d6;var _d7;if(_d5==null||this.stateDOM==_d5){_d6=this.stateNode;_d7=this.stateDOM;}else{_d6=this._getStateNode(_d5);_d7=_d5;}var _d8=this._getSpecificStateNode("selection","state:selection",_d6,_d7);return new com.ibm.portal.state.SelectionAccessor(_d8,_d7);},newSoloStateAccessor:function(_d9){var _da;var _db;if(_d9==null||this.stateDOM==_d9){_da=this.stateNode;_db=this.stateDOM;}else{_da=this._getStateNode(_d9);_db=_d9;}var _dc=this._getSpecificStateNode("solo","state:solo",_da,_db);return new com.ibm.portal.state.SoloStateAccessor(_dc,_db);},newThemeTemplateAccessor:function(_dd){var _de;var _df;if(_dd==null||this.stateDOM==_dd){_de=this.stateNode;_df=this.stateDOM;}else{_de=this._getStateNode(_dd);_df=_dd;}var _e0=this._getSpecificStateNode("theme-template","state:theme-template",_de,_df);return new com.ibm.portal.state.ThemeTemplateAccessor(_e0,_df);},newLocaleAccessor:function(_e1){var _e2;var _e3;if(_e1==null||this.stateDOM==_e1){_e2=this.stateNode;_e3=this.stateDOM;}else{_e2=this._getStateNode(_e1);_e3=_e1;}var _e4=this._getSpecificStateNode("locale","state:locale",_e2,_e3);return new com.ibm.portal.state.LocaleAccessor(_e4,_e3);},newStatePartitionAccessor:function(_e5){var _e6;var _e7;if(_e5==null||this.stateDOM==_e5){_e6=this.stateNode;_e7=this.stateDOM;}else{_e6=this._getStateNode(_e5);_e7=_e5;}var _e8=this._getSpecificStateNode("statepartition","state:statepartition",_e6,_e7);return new com.ibm.portal.state.StatePartitionAccessor(_e8,_e7);},_getStateNode:function(_e9){var _ea="state:root/state:state[@type='navigational']";var _eb=com.ibm.portal.xpath.evaluateXPath(_ea,_e9,this.ns);var _ec=null;if(_eb==null||_eb.length<=0){var _ed=_e9.firstChild;while(_ed&&_ed.nodeType==7){_ed=_ed.nextSibling;}if(_ed==null){_ed=this._createElement(_e9,"root");this._prependChild(_ed,_e9);}_ec=_ed.firstChild;if(_ec==null){_ec=this._createElement(_e9,"state");this._prependChild(_ec,_ed);}_ec.setAttribute("type","navigational");}else{_ec=_eb[0];}return _ec;},_getSpecificStateNode:function(_ee,_ef,_f0,_f1){var _f2=com.ibm.portal.xpath.evaluateXPath(_ef,_f0,this.ns);var _f3;if(_f2==null||_f2.length<=0){_f3=this._createElement(_f1,_ee);this._prependChild(_f3,_f0);}else{_f3=_f2[0];}return _f3;},_prependChild:function(_f4,_f5){_f5.firstChild?_f5.insertBefore(_f4,_f5.firstChild):_f5.appendChild(_f4);},_createElement:function(dom,_f7){var _f8;if(dojo.isIE){_f8=dom.createNode(1,_f7,this.ns.state);}else{_f8=dom.createElementNS(this.ns.state,_f7);}return _f8;}});dojo.declare("com.ibm.portal.state.PortletAccessor",null,{constructor:function(_f9,_fa){this.portletNode=_f9;this.stateDOM=_fa;this.parameters=new com.ibm.portal.state.Parameters(_f9,_fa);this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};this.xsltURL=dojo.moduleUrl("com","ibm/portal/state/");},getPortletMode:function(){var _fb="state:portlet-mode";var _fc=com.ibm.portal.xpath.evaluateXPath(_fb,this.portletNode,this.ns);var _fd=ibm.portal.portlet.PortletMode.VIEW;if(_fc!=null&&_fc.length>0){var _fe=_fc[0].firstChild;if(_fe!=null){_fd=_fe.nodeValue;}}return _fd;},getWindowState:function(){var _ff="state:window-state";var _100=com.ibm.portal.xpath.evaluateXPath(_ff,this.portletNode,this.ns);var _101=ibm.portal.portlet.WindowState.NORMAL;if(_100!=null&&_100.length>0){var _102=_100[0].firstChild;if(_102!=null){_101=_102.nodeValue;}}return _101;},getRenderParameters:function(){return this.parameters;},setPortletMode:function(_103){var expr="state:portlet-mode";var _105=com.ibm.portal.xpath.evaluateXPath(expr,this.portletNode,this.ns);if(_105==null||_105.length<=0){var _106=this._createElement(this.stateDOM,"portlet-mode");this._prependChild(_106,this.portletNode);var _107=this.stateDOM.createTextNode(_103);this._prependChild(_107,_106);}else{_105[0].firstChild.nodeValue=_103;}},setWindowState:function(_108){var expr="state:window-state";var _10a=com.ibm.portal.xpath.evaluateXPath(expr,this.portletNode,this.ns);if(_10a==null||_10a.length<=0){var _10b=this._createElement(this.stateDOM,"window-state");this._prependChild(_10b,this.portletNode);var _10c=this.stateDOM.createTextNode(_108);this._prependChild(_10c,_10b);}else{_10a[0].firstChild.nodeValue=_108;}},getPortletState:function(){var _10d=dojox.data.dom.createDocument();var _10e=com.ibm.portal.state.STATE_MANAGER.newPortletAccessor(this.portletNode.getAttribute("id"),_10d);_10e.setPortletMode(this.getPortletMode());_10e.setWindowState(this.getWindowState());var _10f=this.getRenderParameters().getMap();if(_10f.length>0){_10e.getRenderParameters().putAll(_10f);}return _10d;},setPortletState:function(_110,_111){var _112=com.ibm.portal.state.STATE_MANAGER.newPortletAccessor(this.portletNode.getAttribute("id"),_110);this.setPortletMode(_112.getPortletMode());this.setWindowState(_112.getWindowState());var _113=_112.getRenderParameters().getMap();if(_111==null||_111==false){this.getRenderParameters().clear();}if(_113.length>0){this.getRenderParameters().putAll(_113);}},_prependChild:function(node,_115){_115.firstChild?_115.insertBefore(node,_115.firstChild):_115.appendChild(node);},_createElement:function(dom,name){var _118;if(dojo.isIE){_118=dom.createNode(1,name,this.ns.state);}else{_118=dom.createElementNS(this.ns.state,name);}return _118;}});dojo.declare("com.ibm.portal.state.Parameters",null,{constructor:function(_119,_11a){this.baseNode=_119;this.stateDOM=_11a;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},getMap:function(){var _11b=this.getNames();var map=new Array(_11b.length);for(var i=0;i<_11b.length;i++){var name=_11b[i];map[i]={name:name,values:this.getValues(name)};}return map;},getNames:function(){var expr="state:parameters/state:param";var _120=com.ibm.portal.xpath.evaluateXPath(expr,this.baseNode,this.ns);var _121=new Array();if(_120!=null&&_120.length>0){var _122=_120.length;for(var i=0;i<_122;i++){_121[i]=_120[i].getAttribute("name");}}return _121;},getValue:function(name){var _125=this.getValues(name);var _126=null;if(_125!=null&&_125.length>0){_126=_125[0];}return _126;},getValues:function(name){var expr="state:parameters/state:param[@name='"+name+"']/state:value";var _129=com.ibm.portal.xpath.evaluateXPath(expr,this.baseNode,this.ns);var _12a=null;if(_129!=null&&_129.length>0){_12a=new Array(_129.length);var _12b=_129.length;for(var i=0;i<_12b;i++){var _12d=_129[i].firstChild;if(_12d!=null){_12a[i]=_12d.nodeValue;}}}return _12a;},remove:function(name){var expr="state:parameters/state:param[@name='"+name+"']";var _130=com.ibm.portal.xpath.evaluateXPath(expr,this.baseNode,this.ns);if(_130!=null){var _131=_130[0];if(_131&&_131.parentNode){_131.parentNode.removeChild(_131);}}},putAll:function(map){if(map!=null&&map.length>0){for(var i=map.length-1;i>=0;i--){var _134=map[i].name;var _135=map[i].values;this.setValues(_134,_135);}}},setValue:function(name,_137){this.setValues(name,new Array(_137));},setValues:function(name,_139){var expr="state:parameters/state:param[@name='"+name+"']";var _13b=com.ibm.portal.xpath.evaluateXPath(expr,this.baseNode,this.ns);var _13c;if(_13b==null||_13b.length==0){var _13d=null;if(_13d==null){_13d=this._createElement(this.stateDOM,"parameters");this._prependChild(_13d,this.baseNode);}_13c=this._createElement(this.stateDOM,"param");_13c.setAttribute("name",name);this._prependChild(_13c,_13d);}else{_13c=_13b[0];dojox.data.dom.removeChildren(_13c);}if(_139!=null){for(var i=_139.length-1;i>=0;i--){var _13f=this._createElement(this.stateDOM,"value");this._prependChild(_13f,_13c);var _140=_139[i];if(_140!=null){var _141=this.stateDOM.createTextNode(_140);this._prependChild(_141,_13f);}}}},clear:function(){var expr="state:parameters";var _143=com.ibm.portal.xpath.evaluateXPath(expr,this.baseNode,this.ns);if(_143!=null){var _144=_143[0];if(_144&&_144.parentNode){_144.parentNode.removeChild(_144);}}},_getFirstChildWithTag:function(_145,_146){if(!_145||!_146){return null;}var node=_145.firstChild;while(node){if(node.nodeType==1&&node.tagName&&node.tagName.toLowerCase()==_146.toLowerCase()){return node;}node=node.nextSibling;}return null;},_prependChild:function(node,_149){_149.firstChild?_149.insertBefore(node,_149.firstChild):_149.appendChild(node);},_createElement:function(dom,name){var _14c;if(dojo.isIE){_14c=dom.createNode(1,name,this.ns.state);}else{_14c=dom.createElementNS(this.ns.state,name);}return _14c;}});dojo.declare("com.ibm.portal.state.PortletListAccessor",null,{constructor:function(_14d,_14e){this.stateNode=_14d;this.stateDOM=_14e;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},getPortlets:function(){var expr="state:portlet";var _150=com.ibm.portal.xpath.evaluateXPath(expr,this.stateNode,this.ns);var _151=null;if(_150!=null&&_150.length>0){_151=new Array(_150.length);for(var i=0;i<_150.length;i++){var node=_150[i];_151[i]=node.getAttribute("id");}}return _151;}});dojo.declare("com.ibm.portal.state.SelectionAccessor",null,{constructor:function(_154,_155){this.selectionNode=_154;this.stateDOM=_155;this.parameters=new com.ibm.portal.state.Parameters(this.selectionNode,_155);this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},getPageSelection:function(){return this.selectionNode.getAttribute("selection-node");},getFragmentSelection:function(){var _156=this.getParameters();var _157=_156.getValues("frg");var _158=null;if(_157!=null&&_157.length>0){_158=_157[0];if(_157.length>1){if(_158=="pw"){_158=_157[1];}}}return _158;},getMapping:function(_159){var expr="state:mapping[@src='"+_159+"']";var _15b=com.ibm.portal.xpath.evaluateXPath(expr,this.selectionNode,this.ns);var _15c=null;if(_15b!=null&&_15b.length>0){var _15d=_15b[0];_15c=_15d.getAttribute("dst");}return _15c;},getParameters:function(){return this.parameters;},setPageSelection:function(_15e){this.selectionNode.setAttribute("selection-node",_15e);},setFragmentSelection:function(_15f,_160){var _161=this.getParameters();if(_160==null||_160==true){var _162=new Array(2);_162[0]=_15f;_162[1]="pw";_161.setValues("frg",_162);}else{_161.setValue("frg",_15f);}},setMapping:function(_163,_164){if(_164!=null){var expr="state:mapping[@src='"+_163+"']";var _166=com.ibm.portal.xpath.evaluateXPath(expr,this.selectionNode,this.ns);var _167;if(_166!=null&&_166.length>0){_167=_166[0];}else{_167=this._createElement(this.stateDOM,"mapping");this._prependChild(_167,this.selectionNode);_167.setAttribute("src",_163);}_167.setAttribute("dst",_164);}else{this.removeMapping(_163);}},removeMapping:function(_168){var expr="state:mapping[@src='"+_168+"']";var _16a=com.ibm.portal.xpath.evaluateXPath(expr,this.selectionNode,this.ns);var _16b=false;if(_16a!=null&&_16a.length>0){for(var i=0;i<_16a.length;i++){var _16d=_16a[i];if(_16d&&_16d.parentNode){_16d.parentNode.removeChild(_16d);}}_16b=true;}return _16b;},_prependChild:function(node,_16f){_16f.firstChild?_16f.insertBefore(node,_16f.firstChild):_16f.appendChild(node);},_createElement:function(dom,name){var _172;if(dojo.isIE){_172=dom.createNode(1,name,this.ns.state);}else{_172=dom.createElementNS(this.ns.state,name);}return _172;},getSelection:function(){return this.getPageSelection();},setSelection:function(_173){this.setPageSelection(_173);}});dojo.declare("com.ibm.portal.state.SoloStateAccessor",null,{constructor:function(_174,_175){this.soloNode=_174;this.stateDOM=_175;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},setSoloPortlet:function(_176){dojox.data.dom.removeChildren(this.soloNode);if(_176!=null){var _177=this.stateDOM.createTextNode(_176);this._prependChild(_177,this.soloNode);}},getSoloPortlet:function(){var _178=this.soloNode.firstChild;if(_178!=null){return _178.nodeValue;}else{return null;}},setReturnSelection:function(_179){this.soloNode.setAttribute("return-selection",_179);},getReturnSelection:function(){return this.soloNode.getAttribute("return-selection");},_prependChild:function(node,_17b){_17b.firstChild?_17b.insertBefore(node,_17b.firstChild):_17b.appendChild(node);}});dojo.declare("com.ibm.portal.state.ThemeTemplateAccessor",null,{constructor:function(_17c,_17d){this.themeTemplateNode=_17c;this.stateDOM=_17d;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},setThemeTemplate:function(_17e){dojox.data.dom.removeChildren(this.themeTemplateNode);if(_17e!=null){var _17f=this.stateDOM.createTextNode(_17e);this._prependChild(_17f,this.themeTemplateNode);}},getThemeTemplate:function(){var _180=this.themeTemplateNode.firstChild;if(_180!=null){return _180.nodeValue;}else{return null;}},_prependChild:function(node,_182){_182.firstChild?_182.insertBefore(node,_182.firstChild):_182.appendChild(node);}});dojo.declare("com.ibm.portal.state.LocaleAccessor",null,{constructor:function(_183,_184){this.localeNode=_183;this.stateDOM=_184;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},setLocale:function(_185){dojox.data.dom.removeChildren(this.localeNode);if(_185!=null){var _186=this.stateDOM.createTextNode(_185);this._prependChild(_186,this.localeNode);}},getLocale:function(){var _187=this.localeNode.firstChild;if(_187!=null){return _187.nodeValue;}else{return null;}},_prependChild:function(node,_189){_189.firstChild?_189.insertBefore(node,_189.firstChild):_189.appendChild(node);}});dojo.declare("com.ibm.portal.state.StatePartitionAccessor",null,{constructor:function(_18a,_18b){this.statePartitionNode=_18a;this.stateDOM=_18b;this.ns={"state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};},includeStatePartition:function(){dojox.data.dom.removeChildren(this.statePartitionNode);var _18c=this.stateDOM.createTextNode(this._generateID());this._prependChild(_18c,this.statePartitionNode);},_prependChild:function(node,_18e){_18e.firstChild?_18e.insertBefore(node,_18e.firstChild):_18e.appendChild(node);},_generateID:function(){return Math.floor(Math.random()*100);}});dojo.declare("com.ibm.portal.state.SerializationManager",null,{STATE_URI_SCHEME:"state",STATE_URI_POST:"state:encode",DOWNLOAD_MODE:"download",STATUS_UNDEFINED:0,STATUS_OK:1,STATUS_ERROR:2,STATE_NS_URI:"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state",STATE_THRESHOLD:1024,constructor:function(_18f){this.serviceURL=_18f;},serialize:function(_190,_191,_192,_193){ibm.portal.debug.entry("SerializationManager.serialize",[dojox.data.dom.innerXML(_190),_191,_192,_193]);var _194=dojox.data.dom.innerXML(_190).replace(/[\r\n]/mg,"");var _195=escape(_194);var _196=this._getMimeType();var _197=null;var me=this;ibm.portal.debug.text("Mime type for response: "+_196);ibm.portal.debug.text("Length of encoded state XML is: "+_195.length);ibm.portal.debug.text("Encoded state XML is: "+_195);var _199=com.ibm.portal.services.PortalRestServiceConfig.digest;ibm.portal.debug.text("Digest: "+_199);if(_195.length<=this.STATE_THRESHOLD){var _19a=this.STATE_URI_SCHEME+":"+_195;var _19b;_191=(_191!=null&&_191==true);if(_191==true){if(_199!=null){_19b={"uri":_19a,"mode":this.DOWNLOAD_MODE,"xmlns":this.STATE_NS_URI,"sessionDependencyAllowed":"true","preprocessors":"true","digest":_199};}else{_19b={"uri":_19a,"mode":this.DOWNLOAD_MODE,"xmlns":this.STATE_NS_URI,"sessionDependencyAllowed":"true","preprocessors":"true"};}}else{if(_199!=null){_19b={"uri":_19a,"mode":this.DOWNLOAD_MODE,"xmlns":this.STATE_NS_URI,"sessionDependencyAllowed":"true","digest":_199};}else{_19b={"uri":_19a,"mode":this.DOWNLOAD_MODE,"xmlns":this.STATE_NS_URI,"sessionDependencyAllowed":"true"};}}if(_193===true){_19b.forceAbsolute=true;}ibm.portal.debug.text("Doing a GET request: { url: \""+this.serviceURL+"\", sync: "+((_192)?false:true)+", content: "+_19b+", handleAs: "+_196+", transport: XMLHTTPRequest");ibm.portal.debug.text("Parameters: uri=\""+_19b.uri+"\" mode=\""+_19b.mode+"\" xmlns=\""+_19b.xmlns+"\""+"forceAbsolute=\""+_19b.forceAbsolute+"\"");dojo.xhrGet({url:this.serviceURL,sync:(_192)?false:true,content:_19b,handleAs:_196,handle:function(_19c,_19d){ibm.portal.debug.text("Response: "+_19c);_197=me._handleSerializationResponse.call(me,_19c,_192,_190,_191);return _19c;},transport:"XMLHTTPTransport"});}else{ibm.portal.debug.text("Doing a POST request.");if(dojo.isIE){var idx=_194.indexOf("UTF-16");if(idx>=0){_194=_194.replace(/UTF-16/,"UTF-8");}}var url=this.serviceURL+"?uri="+this.STATE_URI_POST+"&xmlns="+this.STATE_NS_URI+"&sessionDependencyAllowed=true";if(_199!=null){url+="&digest="+_199;}if(_191===true){url+="&preprocessors=true";}if(_193===true){url+="&forceAbsolute=true";}dojo.rawXhrPost({url:url,sync:(_192)?false:true,postData:_194,handleAs:_196,headers:{"Content-Type":"text/xml"},handle:function(_1a0,_1a1){_197=me._handleSerializationResponse.call(me,_1a0,_192,_190,_191);return _1a0;},transport:"XMLHTTPTransport"});}ibm.portal.debug.exit("SerializationManager.serialize",_197);return _197;},deserialize:function(url,_1a3){var _1a4=this.STATE_URI_SCHEME+":"+url;var _1a5=null;var _1a6=this._getMimeType();var me=this;var _1a8=com.ibm.portal.services.PortalRestServiceConfig.digest;ibm.portal.debug.text("Digest: "+_1a8);var _1a9;if(_1a8!=null){_1a9={"uri":_1a4,"mode":this.DOWNLOAD_MODE,"xmlns":this.STATE_NS_URI,"digest":_1a8,"preprocessors":"true"};}else{_1a9={"uri":_1a4,"mode":this.DOWNLOAD_MODE,"xmlns":this.STATE_NS_URI,"preprocessors":"true"};}dojo.xhrGet({url:this.serviceURL,sync:(_1a3)?false:true,content:_1a9,handleAs:_1a6,handle:function(_1aa,_1ab){var type=(_1aa instanceof Error)?"error":"load";if(type=="load"){var _1ad=me._getResponseXML(_1aa);if(_1ad.documentElement.nodeName=="parsererror"){_1ad=dojox.data.dom.createDocument();}if(_1a3){_1a3(1,url,_1ad);}else{_1a5={"status":1,"input":me.serviceURL,"url":me.serviceURL,"returnObject":_1ad,"state":_1ad};}}else{if(type=="error"){if(_1a3){_1a3(2,url,null);}else{_1a5={"status":2,"input":me.serviceURL,"url":me.serviceURL,"returnObject":null,"state":null};}}}},transport:"XMLHTTPTransport"});return _1a5;},_handleSerializationResponse:function(_1ae,_1af,_1b0,_1b1){var _1b2=null;var type=(_1ae instanceof Error)?"error":"load";if(type=="load"){var _1b4=this._getResponseXML(_1ae);var _1b5="atom:entry/atom:link";var ns={"atom":"http://www.w3.org/2005/Atom","state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state"};var _1b7=null;var _1b8=com.ibm.portal.xpath.evaluateXPath(_1b5,_1b4,ns);if(_1b8!=null&&_1b8.length>0){_1b7=_1b8[0].getAttribute("href");}var _1b9=_1b0;if(_1b1==true){var _1ba="atom:entry/atom:content/state:root";var _1bb=com.ibm.portal.xpath.evaluateXPath(_1ba,_1b4,ns);if(_1bb!=null&&_1bb.length>0){var _1bc=dojox.data.dom.innerXML(_1bb[0]);_1b9=dojox.data.dom.createDocument(_1bc);}}if(_1af){_1af(1,_1b9,_1b7);}else{_1b2={"status":1,"input":_1b9,"state":_1b9,"returnObject":_1b7,"url":_1b7};}}else{if(type=="error"){if(_1af){_1af(this.STATUS_ERROR,_1b0,null);}else{_1b2={"status":this.STATUS_ERROR,"input":_1b0,"state":_1b0,"returnObject":null,"url":null};}}}return _1b2;},_getMimeType:function(){var _1bd="xml";if(dojo.isIE){_1bd="text";}return _1bd;},_getResponseXML:function(data){var _1bf=data;if(dojo.isIE){_1bf=dojox.data.dom.createDocument(data);}return _1bf;},_encodeAscii:function(str){var ret=str;if(dojo.isString(ret)){var _1c2=escape(ret);var _1c3=/%u([A-F0-9][A-F0-9][A-F0-9][A-F0-9])/i;var _1c4=null;while((_1c4=_1c2.match(_1c3))){ret+=_1c2.substring(0,_1c4.index)+escape(Number("0x"+_1c4[1]));_1c2=_1c2.substring(_1c4.index+_1c4[0].length);}ret+=_1c2;ret=ret.replace(/\+/g,"%2B");}return ret;}});dojo.declare("com.ibm.portal.navigation.controller.StateVaryManager",null,{constructor:function(){this._expr=new Array();},setExpressions:function(id,_1c6){var _1c7=this._findBucket(id);if(_1c7==null){_1c7={"id":id,"expr":null};this._expr.push(_1c7);}_1c7.expr=_1c6;},getExpressions:function(id){var _1c9=null;var _1ca=this._findBucket(id);if(_1ca!=null){_1c9=_1ca.expr;}return _1c9;},_findBucket:function(id){var _1cc=null;for(i=0;i<this._expr.length;i++){var temp=this._expr[i];if(temp.id==id){_1cc=temp;break;}}return _1cc;}});com.ibm.portal.state.STATE_MANAGER=new com.ibm.portal.state.StateManager();com.ibm.portal.state.STATE_MANAGER.reset(dojox.data.dom.createDocument());}if(!dojo._hasResource["com.ibm.ajax.auth"]){dojo._hasResource["com.ibm.ajax.auth"]=true;dojo.provide("com.ibm.ajax.auth");com.ibm.ajax.auth={prepareSecure:function(args,_1cf,_1d0){args._handle=args.handle;args.handle=dojo.partial(this.testAuthenticationHandler,this,_1cf,_1d0);return args;},setAuthenticationHandler:function(_1d1){this.authenticationHandler=_1d1;},setTestAuthenticationHandler:function(_1d2){this.testAuthenticationHandler=_1d2;},setDefaultAuthenticationTests:function(_1d3,_1d4,_1d5){this.checkFromCaller=_1d3;this.checkByContentType=_1d4;this.checkByStatusCode=_1d5;},addAuthenticationCheck:function(_1d6){if(_1d6){this.authenticationChecks.push(_1d6);}},isAuthenticationRequired:function(_1d7,_1d8){var _1d9=_1d8.args.handleAs;var _1da=false;if(!_1d7||dojo.indexOf(["cancel","timeout"],_1d7.dojoType)==-1){if(this.checkByContentType&&dojo.indexOf(["xml","json","json-comment-optional","text"],_1d9)!=-1&&_1d8.xhr&&/^text\/html/.exec(_1d8.xhr.getResponseHeader("Content-Type"))&&_1d8.xhr.status>=200&&_1d8.xhr.status<300){ibm.portal.debug.text("auth::isAuthenticationRequired DEBUG content type does not match request, assume logged out");return true;}else{if(this.checkByStatusCode&&dojo.indexOf(["xml","json","json-comment-optional","text"],_1d9)!=-1&&_1d8.xhr&&_1d8.xhr.status==302){ibm.portal.debug.text("auth::isAuthenticationRequired DEBUG redirect received, assume login request");return true;}else{if(this.checkByStatusCode&&_1d8.xhr&&(_1d8.xhr.status==401||_1d8.xhr.status==0)&&_1d8.xhr.getResponseHeader("WWW-Authenticate")&&_1d8.xhr.getResponseHeader("WWW-Authenticate").indexOf("IBMXHR")!=-1){ibm.portal.debug.text("auth::isAuthenticationRequired DEBUG Portal 401 received, assume login required");return true;}}}}if(!_1da){for(var i=0;i<this.authenticationChecks.length;i++){if(this.authenticationChecks[i](this,_1d7,_1d8)){return true;}}}return false;},testAuthenticationHandler:function(auth,_1dd,_1de,_1df,_1e0){var args=dojo._toArray(arguments).slice(3);var _1e2=false;if(!_1df||dojo.indexOf(["cancel","timeout"],_1df.dojoType)==-1){if(auth.checkFromCaller&&typeof _1dd=="function"&&_1dd(_1df,_1e0)){_1e2=true;}else{_1e2=auth.isAuthenticationRequired(_1df,_1e0,_1dd);}}if(_1e2){var path=auth._parseUri(_1e0.args.url).path;dojo.cookie("WASPostParam",null,{expires:-1,path:path});dojo.cookie("WASReqURL",null,{expires:-1,path:"/"});auth.authenticationHandler(_1df,_1e0,_1de);args[0]=new Error("xhr unauthenticated");args[0].dojoType="unauthenticated";}if(_1e0.args._handle){return _1e0.args._handle.apply(this,args);}else{return (_1df);}},_parseUri:function(uri){if(!uri){return null;}uri=new dojo._Url(uri);var _1e5=this._splitQuery(uri.query);uri.queryParameters=_1e5;return uri;},_splitQuery:function(_1e6){var _1e7={};if(!_1e6){return _1e7;}if(_1e6.charAt(0)=="?"){_1e6=_1e6.substring(1);}var args=_1e6.split("&");for(var i=0;i<args.length;i++){if(args[i].length>0){var _1ea=args[i].indexOf("=");if(_1ea==-1){var key=decodeURIComponent(args[i]);var _1ec=_1e7[key];if(dojo.isArray(_1ec)){_1ec.push("");}else{if(_1ec){_1e7[key]=[_1ec,""];}else{_1e7[key]="";}}}else{if(_1ea>0){var key=decodeURIComponent(args[i].substring(0,_1ea));var _1ed=decodeURIComponent(args[i].substring(_1ea+1));var _1ec=_1e7[key];if(dojo.isArray(_1ec)){_1ec.push(_1ed);}else{if(_1ec){_1e7[key]=[_1ec,_1ed];}else{_1e7[key]=_1ed;}}}}}}return _1e7;},checkFromCaller:true,checkByContentType:true,checkByStatusCode:true,authenticationChecks:[],authenticationHandler:function(){ibm.portal.debug.text("auth::authenticationHandler DEBUG authentication was required");}};}if(!dojo._hasResource["com.ibm.portal.debug"]){dojo._hasResource["com.ibm.portal.debug"]=true;dojo.provide("com.ibm.portal.debug");dojo.provide("ibm.portal.debug");ibm.portal.debug.setTrace=function(_1ee){ibm.portal.debug._traceString=_1ee;};ibm.portal.debug._isDebugEnabled=function(){var _1ef=false;if(typeof (ibmPortalConfig)!="undefined"){if(ibmPortalConfig&&ibmPortalConfig.isDebug){_1ef=true;}}return _1ef;};ibm.portal.debug.text=function(str,_1f1){if(typeof (ibmPortalConfig)!="undefined"){if(ibmPortalConfig&&ibmPortalConfig.isDebug){var _1f2=ibm.portal.debug._traceString;if(_1f2){if(_1f1){if(_1f1.indexOf(_1f2)>=0){window.console.log(str);}}}else{window.console.log(str);}}}};ibm.portal.debug.entry=function(_1f3,args){if(ibm.portal.debug._isDebugEnabled()){var _1f5=_1f3+" --> entry; { ";if(args&&args.length>0){for(arg in args){_1f5=_1f5+args[arg]+" ";}}_1f5=_1f5+" } ";ibm.portal.debug.text(_1f5,_1f3);}};ibm.portal.debug.exit=function(_1f6,_1f7){if(ibm.portal.debug._isDebugEnabled()){var _1f8=_1f6+" --> exit;";if(typeof (_1f7)!="undefined"){_1f8=_1f8+" { "+_1f7+" } ";}ibm.portal.debug.text(_1f8,_1f6);}};ibm.portal.debug.escapeXmlForHTMLDisplay=function(_1f9){_1f9=_1f9.replace(/</g,"&lt;");_1f9=_1f9.replace(/>/g,"&gt;");return _1f9;};}if(!dojo._hasResource["com.ibm.portal.EventBroker"]){dojo._hasResource["com.ibm.portal.EventBroker"]=true;dojo.provide("com.ibm.portal.EventBroker");dojo.require("com.ibm.portal.debug");dojo.declare("com.ibm.portal.Event",null,{constructor:function(_1fa){this.eventName=_1fa;this._listeners=new Array();},fire:function(_1fb){ibm.portal.debug.text("Firing event: "+this.eventName+" with parameters: ");dojo.publish(this.eventName,[_1fb]);},register:function(_1fc,_1fd){if(!_1fd){return dojo.subscribe(this.eventName,null,_1fc);}else{return dojo.subscribe(this.eventName,_1fc,_1fd);}},unregister:function(_1fe){dojo.unsubscribe(_1fe);},cancel:function(_1ff){dojo.publish(this.id+"/cancel");}});dojo.declare("com.ibm.portal.EventBroker",null,{startPage:new com.ibm.portal.Event("portal/StartPage"),endPage:new com.ibm.portal.Event("portal/EndPage"),startFragment:new com.ibm.portal.Event("portal/StartFragment"),endFragment:new com.ibm.portal.Event("portal/EndFragment"),fragmentUpdated:new com.ibm.portal.Event("portal/FragmentUpdated"),startRequest:new com.ibm.portal.Event("portal/StartRequest"),endRequest:new com.ibm.portal.Event("portal/EndRequest"),cancelAll:new com.ibm.portal.Event("portal/CancelAll"),cancelFragmentUpdate:new com.ibm.portal.Event("portal/CancelFragmentUpdate"),stateChanged:new com.ibm.portal.Event("portal/StateChanged"),startScriptHandling:new com.ibm.portal.Event("portal/StartScriptHandling"),endScriptHandling:new com.ibm.portal.Event("portal/EndScriptHandling"),startScriptExecution:new com.ibm.portal.Event("portal/StartScriptExecution"),endScriptExecution:new com.ibm.portal.Event("portal/EndScriptExecution"),javascriptCleanup:new com.ibm.portal.Event("portal/JavascriptCleanup"),beforeSnapShot:new com.ibm.portal.Event("portal/BeforeSnapShot"),afterSnapShot:new com.ibm.portal.Event("portal/AfterSnapShot"),restorePointUpdated:new com.ibm.portal.Event("portal/RestorePointUpdated"),clearRestorePoint:new com.ibm.portal.Event("portal/ClearRestorePoint"),stopEvent:new com.ibm.portal.Event("portal/StopEvent"),redirect:new com.ibm.portal.Event("portal/Redirect")});com.ibm.portal.EVENT_BROKER=new com.ibm.portal.EventBroker();}if(!dojo._hasResource["dojo.dnd.common"]){dojo._hasResource["dojo.dnd.common"]=true;dojo.provide("dojo.dnd.common");dojo.dnd._isMac=navigator.appVersion.indexOf("Macintosh")>=0;dojo.dnd._copyKey=dojo.dnd._isMac?"metaKey":"ctrlKey";dojo.dnd.getCopyKeyState=function(e){return e[dojo.dnd._copyKey];};dojo.dnd._uniqueId=0;dojo.dnd.getUniqueId=function(){var id;do{id=dojo._scopeName+"Unique"+(++dojo.dnd._uniqueId);}while(dojo.byId(id));return id;};dojo.dnd._empty={};dojo.dnd.isFormElement=function(e){var t=e.target;if(t.nodeType==3){t=t.parentNode;}return " button textarea input select option ".indexOf(" "+t.tagName.toLowerCase()+" ")>=0;};dojo.dnd._lmb=dojo.isIE?1:0;dojo.dnd._isLmbPressed=dojo.isIE?function(e){return e.button&1;}:function(e){return e.button===0;};}if(!dojo._hasResource["com.ibm.portal.services.PortalRestServiceRequestQueue"]){dojo._hasResource["com.ibm.portal.services.PortalRestServiceRequestQueue"]=true;dojo.provide("com.ibm.portal.services.PortalRestServiceRequestQueue");dojo.declare("com.ibm.portal.services.PortalRestServiceRequestQueue",null,{maxNumberOfActiveRequests:4,constructor:function(){var _206="PortalRestServiceRequestQueue.constructor";ibm.portal.debug.entry(_206);this._activeRequests=0;this._requestQueue=[];ibm.portal.debug.exit(_206);},add:function(req){var _208="PortalRestServiceRequestQueue.add";ibm.portal.debug.entry(_208,[req]);this._requestQueue.push(req);var me=this;setTimeout(function(){me._executeNextRequest();},5);ibm.portal.debug.exit(_208);},_executeNextRequest:function(){var _20a="PortalRestServiceRequestQueue._executeNextRequest";ibm.portal.debug.entry(_20a);ibm.portal.debug.text(this._requestQueue.length+" request(s) in the queue. "+this._activeRequests+" active request(s) currently.",_20a);if(this._requestQueue.length>0&&this._activeRequests<this.maxNumberOfActiveRequests){var _20b=this._requestQueue.shift();ibm.portal.debug.text("Executing request: "+_20b,_20a);var me=this;setTimeout(function(){me._activeRequests=me._activeRequests+1;_20b.execute(function(){me._notifyComplete();});},1);}else{ibm.portal.debug.text("No request(s) pending or maximum number of requests already currently active.",_20a);}ibm.portal.debug.exit(_20a);},_notifyComplete:function(){var _20d="PortalRestServiceRequestQueue._notifyComplete";this._activeRequests=this._activeRequests-1;if(this._activeRequests<0){this._activeRequests=0;}var me=this;setTimeout(function(){me._executeNextRequest();},5);}});}if(!dojo._hasResource["com.ibm.portal.utilities"]){dojo._hasResource["com.ibm.portal.utilities"]=true;dojo.provide("com.ibm.portal.utilities");com.ibm.portal.utilities={findPortletIdByElement:function(_20f){ibm.portal.debug.entry("findPortletID",[_20f]);var id="";var _211=_20f.parentNode;while(_211&&id.length==0){ibm.portal.debug.text("examining element "+_211.tagName+"; class="+_211.className,"findPortletID");if(_211.className&&(_211.className.match(/\bwpsPortletBody\b/)||_211.className.match(/\bwpsPortletBodyInlineMode\b/))){id=_211.id;var _212=id.indexOf("_mode");if(_212>=0){id=id.substring(0,_212);}}_211=_211.parentNode;}if(id.indexOf("portletActions_")>=0){id=id.substring("portletActions_".length);}ibm.portal.debug.exit("findPortletID",[id]);return id;},findFormByElement:function(_213){var _214=_213;while(_214){if(_214.tagName&&_214.tagName.toLowerCase()=="form"){break;}_214=_214.parentNode;}return _214;},encodeURI:function(uri){ibm.portal.debug.entry("encodeURI",[uri]);var _216=uri;var _217=uri.lastIndexOf(":");while(_217>=0){var _218=_216.substring(0,_217);var part=_216.substring(_217+1);_216=_218+":"+encodeURIComponent(part);_217=_218.lastIndexOf(":");}_216=encodeURIComponent(_216);ibm.portal.debug.exit("encodeURI",[_216]);return _216;},decodeURI:function(uri){ibm.portal.debug.entry("decodeURI",[uri]);var _21b=decodeURIComponent(uri);var _21c=_21b.indexOf(":");while(_21c>=0){var _21d=_21b.substring(0,_21c);var part=_21b.substring(_21c+1);_21b=_21d+":"+decodeURIComponent(part);_21c=_21b.indexOf(":",_21c+1);}ibm.portal.debug.exit("decodeURI",[_21b]);return _21b;},getSelectionNodeId:function(_21f){ibm.portal.debug.entry("getSelectionNodeId",[_21f]);var _220=_21f.split("@oid:");ibm.portal.debug.exit("getSelectionNodeId",[_220[1]]);return _220[1];},getControlId:function(_221){ibm.portal.debug.entry("_getControlId",[_221]);var _222=_221.split("@oid:");var _223=_222[0].split("oid:");ibm.portal.debug.exit("getControlId",[_223[1]]);return _223[1];},overwriteProperty:function(obj,_225,_226,_227){ibm.portal.debug.entry("overwriteProperty",[obj,_225,_226,_227]);if(!obj["_overwritten_"]){obj["_overwritten_"]=new Object();}if(!_227){_227=false;}var _228=(_227&&(obj["_overwritten_"][_225]!=null));if(!_228){if(obj["_overwritten_"][_225]==null){obj["_overwritten_"][_225]=obj[_225];}else{obj["_overwritten_"][_225]=null;}obj[_225]=_226;ibm.portal.debug.text("Property overwrite successful!");}ibm.portal.debug.exit("overwriteProperty");},restoreProperty:function(obj,_22a){ibm.portal.debug.entry("utilities.restoreProperty",[obj,_22a]);var _22b=obj[_22a];if(obj["_overwritten_"]!=null){ibm.portal.debug.text("overwritten property value: "+obj["_overwritten_"]);obj[_22a]=obj["_overwritten_"][_22a];obj["_overwritten_"][_22a]=null;}else{obj[_22a]=null;}ibm.portal.debug.exit("utilities.restoreProperty",_22b);return _22b;},getOverwrittenProperty:function(obj,_22d){if(obj["_overwritten_"]){return obj["_overwritten_"][_22d];}else{return null;}},setOverwrittenProperty:function(obj,_22f,_230){ibm.portal.debug.entry("utilities.setOverwrittenProperty",[obj,_22f,_230]);if(!obj["_overwritten_"]){obj["_overwritten_"]=new Object();}obj["_overwritten_"][_22f]=_230;ibm.portal.debug.exit("utilities.setOverwrittenProperty");},callOverwrittenFunction:function(_231,_232,args){ibm.portal.debug.entry("utilities.callOverwrittenFunction",[_231,_232,args]);var _234=null;var _235=this.getOverwrittenProperty(_231,_232);ibm.portal.debug.text("Overwritten property: "+_235);ibm.portal.debug.text("old property's apply function: "+_235.apply);if(args){_234=_235.apply(_231,args);}else{_234=_235.apply(_231);}ibm.portal.debug.exit("utilities.callOverwrittenFunction",_234);return _234;},clearOverwrittenProperties:function(obj){ibm.portal.debug.entry("utilities.clearOverwrittenProperties",[obj]);if(obj["_overwritten_"]){obj["_overwritten_"]=null;}ibm.portal.debug.exit("utilities.clearOverwrittenProperties");},isExternalUrl:function(_237){ibm.portal.debug.entry("isExternalUrl",[_237]);var host=window.location.host;var _239=window.location.protocol;var _23a=_237.split("?")[0];var _23b=!(_23a.indexOf("://")<0||(_23a.indexOf(_239)==0&&_23a.indexOf(host)==_239.length+2));ibm.portal.debug.text("urlStringNoQuery.indexOf(\"://\") = "+_23a.indexOf("://"));ibm.portal.debug.text("urlStringNoQuery.indexOf(protocol) = "+_23a.indexOf(_239));ibm.portal.debug.exit("isExternalUrl",_23b);return _23b;},isJavascriptUrl:function(_23c){ibm.portal.debug.entry("isJavascriptUrl",[_23c]);var url=com.ibm.portal.utilities.string.trim(_23c.toLowerCase());var _23e=(url.indexOf("javascript:")==0);ibm.portal.debug.exit("isJavascriptUrl",_23e);return _23e;},isPortalUrl:function(_23f){ibm.portal.debug.entry("utilities.isPortalUrl",[_23f]);var _240=(_23f.indexOf(ibmPortalConfig["portalURI"])>=0);ibm.portal.debug.exit("utilities.isPortalUrl",_240);return _240;},addExternalNode:function(doc,node){var _243=null;if(doc.importNode){_243=doc.importNode(node,true);}else{_243=node;}doc.appendChild(_243);},decodeXML:function(_244){ibm.portal.debug.entry("decodeXML",[_244]);var _245=_244.replace(/&amp;/g,"&");var _246=_245.replace(/&amp;/g,"&");_245=_246.replace(/&#039;/g,"'");_246=_245.replace(/&#034;/g,"\"");_246=_246.replace(/&lt;/g,"<");_246=_246.replace(/&gt;/g,">");ibm.portal.debug.exit("decodeXML",[_246]);return _246;},eventHandlerToString:function(_247){var _248=_247.toString();var _249=_248.indexOf("{");var _24a=_248.lastIndexOf("}");onclickStr=_248.substring(_249+1,_24a);return onclickStr;},_waitingForScript:false,_isWaitingForScript:function(){return com.ibm.portal.utilities._waitingForScript;},stopWaitingForScript:function(){com.ibm.portal.utilities._waitingForScript=false;},waitFor:function(_24b,_24c,_24d,args){var _24f=setInterval(function(){if(_24b()){clearInterval(_24f);if(!args){_24d();}else{_24d(args);}}},_24c);},waitForScript:function(_250,args){com.ibm.portal.utilities._waitingForScript=true;com.ibm.portal.utilities.waitFor(function(){return (!com.ibm.portal.utilities._isWaitingForScript());},500,_250,args);}};com.ibm.portal.utilities.string={findNext:function(_252,_253,from){ibm.portal.debug.entry("string.findNext",[_252,_253]);var _255=-1;for(var i=0;i<_253.length;i++){var _257=null;if(from){_257=from+_253[i].length;}var _258=_252.indexOf(_253[i],_257);if(_258>-1&&(_258<_255||_255==-1)){_255=_258;}}ibm.portal.debug.exit("string.findNext",[_255]);return _255;},contains:function(_259,_25a){ibm.portal.debug.entry("string.contains",[_259,_25a]);var _25b=false;if(_259!=null&&_25a!=null){_25b=(_259.indexOf(_25a)!=-1);}ibm.portal.debug.exit("string.contains",[_25b]);return _25b;},strip:function(_25c,_25d){ibm.portal.debug.entry("string.strip",[_25c,_25d]);var _25e=_25c.replace(new RegExp(_25d,"g"),"");ibm.portal.debug.exit("string.strip",[_25e]);return _25e;},properCase:function(_25f){if(_25f==null||_25f.length<1){return "";}ibm.portal.debug.entry("string.properCase",[_25f]);var _260=_25f.charAt(0).toUpperCase();if(_25f.length>1){_260+=_25f.substring(1).toLowerCase();}ibm.portal.debug.exit("string.properCase",[_260]);return _260;},trim:function(_261){ibm.portal.debug.entry("string.trim",[_261]);var _262=_261;_262=_262.replace(/^\s+/,"");_262=_262.replace(/\s+$/,"");ibm.portal.debug.exit("string.trim",_262);return _262;}};dojo.declare("com.ibm.portal.utilities.HttpUrl",null,{constructor:function(_263){this.scheme=window.location.protocol+"//";this.server=this._extractServer(_263);this.port=this._extractPort(_263);this.path=this._extractPath(_263);this.query=this._extractQuery(_263);this.anchor="";},addParameter:function(name,_265){this.query+="&"+name+"="+_265;},toString:function(){var str="";if(this.server!=""){str+=this.scheme+this.server;}if(this.port!=""){str+=":"+this.port;}str+="/"+this.path;if(this.query!=""){str+="?"+this.query;}if(this.anchor!=""){str+="#"+this.anchor;}return str;},_extractServer:function(_267){var _268=_267.indexOf(this.scheme);var _269="";if(_268==0){var _26a=_267.indexOf("/",_268+this.scheme.length);var _26b=_267.substring(_268+this.scheme.length,_26a);_269=_26b.split(":")[0];}return _269;},_extractPort:function(_26c){var _26d=_26c.indexOf(this.server);var _26e="";if(_26d>=0){var _26f=_26c.indexOf("/",_26d);var _270=_26c.substring(_26d,_26f);var _271=_270.split(":");if(_271.length>1){_26e=_271[1];}}return _26e;},_extractPath:function(_272){var _273=_272.indexOf(this.server);var _274="";if(_273>=0){var _275=_272.indexOf("/",_273);var _276=_272.indexOf("?");var _277=_272.lastIndexOf("#");if(_276>=0){_274=_272.substring(_275+1,_276);}else{if(_277>=0){_274=_272.substring(_275+1,_277);}else{_274=_272.substring(_275+1);}}}return _274;},_extractQuery:function(_278){var _279="";var _27a=_278.split("?");if(_27a.length>1){_279=_27a[1].split("#")[0];}return _279;},_extractAnchor:function(_27b){var _27c="";var _27d=_27b.split("#");if(_27d.length>1){_27c=_27d[_27d.length-1];}return _27c;}});}if(!dojo._hasResource["dojo.fx.Toggler"]){dojo._hasResource["dojo.fx.Toggler"]=true;dojo.provide("dojo.fx.Toggler");dojo.declare("dojo.fx.Toggler",null,{constructor:function(args){var _t=this;dojo.mixin(_t,args);_t.node=args.node;_t._showArgs=dojo.mixin({},args);_t._showArgs.node=_t.node;_t._showArgs.duration=_t.showDuration;_t.showAnim=_t.showFunc(_t._showArgs);_t._hideArgs=dojo.mixin({},args);_t._hideArgs.node=_t.node;_t._hideArgs.duration=_t.hideDuration;_t.hideAnim=_t.hideFunc(_t._hideArgs);dojo.connect(_t.showAnim,"beforeBegin",dojo.hitch(_t.hideAnim,"stop",true));dojo.connect(_t.hideAnim,"beforeBegin",dojo.hitch(_t.showAnim,"stop",true));},node:null,showFunc:dojo.fadeIn,hideFunc:dojo.fadeOut,showDuration:200,hideDuration:200,show:function(_280){return this.showAnim.play(_280||0);},hide:function(_281){return this.hideAnim.play(_281||0);}});}if(!dojo._hasResource["dojo.fx"]){dojo._hasResource["dojo.fx"]=true;dojo.provide("dojo.fx");(function(){var d=dojo,_283={_fire:function(evt,args){if(this[evt]){this[evt].apply(this,args||[]);}return this;}};var _286=function(_287){this._index=-1;this._animations=_287||[];this._current=this._onAnimateCtx=this._onEndCtx=null;this.duration=0;d.forEach(this._animations,function(a){this.duration+=a.duration;if(a.delay){this.duration+=a.delay;}},this);};d.extend(_286,{_onAnimate:function(){this._fire("onAnimate",arguments);},_onEnd:function(){d.disconnect(this._onAnimateCtx);d.disconnect(this._onEndCtx);this._onAnimateCtx=this._onEndCtx=null;if(this._index+1==this._animations.length){this._fire("onEnd");}else{this._current=this._animations[++this._index];this._onAnimateCtx=d.connect(this._current,"onAnimate",this,"_onAnimate");this._onEndCtx=d.connect(this._current,"onEnd",this,"_onEnd");this._current.play(0,true);}},play:function(_289,_28a){if(!this._current){this._current=this._animations[this._index=0];}if(!_28a&&this._current.status()=="playing"){return this;}var _28b=d.connect(this._current,"beforeBegin",this,function(){this._fire("beforeBegin");}),_28c=d.connect(this._current,"onBegin",this,function(arg){this._fire("onBegin",arguments);}),_28e=d.connect(this._current,"onPlay",this,function(arg){this._fire("onPlay",arguments);d.disconnect(_28b);d.disconnect(_28c);d.disconnect(_28e);});if(this._onAnimateCtx){d.disconnect(this._onAnimateCtx);}this._onAnimateCtx=d.connect(this._current,"onAnimate",this,"_onAnimate");if(this._onEndCtx){d.disconnect(this._onEndCtx);}this._onEndCtx=d.connect(this._current,"onEnd",this,"_onEnd");this._current.play.apply(this._current,arguments);return this;},pause:function(){if(this._current){var e=d.connect(this._current,"onPause",this,function(arg){this._fire("onPause",arguments);d.disconnect(e);});this._current.pause();}return this;},gotoPercent:function(_292,_293){this.pause();var _294=this.duration*_292;this._current=null;d.some(this._animations,function(a){if(a.duration<=_294){this._current=a;return true;}_294-=a.duration;return false;});if(this._current){this._current.gotoPercent(_294/this._current.duration,_293);}return this;},stop:function(_296){if(this._current){if(_296){for(;this._index+1<this._animations.length;++this._index){this._animations[this._index].stop(true);}this._current=this._animations[this._index];}var e=d.connect(this._current,"onStop",this,function(arg){this._fire("onStop",arguments);d.disconnect(e);});this._current.stop();}return this;},status:function(){return this._current?this._current.status():"stopped";},destroy:function(){if(this._onAnimateCtx){d.disconnect(this._onAnimateCtx);}if(this._onEndCtx){d.disconnect(this._onEndCtx);}}});d.extend(_286,_283);dojo.fx.chain=function(_299){return new _286(_299);};var _29a=function(_29b){this._animations=_29b||[];this._connects=[];this._finished=0;this.duration=0;d.forEach(_29b,function(a){var _29d=a.duration;if(a.delay){_29d+=a.delay;}if(this.duration<_29d){this.duration=_29d;}this._connects.push(d.connect(a,"onEnd",this,"_onEnd"));},this);this._pseudoAnimation=new d._Animation({curve:[0,1],duration:this.duration});var self=this;d.forEach(["beforeBegin","onBegin","onPlay","onAnimate","onPause","onStop"],function(evt){self._connects.push(d.connect(self._pseudoAnimation,evt,function(){self._fire(evt,arguments);}));});};d.extend(_29a,{_doAction:function(_2a0,args){d.forEach(this._animations,function(a){a[_2a0].apply(a,args);});return this;},_onEnd:function(){if(++this._finished==this._animations.length){this._fire("onEnd");}},_call:function(_2a3,args){var t=this._pseudoAnimation;t[_2a3].apply(t,args);},play:function(_2a6,_2a7){this._finished=0;this._doAction("play",arguments);this._call("play",arguments);return this;},pause:function(){this._doAction("pause",arguments);this._call("pause",arguments);return this;},gotoPercent:function(_2a8,_2a9){var ms=this.duration*_2a8;d.forEach(this._animations,function(a){a.gotoPercent(a.duration<ms?1:(ms/a.duration),_2a9);});this._call("gotoPercent",arguments);return this;},stop:function(_2ac){this._doAction("stop",arguments);this._call("stop",arguments);return this;},status:function(){return this._pseudoAnimation.status();},destroy:function(){d.forEach(this._connects,dojo.disconnect);}});d.extend(_29a,_283);dojo.fx.combine=function(_2ad){return new _29a(_2ad);};dojo.fx.wipeIn=function(args){args.node=d.byId(args.node);var node=args.node,s=node.style,o;var anim=d.animateProperty(d.mixin({properties:{height:{start:function(){o=s.overflow;s.overflow="hidden";if(s.visibility=="hidden"||s.display=="none"){s.height="1px";s.display="";s.visibility="";return 1;}else{var _2b3=d.style(node,"height");return Math.max(_2b3,1);}},end:function(){return node.scrollHeight;}}}},args));d.connect(anim,"onEnd",function(){s.height="auto";s.overflow=o;});return anim;};dojo.fx.wipeOut=function(args){var node=args.node=d.byId(args.node),s=node.style,o;var anim=d.animateProperty(d.mixin({properties:{height:{end:1}}},args));d.connect(anim,"beforeBegin",function(){o=s.overflow;s.overflow="hidden";s.display="";});d.connect(anim,"onEnd",function(){s.overflow=o;s.height="auto";s.display="none";});return anim;};dojo.fx.slideTo=function(args){var node=args.node=d.byId(args.node),top=null,left=null;var init=(function(n){return function(){var cs=d.getComputedStyle(n);var pos=cs.position;top=(pos=="absolute"?n.offsetTop:parseInt(cs.top)||0);left=(pos=="absolute"?n.offsetLeft:parseInt(cs.left)||0);if(pos!="absolute"&&pos!="relative"){var ret=d.coords(n,true);top=ret.y;left=ret.x;n.style.position="absolute";n.style.top=top+"px";n.style.left=left+"px";}};})(node);init();var anim=d.animateProperty(d.mixin({properties:{top:args.top||0,left:args.left||0}},args));d.connect(anim,"beforeBegin",anim,init);return anim;};})();}if(!dojo._hasResource["com.ibm.portal.utilities.html"]){dojo._hasResource["com.ibm.portal.utilities.html"]=true;dojo.provide("com.ibm.portal.utilities.html");dojo.require("com.ibm.portal.utilities");dojo.require("dojo.fx");com.ibm.portal.utilities.html={createAnchor:function(_2c3,href,id,_2c6,_2c7){ibm.portal.debug.entry("SkinRenderer.createAnchor",[_2c3,href,id,_2c6,_2c7]);var _2c8=document.createElement("A");_2c8.href=href;if(id){_2c8.id=id;}if(_2c7){_2c8.className=_2c7;}if(_2c6){_2c8.appendChild(document.createTextNode(_2c6));}_2c3.appendChild(_2c8);ibm.portal.debug.exit("SkinRenderer.createAnchor",[_2c8]);return _2c8;},createButton:function(_2c9,href,id,_2cc,_2cd){ibm.portal.debug.entry("SkinRenderer.createButton",[_2c9,href,id,_2cc,_2cd]);var _2ce=document.createElement("BUTTON");if(href){_2ce.href=href;}if(id){_2ce.id=id;}if(_2cd){_2ce.className=_2cd;}if(_2cc){_2ce.appendChild(document.createTextNode(_2cc));}_2c9.appendChild(_2ce);ibm.portal.debug.exit("SkinRenderer.createButton",[_2ce]);return _2ce;},createImage:function(_2cf,src,id,_2d2,_2d3){ibm.portal.debug.entry("SkinRenderer.createImage",[_2cf,src,id,_2d2,_2d3]);var img=document.createElement("IMG");img.src=src;if(id){img.id=id;}if(_2d2){img.alt=_2d2;img.setAttribute("title",_2d2);if(_2cf.nodeName=="BUTTON"){_2cf.setAttribute("title",_2d2);}}if(_2d3){img.className=_2d3;}_2cf.appendChild(img);ibm.portal.debug.exit("SkinRenderer.createImage",[img]);return img;},createImageAnchor:function(_2d5,src,id,_2d8,_2d9){ibm.portal.debug.entry("SkinRenderer.createImageAnchor",[_2d5,src,id,_2d8,_2d9]);var _2da=com.ibm.portal.utilities.html.createAnchor(_2d5,"javascript:void(0);");var img=document.createElement("IMG");img.src=src;if(id){img.id=id;}if(_2d8){img.alt=_2d8;img.title=_2d8;}if(_2d9){img.className=_2d9;}_2da.appendChild(img);ibm.portal.debug.exit("SkinRenderer.createImageAnchor",[img]);return _2da;},createTemporaryMarkupDiv:function(_2dc){ibm.portal.debug.entry("html.createTemporaryMarkupDiv");var _2dd={markup:_2dc,objects:{}};if(dojo.isIE){_2dd=com.ibm.portal.utilities.html.extractObjectElementsFromString(_2dc);}var div=document.createElement("DIV");div.innerHTML="<p style='display: none;'>&nbsp;</p>"+_2dd.markup;ibm.portal.debug.exit("html.createTemporaryMarkupDiv",[div]);return {node:div,objects:_2dd.objects};},extractObjectElementsFromString:function(_2df){var _2e0={};var _2e1=/<object/gi;var _2e2=/<\/object>/gi;var _2e3=_2df;var _2e4=null;try{_2e4=_2e1.exec(_2e3);if(_2e4&&_2e4.index>-1){var _2e5=_2e4.index;var buf;var end;var _2e8;var id;while(_2e5>-1){buf=_2e3.substring(0,_2e5);end=_2e3.indexOf(">",_2e5);if(_2e3.charAt(end-1)=="/"){_2e1.lastIndex=end;_2e4=_2e1.exec(_2e3);if(_2e4){_2e5=_2e4.index;continue;}else{break;}}_2e2.lastIndex=_2e5;_2e4=_2e2.exec(_2e3);if(_2e4){end=_2e4.index;}else{break;}_2e8=_2e3.substring(_2e5,end+9);id=dojo.dnd.getUniqueId();_2e3=buf+"<div id='"+id+"'></div>"+_2e3.substring(end+9);_2e0[id]=_2e8;_2e1.lastIndex=0;_2e4=_2e1.exec(_2e3);if(_2e4){_2e5=_2e4.index;}else{break;}}}_2df=_2e3;}catch(e){_2e0={};}return {markup:_2df,objects:_2e0};},replaceObjectElementsInMarkup:function(_2ea){for(var id in _2ea){var _2ec=dojo.byId(id);if(_2ec){_2ec.outerHTML=_2ea[id];}}},removeNodesOnCondition:function(node,_2ee){if(!_2ee){_2ee=function(){return false;};}if(node&&node.childNodes){for(var i=0;i<node.childNodes.length;i++){if(_2ee(node.childNodes[i])){var _2f0=node.childNodes[i];node.removeChild(_2f0);delete _2f0;i--;}else{this.removeNodesOnCondition(node.childNodes[i],_2ee);}}}},getElementsByTagNames:function(_2f1){ibm.portal.debug.entry("html.getElementsByTagNames",[_2f1]);var _2f2=new Array();for(var i=1;i<arguments.length;i++){var _2f4=_2f1.getElementsByTagName(arguments[i]);ibm.portal.debug.text("found "+_2f4.length+" "+arguments[i]+" tags.");for(var j=0;j<_2f4.length;j++){_2f2.push(_2f4[j]);}}ibm.portal.debug.exit("html.getElementsByTagNames",[_2f2]);return _2f2;},getX:function(elem){ibm.portal.debug.entry("html.getX",[elem]);var size=0;if(elem!=null){if(elem.offsetParent!=null){size+=com.ibm.portal.utilities.html.getX(elem.offsetParent);}if(elem!=null){size+=elem.offsetLeft;}}ibm.portal.debug.exit("html.getX",[size]);return size;},getY:function(elem){ibm.portal.debug.entry("html.getY"[elem]);var size=0;if(elem!=null){if(elem.offsetParent!=null){size+=com.ibm.portal.utilities.html.getY(elem.offsetParent);}if(elem!=null){size+=elem.offsetTop;}}ibm.portal.debug.exit("html.getY",[size]);return size;},convertFormToQuery:function(_2fa,_2fb){ibm.portal.debug.entry("html.convertFormToQuery",[_2fa,_2fb]);var _2fc=this.getElementsByTagNames(_2fa,"input","select","textarea","button");var _2fd="";var _2fe="&";var _2ff="=";var _300=0;for(var i=0;i<_2fc.length;i++){var _302=this.convertInputToNameValuePairs(_2fc[i],_2fb);for(var k=0;k<_302.length;k++){var pair=_302[k];if(pair.name!=""){if(_300!=0){_2fd+=_2fe;}_2fd+=encodeURIComponent(pair.name);for(var j=0;j<pair.values.length;j++){if(j==0){_2fd+=(_2ff+encodeURIComponent(pair.values[j]));}else{_2fd+=(_2fe+encodeURIComponent(pair.name)+_2ff+encodeURIComponent(pair.values[j]));}}_300=_300+1;}}}ibm.portal.debug.exit("html.convertFormToQuery",_2fd);return _2fd;},convertInputToNameValuePairs:function(_306,_307){ibm.portal.debug.entry("html.convertInputToNameValuePairs",[_306,_307]);var type=_306.type;ibm.portal.debug.text("Input type is: "+type);ibm.portal.debug.text("Input name is: "+_306.name);var name="";var _30a=[];var _30b=[];if(!_306.disabled){switch(type.toLowerCase()){case "text":case "password":case "hidden":name=_306.name;_30a.push(_306.value);_30b.push({name:name,values:_30a});break;case "reset":case "button":if(!_307||(_306.name==_307.name&&_306.value==_307.value)){name=_306.name;_30a.push(_306.value);_30b.push({name:name,values:_30a});}break;case "radio":case "checkbox":if(_306.checked){name=_306.name;_30a.push(_306.value);}_30b.push({name:name,values:_30a});break;case "image":if(!_307||_306.name==_307){name=_306.name;if(_306.value){_30a.push(_306.value);_30b.push({name:name,values:_30a});}_30b.push({name:name+".x",values:[this.getX(_306)]});_30b.push({name:name+".y",values:[this.getY(_306)]});}break;case "submit":if(!_307||(_306.name==_307.name&&_306.value==_307.value)){name=_306.name;if(_306.value){_30a.push(_306.value);}_30b.push({name:name,values:_30a});}break;case "select-one":case "select-multiple":name=_306.name;for(var i=0;i<_306.options.length;i++){if(_306.options[i].selected){var _30d=_306.options[i].value?_306.options[i].value:_306.options[i].text;_30a.push(_30d);}}if(_30a.length!=0){_30b.push({name:name,values:_30a});}break;case "file":break;default:name=_306.name;_30a.push(_306.value);_30b.push({name:name,values:_30a});}}ibm.portal.debug.exit("html.convertInputToNameValuePairs",_30b);return _30b;},isHidden:function(node){return dojo.style(node,"display")=="none";},hide:function(node){dojo.fx.wipeOut({node:node,duration:5}).play();},show:function(node){dojo.fx.wipeIn({node:node,duration:5}).play();},isDescendantOf:function(node,ref){var node=node.parentNode;var _313=false;while(node&&!_313){if(node==ref){_313=true;}node=node.parentNode;}return _313;}};}if(!dojo._hasResource["com.ibm.portal.services.PortalRestServiceRequest"]){dojo._hasResource["com.ibm.portal.services.PortalRestServiceRequest"]=true;dojo.provide("com.ibm.portal.services.PortalRestServiceRequest");dojo.require("com.ibm.ajax.auth");dojo.require("com.ibm.portal.EventBroker");dojo.require("com.ibm.portal.services.PortalRestServiceRequestQueue");dojo.declare("com.ibm.portal.services.ContentHandlerURL",null,{constructor:function(uri,_315,verb,_317){ibm.portal.debug.entry("ContentHandlerURL.constructor",[uri,_315,verb,_317]);if(uri==null){return null;}if(!_315){_315=2;}var _318=com.ibm.portal.navigation.controller.NAVIGATION_CONTROLLER.getState();var _319=_318.getLocale();if(_319){if(_317){_317+="&locale="+_319;}else{_317="&locale="+_319;}}this.url="";if(uri.charAt(0)=="?"){this.url=this._fromQueryString(uri,_317);}else{this.url=this._fromURI(uri,_315,"download",_317);}ibm.portal.debug.exit("ContentHandlerURL.constructor");},_fromQueryString:function(_31a,_31b){ibm.portal.debug.entry("fromQueryString",[_31a]);var str=ibmPortalConfig["contentHandlerURI"]+_31a;str=str.replace(/&amp;/g,"&");if(_31b){str=str+_31b;}if(str.indexOf("rep=compact")<0&&str.indexOf("rep=full")<0){str=str+"&rep=compact";}ibm.portal.debug.exit("fromQueryString",[str]);return str;},_fromURI:function(uri,_31e,verb,_320){ibm.portal.debug.entry("ContentHandlerURL._fromURI",[uri,_31e,verb,_320]);uri=com.ibm.portal.utilities.encodeURI(uri);var qStr="?uri="+uri;if(_31e){qStr=qStr+"&levels="+encodeURIComponent(_31e);}if(verb){qStr=qStr+"&mode="+encodeURIComponent(verb);}if(_320){qStr=qStr+_320;}if(qStr.indexOf("rep=compact")<0&&qStr.indexOf("rep=full")<0){qStr=qStr+"&rep=compact";}return this._fromQueryString(qStr);},getURI:function(){ibm.portal.debug.entry("ContentHandlerURL.getURI");return com.ibm.portal.utilities.decodeURI(this._extractParamValue("uri"));},getLevels:function(){return this._extractParamValue("levels");},getVerb:function(){return this._extractParamValue("verb");},_extractParamValue:function(_322){ibm.portal.debug.entry("ContentHandlerURL._extractParamValue",[_322]);var _323=this.url.indexOf(_322);var _324=this.url.indexOf("&",_323);var _325=this.url.slice(_323+_322.length+1,_324);ibm.portal.debug.exit("ContentHandlerURL._extractParamValue",[_325]);return _325;}});dojo.require("com.ibm.portal.utilities.html");dojo.declare("com.ibm.portal.services.PortalRestServiceForm",null,{method:"GET",isMultipart:false,encoding:"application/x-www-form-urlencoded",DomId:null,constructor:function(_326){if(_326.getAttributeNode("method")){this.method=_326.getAttributeNode("method").value;}if(_326.getAttributeNode("encType")){this.encoding=_326.getAttributeNode("encType").value;}if(_326.getAttributeNode("id")){this.DomId=_326.getAttributeNode("id").value;}else{DomId=_326;}this.isMultipart=(this.encoding=="multipart/form-data");},getDOMElement:function(){return dojo.byId(this.DomId);},submit:function(){this.getDOMElement().submit();},toQuery:function(){return com.ibm.portal.utilities.html.convertFormToQuery(this.getDOMElement());}});com.ibm.portal.services.REQUEST_QUEUE=new com.ibm.portal.services.PortalRestServiceRequestQueue();dojo.declare("com.ibm.portal.services.PortalRestServiceRequest",null,{constructor:function(_327,form,_329,sync){ibm.portal.debug.entry("PortalRestServiceRequest.constructor",[_327,form,_329,sync]);this._feedURI=_327.url;this._textOnly=_329;this._sync=sync;this._form=form;this._customResponseValidator=null;this._onauthenticated=null;if(!this._sync){this._sync=false;}ibm.portal.debug.exit("PortalRestServiceRequest.constructor");},cancelled:false,_deferred:undefined,setAuthenticationValidator:function(_32b){this._customResponseValidator=_32b;},setOnAuthenticatedHandler:function(_32c){this._onauthenticated=_32c;},create:function(data,_32e,_32f){if(!this.cancelled){this._doXmlHttpRequest("POST",data,_32e,_32f);}},read:function(_330,_331){ibm.portal.debug.entry("PortalRestServiceRequest.read",[_330,_331]);if(!this.cancelled){if(!this._sync){ibm.portal.debug.text("Queueing request!");var q=com.ibm.portal.services.REQUEST_QUEUE;var me=this;q.add({execute:function(_334){if(!me.cancelled){com.ibm.portal.EVENT_BROKER.startRequest.fire({uri:me._feedURI});var _335=function(arg1,arg2,arg3,arg4){_330(arg1,arg2,arg3,arg4);if(_334){_334();}};if(me._textOnly){me._retrieveRawFeed(_335,_331);}else{me._retrieve(_335,_331);}}else{if(_334){_334();}}}});}else{com.ibm.portal.EVENT_BROKER.startRequest.fire({uri:this._feedURI});if(this._textOnly){this._retrieveRawFeed(_330,_331);}else{this._retrieve(_330,_331);}}}ibm.portal.debug.exit("PortalRestServiceRequest.read");},update:function(data,_33b,_33c){if(!this.cancelled){this._doXmlHttpRequest("Put",data,_33b,_33c);}},remove:function(_33d,_33e){if(!this.cancelled){this._doXmlHttpRequest("Delete",null,_33d,_33e);}},cancel:function(){this.cancelled=true;if(this._deferred!==undefined){this._deferred.cancel();}},_retrieveRawFeed:function(_33f,_340){ibm.portal.debug.entry("_retrieveRawFeed",[_33f,_340]);var me=this;dojo.xhrGet({url:this._feedURI,load:function(type,data,evt){_33f(data,_340);com.ibm.portal.EVENT_BROKER.endRequest.fire({uri:me._feedURI});},sync:this._sync});ibm.portal.debug.exit("_retrieveRawFeed");},_retrieve:function(_345,_346,_347,_348){ibm.portal.debug.entry("_retrieve",[_345]);if(this._form&&this._form.isMultipart){this._doIframeRequest(_345,_346);}else{this._doXmlHttpRequest("Get",null,_345,_346);}ibm.portal.debug.exit("PortalRestServiceRequest._retrieve");},_doIframeRequest:function(_349,_34a){ibm.portal.debug.entry("PortalRestServiceRequest._doIframeRequest",[_349]);var _34b=null;var _34c=dojo.dnd.getUniqueId();if(dojo.isIE){_34b=document.createElement("<iframe name='"+_34c+"' id='"+_34c+"' src='about:blank' onload='com.ibm.portal.aggregation.forms.PORTLET_FORM_HANDLER.handleMultiPartResult(this.id);'></iframe>");com.ibm.portal.aggregation.forms.PORTLET_FORM_HANDLER._callbackfns[_34c]={fn:_349,args:_34a};var url=new com.ibm.portal.utilities.HttpUrl(this._feedURI);url.addParameter("ibm.web2.contentType","text/plain");this._form.getDOMElement().setAttribute("action",url.toString());}else{ibm.portal.debug.text("Creating the iframe... name is: "+_34c+"; url is: "+this._feedURI);_34b=document.createElement("IFRAME");_34b.setAttribute("name",_34c);_34b.setAttribute("id",_34c);var me=this;_34b.onload=function(){var xml=window.frames[_34c].document;_349("load",xml,null,_34a);com.ibm.portal.EVENT_BROKER.endRequest.fire({uri:me._feedURI});};this._form.getDOMElement().setAttribute("action",this._feedURI);}_34b.style.visibility="hidden";_34b.style.height="1px";_34b.style.width="1px";document.body.appendChild(_34b);if(window.frames[_34c].name!=_34c){window.frames[_34c].name=_34c;}ibm.portal.debug.text("Setting the iframe target attribute to: "+_34c);this._form.getDOMElement().setAttribute("target",_34c);this._form.submit();ibm.portal.debug.exit("PortalRestServiceRequest._doIframeRequest");},_doXmlHttpRequest:function(_350,body,_352,_353){ibm.portal.debug.entry("PortalRestServiceRequest._doXmlHttpRequest",[_350,body,_352,_353]);ibm.portal.debug.text("Attempting to retrieve: "+this._feedURI+" using method: "+_350+"; synchronously? "+this._sync);var me=this;var args={url:this._feedURI,content:{},headers:{"X-IBM-XHR":"true"},handle:function(_356,_357){ibm.portal.debug.entry("PortalRestServiceRequest.handle",[_356,_357]);if(_356 instanceof Error&&_356.dojoType==="cancel"){_352("cancel",_356,null,_353);return;}var xhr=_357.xhr;ibm.portal.debug.text("XHR object: "+xhr);var _359=com.ibm.portal.services.PortalRestServiceConfig;var _35a=xhr.getResponseHeader("X-Request-Digest");if(_35a){_359.digest=_35a;}if(xhr.status==200){var data=_356;var loc=xhr.getResponseHeader("IBM-Web2-Location");if(loc){if(loc.indexOf(ibmPortalConfig["portalProtectedURI"])>=0&&me._feedURI.indexOf(ibmPortalConfig["portalPublicURI"])>=0){top.location.href=loc;return;}}var _35d=xhr.getResponseHeader("Content-Type");ibm.portal.debug.text("content-type is: "+_35d);if(/^text\/html/.exec(_35d)&&loc&&(loc.indexOf(ibmPortalConfig["portalProtectedURI"])>-1||loc.indexOf(ibmPortalConfig["portalPublicURI"])>-1)){ibm.portal.debug.text("content-type is text .. follow IBM-Web2-Location");top.location.href=loc;return;}var auth=com.ibm.ajax.auth;var _35f=false;if(me._customResponseValidator){_35f=me._customResponseValidator(_356,_357);}if(!_35f){_35f=auth.isAuthenticationRequired(_356,_357);}if(_35f){auth.authenticationHandler(_356,_357,me._onauthenticated);return;}ibm.portal.debug.text("Read feed: "+me._feedURI);if(dojo.isIE){var doc=dojox.data.dom.createDocument(data);_352("load",doc,xhr,_353);}else{_352("load",data,xhr,_353);}}else{if(xhr.status==401||xhr.status==0){ibm.portal.debug.text("Basic auth 401 found, trigger reload");com.ibm.ajax.auth.authenticationHandler();return;}else{_352("error",_356,xhr,_353);}}com.ibm.portal.EVENT_BROKER.endRequest.fire({uri:me._feedURI});ibm.portal.debug.exit("PortalRestServiceRequest.handle");},sync:this._sync,handleAs:"xml"};if(this._form){args.content=dojo.queryToObject(this._form.toQuery());_350=this._form.method;}_350=_350.toUpperCase();if(_350!="GET"&&_350!="POST"){if(ibmPortalConfig&&ibmPortalConfig.xMethodOverride){args.headers["X-Method-Override"]=_350.toUpperCase();_350="Post";}}if(_350=="PUT"&&body){args.putData=body;}else{if(_350=="POST"&&body){args.postData=body;}}if(dojo.isIE){args.content["ibm.web2.contentType"]="text/xml";args.handleAs="text";}var _361=com.ibm.portal.services.PortalRestServiceConfig;if(_361.timeout){args.timeout=_361.timeout;}if(_361.digest){args.content["digest"]=_361.digest;}_350=com.ibm.portal.utilities.string.properCase(_350);var _362=dojo["xhr"+_350];if(_362){this._deferred=_362(args);}else{throw new Error("Invalid request method attempted: "+_350);}ibm.portal.debug.exit("PortalRestServiceRequest._doXmlHttpRequest");},toString:function(){return this._feedURI;}});com.ibm.portal.services.PortalRestServiceConfig={timeout:null,digest:null};com.ibm.ajax.auth.setAuthenticationHandler(function(){if(typeof (document.isCSA)=="undefined"){top.location.reload();}else{ibm.portal.debug.entry("DefaultAuthenticationHandler");ibm.portal.debug.text("Illegal response content-type detected!");ibm.portal.debug.text("Parameterized redirect URL is: "+ibmPortalConfig["contentModelBlankURL"]);var _363=com.ibm.portal.navigation.controller.NAVIGATION_CONTROLLER.getState();var _364=ibmPortalConfig["contentModelBlankURL"].replace("-----oid-----",_363.getPageSelection());ibm.portal.debug.text("fullPageRefreshURL is currently: "+_364);if(dojo.cookie("WASReqURL")!=null){var _365=_363.createLinkToCurrentState();var _366="WASReqURL="+_365+"; path=/";document.cookie=_366;}ibm.portal.debug.text("Redirecting to: "+_364);com.ibm.portal.EVENT_BROKER.redirect.fire({url:_364});top.location.href=_364;ibm.portal.debug.exit("DefaultAuthenticationHandler");}});}if(!dojo._hasResource["com.ibm.portal.services.PortletFragmentService"]){dojo._hasResource["com.ibm.portal.services.PortletFragmentService"]=true;dojo.provide("com.ibm.portal.services.PortletFragmentService");dojo.require("dojox.data.dom");dojo.require("com.ibm.portal.services.PortalRestServiceRequest");dojo.require("com.ibm.portal.utilities");dojo.require("com.ibm.portal.debug");dojo.require("com.ibm.portal.EventBroker");dojo.declare("com.ibm.portal.services.PortletFragmentURL",null,{constructor:function(uri){if(uri.indexOf("?uri=")==0){this.url=ibmPortalConfig["portalURI"]+uri;this.url=this.url.replace(/&amp;/g,"&");this.url=this.url.replace(/lm:/,"pm:");}else{if(uri.indexOf("lm:")==0){this.url=ibmPortalConfig["portalURI"]+"?uri=fragment:"+uri;this.url=this.url.replace(/lm:/,"pm:");}else{this.url=uri;}}}});dojo.declare("com.ibm.portal.services.PortletInfo",null,{constructor:function(wId,pId,_36a,_36b,_36c,_36d,_36e,_36f,_370,_371,_372,_373){ibm.portal.debug.entry("PortletInfo.constructor",[wId,pId,_36a,_36b,_36c,_36d,_36f,_373]);this.windowId=wId;this.portletId=pId;this.uri="fragment:pm:oid:"+wId+"@oid:"+pId;this.markup=_36a;this.portletModes=_36b;this.windowStates=_36c;this.dependentPortlets=_36d;this.otherPortlets=_36e;this.stateVaryExpressions=_370;this.updatedState=_36f;this.currentMode=_371;this.currentWindowState=_372;this.portletTitle=_373;ibm.portal.debug.exit("PortletInfo.constructor");}});dojo.declare("com.ibm.portal.services.PortletFragmentService",null,{namespaces:{"xsl":"http://www.w3.org/1999/XSL/Transform","thr":"http://purl.org/syndication/thread/1.0","atom":"http://www.w3.org/2005/Atom","xhtml":"http://www.w3.org/1999/xhtml","model":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.0.1/portal-model-elements","base":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.0/ibm-portal-composite-base","portal":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.0.1/portal-model","xsi":"http://www.w3.org/2001/XMLSchema-instance","state":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state","state-vary":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.1/portal-state-vary"},activeRequests:{},constructor:function(){this.staticContext=com.ibm.portal.services.PortletFragmentService.prototype;},_flagPortletUrl:function(url,_375){ibm.portal.debug.entry("PortletFragmentService._flagPortletUrl",[url]);var _376=url.indexOf("uri=fragment:pm:oid:");var _377=new com.ibm.portal.utilities.HttpUrl(url);_377.addParameter("ibm.web2.keepRenderMode","false");if(_376<0){_375=_375.replace(/lm:/g,"fragment:pm:");_377.addParameter("uri",_375);}ibm.portal.debug.exit("PortletFragmentService._flagPortletUrl",[_377.toString()]);return _377.toString();},getPortletInfo:function(_378,_379,_37a,form,_37c){ibm.portal.debug.entry("PortletFragmentService.getPortletInfo",[_378,_379,_37a,form,_37c]);if(_379=="#"||_379==window.location.href+"#"){ibm.portal.debug.text("Illegal portlet url provided: "+_379);ibm.portal.debug.text("Aborting request.");return false;}if(com.ibm.portal.utilities.isJavascriptUrl(_379)){return eval(_379);}var _37d=_379;if(_37d.indexOf(top.location.href)==0){_37d=_37d.substring(top.location.href.length);while(_37d.length>0&&_37d.charAt(0)=="/"){_37d=_37d.substring(1);}}if(_37d.indexOf("?")==0){var _37e=com.ibm.portal.navigation.controller.NAVIGATION_CONTROLLER.getState();_379=_37e.resolveRelativePortletURL(_37d);}if(com.ibm.portal.utilities.isExternalUrl(_379)){self.location.href=_379;}else{var url={url:this._flagPortletUrl(_379,_378)};var _380=ibmPortalConfig.enforceOneActivePortletRequest;if(_380){var _381=this.staticContext.activeRequests;if(_381[_378]!==undefined&&_381[_378]!==null){_381[_378].cancel();com.ibm.portal.EVENT_BROKER.cancelFragmentUpdate.fire({id:_378});_381[_378]=null;}}var _382=new com.ibm.portal.services.PortalRestServiceRequest(url,form);if(!_37c){com.ibm.portal.EVENT_BROKER.startFragment.fire({id:_378});}if(_380){_381[_378]=_382;}var me=this;_382.read(function(type,_385,xhr){if(_380){_381[_378]=null;}if(!_382.cancelled){var _387=null;if(type=="load"){_387=me.createPortletInfo(_385);}if(_385 instanceof Error){_387=_385;}if(!_37c){me._fireEvents(_387,_378,xhr);}if(_37a){_37a(_387,xhr);}}});}ibm.portal.debug.exit("PortletFragmentService.getPortletInfo");},readWindowID:function(_388){ibm.portal.debug.entry("PortletFragmentService.readWindowID",[_388]);var _389="/atom:feed/atom:entry/atom:id";var _38a=com.ibm.portal.xpath.evaluateXPath(_389,_388,this.namespaces);var _38b=dojox.data.dom.textContent(_38a[0]);ibm.portal.debug.exit("PortletFragmentService.readWindowID",[_38b.substring(4)]);return _38b.substring(4);},readPortletID:function(_38c){ibm.portal.debug.entry("PortletFragmentService.readPortletID",[_38c]);var _38d="/atom:feed/atom:id";var _38e=com.ibm.portal.xpath.evaluateXPath(_38d,_38c,this.namespaces);var _38f=dojox.data.dom.textContent(_38e[0]);ibm.portal.debug.exit("PortletFragmentService.readPortletID",[_38f.substring(4)]);return _38f.substring(4);},readMarkup:function(_390){ibm.portal.debug.entry("PortletFragmentService.readMarkup",[_390]);var _391="/atom:feed/atom:entry/atom:content";var _392=com.ibm.portal.xpath.evaluateXPath(_391,_390,this.namespaces);var _393="";if(_392!=null&&_392.length>0){_393=dojox.data.dom.textContent(_392[0]);}ibm.portal.debug.exit("PortletFragmentService.readMarkup",[_393]);return _393;},readPortletModes:function(_394){ibm.portal.debug.entry("PortletFragmentService.readPortletModes",[_394]);var _395="/atom:feed/atom:entry/atom:link[@portal:rel='portlet-mode']";var _396=com.ibm.portal.xpath.evaluateXPath(_395,_394,this.namespaces);var _397=new Array();if(_396!=null&&_396.length>0){var _398=_396.length;for(var i=0;i<_398;i++){_397.push({"link":_396[i].getAttribute("href"),"mode":_396[i].getAttribute("title")});}}ibm.portal.debug.exit("PortletFragmentService.readPortletModes",[_397]);return _397;},readWindowStates:function(_39a){ibm.portal.debug.entry("PortletFragmentService.readWindowStates",[_39a]);var _39b="/atom:feed/atom:entry/atom:link[@portal:rel='window-state']";var _39c=com.ibm.portal.xpath.evaluateXPath(_39b,_39a,this.namespaces);var _39d=new Array();if(_39c!=null&&_39c.length>0){var _39e=_39c.length;for(var i=0;i<_39e;i++){_39d.push({"link":_39c[i].getAttribute("href"),"mode":_39c[i].getAttribute("title")});}}ibm.portal.debug.exit("PortletFragmentService.readWindowStates",[_39d]);return _39d;},readDependentPortlets:function(_3a0){ibm.portal.debug.entry("PortletFragmentService.readDependentPortlets",[_3a0]);var _3a1="/atom:feed/atom:link[@portal:rel='dependent']";var _3a2=com.ibm.portal.xpath.evaluateXPath(_3a1,_3a0,this.namespaces);var _3a3=new Array();if(_3a2!=null&&_3a2.length>0){var _3a4=_3a2.length;for(var i=0;i<_3a4;i++){_3a3.push({"link":_3a2[i].getAttribute("href"),"portlet":_3a2[i].getAttribute("title"),"uri":_3a2[i].getAttribute("portal:uri")?_3a2[i].getAttribute("portal:uri"):_3a2[i].getAttribute("uri")});}}ibm.portal.debug.exit("PortletFragmentService.readDependentPortlets",[_3a3]);return _3a3;},readOtherPortlets:function(_3a6){ibm.portal.debug.entry("PortletFragmentService.readOtherPortlets",[_3a6]);var _3a7="/atom:feed/atom:link[@portal:rel='other']";var _3a8=com.ibm.portal.xpath.evaluateXPath(_3a7,_3a6,this.namespaces);var _3a9=new Array();if(_3a8!=null&&_3a8.length>0){var _3aa=_3a8.length;for(var i=0;i<_3aa;i++){_3a9.push({"link":_3a8[i].getAttribute("href"),"portlet":_3a8[i].getAttribute("title"),"uri":_3a8[i].getAttribute("portal:uri")});}}ibm.portal.debug.exit("PortletFragmentService.readOtherPortlets",[_3a9]);return _3a9;},readStateVaryExpressions:function(_3ac){ibm.portal.debug.entry("PortletFragmentService.readStateVaryExpressions",[_3ac]);var _3ad="/atom:feed/atom:entry/state-vary:state-vary/state-vary:expr";var _3ae=com.ibm.portal.xpath.evaluateXPath(_3ad,_3ac,this.namespaces);var _3af=new Array();if(_3ae!=null&&_3ae.length>0){var _3b0=_3ae.length;for(var i=0;i<_3b0;i++){var _3b2=_3ae[i].firstChild;if(_3b2!=null){_3af.push(_3b2.nodeValue);}}}ibm.portal.debug.exit("PortletFragmentService.readStateVaryExpressions",[_3af]);return _3af;},readPortletState:function(_3b3){return this._readPortletState(_3b3);},_readPortletState:function(_3b4){ibm.portal.debug.entry("PortletFragmentService.readPortletState",[_3b4]);var _3b5="/atom:feed/atom:entry/state:root";var _3b6=com.ibm.portal.xpath.evaluateXPath(_3b5,_3b4,this.namespaces);var _3b7=null;if(_3b6!=null&&_3b6.length>0){var doc=dojox.data.dom.createDocument();com.ibm.portal.utilities.addExternalNode(doc,_3b6[0]);_3b7=doc;}else{_3b5="/atom:feed/state:root";_3b6=com.ibm.portal.xpath.evaluateXPath(_3b5,_3b4,this.namespaces);if(_3b6!=null&&_3b6.length>0){var doc=dojox.data.dom.createDocument();com.ibm.portal.utilities.addExternalNode(doc,_3b6[0]);_3b7=doc;}}ibm.portal.debug.exit("PortletFragmentService.readPortletState",[_3b7]);return _3b7;},readPortletTitle:function(_3b9){return this._readPortletTitle(_3b9);},_readPortletTitle:function(_3ba){ibm.portal.debug.entry("PortletFragmentService.readPortletTitle",[_3ba]);var _3bb="/atom:feed/atom:entry/atom:title";var _3bc=com.ibm.portal.xpath.evaluateXPath(_3bb,_3ba,this.namespaces);var _3bd=dojox.data.dom.textContent(_3bc[0]);ibm.portal.debug.exit("PortletFragmentService.readPortletTitle",_3bd);return _3bd;},_fireEvents:function(_3be,_3bf,xhr){this._fireGlobalPortletStateChange(_3be,_3bf,xhr);},_fireGlobalPortletStateChange:function(_3c1,_3c2,xhr){com.ibm.portal.EVENT_BROKER.endFragment.fire({portletInfo:_3c1,id:_3c2,xhr:xhr});},_fireIndividualPortletStateChange:function(_3c4){},createPortletInfo:function(_3c5){var _3c6=this.readWindowID(_3c5);var _3c7=this.readPortletID(_3c5);var _3c8=this.readMarkup(_3c5);var _3c9=this.readPortletModes(_3c5);var _3ca=this.readWindowStates(_3c5);var _3cb=this.readDependentPortlets(_3c5);var _3cc=this.readOtherPortlets(_3c5);var _3cd=this.readPortletState(_3c5);var _3ce=this.readStateVaryExpressions(_3c5);var _3cf=this.readPortletTitle(_3c5);var _3d0=_3cd;if(_3d0==null){_3d0=this._readPortletState(_3c5);}var _3d1=new com.ibm.portal.state.StateManager();var _3d2=_3d1.newPortletAccessor(_3c6,_3d0);var mode=_3d2.getPortletMode();var _3d4=_3d2.getWindowState();return new com.ibm.portal.services.PortletInfo(_3c6,_3c7,_3c8,_3c9,_3ca,_3cb,_3cc,_3cd,_3ce,mode,_3d4,_3cf);}});dojo.declare("com.ibm.portal.services.IndependentPortletFragmentService",com.ibm.portal.services.PortletFragmentService,{readDependentPortlets:function(_3d5){ibm.portal.debug.entry("DependentPortletFragmentService.readDependentPortlets",[_3d5]);var _3d6=new Array();ibm.portal.debug.exit("DependentPortletFragmentService.readDependentPortlets",[_3d6]);return _3d6;},readOtherPortlets:function(_3d7){ibm.portal.debug.entry("DependentPortletFragmentService.readOtherPortlets",[_3d7]);var _3d8=new Array();ibm.portal.debug.exit("DependentPortletFragmentService.readOtherPortlets",[_3d8]);return _3d8;},readPortletState:function(_3d9){return null;}});}if(!dojo._hasResource["ibm.portal.portlet.portlet"]){dojo._hasResource["ibm.portal.portlet.portlet"]=true;dojo.provide("ibm.portal.portlet.portlet");ibm.portal.portlet._SafeToExecute=false;if(window.addEventListener){window.addEventListener("load",function(){ibm.portal.portlet._SafeToExecute=true;},false);}else{if(window.attachEvent){window.attachEvent("onload",function(){ibm.portal.portlet._SafeToExecute=true;});}}dojo.declare("ibm.portal.portlet.PortletWindow",null,{STATUS_UNDEFINED:0,STATUS_OK:1,STATUS_ERROR:2,constructor:function(_3da){if(_3da==null){return;}this.windowID=_3da;var _3db=document.getElementById("com.ibm.wps.web2.portlet.preferences."+this.windowID);this.preferenceEditID=_3db.getAttribute("editid");this.preferenceConfigID=_3db.getAttribute("configid");this.preferenceEditDefaultsID=_3db.getAttribute("editdefaultsid");this.pageID=_3db.getAttribute("pageid");this.attributes=new Array();this._queuedFuncs=new Array();this.portletState=new ibm.portal.portlet.PortletState(_3da);this.isCSA=false;try{this.isCSA=(typeof (document.isCSA)!="undefined");}catch(e){}var me=this;function executeQueued(){for(var i=0;i<me._queuedFuncs.length;i++){me._queuedFuncs[i]();}};if(window.addEventListener){window.addEventListener("load",function(){if(!ibm.portal.portlet._SafeToExecute){ibm.portal.portlet._SafeToExecute=true;}executeQueued();},false);}else{if(window.attachEvent){window.attachEvent("onload",function(){if(!ibm.portal.portlet._SafeToExecute){ibm.portal.portlet._SafeToExecute=true;}executeQueued();});}}},reportError:function(_3de){var code;if(_3de.getErrorCode()==ibm.portal.portlet.Error.ERROR){code="error";}else{if(_3de.getErrorCode()==ibm.portal.portlet.Error.INFO){code="info";}else{if(_3de.getErrorCode()==ibm.portal.portlet.Error.WARN){code="warning";}}}var _3e0={"_type":code,"_message":_3de.getMessage(),"_details":_3de.getDescription()};if(this.isCSA){dojo.publish("/portal/status",[{message:_3e0}]);}else{if(typeof (console)!="undefined"){if(_3de.getErrorCode()==ibm.portal.portlet.Error.ERROR){console.error(_3e0._message+"\n"+_3e0._details);}else{if(_3de.getErrorCode()==ibm.portal.portlet.Error.INFO){console.info(_3e0._message+"\n"+_3e0._details);}else{if(_3de.getErrorCode()==ibm.portal.portlet.Error.WARN){console.warn(_3e0._message+"\n"+_3e0._details);}}}}else{alert(_3e0._type.toUpperCase()+"\nMessage: "+_3e0._message+"\nDetails: "+_3e0._details);}}},getAttribute:function(name){return this.attributes[name];},setAttribute:function(name,_3e3){var ret=this.attributes[name];this.attributes[name]=_3e3;return ret;},removeAttribute:function(name){this.attributes[name]=null;},clearAttributes:function(){this.attributes=new Array();},getPortletState:function(_3e6){var _3e7=this.portletState;var _3e8=this;var _3e9=null;if(_3e6!=null){_3e6(_3e8,ibm.portal.portlet.PortletWindow.STATUS_OK,_3e7);}else{_3e9={"portletWindow":_3e8,"status":ibm.portal.portlet.PortletWindow.STATUS_OK,"returnObject":_3e7};}return _3e9;},setPortletState:function(_3ea,_3eb){this.portletState=_3ea;if(this.isCSA){if(_3eb==null){var _3ec=com.ibm.portal.navigation.controller.NAVIGATION_CONTROLLER.getState();var url=_3ec.newPortletRenderURL(this.windowID);var _3ee=new com.ibm.portal.services.PortletFragmentService();_3ee.getPortletInfo("lm:oid:"+this.windowID+"@oid:"+this.pageID,url);}}else{var _3ef=new com.ibm.portal.state.StateManager(ibmPortalConfig["contentHandlerURI"]);_3ef.reset(_3ea.portletAccessor.stateDOM);var _3f0=_3ef.getSerializationManager();var _3f1=_3f0.serialize(_3ef.getState());var _3f2=_3f1["returnObject"];var url=_3f2;window.location.href=url;}return this.getPortletState(_3eb);},_queueUp:function(_3f3){this._queuedFuncs.push(_3f3);},_throwInappropriateRequestError:function(_3f4){throw new Error("Cannot execute a synchronous call before the page loads! Please use an onload handler to execute this call to \""+_3f4+"\".");return null;},getPortletPreferences:function(_3f5){if(!ibm.portal.portlet._SafeToExecute){if(_3f5){var me=this;this._queueUp(function(){me.getPortletPreferences(_3f5);});return false;}else{return this._throwInappropriateRequestError("getPortletPreferences");}}var _3f7=this.getPortletState().returnObject.getPortletMode();this.status=ibm.portal.portlet.PortletWindow.STATUS_UNDEFINED;var _3f8=document.getElementById("com.ibm.wps.web2.portlet.root."+this.windowID).innerHTML;var idx=_3f8.indexOf("--portletwindowid--");var _url=_3f8.replace(/--portletwindowid--/g,this.windowID);if(_url.indexOf("?")<0){_url=_url+"?";}_url+="&verb=download&levels=-all&rep=compact&preferences=aggregated";this.requestedPreferenceID="pm:oid:"+this.preferenceEditID;if(_3f7==ibm.portal.portlet.PortletMode.CONFIG){this.requestedPreferenceID="pm:oid:"+this.preferenceConfigID;}else{if(_3f7==ibm.portal.portlet.PortletMode.EDIT_DEFAULTS){this.requestedPreferenceID="pm:oid:"+this.preferenceEditDefaultsID;}}var _3fb=this;var _3fc=null;dojo.xhrGet({url:_url,handleAs:"xml",headers:{"X-IBM-XHR":"true","If-Modified-Since":"Thu, 1 Jan 1970 00:00:00 GMT"},sync:(_3f5)?false:true,handle:function(_3fd,_3fe){if(_3fb.isAuthenticationRequired(_3fe.xhr,_3fe.args.handleAs)){_3fb.doAuthentication();}else{var type=(_3fd instanceof Error)?"error":"load";if(type=="load"){var _400=_3fd;if(!_400||(typeof (dojox.data.dom.innerXML(_3fd))=="undefined")){_400=dojox.data.dom.createDocument(_3fe.xhr.responseText);}var _401=new ibm.portal.portlet.PortletPreferences(_3fb.windowID,_3fb.requestedPreferenceID,_400);if(_3f5){_3f5(_3fb,ibm.portal.portlet.PortletWindow.STATUS_OK,_401);}else{_3fc={"portletWindow":_3fb,"status":ibm.portal.portlet.PortletWindow.STATUS_OK,"returnObject":_401};}}else{if(type=="error"){if(_3f5){_3f5(_3fb,ibm.portal.portlet.PortletWindow.STATUS_ERROR,null);}else{_3fc={"portletWindow":_3fb,"status":ibm.portal.portlet.PortletWindow.STATUS_ERROR,"returnObject":null};}}}}},transport:"XMLHTTPTransport"});return _3fc;},setPortletPreferences:function(_402,_403){if(!ibm.portal.portlet._SafeToExecute){if(_403){var me=this;this._queueUp(function(){me.setPortletPreferences(_402,_403);});return false;}else{return this._throwInappropriateRequestError("setPortletPreferences");}}this.status=ibm.portal.portlet.PortletWindow.STATUS_UNDEFINED;var _405=document.getElementById("com.ibm.wps.web2.portlet.root."+this.windowID).innerHTML;var idx=_405.indexOf("--portletwindowid--");var _url=_405.replace(/--portletwindowid--/g,this.windowID);if(_url.indexOf("?")<0){_url+="?verb=download";}else{_url+="&verb=download";}var _408=_402.requestedPreferenceID;var expr="/atom:feed/atom:entry[atom:id='"+_408+"']";var _40a=ibm.portal.xml.xpath.evaluateXPath(expr,_402.xmlData,_402.ns);var _40b;if(_40a&&_40a.length>0){_40b=_40a[0];}else{return null;}var _40c=_40b.parentNode;expr="/atom:feed/atom:entry";_40a=ibm.portal.xml.xpath.evaluateXPath(expr,_402.xmlData,_402.ns);for(var i=0;i<_40a.length;i++){var node=_40a[i];if(node!=_40b){_40c.removeChild(node);}}var _40f=this;var _410=null;dojo.rawXhrPut({url:_url,sync:(_403)?false:true,putData:dojox.data.dom.innerXML(_402.xmlData),contentType:"application/xml",headers:{"X-IBM-XHR":"true"},handleAs:"xml",handle:function(_411,_412){if(_40f.isAuthenticationRequired(_412.xhr,_412.args.handleAs)){_40f.doAuthentication();}else{var type=(_411 instanceof Error)?"error":"load";if(type=="load"){if(_403){_403(_40f,ibm.portal.portlet.PortletWindow.STATUS_OK,_402);}else{_410={"portletWindow":_40f,"status":ibm.portal.portlet.PortletWindow.STATUS_OK,"returnObject":_402};}}else{if(type=="error"){if(_403){_403(_40f,ibm.portal.portlet.PortletWindow.STATUS_ERROR,null);}else{_410={"portletWindow":_40f,"status":ibm.portal.portlet.PortletWindow.STATUS_ERROR,"returnObject":null};}}}}},transport:"XMLHTTPTransport"});return _410;},getUserProfile:function(_414){if(!ibm.portal.portlet._SafeToExecute){if(_414){var me=this;this._queueUp(function(){me.getUserProfile(_414);});return false;}else{return this._throwInappropriateRequestError("getUserProfile");}}this.status=ibm.portal.portlet.PortletWindow.STATUS_UNDEFINED;var _url=document.getElementById("com.ibm.wps.web2.portlet.user."+this.windowID).innerHTML;var _417=this;var _418=null;dojo.xhrGet({url:_url,headers:{"X-IBM-XHR":"true","If-Modified-Since":"Thu, 1 Jan 1970 00:00:00 GMT"},sync:(_414)?false:true,handleAs:"xml",handle:function(_419,_41a){if(_417.isAuthenticationRequired(_41a.xhr,_41a.args.handleAs)){_417.doAuthentication();}else{var type=(_419 instanceof Error)?"error":"load";if(type=="load"){var _41c=_419;if(!_41c||(typeof (dojox.data.dom.innerXML(_419))=="undefined")){_41c=dojox.data.dom.createDocument(_41a.xhr.responseText);}var _41d=new ibm.portal.portlet.UserProfile(_417.windowID,_41c);if(_414){_414(_417,ibm.portal.portlet.PortletWindow.STATUS_OK,_41d);}else{_418={"portletWindow":_417,"status":ibm.portal.portlet.PortletWindow.STATUS_OK,"returnObject":_41d};}}else{if(type=="error"){if(_414){_414(_417,ibm.portal.portlet.PortletWindow.STATUS_ERROR,null);}else{_418={"portletWindow":_417,"status":ibm.portal.portlet.PortletWindow.STATUS_ERROR,"returnObject":null};}}}}},transport:"XMLHTTPTransport"});return _418;},setUserProfile:function(_41e,_41f){if(!ibm.portal.portlet._SafeToExecute){if(_41f){var me=this;this._queueUp(function(){me.setUserProfile(_41e,_41f);});return false;}else{return this._throwInappropriateRequestError("setUserProfile");}}this.status=ibm.portal.portlet.PortletWindow.STATUS_UNDEFINED;var _url=document.getElementById("com.ibm.wps.web2.portlet.user."+this.windowID).innerHTML;var _422=this;var _423=null;dojo.rawXhrPost({url:_url,sync:(_41f)?false:true,postData:dojox.data.dom.innerXML(_41e.xmlData),contentType:"application/xml",headers:{"X-IBM-XHR":"true"},handleAs:"xml",handle:function(_424,_425){if(_422.isAuthenticationRequired(_425.xhr,_425.args.handleAs)){_422.doAuthentication();}else{var type=(_424 instanceof Error)?"error":"load";if(type=="load"){if(_41f){_41f(_422,ibm.portal.portlet.PortletWindow.STATUS_OK,_41e);}else{_423={"portletWindow":_422,"status":ibm.portal.portlet.PortletWindow.STATUS_OK,"returnObject":_41e};}}else{if(type=="error"){if(_41f){_41f(_422,ibm.portal.portlet.PortletWindow.STATUS_ERROR,null);}else{_423={"portletWindow":_422,"status":ibm.portal.portlet.PortletWindow.STATUS_ERROR,"returnObject":null};}}}}},transport:"XMLHTTPTransport"});return _423;},newXMLPortletRequest:function(){return new ibm.portal.portlet.XMLPortletRequest(this);},isAuthenticationRequired:function(_427,_428){if(_427.readyState!=4){throw new Error("isAuthenticationRequired should only be called with a COMPLETED XMLHttpRequest! The readyState on the given XMLHttpRequest is not 4 (COMPLETE)!");}var _429={dojoType:"valid"};var _42a={xhr:_427,args:{handleAs:_428}};return com.ibm.ajax.auth.isAuthenticationRequired(_429,_42a);},setAuthenticationHandler:function(_42b){this._authenticationFn=_42b;},doAuthentication:function(){if(this._authenticationFn){this._authenticationFn();}else{com.ibm.ajax.auth.authenticationHandler();}}});dojo.declare("ibm.portal.portlet.PortletPreferences",null,{constructor:function(_42c,_42d,data){this.windowID=_42c;this.requestedPreferenceID=_42d;this.xmlData=data;this.xsltURL=dojo.moduleUrl("ibm","portal/portlet/");this.ns={"xsl":"http://www.w3.org/1999/XSL/Transform","thr":"http://purl.org/syndication/thread/1.0","atom":"http://www.w3.org/2005/Atom","xhtml":"http://www.w3.org/1999/xhtml","model":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.0.1/portal-model-elements","base":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.0/ibm-portal-composite-base","portal":"http://www.ibm.com/xmlns/prod/websphere/portal/v6.0.1/portal-model","xsi":"http://www.w3.org/2001/XMLSchema-instance"};this.internal_reset();},getMap:function(){if(this.result_getMap){return this.result_getMap;}var _42f=ibm.portal.xml.xslt.loadXsl(this.xsltURL+"PortletPreferencesMap.xsl");if(_42f.documentElement==null){alert("xslDoc is null");}var _430=ibm.portal.xml.xslt.transform(this.xmlData,_42f,null,{"selectionid":this.requestedPreferenceID},true);if(_430==null){this.result_getNames=null;return null;}var _431=eval(_430);if(_431){_431=_431.preferences;}this.result_getMap=_431;return this.result_getMap;},getNames:function(){if(this.result_getNames){return this.result_getNames;}var _432=ibm.portal.xml.xslt.loadXsl(this.xsltURL+"PortletPreferencesNames.xsl");if(_432.documentElement==null){alert("xslDoc is null");}var _433=ibm.portal.xml.xslt.transform(this.xmlData,_432,null,{"selectionid":this.requestedPreferenceID},true);if(_433==null){this.result_getNames=null;return null;}var _434=eval(_433);if(_434){_434=_434.names;}this.result_getNames=_434;return this.result_getNames;},getValue:function(key,def){var expr="/atom:feed/atom:entry[atom:id='"+this.requestedPreferenceID+"']/atom:content/*/model:portletpreferences[@name='"+key+"']/base:value";var _438=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);var _439;if(_438&&_438.length>0){_439=_438[0].getAttribute("value");}else{_439=def;}return _439;},getValues:function(key,def){var expr="/atom:feed/atom:entry[atom:id='"+this.requestedPreferenceID+"']/atom:content/*/model:portletpreferences[@name='"+key+"']/base:value";var _43d=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);var _43e;if(_43d&&_43d.length>0){_43e=new Array();for(var i=0;i<_43d.length;i++){_43e[i]=_43d[i].getAttribute("value");}}else{_43e=def;}return _43e;},isReadOnly:function(key){var id=this.requestedPreferenceID;var expr="/atom:feed/atom:entry[atom:id='"+id+"']/atom:content/*/model:portletpreferences[@name='"+key+"']";var _443=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);var _444=false;if(_443&&_443.length>0){var temp=_443[0].getAttribute("read-only");if(temp!=null){if(temp=="true"){_444=true;}}}return _444;},reset:function(key){this.internal_reset();var expr="/atom:feed/atom:entry[atom:id='"+this.requestedPreferenceID+"']/atom:content/*/model:portletpreferences[@name='"+key+"']";var _448=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);if(_448&&_448.length>0){var _449=_448[0].parentNode;_449.removeChild(_448[0]);}},setValue:function(key,_44b){var _44c=new Array();_44c[0]=_44b;this.setValues(key,_44c);},setValues:function(key,_44e){this.internal_reset();var expr="/atom:feed/atom:entry[atom:id='"+this.requestedPreferenceID+"']/atom:content/*/model:portletpreferences[@name='"+key+"']";var _450=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);var _451=null;if(_450&&_450.length>0){_451=_450[0];for(var i=_451.childNodes.length-1;i>=0;i--){_451.removeChild(_451.childNodes[i]);}}else{var _453="/atom:feed/atom:entry[atom:id='"+this.requestedPreferenceID+"']/atom:content/*";var _454=ibm.portal.xml.xpath.evaluateXPath(_453,this.xmlData,this.ns);if(dojo.isIE){_451=this.xmlData.createNode(1,"model:portletpreferences",this.ns.model);}else{_451=this.xmlData.createElementNS(this.ns.model,"model:portletpreferences");}_451.setAttribute("name",key);_451.setAttribute("read-only","false");_454[0].appendChild(_451);}for(var i=0;i<_44e.length;i++){var _455;if(dojo.isIE){_455=this.xmlData.createNode(1,"base:value",this.ns.base);var _456=this.xmlData.createNode(2,"xsi:type",this.ns.xsi);_456.nodeValue="String";_455.setAttributeNode(_456);}else{_455=this.xmlData.createElementNS(this.ns.base,"base:value");_455.setAttributeNS(this.ns.xsi,"xsi:type","String");}_455.setAttribute("value",_44e[i]);_451.appendChild(_455);}},internal_reset:function(){this.result_getMap=null;this.result_getNames=null;},clone:function(){var _457=dojox.data.dom.innerXML(this.xmlData);var _458=dojox.data.dom.createDocument(_457);return new ibm.portal.portlet.PortletPreferences(this.windowID,this.requestedPreferenceID,_458);}});dojo.declare("ibm.portal.portlet.PortletMode",null,{VIEW:"view",EDIT:"edit",EDIT_DEFAULTS:"edit_defaults",HELP:"help",CONFIG:"config"});dojo.declare("ibm.portal.portlet.WindowState",null,{NORMAL:"normal",MINIMIZED:"minimized",MAXIMIZED:"maximized"});dojo.declare("ibm.portal.portlet.PortletState",null,{constructor:function(_459,_45a){var _45b=new com.ibm.portal.state.StateManager(ibmPortalConfig["contentHandlerURI"]);if(dojo.isString(_459)){var _45c=this._getExistingState(_459,_45b.getSerializationManager());_45b.reset(_45c);}else{_45b.reset(_459);_459=_45a;}this.portletAccessor=_45b.newPortletAccessor(_459);this.renderParameters=this.portletAccessor.getRenderParameters();},_isCSA:function(){var _45d=false;try{_45d=(typeof (document.isCSA)!="undefined");}catch(e){}return _45d;},_getExistingState:function(_45e,_45f){var _460=null;if(this._isCSA()){_460=com.ibm.portal.navigation.controller.NAVIGATION_CONTROLLER.getState().stateDOM;}else{if(_45f!=null){var _461=_45f.deserialize(location.href);_460=_461.returnObject;}else{_460=dojox.data.dom.createDocument();}}return _460;},getPortletMode:function(){return this.portletAccessor.getPortletMode();},setPortletMode:function(_462){this.portletAccessor.setPortletMode(_462);return _462;},getWindowState:function(){return this.portletAccessor.getWindowState();},setWindowState:function(_463){this.portletAccessor.setWindowState(_463);return _463;},getParameterNames:function(){return this.renderParameters.getNames();},getParameterValue:function(name){return this.renderParameters.getValue(name);},getParameterValues:function(name){return this.renderParameters.getValues(name);},getParameterMap:function(){return this.renderParameters.getMap();},setParameterValue:function(name,_467){this.renderParameters.setValue(name,_467);return _467;},setParameterValues:function(name,_469){this.renderParameters.setValues(name,_469);return _469;},setParameterMap:function(map,_46b){if(_46b==true){this.renderParameters.clear();}this.renderParameters.putAll(map);return this.renderParameters.getMap();},removeParameter:function(name){this.renderParameters.remove(name);}});dojo.require("com.ibm.portal.services.PortletFragmentService");dojo.declare("ibm.portal.portlet.XMLPortletRequest",null,{onreadystatechange:null,readyState:0,responseText:null,responseXML:null,status:null,statusText:null,onportletstateready:null,_location:null,_async:null,constructor:function(_46d){var _46e=this.declaredClass+".constructor";ibm.portal.debug.entry(_46e,[_46d]);this.pageID=_46d.pageID;this.windowID=_46d.windowID;this.windowObj=_46d;ibm.portal.debug.exit(_46e);},_getXHR:function(){var _46f=this.declaredClass+"._getXHR";ibm.portal.debug.entry(_46f);if(!this._xhr){this._xhr=this._createXHR();}ibm.portal.debug.exit(_46f,this._xhr);return this._xhr;},_createXHR:function(){var _470=this.declaredClass+"._createXHR";ibm.portal.debug.entry(_470);var _471=null;if(typeof (XMLHttpRequest)!="undefined"){_471=new XMLHttpRequest();}else{_471=new ActiveXObject("Microsoft.XMLHTTP");}ibm.portal.debug.exit(_470,_471);return _471;},_onreadystatechangehandler:function(){var _472=this.declaredClass+"._onreadystatechangehandler";ibm.portal.debug.entry(_472);if(!this.handled){var xhr=this._getXHR();this.readyState=xhr.readyState;ibm.portal.debug.text("ready state is "+xhr.readyState);if(this.readyState==4){var _474=this.windowObj.isAuthenticationRequired(xhr,"xml");ibm.portal.debug.text("is auth required: "+_474);if(_474){this.windowObj.doAuthentication(xhr);return;}else{this.responseText=xhr.responseText;this.responseXML=xhr.responseXML;this.status=xhr.status;this.statusText=xhr.statusText;var _475=new com.ibm.portal.services.PortletFragmentService();var _476=_475.createPortletInfo(xhr.responseXML);this.responseText=_476.markup;this.responseXML=null;var _477=true;var _478=_476.updatedState;if(this.onportletstateready!=null){var _479=_476.windowId;var _47a=new ibm.portal.portlet.PortletState(_478,_479);_477=this.onportletstateready(_47a);}if(_477&&this._isCSA()){_475._fireGlobalPortletStateChange(_476);}this._handleDependentPortlets(_475.readDependentPortlets(xhr.responseXML),_478);this.handled=true;}}if(this.onreadystatechange!=null){this.onreadystatechange();}}ibm.portal.debug.exit(_472);},_handleDependentPortlets:function(_47b,_47c){var _47d=this.declaredClass+"._handleDependentPortlets";ibm.portal.debug.entry(_47d,[_47b,_47c]);if(!this._isCSA()){if(_47b.length>0){window.location.href=this._newPageURL(_47c);}}ibm.portal.debug.exit(_47d);},_isCSA:function(){var _47e=this.declaredClass+"._isCSA";ibm.portal.debug.entry(_47e);var _47f=false;try{_47f=(typeof (document.isCSA)!="undefined");}catch(e){}ibm.portal.debug.exit(_47e,_47f);return _47f;},_flag:function(_480){var _481=this.declaredClass+"._flag";ibm.portal.debug.entry(_481,[_480]);var id="lm:oid:"+this.windowID+"@oid:"+this.pageID;var _483=new com.ibm.portal.services.PortletFragmentService();var url=_483._flagPortletUrl(_480,id);ibm.portal.debug.exit(_481,url);return url;},_newPageURL:function(_485){var _486=this.declaredClass+"._newPageURL";ibm.portal.debug.entry(_486,[_485]);ibm.portal.debug.text(dojox.data.dom.innerXML(_485));var _487=new com.ibm.portal.state.StateManager(ibmPortalConfig["contentHandlerURI"]);var _488=_485;if(!_485){_488=dojox.data.dom.createDocument();}_487.reset(_488);var _489=_487.getSerializationManager();var _48a=_489.serialize(_488);var _48b=_48a["returnObject"];var url=_48b;ibm.portal.debug.exit(_486,url);return url;},open:function(_48d,uri){var _48f=this.declaredClass+".open";ibm.portal.debug.entry(_48f,[_48d,uri]);this.open(_48d,uri,false);ibm.portal.debug.exit(_48f);},open:function(_490,uri,_492){var _493=this.declaredClass+".open";ibm.portal.debug.entry(_493,[_490,uri,_492]);var xhr=this._getXHR();var me=this;this._location=uri;if(_492==undefined){_492=false;}this._async=_492;xhr.onreadystatechange=function(){me._onreadystatechangehandler();};xhr.open(_490,this._flag(uri),_492);xhr.setRequestHeader("X-IBM-XHR","true");ibm.portal.debug.exit(_493);},setRequestHeader:function(_496,_497){var _498=this.declaredClass+".setRequestHeader";ibm.portal.debug.entry(_498,[_496,_497]);this._getXHR().setRequestHeader(_496,_497);ibm.portal.debug.exit(_498);},send:function(data){var _49a=this.declaredClass+".send";ibm.portal.debug.entry(_49a,[data]);this._getXHR().send(data);if(!this._async){this._onreadystatechangehandler();}ibm.portal.debug.exit(_49a);},abort:function(){var _49b=this.declaredClass+".abort";ibm.portal.debug.entry(_49b);this._getXHR().abort();ibm.portal.debug.exit(_49b);},getAllResponseHeaders:function(){return this._getXHR().getAllResponseHeaders();},getResponseHeader:function(_49c){return this._getXHR().getResponseHeader(_49c);}});dojo.declare("ibm.portal.portlet.UserProfile",null,{constructor:function(_49d,data){this.windowID=_49d;this.xmlData=data;this.ns={"xsl":"http://www.w3.org/1999/XSL/Transform","atom":"http://www.w3.org/2005/Atom","xhtml":"http://www.w3.org/1999/xhtml","xsi":"http://www.w3.org/2001/XMLSchema-instance","um":"http://www.ibm.com/xmlns/prod/websphere/um.xsd"};},getAttribute:function(name){var expr="/atom:entry/atom:content/um:profile[@type='user']/um:attribute[@name='"+name+"']/um:attributeValue";var _4a1=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);var _4a2=null;if(_4a1&&_4a1.length>0){if(_4a1[0].textContent){_4a2=_4a1[0].textContent;}else{_4a2=_4a1[0].text;}}return _4a2;},setAttribute:function(name,_4a4){var expr="/atom:entry/atom:content/um:profile[@type='user']/um:attribute[@name='"+name+"']/um:attributeValue";var _4a6=ibm.portal.xml.xpath.evaluateXPath(expr,this.xmlData,this.ns);var _4a7=null;if(_4a6&&_4a6.length>0){if(_4a6[0].textContent){_4a7=_4a6[0].textContent;_4a6[0].textContent=_4a4;}else{_4a7=_4a6[0].text;_4a6[0].text=_4a4;}}else{var _4a8="/atom:entry/atom:content/um:profile[@type='user']/um:attribute[@name='"+name+"']";var _4a9=ibm.portal.xml.xpath.evaluateXPath(_4a8,this.xmlData,this.ns);var _4aa=null;if(_4a9&&_4a9.length>0){_4aa=_4a9[0];}else{var _4ab="/atom:entry/atom:content/um:profile[@type='user']";var _4ac=ibm.portal.xml.xpath.evaluateXPath(_4ab,this.xmlData,this.ns);if(dojo.isIE){_4aa=this.xmlData.createNode(1,"um:attribute",this.ns.um);}else{_4aa=this.xmlData.createElementNS(this.ns.um,"um:attribute");}_4aa.setAttribute("type","xs:string");_4aa.setAttribute("multiValued","false");_4aa.setAttribute("name",name);_4ac[0].appendChild(_4aa);}var _4ad;if(dojo.isIE){_4ad=this.xmlData.createNode(1,"um:attributeValue",this.ns.um);_4ad.text=_4a4;}else{_4ad=this.xmlData.createElementNS(this.ns.um,"um:attributeValue");_4ad.textContent=_4a4;}_4aa.appendChild(_4ad);}return _4a7;},clone:function(){var _4ae=dojox.data.dom.innerXML(this.xmlData);var _4af=dojox.data.dom.createDocument(_4ae);return new ibm.portal.portlet.UserProfile(this.windowID,_4af);}});dojo.declare("ibm.portal.portlet.Error",null,{INFO:0,WARN:1,ERROR:2,constructor:function(_4b0,_4b1,_4b2){this.errorCode=_4b0;this.message=_4b1;this.description=_4b2;},getErrorCode:function(){return this.errorCode;},getMessage:function(){return this.message;},getDescription:function(){return this.description;}});var com_ibm_portal_portlet_portletwindow=new ibm.portal.portlet.PortletWindow();ibm.portal.portlet.PortletWindow.STATUS_UNDEFINED=com_ibm_portal_portlet_portletwindow.STATUS_UNDEFINED;ibm.portal.portlet.PortletWindow.STATUS_OK=com_ibm_portal_portlet_portletwindow.STATUS_OK;ibm.portal.portlet.PortletWindow.STATUS_ERROR=com_ibm_portal_portlet_portletwindow.STATUS_ERROR;com_ibm_portal_portlet_portletwindow=null;var com_ibm_portal_portlet_portletmode=new ibm.portal.portlet.PortletMode();ibm.portal.portlet.PortletMode.VIEW=com_ibm_portal_portlet_portletmode.VIEW;ibm.portal.portlet.PortletMode.EDIT=com_ibm_portal_portlet_portletmode.EDIT;ibm.portal.portlet.PortletMode.EDIT_DEFAULTS=com_ibm_portal_portlet_portletmode.EDIT_DEFAULTS;ibm.portal.portlet.PortletMode.HELP=com_ibm_portal_portlet_portletmode.HELP;ibm.portal.portlet.PortletMode.CONFIG=com_ibm_portal_portlet_portletmode.CONFIG;com_ibm_portal_portlet_portletmode=null;var com_ibm_portal_portlet_windowstate=new ibm.portal.portlet.WindowState();ibm.portal.portlet.WindowState.NORMAL=com_ibm_portal_portlet_windowstate.NORMAL;ibm.portal.portlet.WindowState.MINIMIZED=com_ibm_portal_portlet_windowstate.MINIMIZED;ibm.portal.portlet.WindowState.MAXIMIZED=com_ibm_portal_portlet_windowstate.MAXIMIZED;com_ibm_portal_portlet_windowstate=null;var com_ibm_portal_portlet_error=new ibm.portal.portlet.Error();ibm.portal.portlet.Error.INFO=com_ibm_portal_portlet_error.INFO;ibm.portal.portlet.Error.WARN=com_ibm_portal_portlet_error.WARN;ibm.portal.portlet.Error.ERROR=com_ibm_portal_portlet_error.ERROR;com_ibm_portal_portlet_error=null;}




  
  	
  
        
 
 
delete djConfig.baseUrl;
