package main

import (
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

	for v := range in {

		data := strconv.Itoa(v.(int))

		crc := DataSignerCrc32(data)
		md5 := DataSignerMd5(data)
		crcFromMd5 := DataSignerCrc32(md5)

		res := crc + "~" + crcFromMd5

		out <- res
	}
}

var MultiHash = func(in, out chan interface{}) {

	for v := range in {

		data := v.(string)
		res := ""

		for th := 0; th <= 5; th++ {

			crc := DataSignerCrc32(strconv.Itoa(th) + data)
			res += crc
		}

		out <- res
	}
}

var CombineResults = func(in, out chan interface{}) {

	arr := []string{}

	for v := range in {
		arr = append(arr, v.(string))
	}

	out <- strings.Join(arr, "_")
}
