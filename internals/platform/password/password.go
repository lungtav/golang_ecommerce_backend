package password

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error) {
	byte ,err:=bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err!=nil{
		return "", err
	}

	return string(byte), nil
}

func Compare(password, password_hash string) bool{
	err:=bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))

	return err == nil
}