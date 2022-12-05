package build

func getGitBranch() string {
	v, err := runError("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "main"
	}
	return string(v)
}

func getGitSha() string {
	v, err := runError("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "XPay-Build"
	}
	return string(v)
}
