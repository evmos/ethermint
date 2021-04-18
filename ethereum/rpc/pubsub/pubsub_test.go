package pubsub

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddTopic(t *testing.T) {
	assert := assert.New(t)

	q := NewEventBus()
	err := q.AddTopic("kek", make(chan interface{}))
	if !assert.NoError(err) {
		return
	}

	err = q.AddTopic("lol", make(chan interface{}))
	if !assert.NoError(err) {
		return
	}

	err = q.AddTopic("lol", make(chan interface{}))
	if !assert.Error(err) {
		return
	}

	assert.EqualValues([]string{"kek", "lol"}, q.Topics())
}

func TestSubscribe(t *testing.T) {
	assert := assert.New(t)

	q := NewEventBus()
	kekSrc := make(chan interface{})
	q.AddTopic("kek", kekSrc)

	lolSrc := make(chan interface{})
	q.AddTopic("lol", lolSrc)

	kekSubC, err := q.Subscribe("kek")
	if !assert.NoError(err) {
		return
	}

	lolSubC, err := q.Subscribe("lol")
	if !assert.NoError(err) {
		return
	}

	lol2SubC, err := q.Subscribe("lol")
	if !assert.NoError(err) {
		return
	}

	wg := new(sync.WaitGroup)
	wg.Add(4)

	go func() {
		defer wg.Done()
		msg := <-kekSubC
		log.Println("kek:", msg)
		assert.EqualValues(1, msg)
	}()

	go func() {
		defer wg.Done()
		msg := <-lolSubC
		log.Println("lol:", msg)
		assert.EqualValues(1, msg)
	}()

	go func() {
		defer wg.Done()
		msg := <-lol2SubC
		log.Println("lol2:", msg)
		assert.EqualValues(1, msg)
	}()

	go func() {
		defer wg.Done()

		time.Sleep(time.Second)
		kekSrc <- 1
		lolSrc <- 1

		close(kekSrc)
		close(lolSrc)
	}()

	wg.Wait()
	time.Sleep(time.Second)
}
