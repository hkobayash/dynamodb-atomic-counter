package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	worker, _ := strconv.ParseInt(os.Args[1], 10, 64)
	log.Printf("worker: %d", worker)

	sigH := NewSigH(func(s os.Signal) {
		log.Printf("received signal: %s\nGraceful shutdown...\n", s.String())
	}, syscall.SIGHUP, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	go sigH.Run(ctx, cancel)

	counter := NewCounter()
	go counter.Watch(ctx)

	db, err := sql.Open("mysql", os.Args[2])
	if err != nil {
		log.Fatalf("sql open err: %#v", err)
	}
	defer db.Close()

	var wg sync.WaitGroup
	wg.Add(int(worker))
	for i := 0; i < int(worker); i++ {
		go func() {
			defer wg.Done()
			for {
				if res, err := db.ExecContext(ctx, "UPDATE sequence SET id=LAST_INSERT_ID(id+1)"); err != nil {
					return
				} else {
					if _, err := res.LastInsertId(); err != nil {
						counter.ErrIncrement()
					} else {
						counter.Increment()
					}
				}
			}
		}()
	}
	wg.Wait()
}
