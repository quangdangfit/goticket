package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/quangdangfit/goticket/internal/user"
	"github.com/quangdangfit/goticket/internal/user/dto"
	"github.com/quangdangfit/goticket/internal/user/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

// Hand-written fake repo to avoid an import cycle between the service
// package and an external mock package that imports user.Repository.
type fakeRepo struct {
	users    map[string]*model.User
	byEmail  map[string]string
	refresh  map[string]*model.RefreshToken
	storeErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:   map[string]*model.User{},
		byEmail: map[string]string{},
		refresh: map[string]*model.RefreshToken{},
	}
}

func (f *fakeRepo) Create(_ context.Context, u *model.User) error {
	if f.storeErr != nil {
		return f.storeErr
	}
	f.users[u.ID] = u
	f.byEmail[u.Email] = u.ID
	return nil
}
func (f *fakeRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	id, ok := f.byEmail[email]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return f.users[id], nil
}
func (f *fakeRepo) GetByID(_ context.Context, id string) (*model.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return u, nil
}
func (f *fakeRepo) StoreRefresh(_ context.Context, t *model.RefreshToken) error {
	f.refresh[t.TokenHash] = t
	return nil
}
func (f *fakeRepo) GetRefresh(_ context.Context, h string) (*model.RefreshToken, error) {
	t, ok := f.refresh[h]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return t, nil
}
func (f *fakeRepo) RevokeRefresh(_ context.Context, h string, at time.Time) error {
	if t, ok := f.refresh[h]; ok {
		t.RevokedAt = &at
	}
	return nil
}

// Fake JWT manager — deterministic, monotonically unique outputs.
type fakeJWT struct {
	now time.Time
	seq int
}

func (j *fakeJWT) IssueAccess(uid, role string) (string, time.Time, error) {
	j.seq++
	return fmt.Sprintf("access-%s-%s-%d", uid, role, j.seq), j.now.Add(time.Minute), nil
}
func (j *fakeJWT) IssueRefresh(uid string) (string, string, time.Time, error) {
	j.seq++
	tok := fmt.Sprintf("refresh-%s-%d", uid, j.seq)
	return tok, j.HashRefresh(tok), j.now.Add(time.Hour), nil
}
func (j *fakeJWT) Verify(string) (string, string, error) { return "", "", errors.New("nope") }
func (j *fakeJWT) HashRefresh(t string) string           { return "hash:" + t }

func newSvc(t *testing.T) (user.Service, *fakeRepo) {
	t.Helper()
	r := newFakeRepo()
	return New(r, &fakeJWT{now: time.Now().UTC()}), r
}

func TestUserService_Register_NewUser_ReturnsTokens(t *testing.T) {
	svc, repo := newSvc(t)
	out, err := svc.Register(context.Background(), dto.RegisterInput{
		Email: "a@b.com", Password: "password1", Name: "Alice",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("missing tokens: %+v", out)
	}
	if out.User.Email != "a@b.com" || out.User.Role != model.RoleUser {
		t.Fatalf("bad profile: %+v", out.User)
	}
	if len(repo.users) != 1 || len(repo.refresh) != 1 {
		t.Fatalf("repo state: users=%d refresh=%d", len(repo.users), len(repo.refresh))
	}
}

func TestUserService_Register_DuplicateEmail_ReturnsConflict(t *testing.T) {
	svc, _ := newSvc(t)
	in := dto.RegisterInput{Email: "a@b.com", Password: "password1", Name: "A"}
	if _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Register(context.Background(), in)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
}

func TestUserService_Login_WrongPassword_Unauthorized(t *testing.T) {
	svc, _ := newSvc(t)
	_, _ = svc.Register(context.Background(), dto.RegisterInput{
		Email: "a@b.com", Password: "password1", Name: "A",
	})
	_, err := svc.Login(context.Background(), dto.LoginInput{Email: "a@b.com", Password: "nope1234"})
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("want unauthorized, got %v", err)
	}
}

func TestUserService_Login_OK(t *testing.T) {
	svc, _ := newSvc(t)
	_, _ = svc.Register(context.Background(), dto.RegisterInput{
		Email: "a@b.com", Password: "password1", Name: "A",
	})
	out, err := svc.Login(context.Background(), dto.LoginInput{Email: "a@b.com", Password: "password1"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if out.AccessToken == "" {
		t.Fatal("missing access token")
	}
}

func TestUserService_Refresh_RotatesAndRevokesOld(t *testing.T) {
	svc, repo := newSvc(t)
	out, _ := svc.Register(context.Background(), dto.RegisterInput{
		Email: "a@b.com", Password: "password1", Name: "A",
	})
	first := out.RefreshToken
	out2, err := svc.Refresh(context.Background(), first)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if out2.RefreshToken == "" {
		t.Fatal("missing new refresh")
	}
	// old token must now be revoked
	old := repo.refresh["hash:"+first]
	if old == nil || old.RevokedAt == nil {
		t.Fatalf("old refresh not revoked: %+v", old)
	}
}

func TestUserService_Profile_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.Profile(context.Background(), "missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
}
