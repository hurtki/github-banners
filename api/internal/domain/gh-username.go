package domain

import "strings"

// NormalizeGithubUsername is the only source of
// github username normalization
// used in repository level, migrations, cache level, domain surroundings
func NormalizeGithubUsername(username string) string {
	return strings.ToLower(username)
}
