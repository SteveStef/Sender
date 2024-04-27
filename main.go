package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "os"
  "strings"
  "time"
  "github.com/joho/godotenv"
)

var ValidDomains = []string{"http://localhost:3000", "https://magic-media.org"}

type EmailPayload struct {
  SendTo   string `json:"sendto"`
  Name     string `json:"name"`
  ReplyTo string `json:"replyTo"`
  IsHTML   bool   `json:"ishtml"`
  Title    string `json:"title"`
  Body     string `json:"body"`
}

var httpClient = &http.Client{
  Timeout: 20 * time.Second, // Set a timeout for requests
}

func SendEmail(name string, email string, info string, title string) error {
  apiURL := os.Getenv("RAPID_API_URL")
  apiKey := os.Getenv("RAPID_API_KEY")
  sendTo := os.Getenv("SEND_TO")

  payload := EmailPayload{
    SendTo:   sendTo,
    Name:     name,
    ReplyTo: email,
    IsHTML:   false,
    Title:    title,
    Body:     info,
  }

  payloadBytes, err := json.Marshal(payload)
  if err != nil {
    return err
  }

  req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
  if err != nil {
    return err
  }

  req.Header.Add("content-type", "application/json")
  req.Header.Add("X-RapidAPI-Key", apiKey)
  req.Header.Add("X-RapidAPI-Host", "mail-sender-api1.p.rapidapi.com")

  res, err := httpClient.Do(req)
  if err != nil {
    return err
  }
  defer res.Body.Close()

  return nil
}

func CorsMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    domain := r.Header.Get("Origin")
    if domain != "" {
      for _, validDomain := range ValidDomains {
        if domain == validDomain {
          w.Header().Set("Access-Control-Allow-Origin", domain)
          break
        }
      }
    }

    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-Control-Allow-Credentials", "true")

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
    Name string `json:"name"`
  }{}

  err := json.NewDecoder(r.Body).Decode(&requestData)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  error := SendEmail(requestData.Name, requestData.Email, requestData.Info, requestData.Title)

  if error != nil {
    fmt.Println(error)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
}

func main() {
  godotenv.Load()

  requiredEnvVars := []string{"RAPID_API_URL", "RAPID_API_KEY", "SEND_TO", "AUTH_TOKEN", "PORT"}
  for _, envVar := range requiredEnvVars {
    if os.Getenv(envVar) == "" {
      fmt.Printf("Environment variable %s is not set\n", envVar)
      os.Exit(1)
    }
  }

  mux := http.NewServeMux()
  mux.HandleFunc("/email", postEmail)
  handler := CorsMiddleware(mux)

  port := os.Getenv("PORT")
  fmt.Println("Listening on port " + port)
  http.ListenAndServe(":"+port, handler)
}
