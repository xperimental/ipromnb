package scaffold

import (
	"context"
	"errors"
	"fmt"
)

const executeQueueSize = 1 << 8

type executeQueueItem struct {
	req  *message
	sock *shellSocket
}

// executeQueue executes execute_requests sequentially.
// Although shellSocket in this package accepts request on a shell socket in parallel,
// we shold not handle multiple execute_requests in parallel because the jupyter client sends
// successive execute_requests to servers before previous execute_requests finishes.
// A jupyter kernel is responsible to handle multiple execute_requests sequentially and
// abort them if one of them fails.
type executeQueue struct {
	serverCtx  context.Context
	queue      chan *executeQueueItem
	iopub      *iopubSocket
	handlers   RequestHandlers
	currentCtx *contextAndCancel
}

func newExecuteQueue(ctx context.Context, iopub *iopubSocket, handlers RequestHandlers) *executeQueue {
	return &executeQueue{
		serverCtx: ctx,
		queue:     make(chan *executeQueueItem, executeQueueSize),
		iopub:     iopub,
		handlers:  handlers,
	}
}

func (q *executeQueue) push(req *message, sock *shellSocket) {
	q.queue <- &executeQueueItem{req, sock}
}

// abortQueue aborts requests in the queue.
// c.f. _abort_queue in https://github.com/ipython/ipykernel/blob/master/ipykernel/kernelbase.py
func (q *executeQueue) abortQueue() {
loop:
	for {
		var item *executeQueueItem
		select {
		case item = <-q.queue:
		default:
			break loop
		}
		err := q.iopub.WithOngoingContext(func(ctx context.Context) error {
			res := newMessageWithParent(item.req)
			res.Header.MsgType = "execute_reply"
			res.Content = &ExecuteResult{
				Status: "abort",
			}
			if err := item.sock.pushResult(res); err != nil {
				return fmt.Errorf("Failed to send execute_reply: %v", err)
			}
			return nil
		}, item.req)
		if err != nil {
			log.Errorf("Failed to abort a execute request: %v", err)
		}
	}
}

func (q *executeQueue) cancelCurrent() {
	cur := q.currentCtx
	if cur != nil {
		cur.cancel()
	}
}

// loop executes execute_requests sequentially.
func (q *executeQueue) loop() {
	var errStatusError = errors.New("execute status error")
loop:
	for {
		var item *executeQueueItem
		select {
		case item = <-q.queue:
		case <-q.serverCtx.Done():
			break loop
		}

		exReq := item.req.Content.(*ExecuteRequest)
		err := q.iopub.WithOngoingContext(func(ctx context.Context) error {
			cur, cancel := context.WithCancel(ctx)
			q.currentCtx = &contextAndCancel{cur, cancel}
			defer func() {
				cancel()
				q.currentCtx = nil
			}()
			result := q.handlers.HandleExecuteRequest(
				cur,
				exReq,
				func(name, text string) {
					q.iopub.sendStream(name, text, item.req)
				}, func(data *DisplayData, update bool) {
					q.iopub.sendDisplayData(data, item.req, update)
				})
			res := newMessageWithParent(item.req)
			res.Header.MsgType = "execute_reply"
			res.Content = &result
			if err := item.sock.pushResult(res); err != nil {
				log.Errorf("Failed to send execute_reply: %v", err)
			}
			if result.Status == "error" {
				return errStatusError
			}
			return nil
		}, item.req)
		if err != nil {
			if err != errStatusError {
				log.Errorf("Failed to handle a execute_request: %v", err)
			}
			if exReq.StopOnError {
				q.abortQueue()
			}
		}
	}
}
