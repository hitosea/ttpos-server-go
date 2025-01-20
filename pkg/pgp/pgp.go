package pgp

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}

func GenerateKeyPair(name, email, passphrase string) (*KeyPair, error) {
	key, err := crypto.GenerateKey(name, email, "x25519", 0)
	if err != nil {
		return nil, err
	}
	defer key.ClearPrivateParams()

	locked, err := key.Lock([]byte(passphrase))
	if err != nil {
		return nil, err
	}

	privateKey, err := locked.ArmorWithCustomHeaders("", "")
	if err != nil {
		return nil, err
	}

	publicKey, err := locked.GetArmoredPublicKeyWithCustomHeaders("", "")
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Passphrase: passphrase,
	}, nil
}

func EncryptMessage(plaintext string, publicKey string) (string, error) {
	var (
		publicKeyObj  *crypto.Key
		publicKeyRing *crypto.KeyRing
		ciphertext    *crypto.PGPMessage
		err           error
	)
	if publicKeyObj, err = crypto.NewKeyFromArmored(publicKey); err != nil {
		return "", err
	}
	if publicKeyObj.IsPrivate() {
		if publicKeyObj, err = publicKeyObj.ToPublic(); err != nil {
			return "", err
		}
	}
	if publicKeyRing, err = crypto.NewKeyRing(publicKeyObj); err != nil {
		return "", err
	}
	message := crypto.NewPlainMessageFromString(plaintext)
	if ciphertext, err = publicKeyRing.Encrypt(message, nil); err != nil {
		return "", err
	}
	return ciphertext.GetArmoredWithCustomHeaders("", "")
}

func DecryptMessage(encryptedText string, privateKey, passphrase string) (string, error) {
	// decrypt armored encrypted message using the private key and obtain plain text
	return helper.DecryptMessageArmored(privateKey, []byte(passphrase), encryptedText)
}
