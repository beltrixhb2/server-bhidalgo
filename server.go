package main

import (
    "errors"
	"strconv"
	"context"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    //"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"fmt"
	"encoding/json"
	"net/http"
	loggly "github.com/jamespearly/loggly"
	"github.com/gorilla/mux"

)

type DynamoDBItem struct {
    Time    int `json:"Time"`
    AircraftList  string `json:"AircraftList"`
}

type StatusResponse struct {
	Table	    string `json:"table"`
	RecordCount int    `json:"recordCount"`
}

type statusResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func NewStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
    return &statusResponseWriter{
        ResponseWriter: w,
        statusCode:     http.StatusOK,
    }
}

func (sw *statusResponseWriter) WriteHeader(statusCode int) {
    sw.statusCode = statusCode
    sw.ResponseWriter.WriteHeader(statusCode)
}

func statusHandler(w http.ResponseWriter, req *http.Request) {

	db, ok := req.Context().Value("dynamodb").(*dynamodb.DynamoDB)
    	if !ok {
        	http.Error(w, "DynamoDB client not found", http.StatusInternalServerError)
        	return
    	}	
    	tableName := "bhidalgo_Aircraft_States"
    	result, err := db.DescribeTable(&dynamodb.DescribeTableInput{
        	TableName: aws.String(tableName),
    	})
    	if err != nil {
        	http.Error(w, err.Error(), http.StatusInternalServerError)
        	return
    	}
	recordCount := *result.Table.ItemCount
	response := StatusResponse{
		Table: tableName,  
		RecordCount: int(recordCount),
	}
	sendJSONResponse(w, response)
}


func getItemsFromDynamoDB(req *http.Request, tableName string, time int, bef int, aft int) ([]DynamoDBItem, error) {
	var items []DynamoDBItem
    var params *dynamodb.ScanInput
    db, ok := req.Context().Value("dynamodb").(*dynamodb.DynamoDB)
    if !ok {
        return items, errors.New("Dynamodb client not found") 
    }
    if (time==-1&&bef==-1&&aft==-1){
        params = &dynamodb.ScanInput{
            TableName: aws.String(tableName),
        }
    }else if time!=-1{
        params = &dynamodb.ScanInput{
            TableName: aws.String(tableName),
            FilterExpression: aws.String("#T = :timeValue"),
            ExpressionAttributeNames: map[string]*string{
                "#T": aws.String("Time"),
            },
            ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
                ":timeValue": {N: aws.String(strconv.Itoa(time))},
            },
        }
    }else if aft!=-1&&bef==-1{
         params = &dynamodb.ScanInput{
            TableName: aws.String(tableName),
            FilterExpression: aws.String("#T > :after"),
            ExpressionAttributeNames: map[string]*string{
                "#T": aws.String("Time"),
            },
            ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
                ":after": {
                    N: aws.String(strconv.Itoa(aft)),
                },
            },
        }
    }else{
        params = &dynamodb.ScanInput{
            TableName: aws.String(tableName),
            FilterExpression: aws.String("#T < :before AND #T > :after"),
            ExpressionAttributeNames: map[string]*string{
                "#T": aws.String("Time"),
            },
            ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
                ":before": {
                    N: aws.String(strconv.Itoa(bef)),
                },
                ":after": {
                    N: aws.String(strconv.Itoa(aft)),
                },
            },
        }
    }
	err := db.ScanPages(params,
			func(page *dynamodb.ScanOutput, lastPage bool) bool {
				for _, item := range page.Items {
					One_Item := DynamoDBItem{}
					for attributeName, attributeValue := range item {
						switch attributeName {
						case "Time":
							One_Item.Time,_ = strconv.Atoi(*attributeValue.N)
						case "AircraftList":
							One_Item.AircraftList = *attributeValue.S
						}
					}
					items = append(items, One_Item)
				}
				return !lastPage
			},
		)
    return items, err
}

func allHandler(w http.ResponseWriter, req *http.Request){
    tableName := "bhidalgo_Aircraft_States"
    items, err := getItemsFromDynamoDB(req, tableName, -1, -1, -1)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
    }
    w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func searchHandler(w http.ResponseWriter, req *http.Request){
    tableName := "bhidalgo_Aircraft_States"
    timeStr:= req.URL.Query().Get("time")
	befStr:= req.URL.Query().Get("bef")
	aftStr:= req.URL.Query().Get("aft")
    var time, bef, aft int
    var err error
    if timeStr!=""&&(befStr!=""||aftStr!=""){
        handleInvalidParameter(w, "too many parameters")
        return
    }
    if timeStr!=""{
        time, err = strconv.Atoi(timeStr)
        if err!= nil||time<0||time>3000000000  {
            handleInvalidParameter(w, "time")
            return
        }

    }else{
        time = -1
    }
    if befStr!=""{
        bef, err = strconv.Atoi(befStr)
        if err != nil||bef<0||bef>3000000000 {
            handleInvalidParameter(w, "bef")
            return
        }
    }else{
        bef= -1
    }
    if aftStr!=""{
        aft, err = strconv.Atoi(aftStr)
        if err != nil||aft<0||aft>3000000000 {
            handleInvalidParameter(w, "aft")
            return
        }
    }else{
        aft= -1
    }
    items, err := getItemsFromDynamoDB(req, tableName, time, bef, aft)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
    }
    items, err = getItemsFromDynamoDB(req, tableName, time, bef, aft)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
    }
    w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func handleInvalidParameter(w http.ResponseWriter, paramName string) {
	errorMessage := fmt.Sprintf("Invalid value for parameter: %s", paramName)
	errorResponse := map[string]string{"error": errorMessage}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResponse)
}

func notAllowedHandler(w http.ResponseWriter, req *http.Request) {
    http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	http.NotFound(w, req)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)
	if err != nil {
		fmt.Printf("½v+\n",err)
    		http.Error(w, err.Error(), http.StatusInternalServerError)
    		return
	}
}

func RequestLoggerMiddleware(client *loggly.ClientType) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sw := NewStatusResponseWriter(w)
		defer func(){
			logMessage := fmt.Sprintf("Method: %s, IP: %s, Path: %s, Status: %d",
                        	req.Method, req.RemoteAddr, req.URL.Path, sw.statusCode)
                	client.EchoSend("info",logMessage)
		}()
		sess := session.Must(session.NewSessionWithOptions(session.Options{
                		SharedConfigState: session.SharedConfigEnable,
       			}))
       		svc := dynamodb.New(sess)
		ctx := context.WithValue(req.Context(), "dynamodb", svc)
            	next.ServeHTTP(sw, req.WithContext(ctx))
		
        })
    }
}

func main() {	
	client := loggly.New("Server")
	r := mux.NewRouter()
	r.Use(RequestLoggerMiddleware(client))
	r.HandleFunc("/bhidalgo/status", statusHandler).Methods(http.MethodGet)
	r.HandleFunc("/bhidalgo/all", allHandler).Methods(http.MethodGet)
	r.HandleFunc("/bhidalgo/search", searchHandler).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(notFoundHandler).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(notAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	fmt.Printf("Server is running\n")
	http.ListenAndServe("0.0.0.0:8080",r) 
}


