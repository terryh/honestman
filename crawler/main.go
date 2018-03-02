package main

import (
	"honestman/app"
	"honestman/crawler/task"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	// AppContext hold share object
	AppContext *app.Context
	// Tasks all running task
	Tasks []task.CrawlerTask
)

// RunTasks fire the execution of each task
func RunTasks(context *app.Context) {
	// Task name && sleep interval seconds
	// one day second = 24 * 60 * 60 = 86400
	// 8 * 60 * 60  = 28800

	Tasks = append(Tasks, task.NewRTmart(context, 28800))
	Tasks = append(Tasks, task.NewCarrefour(context, 28800))

	for _, task := range Tasks {
		log.Println("Running", task)
		go task.Run()
	}
}

// process shut down
func sigHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		os.Interrupt,
		os.Kill,
	)

	for s := range sigChan {
		log.Println("Caught", s)
		break
	}
	os.Exit(0)
}

func main() {

	// init share context
	AppContext = app.NewContext()

	// build task
	RunTasks(AppContext)
	// handle process close
	sigHandler()
}
