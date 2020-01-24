package gorm

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/slim-crown/issue-1-website/internal/services/session"
)

func setUpGormDB(t *testing.T) *gorm.DB {
	const (
		host     = "localhost"
		port     = "5432"
		dbname   = "issue#1website"
		role     = "postgres"
		password = "password1234!@#$"
	)
	dataSourceName := fmt.Sprintf(
		`host=%s port=%s dbname='%s' user='%s' password='%s' sslmode=disable`,
		host, port, dbname, role, password)

	db, err := gorm.Open("postgres", dataSourceName)
	if err != nil {
		t.Fatalf("database connection failed because: %s", err.Error())
	}
	{
		// if !db.HasTable(&session.Session{}) || !db.HasTable(&session.MapPair{}) {
		errs := db.AutoMigrate(&session.Session{}, &session.MapPair{}).GetErrors()
		if len(errs) > 0 {
			log.Fatalf("migration of session failed becauses: %+v", errs)
		}
		// }
	}
	return db
}

func TestSessionGormRepo(t *testing.T) {
	db := setUpGormDB(t)
	defer db.Close()
	repo := &sessionRepo{conn: db}
	sess := &session.Session{
		UUID:           "specialTestUUID01234567890ABCDEF",
		Expires:        time.Now().Add(time.Hour),
		LastAccessTime: time.Now(),
		Data:           make([]session.MapPair, 0),
	}

	t.Run("AddSession", func(t *testing.T) {
		type args struct {
			s *session.Session
		}
		tests := []struct {
			name  string
			repo  *sessionRepo
			args  args
			want  *session.Session
			want1 []error
		}{
			{
				"Success",
				repo,
				args{sess},
				sess,
				make([]error, 0),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, got1 := tt.repo.AddSession(tt.args.s)
				if len(got1) != len(tt.want1) {
					t.Errorf("got1 = %v, want %v", got1, tt.want1)
				}
				if !(got.UUID == tt.want.UUID) {
					t.Errorf("got = %v, want %v", got, tt.want)
				}
			})
		}
	})

	sess.Data = append(sess.Data,
		session.MapPair{
			Key:   "test",
			Value: "testing",
		})

	t.Run("UpdateSession", func(t *testing.T) {
		type args struct {
			s *session.Session
		}
		tests := []struct {
			name  string
			repo  *sessionRepo
			args  args
			want  *session.Session
			want1 []error
		}{
			{
				"Success",
				repo,
				args{sess},
				sess,
				make([]error, 0),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, got1 := tt.repo.UpdateSession(tt.args.s)
				if len(got1) != len(tt.want1) {
					t.Errorf("got1 = %v, want %v", got1, tt.want1)
				}
				if got.UUID != tt.want.UUID || !reflect.DeepEqual(got.Data, tt.want.Data) {
					t.Errorf("got = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("GetSession", func(t *testing.T) {
		type args struct {
			sessionID string
		}
		tests := []struct {
			name  string
			repo  *sessionRepo
			args  args
			want  *session.Session
			want1 []error
		}{
			{
				"Success",
				repo,
				args{sess.UUID},
				sess,
				make([]error, 0),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, got1 := tt.repo.GetSession(tt.args.sessionID)
				if len(got1) != len(tt.want1) {
					t.Errorf("\ngot1 = %v, \nwant %v", got1, tt.want1)
				}
				if got.UUID != tt.want.UUID || !reflect.DeepEqual((got.Data), (tt.want.Data)) {
					t.Errorf("\ngot = %v \nwant %v", got, tt.want)
				}
			})
		}
	})

	t.Run("DeleteSession", func(t *testing.T) {
		// t.Skip("temp skip")
		type args struct {
			sessionID string
		}
		tests := []struct {
			name  string
			repo  *sessionRepo
			args  args
			want  *session.Session
			want1 []error
		}{
			{
				"Success",
				repo,
				args{sess.UUID},
				sess,
				make([]error, 0),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, got1 := tt.repo.DeleteSession(tt.args.sessionID)
				if len(got1) != len(tt.want1) {
					t.Errorf("got1 = %v, want %v", got1, tt.want1)
				}
				if !(got.UUID == tt.want.UUID) {
					t.Errorf("got = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
