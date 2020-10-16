package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("Counter"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String("test"),
			},
		},
		UpdateExpression: aws.String("ADD CountValue :incr"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":incr": {
				N: aws.String("1"),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	var wg sync.WaitGroup
	wg.Add(int(worker))
	for i := 0; i < int(worker); i++ {
		go func() {
			defer wg.Done()
			for {
				if _, err := svc.UpdateItemWithContext(ctx, input); err != nil {
					if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
						return
					} else {
						counter.ErrIncrement()
					}
				} else {
					counter.Increment()
				}
			}
		}()
	}
	wg.Wait()
}
