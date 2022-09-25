package cmd

import (
	"errors"
	"fmt"
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/mitchellh/go-homedir"
)

func GenericAccount() (*users.FullAccount, error) {
	dbx := users.New(config)
	res, err := dbx.GetCurrentAccount()
	if err != nil {
		return nil, err
	}
	return res, nil

}

func Init() error {
	domain := ""
	dir, err := homedir.Dir()
	if err != nil {
		return err
	}
	filePath := path.Join(dir, ".config", "dbxcli", configFileName)
	tokType := "personal"
	conf := oauthConfig(tokType, domain)

	tokenMap, err := readTokens(filePath)
	if tokenMap == nil {
		tokenMap = make(TokenMap)
	}
	if tokenMap[domain] == nil {
		tokenMap[domain] = make(map[string]string)
	}
	tokens := tokenMap[domain]

	if err != nil {
		return errors.New(fmt.Sprintf("Not logged, get the secret from: %s", conf.AuthCodeURL("state")))
	}

	logLevel := dropbox.LogOff

	config = dropbox.Config{
		Token:           tokens[tokType],
		LogLevel:        logLevel,
		Logger:          nil,
		AsMemberID:      "",
		Domain:          domain,
		Client:          nil,
		HeaderGenerator: nil,
		URLGenerator:    nil,
	}

	return nil
}
