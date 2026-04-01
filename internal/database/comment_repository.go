// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含评论数据仓库的核心逻辑，包括：
//   - 评论 CRUD 操作（创建、查询、更新、删除）
//   - 嵌套评论支持（最大深度 5 层）
//   - 评论状态管理（待审核、已批准、已删除）
//   - 评论点赞关联操作
//
// 主要用途：
//
//	用于管理文章评论数据，支持嵌套评论和点赞功能。
//
// 注意事项：
//   - 评论删除采用软删除（标记为 deleted 状态）
//   - 嵌套评论最大深度为 5 层
//   - 点赞操作使用原子更新保证一致性
//
// 作者：xfy
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"rua.plus/cadmus/internal/core/comment"
)

// CommentRepository 评论数据仓库实现。
//
// 负责评论数据的 CRUD 操作，支持嵌套评论和状态管理。
// 所有操作通过连接池执行，确保高效的数据访问。
type CommentRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewCommentRepository 创建评论仓库
func NewCommentRepository(pool *Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

// Create 创建新评论
func (r *CommentRepository) Create(ctx context.Context, input *comment.CreateCommentInput) (*comment.Comment, error) {
	// 计算嵌套深度
	depth := 0
	if input.ParentID != nil {
		parent, err := r.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, comment.ErrParentNotFound
		}
		depth = parent.Depth + 1
		// 最大深度限制为 5
		if depth > 5 {
			return nil, comment.ErrMaxDepthExceeded
		}
	}

	query := `
		INSERT INTO comments (
			id, post_id, user_id, parent_id, depth, content, status, like_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	id := uuid.New()
	now := time.Now()
	status := comment.StatusPending

	_, err := r.pool.Exec(ctx, query,
		id,
		input.PostID,
		input.UserID,
		input.ParentID,
		depth,
		input.Content,
		status,
		0,
		now,
		now,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			// 外键约束失败，文章或用户不存在
			if pgErr.ConstraintName == "comments_post_id_fkey" {
				return nil, comment.ErrPostNotFound
			}
			if pgErr.ConstraintName == "comments_user_id_fkey" {
				return nil, comment.ErrUserNotFound
			}
		}
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return &comment.Comment{
		ID:        id,
		PostID:    input.PostID,
		UserID:    input.UserID,
		ParentID:  input.ParentID,
		Depth:     depth,
		Content:   input.Content,
		Status:    status,
		LikeCount: 0,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetByID 根据ID获取评论
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, depth, content, status, like_count, created_at, updated_at
		FROM comments WHERE id = $1
	`

	c, err := r.scanComment(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, comment.ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}
	return c, nil
}

// GetByPostID 获取文章的所有评论（支持筛选）
func (r *CommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, filters *comment.CommentListFilters) ([]*comment.Comment, error) {
	whereClause := "WHERE post_id = $1"
	args := make([]interface{}, 0)
	args = append(args, postID)
	argIndex := 2

	if filters != nil {
		if filters.Status != "" {
			whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, filters.Status)
			argIndex++
		}
		if filters.ParentID != nil {
			whereClause += fmt.Sprintf(" AND parent_id = $%d", argIndex)
			args = append(args, filters.ParentID)
			argIndex++
		}
		if filters.Depth >= 0 {
			whereClause += fmt.Sprintf(" AND depth = $%d", argIndex)
			args = append(args, filters.Depth)
			argIndex++
		}
	}

	query := fmt.Sprintf(`
		SELECT id, post_id, user_id, parent_id, depth, content, status, like_count, created_at, updated_at
		FROM comments %s ORDER BY created_at ASC
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post id: %w", err)
	}
	defer rows.Close()

	return r.scanComments(rows)
}

// GetByUserID 获取用户的所有评论
func (r *CommentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*comment.Comment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, depth, content, status, like_count, created_at, updated_at
		FROM comments WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by user id: %w", err)
	}
	defer rows.Close()

	return r.scanComments(rows)
}

// GetChildren 获取评论的子评论
func (r *CommentRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*comment.Comment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, depth, content, status, like_count, created_at, updated_at
		FROM comments WHERE parent_id = $1 ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment children: %w", err)
	}
	defer rows.Close()

	return r.scanComments(rows)
}

// Update 更新评论
func (r *CommentRepository) Update(ctx context.Context, c *comment.Comment) error {
	query := `
		UPDATE comments SET content = $2, updated_at = $3 WHERE id = $1
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, c.ID, c.Content, now)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return comment.ErrCommentNotFound
	}

	c.UpdatedAt = now
	return nil
}

