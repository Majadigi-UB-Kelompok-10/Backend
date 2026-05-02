package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/farildzaky/siskaperbapo-service/internal/cache"
	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HargaJob struct {
	BahanPokokID int32
	Harga        int32
	Tanggal      pgtype.Date
	AreaID       int32
	Retries      int
}

type HargaWorkerPool struct {
	jobs    chan HargaJob
	queries *db.Queries
	wg      sync.WaitGroup
}


func Start(dbPool *pgxpool.Pool) *HargaWorkerPool {
	queries := db.New(dbPool)

	pool := NewHargaWorkerPool(queries, 5, 1000)

	go runDailyCron(pool, queries)

	slog.Info("worker.started", slog.Int("workers", 5), slog.Int("buffer", 1000))
	return pool
}

// runDailyCron adalah jam weker yang akan bangun setiap jam 23:50 malam
func runDailyCron(pool *HargaWorkerPool, queries *db.Queries) {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 23, 50, 0, 0, now.Location())

		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		sleepDuration := time.Until(nextRun)
		slog.Info("worker.cron_scheduled", slog.String("next_run", nextRun.Format(time.RFC3339)))

		time.Sleep(sleepDuration)

		slog.Info("worker.cron_executing", slog.String("task", "auto_carry_over"))
		executeAutoCarryOver(pool, queries)
	}
}


func executeAutoCarryOver(pool *HargaWorkerPool, queries *db.Queries) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	today := time.Now().Truncate(24 * time.Hour)
	kemarin := today.AddDate(0, 0, -1)

	areas, err := queries.GetAllArea(ctx)
	if err != nil {
		slog.Error("worker.cron_error_get_areas", slog.String("error", err.Error()))
		return
	}
	areaMap := make(map[string]int32)
	for _, a := range areas {
		areaMap[a.Slug] = a.ID
	}

	bahanList, err := queries.GetAllBahanPokok(ctx, db.GetAllBahanPokokParams{
		Keyword:    pgtype.Text{Valid: false},
		LimitData:  1000,
		OffsetData: 0,
	})
	if err != nil {
		slog.Error("worker.cron_error_get_bahan", slog.String("error", err.Error()))
		return
	}

	kemarinPg := pgtype.Date{Time: kemarin, Valid: true}
	hariIniPg := pgtype.Date{Time: today, Valid: true}
	jobsQueued := 0

	for _, bahan := range bahanList {
		hargaKemarinList, err := queries.GetDaftarHargaArea(ctx, db.GetDaftarHargaAreaParams{
			BahanPokokID: bahan.ID,
			Tanggal:      kemarinPg,
		})
		if err != nil {
			continue 
		}

		for _, hk := range hargaKemarinList {
			areaID, exists := areaMap[hk.AreaSlug]
			if !exists {
				continue
			}

			pool.Enqueue(HargaJob{
				BahanPokokID: bahan.ID,
				AreaID:       areaID,
				Harga:        hk.Harga,
				Tanggal:      hariIniPg,
				Retries:      0,
			})
			jobsQueued++
		}
	}

	slog.Info("worker.cron_carry_over_finished", slog.Int("jobs_enqueued", jobsQueued))
}

func NewHargaWorkerPool(queries *db.Queries, workerCount int, bufferSize int) *HargaWorkerPool {
	pool := &HargaWorkerPool{
		jobs:    make(chan HargaJob, bufferSize),
		queries: queries,
	}

	for i := 1; i <= workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

func (p *HargaWorkerPool) worker(id int) {
	defer p.wg.Done()
	for job := range p.jobs {
		p.processJob(id, job)
	}
}

func (p *HargaWorkerPool) processJob(id int, job HargaJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.queries.UpsertHargaHarian(ctx, db.UpsertHargaHarianParams{
		BahanPokokID: job.BahanPokokID,
		AreaID:       job.AreaID,
		Harga:        job.Harga,
		Tanggal:      job.Tanggal,
	})

	if err == nil {
		tanggalStr := job.Tanggal.Time.Format("2006-01-02")
		cache.GlobalCache.DeleteByPrefix(fmt.Sprintf("all_bahan:%s:", tanggalStr))
		cache.GlobalCache.DeleteByPrefix(fmt.Sprintf("detail_bahan:%d:%s:", job.BahanPokokID, tanggalStr))
		cache.GlobalCache.DeleteByPrefix("tren_bahan:")
		return
	}

	if job.Retries < 3 {
		job.Retries++
		slog.Warn("worker.job_retry",
			slog.Int("worker_id", id),
			slog.Int("bahan_id", int(job.BahanPokokID)),
			slog.Int("area_id", int(job.AreaID)),
			slog.Int("retry", job.Retries),
			slog.String("error", err.Error()),
		)

		delay := time.Duration(job.Retries*100) * time.Millisecond
		time.AfterFunc(delay, func() {
			p.jobs <- job
		})
	} else {
		slog.Error("worker.job_failed_permanently",
			slog.Int("worker_id", id),
			slog.Int("bahan_id", int(job.BahanPokokID)),
			slog.Int("area_id", int(job.AreaID)),
			slog.String("error", err.Error()),
		)
	}
}

func (p *HargaWorkerPool) Enqueue(job HargaJob) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		slog.Warn("worker.queue_full_dropped_job",
			slog.Int("bahan_id", int(job.BahanPokokID)),
			slog.Int("area_id", int(job.AreaID)),
		)
		return false
	}
}

func (p *HargaWorkerPool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
	slog.Info("worker.shutdown_complete")
}