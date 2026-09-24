package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.CreateRule = types.Pointer("")
		users.AuthRule = types.Pointer("")
		users.ListRule = types.Pointer("id = @request.auth.id")
		users.ViewRule = types.Pointer("id = @request.auth.id")
		users.UpdateRule = types.Pointer("id = @request.auth.id")
		users.DeleteRule = types.Pointer("id = @request.auth.id")
		if err := app.Save(users); err != nil {
			return err
		}

		goals := core.NewBaseCollection("learning_goals")
		goals.ListRule = types.Pointer("user = @request.auth.id")
		goals.ViewRule = types.Pointer("user = @request.auth.id")
		goals.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
		goals.UpdateRule = types.Pointer("@request.auth.id != '' && user = @request.auth.id && @request.body.user:changed=false")
		goals.DeleteRule = types.Pointer("user = @request.auth.id")

		minimumMinutes, maximumMinutes := 10.0, 10080.0
		minimumProgress, maximumProgress := 0.0, 1.0
		goals.Fields.Add(
			&core.RelationField{Name: "user", Required: true, MaxSelect: 1, CascadeDelete: true, CollectionId: users.Id},
			&core.TextField{Name: "domain", Required: true, Max: 100},
			&core.SelectField{Name: "level", Required: true, Values: []string{"inicial", "intermedio", "avanzado"}},
			&core.TextField{Name: "purpose", Required: true, Max: 240},
			&core.NumberField{Name: "minutes_per_week", Required: true, OnlyInt: true, Min: &minimumMinutes, Max: &maximumMinutes},
			&core.SelectField{Name: "status", Required: true, Values: []string{"active", "completed"}},
			&core.NumberField{Name: "progress", Min: &minimumProgress, Max: &maximumProgress},
			&core.JSONField{Name: "path", Required: true, MaxSize: 65536},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		goals.AddIndex("idx_learning_goals_user", false, "user", "")
		return app.Save(goals)
	}, func(app core.App) error {
		goals, err := app.FindCollectionByNameOrId("learning_goals")
		if err == nil {
			if err := app.Delete(goals); err != nil {
				return err
			}
		}
		return nil
	})
}
