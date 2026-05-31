package useragent

type Browser string

const (
	BrowserAndroid   Browser = "Android Browser"
	BrowserChrome    Browser = "Chrome"
	BrowserEdge      Browser = "Edge"
	BrowserFirefox   Browser = "Firefox"
	BrowserIE        Browser = "IE"
	BrowserOpera     Browser = "Opera"
	BrowserOperaMini Browser = "Opera Mini"
	BrowserSafari    Browser = "Safari"
	BrowserVivaldi   Browser = "Vivaldi"
	BrowserSilk      Browser = "Amazon Silk"
	BrowserSamsung   Browser = "Samsung Browser"
	BrowserFalkon    Browser = "Falkon"
	BrowserNintendo  Browser = "Nintendo Browser"
	BrowserYandex    Browser = "Yandex Browser"
	BrowserUnknown   Browser = "Unknown"
)

func GetBrowser(ua string) Browser {
	agent := parser.Parse(ua)
	browser := agent.Browser()

	if browser == "" {
		return BrowserUnknown
	}

	return Browser(browser)
}
