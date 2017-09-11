function linkify(tab) {
    browser.tabs.executeScript({file: "linkify.js"});
}
browser.pageAction.onClicked.addListener(linkify);

browser.tabs.onUpdated.addListener((tabId, changeInfo, newTab) => {
    if(newTab.url.toLowerCase().includes("latinlibrary"))
        browser.pageAction.show(newTab.id);
    else
        browser.pageAction.hide(newTab.id);
});

