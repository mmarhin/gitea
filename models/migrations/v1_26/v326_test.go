// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"testing"

	"code.gitea.io/gitea/models/migrations/base"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateReputationLabelTable(t *testing.T) {
	x, deferrable := base.PrepareTestEnv(t, 0)
	defer deferrable()

	require.NoError(t, CreateReputationLabelTable(x))
	tables := base.LoadTableSchemasMap(t, x)
	assert.Contains(t, tables, "reputation_label")
	assert.Contains(t, tables, "user_reputation_label")
}
