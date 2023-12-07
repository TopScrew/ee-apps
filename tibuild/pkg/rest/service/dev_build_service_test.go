package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"tibuild/gojenkins"
	"time"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	saved DevBuild
}

func (m *mockRepo) Create(ctx context.Context, req DevBuild) (resp *DevBuild, err error) {
	req.ID = 1
	m.saved = req
	return &req, nil
}
func (m *mockRepo) Get(ctx context.Context, id int) (resp *DevBuild, err error) {
	return &m.saved, nil
}
func (m *mockRepo) Update(ctx context.Context, id int, req DevBuild) (resp *DevBuild, err error) {
	m.saved = req
	return &req, nil
}
func (m *mockRepo) List(ctx context.Context, option DevBuildListOption) (resp []DevBuild, err error) {
	return nil, nil
}

type mockJenkins struct {
	params map[string]string
	resume chan struct{}
	job    *gojenkins.Build
}

func (m *mockJenkins) BuildJob(ctx context.Context, name string, params map[string]string) (int64, error) {
	m.params = params
	return 1, nil
}
func (m mockJenkins) GetBuildNumberFromQueueID(ctx context.Context, queueid int64, jobname string) (int64, error) {
	if m.resume != nil {
		<-m.resume
	}
	return 2, nil
}

func (mock mockJenkins) GetBuild(ctx context.Context, jobName string, number int64) (*gojenkins.Build, error) {
	return mock.job, nil
}
func (mock mockJenkins) BuildURL(jobName string, number int64) string {
	return fmt.Sprintf("%s/blue/organizations/jenkins/%s/detail/%s/%d/pipeline", "https://cd.pingcap.net/", jobName, jobName, number)
}

