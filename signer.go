package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

type PipelineUnit struct {
	JobFunc job
	In      chan interface{}
	Out     chan interface{}
}

func ExecutePipeline(jobs ...job) {

	pipeline := make([]*PipelineUnit, len(jobs))

	pipeline[0] = &PipelineUnit{
		JobFunc: jobs[0],
		In:      make(chan interface{}),
		Out:     make(chan interface{}),
	}

	for i := 1; i < len(jobs); i++ {

		pipeline[i] = &PipelineUnit{
			JobFunc: jobs[i],
			In:      pipeline[i-1].Out,
			Out:     make(chan interface{}),
		}
	}

	wg := &sync.WaitGroup{}

	for _, unit := range pipeline {

		wg.Add(1)

		go worker(unit, wg)
	}

	wg.Wait()
}

func worker(unit *PipelineUnit, wg *sync.WaitGroup) {

	unit.JobFunc(unit.In, unit.Out)

	close(unit.Out)

	wg.Done()
}

var SingleHash = func(in, out chan interface{}) {

	blocker := make(chan struct{})

	wg := &sync.WaitGroup{}

	for v := range in {
		wg.Add(1)

		go SingleHashWorker(v, wg, out, blocker)
		blocker <- struct{}{}
	}

	wg.Wait()
}

func SingleHashWorker(v interface{}, parentWG *sync.WaitGroup, out chan interface{}, blocker chan struct{}) {

	defer parentWG.Done()

	data := strconv.Itoa(v.(int))

	wg := &sync.WaitGroup{}

	crc := ""
	md5 := ""
	crcFromMd5 := ""

	wg.Add(2)

	go func() {
		md5 = DataSignerMd5(data)

		<-blocker

		crcFromMd5 = DataSignerCrc32(md5)

		wg.Done()
	}()

	go func() {
		crc = DataSignerCrc32(data)

		wg.Done()
	}()

	wg.Wait()

	res := crc + "~" + crcFromMd5

	out <- res
}

var MultiHash = func(in, out chan interface{}) {

	wg := &sync.WaitGroup{}

	for v := range in {
		wg.Add(1)

		go MultiHashWorker(v, wg, out)
	}

	wg.Wait()
}

func MultiHashWorker(v interface{}, parentWG *sync.WaitGroup, out chan interface{}) {

	defer parentWG.Done()

	data := v.(string)

	m := &sync.Map{}
	wg := &sync.WaitGroup{}

	for th := 0; th <= 5; th++ {
		wg.Add(1)

		go func(idx int, m *sync.Map) {

			crc := DataSignerCrc32(strconv.Itoa(idx) + data)

			m.Store(idx, crc)

			wg.Done()
		}(th, m)
	}

	wg.Wait()

	res := ""

	for i := 0; i <= 5; i++ {
		value, _ := m.Load(i)

		res += value.(string)
	}

	out <- res
}

var CombineResults = func(in, out chan interface{}) {

	arr := []string{}

	for v := range in {
		arr = append(arr, v.(string))
	}

	sort.Strings(arr)

	out <- strings.Join(arr, "_")
}
