// Package migrations contains the system PocketBase DB migrations.
package migrations

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/models/settings"
	"github.com/pocketbase/pocketbase/tools/migrate"
	"github.com/pocketbase/pocketbase/tools/types"
)

var AppMigrations migrate.MigrationsList

// Register is a short alias for `AppMigrations.Register()`
// that is usually used in external/user defined migrations.
func Register(
	up func(db dbx.Builder) error,
	down func(db dbx.Builder) error,
	optFilename ...string,
) {
	var optFiles []string
	if len(optFilename) > 0 {
		optFiles = optFilename
	} else {
		_, path, _, _ := runtime.Caller(1)
		optFiles = append(optFiles, filepath.Base(path))
	}
	AppMigrations.Register(up, down, optFiles...)
}

func init() {
	AppMigrations.Register(func(db dbx.Builder) error {
		_, tablesErr := db.NewQuery(`
			create sequence public._admin_id_seq;
			create table _admin
			(
				id varchar default nextval('public._admin_id_seq') not null,
				avatar int default 0 not null,
				email varchar not null,
				"tokenKey" varchar not null,
				"passwordHash" varchar not null,
				"lastResetSentAt" varchar default '' not null,
				created varchar default '' not null,
				updated varchar default '' not null
			);

			create unique index _admin_email_uindex
				on _admin (email);

			create unique index _admin_id_uindex
				on _admin (id);

			create unique index _admin_tokenkey_uindex
				on _admin ("tokenKey");

			alter table _admin
				add constraint _admin_pk
					primary key (id);
			-- end admin

			-- for collection
			create sequence public._collections_id_seq;
			create table _collections
			(
				-- Only integer types can be auto increment
				id varchar default nextval('public._collections_id_seq') not null,
				system bool default false not null,
				type varchar default 'base' not null,
				name varchar not null,
				schema jsonb default '{}'::jsonb not null,
				indexes jsonb default '{}'::jsonb not null,
				"listRule" varchar,
				"viewRule" varchar,
				"createRule" varchar,
				"updateRule" varchar,
				"deleteRule" varchar,
				options jsonb default '{}'::jsonb not null,
				created varchar default '' not null,
				updated varchar default '' not null
			);

			create unique index _collections_id_uindex
				on _collections (id);

			create unique index _collections_name_uindex
				on _collections (name);

			alter table _collections
				add constraint _collections_pk
					primary key (id);
			-- end collection

			-- for params
			create sequence public._params_id_seq;
			create table _params
			(
				-- Only integer types can be auto increment
				id varchar default nextval('public._params_id_seq') not null,
				key varchar not null,
				value jsonb default '{}'::jsonb,
				created varchar default '' not null,
				updated varchar default '' not null
			);

			create unique index _params_id_uindex
				on _params (id);

			create unique index _params_key_uindex
				on _params (key);

			alter table _params
				add constraint _params_pk
					primary key (id);
			-- end params
		`).Execute()
		if tablesErr != nil {
			return tablesErr
		}

		dao := daos.New(db)

		// inserts default settings
		// -----------------------------------------------------------
		defaultSettings := settings.New()
		if err := dao.SaveSettings(defaultSettings); err != nil {
			return err
		}

		// inserts the system profiles collection
		// -----------------------------------------------------------
		usersCollection := &models.Collection{}
		usersCollection.MarkAsNew()
		usersCollection.Id = "_pb_users_auth_"
		usersCollection.Name = "users"
		usersCollection.Type = models.CollectionTypeAuth
		usersCollection.ListRule = types.Pointer("id = @request.auth.id")
		usersCollection.ViewRule = types.Pointer("id = @request.auth.id")
		usersCollection.CreateRule = types.Pointer("")
		usersCollection.UpdateRule = types.Pointer("id = @request.auth.id")
		usersCollection.DeleteRule = types.Pointer("id = @request.auth.id")

		// set auth options
		usersCollection.SetOptions(models.CollectionAuthOptions{
			ManageRule:        nil,
			AllowOAuth2Auth:   true,
			AllowUsernameAuth: true,
			AllowEmailAuth:    true,
			MinPasswordLength: 8,
			RequireEmail:      false,
		})

		// set optional default fields
		usersCollection.Schema = schema.NewSchema(
			&schema.SchemaField{
				Id:      "users_name",
				Type:    schema.FieldTypeText,
				Name:    "name",
				Options: &schema.TextOptions{},
			},
			&schema.SchemaField{
				Id:   "users_avatar",
				Type: schema.FieldTypeFile,
				Name: "avatar",
				Options: &schema.FileOptions{
					MaxSelect: 1,
					MaxSize:   5242880,
					MimeTypes: []string{
						"image/jpeg",
						"image/png",
						"image/svg+xml",
						"image/gif",
						"image/webp",
					},
				},
			},
		)
		return dao.SaveCollection(usersCollection)
	}, func(db dbx.Builder) error {
		tables := []string{
			// "users",
			"_externalAuths",
			"_params",
			"_collections",
			"_admins",
		}
		fmt.Print("masuk sini")
		fmt.Print(tables)
		fmt.Print("masuk sini2")
		for _, name := range tables {
			if _, err := db.DropTable(name).Execute(); err != nil {
				return err
			}
		}

		return nil
	})
}
