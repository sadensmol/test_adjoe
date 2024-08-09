package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-test-task/test-task/src/domain"
)

type SQSTodo struct {
	SQS      *sqs.SQS
	QueueURL string
}

func (s *SQSTodo) Save(ctx context.Context, item domain.ToDoItem) (int64, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second)) //fixme: move to constants
	defer cancel()
	rCh := make(chan int64)
	eCh := make(chan error)

	go func() {
		defer func() {
			close(rCh)
			close(eCh)
		}()

		_, err = s.SQS.SendMessage(&sqs.SendMessageInput{
			MessageBody: aws.String(string(data)),
			QueueUrl:    aws.String(s.QueueURL),
		})
		if err != nil {
			eCh <- err
		}

		rCh <- int64(item.ID)
	}()

	for {
		select {
		case <-ctx.Done():
			return 0, errors.New("sqs timeout")
		case id, ok := <-rCh:
			if ok { //channel not closed
				return id, nil
			}
		case e, ok := <-eCh:
			if ok { //channel not closed
				return 0, e
			}
		}
	}
}
