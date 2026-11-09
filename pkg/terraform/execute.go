package terraform

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"get.porter.sh/porter/pkg/exec/builder"
)

// executeWithRetry runs the action's step, rerunning it on failure as retry allows.
func (m *Mixin) executeWithRetry(ctx context.Context, action *Action, retry *Retry) error {
	attempts, delay := 1, time.Duration(0)
	if retry != nil {
		attempts, delay = retry.Attempts, retry.Delay
	}
	giveUp := func(attempt int, err error) error {
		if errors.Is(err, ctx.Err()) {
			return fmt.Errorf("giving up after attempt %d failed: %w", attempt, err)
		}
		return fmt.Errorf("giving up after attempt %d failed: %w: %w", attempt, err, ctx.Err())
	}

	for attempt := 1; ; attempt++ {
		_, err := builder.ExecuteSingleStepAction(ctx, m.RuntimeConfig, action)
		if err == nil || attempts == 1 {
			return err
		}
		if ctx.Err() != nil {
			return giveUp(attempt, err)
		}
		if attempt >= attempts {
			return fmt.Errorf("giving up after %d attempts: %w", attempts, err)
		}
		// Only a command that ran and failed on its own can pass on a rerun; a signal means something stopped it
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() == -1 {
			return err
		}

		wait := ""
		if delay > 0 {
			wait = " in " + delay.String()
		}
		fmt.Fprintf(m.Err, "Attempt %d of %d failed, retrying%s...\n", attempt, attempts, wait)
		select {
		case <-ctx.Done():
			return giveUp(attempt, err)
		case <-time.After(delay):
		}
	}
}
