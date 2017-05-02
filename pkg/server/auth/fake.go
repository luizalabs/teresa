package auth

type Fake struct{}

func (*Fake) GenerateToken(email string) (string, error) {
	return "good token", nil
}

func (*Fake) ValidateToken(token string) (string, error) {
	return "gopher@luizalabs.com", nil
}

func NewFake() Auth {
	return new(Fake)
}
