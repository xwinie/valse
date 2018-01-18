package casbin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/dgrijalva/jwt-go"
	"github.com/xwinie/valse"
)

//Config casbin需要的配置
type Config struct {
	model string
	User  string
	f     func(string) ([]Permission, error) //基于token的Authorization来进行数据获取
	Open  bool                               // 0代表非开放 1代表开放接口
}

// Permission 需要的权限结果集合
type Permission struct {
	ID     string
	Action string
	Method string
}

//Authorizer 权限结构
type Authorizer struct {
	enforcer *casbin.Enforcer
}

//RestAuth api sign auth
func RestAuth(c Config, next valse.RequestHandler) valse.RequestHandler {
	return func(ctx *valse.Context) error {
		if string(ctx.HeaderParameter("appid")) == "" {
			return valse.NewHTTPMessage(http.StatusForbidden, "miss appid header")
		}
		var e = casbin.NewEnforcer(casbin.NewModel(c.model))
		a := &Authorizer{enforcer: e}

		permissions, err := c.f(string(ctx.HeaderParameter("Authorization")))
		if err != nil {
			return valse.NewHTTPMessage(http.StatusForbidden, "Auth Fail")
		}
		for _, v := range permissions {
			e.AddPermissionForUser(c.User, v.Action, v.Method)
		}
		if !c.Open && !a.CheckPermission(c.User, string(ctx.Method()), string(ctx.Path())) {
			return valse.NewHTTPMessage(http.StatusForbidden, "Auth Fail")
		} else if c.Open && a.CheckPermission(c.User, string(ctx.Method()), string(ctx.Path())) {
			return next(ctx)

		}
		return next(ctx)
	}

}

// CheckPermission checks the user/method/path combination from the request.
// Returns true (permission granted) or false (permission forbidden)
func (a *Authorizer) CheckPermission(user, method, path string) bool {
	return a.enforcer.Enforce(user, path, method)
}

//ParseToken 解析token
func ParseToken(authString, secret string) (*jwt.Token, error) {
	if strings.Split(authString, " ")[1] == "" {
		return nil, valse.NewHTTPMessage(http.StatusForbidden, "AuthString invalid,Token:"+authString)
	}

	kv := strings.Split(authString, " ")
	if len(kv) != 2 || kv[0] != "Bearer" {
		return nil, valse.NewHTTPMessage(http.StatusForbidden, "AuthString invalid,Token:"+authString)
	}
	tokenString := kv[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, valse.NewHTTPMessage(http.StatusForbidden, "That's not even a token")
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, valse.NewHTTPMessage(http.StatusForbidden, "Token is either expired or not active yet")
			} else {
				return nil, valse.NewHTTPMessage(http.StatusForbidden, "Couldn‘t handle this token")
			}
		} else {
			return nil, valse.NewHTTPMessage(http.StatusForbidden, "Parse token is error")
		}
	}
	if !token.Valid {
		return nil, valse.NewHTTPMessage(http.StatusForbidden, "Token invalid:"+authString)
	}

	return token, nil
}
