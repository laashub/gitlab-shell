package uploadarchive

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet/testserver"
	"gitlab.com/gitlab-org/gitlab-shell/internal/testhelper"
	"gitlab.com/gitlab-org/gitlab-shell/internal/testhelper/requesthandlers"
)

func TestUploadPack(t *testing.T) {
	gitalyAddress, _, cleanup := testserver.StartGitalyServer(t)
	defer cleanup()

	requests := requesthandlers.BuildAllowedWithGitalyHandlers(t, gitalyAddress)
	url, cleanup := testserver.StartHttpServer(t, requests)
	defer cleanup()

	output := &bytes.Buffer{}
	input := &bytes.Buffer{}

	userId := "1"
	repo := "group/repo"

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		Args:       &commandargs.Shell{GitlabKeyId: userId, CommandType: commandargs.UploadArchive, SshArgs: []string{"git-upload-archive", repo}},
		ReadWriter: &readwriter.ReadWriter{ErrOut: output, Out: output, In: input},
	}

	hook := testhelper.SetupLogger()

	err := cmd.Execute()
	require.NoError(t, err)

	require.Equal(t, "UploadArchive: "+repo, output.String())
	require.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	require.True(t, strings.Contains(hook.LastEntry().Message, "executing git command"))
	require.True(t, strings.Contains(hook.LastEntry().Message, "command=git-upload-archive"))
}
