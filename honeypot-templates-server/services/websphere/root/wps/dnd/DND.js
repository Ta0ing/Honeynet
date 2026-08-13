/**********************************************************************
 * IBM Confidential
 * (c) Copyright IBM Corp 2005
 *
 * The source code for this program is not published or otherwise
 * divested of its trade secrets, irrespective of what has been
 * deposited with the U.S. Copyright Office.
 *************************************************************************/
//////////////////////////////////////////////////////////////////
// DND default values

//This is the zindex used for dragging markup on the page.
//If a customer has a problem with objects disappearing during a drag
//there is a problem an html with a zindex greater than this default on 
//their page.  They can fix the problem by changing this default value to
//be greater than the zindex of the problem element.
var DNDZindex = 300;

//This is the time in milliseconds it will take for a dropped drag object to slide back
//into place when it is not dropped on a drop zone.  The customer
//can change this value to affect how fast the object will slide back
//into place.
var DNDMoveBackTime = 500;

//This is the default proximity used for drop zones.  Customers can
//change this value to affect the proximity used for drop zones by
//the DnD broker
var DNDProximity = 15;

//This is the default time that we consider to be a "click" in DnD.  Because of the 
//amount of variability in the speed of client machines, customers may wish to 
//modify this value.  This value is only important in multiselect situations (e.g.
//the Content Palette).
var DNDClickTimeout = 1000;

// end DND default values
//////////////////////////////////////////////////////////////////
 
//////////////////////////////////////////////////////////////////
// global functions
function DNDMouseUp(e) {
    if (document._DNDBroker == null) {
		alert("illegal state");
	}
	
    document.omu = document._DNDBroker.originalMouseUp;

    document._DNDBroker.drop();

    if (document.omu != null) {
        document.omu(e);
        document.omu = null;
    }
}

function DNDMouseMove(e) {
	if (document._DNDBroker == null) {
		alert("illegal state");
	}
	var topFrame;
    var selfFrame; 
    
    if(isOpera()){
        topFrame = top;      
        selfFrame = self;
    }else{
        topFrame = window.frames["top"];
        selfFrame = window.frames["self"];
    }
    //this code is here to prevent the browser from thinking that the user
    //is trying to select markup on the page (e.g. highlighting a paragraph to
    //copy & paste).
    if (!isMozilla() && !isIE8()) {
		if (document.selection && document.selection.empty) {
            document.selection.empty();
		}
	}

    var mx = isIE() ? event.clientX : e.clientX;
	var my = isIE() ? event.clientY : e.clientY;

    document._DNDBroker.move(mx, my);
    
    if (document._DNDBroker.originalMouseMove != null) {
        document._DNDBroker.originalMouseMove(e);
    }
}

function isMozilla() {
	return document.getElementById && !document.all;
}

function isIE() {
	return document.getElementById && document.all;
}

function isIE8(){
	 return (navigator.userAgent.indexOf("MSIE 8") >= 0);
}

function isOpera() {
    return (navigator.userAgent.toLowerCase().indexOf("opera") >= 0);
}

function isCtrlDepressed(){
    return document._CtrlKeyDepressed;
}

function keydown(e){
    if (!e) {
        e = window.event;
    }
    if (e.ctrlKey && !isCtrlDepressed()) {

        document._CtrlKeyDepressed = true;
        
        //mozilla doesn't pickup the onKeyUp & onKeyDown events in child frames, so we do this
        for (var i = 0; i < window.frames.length; i++) {
            try {
                window.frames[i].document._CtrlKeyDepressed = true;
            } catch(err) {
              // do something with the error
            }
        }

        setTimeout("document._CtrlKeyDepressed = false;", 500);

    }

    if (document._DNDBroker.oldOnKeyDown != null) {
        document._DNDBroker.oldOnKeyDown(e);
    }
}

function keyup(e){
    if (!e) {
        e = window.event;
    }
    
    //firefox will not recognize ctrlKey on keyUp event so keyCode == 17 used
    if (e.keyCode == 17 || e.ctrlKey) {
        document._CtrlKeyDepressed = false;

        //mozilla doesn't pickup the onKeyUp & onKeyDown events in child frames, so we do this
        for (var i = 0; i < window.frames.length; i++) {
        	// PK71379
        	try {
            	window.frames[i].document._CtrlKeyDepressed = false;
            } catch (err) {
              // do something with the error
            }
        }

    }

    if (document._DNDBroker.oldOnKeyUp != null) {
        document._DNDBroker.oldOnKeyUp(e);
    }
}

function setDNDMoveCursor(){

    if (document.body.style.cursor != "move") {
        document.body.style.cursor="move";
    }

}

function setDNDNoCursor(){

    if (document.body.style.cursor != "not-allowed") {
        document.body.style.cursor="not-allowed";
    }
    
}

function setDNDNormalCursor(){

    if (document.body.style.cursor != "") {
        document.body.style.cursor="";
    }
    
}

//added per LI 4938
function setDNDDragHandleEventHandlers(idOfDragHandle, idOfDragSource){

    var elem = document.getElementById(idOfDragHandle);
    elem.onmousedown = function (e) {   var event = e;
                                        if (event == null) {
                                            event = window.event;
                                        }
                                        document._DNDBroker.beginDrag(idOfDragSource, event.clientX, event.clientY); 
                                        return false;
                                    }
                                   
        elem.onmouseover = setDNDMoveCursor;
        elem.onmouseout = setDNDNormalCursor;
        elem.ondragstart = function() { return false; }

}

//added per LI 4938
function removeDNDDragHandleEventHandlers(idOfDragHandle){

    var elem = document.getElementById(idOfDragHandle);
    elem.onmousedown = null;
    elem.onmouseover = null;
    elem.onmouseout = null;
    elem.ondragstart = null;

}

// end global functions
//////////////////////////////////////////////////////////////////

// logic behind DNDCopyToParent follows...
// In order to eliminate the need to rewrite several functions to specifically
// handle the case where we are dragging between frames, the goal is to convince
// the parent frame to handle dragging the child frame's markup as though the markup
// came from it's own DOM.  To do this, we send the parent frames beginDrag function
// the same X,Y coordinates that are received by the frame where the drag actually begins.
// We changed to use absolute positioning with defect 155005 which simplifies this
// greatly.

function DNDCopyToParent(dndSrc){
    var topFrame;
    var selfFrame; 
    
    if(isOpera()){
        topFrame = top;      
        selfFrame = self;
    }else{
        topFrame = window.frames["top"];
        selfFrame = window.frames["self"];
    }

    var parentDoc = topFrame.document;
    if(parent.federation_iframe) {parentDoc = parent.federation_iframe.document;}
    var srcDiv = document.getElementById(dndSrc.getId());
    
    var newDiv = parentDoc.createElement("div");
    parentDoc.body.appendChild(newDiv);

    //We have to add "_frame" to the id so that any IDs from the child frame don't collide
    //with any ids in the parent frame.  This is a problem because the child frame and parent
    //frame aren't using the same session partition, so we can't guarantee the IDs will be
    //unique across frames.
    dndSrc.id = dndSrc.getId() + "_frame";
    //register this source in the parent's broker
    parentDoc._DNDBroker.registerSource(dndSrc);

    newDiv.setAttribute("id", dndSrc.getId());
    newDiv.className = dndSrc.getCssClassName();

    newDiv.fromChildFrame = "true";
    newDiv.innerHTML = srcDiv.innerHTML;

    newDiv.mouseOffsetX = srcDiv.mouseOffsetX;
    newDiv.mouseOffsetY = srcDiv.mouseOffsetY;
    
    //we "hide" the newDiv so the markup doesn't show up in the parent frame
    newDiv.style.visibility = "hidden";

    parentDoc._DNDBroker.childFrame = selfFrame;

}

//////////////////////////////////////////////////////////////////
// DNDBroker object definition
DNDBroker.prototype				= new Object();
DNDBroker.prototype.constructor = DNDBroker;
DNDBroker.superclass			= null;

function DNDBroker() {
    
    this.DNDSourceArray = new Array(0);
    this.DNDTargetArray = new Array(0);
    this.selectedSource = null;
    this.awareTargetsArray = null;
    this.multiselectSupport = new DNDMultiselectSupport();
    this.dndUtil = new DNDUtil();
    this.dynamicFunctions = new Array();
	this.sourceTgtArray = new Array(0);
    
}

DNDBroker.prototype.registerSource = function(dndSrc){

    //add a new DNDSource to the array of DNDSources
    this.DNDSourceArray.push(dndSrc);

}

//per LI 4938, we are adding and supporting this function for IBM internal use
DNDBroker.prototype.deregisterSource = function(dndSrcId){

    var tempArray = new Array();
    for (var i = 0; i < this.DNDSourceArray.length; i++) {
        if (this.DNDSourceArray[i].getId() != dndSrcId) {
            tempArray.push(this.DNDSourceArray[i]);
        }
    }

    this.DNDSourceArray = tempArray;

}

DNDBroker.prototype.registerTarget = function(dndTgt){

    //add a new DNDTarget to the array of DNDTargets
    this.DNDTargetArray.push(dndTgt);

}

