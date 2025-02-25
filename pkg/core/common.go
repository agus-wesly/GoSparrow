package core

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var IS_HEADLESS = false

func CreateNewContext() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", IS_HEADLESS),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(actx)
	return ctx, acancel
}

func CreateNewContextWithTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", IS_HEADLESS),
	)
	_ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	_ctx, _ = chromedp.NewContext(_ctx)
	ctx, cancel := context.WithTimeout(_ctx, duration)
	return ctx, cancel
}

type ExecFn func(byts []byte)

func ListenEvent(ctx context.Context, eventKey string, exec ExecFn, wg *sync.WaitGroup) error {
	requestIdList := make([]network.RequestID, 0)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, eventKey) {
				requestIdList = append(requestIdList, responseReceivedEvent.RequestID)
			}
		case *network.EventLoadingFinished:
			if !slices.Contains(requestIdList, responseReceivedEvent.RequestID) {
				break
			} else {
				requestIdList = slices.DeleteFunc(requestIdList, func(targetId network.RequestID) bool {
					return targetId == responseReceivedEvent.RequestID
				})
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
                if wg != nil {
                    wg.Add(1)
                }
				go func() {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						fmt.Println("No resource error")
					}
					exec(byts)
				}()
			}
		}
	})
	return nil
}
