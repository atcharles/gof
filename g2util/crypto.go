package g2util

import (
	"strings"

	"github.com/andeya/goutil"
	"github.com/pkg/errors"
)

// Crypto ...
type Crypto struct {
	key string
}

// NewCrypto ...
func NewCrypto(key string) *Crypto {
	return &Crypto{key: key}
}

// Key ...
func (c *Crypto) Key() string {
	c.key = strings.TrimSpace(strings.Replace(c.key, "-", "", -1))
	return c.key
}

// EncryptCBC ...
func (c *Crypto) EncryptCBC(val string) string {
	return goutil.BytesToString(goutil.AESCBCEncrypt([]byte(c.Key()), []byte(val)))
}

// DecryptCBC ...
func (c *Crypto) DecryptCBC(ciphertext string) (string, error) {
	v, err := goutil.AESCBCDecrypt([]byte(c.Key()), []byte(ciphertext))
	if err != nil {
		return "", err
	}
	return goutil.BytesToString(v), nil
}

// Encrypt ...
func (c *Crypto) Encrypt(val string) string {
	return goutil.BytesToString(goutil.AESEncrypt([]byte(c.Key()), []byte(val)))
}

// Decrypt ...
func (c *Crypto) Decrypt(ciphertext string) (string, error) {
	v, err := goutil.AESDecrypt([]byte(c.Key()), []byte(ciphertext))
	if err != nil {
		return "", err
	}
	return goutil.BytesToString(v), nil
}

const key = "3f756b58-1656-11ec-879b-3c7d0a0ab31b"

// CryptKey ...
func CryptKey() string {
	return strings.TrimSpace(strings.Replace(key, "-", "", -1))
}

// EncryptPassword ...
func EncryptPassword(p string) string {
	k1 := CryptKey()[:16]
	return strings.ToUpper(NewCrypto(k1).Encrypt(p))
}

// DecryptPassword ...
func DecryptPassword(p string) (s string, err error) {
	k1 := CryptKey()[:16]
	return NewCrypto(k1).Decrypt(strings.ToLower(p))
}

// CheckPassword ...c cipher_password
// p plaintext
func CheckPassword(p, c string) error {
	s, err := DecryptPassword(c)
	if err != nil {
		return err
	}
	if s != p {
		return ErrPassword
	}
	return nil
}

var ErrPassword = errors.New("密码错误")
