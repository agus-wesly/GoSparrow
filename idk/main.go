package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	// "strings"
)

var BASE string = "https://www.tiktok.com/api/comment/list/?WebIdLastTime=1738839245&aid=1988&app_language=en&app_name=tiktok_web&aweme_id=7448124030651092231&browser_language=en-US&browser_name=Mozilla&browser_online=true&browser_platform=Linux%20x86_64&browser_version=5.0%20%28X11%3B%20Linux%20x86_64%29%20AppleWebKit%2F537.36%20G%28KHTML%2C%20like%20Gecko%29%20Chrome%2F128.0.0.0%20Safari%2F537.36&channel=tiktok_web&cookie_enabled=true&count=20&cursor=27&data_collection_enabled=true&device_id=7468257643267081735&device_platform=web_pc&focus_state=false&from_page=video&history_len=6&is_fullscreen=false&is_page_visible=true&odinId=7468244277371683858&os=linux&priority_region=&referer=&region=ID&screen_height=768&screen_width=1366&tz_name=Asia%2FJakarta&user_is_login=true&webcast_language=en&msToken=KNJ-insWtCe_sXzEa-42_Y4DQg6Ps7uSLGHZm2KbzJRQ3Y9UG7zH6rciiQrp7sFEYwhdDfYKycz2dLUYiEcXuYn28_GxmDuxXACGwp6A3q3KMNfKRXP5IBwq6aoRBVCSzcEJvjiiePyVrhI=&X-Bogus"

func convert(s string, newCursor int) string {
	urlResult, err := url.Parse(s)
	if err != nil {
		log.Fatalln(err)
	}
	q := urlResult.Query()
	q.Set("cursor", strconv.Itoa(newCursor))
	urlResult.RawQuery = q.Encode()
	return urlResult.String()
}

func foo(s string) string {
	if s[0] != '@' {
		return ""
	}
	i := 0
	for i < len(s) && s[i] != ' ' {
		i++
	}
	return strings.Trim(s[i:], " ")
}

// https://x.com/search?q=soft%20spoken%20min_replies%3A100%20min_faves%3A100%20lang%3Aid&src=typed_query

func cleanupContent(content string) string {
	if content == "" || content[0] != '@' {
		return content
	}
	idx := 0
	for idx < len(content) && content[idx] != ' ' {
		idx++
	}
	res := strings.Trim(content[idx:], " ")
	return cleanupContent(res)
}

func main() {
    foo := make(map[string]string)
    foo["bar"] = "buz"
    foo["bar"] = "fc"
    foo["damn"] = "fc"
    fmt.Println(len(foo))
}
