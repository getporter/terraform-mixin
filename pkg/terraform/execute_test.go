package terraform

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestMixin_UnmarshalRetry(t *testing.T) {
	b, err := os.ReadFile("testdata/install-input-retry.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)

	assert.Equal(t, &Retry{Attempts: 3, Delay: time.Millisecond}, action.Steps[0].Retry)
}

func TestMixin_UnmarshalRetry_Invalid(t *testing.T) {
	testcases := []struct {
		name    string
		retry   string
		wantErr string
	}{
		{"attempts missing", "delay: 30s", "retry.attempts is required"},
		{"attempts zero", "attempts: 0", "retry.attempts must be at least 1, got 0"},
		{"delay without unit", "attempts: 2\n      delay: 30", `retry.delay must be a duration such as 30s, got "30"`},
		{"delay negative", "attempts: 2\n      delay: -30s", `retry.delay must not be negative, got "-30s"`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			input := "install:\n- terraform:\n    description: x\n    retry:\n      " + tc.retry + "\n"
			var action Action
			err := yaml.Unmarshal([]byte(input), &action)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestMixin_ExecuteAction_Retry(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		failApplies int
		wantApplies int
		wantErr     string
		wantNotice  string
	}{
		{"no retry by default", "testdata/install-input.yaml", 1, 1, "error running command", ""},
		{"passes on a later attempt", "testdata/install-input-retry.yaml", 2, 3, "", "Attempt 2 of 3 failed"},
		{"fails when attempts are exhausted", "testdata/install-input-retry.yaml", 3, 3, "giving up after 3 attempts", "Attempt 2 of 3 failed"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			h := newInstallTestMixin(t, tc.input)
			applies := 0
			onApply(h, func(n int, cmd *exec.Cmd) {
				applies = n
				if n <= tc.failApplies {
					cmd.Env = append(cmd.Env, test.ExpectedCommandExitCodeEnv+"=1")
				}
			})

			err := h.Install(context.Background())
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
				var exitErr *exec.ExitError
				assert.ErrorAs(t, err, &exitErr, "the command's exit error should stay in the chain")
			}
			assert.Equal(t, tc.wantApplies, applies)
			if tc.wantNotice == "" {
				assert.NotContains(t, h.TestContext.GetError(), "Attempt ")
			} else {
				assert.Contains(t, h.TestContext.GetError(), tc.wantNotice)
			}
		})
	}
}

func TestMixin_ExecuteAction_NoRetryWhenCommandCannotStart(t *testing.T) {
	h := newInstallTestMixin(t, "testdata/install-input-retry.yaml")
	applies := 0
	onApply(h, func(n int, cmd *exec.Cmd) {
		applies = n
		cmd.Path = "/nonexistent/terraform"
	})

	err := h.Install(context.Background())
	require.ErrorContains(t, err, "couldn't run command")
	assert.Equal(t, 1, applies)
	assert.NotContains(t, h.TestContext.GetError(), "Attempt ")
}

func TestMixin_ExecuteAction_RetryStopsWhenCancelled(t *testing.T) {
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

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// newInstallTestMixin prepares a mixin that reads the given install step and expects init then apply.
func newInstallTestMixin(t *testing.T, input string) *TestMixin {
	t.Setenv(test.ExpectedCommandEnv, strings.Join([]string{
		"terraform init -backend=true -backend-config=key=my.tfstate -reconfigure",
		"terraform apply -auto-approve -input=false -var myvar=foo",
	}, "\n"))

	b, err := os.ReadFile(input)
	require.NoError(t, err)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)
	h.config.WorkingDir = h.Getwd()
	return h
}

// onApply calls fn with each terraform apply command before it runs, numbered from 1.
func onApply(h *TestMixin, fn func(n int, cmd *exec.Cmd)) {
	applies := 0
	newCommand := h.NewCommand
	h.NewCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := newCommand(ctx, name, args...)
		if len(args) > 0 && args[0] == "apply" {
			applies++
			fn(applies, cmd)
		}
		return cmd
	}
}
