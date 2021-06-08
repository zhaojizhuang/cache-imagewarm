package docker

import (
	"context"
	"fmt"
	"knative.dev/pkg/logging"
	"sync"
	"time"

	dockermessage "github.com/docker/docker/pkg/jsonmessage"
)

const (
	// defaultTimeout is the default timeout of short running docker operations.
	// Value is slightly offset from 2 minutes to make timeouts due to this
	// constant recognizable.
	defaultTimeout = 2*time.Minute - 1*time.Second

	// defaultShmSize is the default ShmSize to use (in bytes) if not specified.
	defaultShmSize = int64(1024 * 1024 * 64)

	// defaultImagePullingProgressReportInterval is the default interval of image pulling progress reporting.
	defaultImagePullingProgressReportInterval = 10 * time.Second
)

// progress is a wrapper of dockermessage.JSONMessage with a lock protecting it.
type progress struct {
	sync.RWMutex
	// message stores the latest docker json message.
	message *dockermessage.JSONMessage
	// timestamp of the latest update.
	timestamp time.Time
}

func newProgress() *progress {
	return &progress{timestamp: time.Now()}
}

func (p *progress) set(msg *dockermessage.JSONMessage) {
	p.Lock()
	defer p.Unlock()
	p.message = msg
	p.timestamp = time.Now()
}

func (p *progress) get() (string, time.Time) {
	p.RLock()
	defer p.RUnlock()
	if p.message == nil {
		return "No progress", p.timestamp
	}
	// The following code is based on JSONMessage.Display
	var prefix string
	if p.message.ID != "" {
		prefix = fmt.Sprintf("%s: ", p.message.ID)
	}
	if p.message.Progress == nil {
		return fmt.Sprintf("%s%s", prefix, p.message.Status), p.timestamp
	}
	return fmt.Sprintf("%s%s %s", prefix, p.message.Status, p.message.Progress.String()), p.timestamp
}

// progressReporter keeps the newest image pulling progress and periodically report the newest progress.
type progressReporter struct {
	contex context.Context
	*progress
	image                     string
	cancel                    context.CancelFunc
	stopCh                    chan struct{}
	imagePullProgressDeadline time.Duration
}

// newProgressReporter creates a new progressReporter for specific image with specified reporting interval
func newProgressReporter(ctx context.Context, image string, cancel context.CancelFunc, imagePullProgressDeadline time.Duration) *progressReporter {
	return &progressReporter{
		contex:                    ctx,
		progress:                  newProgress(),
		image:                     image,
		cancel:                    cancel,
		stopCh:                    make(chan struct{}),
		imagePullProgressDeadline: imagePullProgressDeadline,
	}
}

// start starts the progressReporter
func (p *progressReporter) start() {
	logger := logging.FromContext(p.contex)
	go func() {

		ticker := time.NewTicker(defaultImagePullingProgressReportInterval)
		defer ticker.Stop()
		for {
			// TODO(random-liu): Report as events.
			select {
			case <-ticker.C:
				progress, timestamp := p.progress.get()
				// If there is no progress for p.imagePullProgressDeadline, cancel the operation.
				if time.Since(timestamp) > p.imagePullProgressDeadline {
					logger.Errorf("Cancel pulling image %q because of no progress for %v, latest progress: %q", p.image, p.imagePullProgressDeadline, progress)
					p.cancel()
					return
				}
				logger.Infof("Pulling image %q: %q", p.image, progress)
			case <-p.stopCh:
				progress, _ := p.progress.get()
				p.cancel()
				logger.Infof("Complete pull image %q: %q", p.image, progress)
				return
			}
		}
	}()
}

// stop stops the progressReporter
func (p *progressReporter) stop() {
	close(p.stopCh)
}
