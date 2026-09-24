package main

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

//go:embed web/*
var webFiles embed.FS

type Assessment struct {
	Score       float64
	Correct     bool
	Explanation string
	Provider    string
	Model       string
}

type Exercise struct {
	ID        string
	ConceptID string
	Prompt    string
	Expected  string
}

type Concept struct {
	ID         string
	Name       string
	Summary    string
	Content    string
	ExerciseID string
	Mastery    float64
}

type Goal struct {
	ID             string
	Domain         string
	Level          string
	Purpose        string
	MinutesPerWeek int
	Progress       float64
	Concepts       []Concept
}

type pageData struct {
	Title      string
	Error      string
	Domain     string
	Purpose    string
	Level      string
	Minutes    int
	Goal       Goal
	Concept    Concept
	Exercise   Exercise
	Assessment Assessment
	NextURL    string
}

type memoryStore struct {
	goals     map[string]Goal
	exercises map[string]Exercise
}

type app struct {
	mu        sync.RWMutex
	store     memoryStore
	templates *template.Template
}

func newApp() *app {
	views := template.Must(template.New("views").Funcs(template.FuncMap{
		"mul": func(value float64, factor int) float64 { return value * float64(factor) },
		"add": func(value, delta int) int { return value + delta },
	}).ParseFS(webFiles, "web/*.html"))
	return &app{
		store:     memoryStore{goals: make(map[string]Goal), exercises: make(map[string]Exercise)},
		templates: views,
	}
}

func (a *app) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.home)
	mux.HandleFunc("POST /goals", a.createGoal)
	mux.HandleFunc("GET /goals/{id}", a.showGoal)
	mux.HandleFunc("GET /concepts/{id}", a.showLesson)
	mux.HandleFunc("POST /exercises/{id}/attempts", a.createAttempt)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	static, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self' https://unpkg.com; connect-src 'self'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func (a *app) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.render(w, http.StatusOK, "home.html", pageData{Title: "Aprende a tu manera", Level: "inicial", Minutes: 90})
}

func (a *app) createGoal(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if err := r.ParseForm(); err != nil {
		a.render(w, http.StatusBadRequest, "home.html", pageData{Title: "Aprende a tu manera", Error: "No pudimos leer el formulario. Inténtalo de nuevo."})
		return
	}
	domain := strings.TrimSpace(r.FormValue("domain"))
	level := strings.TrimSpace(r.FormValue("level"))
	purpose := strings.TrimSpace(r.FormValue("purpose"))
	minutes, err := strconv.Atoi(r.FormValue("minutes_per_week"))
	if domain == "" || len([]rune(domain)) > 100 {
		a.render(w, http.StatusBadRequest, "home.html", pageData{Title: "Aprende a tu manera", Error: "Escribe el tema que quieres aprender (máximo 100 caracteres).", Domain: domain, Purpose: purpose, Level: level})
		return
	}
	if purpose == "" || len([]rune(purpose)) > 240 {
		a.render(w, http.StatusBadRequest, "home.html", pageData{Title: "Aprende a tu manera", Error: "Cuéntanos para qué quieres aprenderlo (máximo 240 caracteres).", Domain: domain, Purpose: purpose, Level: level})
		return
	}
	if level != "inicial" && level != "intermedio" && level != "avanzado" {
		a.render(w, http.StatusBadRequest, "home.html", pageData{Title: "Aprende a tu manera", Error: "Elige un nivel disponible.", Domain: domain, Purpose: purpose, Level: "inicial"})
		return
	}
	if err != nil || minutes < 10 || minutes > 10080 {
		a.render(w, http.StatusBadRequest, "home.html", pageData{Title: "Aprende a tu manera", Error: "Indica entre 10 y 10,080 minutos disponibles por semana.", Domain: domain, Purpose: purpose, Level: level})
		return
	}

	goalID, err := newID("goal")
	if err != nil {
		http.Error(w, "No se pudo crear el objetivo.", http.StatusInternalServerError)
		return
	}
	concepts := make([]Concept, 3)
	labels := []string{"Fundamentos de " + domain, "Ideas clave", "Aplicación práctica"}
	summaries := []string{"Ubica lo que ya sabes y construye una base.", "Conecta las ideas importantes del tema.", "Usa lo aprendido para acercarte a tu objetivo."}
	for i := range concepts {
		conceptID, idErr := newID("concept")
		exerciseID, exerciseErr := newID("exercise")
		if idErr != nil || exerciseErr != nil {
			http.Error(w, "No se pudo crear la ruta.", http.StatusInternalServerError)
			return
		}
		concepts[i] = Concept{
			ID: conceptID, Name: labels[i], Summary: summaries[i],
			Content:    fmt.Sprintf("Esta lección de demostración te invita a explorar %s desde tu nivel %s.\n\nPiensa qué conceptos ya reconoces, cómo se relacionan y qué ejemplo real te ayudaría a avanzar hacia: %s. El contenido específico del tema se conectará con el proveedor de aprendizaje en una siguiente fase.", domain, level, purpose),
			ExerciseID: exerciseID,
		}
	}
	goal := Goal{ID: goalID, Domain: domain, Level: level, Purpose: purpose, MinutesPerWeek: minutes, Concepts: concepts}
	a.mu.Lock()
	a.store.goals[goalID] = goal
	for _, concept := range concepts {
		a.store.exercises[concept.ExerciseID] = Exercise{
			ID: concept.ExerciseID, ConceptID: concept.ID,
			Prompt:   fmt.Sprintf("Explica una idea importante de %s y cómo te acerca a tu objetivo: %s.", domain, purpose),
			Expected: domain + " " + purpose,
		}
	}
	a.mu.Unlock()
	http.Redirect(w, r, "/goals/"+goalID, http.StatusSeeOther)
}

