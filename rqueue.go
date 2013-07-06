package rqueue

import (
	"fmt"
	"github.com/adeven/goenv"
	"github.com/adeven/redis"
	"time"
)

type Queue struct {
	redisClient *redis.Client
	name        string
}

func NewQueue(goenv *goenv.Goenv, name string) *Queue {
	q := &Queue{name: name}
	host, port, db := goenv.GetRedis()
	q.redisClient = redis.NewTCPClient(host+":"+port, "", int64(db))
	return q
}

func (queue *Queue) Put(payload string) error {
	p := &Package{CreatedAt: time.Now(), Payload: payload, Queue: queue}
	answer := queue.redisClient.LPush(InputQueueName(queue), p.GetString())
	return answer.Err()
}

func (queue *Queue) Get(consumer string) (*Package, error) {
	l := queue.redisClient.LLen(WorkingQueueName(queue, consumer))
	if l.Val() != 0 {
		return nil, fmt.Errorf("unacked Packages found!")
	}
	answer := queue.redisClient.BRPopLPush(InputQueueName(queue), WorkingQueueName(queue, consumer), 0)
	return UnmarshalPackage(answer.Val(), queue, consumer), answer.Err()
}

// func (queue *Queue) GetUnAcked(consumer string) *Package {
// 	//TODO
// }

func (queue *Queue) Ack(p *Package) error {
	answer := queue.redisClient.RPop(WorkingQueueName(queue, p.Consumer))
	return answer.Err()
}

func (queue *Queue) Requeue(p *Package) error {
	answer := queue.redisClient.RPopLPush(WorkingQueueName(queue, p.Consumer), InputQueueName(queue))
	return answer.Err()
}

func (queue *Queue) Fail(p *Package) error {
	answer := queue.redisClient.RPopLPush(WorkingQueueName(queue, p.Consumer), FailedQueueName(queue))
	return answer.Err()
}
