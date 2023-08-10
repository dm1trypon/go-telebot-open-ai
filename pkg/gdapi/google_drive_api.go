package gdapi

import (
	"context"
	"io"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

const timeLayout = "2006-01-02T15:04:05.9999999-07:00"

type GoogleDriveAPI struct {
	credentials *oauth2.Config
	token       *oauth2.Token
	timeout     time.Duration
}

func NewGoogleDriveAPI(credCfg CredentialsSettings, tokenCfg TokenSettings, timeout int) (*GoogleDriveAPI, error) {
	expiry, err := time.Parse(timeLayout, tokenCfg.Expiry)
	if err != nil {
		return nil, err
	}
	return &GoogleDriveAPI{
		&oauth2.Config{
			ClientID:     credCfg.ClientID,
			ClientSecret: credCfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  credCfg.AuthURI,
				TokenURL: credCfg.TokenURI,
			},
			RedirectURL: "http://localhost",
			Scopes:      []string{drive.DriveReadonlyScope},
		},
		&oauth2.Token{
			AccessToken:  tokenCfg.AccessToken,
			TokenType:    tokenCfg.TokenType,
			RefreshToken: tokenCfg.RefreshToken,
			Expiry:       expiry,
		},
		time.Duration(timeout) * time.Millisecond,
	}, nil
}

func (g *GoogleDriveAPI) Download(fileID string) ([]byte, error) {
	service, err := drive.New(g.credentials.Client(context.Background(), g.token))
	if err != nil {
		return nil, err
	}
	// Получаем информацию о файле
	_, err = service.Files.Get(fileID).Do()
	if err != nil {
		return nil, err
	}
	resp, err := service.Files.Get(fileID).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
