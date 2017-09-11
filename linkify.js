var URL = "http://www.archives.nd.edu/cgi-bin/wordz.pl?keyword=";
var WORDS_REGEX = /([A-Za-z]+)(?=\s|<br|[.,;?!:\]"]|&nbsp;)/g;
function isInTag(s, offset) {
    let walker = offset;
    for(let c = s.charAt(walker); c = s.charAt(walker); walker++) {
        if(c === "<")
            return false;
        else if(c === ">")
            break;
    }

    walker = offset;
    for(let c = s.charAt(walker); c = s.charAt(walker); walker--) {
        if(c === ">")
            return false;
        else if(c === "<")
            break;
    }
    return true;
}

var ps = document.querySelectorAll("p:not(.pagehead):not(.smallborder):not(.margin)");
function replacer(match, cap, offset, wholeString) {
    if(match.includes("nbsp") || 
        cap.toLowerCase() === "font" ||
        isInTag(wholeString, offset))
        return match;
    else
        return `<a href="${URL}${cap}" target="_blank">${cap}</a>`;
}
for(let i = 0; i < ps.length; i++) {
    ps[i].innerHTML = ps[i].innerHTML.replace(WORDS_REGEX, replacer);
}
