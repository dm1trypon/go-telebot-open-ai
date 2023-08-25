package gdapi

type CredentialsSettings struct {
	ClientID                string
	ProjectID               string
	AuthURI                 string
	TokenURI                string
	AuthProviderX509CertURL string
	ClientSecret            string
}

type TokenSettings struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       string
}