//per LI 4938, we are adding and supporting this function for IBM internal use
DNDBroker.prototype.deregisterTarget = function(dndTgtId){

    var tempArray = new Array();
    for (var i = 0; i < this.DNDTargetArray.length; i++) {
        if (this.DNDTargetArray[i].getId() != dndTgtId) {
            tempArray.push(this.DNDTargetArray[i]);
        }
    }

    this.DNDTargetArray = tempArray;

}

DNDBroker.prototype.beginDrag = function(dndSrc_id, x, y, tgtArray){

    var topFrame;
    var selfFrame; 
    
    if(isOpera()){
        topFrame = top;      
        selfFrame = self;
    }else{
        topFrame = window.frames["top"];
        selfFrame = window.frames["self"];
    }

	this.sourceTgtArray = tgtArray;

    //to handle a bug in IE
    if (this.selectedSource != null) {
        return false;
    }

    //this needs to be done to handle the different ways to retrieve scroll offsets depending on whether
    //the doctype is "transitional" or "strict"
    var docBody = (document.compatMode && document.compatMode != "BackCompat")? document.documentElement : document.body;
    this.currentPageX =  docBody.scrollLeft + x;
    this.currentPageY =  docBody.scrollTop + y;
    this.originalPageX = this.currentPageX;
    this.originalPageY = this.currentPageY;
        
    //try to locate the selected DNDSource
    for (var i = 0; i < this.DNDSourceArray.length; i++) {
        if (this.DNDSourceArray[i].getId() == dndSrc_id) {
            this.selectedSource = this.DNDSourceArray[i];
            break;
        }
    }

    //make sure we have located the DNDSource
    if (this.selectedSource == null) {
        return false;
    }

    if (!document.getElementById(dndSrc_id).fromChildFrame) {
        this.doBeginDragMultiselectSupport();
    }

    this.doBeginDragTargetSelection();

    var selectedSourceIdArray = new Array(0);
    selectedSourceIdArray[0] = this.selectedSource.getId();
    if (this.multiselectSupport.hasSelections()) {
        selectedSourceIdArray = this.multiselectSupport.getSelections();
    }

    var util = new DNDUtil();
    //loop through and calculate mouseOffsetX and mouseOffsetY for each selection
    for (var i = 0; i < selectedSourceIdArray.length; i++) {
        var elem = document.getElementById(selectedSourceIdArray[i]);
        if (!elem.fromChildFrame) {
            elem.mouseOffsetX = this.originalPageX - util.findOffsetX(elem);
            elem.mouseOffsetY = this.originalPageY - util.findOffsetY(elem);
        }
    }

    //frame support
    
    if (topFrame != selfFrame && selfFrame.name == "wpsFLY_flyoutIFrame") {

       var parentDoc=topFrame.document; 
       if(parent.federation_iframe) {parentDoc = parent.federation_iframe.document;}


        for (var i = 0; i < selectedSourceIdArray.length; i++) {

            var currentSource = null;
            
            for (var j = 0; j < this.DNDSourceArray.length; j++) {
                if (this.DNDSourceArray[j].getId() == selectedSourceIdArray[i]) {
                    currentSource = this.DNDSourceArray[j];
                    break;
                }
            }

            var clonedDNDSource = new DNDSource(currentSource.getId(), 
                                                currentSource.getDNDProperties(),
                                                currentSource.getCssClassName(),
                                                currentSource.getDragClone(),
                                                currentSource.getMultiselectable());

            //copy the selected source to the parent frame
            DNDCopyToParent(clonedDNDSource);

        }

        //set the correct state for the multiselect support in the parent
        if (this.multiselectSupport.hasSelections()) {
            //We need to create a new array of selected source ids and append the appropriate "_frame" suffix to them so
            //we can avoid ID collisions across frames
            var newSSIdArray = new Array();
            for (var i = 0; i < selectedSourceIdArray.length; i++) {
                newSSIdArray[i] = selectedSourceIdArray[i] + "_frame";
            }

            parentDoc._DNDBroker.multiselectSupport.selected_src_ids = newSSIdArray;
            parentDoc._DNDBroker.multiselectSupport.selected_src_property_intersection = this.multiselectSupport.getProperties();
        }

        //call the same method in the parent frame with special Id being used for the drag object that was copied to the parent
        parentDoc._DNDBroker.beginDrag(dndSrc_id + "_frame", x, y);

    }

    this.dragCloneSupport = new DNDDragCloneSupport();
    this.dragCloneSupport.beginDrag(selectedSourceIdArray, this.selectedSource.getDragClone());

    this.doBeginDragBSPCreation();
    
    if (this.selectedSource.getCssClassName() != null) {
        document.getElementById(this.selectedSource.getId()).className = this.selectedSource.getCssClassName();
    }

  	if (this.selectedSource.hasOptionalOnDragJavascript()){
  		
  		this.selectedSource.optionalOnDragJavascript();
  	
  	}

    //remember and reassign mouse event handlers
    this.originalMouseMove = document.onmousemove;
    this.originalMouseUp = document.onmouseup;
    document.onmousemove = DNDMouseMove;
    document.onmouseup = DNDMouseUp;
    
    setDNDNoCursor();

}

DNDBroker.prototype.doBeginDragBSPCreation = function(){

    //change the styles for the aware targets and create the array of segments to be
    //used in the BSP algorithm
    var seg_factory = new DNDSegmentFactory();
    var proximity = DNDProximity;
    this.DNDSegArray = new Array(0);
    this.DNDSegArray = seg_factory.createSegments(this.awareTargetsArray, proximity);
    
    //populate the BSP information
    this.DNDBsp = new DNDBsp();
    this.DNDBsp.create(this.DNDSegArray);

}

DNDBroker.prototype.doBeginDragMultiselectSupport = function(){

    this.isClick = true;
    setTimeout("document._DNDBroker.isClick = false", DNDClickTimeout);
    
    if(this.selectedSource.getMultiselectable() && isCtrlDepressed()){
        if (!this.multiselectSupport.alreadySelected(this.selectedSource.getId())) {
            this.multiselectSupport.selectSource(this.selectedSource.getId(), this.selectedSource.getDNDProperties());
        }
    }
    else{
        this.multiselectSupport.clearAllSelections();
    }

}

DNDBroker.prototype.doBeginDragTargetSelection = function(){

    this.awareTargetsArray = new Array(0);
    var current_source_props = this.selectedSource.getDNDProperties();
    //we need to change our source properties if multiselect is happening
    if (this.multiselectSupport.hasSelections()) {
        current_source_props = this.multiselectSupport.getProperties();
    }

    //loop through the targets, matching their action's properties to the selected source
    //properties... doing this, we construct the list of aware targets
    for (var i = 0; i < this.DNDTargetArray.length; i++) {

        var current_target = this.DNDTargetArray[i];
        var current_target_actions = current_target.getActions();

        var target_matches = false;
		
        
		
        for (var j = 0; j < current_target_actions.length; j++) {

            var current_action_matches = true;
            var current_action_properties = current_target_actions[j].getDNDProperties();

            //if doing multiselect and this action doesn't support multiselect, we know this
            //won't work
            if (this.multiselectSupport.hasSelections()) {
                if (!current_target_actions[j].getSupportsMultiselect()) {
                    current_action_matches = false;
                }
            }

            for (var k = 0; k < current_action_properties.length; k++) {
				
				var current_prop_matches = false;
				for (var l = 0; l < current_source_props.length; l++){
					
					if (current_source_props[l].matchesProperty(current_action_properties[k])) {
                    	current_prop_matches=true;
                    	break;
	                }
				}
				
				if (!current_prop_matches) {
                	current_action_matches = false;
	                break;
    	        }
            }
            
            if (current_action_matches){
            	target_matches = true;
            	current_target.active_action = current_target_actions[j];
            	break;
            }
        }

        if (target_matches) {
			var elem = document.getElementById(current_target.getId());
			if(elem != null){
				this.awareTargetsArray.push(current_target);
				//we do this to compute and store the midpoints of the targets - 
				//needed for Bsp ADR work			
				current_target.midpointX = this.dndUtil.findOffsetX(elem) + (elem.offsetWidth / 2);
				current_target.midpointY = this.dndUtil.findOffsetY(elem) + (elem.offsetHeight / 2);
			}
        }

    }

    for (var i = 0; i < this.awareTargetsArray.length; i++) {
		
		if (null != this.awareTargetsArray[i].active_action.getAwareCssClassName()){
			if("tgt_"+this.selectedSource.getId() != this.awareTargetsArray[i].getId()){
				var current_target_element = document.getElementById(this.awareTargetsArray[i].getId());
				current_target_element.className = this.awareTargetsArray[i].active_action.getAwareCssClassName();
			}
		}
	}
}

DNDBroker.prototype.endDrag = function(){

    //we have to reset this stuff before doing the slideback or we get in trouble
    if (document.onmousemove != this.originalMouseMove) {
        document.onmousemove = this.originalMouseMove;
        document.onmouseup = this.originalMouseUp; 
    }

    if (!this.slideBack()) {
        return;
    }

    clearInterval(this.slideBackInterval);
    
    this.killDrag();

}

