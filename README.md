## Go Sparrow
![image](https://github.com/user-attachments/assets/87b59785-3357-42f9-be37-72852f20e3f3)

GoSparrow is an unofficial tool to scrape data from various social media.<br>
It uses `cdp` (https://github.com/chromedp/chromedp) to run the browser.<br>
Its free and open source :D<br> 


*PLEASE NOTE THAT THIS TOOL IS JUST FOR EDUCATIONAL AND RESEARCH PURPOSE ONLY*

GoSparrow has 2 mode : *Single Mode* and *Search Mode*<br>
Currently its support : Twitter and Tiktok<br>
*More social media support will comming soon!*<br>

### Prerequisite
In order to run this tool, you need to have : 
- Golang Installed. I'm using `go1.23.5 linux/amd64`
- Chrome installed

### Running
```
git clone git@github.com:agus-wesly/GoSparrow.git
```
or maybe just download the zip file (https://github.com/agus-wesly/GoSparrow/archive/refs/heads/main.zip)

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
