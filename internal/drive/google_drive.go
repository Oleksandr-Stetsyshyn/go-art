package drive

import (
	"art/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func getDriveService() (*drive.Service, error) {
	b, err := os.ReadFile("ServiceAccountCred.json")
	if err != nil {
		fmt.Printf("Unable to read ServiceAccountCred.json file. Err: %v\n", err)
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)

	if err != nil {
		return nil, err
	}

	client := getClient(config)

	service, err := drive.New(client)

	if err != nil {
		fmt.Printf("Cannot create the Google Drive service: %v\n", err)
		return nil, err
	}

	return service, err
}

func getClient(config *oauth2.Config) *http.Client {

	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

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

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	fmt.Println("Paste Authrization code here :")
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

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving ServiceAccountCred file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func createFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	file, err := service.Files.Create(f).Media(content).Do()

	if err != nil {
		log.Println("Could not create file: " + err.Error())
		return nil, err
	}

	return file, nil
}

func createFolder(service *drive.Service, name string, parentId string) (*drive.File, error) {
	q := fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and name='%s' and '%s' in parents", name, parentId)
	r, err := service.Files.List().Q(q).Do()
	if err != nil {
		log.Println("Could not find folder: " + err.Error())
		return nil, err
	}

	if len(r.Files) > 0 {
		return r.Files[0], nil
	}

	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentId},
	}

	file, err := service.Files.Create(d).Do()

	if err != nil {
		log.Println("Could not create dir: " + err.Error())
		return nil, err
	}

	return file, nil
}

func UploadImages(id string, path string) (models.Photos, error) {
	// Step 1: Get the Google Drive service
	srv, err := getDriveService()
	if err != nil {
		return models.Photos{}, err
	}

	// Step 2: Create directory
	dir, err := createFolder(srv, id, "1BlTd6XF81bGBXKZvspzQr8sgbPPuYOGx")
	if err != nil {
		return models.Photos{}, fmt.Errorf("Could not create dir: %v", err)
	}

	// Use the ID as the folder ID
	folderId := dir.Id

	var links []string

	// Get the list of files in the tmp/tmpFiles directory
	tmpEntries, err := os.ReadDir(path)
	if err != nil {
		return models.Photos{}, err
	}

	// Step 3: Loop over the image paths and upload each image
	for _, tmpEntry := range tmpEntries {
		// Skip if not a file
		if !tmpEntry.Type().IsRegular() {
			continue
		}

		// Open the file
		f, err := os.Open(filepath.Join(path, tmpEntry.Name()))
		if err != nil {
			return models.Photos{}, err
		}
		defer f.Close()

		// Create the file and upload
		file, err := createFile(srv, filepath.Base(f.Name()), "application/octet-stream", f, folderId)
		if err != nil {
			return models.Photos{}, err
		}

		// After the file is uploaded, make it public
		perm := &drive.Permission{
			Type: "anyone",
			Role: "reader",
		}

		_, err = srv.Permissions.Create(file.Id, perm).Do()
		if err != nil {
			return models.Photos{}, err
		}

		// Now get the file metadata, including the web view link
		file, err = srv.Files.Get(file.Id).Fields("webViewLink").Do()
		if err != nil {
			return models.Photos{}, err
		}

		// Add the web view link to the array of links
		links = append(links, file.WebViewLink)
	}

	photos := models.Photos{
		Urls:     links,
		FolderId: folderId,
	}

	return photos, nil
}

func DeleteFolder(id string) error {
	fmt.Printf("Deleting folder with ID: %s\n", id)
	srv, err := getDriveService()
	if err != nil {
		return err
	}

	// Check if the folder exists
	_, err = srv.Files.Get(id).Do()
	if err != nil {
		return fmt.Errorf("Folder with ID '%s' not found", id)
	}

	// Now we directly attempt to delete the folder using the provided id
	err = srv.Files.Delete(id).Do()
	if err != nil {
		return err
	}

	return nil
}
