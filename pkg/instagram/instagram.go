package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Instagram struct {
	AuthToken string
	Log       *terminal.Log
}

func (ins *Instagram) Begin() {
	// Login and verify
	// Instagram
	// Need to login first
	// sessionid
	//		err := network.SetCookie("auth_token", t.AuthToken).
	//			WithDomain(".instagram.com").
	//			WithHTTPOnly(true).
	//			WithExpires(&expr).
	//			WithSecure(true).
	//			WithSameSite("strict").
	//			WithPath("/").
	//			Do(ctx)

	ctx, cancel := core.CreateNewContext()
	defer cancel()
	url := "https://www.instagram.com/p/DFMsjpuxdTC/"

	ins.Listen(ctx, "query")

	chromedp.Run(ctx,
		ins.Login(),
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(5*time.Second),
	)
}

func (ins *Instagram) Login() chromedp.Tasks {
	expr := cdp.TimeSinceEpoch(time.Now().Add(3 * time.Hour))
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := network.SetCookie("sessionid", "14441980786%3AeSO0KgPEDhOvdj%3A18%3AAYdhDo0Ft6e69jWGXuYM4fJPe69dJlt7FhRo1GREnA").
				WithDomain(".instagram.com").
				WithHTTPOnly(true).
				WithExpires(&expr).
				WithSecure(true).
				WithSameSite("None").
				WithPath("/").
				Do(ctx)

			if err != nil {
				return err
			}
			return nil
		}),
	}
}

func (ins *Instagram) Listen(ctx context.Context, eventKey string) error {
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
				go func() {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						panic(err)
					}
					var res interface{}
					err = json.Unmarshal(byts, &res)
					if err != nil {
						panic(err)
					}
					fmt.Println(res)
				}()
			}
		}
	})
	return nil
}
