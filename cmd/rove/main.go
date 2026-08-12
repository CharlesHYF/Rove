/*
 * 文件作用：Rove CLI 入口 -- 执行根命令，错误以 machine-readable 形式输出到 stderr。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package main 是 Rove CLI 的入口。
package main

import (
	"fmt"
	"os"
)

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
