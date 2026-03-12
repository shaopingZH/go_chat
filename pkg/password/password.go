package password

import "golang.org/x/crypto/bcrypt"

func Hash(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func Verify(hashed, raw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw))
	return err == nil
}
