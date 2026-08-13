var isXblibInit = false ;

var isOther = true;	// miscellaneous non-supported browsers

var isNav = false;
var isNav4 = false;
var isNav6 = false ;
var isIE = false ;
var isIE4 = false ;
var isIE5 = false ;
var isIE5Up = false ;
var isIE5_5 = false ;
var isW3C = false ;
var isWin = false;
var isMac = false;
var isX11 = false;
var isMoz5 = false;

var docAll = null;

var verMajor = 0;
var verMinor = 0;

var getElem = function(){}

xblibInit();

if ( isX11 )
{
	document.write( '<LINK REL=stylesheet HREF="/site/css/main_nav4_x.css" TYPE="text/css">' );
}
else if ( isIE && isMac )
{
	document.write( '<LINK REL=stylesheet HREF="/site/css/main_ie_mac.css" TYPE="text/css">' );
}
else if ( isMoz5 )
{
	document.write( '<LINK REL=stylesheet HREF="/site/css/mozilla.css" TYPE="text/css">' );
}
else
{
	document.write( '<LINK REL=stylesheet HREF="/site/css/main.css" TYPE="text/css">' );
}
var contextID = "0";


function xblibInit ()
{
	docAll = document;

	var ver = navigator.appVersion.toLowerCase();
	var agt = navigator.userAgent.toLowerCase(); 

	isWin = (-1 != ver.indexOf( "win" ) ); 
	isMac = (-1 != ver.indexOf( "mac" ) ); 
	isX11 = (-1 != ver.indexOf( "x11" ) ); 
	isMoz5 = (-1 != ver.indexOf( "5.0" ) ); 

	var pos = ver.indexOf( "msie" );
	if ( -1 == pos )
	{
		isNav = true;
		verMajor = parseInt( ver ); 
		verMinor = parseFloat( ver ); 
	}
	else
	{
		verMajor = parseInt( ver.slice( pos + 4 ) ); 
		verMinor = parseFloat( ver.slice( pos + 4 ) ); 
	}

	if ( document.layers )
	{
		// check for releases prior to 4.06 - they stink!
		isNav4 = ( verMinor >= 4.06 );
		isOther = !isNav4;
	}
	else
	{
		if ( document.all )
		{
			isIE = true ;
			isIE4 = true ;
			isOther = false;
			docAll = document.all ;
		}

		if ( document.getElementById )
		{
			isIE4 = false ;
			isIE5Up = (isIE && (verMajor >= 5));
			isIE5_5 = (isIE && (verMinor >= 5.5));
			isIE5 = (isIE && (verMinor < 5.5));
			isMoz5 = (!isIE);
			isW3C = true ;
			isOther = false;
			docAll = document;
			isNav6 = (-1 != agt.indexOf( "netscape" ) );
		}
	}
	isXblibInit = true ;
}
