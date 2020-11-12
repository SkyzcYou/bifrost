package bifrost

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ClessLi/bifrost/internal/pkg/password"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// 认证接口错误返回
	ErrorReasonServerBusy    = "服务器繁忙"
	ErrorReasonRelogin       = "请重新登陆"
	ErrorReasonWrongPassword = "用户或密码错误"
	//ErrorReasonNoneToken     = "请通过认证"
)

// JWTClaims, jwt断言对象，定义认证接口校验的用户信息
type JWTClaims struct { // token里面添加用户信息，验证token后可能会用到用户信息
	jwt.StandardClaims
	UserID      int      `json:"user_id"`
	Password    string   `json:"password"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Permissions []string `json:"permissions"`
}

var (
	ExpireTime = 3600 // token有效期
)

// verifyAction, 认证token有效性函数
// 参数:
//     strToken: token字符串
// 返回值:
//     用户jwt断言对象指针
//     错误
func verifyAction(strToken string) (*JWTClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(strToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(password.Secret), nil
	})
	if err != nil {
		Log(WARN, err.Error())
		//return nil, errors.New(ErrorReasonServerBusy)
		return nil, errors.New(ErrorReasonRelogin)
	}

	// 转换jwt断言对象
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New(ErrorReasonRelogin)
	}
	Log(INFO, "Verify user '%s'...", claims.Username)

	// 认证用户信息
	if !validUser(claims) {
		Log(WARN, "Invalid user '%s' or password '%s'.", claims.Username, claims.Password)
		return nil, errors.New(ErrorReasonWrongPassword)
	}

	if err := token.Claims.Valid(); err != nil {
		return nil, errors.New(ErrorReasonRelogin)
	}
	Log(INFO, "Username '%s' passed verification", claims.Username)

	// 通过返回有效用户jwt断言对象
	return claims, nil
}

// getToken, token生成函数，根据jwt断言对象编码为token
// 参数:
//     claims: 用户jwt断言对象指针
// 返回值:
//     token字符串
//     错误
func getToken(claims *JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(password.Secret))
	if err != nil {
		Log(WARN, err.Error())
		return "", errors.New(ErrorReasonServerBusy)
	}
	return signedToken, nil
}

// validUser, 用户认证函数，判断用户是否有效
// 参数:
//     claims: 用户jwt断言对象指针
// 返回值:
//     用户是否有效
func validUser(claims *JWTClaims) bool {
	if authDBConfig == nil {
		return claims.Username == authConfig.Username && claims.Password == authConfig.Password
	}
	sqlStr := fmt.Sprintf("SELECT `password` FROM `%s`.`user` WHERE `user_name` = \"%s\" LIMIT 1;", authDBConfig.DBName, claims.Username)
	checkPasswd, err := getPasswd(sqlStr)
	if err != nil && err != sql.ErrNoRows {
		Log(ERROR, err.Error())
		return false
	} else if err == sql.ErrNoRows {
		Log(NOTICE, "user '%s' is not exist in bifrost", claims.Username)
		return false
	}

	return password.Password(claims.Password) == checkPasswd
}

// getPasswd, 用户密码查询函数
// 参数:
//     sqlStr: 查询语句
// 返回值:
//     用户加密密码
//     错误
func getPasswd(sqlStr string) (string, error) {
	mysqlUrl := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8", authDBConfig.User, authDBConfig.Password, authDBConfig.Protocol, authDBConfig.Host, authDBConfig.Port, authDBConfig.DBName)
	//fmt.Println(mysqlUrl)
	db, dbConnErr := sql.Open("mysql", mysqlUrl)
	if dbConnErr != nil {
		Log(ERROR, dbConnErr.Error())
		return "", dbConnErr
	}

	defer db.Close()

	rows, queryErr := db.Query(sqlStr)
	if queryErr != nil {
		Log(WARN, queryErr.Error())
		return "", queryErr
	}

	_, rowErr := rows.Columns()
	if rowErr == sql.ErrNoRows {
		return "", rowErr
	}

	for rows.Next() {
		var passwd string
		scanErr := rows.Scan(&passwd)
		if scanErr != nil {
			Log(WARN, scanErr.Error())
			return "", scanErr
		}

		if passwd != "" {
			return passwd, nil
		}
	}

	return "", errors.New("sql: unkown error")
}
