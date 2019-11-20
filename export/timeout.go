package export

import "context"

func (c *Client) SyncSendWithTimeout(ctx context.Context, operator string, param []byte, callback RequestStatusCallback) error {
	ch := make(chan error, 1)
	go func() {
		ch <- c.AsyncSend(operator, param, callback)
	}()

	for {
		select {
		case err := <-ch:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