func (a *app) showGoal(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	goal, ok := a.store.goals[r.PathValue("id")]
	if ok {
		goal.Concepts = append([]Concept(nil), goal.Concepts...)
	}
	a.mu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	a.render(w, http.StatusOK, "goal.html", pageData{Title: goal.Domain, Goal: goal})
}

func (a *app) showLesson(w http.ResponseWriter, r *http.Request) {
	conceptID := r.PathValue("id")
	a.mu.RLock()
	var goal Goal
	var concept Concept
	found := false
	for _, candidate := range a.store.goals {
		for _, item := range candidate.Concepts {
			if item.ID == conceptID {
				goal, concept, found = candidate, item, true
				break
			}
		}
		if found {
			break
		}
	}
	if found {
		goal.Concepts = append([]Concept(nil), goal.Concepts...)
	}
	exercise := a.store.exercises[concept.ExerciseID]
	a.mu.RUnlock()
	if !found {
		http.NotFound(w, r)
		return
	}
	a.render(w, http.StatusOK, "lesson.html", pageData{Title: concept.Name, Goal: goal, Concept: concept, Exercise: exercise})
}

func (a *app) createAttempt(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "No pudimos leer tu respuesta.", http.StatusBadRequest)
		return
	}
	answer := strings.TrimSpace(r.FormValue("answer"))
	if answer == "" || len([]rune(answer)) > 4000 {
		http.Error(w, "Escribe una respuesta de hasta 4,000 caracteres.", http.StatusBadRequest)
		return
	}
	a.mu.Lock()
	exercise, ok := a.store.exercises[r.PathValue("id")]
	if !ok {
		a.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	assessment := (mockEvaluator{}).evaluate(exercise.Expected, answer)
	var goal Goal
	var concept Concept
	for goalID, candidate := range a.store.goals {
		for i, item := range candidate.Concepts {
			if item.ID == exercise.ConceptID {
				item.Mastery = assessment.Score
				candidate.Concepts[i] = item
				candidate.Progress = 0
				for _, c := range candidate.Concepts {
					candidate.Progress += c.Mastery
				}
				candidate.Progress /= float64(len(candidate.Concepts))
				a.store.goals[goalID] = candidate
				goal, concept = candidate, item
				break
			}
		}
	}
	a.mu.Unlock()
	if concept.ID == "" {
		http.NotFound(w, r)
		return
	}
	a.render(w, http.StatusOK, "result.html", pageData{Title: "Tu respuesta", Goal: goal, Concept: concept, Assessment: assessment, NextURL: "/goals/" + goal.ID})
}

func (a *app) render(w http.ResponseWriter, status int, name string, data pageData) {
	var output bytes.Buffer
	if err := a.templates.ExecuteTemplate(&output, name, data); err != nil {
		log.Printf("render %s failed: %v", name, err)
		http.Error(w, "No se pudo mostrar esta página.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}

func newID(prefix string) (string, error) {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(raw[:]), nil
}

type mockEvaluator struct{}

func (mockEvaluator) evaluate(expected, answer string) Assessment {
	expectedWords := words(expected)
	answerWords := words(answer)
	matches := 0
	for word := range expectedWords {
		if len([]rune(word)) >= 3 && answerWords[word] {
			matches++
		}
	}
	threshold := 2
	if len(expectedWords) < threshold {
		threshold = len(expectedWords)
	}
	correct := matches >= threshold && threshold > 0
	assessment := Assessment{Correct: correct, Provider: "mock", Model: "heuristic-demo-v1", Score: 0.25}
	if correct {
		assessment.Score = 0.9
		assessment.Explanation = "Buen comienzo: conectaste el tema con tu objetivo. Esta evaluación es una demostración heurística, no una revisión de IA."
	} else {
		assessment.Explanation = "En esta demostración intenta relacionar el tema con el propósito que elegiste. La evaluación es heurística y no equivale a una revisión de IA."
	}
	return assessment
}

func words(text string) map[string]bool {
	result := make(map[string]bool)
	for _, word := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }) {
		result[word] = true
	}
	return result
}

func main() {
	address := os.Getenv("EDG_ADDR")
	if address == "" {
		address = ":8080"
	}
	server := &http.Server{Addr: address, Handler: newApp().Handler(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Trama listening on %s", address)
	log.Fatal(server.ListenAndServe())
}
