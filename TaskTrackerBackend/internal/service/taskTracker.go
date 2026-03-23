package service

import (
	"bug_tracker/internal/sql"
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type TaskTrackerService struct {
	conn *pgx.Conn
}

func NewTaskTrackerService(conn *pgx.Conn) *TaskTrackerService {
	return &TaskTrackerService{conn: conn}
}

func (t *TaskTrackerService) Register(ctx context.Context, user sql.User) (int, error) {
	existingUser, err := sql.GetByEmail(ctx, t.conn, user)
	if err != nil {
		return 0, err
	}

	if existingUser != nil {
		slog.Warn("registration failed: email already taken", "email", user.Email)
		return 0, errors.New("user already exists")
	}

	id, err := sql.CreateUser(ctx, t.conn, user)
	if err != nil {
		return 0, err
	}

	slog.Info("user registered successfully", "id", strconv.Itoa(id))
	return id, nil
}

func (t *TaskTrackerService) Login(ctx context.Context, user sql.User) (int, error) {
	existingUser, err := sql.GetByEmail(ctx, t.conn, user)
	if err != nil {
		return 0, err
	}

	if existingUser == nil {
		slog.Warn("login failed: there is no user with this email", "email", user.Email)
		return 0, errors.New("user not exist")
	}

	if existingUser.Password != user.Password {
		slog.Warn("login failed: the password is incorrect")
		return 0, errors.New("password incorrect")
	}

	slog.Info("user logined successfully", "id", existingUser.Id)
	return existingUser.Id, nil
}

func (t *TaskTrackerService) GetAllTasks(ctx context.Context) ([]sql.Task, error) {
	tasks, err := sql.GetAllTasks(ctx, t.conn)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
