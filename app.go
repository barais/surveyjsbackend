package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Error handling
	"errors"

	"github.com/thanhpk/randstr"

	// To read password from standard input without echoing
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	//"html/template"
	"net/url"

	"github.com/adelowo/filer"
	"github.com/adelowo/filer/validator"
	"github.com/barais/ipfilter"
	"gopkg.in/cas.v2"
	"gopkg.in/ldap.v3"
)

//TODO
// manage maven crash

var fileserver http.Handler
var login string
var pass string
var uploadfolder string

//var mavenhome string
//var templateProjectPath string
var smtpserver string
var amqpserver string
var ldapserver string
var sendemail bool
var binome bool
var ipfilterconfig string
var zipScheduleO []FileInterval

func main() {
	port := flag.String("p", "8080", "port to serve on")
	directory := flag.String("d", "./public", "the directory of static file to host")
	//dir = *directory
	casURL := flag.String("url", "https://sso-cas.univ-rennes1.fr", "the URL of your cas server")
	paramlogin := flag.String("login", "obarais", "login of smtp server")
	parampass := flag.String("pass", "", "pass of smtp server")
	uploadfolderparam := flag.String("u", "upload", "path of the folder to upload file")
	smtpserverparam := flag.String("smtpserver", "smtps.univ-rennes1.fr:587", "smtp server to use")
	ldapserverparam := flag.String("ldapserver", "ldap.univ-rennes1.fr:389", "ldap server to use")
	sendEmailparam := flag.Bool("sendemail", true, "Send an email")
	binomeparam := flag.Bool("binome", false, "Send Binome Name within the post message")
	ipfilterconfigparam := flag.String("ipfilterconfig", "ipfilter.json", "json file to configure ipfilter")

	flag.Parse()
	fileserver = http.FileServer(http.Dir(*directory))
	login = *paramlogin

	uploadfolder = *uploadfolderparam
	//	mavenhome = *mavenhomeparam
	//	templateProjectPath = *templateProjectPathparam
	smtpserver = *smtpserverparam

	ldapserver = *ldapserverparam
	sendemail = *sendEmailparam
	binome = *binomeparam

	// Password initialisation for smtp
	pass = *parampass
	if sendemail {
		if pass == "" { // Reading password from standard input
			fmt.Println("Enter your smtp password: ")
			password, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Println("Error while reading your smtp password")
				fmt.Println(err)
			}
			pass = string(password)
			if pass == "" {
				fmt.Println("Error while reading your smtp password. Empty password.")
			}
		}
	}

	ipfilterconfig = *ipfilterconfigparam
	mux := http.NewServeMux()

	ipfilterconfigfile, err := os.Open(ipfilterconfig)
	if err != nil {
		log.Fatal(err)
	}
	defer ipfilterconfigfile.Close()
	allowedSchedule, err := ioutil.ReadAll(ipfilterconfigfile)
	var allowedScheduleO []*ipfilter.IPInterval
	json.Unmarshal([]byte(allowedSchedule), &allowedScheduleO)

	mux.HandleFunc("/", testCas)
	url, _ := url.Parse(*casURL)

	client := cas.NewClient(&cas.Options{
		URL: url,
	})

	options := ipfilter.NewOption(allowedScheduleO)

	myProtectedHandler := ipfilter.Wrap(client.Handle(mux), *options)

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+*port, myProtectedHandler))
}

type templateBinding struct {
	Username   string
	Attributes cas.UserAttributes
}

