package sql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatThread struct {
	Id          int    `json:"id"`
	Scope       string `json:"scope"` // org | project | dm
	OrgId       int    `json:"org_id,omitempty"`
	ProjectId   int    `json:"project_id,omitempty"`
	DmUserA     int    `json:"dm_user_a,omitempty"`
	DmUserB     int    `json:"dm_user_b,omitempty"`
	CreatedBy   int    `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	PeerEmail   string `json:"peer_email,omitempty"`
	Title       string `json:"title,omitempty"`
	LastMessage string `json:"last_message,omitempty"`
	LastMessageAt string `json:"last_message_at,omitempty"`
	UnreadCount int `json:"unread_count"`
}

type ChatMessage struct {
	Id        int    `json:"id"`
	ThreadId  int    `json:"thread_id"`
	UserId    int    `json:"user_id"`
	UserEmail string `json:"user_email"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	EditedAt  string `json:"edited_at,omitempty"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

func isOrgMember(ctx context.Context, conn *pgxpool.Pool, userID, orgID int) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM org_member WHERE org_id_fk = $1 AND user_id_fk = $2)`, orgID, userID).Scan(&exists)
	return exists, err
}

func EnsureOrgThread(ctx context.Context, conn *pgxpool.Pool, orgID, actorUserID int) (int, error) {
	ok, err := isOrgMember(ctx, conn, actorUserID, orgID)
	if err != nil || !ok {
		return 0, err
	}
	var threadID int
	err = conn.QueryRow(ctx, `SELECT id_pk FROM chat_thread WHERE scope = 'org' AND org_id_fk = $1`, orgID).Scan(&threadID)
	if err == nil {
		return threadID, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	err = conn.QueryRow(ctx, `INSERT INTO chat_thread (scope, org_id_fk, created_by_fk) VALUES ('org',$1,$2) RETURNING id_pk`, orgID, actorUserID).Scan(&threadID)
	return threadID, err
}

func EnsureProjectThread(ctx context.Context, conn *pgxpool.Pool, projectID, actorUserID int) (int, error) {
	ok, err := CanUserAccessProject(ctx, conn, actorUserID, projectID)
	if err != nil || !ok {
		return 0, err
	}
	var threadID int
	err = conn.QueryRow(ctx, `SELECT id_pk FROM chat_thread WHERE scope = 'project' AND project_id_fk = $1`, projectID).Scan(&threadID)
	if err == nil {
		return threadID, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	err = conn.QueryRow(ctx, `INSERT INTO chat_thread (scope, project_id_fk, created_by_fk) VALUES ('project',$1,$2) RETURNING id_pk`, projectID, actorUserID).Scan(&threadID)
	return threadID, err
}

func EnsureDMThreadByEmail(ctx context.Context, conn *pgxpool.Pool, email string, actorUserID int) (int, error) {
	var otherUserID int
	if err := conn.QueryRow(ctx, `SELECT id_pk FROM "User" WHERE email = $1`, email).Scan(&otherUserID); err != nil {
		return 0, err
	}
	a, b := actorUserID, otherUserID
	if a > b {
		a, b = b, a
	}
	var threadID int
	err := conn.QueryRow(ctx, `SELECT id_pk FROM chat_thread WHERE scope = 'dm' AND dm_user_a_fk = $1 AND dm_user_b_fk = $2`, a, b).Scan(&threadID)
	if err == nil {
		return threadID, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	err = conn.QueryRow(ctx, `INSERT INTO chat_thread (scope, dm_user_a_fk, dm_user_b_fk, created_by_fk) VALUES ('dm',$1,$2,$3) RETURNING id_pk`, a, b, actorUserID).Scan(&threadID)
	return threadID, err
}

func GetOrgThread(ctx context.Context, conn *pgxpool.Pool, orgID, actorUserID int) (*ChatThread, error) {
	ok, err := isOrgMember(ctx, conn, actorUserID, orgID)
	if err != nil || !ok {
		return nil, err
	}
	var t ChatThread
	err = conn.QueryRow(ctx, `
		SELECT id_pk, scope, COALESCE(org_id_fk,0), COALESCE(project_id_fk,0), COALESCE(dm_user_a_fk,0), COALESCE(dm_user_b_fk,0), created_by_fk, created_at::text
		FROM chat_thread
		WHERE scope = 'org' AND org_id_fk = $1
	`, orgID).Scan(&t.Id, &t.Scope, &t.OrgId, &t.ProjectId, &t.DmUserA, &t.DmUserB, &t.CreatedBy, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetProjectThread(ctx context.Context, conn *pgxpool.Pool, projectID, actorUserID int) (*ChatThread, error) {
	ok, err := CanUserAccessProject(ctx, conn, actorUserID, projectID)
	if err != nil || !ok {
		return nil, err
	}
	var t ChatThread
	err = conn.QueryRow(ctx, `
		SELECT id_pk, scope, COALESCE(org_id_fk,0), COALESCE(project_id_fk,0), COALESCE(dm_user_a_fk,0), COALESCE(dm_user_b_fk,0), created_by_fk, created_at::text
		FROM chat_thread
		WHERE scope = 'project' AND project_id_fk = $1
	`, projectID).Scan(&t.Id, &t.Scope, &t.OrgId, &t.ProjectId, &t.DmUserA, &t.DmUserB, &t.CreatedBy, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetDMThreads(ctx context.Context, conn *pgxpool.Pool, actorUserID int) ([]ChatThread, error) {
	rows, err := conn.Query(ctx, `
		SELECT
			t.id_pk, t.scope, COALESCE(t.org_id_fk,0), COALESCE(t.project_id_fk,0), COALESCE(t.dm_user_a_fk,0), COALESCE(t.dm_user_b_fk,0), t.created_by_fk, t.created_at::text,
			CASE WHEN t.dm_user_a_fk = $1 THEN u2.email ELSE u1.email END as peer_email,
			COALESCE(last_msg.body, '') as last_message,
			COALESCE(last_msg.created_at::text, '') as last_message_at,
			COALESCE(unread.cnt, 0) as unread_count
		FROM chat_thread t
		JOIN "User" u1 ON u1.id_pk = t.dm_user_a_fk
		JOIN "User" u2 ON u2.id_pk = t.dm_user_b_fk
		LEFT JOIN LATERAL (
			SELECT m.id_pk, m.body, m.created_at
			FROM chat_message m
			WHERE m.thread_id_fk = t.id_pk
			ORDER BY m.id_pk DESC
			LIMIT 1
		) last_msg ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int as cnt
			FROM chat_message m
			LEFT JOIN chat_read_state rs ON rs.thread_id_fk = t.id_pk AND rs.user_id_fk = $1
			WHERE m.thread_id_fk = t.id_pk
			  AND m.user_id_fk <> $1
			  AND (rs.last_read_message_id IS NULL OR m.id_pk > rs.last_read_message_id)
		) unread ON true
		WHERE t.scope = 'dm' AND ($1 = t.dm_user_a_fk OR $1 = t.dm_user_b_fk)
		ORDER BY COALESCE(last_msg.id_pk, t.id_pk) DESC
	`, actorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChatThread
	for rows.Next() {
		var t ChatThread
		if err := rows.Scan(&t.Id, &t.Scope, &t.OrgId, &t.ProjectId, &t.DmUserA, &t.DmUserB, &t.CreatedBy, &t.CreatedAt, &t.PeerEmail, &t.LastMessage, &t.LastMessageAt, &t.UnreadCount); err != nil {
			return nil, err
		}
		t.Title = t.PeerEmail
		out = append(out, t)
	}
	return out, rows.Err()
}

func GetOrgThreads(ctx context.Context, conn *pgxpool.Pool, orgID, actorUserID int) ([]ChatThread, error) {
	ok, err := isOrgMember(ctx, conn, actorUserID, orgID)
	if err != nil || !ok {
		return nil, err
	}
	rows, err := conn.Query(ctx, `
		SELECT
			t.id_pk, t.scope, COALESCE(t.org_id_fk,0), COALESCE(t.project_id_fk,0), COALESCE(t.dm_user_a_fk,0), COALESCE(t.dm_user_b_fk,0), t.created_by_fk, t.created_at::text,
			o.name as title,
			COALESCE(last_msg.body, '') as last_message,
			COALESCE(last_msg.created_at::text, '') as last_message_at,
			COALESCE(unread.cnt, 0) as unread_count
		FROM chat_thread t
		JOIN organizations o ON o.id_pk = t.org_id_fk
		LEFT JOIN LATERAL (
			SELECT m.id_pk, m.body, m.created_at
			FROM chat_message m
			WHERE m.thread_id_fk = t.id_pk
			ORDER BY m.id_pk DESC
			LIMIT 1
		) last_msg ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int as cnt
			FROM chat_message m
			LEFT JOIN chat_read_state rs ON rs.thread_id_fk = t.id_pk AND rs.user_id_fk = $2
			WHERE m.thread_id_fk = t.id_pk
			  AND m.user_id_fk <> $2
			  AND (rs.last_read_message_id IS NULL OR m.id_pk > rs.last_read_message_id)
		) unread ON true
		WHERE t.scope = 'org' AND t.org_id_fk = $1
		ORDER BY COALESCE(last_msg.id_pk, t.id_pk) DESC
	`, orgID, actorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChatThread
	for rows.Next() {
		var t ChatThread
		if err := rows.Scan(&t.Id, &t.Scope, &t.OrgId, &t.ProjectId, &t.DmUserA, &t.DmUserB, &t.CreatedBy, &t.CreatedAt, &t.Title, &t.LastMessage, &t.LastMessageAt, &t.UnreadCount); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func GetProjectThreads(ctx context.Context, conn *pgxpool.Pool, projectID, actorUserID int) ([]ChatThread, error) {
	ok, err := CanUserAccessProject(ctx, conn, actorUserID, projectID)
	if err != nil || !ok {
		return nil, err
	}
	rows, err := conn.Query(ctx, `
		SELECT
			t.id_pk, t.scope, COALESCE(t.org_id_fk,0), COALESCE(t.project_id_fk,0), COALESCE(t.dm_user_a_fk,0), COALESCE(t.dm_user_b_fk,0), t.created_by_fk, t.created_at::text,
			p.name as title,
			COALESCE(last_msg.body, '') as last_message,
			COALESCE(last_msg.created_at::text, '') as last_message_at,
			COALESCE(unread.cnt, 0) as unread_count
		FROM chat_thread t
		JOIN projects p ON p.id_pk = t.project_id_fk
		LEFT JOIN LATERAL (
			SELECT m.id_pk, m.body, m.created_at
			FROM chat_message m
			WHERE m.thread_id_fk = t.id_pk
			ORDER BY m.id_pk DESC
			LIMIT 1
		) last_msg ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int as cnt
			FROM chat_message m
			LEFT JOIN chat_read_state rs ON rs.thread_id_fk = t.id_pk AND rs.user_id_fk = $2
			WHERE m.thread_id_fk = t.id_pk
			  AND m.user_id_fk <> $2
			  AND (rs.last_read_message_id IS NULL OR m.id_pk > rs.last_read_message_id)
		) unread ON true
		WHERE t.scope = 'project' AND t.project_id_fk = $1
		ORDER BY COALESCE(last_msg.id_pk, t.id_pk) DESC
	`, projectID, actorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChatThread
	for rows.Next() {
		var t ChatThread
		if err := rows.Scan(&t.Id, &t.Scope, &t.OrgId, &t.ProjectId, &t.DmUserA, &t.DmUserB, &t.CreatedBy, &t.CreatedAt, &t.Title, &t.LastMessage, &t.LastMessageAt, &t.UnreadCount); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func CanUserAccessThread(ctx context.Context, conn *pgxpool.Pool, threadID, actorUserID int) (bool, error) {
	var scope string
	var orgID, projectID, dmA, dmB int
	err := conn.QueryRow(ctx, `
		SELECT scope, COALESCE(org_id_fk,0), COALESCE(project_id_fk,0), COALESCE(dm_user_a_fk,0), COALESCE(dm_user_b_fk,0)
		FROM chat_thread WHERE id_pk = $1
	`, threadID).Scan(&scope, &orgID, &projectID, &dmA, &dmB)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	switch scope {
	case "org":
		if orgID <= 0 {
			return false, nil
		}
		return isOrgMember(ctx, conn, actorUserID, orgID)
	case "project":
		if projectID <= 0 {
			return false, nil
		}
		return CanUserAccessProject(ctx, conn, actorUserID, projectID)
	case "dm":
		return dmA == actorUserID || dmB == actorUserID, nil
	default:
		return false, nil
	}
}

func AddChatMessage(ctx context.Context, conn *pgxpool.Pool, threadID, actorUserID int, body string) (int, error) {
	ok, err := CanUserAccessThread(ctx, conn, threadID, actorUserID)
	if err != nil || !ok {
		return 0, err
	}
	var id int
	err = conn.QueryRow(ctx, `
		INSERT INTO chat_message (thread_id_fk, user_id_fk, body)
		VALUES ($1, $2, $3)
		RETURNING id_pk
	`, threadID, actorUserID, body).Scan(&id)
	return id, err
}

func MarkThreadRead(ctx context.Context, conn *pgxpool.Pool, threadID, actorUserID int) error {
	ok, err := CanUserAccessThread(ctx, conn, threadID, actorUserID)
	if err != nil || !ok {
		return err
	}
	var lastID int
	err = conn.QueryRow(ctx, `SELECT id_pk FROM chat_message WHERE thread_id_fk = $1 ORDER BY id_pk DESC LIMIT 1`, threadID).Scan(&lastID)
	if err == pgx.ErrNoRows {
		lastID = 0
	} else if err != nil {
		return err
	}
	var lastIDPtr interface{}
	if lastID > 0 {
		lastIDPtr = lastID
	} else {
		lastIDPtr = nil
	}
	_, err = conn.Exec(ctx, `
		INSERT INTO chat_read_state (thread_id_fk, user_id_fk, last_read_message_id, last_read_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (thread_id_fk, user_id_fk)
		DO UPDATE SET last_read_message_id = EXCLUDED.last_read_message_id, last_read_at = NOW()
	`, threadID, actorUserID, lastIDPtr)
	return err
}

func GetChatMessages(ctx context.Context, conn *pgxpool.Pool, threadID, actorUserID int, limit int, beforeID int) ([]ChatMessage, error) {
	ok, err := CanUserAccessThread(ctx, conn, threadID, actorUserID)
	if err != nil || !ok {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if beforeID <= 0 {
		beforeID = 1<<31 - 1
	}
	rows, err := conn.Query(ctx, `
		SELECT m.id_pk, m.thread_id_fk, m.user_id_fk, u.email, COALESCE(m.body,''), m.created_at::text, COALESCE(m.edited_at::text,''), COALESCE(m.deleted_at::text,'')
		FROM chat_message m
		JOIN "User" u ON u.id_pk = m.user_id_fk
		WHERE m.thread_id_fk = $1
		  AND m.id_pk < $2
		ORDER BY m.id_pk DESC
		LIMIT $3
	`, threadID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.Id, &m.ThreadId, &m.UserId, &m.UserEmail, &m.Body, &m.CreatedAt, &m.EditedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func UpdateChatMessage(ctx context.Context, conn *pgxpool.Pool, messageID, actorUserID int, body string) (bool, error) {
	var updatedID int
	err := conn.QueryRow(ctx, `
		UPDATE chat_message
		SET body = $1, edited_at = NOW()
		WHERE id_pk = $2 AND user_id_fk = $3 AND deleted_at IS NULL
		RETURNING id_pk
	`, body, messageID, actorUserID).Scan(&updatedID)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func DeleteChatMessage(ctx context.Context, conn *pgxpool.Pool, messageID, actorUserID int) (bool, error) {
	var updatedID int
	err := conn.QueryRow(ctx, `
		UPDATE chat_message
		SET body = '', deleted_at = NOW()
		WHERE id_pk = $1 AND user_id_fk = $2 AND deleted_at IS NULL
		RETURNING id_pk
	`, messageID, actorUserID).Scan(&updatedID)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func UpsertTypingState(ctx context.Context, conn *pgxpool.Pool, threadID, actorUserID int, isTyping bool) error {
	ok, err := CanUserAccessThread(ctx, conn, threadID, actorUserID)
	if err != nil || !ok {
		return err
	}
	_, err = conn.Exec(ctx, `
		INSERT INTO chat_typing_state (thread_id_fk, user_id_fk, is_typing, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (thread_id_fk, user_id_fk)
		DO UPDATE SET is_typing = EXCLUDED.is_typing, updated_at = NOW()
	`, threadID, actorUserID, isTyping)
	return err
}

func GetTypingUsers(ctx context.Context, conn *pgxpool.Pool, threadID, actorUserID int) ([]string, error) {
	ok, err := CanUserAccessThread(ctx, conn, threadID, actorUserID)
	if err != nil || !ok {
		return nil, err
	}
	rows, err := conn.Query(ctx, `
		SELECT u.email
		FROM chat_typing_state s
		JOIN "User" u ON u.id_pk = s.user_id_fk
		WHERE s.thread_id_fk = $1
		  AND s.user_id_fk <> $2
		  AND s.is_typing = true
		  AND s.updated_at > NOW() - INTERVAL '7 seconds'
		ORDER BY u.email
	`, threadID, actorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emails []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}
	return emails, rows.Err()
}

