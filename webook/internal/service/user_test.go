package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordEncrypt(t *testing.T) {
	pwd := []byte("123456#123456#11adasfasfsfsf2")
	// 加密
	encrypted, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	// 比较
	println(len(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, pwd)
	require.NoError(t, err)
}
