package main

import( 
"database/sql"
_"database/sql/driver/mysql"
"net/http"
"net/smtp"
"log"
"fmt"
"time"
"net/url"
b64"encoding/base64"
"strings"
"cron"
)

var database *sql.DB
var err error
func main(){
    database,err= sql.Open("mysql","root:sqla@tcp(127.0.0.1:3306)/registration")
    if err!= nil{
        panic (err.Error())
    } 
    cron:= cron.New()
    cron.AddFunc("@daily",func(){currentdate := time.Now()
    fmt.Println("cron started")
    var databaseemail string;
    checkdate:= currentdate.AddDate(0,0,-7) 
    row,err:= database.Query("select email from registrationuser where active='"+checkdate.Format("2006-2-1")+"'")
        if err!= nil{
            log.Fatal(err)
        }else{
                for row.Next(){
                  row.Scan(&databaseemail)
                    email :=[]string{databaseemail}
                    send(email)
                }    
            }
      })
        cron.Start()
    http.HandleFunc("/",loginPage)
    http.Handle("/static/",http.StripPrefix("/static/",http.FileServer( http.Dir("static"))))
    http.HandleFunc("/Registration.html",RegistrationPage)
    http.HandleFunc("/activation_account",activation)   
    http.HandleFunc("/update.html",update)
    http.HandleFunc("/logout",logout)
    http.HandleFunc("/updateinsert", updateinsert)
    err= http.ListenAndServe(":8080",nil)
    if err!=nil{
        log.Fatal(err)    
    }
}

func RegistrationPage(res http.ResponseWriter, req *http.Request){
    if req.Method!= "POST"{
        http.ServeFile(res,req,"Registration.html")
        return
    }
    fmt.Println("Fetch data from form")
    firstname := req.FormValue("First_Name")
    lastname := req.FormValue("Last_Name")
    temp:=req.FormValue("Email")
    email :=[]string{temp}
     password :=req.FormValue("Password")
    confirmpassword :=req.FormValue("ConfirmPassword")
    dob :=req.FormValue("DOB")
    hashedPassword := b64.StdEncoding.EncodeToString([]byte(password))
    fmt.Println("encode the password")
    if(strings.Compare(password,confirmpassword)==0){
            fmt.Println("compare the password")
    _, err = database.Exec("INSERT INTO registrationuser (firstname,lastname,email,dob,password,active)VALUES(?,?,?,?,?,?)",
                           firstname,lastname,temp,dob,hashedPassword,time.Now().Local().Format("2006-2-1"))
    fmt.Println("insert into database")
     if err != nil{
         fmt.Println("<script>alert('user already exist'); window.location.href='/Registration';</script>")
         fmt.Fprintf(res,"<b>User already exist</b>")
            //fmt.Fprintf(res,err) error to display on user page for duplicate user
            http.Error(res,"Unable to insert data into DataBase",500) //500 doubts
            return
        }
        fmt.Println("user created and please activate your account")    // user creadted
        res.Write([]byte("<script>alert('Registration Successfull'); window.location.href='/login';</script>"))
        send(email) // calling send function to send email
        http.Redirect(res,req,"/login",301)

        }else{
        fmt.Fprintf(res,"<script>alert('password not match');</script>")
       // http.Redirect(res,req,"/Registration",301)
        http.ServeFile(res,req,"Registration.html")
    }
}
    func send(to []string){  // to send the activation link to a user
        currentdate:= time.Now().Local().Format("2006-2-1")
        encemail :=b64.StdEncoding.EncodeToString([]byte(to[0]))
        fmt.Println(encemail)
        activation_link :="http://localhost:8080/activation_account?id="+encemail;
        fmt.Println(activation_link)
    auth:=smtp.PlainAuth("","gofirsttime@gmail.com","qwerty@12345","smtp.gmail.com")
    msg:=[]byte("Welcome user Please activate your Account asap from the following link\n"+activation_link)
        fmt.Println(msg)
	err:=smtp.SendMail("smtp.gmail.com:587",auth,"gofirsttime@gmail.com",to,msg)
	if err!=nil{
		log.Fatal(err)
	}else{
    _, err = database.Exec("UPDATE registrationuser SET active='"+currentdate+"'WHERE email ='"+to[0]+"'")
		fmt.Println("Mail sent successfully.")
	}
    }
