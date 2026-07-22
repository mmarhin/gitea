// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/gitea/models/db"

	"xorm.io/builder"
)

// ReputationLabel represent a reputation label
type ReputationLabel struct {
	ID          int64  `xorm:"pk autoincr"`
	Name        string `xorm:"UNIQUE"`
	Description string
	Color       string `xorm:"VARCHAR(7)"`
}

// UserReputationLabel represents a relation between user and //
// ReputationLabel
type UserReputationLabel struct {
	ID                int64 `xorm:"pk autoincr"`
	ReputationLabelID int64
	UserID            int64 `xorm:"INDEX"`
}

func init() {
	db.RegisterModel(new(ReputationLabel))
	db.RegisterModel(new(UserReputationLabel))
}

// GetUserReputationLabels returns the user's reputation labels.
func GetUserReputationLabels(ctx context.Context, u *User) ([]*ReputationLabel, int64, error) {
	sess := db.GetEngine(ctx).
		Select("`reputation_label`.*").
		Join("INNER", "user_reputation_label", "`user_reputation_label`.reputation_label_id=reputation_label.id").
		Where("user_reputation_label.user_id=?", u.ID)

	labels := make([]*ReputationLabel, 0, 8)
	count, err := sess.FindAndCount(&labels)
	return labels, count, err
}

// CreateReputationLabel creates a new reputation label.
func CreateReputationLabel(ctx context.Context, label *ReputationLabel) error {
	_, err := db.GetEngine(ctx).Insert(label)
	return err
}

// SearchReputationLabelOptions are options to search labels for the admin panel
type SearchReputationLabelOptions struct {
	db.ListOptions
	Keyword string
}

// SearchReputationLabels returns all reputation label.
func SearchReputationLabels(ctx context.Context, opts *SearchReputationLabelOptions) ([]*ReputationLabel, error) {
	cond := builder.NewCond()
	if len(opts.Keyword) > 0 {
		likeStr := "%" + strings.ToLower(opts.Keyword) + "%"
		cond = builder.Like{"lower(name)", likeStr}
	}

	opts.SetDefaultValues()

	labels := make([]*ReputationLabel, 0, opts.PageSize)
	err := db.GetEngine(ctx).
		Where(cond).
		Limit(opts.PageSize, (opts.Page-1)*opts.PageSize).
		Find(&labels)

	return labels, err
}

// GetReputationLabel returns a reputation label.
func GetReputationLabel(ctx context.Context, name string) (*ReputationLabel, error) {
	label := new(ReputationLabel)
	has, err := db.GetEngine(ctx).Where("name=?", name).Get(label)
	if !has {
		return nil, err
	}
	return label, err
}

// UpdateReputationLabel updates a label based on its name.
func UpdateReputationLabel(ctx context.Context, label *ReputationLabel) error {
	_, err := db.GetEngine(ctx).Where("name=?", label.Name).Update(label)
	return err
}

// DeleteReputationLabel deletes a label.
func DeleteReputationLabel(ctx context.Context, label *ReputationLabel) error {
	_, err := db.GetEngine(ctx).Where("name=?", label.Name).Delete(label)
	return err
}

// AddUserReputationLabel adds a reputation label to a user.
func AddUserReputationLabel(ctx context.Context, u *User, label *ReputationLabel) error {
	return AddUserReputationLabels(ctx, u, []*ReputationLabel{label})
}

// AddUserReputationLabels adds labels to a user.
func AddUserReputationLabels(ctx context.Context, u *User, labels []*ReputationLabel) error {
	return db.WithTx(ctx, func(ctx context.Context) error {
		for _, label := range labels {
			// hydrate label and check if it exists
			has, err := db.GetEngine(ctx).Where("name=?", label.Name).Get(label)
			if err != nil {
				return err
			} else if !has {
				return fmt.Errorf("label with name %s doesn't exist", label.Name)
			}
			if err := db.Insert(ctx, &UserReputationLabel{
				ReputationLabelID: label.ID,
				UserID:  u.ID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveAllUserReputationLabels removes all labels from a user.
func RemoveAllUserReputationLabels(ctx context.Context, u *User) error {
	_, err := db.GetEngine(ctx).Where("user_id=?", u.ID).Delete(&UserReputationLabel{})
	return err
}

func GetReputationLabelRelated(ctx context.Context, label *ReputationLabel) ([]*User, []*User, error) {
	users := make([]*User, 0)
	orgs := make([]*User, 0)

	err := db.GetEngine(ctx).
		Join("INNER", "user_reputation_label", "user.id = `user_reputation_label`.user_id").
		Where("`user_reputation_label`.reputation_label_id=?", label.ID).
		And("`user`.type=?", UserTypeIndividual).
		Find(&users)

	err = db.GetEngine(ctx).
		Join("INNER", "user_reputation_label", "user.id = `user_reputation_label`.user_id").
		Where("`user_reputation_label`.reputation_label_id=?", label.ID).
		And("`user`.type=?", UserTypeOrganization).
		Find(&orgs)

	return users, orgs, err
}
