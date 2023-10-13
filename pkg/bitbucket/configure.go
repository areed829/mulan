package bitbucket

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func ConfigureBitbucket(setupSsh bool) {
	if setupSsh {
		// _, _, err := createSshKeys()
		// if err != nil {
		// 	panic(err)
		// }
		// addKeyToBitbucket()
		test()
	}
	// var username string
	// var password string

	// prompt := []*survey.Question{
	// 	{
	// 		Name:     "username",
	// 		Prompt:   &survey.Input{Message: "Enter username:"},
	// 		Validate: survey.Required,
	// 	},
	// 	{
	// 		Name:     "password",
	// 		Prompt:   &survey.Password{Message: "Enter password:"},
	// 		Validate: survey.Required,
	// 	},
	// }

	// var answers struct {
	// 	Username string `survey:"username"`
	// 	Password string `survey:"password"`
	// }

	// err := survey.Ask(prompt, &answers)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// username = answers.Username
	// password = answers.Password

	// // use viper to store username and password
	// viper.Set("bitbucket.username", username)
	// viper.Set("bitbucket.password", password)
	// viper.WriteConfig()
}

type SSHKey struct {
	Label string `json:"label"`
	Key   string `json:"key"`
}

type SSHKeyTest struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Key   string `json:"key"`
	URL   string `json:"url"`
}

func test() {
	username := viper.GetString("bitbucket.username")
	password := viper.GetString("bitbucket.password")

	c := bitbucket.NewBasicAuth(username, password)
	profile, err := c.User.Profile()
	if err != nil {
		fmt.Println(err)
		return
	}

	uuid := profile.Uuid
	// urlSafeUuid := url.QueryEscape(uuid)
	requestUrl := fmt.Sprintf("https://api.bitbucket.org/2.0/users/%s/ssh-keys", uuid)

	// Create a new HTTP GET request to the Bitbucket REST endpoint.
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the HTTP request headers.
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s:%s", username, password))

	// Send the HTTP GET request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close the HTTP response body.
	defer resp.Body.Close()

	// Check the HTTP response status code.
	if resp.StatusCode != 200 {
		fmt.Printf("Error getting SSH keys: %s \n", resp.Status)
		return
	}

	// Decode the JSON response body into a slice of SSHKey objects.
	var sshKeys []SSHKeyTest
	err = json.NewDecoder(resp.Body).Decode(&sshKeys)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print all of the SSH keys.
	for _, sshKey := range sshKeys {
		fmt.Println(sshKey)
	}
}

// func addKeyToBitbucket(publickey string) {
// 	username := viper.GetString("bitbucket.username")
// 	password := viper.GetString("bitbucket.password")

// 	c := bitbucket.NewBasicAuth(username, password)
// 	profile, err := c.User.Profile()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	uuid := profile.Uuid

// 	sshKey := SSHKey{
// 		Label: "auto-generated",
// 		Key:   publickey,
// 	}

// 	// Marshal the SSH key object to JSON.
// 	jsonBytes, err := json.Marshal(sshKey)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// Create a new HTTP POST request to the Bitbucket REST endpoint.
// 	urlSafeUuid := url.QueryEscape(uuid)
// 	requestUrl := fmt.Sprintf("https://api.bitbucket.org/2.0/users/%s/ssh-keys", urlSafeUuid)
// 	req, err := http.NewRequest("POST", requestUrl, bytes.NewReader(jsonBytes))
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// Set the HTTP request headers.
// 	req.Header.Set("Authorization", "Basic YOUR_BITBUCKET_USERNAME:YOUR_BITBUCKET_PASSWORD")
// 	req.Header.Set("Content-Type", "application/json")

// 	// Send the HTTP POST request.
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// Close the HTTP response body.
// 	defer resp.Body.Close()

// 	// Check the HTTP response status code.
// 	if resp.StatusCode != 201 {
// 		fmt.Println(fmt.Sprintf("Error creating SSH key: %s", resp.Status))
// 		return
// 	}

// 	// Print the SSH key ID.
// 	fmt.Println(fmt.Sprintf("SSH key created successfully with ID: %s", resp.Header.Get("X-Request-Id")))

// }

func createSshKeys() (string, string, error) {
	// Generate an ssh key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Encode the private key in PEM format
	privateKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Write the private key to a file
	privateKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "test")
	err = os.WriteFile(privateKeyPath, privateKeyBytes, 0600)
	if err != nil {
		return "", "", err
	}

	// Generate the public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	// Encode the public key in OpenSSH format
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	// Write the public key to a file
	publicKeyPath := privateKeyPath + ".pub"
	err = os.WriteFile(publicKeyPath, publicKeyBytes, 0644)
	if err != nil {
		return "", "", err
	}

	publicKeyString := string(publicKeyBytes)
	privateKeyString := string(privateKeyBytes)
	return publicKeyString, privateKeyString, nil
}
