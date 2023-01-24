package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gen2brain/beeep"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	clientID, err := readFile("secrets/client_id.txt")
	if err != nil {
		log.Fatal(err)
	}
	clientSecret, err := readFile("secrets/client_secret.txt")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: redo how this token generation stuff works;
	// i can just make a token by querying
	// https://develop.battle.net/documentation/guides/using-oauth/client-credentials-flow
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"wow.profile"},
		RedirectURL:  "http://127.0.0.1:8080/redirect",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.battle.net/authorize",
			TokenURL: "https://oauth.battle.net/token",
		},
	}
	token, err := getToken(ctx, config)

	client := config.Client(ctx, token)

	realmID, err := getRealmID(client, "Area 52")
	if err != nil {
		log.Fatal(err)
	}

	realmStatus, err := getRealmStatus(client, realmID)
	if err != nil {
		log.Fatal(err)
	}


	for {
		log.Println("Fetching server status...")
		connectedRealmStatus, err := getConnectedRealmStatus(client, realmStatus.ConnectedRealm.HREF)
		if err != nil {
			log.Fatal(err)
		}
		if connectedRealmStatus.Status.Type != "DOWN" {
			log.Println("Notifying that servers are up!")
			err = beeep.Notify("WOW SERVERS ARE UP", "LOG ON YOU NERD", "")
			if err != nil {
				log.Fatal(err)
			}
			break
		}

		time.Sleep(time.Minute * 1)
	}
}

func getConnectedRealmStatus(client *http.Client, connectedRealmURL string) (*connectedRealmStatusResponse, error) {
	return fetchJSON[connectedRealmStatusResponse](
		client,
		connectedRealmURL,
	)
}

type connectedRealmStatusResponse struct {
	Status struct {
		Type string `json:"type"`
	} `json:"status"`
}

func getRealmStatus(client *http.Client, realmID int) (*realmStatusResponse, error) {
	return fetchJSON[realmStatusResponse](
		client,
		fmt.Sprintf("https://us.api.blizzard.com/data/wow/realm/%d?namespace=dynamic-us&locale=en_US", realmID),
	)
}

type realmStatusResponse struct {
	ConnectedRealm struct {
		HREF string `json:"href"`
	} `json:"connected_realm"`
}

func getRealmID(client *http.Client, name string) (int, error) {
	realms, err := getRealms(client)
	if err != nil {
		return -1, err
	}

	for _, realm := range realms.Realms {
		if realm.Name == name {
			return realm.ID, nil
		}
	}

	return -1, errors.New("Could not find matching realm")
}

func getRealms(client *http.Client) (*realmIndexResponse, error) {
	return fetchJSON[realmIndexResponse](
		client,
		"https://us.api.blizzard.com/data/wow/realm/index?namespace=dynamic-us&locale=en_US",
	)
}

type realmIndexResponse struct {
	Realms []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		Slug string `json:"slug"`
	} `json:"realms"`
}

func fetchJSON[T any](client *http.Client, url string) (*T, error) {
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response T
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func getToken(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	token, err := loadToken()
	if err == nil {
		return token, nil
	}
	log.Println("Failed to get token, trying OAuth2:", err)

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println(url)

	codeChan := make(chan string)
	go func() {
		http.ListenAndServe("127.0.0.1:8080", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			query := req.URL.Query()
			if !query.Has("code") {
				res.Write([]byte("No auth code"))
				return
			}

			codeChan <- query.Get("code")
			res.Write([]byte("Received auth code; you can close this page now."))
		}))
	}()

	code := <-codeChan
	token, err = config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	err = saveToken(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func loadToken() (*oauth2.Token, error) {
	file, err := os.Open("secrets/token.json")
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var token *oauth2.Token = &oauth2.Token{}
	err = json.Unmarshal(bytes, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveToken(token *oauth2.Token) error {
	file, err := os.Create("secrets/token.json")
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(token)
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
