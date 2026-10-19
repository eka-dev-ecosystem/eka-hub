package usermodule

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	customErrors "golang/pkg/errors"

	"github.com/google/uuid"
)

var (
	u1 = uuid.New().String()
	u2 = uuid.New().String()
)

// fakeUserRepository implements UserRepository in memory.
type fakeUserRepository struct {
	users   map[string]*User
	byEmail map[string]string // email -> id
}

func newFakeRepo(seed ...*User) *fakeUserRepository {
	r := &fakeUserRepository{users: map[string]*User{}, byEmail: map[string]string{}}
	for _, u := range seed {
		r.users[u.ID] = u
		r.byEmail[u.Email] = u.ID
	}
	return r
}

type fakeTransactor struct {
	called int
}

func (f *fakeTransactor) RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	f.called++
	return fn(ctx)
}

func (f *fakeTransactor) WithTransactionContext(ctx context.Context, fn func(txCtx context.Context) error) error {
	return f.RunInTx(ctx, fn)
}

func newTestUserService(repo UserRepository) UserService {
	return NewUserService(repo, &fakeTransactor{})
}

func (f *fakeUserRepository) GetAll(ctx context.Context) ([]User, error) {
	out := make([]User, 0, len(f.users))
	for _, u := range f.users {
		if u.IsActive {
			out = append(out, *u)
		}
	}
	return out, nil
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if id, ok := f.byEmail[email]; ok {
		return f.users[id], nil
	}
	return nil, nil
}

func (f *fakeUserRepository) Create(ctx context.Context, user *User) error {
	if _, exists := f.byEmail[user.Email]; exists {
		return errors.New("duplicate email")
	}
	f.users[user.ID] = user
	f.byEmail[user.Email] = user.ID
	return nil
}

func (f *fakeUserRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if name, ok := updates["name"].(string); ok {
		u.Name = name
	}
	if email, ok := updates["email"].(string); ok {
		delete(f.byEmail, u.Email)
		u.Email = email
		f.byEmail[email] = id
	}
	if phone, ok := updates["phone"].(string); ok {
		u.Phone = sql.NullString{String: phone, Valid: phone != ""}
	}
	u.UpdatedAt = time.Now()
	return u, nil
}

func (f *fakeUserRepository) Delete(ctx context.Context, id string) error {
	u, ok := f.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	u.IsActive = false
	return nil
}

func (f *fakeUserRepository) HardDelete(ctx context.Context, id string) error {
	if _, ok := f.users[id]; !ok {
		return sql.ErrNoRows
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	_, ok := f.users[id]
	return ok, nil
}

func (f *fakeUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := f.byEmail[email]
	return ok, nil
}

func seedUser(id, email string) *User {
	return &User{
		ID:        id,
		Name:      "Jane",
		Email:     email,
		Phone:     sql.NullString{String: "0811223344", Valid: true},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func mustBeCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	appErr, ok := err.(*customErrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != want {
		t.Errorf("code = %q, want %q", appErr.Code, want)
	}
}

func TestServiceGetAll(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo(seedUser(u1, "a@x.com"), seedUser(u2, "b@x.com")))
	users, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestServiceGetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo())
	_, err := svc.GetByID(ctx, uuid.New().String())
	mustBeCode(t, err, "NOT_FOUND")
}

func TestServiceGetByIDInvalidUUID(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo())
	_, err := svc.GetByID(ctx, "not-a-uuid")
	mustBeCode(t, err, "BAD_REQUEST")
}

func TestServiceCreateDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo(seedUser(u1, "dupe@x.com")))
	_, err := svc.Create(ctx, &CreateUserRequest{Name: "X", Email: "dupe@x.com"})
	mustBeCode(t, err, "CONFLICT")
}

func TestServiceCreateSuccess(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo())
	resp, err := svc.Create(ctx, &CreateUserRequest{Name: "New", Email: "new@x.com", Phone: "081000"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID == "" || resp.Email != "new@x.com" {
		t.Errorf("unexpected response: %#v", resp)
	}
}

func TestServiceUpdateSuccess(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo(seedUser(u1, "a@x.com"))
	svc := newTestUserService(repo)
	resp, err := svc.Update(ctx, u1, &UpdateUserRequest{Name: "Renamed"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Name != "Renamed" {
		t.Errorf("expected renamed, got %#v", resp)
	}
}

func TestServiceUpdateEmailConflict(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo(seedUser(u1, "a@x.com"), seedUser(u2, "taken@x.com"))
	repo.byEmail["taken@x.com"] = u2
	svc := newTestUserService(repo)
	_, err := svc.Update(ctx, u1, &UpdateUserRequest{Email: "taken@x.com"})
	mustBeCode(t, err, "CONFLICT")
}

func TestServiceDeleteNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo())
	err := svc.Delete(ctx, uuid.New().String())
	mustBeCode(t, err, "NOT_FOUND")
}

func TestServiceHardDeleteSuccess(t *testing.T) {
	ctx := context.Background()
	svc := newTestUserService(newFakeRepo(seedUser(u1, "a@x.com")))
	if err := svc.HardDelete(ctx, u1); err != nil {
		t.Fatalf("HardDelete failed: %v", err)
	}
}

func TestServiceCreateWithTransactor(t *testing.T) {
	ctx := context.Background()
	tr := &fakeTransactor{}
	svc := NewUserService(newFakeRepo(), tr)
	resp, err := svc.Create(ctx, &CreateUserRequest{Name: "Alice", Email: "alice@x.com"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp == nil || resp.Name != "Alice" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
	if tr.called != 1 {
		t.Fatalf("expected transactor to be called once, got %d", tr.called)
	}
}
