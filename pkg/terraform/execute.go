package terraform

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"get.porter.sh/porter/pkg/exec/builder"
)

// executeAction runs the action's step, rerunning it on failure as the step's retry block allows.
func (m *Mixin) executeAction(ctx context.Context, action *Action) error {
	attempts, delay := 1, time.Duration(0)
	if retry := action.Steps[0].Retry; retry != nil {
		attempts, delay = retry.Attempts, retry.Delay
	}
	giveUp := func(attempt int, err error) error {
		return fmt.Errorf("%w, giving up after attempt %d failed: %w", ctx.Err(), attempt, err)
	}

	for attempt := 1; ; attempt++ {
		_, err := builder.ExecuteSingleStepAction(ctx, m.RuntimeConfig, action)
		if err == nil || attempts == 1 {
			return err
		}
		if ctx.Err() != nil {
			return giveUp(attempt, err)
		}
		// Only a command that ran and failed can pass on a rerun
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
		if attempt >= attempts {
			return fmt.Errorf("giving up after %d attempts: %w", attempts, err)
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
