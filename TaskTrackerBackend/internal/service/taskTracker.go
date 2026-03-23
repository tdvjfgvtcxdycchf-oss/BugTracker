package service

import (
	"bug_tracker/internal/sql"
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskTrackerService struct {
	conn *pgxpool.Pool
}

func NewTaskTrackerService(conn *pgxpool.Pool) *TaskTrackerService {
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

func (t *TaskTrackerService) GetOtherUsersEmails(ctx context.Context, excludeId int) ([]string, error) {
	return sql.GetOtherUsersEmails(ctx, t.conn, excludeId)
}

func (t *TaskTrackerService) GetAllTasks(ctx context.Context) ([]sql.Task, error) {
	return sql.GetAllTasks(ctx, t.conn)
}

func (t *TaskTrackerService) CreateTask(ctx context.Context, task sql.Task) error {
	return sql.CreateTask(ctx, t.conn, task)
}

func (t *TaskTrackerService) GetAllBugs(ctx context.Context, id int) ([]sql.Bug, error) {
	return sql.GetBugsByTaskId(ctx, t.conn, id)
}

func (t *TaskTrackerService) CreateBug(ctx context.Context, bug sql.Bug) error {
	return sql.CreateBug(ctx, t.conn, bug)
}

func (t *TaskTrackerService) UpdateBug(ctx context.Context, bug sql.Bug, assignedEmail string) error {
	return sql.ChangeBug(ctx, t.conn, bug, assignedEmail)
}

func (t *TaskTrackerService) DeleteTask(ctx context.Context, taskID int, ownerID int) (bool, error) {
	return sql.DeleteTask(ctx, t.conn, taskID, ownerID)
}

func (t *TaskTrackerService) DeleteBug(ctx context.Context, bugID int, creatorID int) (bool, error) {
	return sql.DeleteBug(ctx, t.conn, bugID, creatorID)
}
