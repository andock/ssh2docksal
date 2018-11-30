package ssh2docksal

import (
	"github.com/apex/log"
	"github.com/gliderlabs/ssh"
	"io/ioutil"
)

// NoAuth for local development and testing
func NoAuth() ssh.Option {
	return ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true
	})
}

func validatePublicKeyAuth(authorizedKeysFile string, key ssh.PublicKey) bool {
	authorizedKeysBytes, err := ioutil.ReadFile(authorizedKeysFile)
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
		return false
	}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.WithError(err)
		}
		if ssh.KeysEqual(key, pubKey) {
			return true
		}
		authorizedKeysBytes = rest
	}
	return false
}

// PublicKeyAuth  perform public key authentification
func PublicKeyAuth(authorizedKeysFile string) ssh.Option {
	return ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return validatePublicKeyAuth(authorizedKeysFile, key)
	})
}
