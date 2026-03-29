package worker

import (
    "context"
    "fmt"
    "time"
    "sync"

    "github.com/farildzaky/siskaperbapo-service/internal/cache"
    "github.com/farildzaky/siskaperbapo-service/internal/db"
    "github.com/jackc/pgx/v5/pgtype"
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

func NewHargaWorkerPool(queries *db.Queries, workerCount int, bufferSize int) *HargaWorkerPool {
    pool := &HargaWorkerPool{
        jobs:    make(chan HargaJob, bufferSize),
        queries: queries,
    }

    for i := 0; i < workerCount; i++ {
        pool.wg.Add(1)
        go pool.worker(i)
    }

    return pool
}

func (p *HargaWorkerPool) worker(id int) {
    defer p.wg.Done()
    for job := range p.jobs {
        func() {
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            defer cancel() // ← selalu bersih meski panic

            _, err := p.queries.CreateHargaHarian(ctx, db.CreateHargaHarianParams{
                BahanPokokID: job.BahanPokokID,
                AreaID:       job.AreaID, 
                Harga:        job.Harga,
                Tanggal:      job.Tanggal,
            })

            if err == nil {
                tanggalStr := job.Tanggal.Time.Format("2006-01-02")
                cache.GlobalCache.DeleteByPrefix(fmt.Sprintf("all_bahan:%s:", tanggalStr))
                cache.GlobalCache.DeleteByPrefix(fmt.Sprintf("detail_bahan:%d:%s:", job.BahanPokokID, tanggalStr))
                return
            }

            // err != nil — retry logic
            if job.Retries < 3 {
                job.Retries++
                go func(j HargaJob) {
                    time.Sleep(time.Duration(j.Retries*100) * time.Millisecond)
                    p.jobs <- j
                }(job)
            } else {
                fmt.Printf("🚨 [Worker %d] GAGAL PERMANEN ID %d: %v\n", id, job.BahanPokokID, err)
            }
        }()
    }
}

func (p *HargaWorkerPool) Enqueue(job HargaJob) bool {
    select {
    case p.jobs <- job:
        return true
    default:
        return false
    }
}

func (p *HargaWorkerPool) Shutdown() {
    close(p.jobs) 
    p.wg.Wait()   
}