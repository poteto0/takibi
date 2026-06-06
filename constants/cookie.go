package constants

const (
	CookieSecurePrefix     = "__Secure-"
	CookieHostPrefix       = "__Host-"
	CookieSecurePrefixMode = "secure"
	CookieHostPrefixMode   = "host"

	// MinSignedCookieSecretLen is the minimum required byte length for HMAC signing secrets.
	// gorilla/securecookie recommends 32 or 64 bytes for the hash key.
	MinSignedCookieSecretLen = 32
)
