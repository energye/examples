//go:build !windows

package main

import "github.com/energye/lcl/types"

// Non-Windows builds never run the input diagnostics; keep the symbols so the
// shared main.go / scene.go compile unchanged.
var (
	inputDiagMode bool
	inputTestMode bool
)

func initInputDiag()                                    {}
func diagWndProc(msg *types.TLMessage)                  {}
func diagWndProcAfter(msg *types.TLMessage)             {}
func diagDispatch(e *UIEvent, s *Scene, hit Widget)     {}
func diagDispatchResult(e *UIEvent, s *Scene)           {}
func diagFrame(s *Scene)                                {}
func diagState(s *Scene, tag string)                    {}
func startInputTest(s *Scene)                           {}
