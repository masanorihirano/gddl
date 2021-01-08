package gddl

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mholt/archiver/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	Config = ConfigList{
		CredentialUrl:     "https://gist.githubusercontent.com/masanorihirano/d9464a1d8684d794d7e0a0136347f9bb/raw/gddl-credential",
		RepositoryInfoUrl: "https://gist.githubusercontent.com/masanorihirano/d9464a1d8684d794d7e0a0136347f9bb/raw/gddl-settings",
	}

	url := Config.RepositoryInfoUrl

	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
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
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(ex)
	tokFile := filepath.Join(dir, "gddl_token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err2 := getTokenFromWeb(config)
		if err2 != nil {
			return nil, errors.New(fmt.Sprintf("Error while getting token: %s, %s", err, err2))
		}
		err3 := saveToken(tokFile, tok)
		if err3 != nil {
			return nil, err3
		}
	}
	tok, err = tokenFromFile(tokFile)
	if err != nil {
		return nil, err
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read authorization code %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to retrieve token from web %v", err))
	}
	return tok, nil
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
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to cache oauth token: %v", err))
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}

func getService() (*drive.Service, error) {
	url := Config.CredentialUrl

	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to connect online credential file: %v", err))
	}
	byteArray, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to load online credential file: %v", err))
	}

	config, err := google.ConfigFromJSON(byteArray, drive.DriveScope)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}
	client, err := getClient(config)
	if err != nil {
		return nil, err
	}

	srv, err := drive.New(client)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to retrieve Drive client: %v", err))
	}
	return srv, nil
}

func ListRepository() []string {
	result := make([]string, 0)
	for key, _ := range Repositories {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func ListDirectory(repository string) ([]string, error) {
	service, err := getService()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to get service: %v", err))
	}
	results := make([]string, 0)
	query := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and '%s' in parents", Repositories[repository])).PageSize(1000)
	result, err := query.Do()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
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
			return nil, errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
		}
	}
	sort.Strings(results)
	return results, nil
}

func getDirectory(repository string, directory string) (*drive.Service, *drive.File, error) {
	service, err := getService()
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Unable to get service: %v", err))
	}
	result, err := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and '%s' in parents and name='%s'", Repositories[repository], directory)).PageSize(1).Do()
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
	}
	return service, result.Files[0], nil
}

func ListFiles(repository string, directory string) ([]string, error) {
	service, file, err := getDirectory(repository, directory)
	if err != nil {
		return nil, err
	}
	query := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("'%s' in parents", file.Id)).PageSize(1000)
	results := make([]string, 0)
	result, err := query.Do()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
	}
	for {
		for _, file := range result.Files {
			if strings.HasSuffix(file.Name, ".tar.xz") {
				results = append(results, file.Name[:len(file.Name)-7])
			} else if strings.HasSuffix(file.Name, ".tar.gz") {
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
			return nil, errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
		}
	}
	sort.Strings(results)
	return results, nil
}

func DownloadAndSave(path string, repository string, directory string, fileName string, saveForce bool, unfreeze bool) error {
	service, file, err := getDirectory(repository, directory)
	if err != nil {
		return err
	}
	result, err := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("'%s' in parents and (name='%s' or name='%s.tar.xz' or name='%s.tar.gz')", file.Id, fileName, fileName, fileName)).PageSize(1).Do()
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
	}
	if len(result.Files) != 1 {
		return errors.New("error while searching the targeted file")
	}
	log.Println("Starting download...")
	response, err := service.Files.Get(result.Files[0].Id).Download()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to download: %v", err))
	}
	_, err = os.Stat(filepath.Join(path, result.Files[0].Name))
	if !os.IsNotExist(err) && !saveForce {
		return errors.New(fmt.Sprintf("File already exist: %s", filepath.Join(path, result.Files[0].Name)))
	}
	fp, err := os.Create(filepath.Join(path, result.Files[0].Name))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to make file: %s", filepath.Join(path, result.Files[0].Name)))
	}
	buffer := bufio.NewWriter(fp)
	_, err = buffer.ReadFrom(response.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get data from google drive: %s", filepath.Join(path, result.Files[0].Name)))
	}
	err = fp.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to close file: %s", filepath.Join(path, result.Files[0].Name)))
	}
	log.Println("Ended download...")
	if strings.HasSuffix(result.Files[0].Name, ".tar.xz") && unfreeze {
		log.Println("Starting unfreezing...")
		xzArchiver := archiver.NewTarXz()
		xzArchiver.OverwriteExisting = saveForce
		err = xzArchiver.Unarchive(filepath.Join(path, result.Files[0].Name), filepath.Join(path))
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to unarchive: %s", filepath.Join(path, result.Files[0].Name)))
		}
		if unfreeze {
			err = os.Remove(filepath.Join(path, result.Files[0].Name))
			if err != nil {
				return errors.New(fmt.Sprintf("Failed to delete: %s", filepath.Join(path, result.Files[0].Name)))
			}
		}
		log.Println("Ended unfreezing...")
	} else if strings.HasSuffix(result.Files[0].Name, ".tar.gz") && unfreeze {
		log.Println("Starting unfreezing...")
		gzArchiver := archiver.NewTarGz()
		gzArchiver.SingleThreaded = false
		gzArchiver.OverwriteExisting = saveForce
		err = gzArchiver.Unarchive(filepath.Join(path, result.Files[0].Name), filepath.Join(path, fileName))
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to unarchive: %s", filepath.Join(path, result.Files[0].Name)))
		}
		if unfreeze {
			err := os.Remove(filepath.Join(path, result.Files[0].Name))
			if err != nil {
				return errors.New(fmt.Sprintf("Failed to delete: %s", filepath.Join(path, result.Files[0].Name)))
			}
		}
		log.Println("Ended unfreezing...")
	}
	log.Println("Ended processing")
	return nil
}

func GetFileSize(repository string, directory string, fileName string) (int64, error) {
	service, file, err := getDirectory(repository, directory)
	if err != nil {
		return 0, err
	}
	result, err := service.Files.List().Corpora("teamDrive").IncludeItemsFromAllDrives(true).SupportsTeamDrives(true).TeamDriveId(Repositories[repository]).Q(fmt.Sprintf("'%s' in parents and (name='%s' or name='%s.tar.xz' or name='%s.tar.gz')", file.Id, fileName, fileName, fileName)).PageSize(1).Do()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Unable to list google drive directory: %v", err))
	}
	if len(result.Files) != 1 {
		return 0, errors.New("error while searching the targeted file")
	}
	info, err := service.Files.Get(result.Files[0].Id).SupportsTeamDrives(true).Fields("size").Do()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Unable to get information from google drive: %v", err))
	}
	return info.Size, nil
}
