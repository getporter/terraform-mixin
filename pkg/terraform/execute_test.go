package terraform

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_ExecuteWithRetry(t *testing.T) {
	actions := []struct {
		name string
		key  string
		verb string
		cmd  string
		run  func(h *TestMixin) error
	}{
		{"install", "install", "apply", "terraform apply -auto-approve -input=false -var myvar=foo",
			func(h *TestMixin) error { return h.Install(context.Background()) }},
		{"upgrade", "upgrade", "apply", "terraform apply -auto-approve -input=false -var myvar=foo",
			func(h *TestMixin) error { return h.Upgrade(context.Background()) }},
		{"uninstall", "uninstall", "destroy", "terraform destroy -auto-approve -input=false -var myvar=foo",
			func(h *TestMixin) error { return h.Uninstall(context.Background()) }},
		{"invoke", "custom", "custom", "terraform custom -var myvar=foo",
			func(h *TestMixin) error { return h.Invoke(context.Background(), InvokeOptions{}) }},
	}
	testcases := []struct {
		name       string
		retry      string
		failRuns   int
		wantRuns   int
		wantErr    string
		wantNotice string
	}{
		{"no retry by default", "", 1, 1, "error running command", ""},
		{"a single attempt never retries", "retry:\n      attempts: 1\n", 1, 1, "error running command", ""},
		{"passes on a later attempt", "retry:\n      attempts: 3\n      delay: 1ms\n", 2, 3, "", "Attempt 2 of 3 failed"},
		{"fails when attempts are exhausted", "retry:\n      attempts: 3\n      delay: 1ms\n", 3, 3, "giving up after 3 attempts", "Attempt 2 of 3 failed"},
	}

	for _, a := range actions {
		for _, tc := range testcases {
			t.Run(a.name+"/"+tc.name, func(t *testing.T) {
				input := a.key + ":\n- terraform:\n    description: x\n    arguments:\n      - " + a.verb +
					"\n    backendConfig:\n      key: my.tfstate\n    vars:\n      myvar: foo\n    " + tc.retry
				h := newTestMixinWithInput(t, []byte(input), a.cmd)
				inits, runs := 0, 0
				onCommand(h, "init", func(n int, cmd *exec.Cmd) { inits = n })
				onCommand(h, a.verb, func(n int, cmd *exec.Cmd) {
					runs = n
					if n <= tc.failRuns {
						cmd.Env = append(cmd.Env, test.ExpectedCommandExitCodeEnv+"=1")
					}
				})

				err := a.run(h)
				if tc.wantErr == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, tc.wantErr)
					var exitErr *exec.ExitError
					assert.ErrorAs(t, err, &exitErr, "the command's exit error should stay in the chain")
				}
				assert.Equal(t, tc.wantRuns, runs)
				assert.Equal(t, 1, inits, "init should not be rerun")
				if tc.wantNotice == "" {
					assert.NotContains(t, h.TestContext.GetError(), "Attempt ")
				} else {
					assert.Contains(t, h.TestContext.GetError(), tc.wantNotice)
				}
			})
		}
	}
}

func TestMixin_ExecuteWithRetry_NoRetryWhenCommandCannotStart(t *testing.T) {
	t.Run("on the first attempt", func(t *testing.T) {
		h := newInstallTestMixin(t, "testdata/install-input-retry.yaml")
		applies := 0
		onApply(h, func(n int, cmd *exec.Cmd) {
			applies = n
			cmd.Path = "/nonexistent/terraform"
		})

		err := h.Install(context.Background())
		require.ErrorContains(t, err, "couldn't run command")
		assert.NotContains(t, err.Error(), "giving up")
		assert.Equal(t, 1, applies)
		assert.NotContains(t, h.TestContext.GetError(), "Attempt ")
	})

	t.Run("on the last attempt", func(t *testing.T) {
		h := newInstallTestMixin(t, "testdata/install-input-retry.yaml")
		applies := 0
		onApply(h, func(n int, cmd *exec.Cmd) {
			applies = n
			if n < 3 {
				cmd.Env = append(cmd.Env, test.ExpectedCommandExitCodeEnv+"=1")
			} else {
				cmd.Path = "/nonexistent/terraform"
			}
		})

		err := h.Install(context.Background())
		require.ErrorContains(t, err, "giving up after 3 attempts")
		assert.ErrorContains(t, err, "couldn't run command")
		assert.Equal(t, 3, applies)
	})
}

func TestMixin_ExecuteWithRetry_NoRetryWhenKilledBySignal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("needs a POSIX shell to kill the command")
	}
	h := newInstallTestMixin(t, "testdata/install-input-retry.yaml")
	applies := 0
	onApply(h, func(n int, cmd *exec.Cmd) {
		applies = n
		cmd.Path = "/bin/sh"
		cmd.Args = []string{"sh", "-c", "kill -KILL $$"}
	})

	err := h.Install(context.Background())
	require.ErrorContains(t, err, "signal: killed")
	assert.Equal(t, 1, applies)
	assert.NotContains(t, h.TestContext.GetError(), "Attempt ")
}

