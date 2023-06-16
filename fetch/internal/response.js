function Response(body, init){
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