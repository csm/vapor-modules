package main

import (
    "github.com/csm/vapor-modules"
    "github.com/csm/go-edn/types"
    "bufio"
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/sha512"
    "errors"
    "fmt"
    "hash"
    "io"
    "os"
    "strings"
)

type HashFile struct{}

func (this HashFile) TakesInput() bool {
    return true
}

func (this HashFile) Exec(input types.Value) (out types.Value, err error) {
    defer func() {
        if r := recover(); r != nil {
            out = types.Map{
                types.Keyword("success"): types.Bool(false),
                types.Keyword("error"): types.String(fmt.Sprint(r)),
            }
        }
    }()
    var inputMap types.Map
    if i, ok := input.(types.Map); ok {
        inputMap = i
    } else {
        err = errors.New(fmt.Sprintf("expected a map as input, got a %T", input))
        panic(err)
    }

    var hashName string
    var hashNameElement = inputMap[types.Keyword("hash")]
    if hashNameElement == nil {
        hashName = "sha256"
    } else if h, ok := hashNameElement.(types.String); ok {
        hashName = string(h)
    } else if k, ok := hashNameElement.(types.Keyword); ok {
        hashName = string(k)
    } else if s, ok := hashNameElement.(types.Symbol); ok {
        hashName = string(s)
    } else {
        err = errors.New(":hash must be a string, keyword, or symbol if specified")
        panic(err)
    }
    hashName = strings.ToLower(hashName)

    var hash hash.Hash
    if (hashName == "md5") {
        hash = md5.New()
    } else if (hashName == "sha" || hashName == "sha1" || hashName == "sha-1") {
        hash = sha1.New()
    } else if (hashName == "sha224" || hashName == "sha-224") {
        hash = sha256.New224()
    } else if (hashName == "sha256" || hashName == "sha-256") {
        hash = sha256.New()
    } else if (hashName == "sha384" || hashName == "sha-384") {
        hash = sha512.New384()
    } else if (hashName == "sha512/224" || hashName == "sha-512/224") {
        hash = sha512.New512_224()
    } else if (hashName == "sha512/256" || hashName == "sha-512/256") {
        hash = sha512.New512_256()
    } else if (hashName == "sha512" || hashName == "sha-512") {
        hash = sha512.New()
    } else {
        err = errors.New(fmt.Sprint("unknown hash name: ", hashName))
        return
    }

    var fileName string
    var fileElem = inputMap[types.Keyword("file")]
    if fileElem == nil {
        err = errors.New(":file argument is required")
        panic(err)
    } else if f, ok := fileElem.(types.String); ok {
        fileName = string(f)
    } else {
        err = errors.New(":file argument must be a string")
        panic(err)
    }

    file, err := os.Open(fileName)
    if err != nil {
        panic(err)
    }
    hashOut := bufio.NewWriterSize(hash, hash.BlockSize())
    wrote, err := io.Copy(hashOut, file)
    hashOut.Flush()
    if err != nil {
        panic(err)
    }
    out = types.Map{
        types.Keyword("success"): types.Bool(true),
        types.Keyword("size"): types.Int(wrote),
        types.Keyword("file"): types.String(fileName),
        types.Keyword("hash"): types.Keyword(hashName),
        types.Keyword("digest"): types.String(fmt.Sprintf("%x", hash.Sum(nil))),
    }
    return
}

func main() {
    vapor.RunModule(HashFile{})
}
