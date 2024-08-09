package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-test-task/test-task/src/config"
	"github.com/go-test-task/test-task/src/domain"
	"github.com/go-test-task/test-task/src/service"
	"github.com/go-test-task/test-task/src/storage"
)

type App struct {
	ctx context.Context
	cfg *config.Config
}

type ToDoSaver interface {
	Save(context.Context, domain.ToDoItem) (int64, error)
}
type ToDoLatestReader interface {
	Latest(context.Context) (domain.ToDoItem, error)
}

type ToDoStorage interface {
	ToDoSaver
	ToDoLatestReader
}

var dbStorage ToDoStorage
var sqsStorage ToDoSaver

// fixme: move to service
func (a *App) writeTodoItems(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	var err error
	var id int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			todoItem := domain.ToDoItem{Description: "Test", DueDate: time.Now().Add(24 * time.Hour)}
			id, err = dbStorage.Save(ctx, todoItem)
			if err != nil {
				log.Printf("error saving item: %v", err)
			}
			todoItem = todoItem.WithID(id)
			log.Printf("Saved item: %v", todoItem)
			_, err = sqsStorage.Save(ctx, todoItem)
			if err != nil {
				log.Printf("error saving item to SQS: %v", err)
			}
		}
	}
}

func (a *App) initDB() *sql.DB {
	// Define the DSN (Data Source Name)

	// Open a connection to the database
	var err error
	db, err := sql.Open("mysql", a.cfg.DBURL)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}

	// Ping the database to ensure a connection is established
	err = db.Ping()
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	// Execute the initial script
	initialScript := `
    CREATE TABLE IF NOT EXISTS ToDoItem (
        id INT AUTO_INCREMENT,
        description TEXT NOT NULL,
        due_date DATETIME NOT NULL,
        PRIMARY KEY (id)
    );
    `

	_, err = db.Exec(initialScript)
	if err != nil {
		log.Fatalf("error executing initial script: %v", err)
	}

	fmt.Println("database initialized successfully")

	return db
}

func (a *App) InitSQS() *sqs.SQS {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:                        aws.String("eu-central-1"),
			Endpoint:                      aws.String("http://localstack:4566"),
			Credentials:                   credentials.NewStaticCredentials("test", "test", ""),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},

		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(sess)
	return svc
}

func (a *App) Start(ctx context.Context, cfg *config.Config) {
	a.ctx = ctx
	a.cfg = cfg

	db := a.initDB()
	dbStorage = &storage.MySqlTodo{DB: db}

	sqs := a.InitSQS()
	sqsStorage = &service.SQSTodo{SQS: sqs, QueueURL: cfg.SQSQueueURL}

	go a.writeTodoItems(ctx)

	http.HandleFunc("/todo", func(w http.ResponseWriter, r *http.Request) {
		toDoItem, err := dbStorage.Latest(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json, err := json.Marshal(toDoItem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)

	})

	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
