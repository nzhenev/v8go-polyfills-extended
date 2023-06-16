package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/esoptra/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/fetch/internal"
	. "github.com/nzhenev/v8go-polyfills-extended/internal"
	"github.com/nzhenev/v8go-polyfills-extended/uuid"
)

const (
	UserAgentLocal = "<local>"
	AddrLocal      = "0.0.0.0:0"
)

var defaultLocalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
})

var defaultUserAgentProvider = UserAgentProviderFunc(func(u *url.URL) string {
	if !u.IsAbs() {
		return UserAgentLocal
	}

	return UserAgent()
})

type Fetcher interface {
	GetLocalHandler() http.Handler

	GetFetchFunctionCallback() v8go.FunctionCallback
}

type Fetch struct {
	// Use local handler to handle the relative path (starts with "/") request
	LocalHandler http.Handler

	UserAgentProvider UserAgentProvider
	AddrLocal         string
	ResponseMap       *sync.Map
	InputBody         io.ReadCloser
}

func NewFetcher(opt ...Option) *Fetch {
	ft := &Fetch{
		LocalHandler:      defaultLocalHandler,
		UserAgentProvider: defaultUserAgentProvider,
		AddrLocal:         AddrLocal,
		ResponseMap:       &sync.Map{},
	}

	for _, o := range opt {
		o.apply(ft)
	}

	return ft
}

func (f *Fetch) GetLocalHandler() http.Handler {
	return f.LocalHandler
}

func (f *Fetch) GetFetchFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		args := info.Args()

		resolver, _ := v8go.NewPromiseResolver(ctx)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered from panic", r)
				}
			}()
			if len(args) <= 0 {
				err := errors.New("1 argument required, but only 0 present")
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			var reqInit internal.RequestInit
			var err error
			if len(args) > 1 {
				res, err := getRequestInit(ctx, args[1])
				if err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}
				if res != nil {
					reqInit = *res
				}
			}

			val := args[0]
			var u *url.URL
			if val.IsString() {
				//this happens when invoked as: fetch(url)
				u, err = url.Parse(val.String())
				if err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}
			} else {
				//this happens when invoked as: fetch(new Request(url, options))
				uri, err := val.MarshalJSON()
				if err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}
				var jsReqInit internal.JSRequestInit
				reader := strings.NewReader(string(uri))
				if err := json.NewDecoder(reader).Decode(&jsReqInit); err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}
				reqInit.Method = jsReqInit.Method
				reqInit.Redirect = jsReqInit.Redirect
				reqInit.Body = jsReqInit.Body
				reqInit.Headers = jsReqInit.Headers

				u, err = url.Parse(jsReqInit.Url)
				if err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}
			}

			r, err := f.initRequest(u, reqInit)
			if err != nil {
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			var res *internal.Response

			// do local request
			if !r.URL.IsAbs() {
				res, err = f.fetchLocal(r)
			} else {
				res, err = f.fetchRemote(r)
			}
			if err != nil {
				resolver.Reject(newErrorValue(ctx, err))
				return
			}
			//store a pointer reference with the fetcher
			mini := uuid.NewUuid()
			f.ResponseMap.Store(mini, res.BodyReader)
			res.Body = mini

			resObj, err := newResponseObject(ctx, res)
			if err != nil {
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			resolver.Resolve(resObj)
		}()

		return resolver.GetPromise().Value
	}
}

func (f *Fetch) initRequest(u *url.URL, reqInit internal.RequestInit) (*internal.Request, error) {

	req := &internal.Request{
		URL: u,
		Header: http.Header{
			"Accept":     []string{"*/*"},
			"Connection": []string{"close"},
		},
	}

	if strings.TrimSpace(reqInit.Body) != "" {
		req.Body = strings.NewReader(reqInit.Body)
	} else {
		//supports end to end streaming
		req.Body = f.InputBody
	}

	var ua string
	if f.UserAgentProvider != nil {
		ua = f.UserAgentProvider.GetUserAgent(u)
	} else {
		ua = defaultUserAgentProvider(u)
	}

	req.Header.Set("User-Agent", ua)

	// url has no scheme, its a local request
	if !u.IsAbs() {
		req.RemoteAddr = f.AddrLocal
	}

	for h, v := range reqInit.Headers {
		headerName := http.CanonicalHeaderKey(h)
		req.Header.Set(headerName, v)
	}

	if reqInit.Method != "" {
		req.Method = strings.ToUpper(reqInit.Method)
	} else {
		req.Method = "GET"
	}

	switch r := strings.ToLower(reqInit.Redirect); r {
	case "error", "follow", "manual":
		req.Redirect = r
	case "":
		req.Redirect = internal.RequestRedirectFollow
	default:
		return nil, fmt.Errorf("unsupported redirect: %s", reqInit.Redirect)
	}

	return req, nil
}

