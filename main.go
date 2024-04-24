package main

import (
    "fmt"
    "io"
    "net/http"
    "encoding/json"
    "bytes"
    "os"
    "github.com/joho/godotenv"
)

type EmailResponse struct{
  Error bool `json:"error"`
  Message string `json:"message"`
}

func SendEmail(email string, info string, title string) EmailResponse {
  apiURL := os.Getenv("RAPID_API_URL")
  apiKey:= os.Getenv("RAPID_API_KEY")
  sendTo := os.Getenv("SEND_TO")


  fmt.Println(apiURL)
  fmt.Println(apiKey)
  fmt.Println(sendTo)

  fmt.Println(email)
  fmt.Println(info)
  fmt.Println(title)

	var emailResp EmailResponse

	payload := bytes.NewBufferString(fmt.Sprintf(`{
		"sendto": "%s",
		"name": "Magic Media", 
		"replyTo": "%s",
		"ishtml": false,
		"title": %s,
		"body": %s 
	}`, sendTo, email, title, info))

  fmt.Println(payload)

	req, err := http.NewRequest("POST", apiURL, payload)
	if err != nil {
		return EmailResponse{Error: true, Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RapidAPI-Key", apiKey)
	req.Header.Set("X-RapidAPI-Host", "mail-sender-api1.p.rapidapi.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return EmailResponse{Error: true, Message: fmt.Sprintf("failed to send email: %v", err)}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EmailResponse{Error: true, Message: fmt.Sprintf("failed to read response: %v", err)}
	}

  fmt.Println(string(body))

	err = json.Unmarshal(body, &emailResp)
	if err != nil {
    return EmailResponse{Error: true, Message: fmt.Sprintf("failed to unmarshal response: %v", err)}
	}

	return emailResp 
}

func CorsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

        if r.Method == "OPTIONS" {
            return
        }

        next.ServeHTTP(w, r)
    })
}

func postEmail(w http.ResponseWriter, r *http.Request) {

  auth := r.Header.Get("Authorization")
  if auth != os.Getenv("AUTH_TOKEN") {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  requestData := struct {
    Email string `json:"email"`
    Title string `json:"title"`
    Info string `json:"info"`
  }{}

  err := json.NewDecoder(r.Body).Decode(&requestData)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  emailResp := SendEmail(requestData.Email, requestData.Info, requestData.Title)
  if emailResp.Error {
    w.WriteHeader(http.StatusInternalServerError)
  } else {
    w.WriteHeader(http.StatusOK)
  }
  json.NewEncoder(w).Encode(emailResp)
}

func main() {
  err := godotenv.Load()
  if err != nil {
    fmt.Println("Error loading .env file")
    v := os.Getenv("RAPID_API_URL")
    fmt.Println(v)
  }

  mux := http.NewServeMux()
  mux.HandleFunc("/email", postEmail)
  handler := CorsMiddleware(mux)

  port := os.Getenv("PORT")
  if port != "" {
    fmt.Println("Listening on port " + port)
    http.ListenAndServe(":" + port, handler)
  } else {
    fmt.Println("Listening on port 8080")
    http.ListenAndServe(":8080", handler)
  }
}