func loginPage(res http.ResponseWriter, req *http.Request){
    if req.Method!="POST"{
        fmt.Println("checked login module")
        http.ServeFile(res,req,"login.html")
    }else{
        email := req.FormValue("email")
        password:=req.FormValue("password")
        fmt.Println(email,password)
        var databasePassword string
        var databaseemail string
        var databaseactive string
        fmt.Println(email)
        row,err:=database.Query("SELECT email,password,active from registrationuser where email='"+email+"'")
        // scan from database
        if err!=nil{
            fmt.Printf("error occured in select query of login")
            log.Fatal(err)
            http.Redirect(res,req,"/login",301)
            return
        }
        if row.Next(){
            row.Scan(&databaseemail,&databasePassword,&databaseactive)
        }
        exp:=time.Now().Add(365*24*time.Hour)
        ck:=http.Cookie{Name:"user",Value:databaseemail,Expires:exp}
        http.SetCookie(res,&ck)
        fmt.Println(databaseactive)
        fmt.Println(databasePassword)
        hashedPassword,_:= b64.StdEncoding.DecodeString(databasePassword)
        //decode database encoded password 
        fmt.Println(string(hashedPassword))
        errorpass:= strings.Compare(password,string(hashedPassword)) 
        // compare database password with inserted password
        fmt.Println("Entered pwd: "+password+"\nOriginal pwd: "+string(hashedPassword))
        if errorpass != 0 {
            // if passowrd or email is incorrect
            fmt.Fprintf(res,"<script>alert('Incorrect email or password');</script>")    
            http.ServeFile(res,req,"login.html")
        }else if strings.Compare(databaseactive,"active")==0{
            // if user account is active       
            http.Redirect(res,req,"/update.html",301)
            fmt.Println("hello logger1")
        }else{
            // if user account is not activated
            fmt.Fprintf(res,"<script>alert('Please Activate your account first');</script>")
            fmt.Println("Please activate your account first.")
            http.Redirect(res,req,"/login",301)
        }
    }
}   
func update(res http.ResponseWriter, req *http.Request){   // for update the user information
    if req.Method!="POST"{
        var databaseemail string
        var databasefirstname string
        var databaselastname string
        var databasedob string
        cookie,_:=req.Cookie("user")
        fmt.Println("Database1",cookie.Value)
        row,err:=database.Query("SELECT firstname,lastname,email,dob from registrationuser where email='"+cookie.Value+"'") // get data from the database
        if err!=nil{
            fmt.Println("Database-exp")
            log.Fatal(err)            
        }else{
            row.Next()
            row.Scan(&databasefirstname,&databaselastname,&databaseemail,&databasedob)
            fmt.Println(databasefirstname)
            exp:= time.Now().Add(365*24*time.Hour)// for cookie timing 
            cookies:=http.Cookie{Name:"fname",Value:databasefirstname,Expires:exp}
            http.SetCookie(res,&cookies) // set cookies 
            cookies=http.Cookie{Name:"lname",Value:databaselastname,Expires:exp}
            http.SetCookie(res,&cookies)
            cookies=http.Cookie{Name:"email",Value:databaseemail,Expires:exp}
            http.SetCookie(res,&cookies)
            cookies=http.Cookie{Name:"dob",Value:databasedob,Expires:exp}
            http.SetCookie(res,&cookies)
            http.ServeFile(res,req,"update.html")    
        }
    }else{
        http.ServeFile(res,req,"update.html")
    }
}
    
func activation(res http.ResponseWriter, req *http.Request){  // to activate the accounr of a user
    geturl:=req.URL.RequestURI()
    urlenc,_:=url.Parse(geturl)
    mail,_:=url.ParseQuery(urlenc.RawQuery)
    encmail:=mail["id"][0]
    id,_:=b64.URLEncoding.DecodeString(encmail)
    _,err=database.Exec("update registrationuser set active='active' where email='"+string(id)+"'")
    // make change in the active field of database 
    if err!=nil{
        log.Fatal(err)      
    }else{
     fmt.Fprintf(res,"<script>alert('Account has been Activated');</script>")
       http.Redirect(res,req,"/login",301) 
    }
}
func logout(res http.ResponseWriter, req *http.Request){
    fmt.Print("logout func")
    if req.Method!="POST"{
        fmt.Fprintf(res,"error may occur")
    }else{
        expcookies:= & http.Cookie{Name: "fname",Value: " ",Expires: time.Now()}
        http.SetCookie(res,expcookies)
        expcookies= & http.Cookie{Name: "lname",Value: " ",Expires: time.Now()}
        http.SetCookie(res,expcookies)
        expcookies= & http.Cookie{Name: "email",Value: " ",Expires: time.Now()}
        http.SetCookie(res,expcookies)
        expcookies= & http.Cookie{Name: "dob",Value: " ",Expires: time.Now()}
        http.SetCookie(res,expcookies)
        expcookies= & http.Cookie{Name: "user",Value: " ",Expires: time.Now()}
        http.SetCookie(res,expcookies)
        http.Redirect(res,req,"/login",301)
    }
}

func updateinsert(res http.ResponseWriter , req *http.Request){
    fmt.Println("update insert")
    if req.Method!="POST"{
        fmt.Fprintf(res,"error may occur")
    }else{
       firstname := req.FormValue("First_Name")
       lastname := req.FormValue("Last_Name")
       email := req.FormValue("Email")
      // email :=[]string{temp}
       dob :=req.FormValue("DOB")
       fmt.Println(firstname)
        fmt.Println(string(email))
        _,err=database.Exec("update registrationuser set firstname='"+firstname+"',lastname='"+lastname+"',dob='"+dob+"' where email='" + string(email) + "'")
        fmt.Println("querry executed")
        fmt.Fprintf(res,"<script>alert('Profile Updated');</script>") 
        if err!= nil{
            log.Fatal(err)
        }else{
            http.ServeFile(res,req,"update.html")
        }
      }
}