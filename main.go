package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type ConfigList struct {
	CredentialUrl     string
	RepositoryInfoUrl string
}

var Config ConfigList
var Repositories map[string]string

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		panic("Error while loading config file in gddl.")
	}
	Config = ConfigList{
		CredentialUrl:     cfg.Section("config").Key("credential_url").String(),
		RepositoryInfoUrl: cfg.Section("config").Key("repository_info_url").String(),
	}

	url := Config.RepositoryInfoUrl

	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		panic(err)
	}
	byteArray, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(byteArray, &Repositories)
	if err != nil {
		panic(err)
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getService() *drive.Service {
	url := Config.CredentialUrl

	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		panic(err)
	}
	byteArray, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	config, err := google.ConfigFromJSON(byteArray, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	return srv
}

func ListRepository() []string {
	result := make([]string, 0)
	for key, _ := range Repositories {
		result = append(result, key)
	}
	return result
}

func ListDirectory(repository string) []string {
	service := getService()
	results := make([]string, 0)
	query := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and '%s' in parents", Repositories[repository])).PageSize(1000)
	result, err := query.Do()
	if err != nil {
		panic(err)
	}
	for {
		for _, file := range result.Files {
			results = append(results, file.Name)
		}
		if result.NextPageToken == "" {
			break
		}
		result, err = query.PageToken(result.NextPageToken).Do()
		if err != nil {
			panic(err)
		}
	}
	sort.Strings(results)
	return results
}

func getDirectory(repository string, directory string) (*drive.Service, *drive.File) {
	service := getService()
	result, err := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and '%s' in parents and name='%s'", Repositories[repository], directory)).PageSize(1).Do()
	if err != nil {
		panic(err)
	}
	return service, result.Files[0]
}

func ListFiles(repository string, directory string) []string {
	service, file := getDirectory(repository, directory)
	query := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("'%s' in parents", file.Id)).PageSize(1000)
	results := make([]string, 0)
	result, err := query.Do()
	if err != nil {
		panic(err)
	}
	for {
		for _, file := range result.Files {
			if strings.HasSuffix(file.Name, ".tar.xz") {
				results = append(results, file.Name[:len(file.Name)-7])
			} else {
				results = append(results, file.Name)
			}
		}
		if result.NextPageToken == "" {
			break
		}
		result, err = query.PageToken(result.NextPageToken).Do()
		if err != nil {
			panic(err)
		}
	}
	sort.Strings(results)
	return results
}

func showUsage() {
	fmt.Printf("Usage (only arguments):\n" +
		"\tShow all repositories:\n\t\tshow\n" +
		"\tShow folders in a repository:\n\t\tshow [repository]\n" +
		"\tShow download candidates in a folder:\n\t\tshow [repository] [folder]\n" +
		"\tDownload:\n\t\tdownload [repository] [folder] [target]\n" +
		"\tUpload:\n\t\tupload [repository] [folder] [file/folder]\n")
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		showUsage()
		return
	}
	if flag.Arg(0) == "show" {
		if flag.Arg(1) == "" {
			printStr := ""
			for _, key := range ListRepository() {
				if printStr == "" {
					printStr = key
				} else {
					printStr += "\n" + key
				}
			}
			fmt.Print(printStr)
		} else {
			includes := false
			for _, key := range ListRepository() {
				if key == flag.Arg(1) {
					includes = true
					break
				}
			}
			if !includes {
				fmt.Printf("Cannot found repository: %s", flag.Arg(1))
				return
			}
			if flag.Arg(2) == "" {
				printStr := ""
				for _, key := range ListDirectory(flag.Arg(1)) {
					if printStr == "" {
						printStr = key
					} else {
						printStr += "\n" + key
					}
				}
				fmt.Print(printStr)
			} else {
				includes := false
				for _, key := range ListDirectory(flag.Arg(1)) {
					if key == flag.Arg(2) {
						includes = true
						break
					}
				}
				if !includes {
					fmt.Printf("Cannot found directory: %s", flag.Arg(2))
					return
				}
				printStr := ""
				for _, key := range ListFiles(flag.Arg(1), flag.Arg(2)) {
					if printStr == "" {
						printStr = key
					} else {
						printStr += "\n" + key
					}
				}
				fmt.Print(printStr)
			}
		}
		return
	}
}