DNDBroker.prototype.drop = function(){
    var topFrame;
    var selfFrame; 
    
    if(isOpera()){
        topFrame = top;      
        selfFrame = self;
    }else{
        topFrame = window.frames["top"];
        selfFrame = window.frames["self"];
    }

    var doSubmit = false;

    if (this.activeTarget != null) {

		doSubmit = true;		
		
        if (this.activeTarget.active_action.hasOptionalActionJavascript()){
		
			doSubmit = this.activeTarget.active_action.optionalActionJavascript();
		
		}
		
		if (doSubmit){
            if (this.activeTarget.active_action.ALTERNATE_TARGET_SUBMIT_FUNC == null) {
                this.doTargetSubmit();
            }
            else{
                eval(this.activeTarget.active_action.ALTERNATE_TARGET_SUBMIT_FUNC);
            }
		}
    
    }

    if (!doSubmit) {
        this.slideBackInterval = setInterval("document._DNDBroker.endDrag();", 10);
    }
    //if the parent frame is still "dragging", kill it
    if (topFrame != selfFrame &&
        selfFrame.name == "wpsFLY_flyoutIFrame" &&
        topFrame.document._DNDBroker.childFrame != null) {
        topFrame.document._DNDBroker.killDrag();
        topFrame.document._DNDBroker.childFrame = null;
    }
}

//this is used to clean up when dragging with frame support
DNDBroker.prototype.killDrag = function(){

    var topFrame;
    var selfFrame; 
    
    if(isOpera()){
        topFrame = top;      
        selfFrame = self;
    }else{
        topFrame = window.frames["top"];
        selfFrame = window.frames["self"];
    }

    //code to fix the "freed script" problem in IE.  Bascially, we have to remove references to objects in the
    //child frame from the parent frame at the end of a drag
    if (topFrame == selfFrame) {
        //iterate through the registered sources, removing sources from the child frame
        var newArray = new Array();
        for (var i = 0; i < this.DNDSourceArray.length; i++) {
            if (!(this.DNDSourceArray[i].getId().indexOf("_frame") > -1)) {
                newArray.push(this.DNDSourceArray[i]);
            }
        }

        this.DNDSourceArray = newArray;
    }

    //we have to reset this stuff before doing the slideback or we get in trouble
    if (document.onmousemove != this.originalMouseMove) {
        document.onmousemove = this.originalMouseMove;
        document.onmouseup = this.originalMouseUp; 
    }

    if (this.selectedSource == null) {
        //we check this because sometime DnD gets in a bad state since we can't capture events that happen
        //outside the browser and some of the event firing in both browsers is unreliable
        return;
    }

    var selectedSourceIdArray = new Array(0);
    selectedSourceIdArray[0] = this.selectedSource.getId();
    if (this.multiselectSupport.hasSelections()) {
        selectedSourceIdArray = this.multiselectSupport.getSelections();
    }

    this.dragCloneSupport.endDrag(selectedSourceIdArray, this.selectedSource.getDragClone());
    this.dragCloneSupport = null;

    //if this.isClick has not been set to false, then the mousedown and mouseup have occured within
    //the timeout of each other and we will consider this to be a click for multiselect purposes.
    //IMPORTANT - this code must come after the dragclonesupport code
    if (this.isClick && this.selectedSource.getMultiselectable() && !this.multiselectSupport.alreadySelected(this.selectedSource.getId())) {
        this.multiselectSupport.selectSource(this.selectedSource.getId(), this.selectedSource.getDNDProperties());
    }

    //iterate through the active targets, setting their className attributes for their DOM elements
    //back to the empty string
    for (var i = 0; i < this.awareTargetsArray.length; i++) {

        var curr_tgt = document.getElementById(this.awareTargetsArray[i].getId());
		if(curr_tgt != null){
			
			curr_tgt.className = "";
			curr_tgt.style.cssText = "";
			
			if(curr_tgt.parentNode.getAttribute("style")!=null){
				curr_tgt.parentNode.style.cssText = "";
			}
			
		}
		
        this.awareTargetsArray[i].active_action = null;

    }

    //find out if we are ending a drag from a child frame and are in the parent frame
    //if so, we need to reset the multiselectsupport
    if (this.multiselectSupport.hasSelections()){
        var source_ids = this.multiselectSupport.getSelections();
        if ( document.getElementById(source_ids[0]) == null || document.getElementById(source_ids[0]).fromChildFrame ) {
            this.multiselectSupport = new DNDMultiselectSupport();
        }
    }

   	if(this.selectedSource.id.indexOf("_frame") == -1 ){
   		var originalNode = document.getElementById(this.selectedSource.id);
   		originalNode.className = "";
		originalNode.style.opacity = 1;
		originalNode.style.filter = 'alpha(opacity=' + 100 + ')';
		
		// Fix for Firefox with slideback in vertical containers
		if(isMozilla() && 
		   !this.activeTarget && 
		   this.selectedSource.id.indexOf("_frame") < 0 &&
		   typeof CSA_DND_CHECK != 'undefined'){  //PK80533
			var container = CSADND_getPortletsContainer(originalNode);
			if(container != null){
				if(container.getAttribute("CSAContainerType") == "UCV"){
					var oldRow = originalNode.parentNode;
		   			var newNode = oldRow.removeChild(originalNode);
			
			   		var newRow = document.createElement("TR");
			   		newRow.style.verticalAlign = "top";
			   		
			   		var tbody = oldRow.parentNode;
			   		tbody.insertBefore(newRow, oldRow);
			   		tbody.removeChild(oldRow);
			   		
			   		newRow.appendChild(newNode);
				}
			}
   		}
	}
	
    this.selectedSource = null;
    this.activeTarget = null;
    this.awareTargetsArray = null;
    
    //reset additional form parameters
    this.formParamNames = null;
    this.formParamValues = null;

    this._frameOriginalX = null;
    this._frameOriginalY = null;

    //set the 'normal' cursor
    setDNDNormalCursor();

}

