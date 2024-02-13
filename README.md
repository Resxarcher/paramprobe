# ParamProbe
a regex-base tool written in golang for grab parameters from url's.
# Features
- Fast.
- Extract attrebuites like name and id.
- Extract JSON Object's
- Extract Parameters from anchor tags.
- unique and sort.
# Install
**binary file**

``go install github.com/resxarcher/paramprobe@latest``

**source code**

``go get -u github.com/resxarcher/paramprobe``
# Usage
```
Usage: paramprobe [options]
OPTIONS:
  -delay int
    	durations between each HTTP requests (default -1)
  -domain string
    	single domain to process
  -header value
    	custom header/cookie to include in requests like that: Cookie:value, Origin:value, Host:value
  -lists string
    	file list to process
  -output string
    	path to save output
  -timeout int
    	timeout in seconds (default 10)
```
# Enjoy it ❤️
**if you found any issues, tweet it or open a new issues :)**

<a href="https://www.buymeacoffee.com/resxarcher" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-violet.png" alt="Buy Me A Coffee" style="height: 22px !important;width: 100px !important;" ></a> 

[![Twitter Follow](https://img.shields.io/twitter/follow/resxarcher?style=social)](https://twitter.com/resxarcher)
