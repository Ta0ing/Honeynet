var mimgpath="/site/image";var movercolor="#df0001";	//rgb(0, 128, 128) space seperated.var mnormalcolor="#000000";var mclickcolor="#df0000";	//rgb(0, 128, 127)var moldFocus;var NC6=(navigator.userAgent.indexOf("Netscape6")>0)?true:false;var IE=(document.all)?true:false;function mOver(id){	if (IE)	{
		if ( document.getElementById(id).style.color==mclickcolor )			;
		else			document.getElementById(id).style.color=movercolor;
	} 
	else //if (NC6)
	{ 
	     if ( document.getElementById(id).style.color!="rgb(223, 0, 0)" )
			document.getElementById(id).style.color=movercolor;
	}}function mClick(id){	if (IE)	{
		if ( moldFocus )
			document.getElementById(moldFocus).style.color=mnormalcolor;
		document.getElementById(id).style.color=mclickcolor;
		moldFocus = id;
  }
  else //if(NC6)
  { 		if (moldFocus)
	    	document.getElementById(moldFocus).style.color=mnormalcolor;
		document.getElementById(id).style.color=mclickcolor;
		moldFocus = id;
  }}function mOut(id){	if (IE)	{		if (document.getElementById(id).style.color==mclickcolor)			;
		else			document.getElementById(id).style.color=mnormalcolor; 	}	else //if(NC6)
	{ 
		if (document.getElementById(id).style.color=="rgb(223, 0, 1)")
			document.getElementById(id).style.color=mnormalcolor; 
	}}function mFolderExpand($1,$2,pic){
	if (IE)	{		mExpandIE($1,$2,pic)	} 
	else if(NC6)	{		mExpandNC($1,$2,pic)	}
	else	{		mExpandNC($1,$2,pic)	}}function mExpandIE($1,$2,pic){
	Expanda = eval($1 + "a");
	Expanda.blur()
	ExpandChild = eval($1 + "Child");
   if ($2 != "top")	{ 
		ExpandTree = eval($1 + "Tree");
		ExpandFolder = eval($1 + "Folder");
	}
	if (ExpandChild.style.display == "none")	{
		ExpandChild.style.display = "block";
       if ($2 != "top")		{ 
           if ($2 == "last")			{				ExpandTree.src = mimgpath+"/Lminus.gif";			}
			else			{				ExpandTree.src = mimgpath+"/Tminus.gif";			}
			ExpandFolder.src = mimgpath+"/openfolder"+pic+".gif";	
		}
		else		{			mmTree.src = mimgpath+"/topopen1.gif";		}
	}	else	{
		ExpandChild.style.display = "none";
       if ($2 != "top")		{ 
	        if ($2 == "last")			{				ExpandTree.src = mimgpath+"/Lplus.gif";			}
			else			{				ExpandTree.src = mimgpath+"/Tplus.gif";			}
			ExpandFolder.src = mimgpath+"/folder"+pic+".gif";
		}
		else		{			mmTree.src = mimgpath+"/top1.gif";		}
	}}function mExpandNC($1,$2,pic){
	Expanda = document.getElementById($1 + "a");
	Expanda.blur()
	ExpandChild = document.getElementById($1 + "Child");
   if ($2 != "top") { 
		ExpandTree = document.getElementById($1 + "Tree");
		ExpandFolder = document.getElementById($1 + "Folder");
	}
	if (ExpandChild.style.display == "none")	{
		ExpandChild.style.display = "block";
       if ($2 != "top")		{ 
           if ($2 == "last")			{				ExpandTree.src = mimgpath+"/Lminus.gif";			}
			else			{				ExpandTree.src = mimgpath+"/Tminus.gif";			}
			ExpandFolder.src = mimgpath+"/openfolder"+pic+".gif";	
		}
		else		{			document.getElementById("mmTree").src = mimgpath+"/topopen1.gif";		}
	}	else	{
		ExpandChild.style.display = "none";
       if ($2 != "top")		{ 
	        if ($2 == "last")			{				ExpandTree.src = mimgpath+"/Lplus.gif";			}
			else			{				ExpandTree.src = mimgpath+"/Tplus.gif";			}
			ExpandFolder.src = mimgpath+"/folder"+pic+".gif";
		}
		else		{			document.getElementById("mmTree").src = mimgpath+"/top1.gif";		}
	}}