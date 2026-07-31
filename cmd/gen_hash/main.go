package main

import (
    "fmt"
    "www.abc-ai.cn/FastToken/common"
)

func main() {
    hash, err := common.Password2Hash("FastToken2026!")
    if err != nil {
        fmt.Println("ERROR:", err)
        return
    }
    fmt.Printf("%s", hash)
}
