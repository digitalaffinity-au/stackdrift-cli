package commands

import (
	"fmt"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/auth"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
)

func Login([]string) error {
	return auth.Login(config.BaseURL())
}

func Logout([]string) error {
	baseURL := config.BaseURL()
	if err := config.DeleteCredential(baseURL); err != nil {
		return err
	}
	fmt.Println("Signed out.")
	return nil
}

func Whoami([]string) error {
	client, _, err := authenticatedClient()
	if err != nil {
		return err
	}
	me, err := client.Me()
	if err != nil {
		return err
	}
	if !me.Authenticated {
		return errNotSignedIn
	}
	fmt.Println("Signed in as " + me.Email)
	return nil
}
