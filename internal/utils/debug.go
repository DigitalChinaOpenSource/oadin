// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import "github.com/k0kubun/pp/v3"

// PPprint Colorful pretty printers to help quick debug
func PPprint(a ...interface{}) (n int, err error) {
	return pp.Print(a...)
}

func PPprintf(format string, a ...interface{}) (n int, err error) {
	return pp.Printf(format, a...)
}

func PPprintln(a ...interface{}) (n int, err error) {
	return pp.Println(a...)
}

func PPsprint(a ...interface{}) string {
	return pp.Sprint(a...)
}

func PPsprintf(format string, a ...interface{}) string {
	return pp.Sprintf(format, a...)
}

func PPsprintln(a ...interface{}) string {
	return pp.Sprintln(a...)
}
