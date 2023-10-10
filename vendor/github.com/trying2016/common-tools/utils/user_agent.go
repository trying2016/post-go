package utils

import "strings"

func ConvertUserAgentToReadable(ua string) string {
	info := make([]string, 0, 2)
	uaLow := strings.ToLower(ua)
	if strings.HasPrefix(uaLow, "okhttp/") {
		info = append(info, "Android")
	} else {
		// device
		if strings.Contains(uaLow, "android") {
			info = append(info, "Android")
		} else if strings.Contains(uaLow, "windows") {
			info = append(info, "Windows")
		} else if strings.Contains(uaLow, "linux") {
			info = append(info, "Linux")
		} else if strings.Contains(uaLow, "iphone") {
			info = append(info, "iPhone")
		} else if strings.Contains(uaLow, "ipad") {
			info = append(info, "iPad")
		} else if strings.Contains(uaLow, "mac") {
			info = append(info, "MacOS")
		} else {
			info = append(info, "Unknown Device")
		}

		// soft
		if strings.Contains(uaLow, "trident") {
			info = append(info, "IE")
		} else if strings.Contains(uaLow, "edge") {
			info = append(info, "Edge")
		} else if strings.Contains(uaLow, "safari") {
			info = append(info, "Safari")
		} else if strings.Contains(uaLow, "firefox") {
			info = append(info, "Firefox")
		} else if strings.Contains(uaLow, "chrome") {
			info = append(info, "Chrome")
		} else {
			info = append(info, "Unknown Browser")
		}
	}

	return strings.Join(info, ",")
}
