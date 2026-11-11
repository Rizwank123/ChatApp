package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/domain"
)

// MediaProcessor handles asynchronous image dimension extraction and thumbnailing.
type MediaProcessor struct {
	storage   domain.StorageProvider
	mediaRepo domain.MediaRepository
	queue     chan uuid.UUID
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewMediaProcessor(storage domain.StorageProvider, mediaRepo domain.MediaRepository) *MediaProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	mp := &MediaProcessor{
		storage:   storage,
		mediaRepo: mediaRepo,
		queue:     make(chan uuid.UUID, 500),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start 2 worker goroutines
	for i := 0; i < 2; i++ {
		mp.wg.Add(1)
		go mp.worker()
	}

	return mp
}

func (mp *MediaProcessor) Enqueue(mediaID uuid.UUID) {
	select {
	case mp.queue <- mediaID:
	default:
		log.Printf("[MediaProcessor] queue full, dropping immediate processing for %s", mediaID)
	}
}

func (mp *MediaProcessor) Close() {
	mp.cancel()
	close(mp.queue)
	mp.wg.Wait()
}

func (mp *MediaProcessor) worker() {
	defer mp.wg.Done()
	for {
		select {
		case <-mp.ctx.Done():
			return
		case mediaID, ok := <-mp.queue:
			if !ok {
				return
			}
			mp.processMedia(mediaID)
		}
	}
}

func (mp *MediaProcessor) processMedia(mediaID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	media, err := mp.mediaRepo.GetMediaByID(ctx, mediaID)
	if err != nil {
		return
	}

	if strings.ToUpper(media.MediaType) == "IMAGE" {
		data, err := mp.storage.GetObject(ctx, media.Bucket, media.ObjectKey)
		if err != nil {
			return
		}

		// Decode dimensions
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		width := cfg.Width
		height := cfg.Height

		var thumbKey string
		// Generate thumbnail
		img, _, err := image.Decode(bytes.NewReader(data))
		if err == nil && img != nil {
			thumb := simpleResize(img, 320, 320)
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 80}); err == nil {
				thumbKey = fmt.Sprintf("social-media/users/%s/thumbnails/%s.jpg", media.OwnerID.String(), media.ID.String())
				_ = mp.storage.PutObject(ctx, media.Bucket, thumbKey, buf.Bytes(), "image/jpeg")
			}
		}

		_ = mp.mediaRepo.UpdatePostMediaStatus(ctx, media.ID, "READY", thumbKey, width, height, 0)
	}
}

// simpleResize performs nearest-neighbor scaling for fast thumbnail generation without CGO dependencies.
func simpleResize(src image.Image, maxW, maxH int) image.Image {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return src
	}

	scaleW := float64(maxW) / float64(w)
	scaleH := float64(maxH) / float64(h)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}
	if scale >= 1.0 {
		return src
	}

	targetW := int(float64(w) * scale)
	targetH := int(float64(h) * scale)
	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	for y := 0; y < targetH; y++ {
		srcY := int(float64(y) / scale)
		for x := 0; x < targetW; x++ {
			srcX := int(float64(x) / scale)
			dst.Set(x, y, src.At(b.Min.X+srcX, b.Min.Y+srcY))
		}
	}
	return dst
}