func TestDevBuildCreate(t *testing.T) {
	mockedJenkins := &mockJenkins{resume: make(chan struct{})}
	mockedRepo := mockRepo{}
	server := DevbuildServer{
		Repo:    &mockedRepo,
		Jenkins: mockedJenkins,
		Now:     time.Now,
	}

	t.Run("ok", func(t *testing.T) {
		entity, err := server.Create(context.TODO(),
			DevBuild{Spec: DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", PluginGitRef: "master", GithubRepo: "pingcap/tidb"}},
			DevBuildSaveOption{})
		require.NoError(t, err)
		require.Equal(t, 1, entity.ID)
		require.Equal(t, BuildStatusPending, entity.Status.Status)
		require.Equal(t, int64(0), entity.Status.PipelineBuildID)
		require.Equal(t, map[string]string{"Edition": "enterprise", "GitRef": "pull/23", "Product": "tidb",
			"Version": "v6.1.2", "PluginGitRef": "master", "IsPushGCR": "false", "IsHotfix": "false", "Features": "",
			"GithubRepo": "pingcap/tidb", "TiBuildID": "1", "BuildEnv": "", "ProductDockerfile": "", "BuilderImg": "", "ProductBaseImg": ""}, mockedJenkins.params)
		mockedJenkins.resume <- struct{}{}
		time.Sleep(time.Millisecond)
		require.Equal(t, int64(2), mockedRepo.saved.Status.PipelineBuildID)
	})
	t.Run("auto fill plugin", func(t *testing.T) {
		entity, err := server.Create(context.TODO(),
			DevBuild{Spec: DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23"}},
			DevBuildSaveOption{})
		require.NoError(t, err)
		require.Equal(t, "release-6.1", entity.Spec.PluginGitRef)
	})
	t.Run("auto fill fips", func(t *testing.T) {
		entity, err := server.Create(context.TODO(),
			DevBuild{Spec: DevBuildSpec{Product: ProductTikv, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", Features: "fips"}},
			DevBuildSaveOption{})
		require.NoError(t, err)
		require.Equal(t, FIPS_BUILD_ENV, entity.Spec.BuildEnv)
		require.Equal(t, FIPS_TIKV_BUILDER, entity.Spec.BuilderImg)
		require.Equal(t, FIPS_TIKV_BASE, entity.Spec.ProductBaseImg)
	})
	t.Run("auto fill fips", func(t *testing.T) {
		entity, err := server.Create(context.TODO(),
			DevBuild{Spec: DevBuildSpec{Product: ProductBr, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", Features: "fips"}},
			DevBuildSaveOption{})
		require.NoError(t, err)
		require.Equal(t, FIPS_BUILD_ENV, entity.Spec.BuildEnv)
		require.Equal(t, "https://raw.githubusercontent.com/PingCAP-QE/artifacts/main/dockerfiles/br.Dockerfile", entity.Spec.ProductDockerfile)
	})
	t.Run("bad enterprise plugin", func(t *testing.T) {
		_, err := server.Create(context.TODO(),
			DevBuild{Spec: DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", PluginGitRef: "maste"}},
			DevBuildSaveOption{})
		require.ErrorContains(t, err, "pluginGitRef is not valid")
		require.ErrorIs(t, err, ErrBadRequest)
	})
	t.Run("bad product", func(t *testing.T) {
		_, err := server.Create(context.TODO(), DevBuild{Spec: DevBuildSpec{Product: ""}}, DevBuildSaveOption{})
		require.ErrorIs(t, err, ErrBadRequest)
	})

	t.Run("bad version", func(t *testing.T) {
		_, err := server.Create(context.TODO(), DevBuild{Spec: DevBuildSpec{Product: ProductTidb, Version: "av6.1.2", Edition: CommunityEdition}}, DevBuildSaveOption{})
		require.ErrorContains(t, err, "version is not valid")
		require.ErrorIs(t, err, ErrBadRequest)
	})
	t.Run("validate gitRef", func(t *testing.T) {
		obj := DevBuild{Spec: DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: CommunityEdition, GitRef: "pull/1234 "}}
		_, err := server.Create(context.TODO(), obj, DevBuildSaveOption{})
		require.ErrorContains(t, err, "gitRef is not valid")
		require.ErrorIs(t, err, ErrBadRequest)

		obj.Spec.GitRef = "pull/abcd"
		_, err = server.Create(context.TODO(), obj, DevBuildSaveOption{})
		require.ErrorContains(t, err, "gitRef is not valid")

		obj.Spec.GitRef = "abcde"
		_, err = server.Create(context.TODO(), obj, DevBuildSaveOption{})
		require.ErrorContains(t, err, "gitRef is not valid")

		obj.Spec.GitRef = "branch/tidb-6.5-with-kv-timeout-feature"
		_, err = server.Create(context.TODO(), obj, DevBuildSaveOption{})
		require.NoError(t, err)

		obj.Spec.GitRef = "release-6.1"
		_, err = server.Create(context.TODO(), obj, DevBuildSaveOption{})
		require.NoError(t, err)
	})

	t.Run("bad githubRepo", func(t *testing.T) {
		_, err := server.Create(context.TODO(), DevBuild{Spec: DevBuildSpec{Product: ProductTidb, GitRef: "pull/23",
			Version: "v6.1.2", Edition: CommunityEdition, GithubRepo: "aa/bb/cc"}}, DevBuildSaveOption{})
		require.ErrorContains(t, err, "githubRepo is not valid")
		require.ErrorIs(t, err, ErrBadRequest)
	})
	t.Run("hotfix ok", func(t *testing.T) {
		_, err := server.Create(context.TODO(), DevBuild{Spec: DevBuildSpec{Product: ProductTidb, GitRef: "branch/master",
			Version: "v6.1.2-20230102", Edition: CommunityEdition, IsHotfix: true}}, DevBuildSaveOption{})
		require.NoError(t, err)
	})
	t.Run("hotfix check", func(t *testing.T) {
		_, err := server.Create(context.TODO(), DevBuild{Spec: DevBuildSpec{Product: ProductTidb, GitRef: "pull/23",
			Version: "v6.1.2", Edition: CommunityEdition, IsHotfix: true}}, DevBuildSaveOption{})
		require.ErrorContains(t, err, "verion must be like v7.0.0-20230102... for hotfix")
		require.ErrorIs(t, err, ErrBadRequest)
	})
}

func TestDevBuildUpdate(t *testing.T) {
	mockedJenkins := &mockJenkins{}
	mockedRepo := mockRepo{}
	server := DevbuildServer{
		Repo:    &mockedRepo,
		Jenkins: mockedJenkins,
		Now:     time.Now,
	}

	t.Run("ok", func(t *testing.T) {
		mockedRepo.saved = DevBuild{Spec: DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", PluginGitRef: "master"}}
		entity, err := server.Update(context.TODO(), 1,
			DevBuild{Spec: DevBuildSpec{Product: ProductBr}, Status: DevBuildStatus{Status: BuildStatusSuccess}},
			DevBuildSaveOption{})
		require.NoError(t, err)
		require.Equal(t, BuildStatusSuccess, entity.Status.Status)
		require.Equal(t, ProductTidb, mockedRepo.saved.Spec.Product)
		require.Equal(t, BuildStatusSuccess, mockedRepo.saved.Status.Status)
	})
	t.Run("bad enterprise plugin", func(t *testing.T) {
		_, err := server.Update(context.TODO(),
			1,
			DevBuild{ID: 2},
			DevBuildSaveOption{})
		require.ErrorContains(t, err, "bad id")
		require.ErrorIs(t, err, ErrBadRequest)
	})
	t.Run("with report", func(t *testing.T) {
		req := `
	{
		"status": {
			"status": "SUCCESS",
			"pipelineBuildID": 13,
			"pipelineStartAt": "2023-01-16T20:11:54+08:00",
			"pipelineEndAt": "2023-01-16T20:11:57+08:00",
			"buildReport": {
				"gitHash": "c6ebcec4d7b4d379966bfeb8edd1ab67fc5346b9",
				"images": [
					{
						"platform": "multi-arch",
						"url": "hub.pingcap.net/devbuild/pd:v6.5.0-13"
					}
				],
				"binaries": [
					{
						"platform": "linux/amd64",
						"url": "builds/devbuild/13/pd-linux-amd64.tar.gz",
						"sha256URL": "builds/devbuild/13/pd-linux-amd64.tar.gz.sha256"
					},
					{
						"platform": "linux/arm64",
						"url": "builds/devbuild/13/pd-linux-arm64.tar.gz",
						"sha256URL": "builds/devbuild/13/pd-linux-arm64.tar.gz.sha256"
					}
				]
			}
		}
	}`
		obj := DevBuild{}
		json.Unmarshal([]byte(req), &obj)
		ent, err := server.Update(context.TODO(), 1, obj, DevBuildSaveOption{})
		require.NoError(t, err)
		require.NotNil(t, ent)
		require.Equal(t, int64(13), ent.Status.PipelineBuildID)
	})
}

func TestDevBuildGet(t *testing.T) {
	mockedRepo := mockRepo{}
	mockedJenkins := mockJenkins{}
	server := DevbuildServer{
		Repo:    &mockedRepo,
		Jenkins: &mockedJenkins,
		Now:     time.Now,
	}

	t.Run("ok", func(t *testing.T) {
		mockedRepo.saved = DevBuild{ID: 1,
			Spec:   DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", PluginGitRef: "master"},
			Status: DevBuildStatus{PipelineBuildID: 4}}
		entity, err := server.Get(context.TODO(), 1, DevBuildGetOption{})
		require.NoError(t, err)
		require.Equal(t, "https://cd.pingcap.net//blue/organizations/jenkins/devbuild/detail/devbuild/4/pipeline", entity.Status.PipelineViewURL)
	})
	t.Run("sync", func(t *testing.T) {
		mockedRepo.saved = DevBuild{ID: 1,
			Spec:   DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", PluginGitRef: "master"},
			Status: DevBuildStatus{PipelineBuildID: 4, Status: BuildStatusProcessing}}
		mockedJenkins.job = &gojenkins.Build{Raw: &gojenkins.BuildResponse{Result: "PROCCESSING", Timestamp: time.Unix(1, 0).UnixMilli()}}
		entity, err := server.Get(context.TODO(), 1, DevBuildGetOption{Sync: true})
		require.NoError(t, err)
		require.Equal(t, time.Unix(1, 0).Local(), *entity.Status.PipelineStartAt)
		require.Equal(t, "https://cd.pingcap.net//blue/organizations/jenkins/devbuild/detail/devbuild/4/pipeline", entity.Status.PipelineViewURL)
	})
}

func TestDevBuildRerun(t *testing.T) {
	mockedRepo := mockRepo{}
	server := DevbuildServer{
		Repo:    &mockedRepo,
		Jenkins: &mockJenkins{},
		Now:     time.Now,
	}
	mockedRepo.saved = DevBuild{ID: 2,
		Spec:   DevBuildSpec{Product: ProductTidb, Version: "v6.1.2", Edition: EnterpriseEdition, GitRef: "pull/23", PluginGitRef: "master"},
		Status: DevBuildStatus{PipelineBuildID: 4}}
	entity, err := server.Rerun(context.TODO(), 1, DevBuildSaveOption{})
	require.NoError(t, err)
	require.Equal(t, 1, entity.ID)
	require.Equal(t, "v6.1.2", entity.Spec.Version)
}