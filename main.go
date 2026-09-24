package main

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "edg-learning-mvp/migrations"
)

//go:embed web/public/*
var publicFiles embed.FS

type Assessment struct {
	Score       float64 `json:"score"`
	Correct     bool    `json:"correct"`
	Explanation string  `json:"explanation"`
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
}

type attemptRequest struct {
	Index  int    `json:"index"`
	Answer string `json:"answer"`
}

func newPocketBase(dataDir string) *pocketbase.PocketBase {
	pb := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dataDir,
		HideStartBanner: true,
	})
	migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{})

	pb.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		if err := createInitialLearningSession(e.App, e.Record.Id); err != nil {
			return err
		}
		return e.Next()
	})

	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		public, err := fsSubPublic()
		if err != nil {
			return err
		}
		e.Router.BindFunc(func(event *core.RequestEvent) error {
			event.Response.Header().Set("Cache-Control", "no-store")
			if !strings.HasPrefix(event.Request.URL.Path, "/_/" ) {
				event.Response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				event.Response.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
			}
			return event.Next()
		})
		e.Router.GET("/healthz", func(event *core.RequestEvent) error {
			return event.String(http.StatusOK, "ok\n")
		})
		e.Router.POST("/api/trama/goals/{id}/attempts", createAttemptRoute).Bind(apis.RequireAuth(), apis.BodyLimit(16*1024))
		e.Router.GET("/{path...}", apis.Static(public, true))
		return e.Next()
	})
	return pb
}

func fsSubPublic() (fs.FS, error) {
	return fs.Sub(publicFiles, "web/public")
}

func createInitialLearningSession(app core.App, userID string) error {
	collection, err := app.FindCollectionByNameOrId("learning_goals")
	if err != nil {
		return err
	}

	goal := core.NewRecord(collection)
	goal.Set("user", userID)
	goal.Set("domain", "Aprender a aprender")
	goal.Set("level", "inicial")
	goal.Set("purpose", "Explorar Trama con una ruta privada de demostración.")
	goal.Set("minutes_per_week", 90)
	goal.Set("status", "active")
	goal.Set("progress", 0)
	goal.Set("path", initialPath("Aprender a aprender", "Explorar Trama con una ruta privada de demostración.", "inicial"))
	return app.Save(goal)
}

func initialPath(domain, purpose, level string) []map[string]any {
	return []map[string]any{
		{
			"name":    "Una meta que importa",
			"summary": "Conecta el tema con algo que sí quieres lograr.",
			"content": "Empieza por tu objetivo: " + purpose + " Observa qué sabes ya y qué te gustaría comprender mejor.",
			"prompt":  "¿Qué te gustaría comprender de " + domain + " y por qué te importa?",
			"mastery": 0.0,
		},
		{
			"name":    "Ideas que se conectan",
			"summary": "Busca relaciones entre conceptos, no solo definiciones.",
			"content": "En el nivel " + level + ", elige dos ideas de " + domain + " y explica cómo una ayuda a entender la otra.",
			"prompt":  "¿Qué dos ideas de " + domain + " se relacionan y cómo?",
			"mastery": 0.0,
		},
		{
			"name":    "Una aplicación práctica",
			"summary": "Lleva lo aprendido a un ejemplo cercano.",
			"content": "Piensa en una situación cotidiana en la que podrías usar " + domain + ". Describe un primer paso y qué aprenderías al intentarlo.",
			"prompt":  "Describe una aplicación práctica de " + domain + " vinculada con tu objetivo.",
			"mastery": 0.0,
		},
	}
}

func createAttemptRoute(e *core.RequestEvent) error {
	var input attemptRequest
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("No pudimos leer tu respuesta.", err)
	}
	input.Answer = strings.TrimSpace(input.Answer)
	if input.Index < 0 || len([]rune(input.Answer)) == 0 || len([]rune(input.Answer)) > 4000 {
		return e.BadRequestError("La respuesta o el paso no son válidos.", nil)
	}

	var assessment Assessment
	var progress float64
	err := e.App.RunInTransaction(func(tx core.App) error {
		goal, err := tx.FindRecordById("learning_goals", e.Request.PathValue("id"))
		if err != nil || goal.GetString("user") != e.Auth.Id {
			return errGoalNotFound
		}
		encodedPath, err := json.Marshal(goal.Get("path"))
		if err != nil {
			return err
		}
		var path []map[string]any
		if err := json.Unmarshal(encodedPath, &path); err != nil {
			return err
		}
		if input.Index >= len(path) {
			return errInvalidPathIndex
		}

		assessment = evaluateAnswer(goal.GetString("domain")+" "+goal.GetString("purpose"), input.Answer)
		path[input.Index]["mastery"] = assessment.Score
		path[input.Index]["last_answer"] = input.Answer
		path[input.Index]["feedback"] = assessment.Explanation
		for _, concept := range path {
			if mastery, ok := concept["mastery"].(float64); ok {
				progress += mastery
			}
		}
		progress /= float64(len(path))
		goal.Set("path", path)
		goal.Set("progress", progress)
		return tx.Save(goal)
	})
	if errors.Is(err, errGoalNotFound) {
		return e.NotFoundError("Learning session not found", err)
	}
	if errors.Is(err, errInvalidPathIndex) {
		return e.BadRequestError("The requested learning step does not exist.", err)
	}
	if err != nil {
		return e.InternalServerError("Could not save this attempt.", err)
	}
	return e.JSON(http.StatusOK, map[string]any{
		"assessment": assessment,
		"progress":   progress,
	})
}

var (
	errGoalNotFound     = errors.New("learning session not found")
	errInvalidPathIndex = errors.New("learning step index is invalid")
)

func evaluateAnswer(expected, answer string) Assessment {
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
	for _, word := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		result[word] = true
	}
	return result
}

func main() {
	dataDir := os.Getenv("TRAMA_DATA_DIR")
	if dataDir == "" {
		dataDir = "./pb_data"
	}
	if err := newPocketBase(dataDir).Start(); err != nil {
		log.Fatal(err)
	}
}