DNDBroker.prototype.move = function(x, y){
    
    //this needs to be done to handle the different ways to retrieve scroll offsets depending on whether
    //the doctype is "transitional" or "strict"
    var docBody = (document.compatMode && document.compatMode != "BackCompat")? document.documentElement : document.body;

    //calculate new coordinate offset
    var pageX = docBody.scrollLeft + x;
	var pageY = docBody.scrollTop + y;

    //remember the current pageX and pageY so we can do the slideBack
    this.currentPageX = pageX;
    this.currentPageY = pageY;

    if (isIE() & this.isBidi) {
        pageX = x;
    }

    var selectedSourceIdArray = new Array(0);
    selectedSourceIdArray[0] = this.selectedSource.getId();
    if (this.multiselectSupport.hasSelections()) {
        selectedSourceIdArray = this.multiselectSupport.getSelections();
    }
    
    for (var i = 0; i < selectedSourceIdArray.length; i++) {

        var src_elem = document.getElementById(selectedSourceIdArray[i]);

        //for frame support slideback

        if (src_elem.fromChildFrame && this._frameOriginalX == null && this._frameOriginalY == null) {
            this._frameOriginalX = this.currentPageX - (this.childFrame.document._DNDBroker.currentPageX - this.childFrame.document._DNDBroker.originalPageX);
            this._frameOriginalY = this.currentPageY - (this.childFrame.document._DNDBroker.currentPageY - this.childFrame.document._DNDBroker.originalPageY);

            //call this to reset the child frame
            if (this.childFrame != null) {
                this.childFrame.document._DNDBroker.killDrag();
                this.childFrame = null;
            }
        }

        if (src_elem.style.visibility == "hidden") {
            src_elem.style.visibility = "visible";
        }
            
        if (!document._DNDBroker.isBidi) {
            src_elem.style.left = (this.currentPageX - src_elem.mouseOffsetX) + "px";
        }
        else{
            // BIDI suppport
            if (this.bd == null) {
                this.bd = new DNDBrowserDimensions();
            }

            if (isMozilla() && !src_elem.fromChildFrame) {
                src_elem.style.left = (this.currentPageX - (src_elem.offsetWidth + src_elem.mouseOffsetX)) + "px";
            }
            else if (!src_elem.fromChildFrame){
                /*if (src_elem.xOffsetForBidiInIE == null) {
                    src_elem.xOffsetForBidiInIE = -1 * (this.bd.getHTMLElementWidth() - this.currentPageX);
                }
                src_elem.style.left = (-1 * (this.bd.getHTMLElementWidth() - this.currentPageX) - src_elem.xOffsetForBidiInIE) + "px";
                */
                src_elem.style.left = (this.currentPageX - src_elem.mouseOffsetX) + "px";
            }
            else{
                
                if (isMozilla()) {
                    src_elem.style.left = -1 * (this.bd.getHTMLElementWidth() - (this.currentPageX - src_elem.mouseOffsetX))  + "px";
                }
                else{
                    src_elem.style.left = -1 * (this.bd.getHTMLElementWidth() - this.currentPageX) + "px";
                }
            }
        }
        
        src_elem.style.top = (this.currentPageY - src_elem.mouseOffsetY) + "px";

    }

    //only calculate active target if we have aware targets
    if (this.awareTargetsArray != null && this.awareTargetsArray.length > 0) {
        //remember what target was previously active
        var previousActiveTarget = this.activeTarget;
		var bspPageY = pageY;
		if (navigator.userAgent.toLowerCase().indexOf("safari") >= 0){
			bspPageY = docBody.scrollTop + y + window.pageYOffset;
		}
        this.activeTarget = this.DNDBsp.getTarget(pageX,bspPageY);
        if (this.activeTarget != previousActiveTarget) {
            if (previousActiveTarget != null) {
                if (previousActiveTarget.active_action.getAwareCssClassName() != null) {
					if("tgt_"+this.selectedSource.getId() != previousActiveTarget.getId()){
						
						// display the target with the aware style class
						document.getElementById(previousActiveTarget.getId()).className = previousActiveTarget.active_action.getAwareCssClassName();
					
						if(typeof CSA_DND_CHECK != 'undefined'){
							var tgtElem = dojo.byId(previousActiveTarget.getId());
							tgtElem.removeAttribute("style");
							if(tgtElem.parentNode.getAttribute("style") != null)
								tgtElem.parentNode.removeAttribute("style");
						}
					}
                }
            }

            if (this.activeTarget != null) {
                if (this.activeTarget.active_action.getActiveCssClassName() != null) {
					if("tgt_"+this.selectedSource.getId() != this.activeTarget.getId()){
						if(typeof CSA_DND_CHECK != 'undefined' && this.selectedSource.id.indexOf("wp_") < 0){

							// CSA drag and drop move on page
														
							var selectedElem = document.getElementById(this.selectedSource.getId());
							var tgtElem = document.getElementById(this.activeTarget.getId());
							var tgtContainer = CSADND_getPortletsContainer(document.getElementById(this.activeTarget.getId()));
							var selectedContainer = CSADND_getPortletsContainer(selectedElem);
						
							if(tgtContainer.getAttribute("CSAContainerType") == "UCH"){
								// Horizontal Container
								var numberOfChildNodes = tgtContainer.childNodes[0].childNodes[0].childNodes.length - 1;

								if(tgtContainer.id ==  selectedContainer.id){
									numberOfChildNodes = ( ( (numberOfChildNodes ) /2 ) );
								}else{
									numberOfChildNodes = ( ( (numberOfChildNodes ) /2 ) )+1;
								}
								
								var activeWidth = tgtContainer.offsetWidth / numberOfChildNodes;
								tgtElem.style.cssText = "height: "+ selectedElem.offsetHeight + "px; width: "+ activeWidth +"px;";
								tgtElem.parentNode.style.cssText = "width: "+ activeWidth +"px;";
								tgtElem.className = this.activeTarget.active_action.getActiveCssClassName();

							}else{
								// Vertical Container
								var containerEmpty = false;
								var count = 0;
								
								// check container for portlets
								for(var j = 0; j < tgtContainer.childNodes[0].childNodes.length; j++){
										
									// get the container's control table cell
									var elem = tgtContainer.childNodes[0].childNodes[j].childNodes[0];
										
									if(elem.getAttribute("CSAContainerType") == "C"){
										
										//check to make sure the control is there and not being dragged
										if(selectedElem.id != elem.id){
											count++;
										}
									}
								}
									
								// if no portlet's are in the container then it is empty
								if( count == 0 ){
									containerEmpty = true;
								}else{
									containerEmpty = false;
								}
								
								if(containerEmpty){
									var numberOfChildNodes = 0;
									var containerParent = tgtContainer.parentNode.parentNode;
									for (var i = 0; i < containerParent.childNodes.length; i++){
										numberOfChildNodes++;
									}
									
									var activeWidth = containerParent.offsetWidth / numberOfChildNodes;
									tgtElem.parentNode.style.cssText = "width:"+activeWidth+"px;";
									tgtElem.style.cssText = "height:"+selectedElem.offsetHeight+"px;";	
								}else{
									tgtElem.style.cssText = "height: "+selectedElem.offsetHeight+"px; width:100%; min-width: 10px;";
								}
								
								tgtElem.className = this.activeTarget.active_action.getActiveCssClassName();
							}
						}else if(typeof CSA_DND_CHECK != 'undefined' && this.selectedSource.id.indexOf("wp_") >= 0){
							
							// CSA drag from palette to page
							
							if( (CSADND_getPortletsContainer(dojo.byId(this.activeTarget.getId()))).getAttribute("CSAContainerType") == "UCV"){
								document.getElementById(this.activeTarget.getId()).className = "dndStaticDropActiveVerticalCSA";
							}else{
								document.getElementById(this.activeTarget.getId()).className = "dndStaticDropActiveHorizontalCSA";
							}
							
						}else{
							
							// SSA drag and drop active target
							document.getElementById(this.activeTarget.getId()).className = this.activeTarget.active_action.getActiveCssClassName();
						
						}
					}
                }
            }
        }
    }

    if (this.activeTarget != null) {
        setDNDMoveCursor();
    }
    else{
        setDNDNoCursor();
    }
    
}

DNDBroker.prototype.doTargetSubmit = function(){	
	var topFrame;
    var selfFrame; 
    
    if(isOpera()){
        topFrame = top;      
        selfFrame = self;
    }else{
        topFrame = window.frames["top"];
        selfFrame = window.frames["self"];
    }

	// if the portlet was dropped into same position in layout, do not send off request
	if( this.sourceTgtArray != null){
        var props = this.selectedSource.getDNDProperties();
        var contentType = props[0]["type"];
        //only perform check for page portlets and palette categories
        if( contentType == "portlet_windowID" || contentType == "category_id"){
    		if( (this.activeTarget.getId() == this.sourceTgtArray[0]) || (this.activeTarget.getId() == this.sourceTgtArray[1]) ){
    			this.slideBackInterval = setInterval("document._DNDBroker.endDrag();", 10);
    			//if the parent frame is still "dragging", kill it
    			if (topFrame != window.frames["self"] &&
    				selfFrame.name == "wpsFLY_flyoutIFrame" &&
    				topFrame.document._DNDBroker.childFrame != null) {
    				topFrame.document._DNDBroker.killDrag();
    				topFrame.document._DNDBroker.childFrame = null;
    			}
    			return;
    		}
        }
	}
    
	var src_div = document.getElementById(this.selectedSource.getId());
		
    var formElem = document.createElement("form");
    src_div.appendChild(formElem);

    formElem.setAttribute("action", this.activeTarget.active_action.getAction());
	formElem.setAttribute("method", "POST");
    formElem.setAttribute("id", "DND_SUBMIT_FORM");

    var tgt_props = this.activeTarget.active_action.getDNDProperties();
    var src_props = this.selectedSource.getDNDProperties();
    if (this.multiselectSupport.hasSelections()) {
        src_props = this.multiselectSupport.doDelimiterReplacement(this.multiselectSupport.getProperties(),
                                                                   this.activeTarget.active_action.getMultiselectDelimiter());
    }
    
    for (var i = 0; i < tgt_props.length; i++) {

        var curr_prop = tgt_props[i];

        for (var j = 0; j < src_props.length; j++) {
            if (curr_prop.matchesProperty(src_props[j])) {
                curr_prop.value = src_props[j].getValue();
                break;
            }
        }

        if (curr_prop != null && curr_prop.getName() != null && curr_prop.getValue() != null){

            var inputElem = document.createElement("input");
            inputElem.setAttribute("type", "hidden");
            inputElem.setAttribute("name", curr_prop.getName());
            inputElem.setAttribute("value", curr_prop.getValue());
            formElem.appendChild(inputElem);
        }

    }

    //add additional form parameters
    if (this.formParamNames != null &&
        this.formParamValues != null &&
        this.formParamNames.length == this.formParamValues.length) {

        for (var i = 0; i < this.formParamNames.length; i++) {
            var inputElem = document.createElement("input");
            inputElem.setAttribute("type", "hidden");
            inputElem.setAttribute("name", this.formParamNames[i]);
            inputElem.setAttribute("value", this.formParamValues[i]);
            formElem.appendChild(inputElem);
        }

    }

    formElem.submit();

}

DNDBroker.prototype.slideBack = function(){

    var currentTime = new Date();

    if (!this.slidingBack) {
        
        this.slidingBack = true;
        this.slidingBackBeginTime = currentTime.getTime();
        this.slidingBackEndTime = this.slidingBackBeginTime + DNDMoveBackTime;
        this.slidingBackOriginalX = this.currentPageX;
        this.slidingBackOriginalY = this.currentPageY;
        this.slidingBackDestinationX = this.originalPageX;
        this.slidingBackDestinationY = this.originalPageY;
        if (this._frameOriginalX != null && this._frameOriginalY != null) {
            this.slidingBackDestinationX = this._frameOriginalX;
            this.slidingBackDestinationY = this._frameOriginalY;
        }

    }

    if (currentTime.getTime() > this.slidingBackEndTime || this.isClick) {
        this.slidingBack = false;
        this.slidingBackBeginTime = null;
        this.slidingBackEndTime = null;
        this.slidingBackOriginalX = null;
        this.slidingBackOriginalY = null;
        this.slidingBackDestinationX = null;
        this.slidingBackDestinationY = null;
        return true;
    }

    var currentTimeMillis = currentTime.getTime();

    var fractionOfSlideTimePassed = (this.slidingBackEndTime - currentTimeMillis) / DNDMoveBackTime;
    var dX = Math.round((this.slidingBackOriginalX - this.slidingBackDestinationX) * fractionOfSlideTimePassed);
    var dY = Math.round((this.slidingBackOriginalY - this.slidingBackDestinationY) * fractionOfSlideTimePassed);

    if (this.selectedSource == null) {
        //we check this because sometime DnD gets in a bad state since we can't capture events that happen
        //outside the browser and some of the event firing in both browsers is unreliable
        return true;;
    }

    var selectedSourceIdArray = new Array(0);
    selectedSourceIdArray[0] = this.selectedSource.getId();
    if (this.multiselectSupport.hasSelections()) {
        selectedSourceIdArray = this.multiselectSupport.getSelections();
    }

    for (var j = 0; j < selectedSourceIdArray.length; j++) {

        var elem = document.getElementById(selectedSourceIdArray[j]);
        
        if (!document._DNDBroker.isBidi) {
            elem.style.left = (this.slidingBackDestinationX - elem.mouseOffsetX + dX) + "px";
        }
        else{
            if (this.bd == null) {
                this.bd = new DNDBrowserDimensions();
            }

            if (isMozilla() && !elem.fromChildFrame) {
                elem.style.left = (this.slidingBackDestinationX + elem.mouseOffsetX + dX) + "px";
            }
            else if (!elem.fromChildFrame){
                elem.style.left = (-1 * (this.bd.getHTMLElementWidth() - this.slidingBackDestinationX) - elem.xOffsetForBidiInIE) + dX + "px";
            }
            else{
                if (isMozilla()) {
                    elem.style.left = -1 * (this.bd.getHTMLElementWidth() - (this.slidingBackDestinationX + elem.mouseOffsetX) - dX) + "px";
                }
                else{
                    elem.style.left = -1 * (this.bd.getHTMLElementWidth() - (this.slidingBackDestinationX)) + dX + "px";
                }

            }
        }
        elem.style.top = (this.slidingBackDestinationY - elem.mouseOffsetY + dY) + "px";

    }

    return false;

}

