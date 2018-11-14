function setCookie(c_name,value,exdays)
{
    var exdate=new Date();
    exdate.setDate(exdate.getDate() + exdays);
    var c_value=escape(value) + ((exdays==null)
                                 ? "" : "; expires="+exdate.toUTCString())
                                + "; path=/";
    document.cookie=c_name + "=" + c_value;
}

 function sortStrings(a, b)
 {
    var x = a.toLowerCase();
    var y = b.toLowerCase();
    if (x < y) {return -1;}
    if (x > y) {return 1;}
    return 0;
}

function splitLines(t) { 
	return t.split(/\r\n|\r|\n/); 
}


function stars6(avgScore)
{
	var avgScoreMax6 = 1400
	var avgScoreMin6 = 700
    return Math.max(0, 10 * (avgScore-avgScoreMin6)/(avgScoreMax6-avgScoreMin6))
}