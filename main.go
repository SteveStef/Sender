package main

import (
  "fmt"
  "io"
  "net/http"
  "encoding/json"
  "strings"
  "os"
  "github.com/joho/godotenv"
)

func SendEmail(email string, info string, title string) error {
  apiURL := os.Getenv("RAPID_API_URL")
  apiKey:= os.Getenv("RAPID_API_KEY")
  sendTo := os.Getenv("SEND_TO")

  payload := strings.NewReader("{\n    \"sendto\": \"" + sendTo + "\",\n    \"name\": \"Magic Media\",\n    \"replyTo\": \"" + email + "\",\n    \"ishtml\": \"false\",\n    \"title\": \"" + title + "\",\n    \"body\": \"" + info + "\"\n}")

  req, err := http.NewRequest("POST", apiURL, payload)
  if err != nil {
    return err
  }

	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-RapidAPI-Key", apiKey)
	req.Header.Add("X-RapidAPI-Host", "mail-sender-api1.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
  if err != nil {
    return err
  }

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
  if err != nil {
    return err
  }

	fmt.Println(res)
	fmt.Println(string(body))

	return nil;
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

  error := SendEmail(requestData.Email, requestData.Info, requestData.Title)
  if error != nil {
    fmt.Println(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
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