DNDBroker.prototype.addFormParam = function(name, value){
    if (this.formParamNames == null || this.formParamValues == null) {
        this.formParamNames = new Array(0);
        this.formParamValues = new Array(0);
    }
    if (name != null && value != null && this.formParamNames.length == this.formParamValues.length) {
        this.formParamNames[this.formParamNames.length] = name;
        this.formParamValues[this.formParamValues.length] = value;
    }
}

DNDBroker.prototype.parseAndRegisterSource = function(xml){

    var parser = new DNDXMLParser();
    var srcArray = parser.parseSource(xml);
    for (var i = 0; i < srcArray.length; i++) {
        this.registerSource(srcArray[i]);
    }
}

DNDBroker.prototype.parseAndRegisterTarget = function(xml){
    var parser = new DNDXMLParser();
    var tgtArray = parser.parseTarget(xml);
    for (var i = 0; i < tgtArray.length; i++) {
        this.registerTarget(tgtArray[i]);
    }
}

DNDBroker.prototype.addDynFunc = function (newFunc){
    this.dynamicFunctions.push(newFunc);
}

//wrapper functions for shortening the markup size
DNDBroker.prototype.addDFnc = function (newFunc){
    this.addDynFunc(newFunc);
}

DNDBroker.prototype.pARS = function(xml){
    this.parseAndRegisterSource(xml);
}

DNDBroker.prototype.pART = function(xml){
    this.parseAndRegisterTarget(xml);
}

// end DNDBroker object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDDynFunc object definition
DNDDF.prototype = new Object();
DNDDF.prototype.constructor = DNDDF;
DNDDF.superclass = null;

function DNDDF(iName){
    this.name = iName;
}
// end DNDDynFunc object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDXMLParser object definition
DNDXMLParser.prototype = new Object();
DNDXMLParser.prototype.constructor = DNDXMLParser;
DNDXMLParser.superclass = null;

function DNDXMLParser(){
    //do nothing
}

DNDXMLParser.prototype.parseSource = function(xml){

    var srcArray = new Array();
    while (xml.length > 0 && xml.indexOf("<dndSrc") > -1 && xml.indexOf("</dndSrc>") > xml.indexOf("<dndSrc")) {
        var attributeString = xml.substring(xml.indexOf("<dndSrc") + 8, xml.indexOf(">"));
        var propertyString = xml.substring(xml.indexOf(">") + 1, xml.indexOf("</dndSrc>"));
        xml = xml.substring(xml.indexOf("</dndSrc>") + 9, xml.length);

        var propArray = this.parseProperty(propertyString);

        var id = this.getAttribute(attributeString, "id");
        var cssClassName = this.getAttribute(attributeString, "cssCN");
        var dragCloneStr = this.getAttribute(attributeString, "dc");
        var dragClone = false;
        if (dragCloneStr == "true" || dragCloneStr == "TRUE") {
            dragClone = true;
        }
        var multiselectStr = this.getAttribute(attributeString, "multi");
        var multiselect = false;
        if (multiselectStr == "true" || multiselectStr == "TRUE") {
            multiselect = true;
        }

        var dSource = new DNDSource(id, propArray, cssClassName, dragClone, multiselect);
        
        srcArray.push(dSource);
    }

    return srcArray;

}

DNDXMLParser.prototype.parseTarget = function(xml){

    var tgtArray = new Array();
    while (xml.length > 0 && xml.indexOf("<dndTgt") > -1 && xml.indexOf("</dndTgt>") > xml.indexOf("<dndTgt")) {
        var attributeString = xml.substring(xml.indexOf("<dndTgt") + 8, xml.indexOf(">"));
        var actionString = xml.substring(xml.indexOf(">") + 1, xml.indexOf("</dndTgt>"));
        xml = xml.substring(xml.indexOf("</dndTgt>") + 9, xml.length);

        var actArray = this.parseAction(actionString);

        var id = this.getAttribute(attributeString, "id");

        tgtArray.push(new DNDTarget(id, actArray));
    }

    return tgtArray;

}

DNDXMLParser.prototype.parseAction = function(xml){

    var actArray = new Array();
    while(xml.length > 0 && xml.indexOf("<dndAct") > -1 && xml.indexOf("</dndAct>") > xml.indexOf("<dndAct")){

        var attributeString = xml.substring(xml.indexOf("<dndAct") + 8, xml.indexOf(">"));
        var propertyString = xml.substring(xml.indexOf(">") + 1, xml.indexOf("</dndAct>"));
        xml = xml.substring(xml.indexOf("</dndAct>") + 9, xml.length);

        var propArray = this.parseProperty(propertyString);

        var id = this.getAttribute(attributeString, "id");
        var action = this.getAttribute(attributeString, "a");
        var activeCss = this.getAttribute(attributeString, "actCss");
        var awareCss = this.getAttribute(attributeString, "awCss");
        var multiselectStr = this.getAttribute(attributeString, "ms");
        var multiselect = false;
        if (multiselectStr == "true" || multiselectStr== "TRUE") {
            multiselect = true;
        }
        var multiselectDelim = this.getAttribute(attributeString, "msDelim");

        var nextAction = new DNDAction(id, action, propArray, activeCss, awareCss, multiselect, multiselectDelim);

        actArray.push(nextAction);
    }                                                                                                         
    return actArray;
}

DNDXMLParser.prototype.parseProperty = function(xml){

    var propArray = new Array();

    while (xml.length > 0 && xml.indexOf("<dndProp") > -1 && xml.indexOf("/>") > xml.indexOf("<dndProp")) {
        
        var attributeString = xml.substring(xml.indexOf("<dndProp") + 9, xml.indexOf("/>"));

        xml = xml.substring(xml.indexOf("/>") + 2, xml.length);

        var namespace = this.getAttribute(attributeString, "ns");
        var type = this.getAttribute(attributeString, "t");
        var name = this.getAttribute(attributeString, "n");
        var value = this.getAttribute(attributeString, "v");

        propArray.push(new DNDProperty(namespace, type, name, value));

    }

    return propArray;

}

DNDXMLParser.prototype.getAttribute = function(attrs, key){

    var value = null;
    if (attrs.indexOf(key + "='") > -1) {
        value = attrs.substring(attrs.indexOf(key + "='") + key.length + 2, attrs.length);
        value = value.substring(0, value.indexOf("'"));
    }

    return value;

}

// end DNDXMLParser object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDSegment object definition
DNDSegment.prototype				= new Object();
DNDSegment.prototype.constructor = DNDSegment;
DNDSegment.superclass			= null;

function DNDSegment(target, x1, y1, x2, y2, constraint, norm, xAxis, yAxis) {
	this.init(target, x1, y1, x2, y2, constraint, norm, xAxis, yAxis); 
}

DNDSegment.prototype.init = function(target, x1, y1, x2, y2, constraint, norm, xAxis, yAxis) {
    this._target = target;
	this._x1 = x1;
	this._x2 = x2;
	this._y1 = y1;
	this._y2 = y2;
	this._constraint = constraint;
	this._norm = norm;
	this._xAxis = xAxis;
	this._yAxis = yAxis;
}

DNDSegment.prototype.isInFront = function(x, y) {
	var v = this._xAxis ? x : y;
	return ((this._constraint * this._norm) >= (v*this._norm));
}

DNDSegment.prototype.classify = function(seg) {
	var bool1 = this.isInFront(seg._x1, seg._y1);
	var bool2 = this.isInFront(seg._x2, seg._y2);
	if (bool1 && bool2) {
		return 1;
	} else if (!bool1 && !bool2) {
		return -1;
	} else {
		return 0;
	}
}

