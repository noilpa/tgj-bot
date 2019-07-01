package testHelpers

import (
	"math/rand"

	"tgj-bot/models"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers  = "0123456789"
	all = letters + numbers
)

func String() string {
	return Stringn(10)
}

func Stringn(n int) string {
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		s[i] = byte(all[rand.Intn(len(all))])
	}
	return string(s)
}

func Int() int {
	return rand.Int()
}

func Int32() int32 {
	return rand.Int31()
}

func Int64() int64 {
	return rand.Int63()
}

func Bool() bool {
	return rand.Int() % 2 == 0
}

func Role() models.Role {
	return models.ValidRoles[rand.Intn(len(models.ValidRoles))]
}
