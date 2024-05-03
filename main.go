package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	apiURL          = "http://localhost:8080"
	loginEndpoint   = "/login"
	messageEndpoint = "/message"
	fileEndpoint    = "/file"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type UserSession struct {
	Token string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter username: ")
	scanner.Scan()
	username := scanner.Text()

	fmt.Print("Enter password: ")
	scanner.Scan()
	password := scanner.Text()

	credentials := Credentials{
		Username: username,
		Password: password,
	}

	session, err := loginUser(credentials)
	if err != nil {
		fmt.Println("Login failed:", err)
		return
	}

	for {
		fmt.Print("Enter command (message, file, exit): ")
		scanner.Scan()
		command := scanner.Text()

		switch command {
		case "message":
			fmt.Print("Enter recipient: ")
			scanner.Scan()
			recipient := scanner.Text()

			fmt.Print("Enter message: ")
			scanner.Scan()
			message := scanner.Text()

			err := sendMessage(session.Token, Message{Recipient: recipient, Message: message})
			if err != nil {
				fmt.Println("Error sending message:", err)
			} else {
				fmt.Println("Message sent successfully")
			}
		case "file":
			fmt.Print("Enter recipient: ")
			scanner.Scan()
			recipient := scanner.Text()

			fmt.Print("Enter file path: ")
			scanner.Scan()
			filePath := scanner.Text()

			err := sendFile(session.Token, recipient, filePath)
			if err != nil {
				fmt.Println("Error sending file:", err)
			} else {
				fmt.Println("File sent successfully")
			}
		case "exit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func loginUser(creds Credentials) (*UserSession, error) {
	data, err := json.Marshal(creds)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL+loginEndpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var session UserSession
	err = json.Unmarshal(body, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func sendMessage(token string, message Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", apiURL+messageEndpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send message with status: %d", resp.StatusCode)
	}

	return nil
}

func sendFile(token string, recipient, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	err = writer.WriteField("recipient", recipient)
	if err != nil {
		return err
	}
	writer.Close()

	request, err := http.NewRequest("POST", apiURL+fileEndpoint, body)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send file with status: %d", resp.StatusCode)
	}

	return nil
}
