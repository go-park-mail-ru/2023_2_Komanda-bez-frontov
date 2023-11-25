package api

import (
 "fmt"
 "io/ioutil"
 "io"
 "encoding/json"
 "net/http"
 "go-form-hub/csat_service/repository"
)

type CSATHandler struct {
	csat_service *repository.CSATService
}

type RequestADD struct {
	Rating int `json:"rating"`
}

type User struct {
	ID int64 `json:"id"`
}

type ResponseUser struct {
	User User `json:"current_user"`
}


   

func NewCSATHandler(csat_service *repository.CSATService) *CSATHandler {
 return &CSATHandler{
	csat_service: csat_service,
 }
}

// CheckCSAT обрабатывает запросы по пути /api/v1/csat/check
func (handler *CSATHandler) CheckCSAT(w http.ResponseWriter, r *http.Request) {
	url := "http://localhost:8080/api/v1/is_authorized"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
	 fmt.Println("Error creating the request:", err)
	 return
	}

	sessionID, err := r.Cookie("session_id")
	if err != nil {
		result := map[string]interface{}{
			"render": false,
		}
		respondWithJSON(w, http.StatusInternalServerError, result)
		return
	}
   
	cookie := http.Cookie{Name: "session_id", Value: sessionID.Value}
	req.AddCookie(&cookie)
   
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
	 fmt.Println("Error making the request:", err)
	 return
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)

	if resp.Status != "200 OK" {
		result := map[string]interface{}{
			"render": false,
		}
		respondWithJSON(w, http.StatusOK, result)
		return
	}
   
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	 fmt.Println("Error reading the response:", err)
	 return
	}
   
	var user ResponseUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println("Error unmarshalling the JSON:", err)
		return
	}

	fmt.Println(user.User.ID)

	if ok, _ := handler.csat_service.CheckPassageByUserID(user.User.ID); ok {
		result := map[string]interface{}{
			"render": false,
		}
		respondWithJSON(w, http.StatusOK, result)
		return
	}
   
	fmt.Println("Data from the response:", user.User.ID)

	result := map[string]interface{}{
		"render": true,
	}
	respondWithJSON(w, http.StatusOK, result)
}

// AddCSAT обрабатывает запросы по пути /api/v1/csat/add
func (handler *CSATHandler) AddCSAT(w http.ResponseWriter, r *http.Request) {
	requestJSON, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	var request RequestADD 
	if err := json.Unmarshal(requestJSON, &request); err != nil {
		fmt.Println("unmarshal err: %e", err)
		respondWithJSON(w, http.StatusInternalServerError, "")
		return
	}

	// AAAAAAAAAAAAA
	url := "http://localhost:8080/api/v1/is_authorized"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
	 fmt.Println("Error creating the request:", err)
	 return
	}

	sessionID, err := r.Cookie("session_id")
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "")
		return
	}
   
	cookie := http.Cookie{Name: "session_id", Value: sessionID.Value}
	req.AddCookie(&cookie)
   
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
	 fmt.Println("Error making the request:", err)
	 return
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		result := map[string]interface{}{
			"render": false,
		}
		respondWithJSON(w, http.StatusOK, result)
		return
	}
   
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	 fmt.Println("Error reading the response:", err)
	 return
	}
   
	var user ResponseUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println("Error unmarshalling the JSON:", err)
		return
	}

	err = handler.csat_service.AddPassage(user.User.ID, request.Rating)


	if err != nil {
		respondWithJSON(w, http.StatusConflict, "")
		return
	}

	respondWithJSON(w, http.StatusOK, "")
}

// ResultsCSAT обрабатывает запросы по пути /api/v1/csat/results
func (handler *CSATHandler) ResultsCSAT(w http.ResponseWriter, r *http.Request) {
	results, err := handler.csat_service.Results(); 

	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "")
		return
	}

	respondWithJSON(w, http.StatusOK, results)
}

// respondWithJSON добавляет заголовки и отправляет JSON ответ
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
 response, err := json.Marshal(payload)
 if err != nil {
  w.WriteHeader(http.StatusInternalServerError)
  w.Write([]byte("Internal Server Error"))
  return
 }

 w.Header().Set("Content-Type", "application/json")
 w.WriteHeader(status)
 w.Write(response)
}