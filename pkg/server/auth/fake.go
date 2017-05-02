package auth

type Fake struct{}

func (*Fake) GenerateToken(email string) (string, error) {
	return "good token", nil
}

func NewFake() Auth {
	return new(Fake)
}
