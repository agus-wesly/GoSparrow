## Go Sparrow
GoSparrow is an unofficial tool to scrape data from various social media. 
It uses `cdp` (https://github.com/chromedp/chromedp) to run the browser

*PLEASE NOTE THAT THIS TOOLS IS JUST FOR EDUCATIONAL PURPOSE ONLY*

GoSparrow has 2 mode : *Single Mode* and *Search Mode*
Currently its support : Twitter and Tiktok
*More social media support will comming soon! :D*

### Prerequisite
In order to run this tools, you need to have : 
> Golang Installed. I'm using `go1.23.5 linux/amd64`
> Chrome installed

### Running
```
git clone git@github.com:agus-wesly/GoSparrow.git
```
or maybe just download the zip file

```
go build .

# Windows
GoSparrow.exe --headless=true //Set this to false if you want to turn off the headless mode
# Linux
./GoSparrow --headless=true //Set this to false if you want to turn off the headless mode
```

### Social Media Todo
- [x] Twitter
- [x] Tiktok
- [ ] Instagram Reels
- [ ] Threads
- [ ] Facebook
