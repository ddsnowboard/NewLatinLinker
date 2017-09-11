// I need to make it show on LatinLibrary pages. Then when it's clicked I need to make it
// actually change the text. So that should be... possible. In time.
var gettingActiveTab = browser.tabs.query({active: true, currentWindow: true});
gettingActiveTab.then((tabs) => {
    console.log(tabs[0].id);
});
