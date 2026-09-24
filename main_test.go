package main

import (
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func testPocketBase(t *testing.T) *pocketbase.PocketBase {
	t.Helper()
	app := newPocketBase(t.TempDir())
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap PocketBase: %v", err)
	}
	if err := app.RunAppMigrations(); err != nil {
		t.Fatalf("run PocketBase migrations: %v", err)
	}
	t.Cleanup(func() {
		if err := app.ClearBootstrap(); err != nil {
			t.Errorf("clear PocketBase bootstrap: %v", err)
		}
	})
	return app
}

func TestLearningGoalsRequireAuthenticatedOwner(t *testing.T) {
	app := testPocketBase(t)
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}
	if users.CreateRule == nil || *users.CreateRule != "" {
		t.Fatal("new testers must be able to register")
	}
	if users.ListRule == nil || *users.ListRule != "id = @request.auth.id" {
		t.Fatalf("users list rule = %v; expected self-only", users.ListRule)
	}
	if users.ViewRule == nil || *users.ViewRule != "id = @request.auth.id" {
		t.Fatalf("users view rule = %v; expected self-only", users.ViewRule)
	}

	goals, err := app.FindCollectionByNameOrId("learning_goals")
	if err != nil {
		t.Fatal(err)
	}
	if goals.ListRule == nil || *goals.ListRule != "user = @request.auth.id" {
		t.Fatalf("goal list rule = %v; expected owner-only", goals.ListRule)
	}
	if goals.ViewRule == nil || *goals.ViewRule != "user = @request.auth.id" {
		t.Fatalf("goal view rule = %v; expected owner-only", goals.ViewRule)
	}
	if goals.CreateRule == nil || *goals.CreateRule != "@request.auth.id != '' && @request.body.user = @request.auth.id" {
		t.Fatalf("goal create rule = %v; expected authenticated owner", goals.CreateRule)
	}
	if goals.UpdateRule == nil || *goals.UpdateRule != "@request.auth.id != '' && user = @request.auth.id && @request.body.user:changed=false" {
		t.Fatalf("goal update rule = %v; ownership must not be transferable", goals.UpdateRule)
	}
	if goals.DeleteRule == nil || *goals.DeleteRule != "user = @request.auth.id" {
		t.Fatalf("goal delete rule = %v; expected owner-only", goals.DeleteRule)
	}
	if goals.Fields.GetByName("path") == nil {
		t.Fatal("learning_goals must persist the configured learning path")
	}
}

func TestEachNewAccountGetsItsOwnInitialLearningSession(t *testing.T) {
	app := testPocketBase(t)
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}

	created := make([]*core.Record, 0, 2)
	for _, email := range []string{"first@example.com", "second@example.com"} {
		user := core.NewRecord(users)
		user.Set("email", email)
		user.Set("password", "Trama-test-password-123!")
		if err := app.Save(user); err != nil {
			t.Fatalf("create test user %q: %v", email, err)
		}
		created = append(created, user)
	}

	goalIDs := make(map[string]bool)
	for _, user := range created {
		goals, err := app.FindRecordsByFilter("learning_goals", "user = {:user}", "created", 10, 0, dbx.Params{"user": user.Id})
		if err != nil {
			t.Fatalf("find initial goal for %s: %v", user.Id, err)
		}
		if len(goals) != 1 {
			t.Fatalf("account %s has %d initial learning sessions, want 1", user.Id, len(goals))
		}
		goal := goals[0]
		if goal.GetString("user") != user.Id {
			t.Fatalf("initial session owner = %q, want %q", goal.GetString("user"), user.Id)
		}
		if goal.GetString("domain") == "" || goal.Get("path") == nil {
			t.Fatalf("initial session for %s is missing its starter path", user.Id)
		}
		if goalIDs[goal.Id] {
			t.Fatalf("accounts share initial session record %s", goal.Id)
		}
		goalIDs[goal.Id] = true
	}
}

func TestDemoEvaluationIsNotPresentedAsAI(t *testing.T) {
	assessment := evaluateAnswer("algebra resolver ecuaciones", "Puedo usar algebra para resolver ecuaciones")
	if assessment.Provider != "mock" || assessment.Model != "heuristic-demo-v1" {
		t.Fatalf("evaluation provenance = %q/%q, want explicit demo markers", assessment.Provider, assessment.Model)
	}
	if assessment.Score <= 0 || assessment.Explanation == "" {
		t.Fatalf("expected a scored demo result, got %#v", assessment)
	}
}

func TestLearningDataPersistsAcrossPocketBaseRestart(t *testing.T) {
	dataDir := t.TempDir()
	first := newPocketBase(dataDir)
	if err := first.Bootstrap(); err != nil {
		t.Fatalf("bootstrap first PocketBase instance: %v", err)
	}
	if err := first.RunAppMigrations(); err != nil {
		t.Fatalf("run first PocketBase migrations: %v", err)
	}
	users, err := first.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}
	user := core.NewRecord(users)
	user.Set("email", "persistence@example.com")
	user.Set("password", "Trama-test-password-123!")
	if err := first.Save(user); err != nil {
		t.Fatalf("create persistent test account: %v", err)
	}
	if err := first.ClearBootstrap(); err != nil {
		t.Fatalf("stop first PocketBase instance: %v", err)
	}

	second := newPocketBase(dataDir)
	if err := second.Bootstrap(); err != nil {
		t.Fatalf("bootstrap restarted PocketBase: %v", err)
	}
	t.Cleanup(func() {
		if err := second.ClearBootstrap(); err != nil {
			t.Errorf("stop restarted PocketBase: %v", err)
		}
	})
	if err := second.RunAppMigrations(); err != nil {
		t.Fatalf("run restart migrations: %v", err)
	}
	goals, err := second.FindRecordsByFilter("learning_goals", "user = {:user}", "created", 10, 0, dbx.Params{"user": user.Id})
	if err != nil {
		t.Fatalf("read persisted initial session: %v", err)
	}
	if len(goals) != 1 || goals[0].GetString("domain") != "Aprender a aprender" {
		t.Fatalf("restart did not preserve the initial learning session: %#v", goals)
	}
}
