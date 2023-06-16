Zaza.setHost("http://twintag");
Zaza.setAdminHost("http://admin");
(function (exports, Zaza) {
  "use strict";

  function _interopNamespace(e) {
    if (e && e.__esModule) return e;
    var n = Object.create(null);
    if (e) {
      Object.keys(e).forEach(function (k) {
        if (k !== "default") {
          var d = Object.getOwnPropertyDescriptor(e, k);
          Object.defineProperty(
            n,
            k,
            d.get
              ? d
              : {
                  enumerable: true,
                  get: function () {
                    return e[k];
                  },
                }
          );
        }
      });
    }
    n["default"] = e;
    return Object.freeze(n);
  }

  var Zaza__namespace = /*#__PURE__*/ _interopNamespace(Zaza);

  class Alertis {
    /**
     *
     */
    constructor(event) {
      this.event = event;
    }
    get project() {
      if (!this._project) {
        this._project = new Zaza__namespace.Project(this.event.project.apiKey);
      }
      return this._project;
    }
    get view() {
      if (!this._view) {
        this._view = new Zaza__namespace.View(this.event.view.qid);
      }
      return this._view;
    }
  }

  const encoder = new TextEncoder();
  const decoder = new TextDecoder();
  function concat(...buffers) {
    const size = buffers.reduce((acc, { length }) => acc + length, 0);
    const buf = new Uint8Array(size);
    let i = 0;
    buffers.forEach((buffer) => {
      buf.set(buffer, i);
      i += buffer.length;
    });
    return buf;
  }

  const decodeBase64 = (encoded) => {
    return new Uint8Array(
      atob(encoded)
        .split("")
        .map((c) => c.charCodeAt(0))
    );
  };
  const decode = (input) => {
    let encoded = input;
    if (encoded instanceof Uint8Array) {
      encoded = decoder.decode(encoded);
    }
    encoded = encoded.replace(/-/g, "+").replace(/_/g, "/").replace(/\s/g, "");
    try {
      return decodeBase64(encoded);
    } catch (_a) {
      throw new TypeError("The input to be decoded is not correctly encoded.");
    }
  };

  class JOSEError extends Error {
    constructor(message) {
      var _a;
      super(message);
      this.code = "ERR_JOSE_GENERIC";
      this.name = this.constructor.name;
      (_a = Error.captureStackTrace) === null || _a === void 0
        ? void 0
        : _a.call(Error, this, this.constructor);
    }
    static get code() {
      return "ERR_JOSE_GENERIC";
    }
  }
  class JOSEAlgNotAllowed extends JOSEError {
    constructor() {
      super(...arguments);
      this.code = "ERR_JOSE_ALG_NOT_ALLOWED";
    }
    static get code() {
      return "ERR_JOSE_ALG_NOT_ALLOWED";
    }
  }
  class JOSENotSupported extends JOSEError {
    constructor() {
      super(...arguments);
      this.code = "ERR_JOSE_NOT_SUPPORTED";
    }
    static get code() {
      return "ERR_JOSE_NOT_SUPPORTED";
    }
  }
  class JWSInvalid extends JOSEError {
    constructor() {
      super(...arguments);
      this.code = "ERR_JWS_INVALID";
    }
    static get code() {
      return "ERR_JWS_INVALID";
    }
  }
  class JWSSignatureVerificationFailed extends JOSEError {
    constructor() {
      super(...arguments);
      this.code = "ERR_JWS_SIGNATURE_VERIFICATION_FAILED";
      this.message = "signature verification failed";
    }
    static get code() {
      return "ERR_JWS_SIGNATURE_VERIFICATION_FAILED";
    }
  }

  var crypto$1 = crypto;
  function isCryptoKey(key) {
    try {
      return (
        key != null &&
        typeof key.extractable === "boolean" &&
        typeof key.algorithm.name === "string" &&
        typeof key.type === "string"
      );
    } catch (_a) {
      return false;
    }
  }

  function isCloudflareWorkers() {
    return typeof WebSocketPair === "function";
  }
  function isNodeJs() {
    try {
      return process.versions.node !== undefined;
    } catch (_a) {
      return false;
    }
  }

  function unusable(name, prop = "algorithm.name") {
    return new TypeError(
      `CryptoKey does not support this operation, its ${prop} must be ${name}`
    );
  }
  function isAlgorithm(algorithm, name) {
    return algorithm.name === name;
  }
  function getHashLength(hash) {
    return parseInt(hash.name.substr(4), 10);
  }
  function getNamedCurve(alg) {
    switch (alg) {
      case "ES256":
        return "P-256";
      case "ES384":
        return "P-384";
      case "ES512":
        return "P-521";
      default:
        throw new Error("unreachable");
    }
  }
  function checkUsage(key, usages) {
    if (
      usages.length &&
      !usages.some((expected) => key.usages.includes(expected))
    ) {
      let msg =
        "CryptoKey does not support this operation, its usages must include ";
      if (usages.length > 2) {
        const last = usages.pop();
        msg += `one of ${usages.join(", ")}, or ${last}.`;
      } else if (usages.length === 2) {
        msg += `one of ${usages[0]} or ${usages[1]}.`;
      } else {
        msg += `${usages[0]}.`;
      }
      throw new TypeError(msg);
    }
  }
  function checkSigCryptoKey(key, alg, ...usages) {
    switch (alg) {
      case "HS256":
      case "HS384":
      case "HS512": {
        if (!isAlgorithm(key.algorithm, "HMAC")) throw unusable("HMAC");
        const expected = parseInt(alg.substr(2), 10);
        const actual = getHashLength(key.algorithm.hash);
        if (actual !== expected)
          throw unusable(`SHA-${expected}`, "algorithm.hash");
        break;
      }
      case "RS256":
      case "RS384":
      case "RS512": {
        if (!isAlgorithm(key.algorithm, "RSASSA-PKCS1-v1_5"))
          throw unusable("RSASSA-PKCS1-v1_5");
        const expected = parseInt(alg.substr(2), 10);
        const actual = getHashLength(key.algorithm.hash);
        if (actual !== expected)
          throw unusable(`SHA-${expected}`, "algorithm.hash");
        break;
      }
      case "PS256":
      case "PS384":
      case "PS512": {
        if (!isAlgorithm(key.algorithm, "RSA-PSS")) throw unusable("RSA-PSS");
        const expected = parseInt(alg.substr(2), 10);
        const actual = getHashLength(key.algorithm.hash);
        if (actual !== expected)
          throw unusable(`SHA-${expected}`, "algorithm.hash");
        break;
      }
      case isNodeJs() && "EdDSA": {
        if (
          key.algorithm.name !== "NODE-ED25519" &&
          key.algorithm.name !== "NODE-ED448"
        )
          throw unusable("NODE-ED25519 or NODE-ED448");
        break;
      }
      case isCloudflareWorkers() && "EdDSA": {
        if (!isAlgorithm(key.algorithm, "NODE-ED25519"))
          throw unusable("NODE-ED25519");
        break;
      }
      case "ES256":
      case "ES384":
      case "ES512": {
        if (!isAlgorithm(key.algorithm, "ECDSA")) throw unusable("ECDSA");
        const expected = getNamedCurve(alg);
        const actual = key.algorithm.namedCurve;
        if (actual !== expected)
          throw unusable(expected, "algorithm.namedCurve");
        break;
      }
      default:
        throw new TypeError("CryptoKey does not support this operation");
    }
    checkUsage(key, usages);
  }

  var invalidKeyInput = (actual, ...types) => {
    let msg = "Key must be ";
    if (types.length > 2) {
      const last = types.pop();
      msg += `one of type ${types.join(", ")}, or ${last}.`;
    } else if (types.length === 2) {
      msg += `one of type ${types[0]} or ${types[1]}.`;
    } else {
      msg += `of type ${types[0]}.`;
    }
    if (actual == null) {
      msg += ` Received ${actual}`;
    } else if (typeof actual === "function" && actual.name) {
      msg += ` Received function ${actual.name}`;
    } else if (typeof actual === "object" && actual != null) {
      if (actual.constructor && actual.constructor.name) {
        msg += ` Received an instance of ${actual.constructor.name}`;
      }
    }
    return msg;
  };

  var isKeyLike = (key) => {
    return isCryptoKey(key);
  };
  const types = ["CryptoKey"];

  const isDisjoint = (...headers) => {
    const sources = headers.filter(Boolean);
    if (sources.length === 0 || sources.length === 1) {
      return true;
    }
    let acc;
    for (const header of sources) {
      const parameters = Object.keys(header);
      if (!acc || acc.size === 0) {
        acc = new Set(parameters);
        continue;
      }
      for (const parameter of parameters) {
        if (acc.has(parameter)) {
          return false;
        }
        acc.add(parameter);
      }
    }
    return true;
  };

  function isObjectLike(value) {
    return typeof value === "object" && value !== null;
  }
  function isObject(input) {
    if (
      !isObjectLike(input) ||
      Object.prototype.toString.call(input) !== "[object Object]"
    ) {
      return false;
    }
    if (Object.getPrototypeOf(input) === null) {
      return true;
    }
    let proto = input;
    while (Object.getPrototypeOf(proto) !== null) {
      proto = Object.getPrototypeOf(proto);
    }
    return Object.getPrototypeOf(input) === proto;
  }

  var checkKeyLength = (alg, key) => {
    if (alg.startsWith("RS") || alg.startsWith("PS")) {
      const { modulusLength } = key.algorithm;
      if (typeof modulusLength !== "number" || modulusLength < 2048) {
        throw new TypeError(
          `${alg} requires key modulusLength to be 2048 bits or larger`
        );
      }
    }
  };

  function subtleMapping(jwk) {
    let algorithm;
    let keyUsages;
    switch (jwk.kty) {
      case "oct": {
        switch (jwk.alg) {
          case "HS256":
          case "HS384":
          case "HS512":
            algorithm = { name: "HMAC", hash: `SHA-${jwk.alg.substr(-3)}` };
            keyUsages = ["sign", "verify"];
            break;
          case "A128CBC-HS256":
          case "A192CBC-HS384":
          case "A256CBC-HS512":
            throw new JOSENotSupported(
              `${jwk.alg} keys cannot be imported as CryptoKey instances`
            );
          case "A128GCM":
          case "A192GCM":
          case "A256GCM":
          case "A128GCMKW":
          case "A192GCMKW":
          case "A256GCMKW":
            algorithm = { name: "AES-GCM" };
            keyUsages = ["encrypt", "decrypt"];
            break;
          case "A128KW":
          case "A192KW":
          case "A256KW":
            algorithm = { name: "AES-KW" };
            keyUsages = ["wrapKey", "unwrapKey"];
            break;
          case "PBES2-HS256+A128KW":
          case "PBES2-HS384+A192KW":
          case "PBES2-HS512+A256KW":
            algorithm = { name: "PBKDF2" };
            keyUsages = ["deriveBits"];
            break;
          default:
            throw new JOSENotSupported(
              'Invalid or unsupported JWK "alg" (Algorithm) Parameter value'
            );
        }
        break;
      }
      case "RSA": {
        switch (jwk.alg) {
          case "PS256":
          case "PS384":
          case "PS512":
            algorithm = { name: "RSA-PSS", hash: `SHA-${jwk.alg.substr(-3)}` };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          case "RS256":
          case "RS384":
          case "RS512":
            algorithm = {
              name: "RSASSA-PKCS1-v1_5",
              hash: `SHA-${jwk.alg.substr(-3)}`,
            };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          case "RSA-OAEP":
          case "RSA-OAEP-256":
          case "RSA-OAEP-384":
          case "RSA-OAEP-512":
            algorithm = {
              name: "RSA-OAEP",
              hash: `SHA-${parseInt(jwk.alg.substr(-3), 10) || 1}`,
            };
            keyUsages = jwk.d
              ? ["decrypt", "unwrapKey"]
              : ["encrypt", "wrapKey"];
            break;
          default:
            throw new JOSENotSupported(
              'Invalid or unsupported JWK "alg" (Algorithm) Parameter value'
            );
        }
        break;
      }
      case "EC": {
        switch (jwk.alg) {
          case "ES256":
            algorithm = { name: "ECDSA", namedCurve: "P-256" };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          case "ES384":
            algorithm = { name: "ECDSA", namedCurve: "P-384" };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          case "ES512":
            algorithm = { name: "ECDSA", namedCurve: "P-521" };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          case "ECDH-ES":
          case "ECDH-ES+A128KW":
          case "ECDH-ES+A192KW":
          case "ECDH-ES+A256KW":
            algorithm = { name: "ECDH", namedCurve: jwk.crv };
            keyUsages = jwk.d ? ["deriveBits"] : [];
            break;
          default:
            throw new JOSENotSupported(
              'Invalid or unsupported JWK "alg" (Algorithm) Parameter value'
            );
        }
        break;
      }
      case (isCloudflareWorkers() || isNodeJs()) && "OKP":
        if (jwk.alg !== "EdDSA") {
          throw new JOSENotSupported(
            'Invalid or unsupported JWK "alg" (Algorithm) Parameter value'
          );
        }
        switch (jwk.crv) {
          case "Ed25519":
            algorithm = { name: "NODE-ED25519", namedCurve: "NODE-ED25519" };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          case isNodeJs() && "Ed448":
            algorithm = { name: "NODE-ED448", namedCurve: "NODE-ED448" };
            keyUsages = jwk.d ? ["sign"] : ["verify"];
            break;
          default:
            throw new JOSENotSupported(
              'Invalid or unsupported JWK "crv" (Subtype of Key Pair) Parameter value'
            );
        }
        break;
      default:
        throw new JOSENotSupported(
          'Invalid or unsupported JWK "kty" (Key Type) Parameter value'
        );
    }
    return { algorithm, keyUsages };
  }
  const parse = async (jwk) => {
    console.log("inside parse ", JSON.stringify(jwk))
    var _a, _b;
    const { algorithm, keyUsages } = subtleMapping(jwk);
    const rest = [
      algorithm,
      (_a = jwk.ext) !== null && _a !== void 0 ? _a : false,
      (_b = jwk.key_ops) !== null && _b !== void 0 ? _b : keyUsages,
    ];
    if (algorithm.name === "PBKDF2") {
      return crypto$1.subtle.importKey("raw", decode(jwk.k), ...rest);
    }
    const keyData = { ...jwk };
    delete keyData.alg;
    console.log("___ importkey ", JSON.stringify(rest))
    let res = await crypto$1.subtle.importKey("jwk", keyData, ...rest).catch(err=>{
        console.log("_______ jwk ", JSON.stringify(err))
    });
    return res
  };

  async function importJWK(jwk, alg, octAsKeyObject) {
    if (!isObject(jwk)) {
      console.log("JWK must be an object")
      throw new TypeError("JWK must be an object");
    }
    alg || (alg = jwk.alg);
    if (typeof alg !== "string" || !alg) {
      console.log('"alg" argument is required when "jwk.alg" is not present')
      throw new TypeError(
        '"alg" argument is required when "jwk.alg" is not present'
      );
    }
    console.log("_ jwk.kty ", jwk.kty)
    switch (jwk.kty) {
      case "oct":
        console.log("case oct")
        if (typeof jwk.k !== "string" || !jwk.k) {
          throw new TypeError('missing "k" (Key Value) Parameter value');
        }
        octAsKeyObject !== null && octAsKeyObject !== void 0
          ? octAsKeyObject
          : (octAsKeyObject = jwk.ext !== true);
        if (octAsKeyObject) {
          return parse({ ...jwk, alg, ext: false });
        }
        return decode(jwk.k);
      case "RSA":
        console.log("case rsa")
        if (jwk.oth !== undefined) {
          throw new JOSENotSupported(
            'RSA JWK "oth" (Other Primes Info) Parameter value is not supported'
          );
        }
      case "EC":
      case "OKP": {
        console.log("case okp")
        console.log("going in parse" )
        return parse({ ...jwk, alg });
      }
      default:
        console.log("case deafault")
        throw new JOSENotSupported(
          'Unsupported "kty" (Key Type) Parameter value'
        );
    }
  }

  const symmetricTypeCheck = (key) => {
    if (key instanceof Uint8Array) return;
    if (!isKeyLike(key)) {
      throw new TypeError(invalidKeyInput(key, ...types, "Uint8Array"));
    }
    if (key.type !== "secret") {
      throw new TypeError(
        `${types.join(
          " or "
        )} instances for symmetric algorithms must be of type "secret"`
      );
    }
  };
  const asymmetricTypeCheck = (key, usage) => {
    if (!isKeyLike(key)) {
      throw new TypeError(invalidKeyInput(key, ...types));
    }
    if (key.type === "secret") {
      throw new TypeError(
        `${types.join(
          " or "
        )} instances for asymmetric algorithms must not be of type "secret"`
      );
    }
    if (usage === "sign" && key.type === "public") {
      throw new TypeError(
        `${types.join(
          " or "
        )} instances for asymmetric algorithm signing must be of type "private"`
      );
    }
    if (usage === "decrypt" && key.type === "public") {
      throw new TypeError(
        `${types.join(
          " or "
        )} instances for asymmetric algorithm decryption must be of type "private"`
      );
    }
    if (key.algorithm && usage === "verify" && key.type === "private") {
      throw new TypeError(
        `${types.join(
          " or "
        )} instances for asymmetric algorithm verifying must be of type "public"`
      );
    }
    if (key.algorithm && usage === "encrypt" && key.type === "private") {
      throw new TypeError(
        `${types.join(
          " or "
        )} instances for asymmetric algorithm encryption must be of type "public"`
      );
    }
  };
  const checkKeyType = (alg, key, usage) => {
    const symmetric =
      alg.startsWith("HS") ||
      alg === "dir" ||
      alg.startsWith("PBES2") ||
      /^A\d{3}(?:GCM)?KW$/.test(alg);
    if (symmetric) {
      symmetricTypeCheck(key);
    } else {
      asymmetricTypeCheck(key, usage);
    }
  };

  function validateCrit(
    Err,
    recognizedDefault,
    recognizedOption,
    protectedHeader,
    joseHeader
  ) {
    if (joseHeader.crit !== undefined && protectedHeader.crit === undefined) {
      throw new Err(
        '"crit" (Critical) Header Parameter MUST be integrity protected'
      );
    }
    if (!protectedHeader || protectedHeader.crit === undefined) {
      return new Set();
    }
    if (
      !Array.isArray(protectedHeader.crit) ||
      protectedHeader.crit.length === 0 ||
      protectedHeader.crit.some(
        (input) => typeof input !== "string" || input.length === 0
      )
    ) {
      throw new Err(
        '"crit" (Critical) Header Parameter MUST be an array of non-empty strings when present'
      );
    }
    let recognized;
    if (recognizedOption !== undefined) {
      recognized = new Map([
        ...Object.entries(recognizedOption),
        ...recognizedDefault.entries(),
      ]);
    } else {
      recognized = recognizedDefault;
    }
    for (const parameter of protectedHeader.crit) {
      if (!recognized.has(parameter)) {
        throw new JOSENotSupported(
          `Extension Header Parameter "${parameter}" is not recognized`
        );
      }
      if (joseHeader[parameter] === undefined) {
        throw new Err(`Extension Header Parameter "${parameter}" is missing`);
      } else if (
        recognized.get(parameter) &&
        protectedHeader[parameter] === undefined
      ) {
        throw new Err(
          `Extension Header Parameter "${parameter}" MUST be integrity protected`
        );
      }
    }
    return new Set(protectedHeader.crit);
  }

  const validateAlgorithms = (option, algorithms) => {
    if (
      algorithms !== undefined &&
      (!Array.isArray(algorithms) ||
        algorithms.some((s) => typeof s !== "string"))
    ) {
      throw new TypeError(`"${option}" option must be an array of strings`);
    }
    if (!algorithms) {
      return undefined;
    }
    return new Set(algorithms);
  };

  function subtleDsa(alg, namedCurve) {
    const length = parseInt(alg.substr(-3), 10);
    switch (alg) {
      case "HS256":
      case "HS384":
      case "HS512":
        return { hash: `SHA-${length}`, name: "HMAC" };
      case "PS256":
      case "PS384":
      case "PS512":
        return {
          hash: `SHA-${length}`,
          name: "RSA-PSS",
          saltLength: length >> 3,
        };
      case "RS256":
      case "RS384":
      case "RS512":
        return { hash: `SHA-${length}`, name: "RSASSA-PKCS1-v1_5" };
      case "ES256":
      case "ES384":
      case "ES512":
        return { hash: `SHA-${length}`, name: "ECDSA", namedCurve };
      case (isCloudflareWorkers() || isNodeJs()) && "EdDSA":
        return { name: namedCurve, namedCurve };
      default:
        throw new JOSENotSupported(
          `alg ${alg} is not supported either by JOSE or your javascript runtime`
        );
    }
  }

  function getCryptoKey(alg, key, usage) {
    if (isCryptoKey(key)) {
      checkSigCryptoKey(key, alg, usage);
      return key;
    }
    if (key instanceof Uint8Array) {
      if (!alg.startsWith("HS")) {
        throw new TypeError(invalidKeyInput(key, ...types));
      }
      return crypto$1.subtle.importKey(
        "raw",
        key,
        { hash: `SHA-${alg.substr(-3)}`, name: "HMAC" },
        false,
        [usage]
      );
    }
    throw new TypeError(invalidKeyInput(key, ...types, "Uint8Array"));
  }

  const verify = async (alg, key, signature, data) => {
    const cryptoKey = await getCryptoKey(alg, key, "verify");
    checkKeyLength(alg, cryptoKey);
    const algorithm = subtleDsa(alg, cryptoKey.algorithm.namedCurve);
    try {
      return await crypto$1.subtle.verify(
        algorithm,
        cryptoKey,
        signature,
        data
      );
    } catch (_a) {
      return false;
    }
  };

  async function flattenedVerify(jws, key, options) {
    var _a;
    if (!isObject(jws)) {
      throw new JWSInvalid("Flattened JWS must be an object");
    }
    if (jws.protected === undefined && jws.header === undefined) {
      throw new JWSInvalid(
        'Flattened JWS must have either of the "protected" or "header" members'
      );
    }
    if (jws.protected !== undefined && typeof jws.protected !== "string") {
      throw new JWSInvalid("JWS Protected Header incorrect type");
    }
    if (jws.payload === undefined) {
      throw new JWSInvalid("JWS Payload missing");
    }
    if (typeof jws.signature !== "string") {
      throw new JWSInvalid("JWS Signature missing or incorrect type");
    }
    if (jws.header !== undefined && !isObject(jws.header)) {
      throw new JWSInvalid("JWS Unprotected Header incorrect type");
    }
    let parsedProt = {};
    if (jws.protected) {
      const protectedHeader = decode(jws.protected);
      try {
        parsedProt = JSON.parse(decoder.decode(protectedHeader));
      } catch (_b) {
        throw new JWSInvalid("JWS Protected Header is invalid");
      }
    }
    if (!isDisjoint(parsedProt, jws.header)) {
      throw new JWSInvalid(
        "JWS Protected and JWS Unprotected Header Parameter names must be disjoint"
      );
    }
    const joseHeader = {
      ...parsedProt,
      ...jws.header,
    };
    const extensions = validateCrit(
      JWSInvalid,
      new Map([["b64", true]]),
      options === null || options === void 0 ? void 0 : options.crit,
      parsedProt,
      joseHeader
    );
    let b64 = true;
    if (extensions.has("b64")) {
      b64 = parsedProt.b64;
      if (typeof b64 !== "boolean") {
        throw new JWSInvalid(
          'The "b64" (base64url-encode payload) Header Parameter must be a boolean'
        );
      }
    }
    const { alg } = joseHeader;
    if (typeof alg !== "string" || !alg) {
      throw new JWSInvalid(
        'JWS "alg" (Algorithm) Header Parameter missing or invalid'
      );
    }
    const algorithms =
      options && validateAlgorithms("algorithms", options.algorithms);
    if (algorithms && !algorithms.has(alg)) {
      throw new JOSEAlgNotAllowed(
        '"alg" (Algorithm) Header Parameter not allowed'
      );
    }
    if (b64) {
      if (typeof jws.payload !== "string") {
        throw new JWSInvalid("JWS Payload must be a string");
      }
    } else if (
      typeof jws.payload !== "string" &&
      !(jws.payload instanceof Uint8Array)
    ) {
      throw new JWSInvalid(
        "JWS Payload must be a string or an Uint8Array instance"
      );
    }
    let resolvedKey = false;
    if (typeof key === "function") {
      key = await key(parsedProt, jws);
      resolvedKey = true;
    }
    checkKeyType(alg, key, "verify");
    const data = concat(
      encoder.encode((_a = jws.protected) !== null && _a !== void 0 ? _a : ""),
      encoder.encode("."),
      typeof jws.payload === "string"
        ? encoder.encode(jws.payload)
        : jws.payload
    );
    const signature = decode(jws.signature);
    const verified = await verify(alg, key, signature, data);
    if (!verified) {
      throw new JWSSignatureVerificationFailed();
    }
    let payload;
    if (b64) {
      payload = decode(jws.payload);
    } else if (typeof jws.payload === "string") {
      payload = encoder.encode(jws.payload);
    } else {
      payload = jws.payload;
    }
    const result = { payload };
    if (jws.protected !== undefined) {
      result.protectedHeader = parsedProt;
    }
    if (jws.header !== undefined) {
      result.unprotectedHeader = jws.header;
    }
    if (resolvedKey) {
      return { ...result, key };
    }
    return result;
  }

  class SAOnboard extends Alertis {
    constructor(event) {
      super(event);
      this.request = event.request.body;
    }
    async onboardServiceArticle() {
      try {
        const fetchKeys = async () => {
          //const data = await fetch(`https://login.microsoftonline.com/ec1002d7-0348-40ae-ae4e-90c00506eacd/.well-known/openid-configuration`);
          const data = await fetch(
            `https://login.microsoftonline.com/24b080cd-5874-44ab-9862-8d7e0e0781ab/v2.0/.well-known/openid-configuration`
          );
          const jwksResp = await data.json();
          const keysData = await fetch(jwksResp.jwks_uri);
          return await keysData.json();
        };
        const isValidToken = async () => {
          let data = await fetchKeys();
          let k;
          if (data) {
            k = data.keys.reduce((acc, val) => {
              val.kid && (acc[val.kid] = val);
              return acc;
            });
          }
          let actualtoken =
            "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Imwzc1EtNTBjQ0g0eEJWWkxIVEd3blNSNzY4MCJ9.eyJhdWQiOiJlNjI0YTdjMi0zZTUzLTQ2NTktOGY5Yi1kN2MxOWZjZjAxZjciLCJpc3MiOiJodHRwczovL2xvZ2luLm1pY3Jvc29mdG9ubGluZS5jb20vMjRiMDgwY2QtNTg3NC00NGFiLTk4NjItOGQ3ZTBlMDc4MWFiL3YyLjAiLCJpYXQiOjE2MzkxMzU3NDYsIm5iZiI6MTYzOTEzNTc0NiwiZXhwIjoxNjM5MTM5NjQ2LCJuYW1lIjoiQXNoaXNoIFNoYXJtYSAoRGV2T24pIiwib2lkIjoiOTZmODM2N2QtY2M2NC00NjMwLWI0MGQtYTUwNTVjMjAwOGVkIiwicHJlZmVycmVkX3VzZXJuYW1lIjoiYXNoaXNoLnNoYXJtYUBkZXZvbi5ubCIsInJoIjoiMC5BUUlBellDd0pIUllxMFNZWW8xLURnZUJxOEtuSk9aVFBsbEdqNXZYd1pfUEFmY0NBTzguIiwic3ViIjoiLVNDRE5lR2IwVVc1TzZ5NkoxMERyNWhFZWxIR0lSdU5uNnd3NTZuMHRyMCIsInRpZCI6IjI0YjA4MGNkLTU4NzQtNDRhYi05ODYyLThkN2UwZTA3ODFhYiIsInV0aSI6InJfZERnTWdncGtPN01pQnhPNndTQUEiLCJ2ZXIiOiIyLjAifQ.pG8hmgZAFrtaqOIGgc2eYkZo_Xxb0_0ntgky5fZsAg5QpHoh7f6EfufNaTEcGY4oFMcE9ii5TI5S9-0LJqJcHRHpOd8xcWMQ-YWe5DAEag_90uRubFXDuLnK4mHGrEWRdlkO0vp5YDmhUItykyq_GVMzwBmbRKhRWzVxEao9dsXFZrnrTkQE2rdtE81w5kAvhEnYB8q7Yfy0uRN-7U2wLyazy_TqYfx19tDBN66F32rlV8SdTxvewRj4ZMw12_RBLdaiSoj6phMpjOllFgLdUi1RY-roJjNcdaHH988aopZgqnQvTQ8zOczES0tBqt-oRJyrwOhFvpvTZAgg5-ykUg";

          const splitToken = actualtoken.split(".");
          const head = JSON.parse(atob(splitToken[0]))
          console.log("__ head", JSON.stringify(head))
          const ecPublicKey = await importJWK(
            k[head.kid],
            head.alg
          ).catch(err=>{
            console.log("err in import jwk")
          });
          console.log("___got keys? ", JSON.stringify(ecPublicKey));
          const { payload, protectedHeader } = await flattenedVerify(
            {
              payload: splitToken[1],
              signature: splitToken[2],
              protected: splitToken[0],
            },
            ecPublicKey
          ).catch(err=>{
            console.log("err in flatten verify")
          });;
        };
        const tokenres = await isValidToken();
        if (!(tokenres && tokenres.verified)) {
          return { message: "invalid token. try again" };
        }
        return tokenres;
      } catch (error) {
        console.log("error ", JSON.stringify(error));
        return {
          message: "Something went wrong, Please try again.",
          error: error,
        };
      }
    }
  }

  class RequestHandler extends Alertis {
    constructor(event) {
      super(event);
      this.request = event.request.body;
    }
    async serve() {
      return new SAOnboard(this.event).onboardServiceArticle();
    }
  }

  async function epsilon(event) {
    try {
      const response = await new RequestHandler(event).serve();
      console.log(JSON.stringify(response));
      return new Response(JSON.stringify(response));
    } catch (e) {
      console.log(e);
      return new Response(JSON.stringify({ error: e }));
    }
  }

  // function atob(a: string) {
  const root = typeof globalThis !== "undefined" ? globalThis : window;
  root.epsilon = epsilon;

  exports.root = root;

  Object.defineProperty(exports, "__esModule", { value: true });

  return exports;
})({}, Zaza);
//# sourceMappingURL=alertis.epsilon.js.map

let res = epsilon({
  bag: { id: "f2da55eb9d498772f91f9ca35e81ba0f" },
  view: { qid: "660f26e61d34957c181955b12852054b", rights: 31 },
  project: {
    qid: "3b250f5c1686330bcdd0a1dbebe760ce",
    apiKey:
      "eyJhbGciOiJIUzI1NiIsImtpZCI6IjQwZjRiOWY2NTMxYzQyZmIyMzgyMmVhMDZlNzgyYWVlIiwidHlwIjoiSldUIn0.eyJWaWV3SWQiOiIiLCJSaWdodHMiOjAsIkJsZXNzZWQiOmZhbHNlLCJCYWdTdG9yYWdlUWlkIjoiIiwiUHJvamVjdElkIjoiM2IyNTBmNWMxNjg2MzMwYmNkZDBhMWRiZWJlNzYwY2UiLCJTY29wZSI6InByb2plY3QgdmlldyIsImlhdCI6MTYzODE5NDI4Nn0.eqaMhnoTGIrt-L8HIhSyLCp9HgI8fPgO6NHgtNfV5Sg",
  },
  request: { url: "", body: null },
});
Promise.resolve(res);
