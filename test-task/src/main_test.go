package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const URL = "http://localhost:3333"

type MainTestSuite struct {
	suite.Suite
}

func TestMainTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &MainTestSuite{})
}

func getResponseAsMap(t *testing.T, res *http.Response) map[string]interface{} {
	t.Helper()

	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var respMap map[string]interface{}
	err = json.Unmarshal(respBody, &respMap)
	require.NoError(t, err)

	return respMap
}

func (s *MainTestSuite) TestRequest() {
	t := s.T()
	fmt.Println("TestRequest")
	apitest.New().EnableNetworking().Debug().
		Observe(func(res *http.Response, req *http.Request, apiTest *apitest.APITest) {
			respArray := getResponseAsMap(t, res)
			require.Equal(t, 3, len(respArray))
			require.Equal(t, "Test", respArray["description"])
		}).
		Get(fmt.Sprintf("%s/todo", URL)).
		Expect(t).
		Status(200).
		End()
}

func (s *MainTestSuite) TestSQS() {
	t := s.T()
	qURL := "http://sqs.eu-central-1.localhost.localstack.cloud:4566/000000000000/my-queue"

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String("eu-central-1"),
			Endpoint:    aws.String("http://localhost:4566"),
			Credentials: credentials.NewStaticCredentials("test", "test", ""),

			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	}))
	sqc := sqs.New(sess)

	result2, err := sqc.ReceiveMessage(&sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(3),
		QueueUrl:            aws.String(qURL),
	})
	require.NoError(t, err)
	for _, msg := range result2.Messages {
		// Process the received message
		fmt.Println(*msg.Body)
		// Delete the message
	}
}
