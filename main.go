package main

import (
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"log"
	"net/http"

	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"math/rand"
	"strconv"
	"flag"
	"os"
	"time"
)

var (
	appPort = flag.Int("port", 8090, "configure http port")
	zipkinApiUrl = flag.String("zurl", "http://localhost:9411/api/v1", "zipkin api url")
	zipkinEnableDebug = flag.Bool("zdebug", false, "zipkin enable debugging?")
	appName = flag.String("name", "go-product-service", "service name to display in zipkin")
	tracer = CreateTracer()
)

const (
	appHost           = "localhost"
)

type Product struct {
	Name string `json:"name"`
}

func main() {
	flag.Parse()
	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/{id}", ProductHandler)
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(*appPort), router))
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Demo service, pass product id to get products")
}

func ProductHandler(res http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]

	parentSpan := CreateSpanFromRequest(
		tracer,
		req,
		"products.assemble",
	)

	defer parentSpan.Finish()

	parentSpan.LogEvent("products.retrieve")
	products := DbService(id, parentSpan.Context())

	parentSpan.LogEvent("products.filter")
	products = AnalyticService(products, parentSpan.Context())

	res.Header().Set(
		"Content-Type",
		"application/json; charset=UTF-8",
	)

	if err := json.NewEncoder(res).Encode(products); err != nil {
		panic(err)
	}
}

func DbService(id string, ctx opentracing.SpanContext) []Product {
	return SimulateDbJob(id, ctx)
}

func SimulateDbJob(id string, ctx opentracing.SpanContext) []Product {
	span := StartSpan(ctx, "SELECT", "mongo")
	defer span.Finish()
	if len(id) > 3 {
		ext.Error.Set(span, true)
	}
	span.SetTag("query", "db.products.find()")

	time.Sleep(30 * time.Millisecond)

	return []Product{
		{Name: fmt.Sprintf("Product A-%s", id)},
		{Name: fmt.Sprintf("Product B-%s", id)},
		{Name: fmt.Sprintf("Product C-%s", id)},
	}
}

func AnalyticService(products []Product, ctx opentracing.SpanContext) []Product {
	return SimulateAnalyticJob(products, ctx)
}

func SimulateAnalyticJob(products []Product, ctx opentracing.SpanContext) []Product {
	span := StartSpan(ctx, "filter", "apache-spark")
	defer span.Finish()
	time.Sleep(60 * time.Millisecond)

	if rand.Float64() > 0.5 {
		span.LogEvent("error")
	}

	return products[:2]
}

func StartSpan(ctx opentracing.SpanContext, operationName, serviceName string) opentracing.Span {
	span := opentracing.StartSpan(operationName, opentracing.ChildOf(ctx))
	// Peer tag: creating custom tag
	//wrap a call to an external service which is not instrumented
	ext.PeerHostname.Set(span, "localhost") // remote host
	ext.PeerPort.Set(span, 27017)           // remote port
	ext.PeerService.Set(span, serviceName)  //remote service name
	ext.SpanKind.Set(span, "resource")      // demonstrating in-process activity
	// Use it when the resource that you want to trace have no server side support

	return span
}

func CreateTracer() opentracing.Tracer {
	tracer, err := zipkin.NewTracer(
		CreateRecorder(CreateCollector(*zipkinApiUrl+"/spans")),
		zipkin.ClientServerSameSpan(true),
	)

	if err != nil {
		fmt.Errorf("error creating tracker %v", err)
		os.Exit(-1)
	}

	opentracing.InitGlobalTracer(tracer)
	return tracer
}

func CreateCollector(url string) zipkin.Collector {
	collector, err := zipkin.NewHTTPCollector(url)
	if err != nil {
		fmt.Errorf("error creating collector %v", err)
		os.Exit(-1)
	}

	return collector
}

func CreateRecorder(collector zipkin.Collector) zipkin.SpanRecorder {
	return zipkin.NewRecorder(collector,
		*zipkinEnableDebug,
		appHost+":"+strconv.Itoa(*appPort),
		*appName)
}

func CreateSpanFromRequest(tracer opentracing.Tracer, req *http.Request, operationName string) opentracing.Span {
	// Carries tracing state
	carrier := opentracing.HTTPHeadersCarrier(req.Header)

	//Extract SpanContext from incoming request
	extractedContext, err := tracer.Extract(
		opentracing.HTTPHeaders,
		carrier,
	)
	var span opentracing.Span

	//should we start or continue a trace ?
	if err != nil {
		//Starts a new span
		span = tracer.StartSpan(operationName)
	} else {
		//continue
		span = tracer.StartSpan(
			operationName,
			opentracing.ChildOf(extractedContext))
	}

	ext.SpanKindRPCServer.Set(span)
	span.SetTag("http.method", req.Method)
	span.SetTag("http.url", req.URL)

	return span
}
