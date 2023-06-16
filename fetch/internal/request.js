
// HTTP methods whose capitalization should be normalized
var methods = ['DELETE', 'GET', 'HEAD', 'OPTIONS', 'POST', 'PUT']

function normalizeMethod(method) {
    var upcased = method.toUpperCase()
    return methods.indexOf(upcased) > -1 ? upcased : method
}

function Request(input, options) {
    options = options || {}
    var body = options.body

    console.log("request body", body) 
    if (input instanceof Request) {
        if (input.bodyUsed) {
            throw new TypeError('Already read')
        }
        this.url = input.url
        this.credentials = input.credentials
        if (!options.headers) {
            this.headers = new Headers(input.headers)
        }
        this.method = input.method
        this.mode = input.mode
        this.signal = input.signal
        if (!body && input._bodyInit != null) {
            body = input._bodyInit
            input.bodyUsed = true
        }
    } else {
        console.log("url", input)
        this.url = String(input)
    }

    this.credentials = options.credentials || this.credentials || 'same-origin'
    if (options.headers || !this.headers) {
        this.headers = new Headers(options.headers)
    }
    this.method = normalizeMethod(options.method || this.method || 'GET')
    this.mode = options.mode || this.mode || null
    this.signal = options.signal || this.signal
    this.referrer = null

    console.log("check body", body)
    if ((this.method === 'GET' || this.method === 'HEAD') && body) {
        throw new TypeError('Body not allowed for GET or HEAD requests')
    }

    console.log("init body", body)
    this._initBody(body)
}

Body.call(Request.prototype)

Request.prototype.clone = function () {
    console.log("request clone")
    return new Request(this, { body: this._bodyInit })
}