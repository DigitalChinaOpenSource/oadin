//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package utils

import (
	"container/list"
	"sync"
)

type SafeList struct {
	mut  sync.Mutex
	list *list.List
}

func NewSafeList() *SafeList {
	sl := new(SafeList)
	sl.list = list.New()

	return sl
}

func (sl *SafeList) Back() *list.Element {
	return sl.list.Back()
}

func (sl *SafeList) Front() *list.Element {
	return sl.list.Front()
}

func (sl *SafeList) InsertAfter(v any, mark *list.Element) *list.Element {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	return sl.list.InsertAfter(v, mark)
}

func (sl *SafeList) InsertBefore(v any, mark *list.Element) *list.Element {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	return sl.list.InsertBefore(v, mark)
}

func (l *SafeList) Len() int {
	return l.list.Len()
}

func (sl *SafeList) MoveAfter(e, mark *list.Element) {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	sl.list.MoveAfter(e, mark)
}

func (sl *SafeList) MoveBefore(e, mark *list.Element) {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	sl.list.MoveBefore(e, mark)
}

func (sl *SafeList) Remove(e *list.Element) any {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	return sl.list.Remove(e)
}

func (sl *SafeList) PushBack(v any) *list.Element {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	return sl.list.PushBack(v)
}

func (sl *SafeList) PushFront(v any) *list.Element {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	return sl.list.PushFront(v)
}

func (sl *SafeList) MoveToBack(e *list.Element) {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	sl.list.MoveToBack(e)
}

func (sl *SafeList) MoveToFront(e *list.Element) {
	sl.mut.Lock()
	defer sl.mut.Unlock()

	sl.list.MoveToFront(e)
}
