// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"xorm.io/xorm"
)

func CreateReputationLabelTable(x *xorm.Engine) error {
	type ReputationLabel struct {
		ID          int64  `xorm:"pk autoincr"`
		Name        string `xorm:"UNIQUE"`
		Description string
		Color       string `xorm:"VARCHAR(7)"`
	}

	type UserReputationLabel struct {
		ID                int64 `xorm:"pk autoincr"`
		ReputationLabelID int64
		UserID            int64 `xorm:"INDEX"`
	}

	if err := x.Sync(new(ReputationLabel)); err != nil {
		return err
	}
	return x.Sync(new(UserReputationLabel))
}
