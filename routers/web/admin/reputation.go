// Copyright 2026 The Gitea Authors.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"

	"code.gitea.io/gitea/models/db"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/templates"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/services/forms"
)

const (
	tplLabels    templates.TplName = "admin/reputation/list"
	tplLabelsNew templates.TplName = "admin/reputation/new"
	tplLabelView templates.TplName = "admin/reputation/view"
)

// ReputationLabels show all reputation labels
func ReputationLabels(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.reputation.labels")
	ctx.Data["PageIsAdminReputation"] = true
	keyword := ctx.FormTrim("q")

	opts := &user_model.SearchReputationLabelOptions{
		ListOptions: db.ListOptions{
			PageSize: setting.UI.Admin.UserPagingNum,
			Page:     ctx.FormInt("page"),
		},
		Keyword: keyword,
	}

	if opts.Page <= 1 {
		opts.Page = 1
	}

	labels, err := user_model.SearchReputationLabels(ctx, opts)
	if err != nil {
		labels = []*user_model.ReputationLabel{}
	}
	ctx.Data["Labels"] = labels
	ctx.Data["Total"] = len(labels)

	ctx.Data["SortType"] = ctx.FormString("sort")
	ctx.Data["Keyword"] = keyword

	ctx.HTML(http.StatusOK, tplLabels)
}

// NewReputationLabel render adding a new label page
func NewReputationLabel(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.reputation.new_label")
	ctx.Data["PageIsAdminReputation"] = true
	ctx.HTML(http.StatusOK, tplLabelsNew)
}

// NewReputationLabelPost response for adding a new label
func NewReputationLabelPost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.AdminCreateReputationLabelForm)
	ctx.Data["Title"] = ctx.Tr("admin.reputation.new_label")
	ctx.Data["PageIsAdminReputation"] = true
	if ctx.HasError() {
		ctx.HTML(http.StatusOK, tplLabelsNew)
		return
	}

	l := &user_model.ReputationLabel{
		Name:        form.Name,
		Color:       form.Color,
		Description: form.Description,
	}

	if err := user_model.CreateReputationLabel(ctx, l); err != nil {
		ctx.Flash.Error(ctx.Tr("admin.reputation.new_error", l.Name, err))
		ctx.Redirect(setting.AppSubURL + "/-/admin/reputation")
		return
	}

	ctx.Flash.Success(ctx.Tr("admin.reputation.new_success", l.Name))
	ctx.Redirect(setting.AppSubURL + "/-/admin/reputation")
}

// DeleteReputationLabel response for deleting a label
func DeleteReputationLabel(ctx *context.Context) {
	l, err := user_model.GetReputationLabel(ctx, ctx.FormString("name"))
	if err != nil {
		ctx.ServerError("GetReputationLabel", err)
		return
	}

	if err = user_model.DeleteReputationLabel(ctx, l); err != nil {
		ctx.ServerError("DeleteReputationLabel", err)
		return
	}

	log.Trace("Reputation Label deleted by admin (%s): %s", ctx.Doer.Name, l.Name)
	ctx.Flash.Success(ctx.Tr("admin.reputation.deletion_success"))
	ctx.JSONRedirect("")
}

// ReputationLabels show all reputation labels
func ViewReputationLabel(ctx *context.Context) {
	l, err := user_model.GetReputationLabel(ctx, ctx.PathParam("name"))
	if err != nil {
		ctx.ServerError("GetReputationLabel", err)
		return
	}

	ctx.Data["Title"] = ctx.Tr("admin.reputation.label", l.Name)
	ctx.Data["PageIsAdminReputation"] = true
	ctx.Data["Label"] = l

	// TODO: Get all related orgs, users
	users, orgs, err := user_model.GetReputationLabelRelated(ctx, l)
	if err != nil {
		ctx.ServerError("GetReputationLabelRelated", err)
		return
	}

	type UserList struct {
		Users []*user_model.User
		ShowUserEmail bool
		IsSigned bool
		PageIsAdminUsers bool
	}

	ctx.Data["Users"] = UserList{users, true, false, false}
	ctx.Data["UsersTotal"] = len(users)
	ctx.Data["Orgs"] = UserList{orgs, true, false, false}
	ctx.Data["OrgsTotal"] = len(orgs)

	ctx.HTML(http.StatusOK, tplLabelView)
}

// AddReputationLabel response for adding a label to a user
func AddReputationLabel(ctx *context.Context) {
	l, err := user_model.GetReputationLabel(ctx, ctx.PathParam("name"))
	if err != nil {
		ctx.ServerError("GetReputationLabel", err)
		return
	}
	uname := ctx.FormString("uname")
	var user *user_model.User
	user, err = user_model.GetUserByName(ctx, uname)
	if err != nil {
		ctx.ServerError("GetUserByName", err)
		return
	}
	err = user_model.AddUserReputationLabel(ctx, user, l)
	if err != nil {
		ctx.ServerError("AddUserReputationLabel", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("admin.reputation.add_success", uname))
	ctx.Flash.Success("AAAAAAAAAAA" + user.Name + l.Name)
	ctx.Redirect(setting.AppSubURL + "/-/admin/reputation/" + l.Name)
}