func getJSON(w http.ResponseWriter, r *http.Request, binding *templateBinding) {

	//			w.Header().Add("Content-Disposition", "Attachment")
	w.Header().Set("Content-Type", "application/json")

	jsonfile, err := os.Open("exam.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonfile.Close()
	jsonfilebyt, err := ioutil.ReadAll(jsonfile)
	w.Write(jsonfilebyt)
	return

}

func testCas(w http.ResponseWriter, r *http.Request) {
	if !cas.IsAuthenticated(r) {
		cas.RedirectToLogin(w, r)
		return
	}
	if r.URL.Path == "/logout" {
		cas.RedirectToLogout(w, r)
		return
	}
	log.Println(cas.Username(r))
	binding := &templateBinding{
		Username:   cas.Username(r),
		Attributes: cas.Attributes(r),
	}
	if r.URL.Path == "/upload" {
		uploadProgress(w, r, binding)
		return
	}
	if r.URL.Path == "/casusername" {
		casUsername(w, r, binding)
		return
	}
	if r.URL.Path == "/exam.json" {
		getJSON(w, r, binding)
		return
	}

	fileserver.ServeHTTP(w, r)
}

type FileInterval struct {
	Lower       *time.Time
	Upper       *time.Time
	fileToServe string
}

func (l *FileInterval) UnmarshalJSON(j []byte) error {
	var rawStrings map[string]string

	err := json.Unmarshal(j, &rawStrings)
	if err != nil {
		return err
	}

	for k, v := range rawStrings {
		if strings.ToLower(k) == "lower" {
			t, err := time.Parse(time.RFC822, v)
			if err != nil {
				return err
			}
			l.Lower = &t

		}
		if strings.ToLower(k) == "upper" {
			t, err := time.Parse(time.RFC822, v)
			if err != nil {
				return err
			}
			l.Upper = &t

		}
		if strings.ToLower(k) == "file" {
			l.fileToServe = v
		}
	}

	return nil
}

func casUsername(w http.ResponseWriter, r *http.Request, binding *templateBinding) {

	//w.WriteHeader(200)
	fmt.Fprint(w, ""+binding.Username+"\n", http.StatusOK)
	return

}

func uploadProgress(w http.ResponseWriter, r *http.Request, binding *templateBinding) {
	mr, err := r.MultipartReader()
	if err != nil {
		fmt.Fprint(w, "Error on server, could not upload, please contact your admin\n")
		return
	}
	//    length := r.ContentLength
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		token := randstr.String(24) // generate a random 16 character length string
		timestamp := time.Now().Format("20060102150405")
		//		var p float32

		bindingname := strings.Replace(binding.Username, " ", "", -1)

		path := uploadfolder + "/" + bindingname + "_" + timestamp + "_" + token + ".json"
		dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("Unable to create file " + path)
			log.Printf("Local directory " + uploadfolder + " probably unexisting.")
			log.Printf("Please create one, and re-start server with option -u set to the created path.")
			fmt.Fprint(w, "Error on server, please contact your admin\n")
			return
		}
		var read int64

		for {
			buffer := make([]byte, 1000000)
			cBytes, err := part.Read(buffer)
			read = read + int64(cBytes)
			//fmt.Printf("read: %v \n",read )
			//p = float32(read) / float32(length) *100
			//fmt.Printf("progress: %v \n",p )
			dst.Write(buffer[0:cBytes])
			if err == io.EOF {
				break
			}
		}
		dst.Close()

		max, _ := filer.LengthInBytes("20MB")
		min, _ := filer.LengthInBytes("0KB")
		var val2 = validator.NewSizeValidator(max, min)
		var file, _ = os.Open(path)
		var val3 = validator.NewExtensionValidator([]string{"json"})
		// var val2 = validator.NewSizeValidator((1024 * 1024 * 10), (1 * 1)) //10MB(maxSize) and 1B(minSize)

		var errg error
		if _, err := val2.Validate(file); err != nil {
			errg = err
		}
		if _, err := val3.Validate(file); err != nil {
			errg = err
		}
		if errg != nil {
			log.Printf("Validation failed")
			file.Close()
			os.Remove(path)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprint(w, "Error. Your file must be a json file\n")
			return
		}

		if binome {
			log.Println("ok pour binome")
			part1, _ := mr.NextPart()
			buffer1 := make([]byte, 1000000)
			cBytes1, _ := part1.Read(buffer1)
			binomename := fmt.Sprintf("%s", buffer1[0:cBytes1])
			binomefinal := strings.Replace(binomename, " ", "", -1)
			binomefinal = strings.Replace(binomefinal, "ç", "c", -1)
			binomefinal = strings.Replace(binomefinal, "é", "e", -1)
			binomefinal = strings.Replace(binomefinal, "è", "e", -1)
			binomefinal = strings.Replace(binomefinal, "ê", "e", -1)
			binomefinal = strings.Replace(binomefinal, "ô", "o", -1)
			binomefinal = strings.Replace(binomefinal, "î", "i", -1)
			binomefinal = strings.Replace(binomefinal, "à", "a", -1)
			binomefinal = strings.Replace(binomefinal, "ù", "u", -1)
			bindingname = bindingname + "_" + binomefinal
			//			log.Println(binomefinal)
			os.Rename(path, uploadfolder+"/"+bindingname+"_"+timestamp+"_"+token+".json")
		}

		// Project info and report
		projName := filepath.Base(filepath.Dir(path))
		log.Printf("Current reference name for project : %s\n ", projName)

		preambule := "Vous venez de déposer un exam sur l'interface en ligne de " + fmt.Sprintf("%s", projName) + ".\n\n"

		postambule := "\n" +
			"Le nom du fichier sur le serveur est " + filepath.Base(path) + ". \n\n"

		mailpostambule := "Gardez trace de cet email en cas de litige.\n\n" +
			"Ceci est un mail automatique, merci de ne pas y répondre.\n\n" +
			"Sincèrement,\n L'équipe pédagogique de l'ISTIC"

		// Project info and report
		report := ""
		mailaddr := ""
		if sendemail {
			mailaddr, err = getMail(binding.Username)

			if (err != nil) || (!sendEmail("Bonjour "+binding.Username+",\n\n"+preambule+report+postambule+mailpostambule,
				"Rendu TP "+fmt.Sprintf("%s", projName), mailaddr)) {
				fmt.Fprintf(w, "Error. cannot send the email<BR>")
				return
			}
		}
		/*
			1 : Cannot createTmpFile
			2 : Cannot copy template
			3 : Cannot copy unzip src to template copy
			4 : Cannot copy unzip resources to template copy
			5 : Cannot copy list all jar files
			6 : Cannot copy all jar files
			7 : Cannot generate maven pom.xml
			8 : Cannot execute maven
			9 : Cannot load surfire reports
			10 : Cannot load scalastype report
			11 : Cannot send an email
		*/
		fmt.Fprintf(w, preambule+report+postambule)
		if mailaddr == "" {
			mailaddr = "barais@irisa.fr"
		}
		return

	}
}

