// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package files

import (
	"testing"
	"time"

	"code.gitea.io/gitea/models/unittest"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/routers/api/v1/utils"
	"code.gitea.io/gitea/services/contexttest"

	_ "code.gitea.io/gitea/models/actions"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}

func getExpectedReadmeContentsResponse() *api.ContentsResponse {
	treePath := "README.md"
	sha := "4b4851ad51df6a7d9f25c979345979eaeb5b349f"
	encoding := "base64"
	content := "IyByZXBvMQoKRGVzY3JpcHRpb24gZm9yIHJlcG8x"
	selfURL := "https://try.gitea.io/api/v1/repos/user2/repo1/contents/" + treePath + "?ref=master"
	htmlURL := "https://try.gitea.io/user2/repo1/src/branch/master/" + treePath
	gitURL := "https://try.gitea.io/api/v1/repos/user2/repo1/git/blobs/" + sha
	downloadURL := "https://try.gitea.io/user2/repo1/raw/branch/master/" + treePath
	return &api.ContentsResponse{
		Name:              treePath,
		Path:              treePath,
		SHA:               "4b4851ad51df6a7d9f25c979345979eaeb5b349f",
		LastCommitSHA:     "65f1bf27bc3bf70f64657658635e66094edbcb4d",
		LastCommitterDate: time.Date(2017, time.March, 19, 16, 47, 59, 0, time.FixedZone("", -14400)),
		LastAuthorDate:    time.Date(2017, time.March, 19, 16, 47, 59, 0, time.FixedZone("", -14400)),
		Type:              "file",
		Size:              30,
		Encoding:          &encoding,
		Content:           &content,
		URL:               &selfURL,
		HTMLURL:           &htmlURL,
		GitURL:            &gitURL,
		DownloadURL:       &downloadURL,
		Links: &api.FileLinksResponse{
			Self:    &selfURL,
			GitURL:  &gitURL,
			HTMLURL: &htmlURL,
		},
	}
}

func TestGetContents(t *testing.T) {
	unittest.PrepareTestEnv(t)
	ctx, _ := contexttest.MockContext(t, "user2/repo1")
	ctx.SetPathParam("id", "1")
	contexttest.LoadRepo(t, ctx, 1)
	contexttest.LoadRepoCommit(t, ctx)
	contexttest.LoadUser(t, ctx, 2)
	contexttest.LoadGitRepo(t, ctx)
	defer ctx.Repo.GitRepo.Close()
	repo, gitRepo := ctx.Repo.Repository, ctx.Repo.GitRepo
	refCommit, err := utils.ResolveRefCommit(ctx, ctx.Repo.Repository, ctx.Repo.Repository.DefaultBranch)
	require.NoError(t, err)

	t.Run("GetContentsOrList(README.md)-MetaOnly", func(t *testing.T) {
		expectedContentsResponse := getExpectedReadmeContentsResponse()
		expectedContentsResponse.Encoding = nil // because will be in a list, doesn't have encoding and content
		expectedContentsResponse.Content = nil
		extResp, err := GetContentsOrList(ctx, repo, gitRepo, refCommit, GetContentsOrListOptions{TreePath: "README.md", IncludeSingleFileContent: false})
		assert.Equal(t, expectedContentsResponse, extResp.FileContents)
		assert.NoError(t, err)
	})

	t.Run("GetContentsOrList(README.md)", func(t *testing.T) {
		expectedContentsResponse := getExpectedReadmeContentsResponse()
		extResp, err := GetContentsOrList(ctx, repo, gitRepo, refCommit, GetContentsOrListOptions{TreePath: "README.md", IncludeSingleFileContent: true})
		assert.Equal(t, expectedContentsResponse, extResp.FileContents)
		assert.NoError(t, err)
	})

	t.Run("GetContentsOrList(RootDir)", func(t *testing.T) {
		readmeContentsResponse := getExpectedReadmeContentsResponse()
		readmeContentsResponse.Encoding = nil // because will be in a list, doesn't have encoding and content
		readmeContentsResponse.Content = nil
		expectedContentsListResponse := []*api.ContentsResponse{readmeContentsResponse}
		// even if IncludeFileContent is true, it has no effect for directory listing
		extResp, err := GetContentsOrList(ctx, repo, gitRepo, refCommit, GetContentsOrListOptions{TreePath: "", IncludeSingleFileContent: true})
		assert.Equal(t, expectedContentsListResponse, extResp.DirContents)
		assert.NoError(t, err)
	})

	t.Run("GetContentsOrList(NoSuchTreePath)", func(t *testing.T) {
		extResp, err := GetContentsOrList(ctx, repo, gitRepo, refCommit, GetContentsOrListOptions{TreePath: "no-such/file.md"})
		assert.Error(t, err)
		assert.EqualError(t, err, "object does not exist [id: , rel_path: no-such]")
		assert.Nil(t, extResp.DirContents)
		assert.Nil(t, extResp.FileContents)
	})

	t.Run("GetBlobBySHA", func(t *testing.T) {
		sha := "65f1bf27bc3bf70f64657658635e66094edbcb4d"
		ctx.SetPathParam("id", "1")
		ctx.SetPathParam("sha", sha)
		gbr, err := GetBlobBySHA(ctx.Repo.Repository, ctx.Repo.GitRepo, ctx.PathParam("sha"))
		expectedGBR := &api.GitBlobResponse{
			Content:  util.ToPointer("dHJlZSAyYTJmMWQ0NjcwNzI4YTJlMTAwNDllMzQ1YmQ3YTI3NjQ2OGJlYWI2CmF1dGhvciB1c2VyMSA8YWRkcmVzczFAZXhhbXBsZS5jb20+IDE0ODk5NTY0NzkgLTA0MDAKY29tbWl0dGVyIEV0aGFuIEtvZW5pZyA8ZXRoYW50a29lbmlnQGdtYWlsLmNvbT4gMTQ4OTk1NjQ3OSAtMDQwMAoKSW5pdGlhbCBjb21taXQK"),
			Encoding: util.ToPointer("base64"),
			URL:      "https://try.gitea.io/api/v1/repos/user2/repo1/git/blobs/65f1bf27bc3bf70f64657658635e66094edbcb4d",
			SHA:      "65f1bf27bc3bf70f64657658635e66094edbcb4d",
			Size:     180,
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedGBR, gbr)
	})
}