// UpdateStatus 更新评论状态
func (r *CommentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status comment.CommentStatus) error {
	if !status.IsValid() {
		return comment.ErrInvalidStatus
	}

	query := `UPDATE comments SET status = $2, updated_at = $3 WHERE id = $1`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, id, status, now)
	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return comment.ErrCommentNotFound
	}

	return nil
}

// Delete 删除评论（软删除，标记为 deleted）
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.UpdateStatus(ctx, id, comment.StatusDeleted)
}

// CountByPostID 统计文章的评论数量
func (r *CommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1 AND status = $2`

	var count int
	err := r.pool.QueryRow(ctx, query, postID, comment.StatusApproved).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by post id: %w", err)
	}

	return count, nil
}

// CountByUserID 统计用户的评论数量
func (r *CommentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE user_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by user id: %w", err)
	}

	return count, nil
}

// List 分页获取评论列表
func (r *CommentRepository) List(ctx context.Context, filters *comment.CommentListFilters, offset, limit int) ([]*comment.Comment, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if filters != nil {
		if filters.PostID != uuid.Nil {
			whereClause += fmt.Sprintf(" AND post_id = $%d", argIndex)
			args = append(args, filters.PostID)
			argIndex++
		}
		if filters.UserID != uuid.Nil {
			whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
			args = append(args, filters.UserID)
			argIndex++
		}
		if filters.Status != "" {
			whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, filters.Status)
			argIndex++
		}
		if filters.ParentID != nil {
			whereClause += fmt.Sprintf(" AND parent_id = $%d", argIndex)
			args = append(args, filters.ParentID)
			argIndex++
		}
		if filters.Depth >= 0 {
			whereClause += fmt.Sprintf(" AND depth = $%d", argIndex)
			args = append(args, filters.Depth)
			argIndex++
		}
	}

	query := fmt.Sprintf(`
		SELECT id, post_id, user_id, parent_id, depth, content, status, like_count, created_at, updated_at
		FROM comments %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	return r.scanComments(rows)
}

