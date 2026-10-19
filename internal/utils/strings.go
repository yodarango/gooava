package utils

import (
	"crypto/rand"
	"math/big"
)

/**************************************************************************************
* GenerateRandomString generates a random string of specified length.
*
* Used for verification codes and temporary passwords, so it MUST be
* cryptographically secure: crypto/rand via rand.Int (which also avoids the
* modulo bias of naive indexing). math/rand output is predictable and must
* never be used here.
****************************************************************************************/
func GenerateRandomString(length int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	max := big.NewInt(int64(len(letters)))

	for i := range s {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// crypto/rand only fails when the OS entropy source is broken —
			// there is no safe way to keep generating security codes
			panic("crypto/rand failed: " + err.Error())
		}
		s[i] = letters[n.Int64()]
	}
	return string(s)
}