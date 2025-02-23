/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package peer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"

	"d7y.io/dragonfly/v2/client/config"
	"d7y.io/dragonfly/v2/client/daemon/metrics"
	"d7y.io/dragonfly/v2/client/daemon/storage"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/idgen"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/scheduler"
)

type FileTaskRequest struct {
	scheduler.PeerTaskRequest
	Output            string
	Limit             float64
	DisableBackSource bool
	Pattern           string
	Callsystem        string
}

// FileTask represents a peer task to download a file
type FileTask interface {
	Start(ctx context.Context) (chan *FileTaskProgress, error)
}

type fileTask struct {
	*logger.SugaredLoggerOnWith
	ctx               context.Context
	span              trace.Span
	peerTaskConductor *peerTaskConductor
	pieceCh           chan *pieceInfo

	request *FileTaskRequest
	// progressCh holds progress status
	progressCh     chan *FileTaskProgress
	progressStopCh chan bool

	// disableBackSource indicates not back source when failed
	disableBackSource bool
	pattern           string
	callsystem        string
}

type ProgressState struct {
	Success bool
	Code    base.Code
	Msg     string
}

type FileTaskProgress struct {
	State           *ProgressState
	TaskID          string
	PeerID          string
	ContentLength   int64
	CompletedLength int64
	PeerTaskDone    bool
	DoneCallback    func()
}

func (ptm *peerTaskManager) newFileTask(
	ctx context.Context,
	request *FileTaskRequest,
	limit rate.Limit) (context.Context, *fileTask, error) {
	metrics.FileTaskCount.Add(1)
	ptc, err := ptm.getPeerTaskConductor(ctx, idgen.TaskID(request.Url, request.UrlMeta), &request.PeerTaskRequest, limit)
	if err != nil {
		return nil, nil, err
	}
	// prefetch parent request
	if ptm.enablePrefetch && request.UrlMeta.Range != "" {
		go ptm.prefetch(&request.PeerTaskRequest)
	}
	ctx, span := tracer.Start(ctx, config.SpanFileTask, trace.WithSpanKind(trace.SpanKindClient))

	pt := &fileTask{
		SugaredLoggerOnWith: ptc.SugaredLoggerOnWith,
		ctx:                 ctx,
		span:                span,
		peerTaskConductor:   ptc,
		pieceCh:             ptc.broker.Subscribe(),
		request:             request,

		progressCh:        make(chan *FileTaskProgress),
		progressStopCh:    make(chan bool),
		disableBackSource: request.DisableBackSource,
		pattern:           request.Pattern,
		callsystem:        request.Callsystem,
	}
	return ctx, pt, nil
}

func (f *fileTask) Start(ctx context.Context) (chan *FileTaskProgress, error) {
	go f.syncProgress()
	// return a progress channel for request download progress
	return f.progressCh, nil
}

func (f *fileTask) syncProgress() {
	for {
		select {
		case <-f.peerTaskConductor.successCh:
			f.storeToOutput()
			return
		case <-f.peerTaskConductor.failCh:
			f.sendFailProgress(f.peerTaskConductor.failedCode, f.peerTaskConductor.failedReason)
			return
		case <-f.ctx.Done():
		case piece := <-f.pieceCh:
			if piece.finished {
				continue
			}
			pg := &FileTaskProgress{
				State: &ProgressState{
					Success: true,
					Code:    base.Code_Success,
					Msg:     "downloading",
				},
				TaskID:          f.peerTaskConductor.GetTaskID(),
				PeerID:          f.peerTaskConductor.GetPeerID(),
				ContentLength:   f.peerTaskConductor.GetContentLength(),
				CompletedLength: f.peerTaskConductor.completedLength.Load(),
				PeerTaskDone:    false,
			}

			select {
			case <-f.progressStopCh:
			case f.progressCh <- pg:
				f.Debugf("progress sent, %d/%d", pg.CompletedLength, pg.ContentLength)
			case <-f.ctx.Done():
				f.Warnf("send progress failed, file task context done due to %s", f.ctx.Err())
				return
			}
		}
	}
}

func (f *fileTask) storeToOutput() {
	err := f.peerTaskConductor.storageManager.Store(
		f.ctx,
		&storage.StoreRequest{
			CommonTaskRequest: storage.CommonTaskRequest{
				PeerID:      f.peerTaskConductor.GetPeerID(),
				TaskID:      f.peerTaskConductor.GetTaskID(),
				Destination: f.request.Output,
			},
			MetadataOnly: false,
			TotalPieces:  f.peerTaskConductor.GetTotalPieces(),
		})
	if err != nil {
		f.sendFailProgress(base.Code_ClientError, err.Error())
		return
	}
	f.sendSuccessProgress()
}

func (f *fileTask) sendSuccessProgress() {
	var progressDone bool
	pg := &FileTaskProgress{
		State: &ProgressState{
			Success: true,
			Code:    base.Code_Success,
			Msg:     "done",
		},
		TaskID:          f.peerTaskConductor.GetTaskID(),
		PeerID:          f.peerTaskConductor.GetPeerID(),
		ContentLength:   f.peerTaskConductor.GetContentLength(),
		CompletedLength: f.peerTaskConductor.completedLength.Load(),
		PeerTaskDone:    true,
		DoneCallback: func() {
			progressDone = true
			close(f.progressStopCh)
		},
	}
	// send progress
	select {
	case f.progressCh <- pg:
		f.Infof("finish progress sent")
	case <-f.ctx.Done():
		f.Warnf("finish progress sent failed, context done")
	}

	// wait progress stopped
	select {
	case <-f.progressStopCh:
		f.Infof("progress stopped")
	case <-f.ctx.Done():
		if progressDone {
			f.Debugf("progress stopped and context done")
		} else {
			f.Warnf("wait progress stopped failed, context done, but progress not stopped")
		}
	}
}

func (f *fileTask) sendFailProgress(code base.Code, msg string) {
	var progressDone bool
	pg := &FileTaskProgress{
		State: &ProgressState{
			Success: false,
			Code:    code,
			Msg:     msg,
		},
		TaskID:          f.peerTaskConductor.GetTaskID(),
		PeerID:          f.peerTaskConductor.GetPeerID(),
		ContentLength:   f.peerTaskConductor.GetContentLength(),
		CompletedLength: f.peerTaskConductor.completedLength.Load(),
		PeerTaskDone:    true,
		DoneCallback: func() {
			progressDone = true
			close(f.progressStopCh)
		},
	}

	// wait client received progress
	f.Infof("try to send unfinished progress, completed length: %d, state: (%t, %d, %s)",
		pg.CompletedLength, pg.State.Success, pg.State.Code, pg.State.Msg)
	select {
	case f.progressCh <- pg:
		f.Debugf("unfinished progress sent")
	case <-f.ctx.Done():
		f.Debugf("send unfinished progress failed, context done: %v", f.ctx.Err())
	}
	// wait progress stopped
	select {
	case <-f.progressStopCh:
		f.Infof("progress stopped")
	case <-f.ctx.Done():
		if progressDone {
			f.Debugf("progress stopped and context done")
		} else {
			f.Warnf("wait progress stopped failed, context done, but progress not stopped")
		}
	}
}