func (f *Fetch) fetchLocal(r *internal.Request) (*internal.Response, error) {
	if f.LocalHandler == nil {
		return nil, errors.New("no local handler present")
	}

	var body io.Reader
	if r.Method != "GET" {
		body = r.Body
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.RemoteAddr = r.RemoteAddr
	req.Header = r.Header

	rcd := httptest.NewRecorder()

	f.LocalHandler.ServeHTTP(rcd, req)

	return internal.HandleHttpResponse(rcd.Result(), r.URL.String(), false)
}

func (f *Fetch) fetchRemote(r *internal.Request) (*internal.Response, error) {
	var body io.Reader
	if r.Method != "GET" {
		body = r.Body
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	redirected := false
	client := &http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			switch r.Redirect {
			case internal.RequestRedirectError:
				return errors.New("redirects are not allowed")
			default:
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
			}

			redirected = true
			return nil
		},
		Timeout: 20 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return internal.HandleHttpResponse(res, r.URL.String(), redirected)
}

func newResponseObject(ctx *v8go.Context, res *internal.Response) (*v8go.Object, error) {
	iso := ctx.Isolate()

	headers, err := newHeadersObject(ctx, res.Header)
	if err != nil {
		return nil, err
	}

	textFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		resolver, _ := v8go.NewPromiseResolver(ctx)

		go func() {
			defer res.BodyReader.Close()
			resBody, err := ioutil.ReadAll(res.BodyReader)
			if err != nil {
				resBody = nil
			}
			res.Body = string(resBody)
			//fmt.Println("respbody =>", res.Body)
			v, _ := v8go.NewValue(iso, res.Body)
			resolver.Resolve(v)
		}()

		return resolver.GetPromise().Value
	})
	if err != nil {
		return nil, err
	}

	jsonFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()

		resolver, _ := v8go.NewPromiseResolver(ctx)

		go func() {
			defer res.BodyReader.Close()
			resBody, err := ioutil.ReadAll(res.BodyReader)
			if err != nil {
				resBody = nil
			}
			res.Body = string(resBody)
			val, err := v8go.JSONParse(ctx, res.Body)
			if err != nil {
				rejectVal, _ := v8go.NewValue(iso, err.Error())
				resolver.Reject(rejectVal)
				return
			}

			resolver.Resolve(val)
		}()

		return resolver.GetPromise().Value
	})
	if err != nil {
		return nil, err
	}

	resTmp := v8go.NewObjectTemplate(iso)

	for _, f := range []struct {
		Name string
		Tmp  interface{}
	}{
		{Name: "text", Tmp: textFnTmp},
		{Name: "json", Tmp: jsonFnTmp},
	} {
		if err := resTmp.Set(f.Name, f.Tmp, v8go.ReadOnly); err != nil {
			return nil, err
		}
	}

	resObj, err := resTmp.NewInstance(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range []struct {
		Key string
		Val interface{}
	}{
		{Key: "headers", Val: headers},
		{Key: "ok", Val: res.OK},
		{Key: "redirected", Val: res.Redirected},
		{Key: "status", Val: res.Status},
		{Key: "statusText", Val: res.StatusText},
		{Key: "url", Val: res.URL},
		{Key: "body", Val: res.Body},
	} {
		//fmt.Println(v.Key, v.Val)
		if err := resObj.Set(v.Key, v.Val); err != nil {
			return nil, err
		}
	}

	return resObj, nil
}

func newHeadersObject(ctx *v8go.Context, h http.Header) (*v8go.Object, error) {
	iso := ctx.Isolate()
	// https://developer.mozilla.org/en-US/docs/Web/API/Headers/get
	getFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) <= 0 {
			// TODO: this should return an error, but v8go not supported now
			val, _ := v8go.NewValue(iso, "")
			return val
		}

		key := http.CanonicalHeaderKey(args[0].String())
		val, _ := v8go.NewValue(iso, h.Get(key))
		return val
	})

	// https://developer.mozilla.org/en-US/docs/Web/API/Headers/has
	hasFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) <= 0 {
			val, _ := v8go.NewValue(iso, false)
			return val
		}
		key := http.CanonicalHeaderKey(args[0].String())

		val, _ := v8go.NewValue(iso, h.Get(key) != "")
		return val
	})

	// create a header template,
	// TODO: if v8go supports Map in the future, change this to a Map Object
	headersTmp := v8go.NewObjectTemplate(iso)

	for _, f := range []struct {
		Name string
		Tmp  interface{}
	}{
		{Name: "get", Tmp: getFnTmp},
		{Name: "has", Tmp: hasFnTmp},
	} {
		if err := headersTmp.Set(f.Name, f.Tmp, v8go.ReadOnly); err != nil {
			return nil, err
		}
	}

	headers, err := headersTmp.NewInstance(ctx)
	if err != nil {
		return nil, err
	}

	for k, v := range h {
		var vv string
		if len(v) > 0 {
			// get the first element, like http.Header.Get
			vv = v[0]
		}

		if err := headers.Set(k, vv); err != nil {
			return nil, err
		}
	}

	return headers, nil
}

// v8go currently not support reject a *v8go.Object,
// so we should new *v8go.Value here
func newErrorValue(ctx *v8go.Context, err error) *v8go.Value {
	iso := ctx.Isolate()
	e, _ := v8go.NewValue(iso, fmt.Sprintf("fetch: %v", err))
	return e
}

