# Project overview

This project provides an infrastructure for saving folder quizz based on [surveyJS](https://surveyjs.io/Documentation/Library). The code of the front is [there](https://github.com/barais/examSurveyJSFront).

This project is developed using [GoLang](https://golang.org/).

It consists of a [library](https://github.com/barais/ipfilter/) to open the web application based on a set of IP addresses and time range, a [module](https://github.com/barais/surveyjsbackend/) to provide an http service to provide an exam, use CAS for the student authentification and save the result in json. 


## Surveyjsbackend

The main features of the surveyjsbackend module are the following:

- Filter By IP And Date(Open app only for specific lab room and lab schedule)
- Student authentification with CAS
- Custom Html Page (I provide an example with [SurveyJS](https://github.com/barais/examSurveyJSFront))
- Send Email to Student


# Installation

## Surveyjsbackend

```bash
git get github.com/barais/surveyjsbackend
cd $GOPATH/src/github.com/barais/surveyjsbackend
go build
./surveyjsbackend
```

**Typical usage**

```bash
./surveyjsbackend -p 9511 -login "YOURSMTPLOGIN"  -pass "YOURSMTPPASS" -ipfilterconfig ipfilter.json -d public/  -u upload/ -sendemail=true
```

**Parameter**

```txt
Usage of ./surveyjsbackend:
  -alsologtostderr
    	log to standard error as well as files
  -binome
    	Send Binome Name within the post message
  -d string
    	the directory of static file to host (default "./public")
  -ipfilterconfig string
    	json file to configure ipfilter (default "ipfilter.json")
  -ldapserver string
    	ldap server to use (default "ldap.univ-rennes1.fr:389")
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -login string
    	login of smtp server (default "obarais")
  -logtostderr
    	log to standard error instead of files
  -p string
    	port to serve on (default "8080")
  -pass string
    	pass of smtp server
  -sendemail
    	Send an email (default true)
  -smtpserver string
    	smtp server to use (default "smtps.univ-rennes1.fr:587")
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -u string
    	path of the folder to upload file (default "upload")
  -url string
    	the URL of your cas server (default "https://sso-cas.univ-rennes1.fr")
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging

```

**IPFilter config file**

*ipfilter.json* support the following syntax. It is a list of period associated with a list of allowed IPs. 

```js
[
{
    "lower": "01 Jan 18 08:00 CET",
    "upper": "31 Dec 19 18:00 CET",
    "allowedips": "*"
},
{
    "lower": "10 Jan 19 16:15 CET",
    "upper": "10 Jan 19 18:20 CET",
    "allowedips": "148.60.1.65-148.60.1.74"
},
{
    "lower": "10 Jan 19 16:15 CET",
    "upper": "10 Jan 19 18:20 CET",
    "allowedips": "148.60.1.4"
}
]
```



## Supervisor

```bash
sudo apt-get install supervisor
```

Edit config file

```bash
nano -w /etc/supervisor/conf.d/surveyjsbackend.conf
```

Sample of supervisor (folder name must be adapted to your installation)

```txt
[program:exam]
directory=/opt/surveyjsbackend
command=/opt/surveyjsbackend/surveyjsbackend -p 9511 -login "YOURSMTPLOGIN"  -pass "YOURSMTPPASS" -ipfilterconfig /opt/surveyjsbackend/ipfilter.json -d /opt/surveyjsbackend/public  -u /opt/surveyjsbackend/upload -sendemail=true
autostart=true
autorestart=true
redirect_stderr=true
```



## nginx

sample of nginx proxy. Change APP.yourappdomainname.fr with your app domain. Use [letsencrypt](https://letsencrypt.org/) if you need https endpoint (required if you want to take photos from the front). 


```bash
server {
        listen   80;
        # Must be changed, must be a subdomain of your university to use the cas
        server_name     APP.yourappdomainname.fr;
	 location / {
		proxy_connect_timeout 159s;
                proxy_send_timeout   600;
                proxy_read_timeout   600;
                proxy_http_version 1.1;
                proxy_pass         http://127.0.0.1:9501/;
		proxy_set_header Host APP.yourappdomainname.fr;
    	proxy_set_header X-Host APP.yourappdomainname.fr;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
#		proxy_set_header X-Forwarded-Proto $proxy_x_forwarded_proto;
		proxy_buffering off;
        }
    error_page 502 /HTTP502.html;
    location /HTTP502.html {
        root /etc/nginx/errorpages/HttpErrorPages/dist/;
    }

}
```


An example of HTTP502 is available [here](https://github.com/barais/surveyjsbackend/tree/master/nginxerrorpage)



