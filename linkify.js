var URL = "http://www.archives.nd.edu/cgi-bin/wordz.pl?keyword=";
var WORDS_REGEX = /([A-Za-z]+)(?=\s|<br|[.,;?!:])/g;

var ps = document.querySelectorAll("p:not(.pagehead):not(.smallborder):not(.margin)");
function replacer(match, cap, _, _) {
    if(match.includes("nbsp"))
        return match;
    else
        return `<a href="${URL}${cap}" target="_blank">${cap}</a>`;
}
for(let i = 0; i < ps.length; i++) {
    ps[i].innerHTML = ps[i].innerHTML.replace(WORDS_REGEX, replacer);
}
