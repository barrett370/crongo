# crongo


Very simple task scheduler library. 

Given a struct implemeting in the `Tasker` interface, a `name` and an `time.Interval`, an instance of the `Scheduler` will run the task every interval. 



## Example 

```go

type myTask struct{}

func (myTask) Run(context.Context) error {
    println("work!")
    return nil
}

...
scheduler := crongo.New("example", myTask{}, time.Second)
scheduler.Start()
...
scheduler.Stop()
...
```