package domain

import "strings"

func NormalizeGithubUsername(username string) string {
	return strings.ToLower(username)
}
