package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword солит и хэширует пароль перед сохранением в базу.
// Сам пароль в базе не хранится нигде и никогда.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword сравнивает пароль из запроса с хэшем из базы.
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
