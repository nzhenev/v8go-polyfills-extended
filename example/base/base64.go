package main

import (
	stdBase64 "encoding/base64"
	"fmt"
	"os"

	"github.com/nzhenev/v8go"
	"github.com/nzhenev/v8go-polyfills-extended/base64"
	"github.com/nzhenev/v8go-polyfills-extended/console"
)

const script = `
let result = atob('CN7ySxWjjWNpbYPB1n/TBR8avujZS2cHdXRR5ZM7Fi6QBjlVzBqPu0CI9pQXcW9fvbMXdqU+57XY/QdGozJT19+gbQDM0ZQVzjwqtpyLorcPHjMqum+7dHD6XF5cXo3NZlKGsxcnvSVyClBDU5M1dUCe8bB9yV0wVM6ge+0WAmTX2GYbncilTjDw0bSJI1Z+71NT8UQCfmimKhVxJiKrnkaTrTw2Ma/1I2w4Dny3cRlFtCtob9cvNOeeIm8HtQoi/7HXoE0uFr1C39OL2hCC1TJsxX94djtNFqd9aUOPYrwT+zErSokSvbNYS5WpEjEpRJze9+TCV9NLmqCnARK4Bw'); 
let e = result.split("").map((c) => c.charCodeAt(0));
console.log("console value: ", e)
`

func main() {
	iso := v8go.NewIsolate()
	global := v8go.NewObjectTemplate(iso)

	if err := base64.InjectTo(iso, global); err != nil {
		panic(err)
	}

	ctx := v8go.NewContext(iso, global)
	if err := console.InjectTo(ctx, console.WithOutput(os.Stdout)); err != nil {
		panic(err)
	}

	_, err := ctx.RunScript(script, "fetch.js")
	if err != nil {
		panic(err)
	}
	//GO bytes
	byts, err := stdBase64.RawStdEncoding.DecodeString("CN7ySxWjjWNpbYPB1n/TBR8avujZS2cHdXRR5ZM7Fi6QBjlVzBqPu0CI9pQXcW9fvbMXdqU+57XY/QdGozJT19+gbQDM0ZQVzjwqtpyLorcPHjMqum+7dHD6XF5cXo3NZlKGsxcnvSVyClBDU5M1dUCe8bB9yV0wVM6ge+0WAmTX2GYbncilTjDw0bSJI1Z+71NT8UQCfmimKhVxJiKrnkaTrTw2Ma/1I2w4Dny3cRlFtCtob9cvNOeeIm8HtQoi/7HXoE0uFr1C39OL2hCC1TJsxX94djtNFqd9aUOPYrwT+zErSokSvbNYS5WpEjEpRJze9+TCV9NLmqCnARK4Bw")
	fmt.Println("go value: : ", byts)
}
