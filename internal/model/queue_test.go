// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package model

import (
	"fmt"
	"reflect"
	"testing"
)

func TestJobQueue(t *testing.T) {
	// Some random actions
	q := newJobQueue()
	q.Push("f1")
	q.Push("f2")
	q.Push("f3")
	q.Push("f4")

	progress, queued := q.Jobs()
	if len(progress) != 0 || len(queued) != 4 {
		t.Fatal("Wrong length")
	}

	for i := 1; i < 5; i++ {
		n, ok := q.Pop()
		if !ok || n != fmt.Sprintf("f%d", i) {
			t.Fatal("Wrong element")
		}
		progress, queued = q.Jobs()
		if len(progress) != 1 || len(queued) != 3 {
			t.Log(progress)
			t.Log(queued)
			t.Fatal("Wrong length")
		}

		q.Done(n)
		progress, queued = q.Jobs()
		if len(progress) != 0 || len(queued) != 3 {
			t.Fatal("Wrong length", len(progress), len(queued))
		}

		q.Push(n)
		progress, queued = q.Jobs()
		if len(progress) != 0 || len(queued) != 4 {
			t.Fatal("Wrong length")
		}

		q.Done("f5") // Does not exist
		progress, queued = q.Jobs()
		if len(progress) != 0 || len(queued) != 4 {
			t.Fatal("Wrong length")
		}
	}

	if len(q.progress) > 0 || len(q.queued) != 4 {
		t.Fatal("Wrong length")
	}

	for i := 4; i > 0; i-- {
		progress, queued = q.Jobs()
		if len(progress) != 4-i || len(queued) != i {
			t.Fatal("Wrong length")
		}

		s := fmt.Sprintf("f%d", i)

		q.BringToFront(s)
		progress, queued = q.Jobs()
		if len(progress) != 4-i || len(queued) != i {
			t.Fatal("Wrong length")
		}

		n, ok := q.Pop()
		if !ok || n != s {
			t.Fatal("Wrong element")
		}
		progress, queued = q.Jobs()
		if len(progress) != 5-i || len(queued) != i-1 {
			t.Fatal("Wrong length")
		}

		q.Done("f5") // Does not exist
		progress, queued = q.Jobs()
		if len(progress) != 5-i || len(queued) != i-1 {
			t.Fatal("Wrong length")
		}
	}

	_, ok := q.Pop()
	if len(q.progress) != 4 || ok {
		t.Fatal("Wrong length")
	}

	q.Done("f1")
	q.Done("f2")
	q.Done("f3")
	q.Done("f4")
	q.Done("f5") // Does not exist

	_, ok = q.Pop()
	if len(q.progress) != 0 || ok {
		t.Fatal("Wrong length")
	}

	progress, queued = q.Jobs()
	if len(progress) != 0 || len(queued) != 0 {
		t.Fatal("Wrong length")
	}
	q.BringToFront("")
	q.Done("f5") // Does not exist
	progress, queued = q.Jobs()
	if len(progress) != 0 || len(queued) != 0 {
		t.Fatal("Wrong length")
	}
}

func TestBringToFront(t *testing.T) {
	q := newJobQueue()
	q.Push("f1")
	q.Push("f2")
	q.Push("f3")
	q.Push("f4")

	_, queued := q.Jobs()
	if !reflect.DeepEqual(queued, []string{"f1", "f2", "f3", "f4"}) {
		t.Errorf("Incorrect order %v at start", queued)
	}

	q.BringToFront("f1") // corner case: does nothing

	_, queued = q.Jobs()
	if !reflect.DeepEqual(queued, []string{"f1", "f2", "f3", "f4"}) {
		t.Errorf("Incorrect order %v", queued)
	}

	q.BringToFront("f3")

	_, queued = q.Jobs()
	if !reflect.DeepEqual(queued, []string{"f3", "f1", "f2", "f4"}) {
		t.Errorf("Incorrect order %v", queued)
	}

	q.BringToFront("f2")

	_, queued = q.Jobs()
	if !reflect.DeepEqual(queued, []string{"f2", "f3", "f1", "f4"}) {
		t.Errorf("Incorrect order %v", queued)
	}

	q.BringToFront("f4") // corner case: last element

	_, queued = q.Jobs()
	if !reflect.DeepEqual(queued, []string{"f4", "f2", "f3", "f1"}) {
		t.Errorf("Incorrect order %v", queued)
	}
}

func BenchmarkJobQueueBump(b *testing.B) {
	files := genFiles(b.N)

	q := newJobQueue()
	for _, f := range files {
		q.Push(f.Name)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.BringToFront(files[i].Name)
	}
}

func BenchmarkJobQueuePushPopDone10k(b *testing.B) {
	files := genFiles(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := newJobQueue()
		for _, f := range files {
			q.Push(f.Name)
		}
		for _ = range files {
			n, _ := q.Pop()
			q.Done(n)
		}
	}

}
