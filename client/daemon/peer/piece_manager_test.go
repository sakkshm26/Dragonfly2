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
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	testifyassert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"d7y.io/dragonfly/v2/client/clientutil"
	"d7y.io/dragonfly/v2/client/config"
	"d7y.io/dragonfly/v2/client/daemon/storage"
	"d7y.io/dragonfly/v2/client/daemon/test"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	_ "d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/server"
	"d7y.io/dragonfly/v2/pkg/rpc/scheduler"
	"d7y.io/dragonfly/v2/pkg/source"
	"d7y.io/dragonfly/v2/pkg/source/httpprotocol"
)

func TestPieceManager_DownloadSource(t *testing.T) {
	assert := testifyassert.New(t)
	ctrl := gomock.NewController(t)
	source.UnRegister("http")
	require.Nil(t, source.Register("http", httpprotocol.NewHTTPSourceClient(), httpprotocol.Adapter))
	defer source.UnRegister("http")
	testBytes, err := os.ReadFile(test.File)
	assert.Nil(err, "load test file")

	var (
		peerID = "peer0"
		taskID = "task0"
		output = "../test/testdata/test.output"
	)

	pieceDownloadTimeout := 30 * time.Second
	storageManager, _ := storage.NewStorageManager(
		config.SimpleLocalTaskStoreStrategy,
		&config.StorageOption{
			DataPath: t.TempDir(),
			TaskExpireTime: clientutil.Duration{
				Duration: -1 * time.Second,
			},
		}, func(request storage.CommonTaskRequest) {})

	hash := md5.New()
	hash.Write(testBytes)
	digest := hex.EncodeToString(hash.Sum(nil)[:16])

	testCases := []struct {
		name              string
		pieceSize         uint32
		withContentLength bool
		checkDigest       bool
	}{
		{
			name:              "multiple pieces with content length, check digest",
			pieceSize:         1024,
			checkDigest:       true,
			withContentLength: true,
		},
		{
			name:              "multiple pieces with content length",
			pieceSize:         1024,
			checkDigest:       false,
			withContentLength: true,
		},
		{
			name:              "multiple pieces without content length, check digest",
			pieceSize:         1024,
			checkDigest:       true,
			withContentLength: false,
		},
		{
			name:              "multiple pieces without content length",
			pieceSize:         1024,
			checkDigest:       false,
			withContentLength: false,
		},
		{
			name:              "one pieces with content length case 1",
			pieceSize:         uint32(len(testBytes)),
			withContentLength: true,
		},
		{
			name:              "one pieces without content length case 1",
			pieceSize:         uint32(len(testBytes)),
			withContentLength: false,
		},
		{
			name:              "one pieces with content length case 2",
			pieceSize:         uint32(len(testBytes)) + 1,
			withContentLength: true,
		},
		{
			name:              "one pieces without content length case 2",
			pieceSize:         uint32(len(testBytes)) + 1,
			withContentLength: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/********** prepare test start **********/
			mockPeerTask := NewMockTask(ctrl)
			var (
				totalPieces = &atomic.Int32{}
				taskStorage storage.TaskStorageDriver
			)
			mockPeerTask.EXPECT().SetContentLength(gomock.Any()).AnyTimes().DoAndReturn(
				func(arg0 int64) error {
					return nil
				})
			mockPeerTask.EXPECT().SetTotalPieces(gomock.Any()).AnyTimes().DoAndReturn(
				func(arg0 int32) {
					totalPieces.Store(arg0)
				})
			mockPeerTask.EXPECT().GetTotalPieces().AnyTimes().DoAndReturn(
				func() int32 {
					return totalPieces.Load()
				})
			mockPeerTask.EXPECT().GetPeerID().AnyTimes().DoAndReturn(
				func() string {
					return peerID
				})
			mockPeerTask.EXPECT().GetTaskID().AnyTimes().DoAndReturn(
				func() string {
					return taskID
				})
			mockPeerTask.EXPECT().GetStorage().AnyTimes().DoAndReturn(
				func() storage.TaskStorageDriver {
					return taskStorage
				})
			mockPeerTask.EXPECT().AddTraffic(gomock.Any()).AnyTimes().DoAndReturn(func(int642 uint64) {})
			mockPeerTask.EXPECT().ReportPieceResult(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
				func(*DownloadPieceRequest, *DownloadPieceResult, error) {

				})
			mockPeerTask.EXPECT().PublishPieceInfo(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
				func(pieceNum int32, size uint32) {

				})
			mockPeerTask.EXPECT().Context().AnyTimes().DoAndReturn(func() context.Context {
				return context.Background()
			})
			mockPeerTask.EXPECT().Log().AnyTimes().DoAndReturn(func() *logger.SugaredLoggerOnWith {
				return logger.With("test case", tc.name)
			})
			taskStorage, err = storageManager.RegisterTask(context.Background(),
				storage.RegisterTaskRequest{
					CommonTaskRequest: storage.CommonTaskRequest{
						PeerID:      mockPeerTask.GetPeerID(),
						TaskID:      mockPeerTask.GetTaskID(),
						Destination: output,
					},
					ContentLength: int64(len(testBytes)),
				})
			assert.Nil(err)
			defer storageManager.CleanUp()
			defer os.Remove(output)
			/********** prepare test end **********/
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.withContentLength {
					w.Header().Set("Content-Length",
						fmt.Sprintf("%d", len(testBytes)))
				}
				n, err := io.Copy(w, bytes.NewBuffer(testBytes))
				assert.Nil(err)
				assert.Equal(int64(len(testBytes)), n)
			}))
			defer ts.Close()

			pm, err := NewPieceManager(storageManager, pieceDownloadTimeout)
			assert.Nil(err)
			pm.(*pieceManager).computePieceSize = func(length int64) uint32 {
				return tc.pieceSize
			}

			request := &scheduler.PeerTaskRequest{
				Url: ts.URL,
				UrlMeta: &base.UrlMeta{
					Digest: "",
					Range:  "",
					Header: nil,
				},
			}
			if tc.checkDigest {
				request.UrlMeta.Digest = digest
			}

			err = pm.DownloadSource(context.Background(), mockPeerTask, request)
			assert.Nil(err)

			err = storageManager.Store(context.Background(),
				&storage.StoreRequest{
					CommonTaskRequest: storage.CommonTaskRequest{
						PeerID:      peerID,
						TaskID:      taskID,
						Destination: output,
					},
				})
			assert.Nil(err)

			outputBytes, err := os.ReadFile(output)
			assert.Nil(err, "load output file")
			assert.Equal(testBytes, outputBytes, "output and desired output must match")
		})
	}
}