func UserAgent() string {
	return fmt.Sprintf("v8go-polyfills/%s (v8go/%s)", Version, v8go.Version())
}

func getRequestInit(ctx *v8go.Context, options *v8go.Value) (*internal.RequestInit, error) {
	str, err := v8go.JSONStringify(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("Failed to JSONStringify: %v", err)
	}

	var reqInit internal.RequestInit
	reader := strings.NewReader(str)
	if err := json.NewDecoder(reader).Decode(&reqInit); err != nil {
		if c, ok := err.(*json.UnmarshalTypeError); ok && c.Struct == "RequestInit" && c.Field == "headers" {
			//this happens when invoked as: fetch(url, options) with JS headers
			reqInitObj, err := options.AsObject()
			if err != nil {
				return nil, fmt.Errorf("Error parsing args[1] as JSON: %#v", err)
			}

			//assign method from reqInit
			if reqInitObj.Has("method") {
				method, err := reqInitObj.Get("method")
				if err != nil {
					return nil, fmt.Errorf("Error parsing method from args[1] as obj: %#v", err)
				}
				reqInit.Method = method.String()
			}

			//assign body from reqInit
			if reqInitObj.Has("body") {
				body, err := reqInitObj.Get("body")
				if err != nil {
					return nil, fmt.Errorf("Error parsing body from args[1] as obj: %#v", err)
				}
				if body.String() != "null" {
					reqInit.Body = body.String()
				}
			}
			//assign redirect from reqInit
			if reqInitObj.Has("redirect") {
				redirect, err := reqInitObj.Get("redirect")
				if err != nil {
					return nil, fmt.Errorf("Error parsing redirect from args[1] as obj: %#v", err)
				}
				reqInit.Redirect = redirect.String()
			}

			//assign headers from reqInit
			if reqInitObj.Has("headers") {
				headersVal, err := reqInitObj.Get("headers")
				if err != nil {
					return nil, fmt.Errorf("Error parsing headers from args[1] as value: %#v", err)
				}

				if headersVal.String() != "null" {
					headers, err := headersVal.AsObject()
					if err != nil {
						return nil, fmt.Errorf("Error parsing headers from args[1] as obj: %#v", err)
					}
					if headers.Has("map") {
						headersMap, err := headers.Get("map")
						if err != nil {
							return nil, fmt.Errorf("Error parsing headersMap from args[1] as obj: %#v", err)
						}
						if headersMap.String() != "null" {
							headersMapString, err := headersMap.MarshalJSON()
							if err != nil {
								return nil, fmt.Errorf("Error marshalling headersMap from args[1] as json: %#v", err)
							}
							var goHeadersMap map[string]string
							reader := strings.NewReader(string(headersMapString))
							if err := json.NewDecoder(reader).Decode(&goHeadersMap); err != nil {
								return nil, fmt.Errorf("Error decoding JSON args[1].headers.map as map[string]string: %#v", err)
							}
							reqInit.Headers = goHeadersMap
						}
					}
				}
			}
		} else {
			return nil, fmt.Errorf("Failed to Decode: %v", err)
		}
	}

	return &reqInit, nil

}

func RequestCallbackFunc(info *v8go.FunctionCallbackInfo) *v8go.Value {
	args := info.Args()
	ctx := info.Context()
	iso := ctx.Isolate()
	if len(args) <= 0 {
		strErr, _ := v8go.NewValue(iso, "1 argument required, but only 0 present")
		return iso.ThrowException(strErr)
	}

	uri := args[0].String()
	u, err := url.Parse(uri)
	if err != nil {
		strErr, _ := v8go.NewValue(iso, fmt.Sprintf("Invalid URL %q: %#v", uri, err))
		return iso.ThrowException(strErr)
	}
	if u.Scheme == "" || u.Host == "" {
		strErr, _ := v8go.NewValue(iso, fmt.Sprintf("Invalid URL %q", uri))
		return iso.ThrowException(strErr)
	}
	res := &internal.JSRequestInit{
		Url: uri,
	}
	if len(args) > 1 {
		reqInit, err := getRequestInit(ctx, args[1])
		if err != nil {
			strErr, _ := v8go.NewValue(iso, fmt.Sprintf("Error getRequestInit: %#v", err))
			return iso.ThrowException(strErr)
		}
		if reqInit != nil {
			res.Method = reqInit.Method
			res.Body = reqInit.Body
			res.Redirect = reqInit.Redirect
			res.Headers = reqInit.Headers
		}
	}

	data, err := json.Marshal(res)
	if err != nil {
		strErr, _ := v8go.NewValue(iso, fmt.Sprintf("Error Marshalling: %#v", err))
		return iso.ThrowException(strErr)
	}
	v, err := v8go.JSONParse(info.Context(), string(data))
	if err != nil {
		strErr, _ := v8go.NewValue(iso, fmt.Sprintf("error jsonParse on result: %#v", err))
		return iso.ThrowException(strErr)
	}

	return v
}
