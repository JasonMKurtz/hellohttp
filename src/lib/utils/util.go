package utils

func InList(list []string, needle string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}

	return false
}
