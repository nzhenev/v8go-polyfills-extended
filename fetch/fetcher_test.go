package fetch

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/esoptra/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/console"
	"github.com/nzhenev/v8go-polyfills-extended/uuid"
)

func TestNewFetcher(t *testing.T) {
	t.Parallel()

	f1 := NewFetcher()
	if f1 == nil {
		t.Error("create fetcher failed")
		return
	}

	if h := f1.GetLocalHandler(); h == nil {
		t.Error("local handler is <nil>")
	}

	f2 := NewFetcher(WithLocalHandler(nil))
	if f2 == nil {
		t.Error("create fetcher with local handler failed")
		return
	}

	if h := f2.GetLocalHandler(); h != nil {
		t.Error("set fetcher local handler to <nil> failed")
		return
	}
}

func TestFetchJSON(t *testing.T) {
	t.Parallel()

	ctx, err := newV8ContextWithFetch()
	if err != nil {
		t.Errorf("create v8: %s", err)
		return
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; utf-8")
		_, _ = w.Write([]byte(`{"status": true}`))
	}))

	val, err := ctx.RunScript(fmt.Sprintf("fetch('%s').then(res => res.json())", srv.URL), "fetch_json.js")
	if err != nil {
		t.Error(err)
		return
	}

	proms, err := val.AsPromise()
	if err != nil {
		t.Error(err)
		return
	}

	for proms.State() == v8go.Pending {
		continue
	}

	res, err := proms.Result().AsObject()
	if err != nil {
		t.Error(err)
		return
	}

	status, err := res.Get("status")
	if err != nil {
		t.Error(err)
		return
	}

	if !status.Boolean() {
		t.Error("status should be true")
	}
}

func TestHeaders(t *testing.T) {
	t.Parallel()

	iso := v8go.NewIsolate()

	ctx := v8go.NewContext(iso)

	obj, err := newHeadersObject(ctx, http.Header{
		"AA": []string{"aa"},
		"BB": []string{"bb"},
	})
	if err != nil {
		t.Error(err)
		return
	}

	aa, err := obj.Get("AA")
	if err != nil {
		t.Error(err)
		return
	}

	if aa.String() != "aa" {
		t.Errorf("should be 'aa' but is '%s'", aa.String())
		return
	}

	fn, err := obj.Get("get")
	if err != nil {
		t.Error(err)
		return
	}

	if !fn.IsFunction() {
		t.Error("should be function")
		return
	}
}

func newV8ContextWithFetch(opt ...Option) (*v8go.Context, error) {
	iso := v8go.NewIsolate()
	global := v8go.NewObjectTemplate(iso)

	if err := InjectTo(iso, global, opt...); err != nil {
		return nil, err
	}

	return v8go.NewContext(iso, global), nil
}

func testFetchBodyWithLazyLoad(t *testing.T, script string) {

	iso := v8go.NewIsolate()
	defer iso.Dispose()
	global := v8go.NewObjectTemplate(iso)

	fetcher := NewFetcher()
	if err := InjectWithFetcherTo(iso, global, fetcher); err != nil {
		t.Error(err)
		return
	}

	ctx := v8go.NewContext(iso, global)
	if err := console.InjectTo(ctx); err != nil {
		panic(err)
	}
	if err := InjectHTTPProperties(ctx); err != nil {
		panic(err)
	}

	fn := `function Response(body, init){
		console.log("Response >> "+body)
		if(init == null || init == undefined){
			init =  { "status": 200, "statusText": "OK" }
		}
		if(body == null || body == undefined){
			this.body = ''
		}else if (body.body){
			this.body = body.body
		}else{
			this.body = body
		}
		this.status = init.status
		this.statusText = init.statusText
		this.headers = init.headers
	}
	` + script + `
	let res = epsilon();
	Promise.resolve(res)
	`

	//fmt.Println(fn)
	val, err := ctx.RunScript(fn, "fetch_json.js")
	if err != nil {
		t.Error(err)
		return
	}

	proms, err := val.AsPromise()
	if err != nil {
		t.Error(err)
		return
	}

	for proms.State() == v8go.Pending {
		continue
	}

	res, err := proms.Result().AsObject()
	if err != nil {
		t.Error(err)
		return
	}

	status, err := res.Get("status")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("status : ", status.String())

	body, err := res.Get("body")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("body : ", body.String())
	if uuid.IsUUID(body.String()) {
		val, ok := fetcher.ResponseMap.Load(body.String())
		if ok {
			result := val.(io.ReadCloser)
			defer result.Close()
			bodyBytes, err := ioutil.ReadAll(result)
			if err != nil {
				t.Error(fmt.Errorf("Error while getting status of epsilon execution : %#v", err))
				return
			}
			fmt.Printf("status %s, bodyBytes %s", status.String(), string(bodyBytes))
			return
		}
	}
	fmt.Printf("status %s, body %s", status.String(), body.String())

}

func TestMultipleFetchWithSequentialResponseConsumptionInJS(t *testing.T) {
	addr := "localhost:10000"
	go StartHttpServer(addr)
	time.Sleep(time.Second * 5)

	testFetchBodyWithLazyLoad(t, `epsilon = async (event) => {
			const url = 'http://127.0.0.1:10000/'
			let resp = await fetch(url)
			let out = await resp.text()
			console.log(">> "+out)
			const url1 = 'http://127.0.0.1:10000/public'
			let resp1 = await fetch(url1)
			//let resp1text = await resp1.text()
			//console.log(">> "+resp1text)
			//multiple fetch with all response consumed multiple times in script
			return new Response(resp1)
		}`)

	time.Sleep(time.Second * 5)
	testFetchBodyWithLazyLoad(t, `epsilon = async (event) => {
			const url = 'http://127.0.0.1:10000/'
			let resp = await fetch(url)

			const url1 = 'http://127.0.0.1:10000/public'
			let resp1 = await fetch(url1)

			//multiple fetch with all response consumed multiple times in script
			return new Response(resp)
		}`)
	// es.Expectf(err == nil, "Error %#v", err)
	// es.Expectf(resp != nil, "expected not nil")
	// es.Expectf(strings.ToLower(resp.Status) == "200", "unexpected status %q", resp.Status)
	// es.Expectf(strings.ToLower(resp.Body) == "home page", "unexpected status %q", resp.Body)
}