// scanComment 扫描单行评论数据
func (r *CommentRepository) scanComment(row pgx.Row) (*comment.Comment, error) {
	c := &comment.Comment{}
	err := row.Scan(
		&c.ID,
		&c.PostID,
		&c.UserID,
		&c.ParentID,
		&c.Depth,
		&c.Content,
		&c.Status,
		&c.LikeCount,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// scanComments 扫描多行评论数据
func (r *CommentRepository) scanComments(rows pgx.Rows) ([]*comment.Comment, error) {
	comments := make([]*comment.Comment, 0)
	for rows.Next() {
		c := &comment.Comment{}
		err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&c.ParentID,
			&c.Depth,
			&c.Content,
			&c.Status,
			&c.LikeCount,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// CommentLikeRepository 评论点赞仓库实现
type CommentLikeRepository struct {
	pool *Pool
}

// NewCommentLikeRepository 创建评论点赞仓库
func NewCommentLikeRepository(pool *Pool) *CommentLikeRepository {
	return &CommentLikeRepository{pool: pool}
}

// Create 创建点赞记录
func (r *CommentLikeRepository) Create(ctx context.Context, commentID, userID uuid.UUID) (*comment.CommentLike, error) {
	query := `
		INSERT INTO comment_likes (id, comment_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	id := uuid.New()
	now := time.Now()

	_, err := r.pool.Exec(ctx, query, id, commentID, userID, now)
	if err != nil {
		if isUniqueViolation(err, "comment_likes_comment_id_user_id_key") {
			return nil, comment.ErrAlreadyLiked
		}
		return nil, fmt.Errorf("failed to create comment like: %w", err)
	}

	return &comment.CommentLike{
		ID:        id,
		CommentID: commentID,
		UserID:    userID,
		CreatedAt: now,
	}, nil
}

// CreateIfNotExists 创建点赞记录（使用 ON CONFLICT DO NOTHING），返回是否实际创建
// 同时原子更新评论的点赞计数
func (r *CommentLikeRepository) CreateIfNotExists(ctx context.Context, commentID, userID uuid.UUID) (created bool, err error) {
	query := `
		INSERT INTO comment_likes (id, comment_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (comment_id, user_id) DO NOTHING
	`

	id := uuid.New()
	now := time.Now()

	result, err := r.pool.Exec(ctx, query, id, commentID, userID, now)
	if err != nil {
		return false, fmt.Errorf("failed to create comment like: %w", err)
	}

	created = result.RowsAffected() > 0

	// 只有实际创建点赞记录时才更新计数
	if created {
		updateQuery := `UPDATE comments SET like_count = like_count + 1 WHERE id = $1`
		_, err = r.pool.Exec(ctx, updateQuery, commentID)
		if err != nil {
			return false, fmt.Errorf("failed to update like count: %w", err)
		}
	}

	return created, nil
}

// GetByCommentAndUser 获取用户对评论的点赞记录
func (r *CommentLikeRepository) GetByCommentAndUser(ctx context.Context, commentID, userID uuid.UUID) (*comment.CommentLike, error) {
	query := `
		SELECT id, comment_id, user_id, created_at
		FROM comment_likes WHERE comment_id = $1 AND user_id = $2
	`

	l := &comment.CommentLike{}
	err := r.pool.QueryRow(ctx, query, commentID, userID).Scan(
		&l.ID,
		&l.CommentID,
		&l.UserID,
		&l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, comment.ErrNotLiked
		}
		return nil, fmt.Errorf("failed to get comment like: %w", err)
	}
	return l, nil
}

// Delete 删除点赞记录（取消点赞）
func (r *CommentLikeRepository) Delete(ctx context.Context, commentID, userID uuid.UUID) error {
	query := `DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment like: %w", err)
	}

	if result.RowsAffected() == 0 {
		return comment.ErrNotLiked
	}

	return nil
}

// DeleteIfExists 删除点赞记录（返回是否实际删除）
// 同时原子更新评论的点赞计数
func (r *CommentLikeRepository) DeleteIfExists(ctx context.Context, commentID, userID uuid.UUID) (deleted bool, err error) {
	query := `DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, commentID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete comment like: %w", err)
	}

	deleted = result.RowsAffected() > 0

	// 只有实际删除点赞记录时才更新计数
	if deleted {
		updateQuery := `UPDATE comments SET like_count = like_count - 1 WHERE id = $1 AND like_count > 0`
		_, err = r.pool.Exec(ctx, updateQuery, commentID)
		if err != nil {
			return false, fmt.Errorf("failed to update like count: %w", err)
		}
	}

	return deleted, nil
}

// Exists 检查用户是否已点赞评论
func (r *CommentLikeRepository) Exists(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, commentID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check comment like exists: %w", err)
	}
	return exists, nil
}

// CountByCommentID 统计评论的点赞数量
func (r *CommentLikeRepository) CountByCommentID(ctx context.Context, commentID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM comment_likes WHERE comment_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, commentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comment likes: %w", err)
	}
	return count, nil
}

// GetByUserID 获取用户的所有点赞记录
func (r *CommentLikeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*comment.CommentLike, error) {
	query := `
		SELECT id, comment_id, user_id, created_at
		FROM comment_likes WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment likes by user id: %w", err)
	}
	defer rows.Close()

	likes := make([]*comment.CommentLike, 0)
	for rows.Next() {
		l := &comment.CommentLike{}
		err := rows.Scan(
			&l.ID,
			&l.CommentID,
			&l.UserID,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment like: %w", err)
		}
		likes = append(likes, l)
	}
	return likes, nil
}

// GetLikesBatch 批量检查用户对多个评论的点赞状态
// 返回一个 map，key 是 commentID，value 是是否已点赞
func (r *CommentLikeRepository) GetLikesBatch(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(commentIDs) == 0 || userID == uuid.Nil {
		return make(map[uuid.UUID]bool), nil
	}

	query := `
		SELECT comment_id FROM comment_likes
		WHERE user_id = $1 AND comment_id = ANY($2)
	`

	rows, err := r.pool.Query(ctx, query, userID, commentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch check comment likes: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]bool)
	for rows.Next() {
		var commentID uuid.UUID
		if err := rows.Scan(&commentID); err != nil {
			return nil, fmt.Errorf("failed to scan comment_id: %w", err)
		}
		result[commentID] = true
	}

	return result, nil
}