DNDSegment.prototype.split = function(seg, segFront, segBack) {
	var intersectX = this._xAxis ? this._constraint : seg._constraint;
	var intersectY = this._yAxis ? this._constraint : seg._constraint;
	var frontX = seg.isInFront(this._x1, this._y1) ? this._x1 : this._x2;
	var frontY = seg.isInFront(this._x1, this._y1) ? this._y1 : this._y2;
	var backX = seg.isInFront(this._x1, this._y1) ? this._x2 : this._x1;
	var backY = seg.isInFront(this._x1, this._y1) ? this._y2 : this._y1;
	segFront.init(this._target, frontX, frontY, intersectX, intersectY, this._constraint, this._norm, this._xAxis, this._yAxis);
	segBack.init(this._target, backX, backY, intersectX, intersectY, this._constraint, this._norm, this._xAxis, this._yAxis);
}

DNDSegment.prototype.getDNDTarget = function(){
    return this._target;
}

DNDSegment.prototype.toString = function() {
	return "DNDSegment { target:" + this._target.toString() + ", x1:" + this._x1 + ", y1:" + this._y1 + ", x2:" + this._x2 + ", y2:" + this._y2 + "}";
}

// end DNDSegment object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDSegmentFactory object definition
DNDSegmentFactory.prototype				= new Object();
DNDSegmentFactory.prototype.constructor = DNDSegmentFactory;
DNDSegmentFactory.superclass			= null;

function DNDSegmentFactory() {
	this.dndUtil = new DNDUtil();
}

DNDSegmentFactory.prototype.createSegments = function(dnd_target_array, proximity) {

    var segs = new Array(0);

    for (var i = 0; i < dnd_target_array.length; i++) {
        
        var dndTarget = dnd_target_array[i];	
        var div = document.getElementById(dndTarget.getId());	
        var left = this.dndUtil.findOffsetX(div) - proximity;
        var top = this.dndUtil.findOffsetY(div) - proximity;	
        var right = div.offsetWidth + left + proximity + proximity;
        var bottom = div.offsetHeight + top + proximity + proximity;	

        segs[segs.length] = new DNDSegment(dndTarget, left, top, left, bottom, left, -1, true, false);
        segs[segs.length] = new DNDSegment(dndTarget, left, bottom, right, bottom, bottom, 1, false, true);	
        segs[segs.length] = new DNDSegment(dndTarget, right, bottom, right, top, right, 1, true, false);
        segs[segs.length] = new DNDSegment(dndTarget, right, top, left, top, top, -1, false, true);
        
    }

    return segs;
}
// end DNDSegmentFactory object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDTreeWalker object definition
DNDTreeWalker.prototype         = new Object();
DNDTreeWalker.prototype.constructor = DNDTreeWalker;
DNDTreeWalker.superclass        = null;

function DNDTreeWalker(){

    this.targetList = new Array();
    this.impossibleTargetList = new Array();

}

DNDTreeWalker.prototype.addTarget = function (dndTarget){
    
    var notInList = true;
    for (var i = 0; i < this.targetList.length; i++) {
        if (this.targetList[i].getId() == dndTarget.getId()) {
            notInList = false;
        }
    }
    if (notInList) {
        this.targetList[this.targetList.length] = dndTarget;
    }

}

DNDTreeWalker.prototype.impossibleTarget = function (dndTarget){
    
    var notInList = true;
    for (var i = 0; i < this.impossibleTargetList.length; i++) {
        if (this.impossibleTargetList[i].getId() == dndTarget.getId()) {
            notInList = false;
        }
    }
    if (notInList) {
        this.impossibleTargetList[this.impossibleTargetList.length] = dndTarget;
    }

}

DNDTreeWalker.prototype.getTarget = function (x, y){
    
    var targetArray = new Array();

    //we loop through the targetList, adding all targets that aren't in the impossibleTargetList to 
    //the list of potential targets to return
    for (var i = 0; i < this.targetList.length; i++) {

        var addToList = true;
        for (var j = 0; j < this.impossibleTargetList.length; j++) {
            if (this.targetList[i].getId() == this.impossibleTargetList[j].getId()) {
                addToList = false;
            }
        }

        if (addToList) {
            targetArray[targetArray.length] = this.targetList[i];
        }

    }

    var tgt = null;
    if (targetArray.length == 1) {
        tgt = targetArray[0];
    }
    else if (targetArray.length > 1) {
        
        for (var i = 0; i < targetArray.length; i++) {
            if (tgt == null) {
                tgt = targetArray[i];
            }
            else{

                var currentDistance = Math.sqrt(Math.pow((x - tgt.midpointX),2) + Math.pow((y - tgt.midpointY),2));
                var newDistance = Math.sqrt(Math.pow((x - targetArray[i].midpointX),2) + Math.pow((y - targetArray[i].midpointY), 2));
                if (newDistance <= currentDistance) {
                    tgt = targetArray[i];
                }
            }
        }
    }

    return tgt;

}

// end DNDTreeWalker object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDBsp (Binary Space Partition) object definition
DNDBsp.prototype                = new Object();
DNDBsp.prototype.constructor    = DNDBsp;
DNDBsp.superclass               = null;

function DNDBsp() { 
    //do nothing
}

DNDBsp.prototype.create = function(segArray) {
	this._seg = segArray[0];
	if (segArray.length > 1) {
		var frontArray = new Array();
		var backArray = new Array();
		for (var i=1; i < segArray.length; i++) {
			var retval = this._seg.classify(segArray[i]);
			if (retval == 1) {
				frontArray[frontArray.length] = segArray[i];
			} else if (retval == -1) {
				backArray[backArray.length] = segArray[i];
			} else {
				var segFront = new DNDSegment();
				var segBack = new DNDSegment();
				segArray[i].split(this._seg, segFront, segBack);
				frontArray[frontArray.length] = segFront;
				backArray[backArray.length] = segBack;
			}
		}
		if (frontArray.length > 0) {
			this._frontBsp = new DNDBsp();
			this._frontBsp.create(frontArray);
		}
		if (backArray.length > 0) {
			this._backBsp = new DNDBsp();
			this._backBsp.create(backArray);
		}
	}
}

DNDBsp.prototype.getTarget = function(x, y, treeWalker) {
    if (treeWalker == null) {
        treeWalker = new DNDTreeWalker();
    }
	var retval = this._seg.isInFront(x, y);
	if (retval == 1) {

        treeWalker.addTarget(this._seg._target);

		if (this._frontBsp == null) {
			return treeWalker.getTarget(x, y);
		} else {
			return this._frontBsp.getTarget(x, y, treeWalker);
		}
	} else {

        treeWalker.impossibleTarget(this._seg._target);

		if (this._backBsp == null) {
			return treeWalker.getTarget(x, y);
		} else {
			return this._backBsp.getTarget(x,y, treeWalker);
		}
	}
}
// end DNDBsp object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDUtil object definition
DNDUtil.prototype				= new Object();
DNDUtil.prototype.constructor = DNDUtil;
DNDUtil.superclass			= null;

function DNDUtil() {
	//do nothing
}

DNDUtil.prototype.findOffsetX = function(html_obj) {
	return ( html_obj.offsetParent==null ? html_obj.offsetLeft : html_obj.offsetLeft + this.findOffsetX(html_obj.offsetParent) );
}

DNDUtil.prototype.findOffsetY = function(html_obj) {
	return ( html_obj.offsetParent==null ? html_obj.offsetTop : html_obj.offsetTop + this.findOffsetY(html_obj.offsetParent) );
}

DNDUtil.prototype.replaceString = function(originalStr, replace, replaceWith){

    var index = originalStr.indexOf(replace);
    while (index >= 0) {
        
        originalStr = originalStr.substring(0, index) + replaceWith + originalStr.substring(index + replace.length, originalStr.length); 
        index = originalStr.indexOf(replace);

    }

    return originalStr;

}

// end DNDUtil object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDMultiselectSupport object definition
DNDMultiselectSupport.prototype				= new Object();
DNDMultiselectSupport.prototype.constructor = DNDMultiselectSupport;
DNDMultiselectSupport.superclass			= null;

function DNDMultiselectSupport() {
	this.selected_src_ids = new Array(0);
    this.selected_src_property_intersection = new Array(0);
}

//my constant for delimiting property values during the multiselection process
DNDMultiselectSupport.prototype.DELIM_CONSTANT = ":==--==:";

DNDMultiselectSupport.prototype.selectSource = function(dndSrcId, dndSrcPropArray){

    //try to locate the selected DNDSource
    var currentSourceObject = null;
    for (var i = 0; i < document._DNDBroker.DNDSourceArray.length; i++) {
        if (document._DNDBroker.DNDSourceArray[i].getId() == dndSrcId) {
            currentSourceObject = document._DNDBroker.DNDSourceArray[i];
            break;
        }
    }

    var elem = document.getElementById(dndSrcId);
    
    if (this.selected_src_ids.length == 0) {

        this.selected_src_property_intersection = dndSrcPropArray;
        this.selected_src_ids[this.selected_src_ids.length] = dndSrcId;
    
        elem.className = currentSourceObject.getCssClassName();
        
    }

    //deselect the current selection and select the new item
    else if(this.selected_src_ids.length > 0 && !isCtrlDepressed()){

        this.clearAllSelections();

        this.selected_src_property_intersection = dndSrcPropArray;
        this.selected_src_ids[this.selected_src_ids.length] = dndSrcId;
        elem.className = currentSourceObject.getCssClassName();

    }

    else if (this.selected_src_ids.length > 0 && isCtrlDepressed()) {

        //calculate matching subset of properties and update property values
        
        this.selected_src_property_intersection = this.computeMatchingSubset(this.selected_src_property_intersection, dndSrcPropArray);

        //if subset has no properties, deselect everything and exit
        if (this.selected_src_property_intersection.length == 0) {
            this.clearAllSelections();
            return;
        }
        
        // update id list, and style for new selection
        this.selected_src_ids[this.selected_src_ids.length] = dndSrcId;
        elem.className = currentSourceObject.getCssClassName();

    }

}

