package auth

import "encoding/base64"

func BasicAuthHeader(email, token string) string {
	credentials := email + ":" + token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}
