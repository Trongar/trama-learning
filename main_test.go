package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

func TestHomeOffersLearningGoalForm(t *testing.T) {
	application := newApp()
	server := httptest.NewServer(application.Handler())
	defer server.Close()

	response, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("home status = %d, want 200", response.StatusCode)
	}
	for _, marker := range []string{"¿Qué quieres aprender?", "name=\"domain\"", "name=\"purpose\"", "name=\"minutes_per_week\"", "Crear mi ruta"} {
		if !strings.Contains(string(body), marker) {
			t.Errorf("home page missing %q", marker)
		}
	}
}

func TestLearnerCanCreateGoalStudyAndSubmitAnswer(t *testing.T) {
	application := newApp()
	server := httptest.NewServer(application.Handler())
	defer server.Close()

	form := url.Values{
		"domain":           {"álgebra"},
		"level":            {"inicial"},
		"purpose":          {"resolver ecuaciones"},
		"minutes_per_week": {"90"},
	}
	request, err := http.NewRequest(http.MethodPost, server.URL+"/goals", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("create goal status = %d, want 303", response.StatusCode)
	}
	goalURL := response.Header.Get("Location")
	if !strings.HasPrefix(goalURL, "/goals/") {
		t.Fatalf("goal redirect = %q, want /goals/{id}", goalURL)
	}

	goalResponse, err := http.Get(server.URL + goalURL)
	if err != nil {
		t.Fatal(err)
	}
	goalBody, _ := io.ReadAll(goalResponse.Body)
	goalResponse.Body.Close()
	if goalResponse.StatusCode != http.StatusOK {
		t.Fatalf("goal page status = %d, want 200", goalResponse.StatusCode)
	}
	if !strings.Contains(string(goalBody), `data-testid="concept-map"`) {
		t.Fatal("goal page does not render the concept map")
	}

	goalID := strings.TrimPrefix(goalURL, "/goals/")
	goal, ok := application.store.goals[goalID]
	if !ok || len(goal.Concepts) == 0 {
		t.Fatal("created goal or its concepts were not stored")
	}
	conceptURL := "/concepts/" + goal.Concepts[0].ID
	lessonResponse, err := http.Get(server.URL + conceptURL)
	if err != nil {
		t.Fatal(err)
	}
	lessonBody, _ := io.ReadAll(lessonResponse.Body)
	lessonResponse.Body.Close()
	if lessonResponse.StatusCode != http.StatusOK || !strings.Contains(string(lessonBody), `data-testid="answer-form"`) {
		t.Fatalf("lesson did not render answer form: status=%d", lessonResponse.StatusCode)
	}

	answer := url.Values{"answer": {"Una ecuación representa una igualdad para resolver ecuaciones de álgebra."}}
	attemptResponse, err := http.PostForm(server.URL+"/exercises/"+goal.Concepts[0].ExerciseID+"/attempts", answer)
	if err != nil {
		t.Fatal(err)
	}
	attemptBody, _ := io.ReadAll(attemptResponse.Body)
	attemptResponse.Body.Close()
	if attemptResponse.StatusCode != http.StatusOK {
		t.Fatalf("attempt status = %d, want 200: %s", attemptResponse.StatusCode, attemptBody)
	}
	for _, marker := range []string{"data-testid=\"evaluation-result\"", "RESPUESTA EVALUADA", "mock", "data-testid=\"progress\""} {
		if !strings.Contains(string(attemptBody), marker) {
			t.Errorf("attempt result missing %q", marker)
		}
	}
}

func TestGoalFormRejectsInvalidInput(t *testing.T) {
	application := newApp()
	server := httptest.NewServer(application.Handler())
	defer server.Close()

	response, err := http.PostForm(server.URL+"/goals", url.Values{"domain": {""}})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid goal status = %d, want 400", response.StatusCode)
	}
	if !strings.Contains(string(body), "Escribe el tema") {
		t.Fatalf("invalid goal response should explain the error, got %s", body)
	}
}

func TestConcurrentProgressReadsAndUpdatesAreSafe(t *testing.T) {
	application := newApp()
	server := httptest.NewServer(application.Handler())
	defer server.Close()

	form := url.Values{
		"domain":           {"álgebra"},
		"level":            {"inicial"},
		"purpose":          {"resolver ecuaciones"},
		"minutes_per_week": {"90"},
	}
	request, err := http.NewRequest(http.MethodPost, server.URL+"/goals", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	goalID := strings.TrimPrefix(response.Header.Get("Location"), "/goals/")
	response.Body.Close()
	application.mu.RLock()
	goal := application.store.goals[goalID]
	exerciseID := goal.Concepts[0].ExerciseID
	application.mu.RUnlock()

	const requests = 60
	start := make(chan struct{})
	failures := make(chan string, requests*2)
	var workers sync.WaitGroup
	for i := 0; i < requests; i++ {
		workers.Add(2)
		go func() {
			defer workers.Done()
			<-start
			result, err := http.Get(server.URL + "/goals/" + goalID)
			if err != nil {
				failures <- err.Error()
				return
			}
			io.Copy(io.Discard, result.Body)
			result.Body.Close()
			if result.StatusCode != http.StatusOK {
				failures <- "goal read returned " + result.Status
			}
		}()
		go func() {
			defer workers.Done()
			<-start
			result, err := http.PostForm(server.URL+"/exercises/"+exerciseID+"/attempts", url.Values{"answer": {"álgebra resolver ecuaciones"}})
			if err != nil {
				failures <- err.Error()
				return
			}
			io.Copy(io.Discard, result.Body)
			result.Body.Close()
			if result.StatusCode != http.StatusOK {
				failures <- "attempt returned " + result.Status
			}
		}()
	}
	close(start)
	workers.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
}

func TestUserContentIsEscapedInRenderedPages(t *testing.T) {
	application := newApp()
	server := httptest.NewServer(application.Handler())
	defer server.Close()

	form := url.Values{
		"domain":           {`<script>alert("x")</script>`},
		"level":            {"inicial"},
		"purpose":          {"aprender"},
		"minutes_per_week": {"60"},
	}
	response, err := http.PostForm(server.URL+"/goals", form)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	page, err := http.Get(server.URL + response.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(page.Body)
	page.Body.Close()
	if strings.Contains(string(body), "<script>") {
		t.Fatal("user-provided content was rendered as executable HTML")
	}
}
