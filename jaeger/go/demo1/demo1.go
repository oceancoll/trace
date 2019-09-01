package demo1
// 最简单的脚本使用

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/opentracing/opentracing-go/log"
	"io"
	"time"
	"context"
)

// 初始化jaeger
func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "106.12.200.182:6831",
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func foo3(req string, ctx context.Context) (reply string){
	//1.创建子span
	span, _ := opentracing.StartSpanFromContext(ctx, "span_foo3")
	defer func() {
		//4.接口调用完，在tag中设置request和reply
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()
	// 会与上述tag一同被展示为kv格式
	span.SetTag("testkey", "testval")
	println(req)
	//2.模拟处理耗时
	time.Sleep(time.Second/2)
	//3.返回reply
	reply = "foo3Reply"
	return
}

//跟foo3一样逻辑
func foo4(req string, ctx context.Context) (reply string){
	span, _ := opentracing.StartSpanFromContext(ctx, "span_foo4")
	defer func() {
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	println(req)
	time.Sleep(time.Second/2)
	reply = "foo4Reply"
	return
}

func main() {
	// 初始化jaeger, "jaeger-demo"是servicenname
	tracer, closer := initJaeger("jaeger-demo")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)//StartspanFromContext创建新span时会用到
	//创建父span, "span_root"是operationname
	span := tracer.StartSpan("span_root")
	//生成contaxt，用于传递上下文
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	// 模拟调用两个服务
	r1 := foo3("Hello foo3", ctx)
	r2 := foo4("Hello foo4", ctx)
	fmt.Println(r1, r2)
	//测试log使用
	helloStr := fmt.Sprintf("Hello, %s!", "tyt")
	//生成1条log，kv格式
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)

	println(helloStr)
	//再生成一条log，也就是生成log的方式，推荐使用这种
	span.LogKV("event", "println")
	span.Finish()
}