func sendEmail(body string, subj string, tos string) bool {
	from := mail.Address{"Responsable L1 SI2", "resp-l1-si2@univ-rennes1.fr"}
	to := mail.Address{"", tos}
	log.Println(to)
	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	host, _, _ := net.SplitHostPort(smtpserver)

	auth := smtp.PlainAuth("", login, pass, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	c, err := smtp.Dial(smtpserver)
	if err != nil {
		log.Printf("could not connect to smtp server " + smtpserver)
		return false
	}

	c.StartTLS(tlsconfig)

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Printf("Could not login to SMTP server")
		return false
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		log.Printf("Bad from address")
		return false
	}

	if err = c.Rcpt(to.Address); err != nil {
		log.Printf("Bad to address")
		return false
	}

	// Data
	w, err := c.Data()
	if err != nil {
		//log.Panic(err)
		return false
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		//log.Panic(err)
		return false
	}

	err = w.Close()
	if err != nil {
		//log.Panic(err)
		return false
	}

	c.Quit()
	return true
}

func getMail(uid string) (string, error) {
	deft := "barais@irisa.fr"

	l, err := ldap.Dial("tcp", ldapserver) //fmt.Sprintf("%s:%d", "ldap.univ-rennes1.fr", 389))
	if err != nil {
		return deft, err
	}
	defer l.Close()
	// Reconnect with TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return deft, err
	}

	//15010426
	searchRequest := ldap.NewSearchRequest(
		"ou=People,dc=univ-rennes1,dc=fr", // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(uid="+uid+"))", // The filter to apply
		[]string{"mail"},   // A list attributes to retrieve
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return deft, err
	}

	for _, entry := range sr.Entries {
		return entry.GetAttributeValue("mail"), nil
		//fmt.Printf("%s: %v\n", entry.DN, entry.GetAttributeValue("mail"))
	}
	//	fmt.Println("ok")
	return deft, errors.New("getMail : couldn't find student email address")
}
