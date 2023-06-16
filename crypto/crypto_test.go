/*
 * Copyright (c) 2021 Twintag
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package crypto

import (
	"encoding/json"
	"testing"

	"github.com/nzhenev/v8go"

	"github.com/nzhenev/v8go-polyfills-extended/base64"
	"github.com/nzhenev/v8go-polyfills-extended/console"
	"github.com/nzhenev/v8go-polyfills-extended/fetch"
	"github.com/nzhenev/v8go-polyfills-extended/textEncoder"
)

func TestCrypto(t *testing.T) {
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	con := v8go.NewObjectTemplate(iso)
	if err := fetch.InjectTo(iso, con); err != nil {
		t.Error(err)
		return
	}
	ctx, err := textEncoder.InjectWith(iso, con)
	if err != nil {
		t.Error(err)
	}

	if err := InjectWith(iso, ctx); err != nil {
		t.Error(err)
		return
	}

	if err := console.InjectTo(ctx); err != nil {
		t.Error(err)
	}

	val, err := ctx.RunScript(`const fetchKeys = async () => {
		const data = await fetch('https://login.microsoftonline.com/24b080cd-5874-44ab-9862-8d7e0e0781ab/v2.0/.well-known/openid-configuration')
		const jwksResp = await data.json();
		//console.log(jwksResp.jwks_uri)
		const keysData = await fetch(jwksResp.jwks_uri);
		return await keysData.json();
	  };

	  epsilon = async (event) => {

let data = await fetchKeys()

const algo = {
  name: "RSA-OAEP",
  modulusLength: 4096,
  publicExponent: new Uint8Array([1, 0, 1]),
  hash: "SHA-256"
};
let importedKey = await crypto.subtle.importKey('jwk', data, algo, true, ["encrypt", "decrypt"]);
//console.log(importedKey.kid);

const token = 'eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Im5PbzNaRHJPRFhFSzFqS1doWHNsSFJfS1hFZyJ9.eyJhdWQiOiIwOGQ0NWY3Zi0xNmM5LTQ1ZGUtYmFkZC05NDc3ZGRjZTVlMzYiLCJpc3MiOiJodHRwczovL2xvZ2luLm1pY3Jvc29mdG9ubGluZS5jb20vZWMxMDAyZDctMDM0OC00MGFlLWFlNGUtOTBjMDA1MDZlYWNkL3YyLjAiLCJpYXQiOjE2MjM3NDMzNDcsIm5iZiI6MTYyMzc0MzM0NywiZXhwIjoxNjIzNzQ3MjQ3LCJhaW8iOiJBV1FBbS84VEFBQUFtbEJHSkpMT25BbllSMFRTVFUvdS9NQzhQNDdTSWJncWltY0xva1B0OFpLbGpvdW5IdmJ3M0d4UHJqTnlWNDkrbW5JTUhUNW84R2tXSHprelN4Qk5QSE0vWmVDWFdSc0FITmxtNW1oTURaWEMvWkJ5N3UwT0xVbTMrU0ZBSWR0NiIsImlkcCI6Imh0dHBzOi8vc3RzLndpbmRvd3MubmV0LzU2OTc1MzBiLWRmYzMtNDhlZi1iMWFjLTk4ZTRmZTY4MTI1Mi8iLCJuYW1lIjoiYmFja2tlbSIsIm9pZCI6IjIyNjZhMGYxLWFiZWYtNGRmNC1hY2UwLTNhZDk4NTcxOWRjMCIsInByZWZlcnJlZF91c2VybmFtZSI6Ik1pY2hpZWwuZGViYWNra2VyQHR3aW50YWcuY29tIiwicmgiOiIwLkFRd0Exd0lRN0VnRHJrQ3VUcERBQlFicXpYOWYxQWpKRnQ1RnV0MlVkOTNPWGpZTUFLVS4iLCJzdWIiOiJRR1BSamZrWUhaLTl2MFlrU2lQdm1YX3BhQTAzYzRDbGZrcUlkQWpoMDFvIiwidGlkIjoiZWMxMDAyZDctMDM0OC00MGFlLWFlNGUtOTBjMDA1MDZlYWNkIiwidXRpIjoiOHRaWlB1WHdQVUtPRUVLOGhraWxBQSIsInZlciI6IjIuMCJ9.j58zhFkqOPtcxB-gA1LdLYJYQw_oVZ2vDiZXD6M9nZNWbgAmFFkvN7CuhQFYR5rM9XaGrO-Rn4X6X389aFk-sZKQUOtVqmW4VT8_yT2iSGVspL5BcwWYeR0vEjO_5UNoavSunXz_qOFzzQqUYZ2-ex3KG9x7cL1Tc1kVv2JmAtUB-yK5t5yZU1BzNteIDCC4QEUa_vBxZrTwVEkRW_fT26TonWZTikYvi80COSFlMRiDD-gK2QFHrjcyPvhETTYDzXYhHoJDolcey59ERu9301SE9flTMigVpJlL5SreMIWhy1-vWt5lbCPOA246o3hEa_HAmAVgIdC1t1tSsj61hw'
const splitToken = token.split('.')

const encoder = new TextEncoder()
const signature = encoder.encode(splitToken[2])
console.log("signature", signature)
const payload = encoder.encode(splitToken[0]+'.'+splitToken[1])
console.log("payload", payload)

let isvalid = await crypto.subtle.verify(algo, importedKey, signature, payload)

//console.log(isvalid)

let returnVal = 'failed'
if (isvalid){
  returnVal = 'success'
}
return returnVal;
};
let res = epsilon();
Promise.resolve(res)`, "crypto.js")
	if err != nil {
		t.Error(err)
	}

	proms, err := val.AsPromise()
	if err != nil {
		t.Error(err)
		return
	}

	for proms.State() == v8go.Pending {
		continue
	}

	res := proms.Result().String()
	//fmt.Println("returned val is", res)
	if res != "success" {
		t.Errorf("Expected success; received %q", res)
	}

}

func TestImportKey(t *testing.T) {
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	con := v8go.NewObjectTemplate(iso)
	if err := fetch.InjectTo(iso, con); err != nil {
		t.Error(err)
		return
	}
	ctx, err := textEncoder.InjectWith(iso, con)
	if err != nil {
		t.Error(err)
	}

	if err := InjectWith(iso, ctx); err != nil {
		t.Error(err)
		return
	}

	if err := console.InjectTo(ctx); err != nil {
		t.Error(err)
	}

	val, err := ctx.RunScript(`

			epsilon = async (event) => {
				const key = {"kty":"RSA","use":"sig","kid":"l3sQ-50cCH4xBVZLHTGwnSR7680","x5t":"l3sQ-50cCH4xBVZLHTGwnSR7680","n":"sfsXMXWuO-dniLaIELa3Pyqz9Y_rWff_AVrCAnFSdPHa8__Pmkbt_yq-6Z3u1o4gjRpKWnrjxIh8zDn1Z1RS26nkKcNg5xfWxR2K8CPbSbY8gMrp_4pZn7tgrEmoLMkwfgYaVC-4MiFEo1P2gd9mCdgIICaNeYkG1bIPTnaqquTM5KfT971MpuOVOdM1ysiejdcNDvEb7v284PYZkw2imwqiBY3FR0sVG7jgKUotFvhd7TR5WsA20GS_6ZIkUUlLUbG_rXWGl0YjZLS_Uf4q8Hbo7u-7MaFn8B69F6YaFdDlXm_A0SpedVFWQFGzMsp43_6vEzjfrFDJVAYkwb6xUQ","e":"AQAB","x5c":["MIIDBTCCAe2gAwIBAgIQWPB1ofOpA7FFlOBk5iPaNTANBgkqhkiG9w0BAQsFADAtMSswKQYDVQQDEyJhY2NvdW50cy5hY2Nlc3Njb250cm9sLndpbmRvd3MubmV0MB4XDTIxMDIwNzE3MDAzOVoXDTI2MDIwNjE3MDAzOVowLTErMCkGA1UEAxMiYWNjb3VudHMuYWNjZXNzY29udHJvbC53aW5kb3dzLm5ldDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALH7FzF1rjvnZ4i2iBC2tz8qs/WP61n3/wFawgJxUnTx2vP/z5pG7f8qvumd7taOII0aSlp648SIfMw59WdUUtup5CnDYOcX1sUdivAj20m2PIDK6f+KWZ+7YKxJqCzJMH4GGlQvuDIhRKNT9oHfZgnYCCAmjXmJBtWyD052qqrkzOSn0/e9TKbjlTnTNcrIno3XDQ7xG+79vOD2GZMNopsKogWNxUdLFRu44ClKLRb4Xe00eVrANtBkv+mSJFFJS1Gxv611hpdGI2S0v1H+KvB26O7vuzGhZ/AevRemGhXQ5V5vwNEqXnVRVkBRszLKeN/+rxM436xQyVQGJMG+sVECAwEAAaMhMB8wHQYDVR0OBBYEFLlRBSxxgmNPObCFrl+hSsbcvRkcMA0GCSqGSIb3DQEBCwUAA4IBAQB+UQFTNs6BUY3AIGkS2ZRuZgJsNEr/ZEM4aCs2domd2Oqj7+5iWsnPh5CugFnI4nd+ZLgKVHSD6acQ27we+eNY6gxfpQCY1fiN/uKOOsA0If8IbPdBEhtPerRgPJFXLHaYVqD8UYDo5KNCcoB4Kh8nvCWRGPUUHPRqp7AnAcVrcbiXA/bmMCnFWuNNahcaAKiJTxYlKDaDIiPN35yECYbDj0PBWJUxobrvj5I275jbikkp8QSLYnSU/v7dMDUbxSLfZ7zsTuaF2Qx+L62PsYTwLzIFX3M8EMSQ6h68TupFTi5n0M2yIXQgoRoNEDWNJZ/aZMY/gqT02GQGBWrh+/vJ"],"issuer":"https://login.microsoftonline.com/24b080cd-5874-44ab-9862-8d7e0e0781ab/v2.0","alg":"RS256"}
				const algo = {"name":"RSASSA-PKCS1-v1_5","hash":"SHA-256"}
	let importedKey = await crypto.subtle.importKey('jwk', key, algo, false, ["verify"]);
	return importedKey
	};
	let res = epsilon();
	Promise.resolve(res)`, "crypto.js")
	if err != nil {
		t.Error(err)
	}

	proms, err := val.AsPromise()
	if err != nil {
		t.Error(err)
		return
	}

	for proms.State() == v8go.Pending {
		continue
	}
	if !proms.Result().IsObject() {
		t.Error("expected object type but got error")
	}

	res := proms.Result().Object()
	lala, err := res.MarshalJSON()
	if err != nil {
		t.Errorf("error in marshalling json %#v", err)
	}

	var jsonMap map[string]interface{}
	json.Unmarshal(lala, &jsonMap)
	if jsonMap == nil {
		t.Errorf("expected json but got nil")
	}
	if jsonMap["algorithm"] == nil {
		t.Errorf("expected algorith but got nil")
	}
}
func TestVerify(t *testing.T) {
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	con := v8go.NewObjectTemplate(iso)
	if err := fetch.InjectTo(iso, con); err != nil {
		t.Error(err)
		return
	}
	if err := base64.InjectTo(iso, con); err != nil {
		t.Error(err)
		return
	}
	err := textEncoder.InjectTo(iso, con)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := v8go.NewContext(iso, con)
	if err := InjectWith(iso, ctx); err != nil {
		t.Error(err)
		return
	}

	if err := console.InjectTo(ctx); err != nil {
		t.Error(err)
		return
	}

	val, err := ctx.RunScript(`
	epsilon = async (event) => {
		try{

			const key = { "kty": "RSA", "use": "sig", "kid": "l3sQ-50cCH4xBVZLHTGwnSR7680", "x5t": "l3sQ-50cCH4xBVZLHTGwnSR7680", "n": "sfsXMXWuO-dniLaIELa3Pyqz9Y_rWff_AVrCAnFSdPHa8__Pmkbt_yq-6Z3u1o4gjRpKWnrjxIh8zDn1Z1RS26nkKcNg5xfWxR2K8CPbSbY8gMrp_4pZn7tgrEmoLMkwfgYaVC-4MiFEo1P2gd9mCdgIICaNeYkG1bIPTnaqquTM5KfT971MpuOVOdM1ysiejdcNDvEb7v284PYZkw2imwqiBY3FR0sVG7jgKUotFvhd7TR5WsA20GS_6ZIkUUlLUbG_rXWGl0YjZLS_Uf4q8Hbo7u-7MaFn8B69F6YaFdDlXm_A0SpedVFWQFGzMsp43_6vEzjfrFDJVAYkwb6xUQ", "e": "AQAB", "x5c": ["MIIDBTCCAe2gAwIBAgIQWPB1ofOpA7FFlOBk5iPaNTANBgkqhkiG9w0BAQsFADAtMSswKQYDVQQDEyJhY2NvdW50cy5hY2Nlc3Njb250cm9sLndpbmRvd3MubmV0MB4XDTIxMDIwNzE3MDAzOVoXDTI2MDIwNjE3MDAzOVowLTErMCkGA1UEAxMiYWNjb3VudHMuYWNjZXNzY29udHJvbC53aW5kb3dzLm5ldDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALH7FzF1rjvnZ4i2iBC2tz8qs/WP61n3/wFawgJxUnTx2vP/z5pG7f8qvumd7taOII0aSlp648SIfMw59WdUUtup5CnDYOcX1sUdivAj20m2PIDK6f+KWZ+7YKxJqCzJMH4GGlQvuDIhRKNT9oHfZgnYCCAmjXmJBtWyD052qqrkzOSn0/e9TKbjlTnTNcrIno3XDQ7xG+79vOD2GZMNopsKogWNxUdLFRu44ClKLRb4Xe00eVrANtBkv+mSJFFJS1Gxv611hpdGI2S0v1H+KvB26O7vuzGhZ/AevRemGhXQ5V5vwNEqXnVRVkBRszLKeN/+rxM436xQyVQGJMG+sVECAwEAAaMhMB8wHQYDVR0OBBYEFLlRBSxxgmNPObCFrl+hSsbcvRkcMA0GCSqGSIb3DQEBCwUAA4IBAQB+UQFTNs6BUY3AIGkS2ZRuZgJsNEr/ZEM4aCs2domd2Oqj7+5iWsnPh5CugFnI4nd+ZLgKVHSD6acQ27we+eNY6gxfpQCY1fiN/uKOOsA0If8IbPdBEhtPerRgPJFXLHaYVqD8UYDo5KNCcoB4Kh8nvCWRGPUUHPRqp7AnAcVrcbiXA/bmMCnFWuNNahcaAKiJTxYlKDaDIiPN35yECYbDj0PBWJUxobrvj5I275jbikkp8QSLYnSU/v7dMDUbxSLfZ7zsTuaF2Qx+L62PsYTwLzIFX3M8EMSQ6h68TupFTi5n0M2yIXQgoRoNEDWNJZ/aZMY/gqT02GQGBWrh+/vJ"], "issuer": "https://login.microsoftonline.com/24b080cd-5874-44ab-9862-8d7e0e0781ab/v2.0", "alg": "RS256" }

			const algo = { "name": "RSASSA-PKCS1-v1_5", "hash": "SHA-256" }
			let importedKey = await crypto.subtle.importKey('jwk', key, algo, false, ["verify"]);
			let actualtoken =
				"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Imwzc1EtNTBjQ0g0eEJWWkxIVEd3blNSNzY4MCJ9.eyJhdWQiOiJlNjI0YTdjMi0zZTUzLTQ2NTktOGY5Yi1kN2MxOWZjZjAxZjciLCJpc3MiOiJodHRwczovL2xvZ2luLm1pY3Jvc29mdG9ubGluZS5jb20vMjRiMDgwY2QtNTg3NC00NGFiLTk4NjItOGQ3ZTBlMDc4MWFiL3YyLjAiLCJpYXQiOjE2MzkxMzU3NDYsIm5iZiI6MTYzOTEzNTc0NiwiZXhwIjoxNjM5MTM5NjQ2LCJuYW1lIjoiQXNoaXNoIFNoYXJtYSAoRGV2T24pIiwib2lkIjoiOTZmODM2N2QtY2M2NC00NjMwLWI0MGQtYTUwNTVjMjAwOGVkIiwicHJlZmVycmVkX3VzZXJuYW1lIjoiYXNoaXNoLnNoYXJtYUBkZXZvbi5ubCIsInJoIjoiMC5BUUlBellDd0pIUllxMFNZWW8xLURnZUJxOEtuSk9aVFBsbEdqNXZYd1pfUEFmY0NBTzguIiwic3ViIjoiLVNDRE5lR2IwVVc1TzZ5NkoxMERyNWhFZWxIR0lSdU5uNnd3NTZuMHRyMCIsInRpZCI6IjI0YjA4MGNkLTU4NzQtNDRhYi05ODYyLThkN2UwZTA3ODFhYiIsInV0aSI6InJfZERnTWdncGtPN01pQnhPNndTQUEiLCJ2ZXIiOiIyLjAifQ.pG8hmgZAFrtaqOIGgc2eYkZo_Xxb0_0ntgky5fZsAg5QpHoh7f6EfufNaTEcGY4oFMcE9ii5TI5S9-0LJqJcHRHpOd8xcWMQ-YWe5DAEag_90uRubFXDuLnK4mHGrEWRdlkO0vp5YDmhUItykyq_GVMzwBmbRKhRWzVxEao9dsXFZrnrTkQE2rdtE81w5kAvhEnYB8q7Yfy0uRN-7U2wLyazy_TqYfx19tDBN66F32rlV8SdTxvewRj4ZMw12_RBLdaiSoj6phMpjOllFgLdUi1RY-roJjNcdaHH988aopZgqnQvTQ8zOczES0tBqt-oRJyrwOhFvpvTZAgg5-ykUg";
			
		
			const splitToken = actualtoken.split(".");
			const sign = new Uint8Array(
				atob(splitToken[2].replace(/-/g, "+").replace(/_/g, "/").replace(/\s/g, ""))
					.split("")
					.map((c) => c.charCodeAt(0))
			);
			const encoder = new TextEncoder();
			concat = (...buffers)=> {
				const size = buffers.reduce((acc, { length }) => acc + length, 0);
				const buf = new Uint8Array(size);
				let i = 0;
				buffers.forEach((buffer) => {
				  buf.set(buffer, i);
				  i += buffer.length;
				});
				return buf;
			  }
			const jws = {
				payload: splitToken[1],
				signature: splitToken[2],
				protected: splitToken[0],
			}

			const data = concat(
				encoder.encode((_a = jws.protected) !== null && _a !== void 0 ? _a : ""),
				encoder.encode("."),
				typeof jws.payload === "string"
					? encoder.encode(jws.payload)
					: jws.payload
			);

			let verfied = await crypto.subtle.verify(algo, importedKey, sign, data);
			return verfied
		}catch (err){
			return JSON.stringify(err)
		}
	};
	let res = epsilon();
	Promise.resolve(res)
	`, "crypto.js")
	if err != nil {
		t.Error(err)
	}

	proms, err := val.AsPromise()
	if err != nil {
		t.Error(err)
		return
	}

	for proms.State() == v8go.Pending {
		continue
	}
	if !proms.Result().IsBoolean() {
		t.Error("expected boolean in result, but got error")
	}

	res := proms.Result().Boolean()
	if !res {
		t.Errorf("expected true, but got %t", res)
	}

}
