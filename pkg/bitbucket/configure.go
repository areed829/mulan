package bitbucket

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type SshKeyResponse struct {
	Values  []interface{} `json:"values"`
	Pagelen int           `json:"pagelen"`
	Size    int           `json:"size"`
	Page    int           `json:"page"`
}

type SSHKey struct {
	Label string `json:"label"`
	Key   string `json:"key"`
}

type BitbucketConfigurationSettings struct {
	SetupSsh                 bool
	SetupUsernameAndPassword bool
}

func ConfigureBitbucket(configuration *BitbucketConfigurationSettings) {
	if configuration.SetupUsernameAndPassword {
		setupUsernameAndPassword()
	}
	if configuration.SetupSsh {
		publicKey, _, err := createSshKeys()
		if err != nil {
			panic(err)
		}
		addKeyToBitbucket(publicKey)
		addPrivateKeyToAgent()
	}
}

func setupUsernameAndPassword() {
	var username string
	var password string

	prompt := []*survey.Question{
		{
			Name:     "username",
			Prompt:   &survey.Input{Message: "Enter username:"},
			Validate: survey.Required,
		},
		{
			Name:     "password",
			Prompt:   &survey.Password{Message: "Enter password:"},
			Validate: survey.Required,
		},
	}

	var answers struct {
		Username string `survey:"username"`
		Password string `survey:"password"`
	}

	err := survey.Ask(prompt, &answers)
	if err != nil {
		fmt.Println(err)
		return
	}

	username = answers.Username
	password = answers.Password

	// use viper to store username and password
	viper.Set("bitbucket.username", username)
	viper.Set("bitbucket.password", password)
	viper.WriteConfig()
}

func addKeyToBitbucket(publickey string) {
	username := viper.GetString("bitbucket.username")
	password := viper.GetString("bitbucket.password")

	c := bitbucket.NewBasicAuth(username, password)
	profile, err := c.User.Profile()
	if err != nil {
		fmt.Println(err)
		return
	}

	uuid := profile.Uuid

	sshKey := SSHKey{
		Label: "auto-generated",
		Key:   publickey,
	}

	// Marshal the SSH key object to JSON.
	jsonBytes, err := json.Marshal(sshKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a new HTTP POST request to the Bitbucket REST endpoint.

	requestUrl := fmt.Sprintf("https://api.bitbucket.org/2.0/users/%s/ssh-keys", uuid)
	req, err := http.NewRequest("POST", requestUrl, bytes.NewReader(jsonBytes))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the HTTP request headers.
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encodedAuth))
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP POST request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close the HTTP response body.
	defer resp.Body.Close()

	// Check the HTTP response status code.
	if resp.StatusCode != 201 {
		fmt.Printf("Error creating SSH key: %s\n", resp.Status)
		return
	}

	// Print the SSH key ID.
	fmt.Printf("SSH key created successfully with ID: %s\n", resp.Header.Get("X-Request-Id"))

}

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
	privateKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "bitbucket")
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

func addPrivateKeyToAgent() {
	if !isSSHAgentRunning() {
		startSSHAgent()
	}
	privateKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "bitbucket")
	cmd := exec.Command("ssh-add", privateKeyPath)
	cmd.Env = os.Environ()
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error adding private key to ssh-agent:", err)
	}
}

func startSSHAgent() {
	cmd := exec.Command("ssh-agent", "-s")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error starting ssh-agent:", err)
		return
	}

	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Contains(line, "SSH_AUTH_SOCK") || strings.Contains(line, "SSH_AGENT_PID") {
			parts := strings.Split(line, ";")
			envVar := strings.TrimSpace(parts[0])
			os.Setenv(strings.Split(envVar, "=")[0], strings.Split(envVar, "=")[1])
		}
	}
}

func isSSHAgentRunning() bool {
	// Check for the presence of SSH_AUTH_SOCK environment variable
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return false
	}

	// Run ssh-add -L to check if the agent is operational
	cmd := exec.Command("ssh-add", "-L")
	err := cmd.Run()
	return err == nil
}
