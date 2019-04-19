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
    PREFIX   = "eb_"
)

var (
    DB sql.DB

)
type Sqlstruct struct{
    where []map[string]string
    join []string
    order string
    limit string
    alias string
    table string
    group string
    update string
    save string
    field string
    DB sql.DB
    Rows interface{}
    Row interface{}
}
var Dbstruct Sqlstruct

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

func(sqlstruct *Sqlstruct) Table(table string) Sqlstruct {
    Dbstruct.table = PREFIX+table
    return Dbstruct
}
func(sqlstruct *Sqlstruct) Alias(alias string) Sqlstruct {
     Dbstruct.alias = alias
     return Dbstruct
}
func(sqlstruct *Sqlstruct) Where (wher map[string]string) Sqlstruct {
    Dbstruct.where = append(Dbstruct.where,wher)
    return Dbstruct

}
func(sqlstruct *Sqlstruct) Order(order string) Sqlstruct {
    if order != ""{
        Dbstruct.order = " order by " + order
    }
    return Dbstruct
    
}
func GetOrder() string{
    return Dbstruct.order
}
func(sqlstruct *Sqlstruct) Join (join string,r string) Sqlstruct {
    Dbstruct.join = append(Dbstruct.join,join+" "+r)
    return Dbstruct
}
func(sqlstruct *Sqlstruct) Count () {

}
func(sqlstruct *Sqlstruct) Field(field string) Sqlstruct {
    if field != ""{
        Dbstruct.field = field
    }
    return Dbstruct
}
func GetField() string{
    return Dbstruct.field
}
func SelectSql () string {
    query := ""
    query += "select " +GetField() + " " + " from " + Dbstruct.table + " "+ Dbstruct.alias 
    query += " "+Getjoin() + " " + Getwhere() + " " + GetOrder() + " " + GetGroup() + " " + GetLimit()
    return query
}
func(sqlstruct *Sqlstruct) Select (struc interface{}) interface{} {
    query := SelectSql()
    fmt.Println(query)
    rows,err := Dbstruct.DB.Query(query)
    defer func() {
        if rows != nil {
            rows.Close()
            Dbstruct.where = Dbstruct.where[0:0]
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
    
    return result

}
func(sqlstruct *Sqlstruct) Find (find interface{}) interface{} {
    query := SelectSql()
    fmt.Println(query)
    rows := Dbstruct.DB.QueryRow(query)
    defer func() {
            Dbstruct.where = Dbstruct.where[0:0]
        
    }()
    result := make([]interface{},0)
    dest := reflect.ValueOf(find)

    el := dest.Elem()
    len := el.NumField()
    vals := make([]interface{},len)
    for i,_ := range vals{
        vals[i] = el.Field(i).Addr().Interface()
    }

    rows.Scan(vals...)

    result = append(result,el.Interface())
    

    return result
}
func(sqlstruct *Sqlstruct) Update ( update map[string]string ) sql.Result{
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
    query := "update "+Dbstruct.table + params
    relust,err := Dbstruct.DB.Exec(query)
    if err != nil{
        fmt.Println("Query failed,err:%v", err)
    }
    return relust
}

func(sqlstruct *Sqlstruct) Save (save map[string]string ) sql.Result {
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
    Dbstruct.save = into + value
    query := "insert "+Dbstruct.table + into + value
    result,err := Dbstruct.DB.Exec(query)
    if err != nil{
        fmt.Println("Query failed,err:%v", err)
    }
    return result
}
func GetSave() string{
    return Dbstruct.save
}
func(sqlstruct *Sqlstruct) Del () {

}
func(sqlstruct *Sqlstruct) Limit(limit string) {
    if limit != ""{
        Dbstruct.limit = " limit "+ limit +""
    }
}
func GetLimit() string{
    return Dbstruct.limit
}
func Group(group string) {
    if group != ""{
        Dbstruct.group = " group by "+group

    }
    
}
func GetGroup() string{
    return Dbstruct.group
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
    if Dbstruct.join != nil{
        join += "join "
        params :=  Dbstruct.join

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