func TestMixin_ExecuteWithRetry_StopsWhenCancelled(t *testing.T) {
	t.Run("before an attempt", func(t *testing.T) {
		h := newInstallTestMixin(t, "testdata/install-input-retry.yaml")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		applies := 0
		onApply(h, func(n int, cmd *exec.Cmd) {
			applies = n
			cancel()
		})

		err := h.Install(ctx)
		require.ErrorIs(t, err, context.Canceled)
		assert.ErrorContains(t, err, "giving up after attempt 1 failed")
		assert.Equal(t, 1, strings.Count(err.Error(), "context canceled"), "the cause should be named once")
		assert.Equal(t, 1, applies)
	})

	t.Run("during the delay", func(t *testing.T) {
		h := newInstallTestMixin(t, "testdata/install-input-retry-long-delay.yaml")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		applies := 0
		onApply(h, func(n int, cmd *exec.Cmd) {
			applies = n
			cmd.Env = append(cmd.Env, test.ExpectedCommandExitCodeEnv+"=1")
		})
		// Cancel as the retry notice is printed, which is right before the wait starts
		stderr := h.Err
		h.Err = writerFunc(func(p []byte) (int, error) {
			if bytes.Contains(p, []byte("retrying")) {
				cancel()
			}
			return stderr.Write(p)
		})

		start := time.Now()
		err := h.Install(ctx)
		require.ErrorIs(t, err, context.Canceled)
		assert.ErrorContains(t, err, "giving up after attempt 1 failed")
		assert.Equal(t, 1, applies)
		assert.Less(t, time.Since(start), 5*time.Second, "the 10s delay should be cut short")
	})
}

func TestMixin_ExecuteWithRetry_StopsWhenStateWasNotSaved(t *testing.T) {
	testcases := []struct {
		name        string
		attempts    string
		unsavedOn   int
		wantApplies int
		wantErr     string
	}{
		{"on the first attempt", "3", 1, 1, "attempt 1 of 3 failed and terraform could not save its state, which it wrote to errored.tfstate"},
		{"on the last attempt", "3", 3, 3, "attempt 3 of 3 failed and terraform could not save its state, which it wrote to errored.tfstate"},
		{"on the only attempt", "1", 1, 1, "attempt 1 of 1 failed and terraform could not save its state, which it wrote to errored.tfstate"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			input := "install:\n- terraform:\n    description: x\n    backendConfig:\n      key: my.tfstate\n    vars:\n      myvar: foo\n    retry:\n      attempts: " + tc.attempts + "\n"
			h := newTestMixinWithInput(t, []byte(input), "terraform apply -auto-approve -input=false -var myvar=foo")
			applies := 0
			onApply(h, func(n int, cmd *exec.Cmd) {
				applies = n
				cmd.Env = append(cmd.Env, test.ExpectedCommandExitCodeEnv+"=1")
				if n == tc.unsavedOn {
					require.NoError(t, h.FileSystem.WriteFile(erroredStateFilename, []byte("{}"), 0600))
				}
			})

			err := h.Install(context.Background())
			require.ErrorContains(t, err, tc.wantErr)
			var exitErr *exec.ExitError
			assert.ErrorAs(t, err, &exitErr, "the command's exit error should stay in the chain")
			assert.Equal(t, tc.wantApplies, applies)
			assert.Equal(t, tc.wantApplies-1, strings.Count(h.TestContext.GetError(), "Attempt "), "only the attempts before the unsaved state should be retried")
		})
	}
}

func TestMixin_ExecuteWithRetry_StopsWhenStateCheckFails(t *testing.T) {
	h := newInstallTestMixin(t, "testdata/install-input-retry.yaml")
	h.FileSystem.Afero = &afero.Afero{Fs: statFailsFs{h.FileSystem.Afero.Fs}}
	applies := 0
	onApply(h, func(n int, cmd *exec.Cmd) {
		applies = n
		cmd.Env = append(cmd.Env, test.ExpectedCommandExitCodeEnv+"=1")
	})

	err := h.Install(context.Background())
	require.ErrorContains(t, err, "attempt 1 of 3 failed and looking for errored.tfstate failed too")
	assert.ErrorContains(t, err, "stat refused")
	assert.Equal(t, 1, applies)
	assert.NotContains(t, h.TestContext.GetError(), "Attempt ")
}

// statFailsFs fails every stat of the errored state file.
type statFailsFs struct{ afero.Fs }

func (f statFailsFs) Stat(name string) (os.FileInfo, error) {
	if strings.HasSuffix(name, erroredStateFilename) {
		return nil, errors.New("stat refused")
	}
	return f.Fs.Stat(name)
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// newInstallTestMixin prepares a mixin that reads the given install step and expects init then apply.
func newInstallTestMixin(t *testing.T, input string) *TestMixin {
	b, err := os.ReadFile(input)
	require.NoError(t, err)
	return newTestMixinWithInput(t, b, "terraform apply -auto-approve -input=false -var myvar=foo")
}

// newTestMixinWithInput prepares a mixin that reads input and expects init then cmd.
func newTestMixinWithInput(t *testing.T, input []byte, cmd string) *TestMixin {
	t.Setenv(test.ExpectedCommandEnv, strings.Join([]string{
		"terraform init -backend=true -backend-config=key=my.tfstate -reconfigure",
		cmd,
	}, "\n"))

	h := NewTestMixin(t)
	h.In = bytes.NewReader(input)
	h.config.WorkingDir = h.Getwd()
	return h
}

// onApply calls fn with each terraform apply command before it runs, numbered from 1.
func onApply(h *TestMixin, fn func(n int, cmd *exec.Cmd)) {
	onCommand(h, "apply", fn)
}

// onCommand calls fn with each terraform command whose first argument is verb before it runs, numbered from 1.
func onCommand(h *TestMixin, verb string, fn func(n int, cmd *exec.Cmd)) {
	runs := 0
	newCommand := h.NewCommand
	h.NewCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := newCommand(ctx, name, args...)
		if len(args) > 0 && args[0] == verb {
			runs++
			fn(runs, cmd)
		}
		return cmd
	}
}
