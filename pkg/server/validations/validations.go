package validations

import "regexp"

var emailRE = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func ValidateEmail(email string) bool {
	return emailRE.MatchString(email)
}
