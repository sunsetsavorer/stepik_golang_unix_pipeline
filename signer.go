package main

import "sync"

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

}

var MultiHash = func(in, out chan interface{}) {

}

var CombineResults = func(in, out chan interface{}) {

}
