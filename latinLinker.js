// I need to make it show on LatinLibrary pages. Then when it's clicked I need to make it
// actually change the text. So that should be... possible. In time.
var gettingActiveTab = browser.tabs.query({active: true, currentWindow: true});
gettingActiveTab.then((tabs) => {
    browser.pageAction.show(tabs[0].id);
});

/*
 * Restart alarm for the currently active tab, whenever the user navigates.
 * */
browser.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (!changeInfo.url) {
        return;
    }
    var gettingActiveTab = browser.tabs.query({active: true, currentWindow: true});
    gettingActiveTab.then((tabs) => {
        if (tabId == tabs[0].id) {
            restartAlarm(tabId);
        }
    });
});