DNDMultiselectSupport.prototype.clearAllSelections = function(){

    //set the current selection(s) back to normal
    for (var i = 0; i < this.selected_src_ids.length; i++) {
        var oldSelection = document.getElementById(this.selected_src_ids[i]);
        oldSelection.className = "";
    }

    this.selected_src_ids = new Array(0);
    this.selected_src_property_intersection = new Array(0);


}

DNDMultiselectSupport.prototype.computeMatchingSubset = function(propArray1, propArray2){

    var new_prop_array = new Array(0);
    for (var i = 0; i < propArray1.length; i++) {
        for (var j = 0; j < propArray2.length; j++) {
            if (propArray1[i].matchesProperty(propArray2[j])) {
                new_prop_array[new_prop_array.length] = new DNDProperty(propArray2[j].getNamespace(), 
                                                                        propArray2[j].getType(), 
                                                                        propArray2[j].getName(), 
                                                                        propArray1[i].getValue() + this.DELIM_CONSTANT + propArray2[j].getValue());
            }
        }
    }

    return new_prop_array;
}

DNDMultiselectSupport.prototype.alreadySelected = function(dndSrcId){

    for (var i = 0; i < this.selected_src_ids.length; i++) {
        if (this.selected_src_ids[i] == dndSrcId) {
            return true;
        }
    }

    return false;
}

DNDMultiselectSupport.prototype.hasSelections = function(){

    return (this.selected_src_ids.length > 0);

}

DNDMultiselectSupport.prototype.getSelections = function(){

    return this.selected_src_ids;

}

DNDMultiselectSupport.prototype.getProperties = function(){
    return this.selected_src_property_intersection;
}

DNDMultiselectSupport.prototype.doDelimiterReplacement = function(props, delim){

    var newPropArray = new Array(0);
    var util = new DNDUtil();

    for (var i = 0; i < props.length; i++) {

        var newProp = new DNDProperty(props[i].getNamespace(),
                                      props[i].getType(),
                                      props[i].getName(),                  
                                      util.replaceString(props[i].getValue(), this.DELIM_CONSTANT, delim));

        newPropArray[newPropArray.length] = newProp

    }

    return newPropArray;

}

// end DNDMultiselectSupport object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// DNDDragCloneSupport object definition
DNDDragCloneSupport.prototype				= new Object();
DNDDragCloneSupport.prototype.constructor = DNDDragCloneSupport;
DNDDragCloneSupport.superclass			= null;

function DNDDragCloneSupport() {

}

DNDDragCloneSupport.prototype.beginDrag = function(selectedSourceArray, dragClone) {
    for (var i = 0; i < selectedSourceArray.length; i++) {
        
        //first we want to grab the original markup and change its id to have suffix ".orig"
        var originalMarkup = document.getElementById(selectedSourceArray[i]);

        //for performance purposes in IE, we want to make sure we're not copying any more markup than is necessary.
        if (dragClone && !originalMarkup.fromChildFrame) {
            originalMarkup.id = originalMarkup.id + ".orig";
            
            //now we want to create a clone with id=selectedSource.getId() that will be the markup
            //dragged by the broker
            var clonedMarkup = originalMarkup.cloneNode(true);
            clonedMarkup.id = selectedSourceArray[i];
            //change the zIndex to move above the other markup
            clonedMarkup.style.zIndex = DNDZindex;
            //change the "height" and "width" of the new div so it doesn't fill the page
            //clonedMarkup.style.height=originalMarkup.offsetHeight;
            clonedMarkup.style.width=originalMarkup.offsetWidth;

            clonedMarkup.mouseOffsetX = originalMarkup.mouseOffsetX;
            clonedMarkup.mouseOffsetY = originalMarkup.mouseOffsetY;

            //add the clonedMarkup to the document
            document.body.appendChild(clonedMarkup);

            //set the x and y coordinates for the clonedMarkup
            var util = new DNDUtil();
            clonedMarkup.style.position = "absolute";

            clonedMarkup.style.left =  document._DNDBroker.currentPageX - clonedMarkup.mouseOffsetX + "px";
            clonedMarkup.style.top =  document._DNDBroker.currentPageY - clonedMarkup.mouseOffsetY + "px";

        }
        else{

            if (!document._DNDBroker.isBidi) {

                // remember original values before we change anything
                originalMarkup.oldZIndex = originalMarkup.style.zIndex;
                originalMarkup.oldPositionStyle = originalMarkup.style.position;
                originalMarkup.oldLeft = originalMarkup.style.left;
                originalMarkup.oldTop = originalMarkup.style.top;
                originalMarkup.oldDisplay = originalMarkup.style.display;
                originalMarkup.oldWidth = originalMarkup.style.width;
                originalMarkup.oldTextAlign = originalMarkup.style.textAlign;
                originalMarkup.oldOpacity = originalMarkup.style.opacity;
                originalMarkup.oldFilter = originalMarkup.style.filter;
                
                //now make changes
                originalMarkup.style.display = "block";
                var offsetWidth = originalMarkup.offsetWidth;
                originalMarkup.style.zIndex = DNDZindex;
                originalMarkup.style.position = "absolute";
                originalMarkup.style.left = document._DNDBroker.currentPageX - originalMarkup.mouseOffsetX + "px";
                originalMarkup.style.top = document._DNDBroker.currentPageY - originalMarkup.mouseOffsetY + "px";
                originalMarkup.style.width = offsetWidth + "px";
                originalMarkup.style.textAlign="left";
            }
            else{
                if (document._DNDBroker.bd == null) {
                    document._DNDBroker.bd = new DNDBrowserDimensions();
                }

                // remember original values before we change anything
                originalMarkup.oldZIndex = originalMarkup.style.zIndex;
                originalMarkup.oldPositionStyle = originalMarkup.style.position;
                originalMarkup.oldLeft = originalMarkup.style.left;
                originalMarkup.oldTop = originalMarkup.style.top;
                originalMarkup.oldDisplay = originalMarkup.style.display;
                originalMarkup.oldWidth = originalMarkup.style.width;
                originalMarkup.oldTextAlign = originalMarkup.style.textAlign;
                originalMarkup.oldOpacity = originalMarkup.style.opacity;
                originalMarkup.oldFilter = originalMarkup.style.filter;
                
                //now make changes
                originalMarkup.style.display = "block";
                var offsetWidth = originalMarkup.offsetWidth;
                originalMarkup.style.zIndex = DNDZindex;
                originalMarkup.style.position = "absolute";
                originalMarkup.style.top = document._DNDBroker.currentPageY - originalMarkup.mouseOffsetY + "px";
                originalMarkup.style.width = offsetWidth + "px";
                originalMarkup.style.textAlign="right";
                
                if (isMozilla() && !originalMarkup.fromChildFrame) {
                    originalMarkup.style.left = (document._DNDBroker.currentPageX - (offsetWidth + originalMarkup.mouseOffsetX)) + "px";
                }
                else if (!originalMarkup.fromChildFrame){
                    originalMarkup.style.left = (document._DNDBroker.currentPageX - originalMarkup.mouseOffsetX) + "px";
                }
                else{
                    
                    if (isMozilla()) {
                        originalMarkup.style.left = -1 * (document._DNDBroker.bd.getHTMLElementWidth() - (document._DNDBroker.currentPageX - originalMarkup.mouseOffsetX))  + "px";
                    }
                    else{
                        originalMarkup.style.left = -1 * (document._DNDBroker.bd.getHTMLElementWidth() - document._DNDBroker.currentPageX) + "px";
                    
                    }
                }
    
            }
    
            if(originalMarkup.id.indexOf("_frame") == -1 ){
                originalMarkup.style.opacity = 0.4;
    	        originalMarkup.style.filter = 'alpha(opacity=' + 40 + ')';
    	    }   
        }
    }
}
DNDDragCloneSupport.prototype.endDrag = function(selectedSourceArray, dragClone) {
    for (var i = 0; i < selectedSourceArray.length; i++) {
        
        var markup = document.getElementById(selectedSourceArray[i]);

        //for performance purposes in IE, we want to make sure we're not copying any more markup than is necessary.
        if (dragClone && !markup.fromChildFrame) {
            var originalMarkup = document.getElementById(selectedSourceArray[i] + ".orig");

            document.body.removeChild(markup);
            markup = null;
            originalMarkup.id = selectedSourceArray[i];

        }
        else{
            //go back to the old zIndex
            markup.style.zIndex = markup.oldZIndex;

            //go back to the original position style
            markup.style.position = markup.oldPositionStyle;

            //go back to the original location
            markup.style.left = markup.oldLeft;
            markup.style.top = markup.oldTop;
            markup.style.display = markup.oldDisplay;
            markup.style.width = markup.oldWidth;
            markup.style.textAlign = markup.oldTextAlign;
            markup.style.opacity = markup.oldOpacity;
            markup.style.filter = markup.oldFilter;

            //reset all the variables
            markup.oldZIndex = null;
            markup.oldPositionStyle = null;
            markup.oldLeft = null;
            markup.oldTop = null;
            markup.oldDisplay = null;
            markup.oldWidth = null;
            markup.oldTextAlign = null;
            markup.oldOpacity = null;
            markup.oldFilter = null;

            //to clean up during frame support drag
            if (markup.fromChildFrame) {
                document.body.removeChild(markup);
            }
        }
    }
}

