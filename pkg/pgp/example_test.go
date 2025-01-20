package pgp

import (
	"fmt"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, _ := GenerateKeyPair("aa", "aa@qq.com", "ccc")
	fmt.Printf("%v", kp)
}

func TestEncryptData(t *testing.T) {
	publicKey := `
-----BEGIN PGP PUBLIC KEY BLOCK-----

xjMEZ3/gHBYJKwYBBAHaRw8BAQdAfGtiNvfoSkQ++UBhdG1JBnWNsze0VSCjZVwz
JgT8mdvNEzZkd3V5amJmIDxhYUBiYi5jYz7CvwQTFggAcQWCZ3/gHAMLCQcJkMMe
9jS+IqcpNRQAAAAAABwAEHNhbHRAbm90YXRpb25zLm9wZW5wZ3Bqcy5vcmcB8Q9r
uBQKB2/7fKLN71eqAhUIAxYAAgIZAQKbAwIeARYhBNA3cP0KO/kiK6CiksMe9jS+
IqcpAABVSAEA94O0m08CCstxnI8oGTLdgirVKqmCUXwoGFVoIvPW8VsBAL5gDLlI
kPB40G+VmR0uxRbjYNu2IJwUNVSfGCtqrqAKzjgEZ3/gHBIKKwYBBAGXVQEFAQEH
QHt1uL3r3v+XE75f8wjU5rfhsl5AYU8yFYiVKVVB6ukvAwEKCcKuBBgWCABgBYJn
f+AcCZDDHvY0viKnKTUUAAAAAAAcABBzYWx0QG5vdGF0aW9ucy5vcGVucGdwanMu
b3JnaWrNcmx9l1sRq8cJMUvx3AKbDBYhBNA3cP0KO/kiK6CiksMe9jS+IqcpAAAE
rAD+OyQoNBXqdXnwD05Oy4Ma04fdN/MpdHd/BI+58H57XJ0A/0Xw9Q4EFmDX55kg
booWIudAxXMiewMHK3MwBxtXzDQJ
=FGIs
-----END PGP PUBLIC KEY BLOCK-----
`
	encryptData, _ := EncryptMessage("foo", publicKey)
	fmt.Printf("%v", encryptData)
}

func TestDecryptData(t *testing.T) {
	encrypted := `
-----BEGIN PGP MESSAGE-----

wV4D+ProJQoioP8SAQdAwcv1b9jIPTz0mgxU0Je4+vg+5YeTEEZbwSE7PlgICXYw
sMkZaEgigpiYNTxa3LoStBKT8GfKU9FZ++4Jr7vRFGDoX7moeDjI+Y/qZDIR9NaZ
0j4Bsb+hYSgtnUbPreBx9bfoIR9E31shafNVjZaMY6dHbI0xFR9VF/hl5/5MJT3E
4oEbCpTHSbHrsMmejrKd9A==
=Y9MV
-----END PGP MESSAGE--------
`
	privateKey := `
-----BEGIN PGP PRIVATE KEY BLOCK-----

xYYEZ3/eHBYJKwYBBAHaRw8BAQdAXHZk/9V34EVtIxxsbCdweCI24gkcLjyvQIw/
X3dfglf+CQMIzixB6GemGtxgtjF2PvfItxD+CuOZYKeYeFz1vKtEKBIAIuMPaleI
1tyvHQegd6H1Kv7LoYWAKtvpfFLtvF6qhi/I5NPWrLSGzze27LsCeM0TaXd5bmVl
eW8gPGFhQGJiLmNjPsK/BBMWCABxBYJnf94cAwsJBwmQDFLZpOYPkes1FAAAAAAA
HAAQc2FsdEBub3RhdGlvbnMub3BlbnBncGpzLm9yZ3otPnHnVQDovovTHhSyh4kC
FQgDFgACAhkBApsDAh4BFiEEDuhtveC0lZrjx3hLDFLZpOYPkesAAEjxAQCGZJgH
l35JXwquYCfaqtXRxi+fV2K/zMNZ1q2Wg820wQEA6H/FSDycTZCEdYYDWSI1dGHn
1EtKgWV/jDV/7Dt5sQrHiwRnf94cEgorBgEEAZdVAQUBAQdAO+6xoQnFPw5MdBEG
XM3AiW7+dGnE2q/7bFA97u7uyE4DAQoJ/gkDCK72jmFQtxFJYF24gqpVAqOCcbpP
Aiwy7+19pHnxla6D+8VNgcsURITQ0v2pWVkXq4WbxTcnH14VMv4dQAriA0urZHa0
EL9sazenVhTw1xjCrgQYFggAYAWCZ3/eHAmQDFLZpOYPkes1FAAAAAAAHAAQc2Fs
dEBub3RhdGlvbnMub3BlbnBncGpzLm9yZ1Yn/CqSdrAr8rrO2zi9S3gCmwwWIQQO
6G294LSVmuPHeEsMUtmk5g+R6wAAvP4BAPThIc8gxCIzKEok6UhuMs7ghJ5J1ehw
AiBL/mOota4GAPsErWZelf6wyXD9bIZlSODhg9h4gtusJgAIprI9bbvnAQ==
=T1/x
-----END PGP PRIVATE KEY BLOCK-----
`
	raw, e := DecryptMessage(encrypted, privateKey, "fjbfubu4")
	if e != nil {
		panic(e)
	}
	fmt.Printf("%v", raw)
}
