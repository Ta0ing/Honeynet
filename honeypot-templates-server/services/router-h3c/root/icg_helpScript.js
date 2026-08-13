function popHelp(thepage)
{
	if(window.win1 && !window.win1.closed)
	{
		 window.win1.close();
	}
var x=screen.availWidth / 4;
var y=screen.availHeight / 4;
win1=window.open(thepage,"Help",'width=600,height=450,toolbar=no,status=no,scrollbars=yes,resizable=yes,menubar=no');
win1.moveTo(x,y);
}
function popRestoreHelp(thepage)
{
	if(window.win1 && !window.win1.closed)
	{
		 window.win1.close();
	}
	var x=screen.availWidth / 4;
	var y=screen.availHeight / 4;
	
	win1=window.open(thepage,"Help",'width=700,height=450,toolbar=no,status=no,scrollbars=yes,resizable=yes,menubar=no');
	
	
	win1.moveTo(x,y);
}
function pop_faq()
{ 
    if(window.win1 && !window.win1.closed)
    {
        window.win1.close();
    }
    var x=screen.availWidth / 4;
    var y=screen.availHeight / 4;

    win1=window.open("faq.htm","ÓÃ»§FAQ",'width=700,height=450,toolbar=no,status=no,scrollbars=yes,resizable=yes,menubar=no');
    win1.moveTo(x,y);
}
function popupHelp(thePage)
{
	if(window.win1 && !window.win1.closed)
	{
		 window.win1.close();
	}
var x=screen.availWidth / 4;
var y=screen.availHeight / 4;
win1=window.open(thePage,"Help",'width=600,height=450,toolbar=no,status=no,scrollbars=yes,resizable=yes,menubar=no');
win1.moveTo(x,y);
}
function GURL(x){location=x}
function makesure(p,l){if (confirm(p)) GURL(l);}
function wizquit()
{
if(window.confirm("The Wizard has not yet finished setting up your OfficeConnect Wireless Cable/DSL Gateway.\n\nAre you sure you want to exit?")==true)
{
GURL('QS_ethernet_dynamic.htm');
}
}
function closePopup()
{   top.close()
}