// end DNDDragCloneSupport object definition
//////////////////////////////////////////////////////////////////


//////////////////////////////////////////////////////////////////
// begin DNDAction object definition
DNDAction.prototype				= new Object();
DNDAction.prototype.constructor = DNDAction;
DNDAction.superclass			= null;

function DNDAction(id, action, dndPropertyArray, activeCssClassName, awareCssClassName, supportsMultiselect, multiselectDelimiter) {
    
    this.id = id;
    this.action = action;
    this.properties = dndPropertyArray;
    this.activeCssClassName = activeCssClassName;
    this.awareCssClassName = awareCssClassName;
    this.supportsMultiselect = supportsMultiselect;
    this.multiselectDelimiter = multiselectDelimiter;

    //adding this per LI 4938
    this.ALTERNATE_TARGET_SUBMIT_FUNC = null;

}

DNDAction.prototype.getId = function(){
    return this.id;
}

DNDAction.prototype.getAction = function(){

    return this.action;

}

DNDAction.prototype.getDNDProperties = function(){

	return this.properties;

}

DNDAction.prototype.optionalActionJavascript = function(){

    return this.oaj();

}

DNDAction.prototype.hasOptionalActionJavascript = function(){

    if (this.oaj == null) {
        var dynFuncArray = document._DNDBroker.dynamicFunctions;
        for (var i = 0; i < dynFuncArray.length; i++) {
            if (dynFuncArray[i].name == this.getId()) {
                this.oaj = dynFuncArray[i].dFnc;
            }
        }
    }
	
	return this.oaj != null;
	
}

DNDAction.prototype.getActiveCssClassName = function(){

    return this.activeCssClassName;

}

DNDAction.prototype.getAwareCssClassName = function(){

    return this.awareCssClassName;

}

DNDAction.prototype.getSupportsMultiselect = function(){

    return this.supportsMultiselect;

}

DNDAction.prototype.getMultiselectDelimiter= function(){

    return this.multiselectDelimiter;

}

DNDAction.prototype.toString = function() {
    
    var str = "DNDAction.toString() - (id: " + this.getId() + 
        "), (action: " + this.getAction() + 
        "), (activeCssClassName: " + this.getActiveCssClassName() + 
	    "), (awareCssClassName: " + this.getAwareCssClassName() + 
        "), (supportsMultiselect: " + this.getSupportsMultiselect() + 
        "), (multiselectDelimiter: " + this.getMultiselectDelimiter() + 
        "), (DNDProperties: "; 

    var properties_array = this.getDNDProperties();
    for (var i = 0; i < properties_array.length; i++) {
        if (i > 0) {
            str = str + ", ";
        }
        str = str + "[" + properties_array[i].toString() + "]";
    }

    str = str + ")";

    return str;

}
// end DNDAction object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// begin DNDProperty object definition
DNDProperty.prototype				= new Object();
DNDProperty.prototype.constructor = DNDProperty;
DNDProperty.superclass			= null;

function DNDProperty(namespace, type, name, value) {

    this.namespace = namespace;
    this.type = type;
    this.name = name;
    this.value = value;

}

DNDProperty.prototype.getNamespace = function(){

    return this.namespace;

}

DNDProperty.prototype.getType = function(){

    return this.type;

}

DNDProperty.prototype.getName = function(){

    return this.name;

}

DNDProperty.prototype.getValue = function(){

    return this.value;

}

DNDProperty.prototype.matchesProperty = function(prop){

    if (this.getNamespace() == prop.getNamespace() &&
        this.getType() == prop.getType()) {
        return true;
    }

    return false;

}

DNDProperty.prototype.toString = function() {
    
    var str = "DNDProperty.toString() - (namespace: " + this.getNamespace() + 
        "), (type: " + this.getType() + 
        "), (name: " + this.getName() + 
        "), (value: " + this.getValue() +  ")"; 
    return str;

}
// end DNDProperty object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// begin DNDTarget object definition
DNDTarget.prototype				= new Object();
DNDTarget.prototype.constructor = DNDTarget;
DNDTarget.superclass			= null;

function DNDTarget(id, dndActionArray) {
    
    this.id = id;
    this.actions = dndActionArray;       

}

DNDTarget.prototype.getId = function() {

    return this.id;

}

DNDTarget.prototype.getActions = function(){

	return this.actions;

}

DNDTarget.prototype.toString = function() {
    
    var str = "DNDTarget.toString() - (id: " + this.getId() + 
        "), (actions: "; 

    var actions_array = this.getActions();
    for (var i = 0; i < actions_array.length; i++) {
        if (i > 0) {
            str = str + ", ";
        }
        str = str + "[" + actions_array[i].toString() + "]";
    }

    str = str + ")";

    return str;

}

// end DNDTarget object definition
//////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////
// begin DNDSource object definition
DNDSource.prototype				= new Object();
DNDSource.prototype.constructor = DNDSource;
DNDSource.superclass			= null;

function DNDSource(id, dndPropertyArray, cssClassName, dragClone, multiselectable) {
    
    this.id = id;
    this.properties = dndPropertyArray;
    this.cssClassName = cssClassName;
    this.dragClone = dragClone;
    this.multi = multiselectable;

}

DNDSource.prototype.getId = function() {

    return this.id;

}

DNDSource.prototype.optionalOnDragJavascript = function(){

    return this.oodj();

}

DNDSource.prototype.hasOptionalOnDragJavascript = function(){

    if (this.oodj == null) {
        var dynFuncArray = document._DNDBroker.dynamicFunctions;
        for (var i = 0; i < dynFuncArray.length; i++) {
            if (dynFuncArray[i].name == this.getId()) {
                this.oodj = dynFuncArray[i].dFnc;
            }
        }
    }
	
    return this.oodj != null;

}

DNDSource.prototype.getDNDProperties = function(){

	return this.properties;

}

DNDSource.prototype.getCssClassName = function(){

	return this.cssClassName;

}

DNDSource.prototype.getDragClone = function(){

	return this.dragClone;

}

DNDSource.prototype.getMultiselectable = function(){

	return this.multi;

}

DNDSource.prototype.toString = function() {
    
    var str = "DNDSource.toString() - (id: " + this.getId() + 
        "), (cssClassName: " + this.getCssClassName() + 
        "), (dragClone: " + this.getDragClone() + 
        "), (DNDProperties: "; 

    var properties_array = this.getDNDProperties();
    for (var i = 0; i < properties_array.length; i++) {
        if (i > 0) {
            str = str + ", ";
        }
        str = str + "[" + properties_array[i].toString() + "]";
    }

    str = str + ")";

    return str;

}

// end DNDSource object definition
//////////////////////////////////////////////////////////////////


//////////////////////////////////////////////////////////////////
// begin DNDBrowserDimensions object definition
DNDBrowserDimensions.prototype				= new Object();
DNDBrowserDimensions.prototype.constructor = DNDBrowserDimensions;
DNDBrowserDimensions.superclass			= null;

function DNDBrowserDimensions(){

    this.body = document.body;
    if (this.isStrictDoctype() && !this.isSafari()) {
        this.body = document.documentElement;
    }

}

DNDBrowserDimensions.prototype.getScrollFromLeft = function(){
    return this.body.scrollLeft ;
}

DNDBrowserDimensions.prototype.getScrollFromTop = function(){
    return this.body.scrollTop ;
}

DNDBrowserDimensions.prototype.getViewableAreaWidth = function(){
    return this.body.clientWidth ;
}

DNDBrowserDimensions.prototype.getViewableAreaHeight = function(){
    return this.body.clientHeight ;
}

DNDBrowserDimensions.prototype.getHTMLElementWidth = function(){
    return this.body.scrollWidth ;
}

DNDBrowserDimensions.prototype.getHTMLElementHeight = function(){
    return this.body.scrollHeight ;
}

DNDBrowserDimensions.prototype.isStrictDoctype = function(){

    return (document.compatMode && document.compatMode != "BackCompat");

}

DNDBrowserDimensions.prototype.isSafari = function(){

    return (navigator.userAgent.toLowerCase().indexOf("safari") >= 0);

}

// end DNDBrowserDimensions object definition
//////////////////////////////////////////////////////////////////


//////////////////////////////////////////////////////////////////
// initialization code

document._DNDBroker = new DNDBroker();
//use this one to shorten markup
document._DNDB = document._DNDBroker;
document._DNDBroker.oldOnKeyDown = document.onkeydown;
document._DNDBroker.oldOnKeyUp = document.onkeyup;
document.onkeydown = keydown;
document.onkeyup = keyup;

// end initialization code
//////////////////////////////////////////////////////////////////
