package layouts

func inSection(active string, members ...string) bool {
	for _, m := range members {
		if active == m {
			return true
		}
	}
	return false
}
