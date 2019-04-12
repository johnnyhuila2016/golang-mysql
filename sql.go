package mysql

import (
	_"database/sql/driver"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"time"
	"fmt"
    "reflect"
    "regexp"
    _"strconv"
)
const (
    USERNAME = ""
    PASSWORD = ""
    NETWORK  = "tcp"
    SERVER   = "localhost"
    PORT     = 3306
    DATABASE = ""
    PREFIX   = ""
)

var (
    DB sql.DB

)
type sqlstruct struct{
    Where []map[string]string
    Join []string
    Order string
    Limit string
    Alias string
    Table string
    Group string
    Update string
    Save string
    Field string
    DB sql.DB
    Rows interface{}
    Row interface{}
}
var Dbstruct sqlstruct

func init () {
    Dbstruct.DB = *Connt();
}
func Connt () *sql.DB {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",USERNAME,PASSWORD,NETWORK,SERVER,PORT,DATABASE)
    db,err := sql.Open("mysql",dsn)
    if err != nil{
        fmt.Printf("Open mysql failed,err:%v\n",err)
    }
    db.SetConnMaxLifetime(100*time.Second)
    db.SetMaxOpenConns(100)
    db.SetMaxIdleConns(16)
    return db
}

func Table(table string){
    Dbstruct.Table = PREFIX+table
}
func Alias(alias string){
     Dbstruct.Alias = alias
}
func Where (wher map[string]string) {
    Dbstruct.Where = append(Dbstruct.Where,wher)

}
func Order(order string) {
    if order != ""{
        Dbstruct.Order = " order by " + order
    }
    
}
func GetOrder() string{
    return Dbstruct.Order
}
func Join (join string,r string){
    Dbstruct.Join = append(Dbstruct.Join,join+" "+r)
}
func Count () {

}
func Field(field string){
    if field != ""{
        Dbstruct.Field = field
    }
}
func GetField() string{
    return Dbstruct.Field
}
func SelectSql () string {
    query := ""
    query += "select " +GetField() + " " + " from " + Dbstruct.Table + Dbstruct.Alias 
    query += " "+Getjoin() + " " + Getwhere() + " " + GetOrder() + " " + GetGroup() + " " + GetLimit()
    return query
}
func Select (struc interface{}) sqlstruct {
    query := SelectSql()
    fmt.Println(query)
    rows,err := Dbstruct.DB.Query(query)
    defer func() {
        if rows != nil {
            rows.Close()
            Dbstruct.Where = Dbstruct.Where[0:0]
        }
    }()
    if err != nil {
        fmt.Println("Query failed,err:%v", err)
    }


    result := make([]interface{},0)
    s := reflect.ValueOf(struc).Elem()
    leng := s.NumField()

     // cols,_ := rows.Columns()
     vals := make([]interface {},leng)
     for i,_ := range vals{
        vals[i]=s.Field(i).Addr().Interface()
     }

    for rows.Next() {
        err = rows.Scan(vals...)
        if err != nil {
            panic(err)
        }
        result = append(result, s.Interface())
    }
    //fmt.Printf("%T",result)
    //return &result
    Dbstruct.Rows = result
    
    return Dbstruct

}
func Find () *sql.Row {
    query := SelectSql()
    row := Dbstruct.DB.QueryRow(query)
    defer func() {
        if row != nil {
            Dbstruct.DB.Close()
        }
    }()
    return row
}
func Update ( update map[string]string ) sql.Result{
    params := ""
    if update != nil{
        params := " set " 
        i := len(update)
        for key := range update {
            i--
            params += key + " = " + update[key]
            if len(update)>0 && i>0  {
                params += " , "
            }
        }
    }
    query := "update "+Dbstruct.Table + params
    relust,err := Dbstruct.DB.Exec(query)
    if err != nil{
        fmt.Println("Query failed,err:%v", err)
    }
    return relust
}

func Save (save map[string]string ) sql.Result {
    into,value := "",""
    if save != nil{
        into := " into ( "
        value := " )( "
        i := len(save)
        for key := range save {
            i--
            into += key
            value += save[key]
            if len(save)>0 && i>0  {
                into += " , "
                value += " , "
            }
        }
        into += " ) "
        value += " ) "
        
    }
    Dbstruct.Save = into + value
    query := "insert "+Dbstruct.Table + into + value
    result,err := Dbstruct.DB.Exec(query)
    if err != nil{
        fmt.Println("Query failed,err:%v", err)
    }
    return result
}
func GetSave() string{
    return Dbstruct.Save
}
func Del () {

}
func Limit(limit string) {
    if limit != ""{
        Dbstruct.Limit = " limit "+ limit +""
    }
}
func GetLimit() string{
    return Dbstruct.Limit
}
func Group(group string) {
    if group != ""{
        Dbstruct.Group = " group by "+group

    }
    
}
func GetGroup() string{
    return Dbstruct.Group
}

func Getwhere() string{
    str := Analyticwhere(Dbstruct)
    if str !=""{
        return " where " + str
    }else{
        return ""
    }
    
}
func Getjoin() string{
    join := ""
    if Dbstruct.Join != nil{
        params :=  Dbstruct.Join
        for key := range params{
            join += params[key]
        }
    }
    return join
}
func Analyticwhere(structName interface{}) string  {
    t := reflect.TypeOf(structName)
    v := reflect.ValueOf(structName)
    var data string
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    if t.Kind() != reflect.Struct {
        fmt.Println("Check type error not Struct")
    }
    fieldNum := t.NumField()
    //result := make([]string,0,fieldNum)
    for i:= 0;i<fieldNum;i++ {
        //result = append(result,t.Field(i).Name)
        if t.Field(i).Name == "Where" {
            val := v.Field(i).Interface()
                vl := val.([]map[string]string)
                where := ""
                for key := range vl {
                    i := len(vl[key])
                    for k := range vl[key]{
                        i--
                        reg := regexp.MustCompile(`!=|>|<|>=|<=|like|not between|between|in|not in`)
                        cdtion := reg.FindAllString(k, -1)
                        if len(cdtion)==0 {
                            where+=(k + " = "+ vl[key][k])
                        }else{
                            where+=(k +" " + vl[key][k])
                        }

                        if len(vl[key])>0 && i>0  {
                            where += " and "
                        }
                        
                    }
                    
                } 
                data =where
            }

    }
    return data
}